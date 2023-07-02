// 参考：https://bye04t46if.execute-api.ap-southeast-1.amazonaws.com/default/Web?url=https://medium.com/golangspec/making-debugger-for-golang-part-i-53124284b7c8
package main

import (
	"fmt"
	"github.com/go-delve/delve/pkg/proc"
	"github.com/go-delve/delve/pkg/proc/native"
	"os"
	"strings"
)

var gStack = make(map[int64][]int64)
var gFiles = make(map[int64]*os.File)

func main() {
	if len(os.Args) != 2 {
		println("usage: gocallstack /path/to/exe")
		return
	}
	targetGroup, err := native.Launch(os.Args[1:], "", 0, []string{}, "", [3]string{})
	if err != nil {
		panic(err)
	}
	targetList := targetGroup.Targets()
	for _, target := range targetList {
		for bid, fn := range target.BinInfo().Functions {
			if fn.Entry == 0 || strings.HasPrefix(fn.Name, "runtime.") {
				continue
			}
			if fn.Name == "gosave_systemstack_switch" {
				continue
			}
			if fn.Name == "internal/bytealg.IndexByteString" {
				continue
			}
			if fn.Name == "indexbytebody" {
				continue
			}
			if fn.Name == "aeshashbody" {
				continue
			}

			_, err = target.SetBreakpoint(bid, fn.Entry, proc.UserBreakpoint, nil)
			if err != nil {
				fmt.Println(fn.Name, err.Error())
			}
		}
	}

	err = targetGroup.Continue()
	for err == nil {
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
			indents := getIndents(goroutine.ID, stackFlames[0].FramePointerOffset())

			printf(goroutine.ID, "%10d %s%s\n", goroutine.ID, indents, breakpoint.FunctionName)
		}
		err = targetGroup.Continue()
	}
	fmt.Println(err.Error())
}

func printf(gid int64, format string, args ...any) {
	gFile, ok := gFiles[gid]
	if !ok {
		gFile, _ = os.Create(fmt.Sprintf("gocallstack-%d.sql", gid))
		gFiles[gid] = gFile
	}
	_, _ = fmt.Fprintf(gFile, format, args...)
}

func getIndents(gid int64, offset int64) string {
	gSlice, ok := gStack[gid]
	if !ok {
		gSlice = make([]int64, 0)
		gStack[gid] = gSlice
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
	gStack[gid] = gSlice
	return indents
}
