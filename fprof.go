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
import "fprof/report/html"

var reportDir = "fprof"
var reportType = "txt"

func main() {
	var pReportDir = flag.String("o", reportDir, "Directory to generate profile reports")
	var pReportType = flag.String("t", reportType, "Report type txt or html (default)")
	flag.Parse()

	if *pReportDir != reportDir {
		reportDir = *pReportDir
	}
	if *pReportType != reportType {
		reportType = *pReportType
	}

	profileFor := make(report.LineMetricForFiles)

	scanner := bufio.NewScanner(os.Stdin)
	header := ""
	var reporter report.Reporter
	if reportType == "txt" {
		reporter = text.New(reportDir)
	} else {
		reporter = html.New(reportDir)
	}

	if scanner.Scan() {
		header = scanner.Text()
		reporter.Prolog(header)
	}
	for scanner.Scan() {
		reporter.PopulateProfile(profileFor, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("reading standard input:", err)
	}
	generateMetricFiles(profileFor)
	reporter.Epilog()
}

func generateMetricFiles(profileFor report.LineMetricForFiles) {
	lastLine := 0
	lineMetricGenerator := func(file io.Writer, metrics []report.LineMetric) func(int, string) {
		return func(line int, text string) {
			metric := metrics[line-1]
			/* The metric width must match the width set by
			 * ferite_profile.c write_profile_line_entry()
			 */
			fmt.Fprintf(file, "%62v %v\n", metric, text)
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
