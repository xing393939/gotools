// 参考：https://bye04t46if.execute-api.ap-southeast-1.amazonaws.com/default/Web?url=https://medium.com/golangspec/making-debugger-for-golang-part-i-53124284b7c8
package main

import (
	"fmt"
	"github.com/go-delve/delve/pkg/proc"
	"github.com/go-delve/delve/pkg/proc/native"
	"os"
	"strings"
)

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
		for bid, fn := range target.Process.BinInfo().Functions {
			if fn.Entry == 0 || strings.HasPrefix(fn.Name, "runtime.") {
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
		for _, target := range targetList {
			gid := int64(0)
			tid := target.CurrentThread().ThreadID()
			bp := target.CurrentThread().Breakpoint().Breakpoint
			if bp == nil {
				continue
			}
			if target.SelectedGoroutine() != nil {
				gid = target.SelectedGoroutine().ID
			}
			fmt.Printf("%8d %8d %s-%x\n", tid, gid, bp.FunctionName, bp.Addr)
		}
		err = targetGroup.Next()
		if err != nil && err.Error() == "next while nexting" {
			// 线程正在nexting的时候被中断，如果继续Next()会报错，所以用Continue()
			err = targetGroup.Continue()
		}
	}
	fmt.Println(err.Error())
}
