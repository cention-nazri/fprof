package text

import (
	"fmt"
)

import "fprof/report"
import "fprof/helper"

type TextReporter struct {
	*report.Reporter
}

func New(reportDir string) *TextReporter {
	helper.CreateDir(reportDir)
	reporter := &TextReporter{ }
	reporter.ReportDir = reportDir
	reporter.ProfileFile = helper.CreateFile(reportDir + "/profile.txt")
	return reporter
}

func (reporter *TextReporter) PrintHeader(header string) {
	fmt.Fprintln(reporter.ProfileFile, header)
}
