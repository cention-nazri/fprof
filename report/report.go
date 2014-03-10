package report

import (
	"io"
	"log"
	"strings"
	"strconv"
)

const FilesDir = "files"

type LineMetric string
type LineMetricForFiles map[string][]LineMetric

type Report struct {
	ReportDir string
	ProfileFile io.Writer
}

type Reporter interface {
	PrintMetrics(filesDir string, timings LineMetric, filenameAndLine string)
}

func GetFilenameAndLineNumber(filenameAndLine string) (string, int) {
	colon := strings.LastIndex(filenameAndLine, ":")
	filename := filenameAndLine[0:colon]
	line, err := strconv.Atoi(filenameAndLine[colon+1:])
	if err != nil {
		log.Fatal("Expecting line number.", err)
	}
	return filename, line
}

func GetTimingsAndFilenameLineInfo(record string) (LineMetric, string) {
	firstSlash := strings.Index(record, "/")
	if firstSlash == -1 {
		log.Fatal("Error no slash found in profile record ", record)
	}

	return LineMetric(record[0:firstSlash]), record[firstSlash:]
}
