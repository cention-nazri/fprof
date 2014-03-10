package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
)

import "fprof/helper"
import "fprof/report"
import "fprof/report/text"

var reportDir = "fproftxt"

func main() {
	var pReportDir = flag.String("o", reportDir, "Directory to generate profile reports")
	flag.Parse()

	if *pReportDir != reportDir {
		reportDir = *pReportDir
	}

	profileFor := make(report.LineMetricForFiles)

	scanner := bufio.NewScanner(os.Stdin)
	header := ""
	reporter := text.New(reportDir)

	if scanner.Scan() {
		header = scanner.Text()
		reporter.PrintHeader(header)
	}
	for scanner.Scan() {
		reporter.PopulateProfile(profileFor, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("reading standard input:", err)
	}
	generateMetricFiles(profileFor)
}

func generateMetricFiles(profileFor report.LineMetricForFiles) {
	lastLine := 0
	lineMetricGenerator := func(file io.Writer, metrics []report.LineMetric) func(int, string) {
		return func(line int, text string) {
			metric := metrics[line-1]
			fmt.Fprintf(file, "%56v %v\n", metric, text)
			lastLine = line
		}
	}

	filePrefix := reportDir + "/" + report.FilesDir
	for filename, lineMetrics := range profileFor {
		profileFilename := filePrefix + filename
		helper.CreateDir(path.Dir(profileFilename))
		file, err := os.Create(profileFilename)
		if err != nil {
			log.Fatal(profileFilename, ":", err)
		}
		defer file.Close()

		printer := lineMetricGenerator(file, lineMetrics)
		helper.ForEachLineInFile(filename, printer)
		printer(lastLine+1, "")
	}
}
