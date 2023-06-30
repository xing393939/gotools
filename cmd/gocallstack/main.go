// 参考：https://bye04t46if.execute-api.ap-southeast-1.amazonaws.com/default/Web?url=https://medium.com/golangspec/making-debugger-for-golang-part-i-53124284b7c8
package main

import (
	"fmt"
	"github.com/go-delve/delve/pkg/proc"
	"github.com/go-delve/delve/pkg/proc/native"
	"os"
	"strings"
)

var gMap = make(map[int64][]uint64)

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

			breakpoint = thread.Breakpoint().Breakpoint
			regs, _ := thread.Registers()
			indents := getIndents(goroutine.ID, regs.SP())

			fmt.Printf("%10d %s %s\n", goroutine.ID, indents, breakpoint.FunctionName)
		}
		err = targetGroup.Continue()
	}
	fmt.Println(err.Error())
}

func getIndents(gid int64, gsp uint64) string {
	gSlice, ok := gMap[gid]
	if !ok {
		gSlice = make([]uint64, 1)
		gMap[gid] = gSlice
	}

	indents := ""
	for _, sp := range gSlice {
		if gsp < sp {
			indents = indents + " "
		}
	}
	indentLen := len(indents)
	if indentLen < len(gSlice) {
		gSlice[indentLen] = gsp
		gSlice = gSlice[:indentLen+1]
	} else {
		gSlice = append(gSlice, gsp)
	}
	gMap[gid] = gSlice
	return indents
}
