package main

import (
	"github.com/xing393939/gotools/pkg/functrace"
	"net/http"
	"time"
)

func main() {
	defer functrace.Trace()()
	client := http.Client{
		Timeout: 3 * time.Second,
	}
	res, err := client.Get("http://httpbin.org/delay/10")
	if err == nil {
		println(res.Body)
	} else {
		println(err.Error())
	}
}
