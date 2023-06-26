package main

import "github.com/xing393939/gotools/pkg/functrace"

//go:noinline
func funcA(a int) int {
	defer functrace.Trace()()
	if a > 5 {
		return a + 1
	} else {
		return funcA(a + 1)
	}
}

func main() {
	defer functrace.Trace()()
	a := 1
	a = funcA(a)
	mainFunc()
	println(a)
}
