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

func main() {
	profileFor := make(LineMetricForFiles)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		populateProfile(profileFor, scanner.Text())
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

func generateMetricFiles(profileFor LineMetricForFiles) {
	lastLine := 0
	lineMetricGenerator := func(file io.Writer, metrics []LineMetric) func(int, string) {
		return func(line int, text string) {
			metric := metrics[line-1]
			fmt.Fprintf(file, "%34v %v\n", metric, text)
			lastLine = line
		}
	}

	for filename, lineMetrics := range profileFor {
		reportFilename := reportDir + "/" + filename
		createDir(path.Dir(reportFilename))
		file, err := os.Create(reportFilename)
		if err != nil {
			log.Fatal(reportFilename, ":", err)
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

func populateProfile(profileFor LineMetricForFiles, record string) {
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
}
