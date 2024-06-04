package callstack

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/go-delve/delve/pkg/proc"
	"io"
	"net/http"
	"os"
)

type logBodyStruct struct {
	ActionList   [][2][]int64
	FuncNameList *UniqueList
	FuncArgList  *UniqueList
	FileList     *UniqueList
}

var logBody = logBodyStruct{
	FuncNameList: NewUniqueList(),
	FuncArgList:  NewUniqueList(),
	FileList:     NewUniqueList(),
}

func LogPrint(gId, duration, gIndents int64, funcName, file string, line int, args []*proc.Variable) {
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
	if len(args) > 0 {
		for _, arg := range args {
			if (arg.Flags & proc.VariableReturnArgument) != 0 {
				continue
			}
			if len(arg.Children) > 0 {
				i := logBody.FuncArgList.Insert(arg.Children[0].TypeString())
				action[1] = append(action[1], i)
			} else {
				i := logBody.FuncArgList.Insert(arg.TypeString())
				action[1] = append(action[1], i)
			}
		}
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