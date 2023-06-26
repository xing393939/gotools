package functrace

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"sync"
)

var (
	// 因为map的key是协程id，是可以保证不会存在竞争的
	indentMap = make(map[uint64]int)
	prevTyp   = uint64(0)
	printLock = sync.Mutex{}
)

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func printTrace(id, typ uint64, name string, indent int) {
	printLock.Lock()
	indents := ""
	for i := 0; i < indent; i++ {
		indents += " "
	}
	currTyp := id<<1 | typ
	// 同一协程两次都是enter
	if typ == 1 && prevTyp == currTyp {
		fmt.Printf(" {\n")
	}
	// 切换协程需要给大括号
	if prevTyp&1 == 1 && prevTyp != currTyp|1 {
		fmt.Printf(" {\n")
	}
	if typ == 1 {
		fmt.Printf("%8d %s%s()", id, indents, name)
	} else if typ == 0 && prevTyp == currTyp+1 {
		fmt.Printf("\n")
	} else {
		fmt.Printf("%8d %s}\n", id, indents)
	}
	prevTyp = currTyp
	printLock.Unlock()
}

func Trace() func() {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("not found caller")
	}

	id := getGID()
	fn := runtime.FuncForPC(pc)
	name := fn.Name()

	indent := indentMap[id]
	indentMap[id] = indent + 1
	printTrace(id, 1, name, indent+1)
	return func() {
		indent = indentMap[id]
		indentMap[id] = indent - 1
		printTrace(id, 0, name, indent)
	}
}
