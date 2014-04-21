package osutil

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
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

func RunCommand(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	return cmd.Wait()
}

func CreateFiles(files map[string]string) {
	for filename, content := range files {
		CreateDir(path.Dir(filename))
		file := CreateFile(filename)
		fmt.Fprint(file, content)
	}
}
