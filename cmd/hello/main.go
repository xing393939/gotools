package main

//go:noinline
func funcA(a int) int {

	if a > 5 {
		return a + 1
	} else {
		return funcA(a + 1)
	}
}

func main() {

	a := 1
	a = funcA(a)
	mainFunc()
	println(a)
}
