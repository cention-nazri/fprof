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
import "fprof/jsonprofile"

var reportDir = "fprof"
var reportType = "txt"
var runBrowser = true
var browser = "google-chrome"
var jsonfile = "-"

func main() {
	var pNoBrowser = flag.Bool("w", false, "Do not start the browser")
	var pBrowser = flag.String("b", browser, "Use the given browser to open the profiling results")

	var pReportDir = flag.String("o", reportDir, "Directory to generate profile reports")
	var pReportType = flag.String("t", reportType, "Report type txt or html (default)")
	flag.Parse()

	if *pNoBrowser {
		runBrowser = false
	}
	if *pReportDir != reportDir {
		reportDir = *pReportDir
	}
	if *pReportType != reportType {
		reportType = *pReportType
	}
	if *pBrowser != browser {
		browser = *pBrowser
	}

	args := flag.Args()
	if len(args) == 1 {
		jsonfile = args[0]
		if jsonfile != "-" && reportDir == "fprof" {
			reportDir = jsonfile + ".d"
		}
	}

	reportFromJson()

	if runBrowser {
		openInBrowser(reportDir + "/functions.html")
	}

	//reportFromTxt()
}

func openInBrowser(htmlfile string) {
	log.Println(browser, htmlfile)
	err := helper.RunCommand(browser, htmlfile)
	if err != nil {
		log.Fatal(browser, ":", err)
	}
}

func reportFromJson() {
	in := os.Stdin
	if jsonfile != "-" {
		var err error
		in, err = os.Open(jsonfile)
		if err != nil {
			log.Fatal(err)
		}
	}
	profile := jsonprofile.From(in)

	var reporter report.Reporter
	reporter = html.New(reportDir)
	reporter.ReportFunctions(profile)
}

func reportFromTxt() {
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
