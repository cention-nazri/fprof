package html

import (
	"fmt"
	"log"
	"strings"
)

import "fprof/report"
import "fprof/helper"

type HtmlReporter struct {
	report.Report
}

func New(reportDir string) *HtmlReporter {
	helper.CreateDir(reportDir)
	reporter := HtmlReporter{}
	reporter.ReportDir = reportDir
	reporter.ProfileFile = helper.CreateFile(reportDir + "/profile.html")
	return &reporter
}

func (reporter *HtmlReporter) Prolog(header string) {
	fmt.Fprint(reporter.ProfileFile, "<table><tr>")
	for _, head := range(strings.Fields(header)) {
		fmt.Fprintf(reporter.ProfileFile, "<th>%v</th>", head)
	}
	fmt.Fprintln(reporter.ProfileFile, "<th></th></tr>");
}

func (reporter *HtmlReporter) PrintMetrics(filesDir string, timings report.LineMetric, filenameAndLine string) {
	fmt.Fprint(reporter.ProfileFile, "<tr>");
	nPrinted := 0
	for _, metric := range(strings.Fields(string(timings))) {
		fmt.Fprintf(reporter.ProfileFile, "<td>%v</td>", metric)
		nPrinted++
	}
	for i := nPrinted; i < 5; i++ {
		fmt.Fprint(reporter.ProfileFile, "<td></td>");
	}
	fmt.Fprintf(reporter.ProfileFile, "<td>%v%v</td></tr>\n", filesDir, filenameAndLine)
}

func (reporter *HtmlReporter) PopulateProfile(profileFor report.LineMetricForFiles, record string) {
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

func (reporter *HtmlReporter) Epilog() {
	fmt.Fprintln(reporter.ProfileFile, "</table>");
}
