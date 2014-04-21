package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
)

import "fprof/osutil"
import "fprof/report"
import "fprof/report/html"
import "fprof/json"

var reportDir = "fprof"
var runBrowser = true
var browser = "google-chrome"
var jsonfile = "-"

func main() {
	var pNoBrowser = flag.Bool("w", false, "Do not start the browser")
	var pBrowser = flag.String("b", browser, "Use the given browser to open the profiling results")

	var pReportDir = flag.String("o", reportDir, "Directory to generate profile reports")
	flag.Parse()

	if *pNoBrowser {
		runBrowser = false
	}
	if *pReportDir != reportDir {
		reportDir = *pReportDir
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
	err := osutil.RunCommand(browser, htmlfile)
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
	profile := json.From(in)

	var reporter report.Reporter
	reporter = html.New(reportDir)
	reporter.ReportFunctions(profile)
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
		osutil.CreateDir(path.Dir(profileFilename))
		file, err := os.Create(profileFilename)
		if err != nil {
			log.Fatal(profileFilename, ":", err)
		}
		defer file.Close()

		printer := lineMetricGenerator(file, lineMetrics)
		osutil.ForEachLineInFile(filename, printer)
		printer(lastLine+1, "")
	}
}
