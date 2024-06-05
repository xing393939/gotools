package callstack

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/go-delve/delve/pkg/proc"
	"github.com/go-delve/delve/service/api"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type logBodyStruct struct {
	ActionList   [][3][]int64
	FuncNameList *UniqueList
	FuncArgList  *UniqueList
	FileList     *UniqueList
}

var logBody = logBodyStruct{
	FuncNameList: NewUniqueList(),
	FuncArgList:  NewUniqueList(),
	FileList:     NewUniqueList(),
}

func LogPrint(gId, duration, gIndents int64, funcName, file string, line int, args []*proc.Variable, evals *proc.Variable) {
	if gIndents == 0 {
		funcName = fmt.Sprintf("goroutine-%d created by %s", gId, funcName)
	}
	actionMain := []int64{
		gId, duration, gIndents,
		logBody.FuncNameList.Insert(funcName),
		logBody.FileList.Insert(file), int64(line),
	}
	action := [3][]int64{
		actionMain,
		nil,
		nil,
	}
	for _, arg := range args {
		if (arg.Flags & proc.VariableArgument) != 0 {
			action[1] = append(action[1], logBody.FuncArgList.Insert(getArgStr(arg)))
		}
	}
	if evals != nil {
		str := api.ConvertVar(evals).MultilineString("", "")
		action[2] = append(action[2], logBody.FuncArgList.Insert(str))
	}
	logBody.ActionList = append(logBody.ActionList, action)
}

func getArgStr(arg *proc.Variable) (str string) {
	if arg.Kind == reflect.Interface && len(arg.Children) > 0 {
		str = arg.Name + "=" + arg.Children[0].TypeString()
	} else {
		str = arg.Name + "=" + arg.TypeString()
	}
	return
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

func GetAddrByPath(bi *proc.BinaryInfo, bp string) (uint64, string) {
	path, expr, _ := strings.Cut(bp, " ")
	filename, lineStr, _ := strings.Cut(path, ":")
	if lineStr == "" {
		fnList, _ := bi.FindFunction(path)
		if len(fnList) > 0 {
			return fnList[0].Entry, expr
		}
		return 0, expr
	}
	line, _ := strconv.Atoi(lineStr)
	rs := bi.AllPCsForFileLines(filename, []int{line})
	if len(rs[line]) > 0 {
		return rs[line][0], expr
	}
	return 0, expr
}
