package main

import (
	"flag"
	"github.com/xing393939/gotools/pkg/protolib"
	"log"
)

var folder = flag.String("folder", "", "Specify the folder with the proto files. Only .pb.go-Files are processed")
var recursive = flag.Bool("recursive", false, "Recursive folder traversal")
var file = flag.String("file", "", "Specify only when folder not set. Processes the specified file only")

func main() {
	flag.Parse()

	var err error
	var files []string
	if *file != "" {
		files = append(files, *file)
	} else {
		files, err = protolib.SearchFiles(*folder, *recursive)
		if err != nil {
			log.Fatal("Error Traversing folder ", err)
		}
	}

	for _, v := range files {
		err = protolib.ReplaceOmits(v)
		if err != nil {
			log.Println(err)
		}
	}

	log.Println("Done. Good bye.")
}
