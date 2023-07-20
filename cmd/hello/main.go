package main

import (
	"net/http"
	"time"
)

func main() {
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
