package main

//go:noinline
func funcA(a int) int {
	return a + 1
}

func main() {
	a := 1
	a = funcA(a)
	println(a)
}
