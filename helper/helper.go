package helper

import (
	"os"
	"io"
	"log"
)

func CreateDir(dir string) {
	err := os.MkdirAll(dir, os.ModeDir|os.ModePerm)
	if err != nil {
		log.Fatal(dir, " ", err)
	}
}

func CreateFile(path string) io.Writer {
	file, err := os.Create(path)
	if err != nil {
		log.Fatal(file, ":", err)
	}
	return file
}
