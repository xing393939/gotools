package main

import (
	"github.com/go-delve/delve/pkg/proc"
	"github.com/go-delve/delve/pkg/proc/native"
)

func main() {
	targetGroup, err := native.Launch([]string{"./hello"}, "", 0, nil, "", [3]string{})
	fn, _ := targetGroup.Selected.BinInfo().FindFunction("main.main")
	_, err = targetGroup.Selected.SetBreakpoint(0, fn[0].Entry, proc.UserBreakpoint, nil)
	err = targetGroup.Continue()
	println(err)
	_ = targetGroup.Detach(true)
}
