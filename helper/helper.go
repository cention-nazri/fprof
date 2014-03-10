package helper

import (
	"os"
	"bufio"
	"io"
	"log"
)

type fileLineHandler func(line int, text string)

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

func ForEachLineInFile(filename string, sp fileLineHandler) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		sp(lineNo, scanner.Text())
	}
}

func GetLineCount(filename string) int {
	lineCount := 0
	increaseLineCount := func(line int, text string) {
		lineCount++
	}
	ForEachLineInFile(filename, increaseLineCount)
	return lineCount
}
