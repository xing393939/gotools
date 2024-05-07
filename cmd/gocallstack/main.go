// 参考：https://bye04t46if.execute-api.ap-southeast-1.amazonaws.com/default/Web?url=https://medium.com/golangspec/making-debugger-for-golang-part-i-53124284b7c8
package main

import (
	"flag"
	"fmt"
	"github.com/go-delve/delve/pkg/proc"
	"github.com/go-delve/delve/pkg/proc/native"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var logFormat = "%10d%12.6f %s%s at %s#L%d\n"
var logBody strings.Builder
var fCount = make(map[uint64]uint64)
var gStack = make(map[int64][]int64)
var gAddr = make(map[int64]uint64)
var gFile *os.File
var start = time.Now()

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gocallstack [exe|pid]")
		return
	}
	packageIncluded := flag.String("p", "", "included package")
	packageExcluded := flag.String("P", "", "excluded package")
	flag.Parse()

	// 编译正则表达式
	regIncluded, regErr1 := regexp.Compile(*packageIncluded)
	regExcluded, regErr2 := regexp.Compile(*packageExcluded)
	if regErr1 != nil || regErr2 != nil {
		fmt.Println("Error compiling regex:", regErr1, regErr2)
		return
	}

	// 挂载debug程序
	killFlag := [2]bool{false, true}
	targetGroup, err := native.Launch(flag.Args(), "", 0, nil, "", [3]string{})
	if err != nil {
		pid, _ := strconv.Atoi(flag.Args()[0])
		targetGroup, err = native.Attach(pid, nil)
		if err != nil {
			fmt.Println("exe|pid not found")
			return
		}
		killFlag[1] = false
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-quit
		killFlag[0] = true
		_ = targetGroup.RequestManualStop()
	}()

	fnList := make([]uint64, 0, len(targetGroup.Selected.BinInfo().Functions))
	for _, fn := range targetGroup.Selected.BinInfo().Functions {
		if fn.Entry == 0 {
			continue
		}

		switch fn.Name {
		case "gosave_systemstack_switch", "gogo":
			continue
		case "aeshashbody", "indexbytebody", "countbody", "cmpbody", "indexbody", "memeqbody":
			continue
		}

		fnPackageName := fn.PackageName()
		switch fnPackageName {
		case "encoding/json", "compress/flate", "internal/bytealg", "regexp/syntax":
			continue
		case "reflect", "strings", "runtime", "syscall", "regexp":
			continue
		}

		if len(*packageIncluded) > 0 && !regIncluded.MatchString(fnPackageName) && fnPackageName != "main" {
			continue
		}
		if len(*packageExcluded) > 0 && regExcluded.MatchString(fnPackageName) {
			continue
		}
		fnList = append(fnList, fn.Entry)
	}

	fmt.Printf("SetBreakpoint")
	for bid, fn := range fnList {
		_, err = targetGroup.Selected.SetBreakpoint(bid, fn, proc.UserBreakpoint, nil)
		if err != nil {
			fmt.Printf("\rSetBreakpoint err\n")
		}
		if bid%100 == 0 {
			fmt.Printf("\rSetBreakpoint %d/%d", bid, len(fnList))
		}
	}
	fmt.Printf("\rSetBreakpoint %d/%d\n", len(fnList), len(fnList))

	err = targetGroup.Continue()
	for err == nil {
		if killFlag[0] {
			_ = targetGroup.Detach(killFlag[1])
			err = fmt.Errorf("manual stop")
			break
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

			indents := getIndents(goroutine, &stackFlames[0], targetGroup.Selected.BinInfo())
			duration := time.Since(start).Seconds()
			logPrint(
				logFormat, goroutine.ID, duration, indents,
				breakpoint.FunctionName, breakpoint.File, breakpoint.Line,
			)
		}
		err = targetGroup.Continue()
	}
	printTop10Func(targetGroup.Selected.BinInfo())
	uploadToS3()
	fmt.Printf("%s\n", err.Error())
}

func logPrint2(format string, args ...any) {
	if gFile == nil {
		gFile, _ = os.Create(fmt.Sprintf("stack.log"))
	}
	_, _ = fmt.Fprintf(gFile, format, args...)
}

func logPrint(format string, args ...any) {
	logBody.WriteString(fmt.Sprintf(format, args...))
}

func uploadToS3() {
	host := "https://5xfd05tkng.execute-api.cn-northwest-1.amazonaws.com.cn/callstack"
	resp, err := http.PostForm(host, url.Values{"body": {logBody.String()}})
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("%s?demo=%s\n", host, string(body))
}

func printTop10Func(bi *proc.BinaryInfo) {
	tmpSlice := make([][2]uint64, 0, len(fCount))
	for k, v := range fCount {
		tmpSlice = append(tmpSlice, [2]uint64{k, v})
	}
	sort.Slice(tmpSlice, func(i, j int) bool {
		return tmpSlice[i][1] > tmpSlice[j][1]
	})
	fmt.Println()
	for _, row := range tmpSlice[:10] {
		fnObj := bi.PCToFunc(row[0])
		fmt.Printf("%10d %s\n", row[1], fnObj.Name)
	}
}

func getIndents(g *proc.G, sf *proc.Stackframe, bi *proc.BinaryInfo) string {
	// 统计函数调用次数
	if _, ok := fCount[sf.Call.PC]; ok {
		fCount[sf.Call.PC]++
	} else {
		fCount[sf.Call.PC] = 1
	}

	gSlice, ok := gStack[g.ID]
	offset := sf.FramePointerOffset()
	if !ok {
		gSlice = make([]int64, 1)
		gSlice[0] = 1
		gStack[g.ID] = gSlice
		gIndents := fmt.Sprintf("goroutine-%d created by ", g.ID)
		if g.StartPC == sf.Call.PC {
			return gIndents
		}
		duration := time.Since(start).Seconds()
		fnObj := bi.PCToFunc(g.StartPC)
		file, line := bi.EntryLineForFunc(fnObj)
		logPrint(logFormat, g.ID, duration, gIndents, fnObj.Name, file, line)
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
