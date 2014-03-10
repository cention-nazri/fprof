package report

import (
	"io"
)

type Reporter struct {
	ReportDir string
	ProfileFile io.Writer
}
