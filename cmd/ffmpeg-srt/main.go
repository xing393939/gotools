package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("无法获取当前工作目录:", err)
		return
	}

	wd = "C:\\Users\\bookan\\Downloads\\MIT RES.6-012 Introduction to Probability, Spring 2018麻省理工：概率学导论，2018春季"

	// 遍历当前工作目录下的所有子文件夹
	err = filepath.WalkDir(wd, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 检查文件夹名称是否以 "MIT RES.6-012" 开头
		if d.IsDir() && strings.HasPrefix(d.Name(), "MIT RES.6-012") {
			mergeFilesInDir(path)
		}
		return nil
	})

	if err != nil {
		fmt.Println("遍历文件夹时出错:", err)
	}
}

func mergeFilesInDir(dir string) {
	mp4Files := make(map[string]string)
	srtFiles := make(map[string]string)

	// 遍历子文件夹中的文件
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			if strings.HasSuffix(d.Name(), " - 超清.mp4") {
				baseName := strings.TrimSuffix(d.Name(), " - 超清.mp4")
				mp4Files[baseName] = path
			} else if strings.HasSuffix(d.Name(), " - 中文字幕.srt") {
				baseName := strings.TrimSuffix(d.Name(), " - 中文字幕.srt")
				srtFiles[baseName] = path
			} else if strings.HasSuffix(d.Name(), ".txt") {
				_ = os.Remove(path)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("遍历子文件夹时出错:", err)
		return
	}
	_ = os.Mkdir(filepath.Join(dir, "../dist"), os.ModeDir)
	// 查找并合并成对的文件
	for baseName, mp4File := range mp4Files {
		if srtFile, exists := srtFiles[baseName]; exists {
			outputFile := filepath.Join(dir, "../dist", baseName+"_merged.mkv")
			execFile := "C:\\Users\\bookan\\scoop\\shims\\ffmpeg.exe"
			cmd := exec.Command(
				execFile, "-i", mp4File, "-i", srtFile,
				"-c", "copy", "-c:s", "srt", outputFile)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err = cmd.Run(); err != nil {
				fmt.Println("合并文件时出错:", err)
			} else {
				fmt.Println("合并成功:", outputFile)
			}
		}
	}
}
