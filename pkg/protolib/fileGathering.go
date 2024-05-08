package protolib

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const filenameEnd = `.pb.go`

func SearchFiles(folder string, recursive bool) ([]string, error) {
	files := make([]string, 0)

	folderPath := path.Clean(folder)

	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && !strings.EqualFold(folderPath, folder) && !recursive {
			return filepath.SkipDir
		}
		if strings.HasSuffix(info.Name(), filenameEnd) {
			files = append(files, path)
			log.Println(path)
		}

		return nil
	})

	return files, err
}
