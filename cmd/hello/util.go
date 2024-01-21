package main

import (
	"github.com/xing393939/gotools/pkg/functrace"
	"sync"
)

func A1() {
	defer functrace.Trace()()

	B1()
}

func B1() {
	defer functrace.Trace()()

	C1()
}

func C1() {
	defer functrace.Trace()()

	D()
}

func A2() {
	defer functrace.Trace()()

	B2()
}

func B2() {
	defer functrace.Trace()()

	C2()
}

func C2() {
	defer functrace.Trace()()

	D()
}

func D() {
	defer functrace.Trace()()
	println("D")
}

func mainFunc() {
	defer functrace.Trace()()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		A1()
		wg.Done()
		A2()
	}()

	A2()
	wg.Wait()
}
