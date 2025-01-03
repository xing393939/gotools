// 参考：https://bye04t46if.execute-api.ap-southeast-1.amazonaws.com/default/Web?url=https://medium.com/golangspec/making-debugger-for-golang-part-i-53124284b7c8
package main

import (
	"flag"
	"fmt"
	"github.com/go-delve/delve/pkg/proc"
	"github.com/go-delve/delve/pkg/proc/native"
	"github.com/xing393939/gotools/pkg/callstack"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strconv"
	"syscall"
	"time"
)

var fCount = make(map[uint64]uint64)
var gStack = make(map[int64][]int64)
var gAddr = make(map[int64]*proc.Stackframe)
var start = time.Now()

func main() {
	var importantBreakpoints callstack.MultiString
	flag.Var(&importantBreakpoints, "bp", "important breakpoints")
	packageIncluded := flag.String("p", "", "included package")
	packageExcluded := flag.String("P", "", "excluded package")
	isDebug := flag.Bool("debug", false, "save debug log")
	flag.Usage = func() {
		fmt.Println("Usage: gocallstack exe|pid")
		flag.PrintDefaults()
	}
	flag.Parse()

	// 编译正则表达式
	regIncluded, regErr1 := regexp.Compile(*packageIncluded)
	regExcluded, regErr2 := regexp.Compile(*packageExcluded)
	if regErr1 != nil || regErr2 != nil {
		fmt.Println("Error compiling regex:", regErr1, regErr2)
		return
	}
	if flag.NArg() < 1 {
		flag.Usage()
		return
	}

	// 挂载debug程序
	killFlag := [2]bool{false, true}
	targetGroup, err := native.Launch(flag.Args(), "", 0, nil, "", [3]string{})
	if err != nil {
		pid, _ := strconv.Atoi(flag.Arg(0))
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
	importantAddrMap := make(map[uint64][]string)
	for _, path := range importantBreakpoints {
		fn, exprArr := callstack.GetAddrByPath(targetGroup.Selected.BinInfo(), path)
		if fn > 0 {
			importantAddrMap[fn] = exprArr
			fnList = append(fnList, fn)
		} else {
			fmt.Println("Unknown bp:", path)
		}
	}
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
		if _, ok := importantAddrMap[fn.Entry]; ok {
			continue
		}
		fnList = append(fnList, fn.Entry)
	}

	fmt.Printf("SetBreakpoint")
	for bid, fn := range fnList {
		_, err = targetGroup.Selected.SetBreakpoint(bid, fn, proc.UserBreakpoint, nil)
		if err != nil {
			fmt.Printf("\rSetBreakpoint err: %s\n", err)
		}
		if bid%100 == 0 {
			fmt.Printf("\rSetBreakpoint %d/%d", bid, len(fnList))
		}
	}
	fmt.Printf("\rSetBreakpoint %d/%d\n", len(fnList), len(fnList))

	evalCfg := proc.LoadConfig{MaxStringLen: 64, MaxStructFields: 3}
	evalImportantCfg := proc.LoadConfig{
		FollowPointers:     true,
		MaxVariableRecurse: 1,
		MaxStringLen:       256,
		MaxArrayValues:     64,
		MaxStructFields:    64,
		MaxMapBuckets:      64,
	}
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
			gCurr := &stackFlames[0]
			if gPrev, ok := gAddr[goroutine.ID]; ok &&
				gPrev.FramePointerOffset() == gCurr.FramePointerOffset() &&
				gPrev.Call.PC == gCurr.Call.PC &&
				gPrev.Ret == gCurr.Ret {
				continue
			}
			gAddr[goroutine.ID] = gCurr

			evalScope := proc.FrameToScope(targetGroup.Selected, thread.ProcessMemory(), nil, stackFlames...)
			args, _ := evalScope.FunctionArguments(evalCfg)
			var evalList []*proc.Variable
			if exprArr, ok := importantAddrMap[breakpoint.Addr]; ok {
				for _, expr := range exprArr {
					evalV, evalE := evalScope.EvalExpression(expr, evalImportantCfg)
					if evalE != nil {
						evalV = &proc.Variable{}
					}
					evalList = append(evalList, evalV)
				}
			}

			duration := time.Since(start).Microseconds()
			indents, startPC := getIndents(goroutine, gCurr)
			if startPC > 0 {
				callstack.LogPrint(
					targetGroup.Selected.BinInfo(),
					goroutine.ID, duration, 0, startPC, 0, nil, nil,
				)
			}
			callstack.LogPrint(
				targetGroup.Selected.BinInfo(),
				goroutine.ID, duration, indents, gCurr.Call.PC, gCurr.Ret, args, evalList,
			)
		}
		err = targetGroup.Continue()
	}
	printTop10Func(targetGroup.Selected.BinInfo())
	callstack.PrintDebug(*isDebug)
	callstack.UploadToS3()
	fmt.Printf("Error: %s\n", err.Error())
}

func printTop10Func(bi *proc.BinaryInfo) {
	tmpSlice := make([][2]uint64, 0, len(fCount))
	for k, v := range fCount {
		tmpSlice = append(tmpSlice, [2]uint64{k, v})
	}
	sort.Slice(tmpSlice, func(i, j int) bool {
		return tmpSlice[i][1] > tmpSlice[j][1]
	})
	if len(tmpSlice) > 10 {
		tmpSlice = tmpSlice[:10]
	}
	fmt.Println("Frequently callee top 10:")
	for _, row := range tmpSlice {
		fnObj := bi.PCToFunc(row[0])
		fnPkg := fnObj.PackageName()
		fmt.Printf("%10d %s %s\n", row[1], fnPkg, fnObj.Name[len(fnPkg):])
	}
}

func getIndents(g *proc.G, sf *proc.Stackframe) (int64, uint64) {
	startPC := uint64(0)
	// 统计函数调用次数
	if _, ok := fCount[sf.Call.PC]; ok {
		fCount[sf.Call.PC]++
	} else {
		fCount[sf.Call.PC] = 1
	}

	gSlice, ok := gStack[g.ID]
	if !ok {
		gSlice = make([]int64, 1)
		gSlice[0] = 1
		gStack[g.ID] = gSlice
		if g.StartPC == sf.Call.PC {
			return 0, 0
		}
		startPC = g.StartPC
	}

	indentLen := 0
	offset := sf.FramePointerOffset()
	for _, sp := range gSlice {
		if offset < sp {
			indentLen = indentLen + 1
		}
	}
	if indentLen < len(gSlice) {
		gSlice[indentLen] = offset
		gSlice = gSlice[:indentLen+1]
	} else {
		gSlice = append(gSlice, offset)
	}
	gStack[g.ID] = gSlice
	return int64(indentLen), startPC
}
