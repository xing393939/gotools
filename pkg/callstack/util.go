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

func LogPrint(bi *proc.BinaryInfo, gId, duration, gIndents int64, pc, ret uint64, args, evals []*proc.Variable) {
	fnFile, fnLine, fnObj := bi.PCToLine(pc)
	fnName := fnObj.Name
	if gIndents == 0 {
		fnName = fmt.Sprintf("goroutine-%d created by %s", gId, fnName)
	}
	retFile, retLine, _ := bi.PCToLine(ret)
	actionMain := []int64{
		gId, duration, gIndents,
		logBody.FuncNameList.Insert(fnName),
		logBody.FileList.Insert(fnFile), int64(fnLine),
		logBody.FileList.Insert(retFile), int64(retLine),
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
	for _, eval := range evals {
		str := api.ConvertVar(eval).MultilineString("", "")
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

func GetAddrByPath(bi *proc.BinaryInfo, bp string) (uint64, []string) {
	paths := strings.Split(strings.TrimSpace(bp), " ")
	v := strings.Split(paths[0], ":")
	if len(v) > 2 {
		v = []string{strings.Join(v[0:len(v)-1], ":"), v[len(v)-1]}
	}
	if len(v) != 2 {
		fnList, _ := bi.FindFunction(paths[0])
		if len(fnList) > 0 {
			return fnList[0].Entry, paths[1:]
		}
		return 0, paths[1:]
	}
	line, _ := strconv.Atoi(v[1])
	rs := bi.AllPCsForFileLines(v[0], []int{line})
	if len(rs[line]) > 0 {
		return rs[line][0], paths[1:]
	}
	return 0, paths[1:]
}
