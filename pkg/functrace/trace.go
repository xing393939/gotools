//go:build trace
// +build trace

package functrace

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var (
	// 因为map的key是协程id，是可以保证不会存在竞争的
	indentMap = sync.Map{}
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
	millTime := time.Now().Format("150405.000")
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
		fmt.Printf("%s%8d %s%s()", millTime, id, indents, name)
	} else if typ == 0 && prevTyp == currTyp+1 {
		fmt.Printf("\n")
	} else {
		fmt.Printf("%s%8d %s}\n", millTime, id, indents)
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

	indentObj, _ := indentMap.LoadOrStore(id, 0)
	indentVal := indentObj.(int)
	indentMap.Store(id, indentVal+1)
	printTrace(id, 1, name, indentVal+1)
	return func() {
		indentObj, _ = indentMap.LoadOrStore(id, 0)
		indentVal = indentObj.(int)
		indentMap.Store(id, indentVal-1)
		printTrace(id, 0, name, indentVal)
	}
}
