// 参考：https://bye04t46if.execute-api.ap-southeast-1.amazonaws.com/default/Web?url=https://medium.com/golangspec/making-debugger-for-golang-part-i-53124284b7c8
package main

import (
	"fmt"
	"github.com/go-delve/delve/pkg/proc"
	"github.com/go-delve/delve/pkg/proc/native"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var gStack = make(map[int64][]int64)
var gAddr = make(map[int64]uint64)
var start = time.Now()

func main() {
	if len(os.Args) < 2 {
		logPrint("usage: gocallstack [exe|pid]\n")
		return
	}
	killFlag := [2]bool{false, true}
	targetGroup, err := native.Launch(os.Args[1:], "", 0, nil, "", [3]string{})
	if err != nil {
		pid, _ := strconv.Atoi(os.Args[1])
		targetGroup, err = native.Attach(pid, nil)
		if err != nil {
			logPrint("exe|pid not found\n")
			return
		}
		killFlag[1] = false
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-quit
		killFlag[0] = true
		_ = syscall.Kill(targetGroup.Selected.Pid(), syscall.SIGSTOP)
	}()

	for bid, fn := range targetGroup.Selected.BinInfo().Functions {
		if fn.Entry == 0 {
			continue
		}
		switch fn.Name {
		case "gosave_systemstack_switch", "gogo":
			continue
		case "aeshashbody", "indexbytebody", "countbody", "cmpbody", "indexbody", "memeqbody":
			continue
		}

		switch fn.PackageName() {
		case "encoding/json", "compress/flate", "internal/bytealg":
			continue
		case "reflect", "strings", "runtime", "syscall":
			continue
		}

		_, err = targetGroup.Selected.SetBreakpoint(bid, fn.Entry, proc.UserBreakpoint, nil)
		if err != nil {
			logPrint("%s %s\n", fn.Name, err.Error())
		}
	}

	err = targetGroup.Continue()
	for err == nil {
		if killFlag[0] {
			_ = targetGroup.Detach(killFlag[1])
			os.Exit(0)
		}

		breakpoint := targetGroup.Selected.CurrentThread().Breakpoint().Breakpoint
		if breakpoint == nil {
			continue
		}

		for _, thread := range targetGroup.ThreadList() {
			if thread.Breakpoint().Breakpoint == nil || !thread.Breakpoint().Active {
				continue
			}

			goroutine, _ := proc.GetG(thread)
			if goroutine == nil || goroutine.SystemStack {
				continue
			}

			stackFlames, _ := proc.ThreadStacktrace(thread, 1)
			if len(stackFlames) == 0 {
				continue
			}

			breakpoint = thread.Breakpoint().Breakpoint
			if gPrev, ok := gAddr[goroutine.ID]; ok && gPrev == breakpoint.Addr {
				continue
			}
			gAddr[goroutine.ID] = breakpoint.Addr

			indents := getIndents(goroutine, stackFlames[0].FramePointerOffset())
			duration := time.Since(start).Seconds()
			logPrint("%10d%12.6f %s%s at %s#L%d\n", goroutine.ID, duration, indents, breakpoint.FunctionName, breakpoint.File, breakpoint.Line)
		}
		err = targetGroup.Continue()
	}
	logPrint("%s\n", err.Error())
}

func logPrint(format string, args ...any) {
	fmt.Printf(format, args...)
}

func getIndents(g *proc.G, offset int64) string {
	gSlice, ok := gStack[g.ID]
	if !ok {
		gSlice = make([]int64, 1)
		gSlice[0] = 1
		gStack[g.ID] = gSlice

		if g.ID != 1 {
			return fmt.Sprintf("goroutine-%d created by ", g.ID)
		}
		duration := time.Since(start).Seconds()
		logPrint("%10d%12.6f goroutine-%d runtime.main\n", g.ID, duration, g.ID)
	}

	indents := ""
	for _, sp := range gSlice {
		if offset < sp {
			indents = indents + "."
		}
	}
	indentLen := len(indents)
	if indentLen < len(gSlice) {
		gSlice[indentLen] = offset
		gSlice = gSlice[:indentLen+1]
	} else {
		gSlice = append(gSlice, offset)
	}
	gStack[g.ID] = gSlice
	return indents
}
