package callstack

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type logBodyStruct struct {
	ActionList   [][2][]int64
	FuncNameList *UniqueList
	FileList     *UniqueList
}

var logBody = logBodyStruct{
	FuncNameList: NewUniqueList(),
	FileList:     NewUniqueList(),
}

func LogPrint(gId, duration, gIndents int64, funcName, file string, line int) {
	if gIndents == 0 {
		funcName = fmt.Sprintf("goroutine-%d created by %s", gId, funcName)
	}
	actionMain := []int64{
		gId, duration, gIndents,
		logBody.FuncNameList.Insert(funcName),
		logBody.FileList.Insert(file), int64(line),
	}
	action := [2][]int64{
		actionMain,
		nil,
	}
	logBody.ActionList = append(logBody.ActionList, action)
}

func UploadToS3() {
	host := "https://5xfd05tkng.execute-api.cn-northwest-1.amazonaws.com.cn/callstack"
	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	logBodyBytes, _ := json.Marshal(logBody)
	if _, err := g.Write(logBodyBytes); err != nil {
		return
	}
	if err := g.Close(); err != nil {
		return
	}
	req, _ := http.NewRequest("POST", host, &buf)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "identity")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Webui: %s?demo=%s\n", host, string(body))
}

func PrintDebug(isDebug bool) {
	if !isDebug {
		return
	}
	gFile, _ := os.Create(fmt.Sprintf("stack.log"))
	logBodyBytes, _ := json.Marshal(logBody)
	_, _ = gFile.Write(logBodyBytes)
}
