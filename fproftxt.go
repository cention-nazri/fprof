package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

const reportDir = "fproftxt"

type LineMetric string
type LineMetricForFiles map[string][]LineMetric
type fileLineHandler func(line int, text string)

var filesDir = reportDir + "/files"

func main() {
	profileFor := make(LineMetricForFiles)

	scanner := bufio.NewScanner(os.Stdin)
	header := ""
	createDir(reportDir)
	profileFile := createFile(reportDir + "/profile.txt")
	defer (profileFile).(*os.File).Close()
	if scanner.Scan() {
		header = scanner.Text()
		fmt.Fprintln(profileFile, header)
	}
	for scanner.Scan() {
		populateProfile(profileFile, profileFor, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("reading standard input:", err)
	}
	generateMetricFiles(profileFor)
}

func createDir(dir string) {
	err := os.MkdirAll(dir, os.ModeDir|os.ModePerm)
	if err != nil {
		log.Fatal(dir, " ", err)
	}
}

func createFile(path string) io.Writer {
	file, err := os.Create(path)
	if err != nil {
		log.Fatal(file, ":", err)
	}
	return file
}

func generateMetricFiles(profileFor LineMetricForFiles) {
	lastLine := 0
	lineMetricGenerator := func(file io.Writer, metrics []LineMetric) func(int, string) {
		return func(line int, text string) {
			metric := metrics[line-1]
			fmt.Fprintf(file, "%56v %v\n", metric, text)
			lastLine = line
		}
	}

	for filename, lineMetrics := range profileFor {
		profileFilename := filesDir + "/" + filename
		createDir(path.Dir(profileFilename))
		file, err := os.Create(profileFilename)
		if err != nil {
			log.Fatal(profileFilename, ":", err)
		}
		defer file.Close()

		printer := lineMetricGenerator(file, lineMetrics)
		forEachLineInFile(filename, printer)
		printer(lastLine+1, "")
	}
}

func forEachLineInFile(filename string, sp fileLineHandler) {
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

func getLineCount(filename string) int {
	lineCount := 0
	increaseLineCount := func(line int, text string) {
		lineCount++
	}
	forEachLineInFile(filename, increaseLineCount)
	return lineCount
}

func getTimingsAndFilenameLineInfo(record string) (LineMetric, string) {
	firstSlash := strings.Index(record, "/")
	if firstSlash == -1 {
		log.Fatal("Error no slash found in profile record ", record)
	}

	return LineMetric(record[0:firstSlash]), record[firstSlash:]
}

func getFilenameAndLineNumber(filenameAndLine string) (string, int) {
	colon := strings.LastIndex(filenameAndLine, ":")
	filename := filenameAndLine[0:colon]
	line, err := strconv.Atoi(filenameAndLine[colon+1:])
	if err != nil {
		log.Fatal("Expecting line number.", err)
	}
	return filename, line
}

func populateProfile(profileFile io.Writer, profileFor LineMetricForFiles, record string) {
	timings, filenameAndLine := getTimingsAndFilenameLineInfo(record)
	filename, line := getFilenameAndLineNumber(filenameAndLine)

	lineMetrics, exists := profileFor[filename]
	if exists {
		if cap(lineMetrics) < line {
			log.Fatal(line, " is more than line count for file", filename, cap(lineMetrics))
		}
	} else {
		lineCount := getLineCount(filename)
		lineMetrics = make([]LineMetric, lineCount+1)
		profileFor[filename] = lineMetrics
	}
	//fmt.Println("line count for",filename,"is", cap(lineMetrics))
	//fmt.Println("line is", line)
	profileFor[filename][line-1] = timings
	fmt.Fprintf(profileFile, "%v %v%v\n", timings, filesDir, filenameAndLine)
}
