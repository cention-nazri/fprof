package html

import (
	"fmt"
	"log"
	"io"
	"strings"
	"path"
)

import "fprof/report"
import "fprof/helper"
import "fprof/jsonprofile"

var indent = 0

type HtmlReporter struct {
	report.Report
}

type HtmlWriter struct {
	w io.Writer
	indent int
}

func NewHtmlWriter(w io.Writer) *HtmlWriter {
	hw := HtmlWriter{w, 0}
	return &hw
}

func (hw *HtmlWriter) spaces() {
	str := ""
	for i := 0; i < hw.indent; i++ {
		str += " "
	}
	fmt.Fprint(hw.w, str)
}

func (hw *HtmlWriter) write(v interface{}) {
	fmt.Fprintf(hw.w, "%v", v)
}

func (hw *HtmlWriter) writeln(v interface{}) {
	hw.spaces()
	fmt.Fprintf(hw.w, "%v\n", v)
}

func (hw *HtmlWriter) begin(el string, attrs ...string) {
	hw.spaces()
	hw.write("<" + el)
	for _, v := range(attrs) {
		hw.write(" "+v)
	}
	hw.write(">")
	hw.indent++
}

func (hw *HtmlWriter) end(el string) {
	hw.indent--
	hw.writeln("</" + el + ">")
}

func (hw *HtmlWriter) HtmlOpen() {
	hw.writeln("<html>\n<body>")
	indent++
}

func (hw *HtmlWriter) HtmlClose() {
	hw.writeln("</body>\n</html>")
	indent--
}

func (hw *HtmlWriter) Html(v ...interface{}) {
	for _, e := range(v) {
		hw.writeln(e)
	}
}

func (hw *HtmlWriter) in(el string, v interface{}) {
	hw.begin(el)
	hw.Html(v)
	hw.end(el)
}

func (hw *HtmlWriter) repeatIn(el string, items ...interface{}) {
	for _, v := range(items) {
		hw.in(el, v)
	}
}

func (hw *HtmlWriter) TableOpen(attrs ...string) { hw.begin("table", attrs...) }
func (hw *HtmlWriter) TableClose() { hw.end("table") }
func (hw *HtmlWriter) TrOpen() { hw.begin("tr") }
func (hw *HtmlWriter) TrClose() { hw.end("tr") }
func (hw *HtmlWriter) ThOpen() { hw.begin("th") }
func (hw *HtmlWriter) ThClose() { hw.end("th") }
func (hw *HtmlWriter) TdOpen() { hw.begin("td") }
func (hw *HtmlWriter) htmlTdClose() { hw.end("td") }
func (hw *HtmlWriter) DivOpen() { hw.begin("div") }
func (hw *HtmlWriter) DivClose() { hw.end("div") }
func (hw *HtmlWriter) Th(v ...interface{}) { hw.repeatIn("th", v...) }
func (hw *HtmlWriter) Td(v ...interface{}) { hw.repeatIn("td", v...) }
func (hw *HtmlWriter) Tr(v ...interface{}) { hw.repeatIn("tr", v...) }

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

func (reporter *HtmlReporter) htmlLineFilename(file string) string {
	return report.FilesDir+file+"-line.html"
}

func (reporter *HtmlReporter) htmlLink(fc *jsonprofile.FunctionProfile) string {
	href := fmt.Sprintf("%s#%d", reporter.htmlLineFilename(fc.Filename), fc.StartLine)
	return `<a href="` + href + `">` + fc.Name + `</a>`
}

func (reporter *HtmlReporter) GenerateHtmlFiles(fileProfiles jsonprofile.FileProfile) {

	dir := reporter.ReportDir
	//helper.CreateFile(reporter.ReportDir +"/"+ report.FilesDir)
	for file, lineProfiles := range fileProfiles {
		file := dir +"/"+ reporter.htmlLineFilename(file)
		helper.CreateDir(path.Dir(file))
		hw := NewHtmlWriter(helper.CreateFile(file))
		hw.HtmlOpen()
		hw.TableOpen(`border="0"`, `cellpadding="0"`)
		for lineNo, lineProfile := range lineProfiles {
			hw.TrOpen()
			hw.Td(lineNo)
			if lineProfile == nil || lineProfile.Function == nil {
				continue
			}
			lineProfile.Function.Filename = file
			lineProfile.Function.StartLine = uint64(lineNo)
			hw.TrClose()
		}
		hw.TableClose()
		hw.HtmlClose()
	}
}

func (reporter *HtmlReporter) ReportFunctions(fileProfiles jsonprofile.FileProfile) {
	reporter.GenerateHtmlFiles(fileProfiles)
	functionCalls := fileProfiles.GetFunctionsSortedByExlusiveTime()
	html := helper.CreateFile(reporter.ReportDir + "/functions.html")
	hw := NewHtmlWriter(html)

	hw.HtmlOpen()
	hw.Html("Functions")
	hw.TableOpen(`border="0"`, `cellpadding="0"`)
	hw.Th("Calls", "Places", "Files", "Exclusive", "Inclusive")
	for _, fc := range(functionCalls) {
		hw.TrOpen()
		hw.Td(fc.Hits, "TODO", "TODO",
			fc.ExclusiveDuration.NonZeroMsOrNone(),
			fc.InclusiveDuration.NonZeroMsOrNone(),
			reporter.htmlLink(fc))
		//fmt.Println(fc.ExclusiveDuration.InMilliseconds(), fc.Name)
		hw.TrClose()
		//fmt.Println(lines)
	}
	hw.TableClose()
	hw.HtmlClose()
}

func (reporter *HtmlReporter) Epilog() {
	fmt.Fprintln(reporter.ProfileFile, "</table>");
}
