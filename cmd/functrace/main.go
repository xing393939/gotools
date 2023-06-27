package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/xing393939/gotools/pkg/generator"
)

func main() {
	if len(os.Args) != 2 {
		println("usage: functrace /path/to/your/folder")
		return
	}

	err := filepath.Walk(os.Args[1], func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".go" {
			newSrc, _ := generator.Rewrite(path)
			if len(newSrc) == 0 {
				return nil
			}
			if err = ioutil.WriteFile(path, newSrc, 0666); err != nil {
				panic(err)
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}
