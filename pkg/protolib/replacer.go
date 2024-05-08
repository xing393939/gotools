package protolib

import (
	"bufio"
	"io"
	"os"
	"strings"
)

func ReplaceOmits(f string) (err error) {
	var (
		file *os.File
		temp *os.File
		line string
	)

	// Open the source file
	file, err = os.OpenFile(f, os.O_RDWR, 0644)
	if err != nil {
		return
	}

	// Create a temp file for saving stripped text
	temp, err = os.CreateTemp("", "replaced")
	if err != nil {
		return
	}
	defer func() {
		_ = file.Close()
		_ = temp.Close()
		if err != nil {
			err = os.Remove(temp.Name())
		} else {
			err = os.Rename(temp.Name(), file.Name())
		}
	}()

	reader := bufio.NewReader(file)
	writer := bufio.NewWriter(temp)

	// Going through lines, match the regex and replace omitempty on primitives and enums
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}

		cleaned := strings.Replace(line, ",omitempty", "", -1)
		_, err = writer.WriteString(cleaned)
		if err != nil {
			return
		}

		err = writer.Flush()
		if err != nil {
			return
		}
		err = temp.Sync()
		if err != nil {
			return
		}
	}
}
