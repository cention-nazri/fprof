package text

import (
	"fmt"
	"log"
)

import "fprof/report"
import "fprof/helper"
import "fprof/jsonprofile"

type TextReporter struct {
	report.Report
}

func New(reportDir string) *TextReporter {
	helper.CreateDir(reportDir)
	reporter := TextReporter{}
	reporter.ReportDir = reportDir
	reporter.ProfileFile = helper.CreateFile(reportDir + "/profile.txt")
	return &reporter
}

func (reporter *TextReporter) Prolog(header string) {
	fmt.Fprintln(reporter.ProfileFile, header)
}

func (reporter *TextReporter) PrintMetrics(filesDir string, timings report.LineMetric, filenameAndLine string) {
	fmt.Fprintf(reporter.ProfileFile, "%v%v%v\n", timings, filesDir, filenameAndLine)
}

func (reporter *TextReporter) PopulateProfile(profileFor report.LineMetricForFiles, record string) {
	timings, filenameAndLine := report.GetTimingsAndFilenameLineInfo(record)
	filename, line := report.GetFilenameAndLineNumber(filenameAndLine)

	lineMetrics, exists := profileFor[filename]
	if exists {
		if cap(lineMetrics) < line {
			log.Fatal(line, " is more than line count for file", filename, cap(lineMetrics))
		}
	} else {
		lineCount := helper.GetLineCount(filename)
		lineMetrics = make([]report.LineMetric, lineCount+1)
		profileFor[filename] = lineMetrics
	}
	//fmt.Println("line count for",filename,"is", cap(lineMetrics))
	//fmt.Println("line is", line)
	profileFor[filename][line-1] = timings
	reporter.PrintMetrics(report.FilesDir, timings, filenameAndLine)
}

func (reporter *TextReporter) Epilog() {
	reporter.generateCtags()
}

func (reporter *TextReporter) ReportFunctions(profile *jsonprofile.Profile) {
}

func (reporter *TextReporter) generateCtags() {
	err := helper.RunCommand(
		"ctags",
		"-R",
		"--tag-relative",
		"-o",
		reporter.ReportDir+"/tags",
		reporter.ReportDir+"/"+report.FilesDir)
	if err != nil {
		log.Printf("ctags ran with error: %v", err)
	}
}
