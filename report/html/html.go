package html

import (
	"fmt"
	"log"
	"io"
	"os"
	"bufio"
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
	//hw.spaces()
	fmt.Fprintf(hw.w, "%v\n", v)
}

func getFirstWhiteSpaces(str string) string {

	for i, v := range(str) {
		if v != ' ' && v != '\t' {
			return str[0:i]
		}
	}
	return ""
}

func (hw *HtmlWriter) commentln(indent, format string, args ...interface{}) {
	comment := fmt.Sprintf(format, args...)
	fmt.Fprintf(hw.w, "%s// %s\n", indent, comment)
}

func (hw *HtmlWriter) begin(el string, attrs ...string) {
	fmt.Fprintln(hw.w, "")
	//hw.writeln("")
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

func (hw *HtmlWriter) Html(v ...interface{}) {
	for _, e := range(v) {
		hw.write(e)
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

func (hw *HtmlWriter) HtmlOpen() { hw.begin("html") }
func (hw *HtmlWriter) HtmlClose() { hw.end("html") }
func (hw *HtmlWriter) HeadOpen() { hw.begin("head") }
func (hw *HtmlWriter) LinkCss(cssFile string) { hw.begin("link", `rel="stylesheet"`, `type="text/css"`, fmt.Sprintf(`href="%s"`, cssFile)) }
func (hw *HtmlWriter) HeadClose() { hw.end("head") }
func (hw *HtmlWriter) BodyOpen() { hw.begin("body") }
func (hw *HtmlWriter) BodyClose() { hw.end("body") }
func (hw *HtmlWriter) TableOpen(attrs ...string) { hw.begin("table", attrs...) }
func (hw *HtmlWriter) TableClose() { hw.end("table") }
func (hw *HtmlWriter) TrOpen() { hw.begin("tr") }
func (hw *HtmlWriter) TrClose() { hw.end("tr") }
func (hw *HtmlWriter) ThOpen() { hw.begin("th") }
func (hw *HtmlWriter) ThClose() { hw.end("th") }
func (hw *HtmlWriter) TdOpen(attrs ...string) { hw.begin("td", attrs...) }
func (hw *HtmlWriter) TdClose() { hw.end("td") }
func (hw *HtmlWriter) DivOpen() { hw.begin("div") }
func (hw *HtmlWriter) DivClose() { hw.end("div") }
func (hw *HtmlWriter) Th(v ...interface{}) { hw.repeatIn("th", v...) }
func (hw *HtmlWriter) Td(v ...interface{}) { hw.repeatIn("td", v...) }
func (hw *HtmlWriter) Tr(v ...interface{}) { hw.repeatIn("tr", v...) }
func (hw *HtmlWriter) Div(v ...interface{}) { hw.repeatIn("div", v...) }

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

func (reporter *HtmlReporter) htmlLink(name, sourceFile string, lineNo jsonprofile.Counter) string {
	href := fmt.Sprintf("%s#%d", reporter.htmlLineFilename(sourceFile), lineNo)
	return `<a href="` + href + `">` + name + `</a>`
}

func pathToRoot(file string) string { 
	up := ""
	for d := path.Dir(file); len(d) > 1; d = path.Dir(d) {
		up += "../"
	}
	return up
}

func (reporter *HtmlReporter) writeOneTableRow(hw *HtmlWriter, lineNo int, lp *jsonprofile.LineProfile, scanner *bufio.Scanner) {
	hasSourceLine := false
	sourceLine := ""
	indent := ""
	hw.TrOpen()
	hw.TdOpen()
	hw.writeln(fmt.Sprintf(`<a id="%d"></a>`, lineNo))
	hw.writeln(lineNo)
	hw.TdClose()
	if scanner.Scan() {
		hasSourceLine = true
		sourceLine = scanner.Text()
		indent = getFirstWhiteSpaces(sourceLine)
	}
	if lp == nil {
		hw.Td("","","", "")
		hw.TdOpen(`class="s"`)
	} else {
		hw.Td(lp.Hits, lp.TotalDuration.InMillisecondsStr())
		hw.Td(lp.CallsMade.EmptyIfZero(), lp.TimeInFunctions.NonZeroMsOrNone())
		hw.TdOpen(`class="s"`)
		f := lp.Function
		if f != nil  {
			nCallers := len(f.Callers)
			if nCallers == 0 {
				hw.commentln(indent, "Spent %vms within %v()",
				f.InclusiveDuration.InMillisecondsStr(),
				f.Name)
			} else {

				hw.commentln(indent, "Spent %vms within %v() which was called:",
				f.InclusiveDuration.InMillisecondsStr(),
				f.Name)
				for _, c := range(f.Callers) {
					hw.commentln(indent, "%d time(s) (%vms) by %s at line %d, avg %.3fms/call",
					c.Frequency, c.TotalDuration.InMillisecondsStr(),
					reporter.htmlLink(c.Name + "()", c.Filename, c.At),
					
					c.At, c.TotalDuration.AverageInMilliseconds(c.Frequency),
				)
			}
		}
		}
	}
	if hasSourceLine {
		hw.writeln(sourceLine)
	}
	hw.TdClose()
	hw.TrClose()
}
func (reporter *HtmlReporter) writeOneHtmlFile(file string, lineProfiles []*jsonprofile.LineProfile) {
	htmlfile := reporter.ReportDir +"/"+ reporter.htmlLineFilename(file)
	helper.CreateDir(path.Dir(htmlfile))
	hw := NewHtmlWriter(helper.CreateFile(htmlfile))
	hw.HtmlWithCssBodyOpen(pathToRoot(file) + "../style.css")
	hw.Html(file)
	hw.TableOpen(`border="1"`, `cellpadding="0"`)
	hw.Th("Line", "Hits", "Time on line (ms)", "Calls Made", "Time in functions", "Code")

	sourceFile, err := os.Open(file)
	if err != nil {
		log.Println("Error reading %v:%v", file, err)
		return
	}
	scanner := bufio.NewScanner(sourceFile)

	for lineNo, lp := range lineProfiles {
		if lineNo == 0 { continue; }
		reporter.writeOneTableRow(hw, lineNo, lp, scanner)
	}
	hw.TableClose()
	for i := 0; i < 50; i++ {
		hw.writeln("<br>")
	}
	hw.BodyClose()
	hw.HtmlClose()
}

func (reporter *HtmlReporter) GenerateCssFile() {
	css := helper.CreateFile(reporter.ReportDir + "/style.css")
	fmt.Fprint(css,
`
body {
	font-family: sans-serif;
}
table {
	border-spacing: 0;
	border-collapse: collapse;
	border-color: gray;
}
tr {
	vertical-align: top;
}
th {
	text-align: center;
}
th, td {
	padding: 0 .4em
}
td {
	vertical-align: inherit;
	text-align: right;
}
td.s {
	text-align: left;
	font-family: monospace;
	white-space: pre;
}
`)
}

func (reporter *HtmlReporter) GenerateHtmlFiles(fileProfiles jsonprofile.FileProfile) {
	//helper.CreateFile(reporter.ReportDir +"/"+ report.FilesDir)
	done := make(chan bool)
	for file, lineProfiles := range fileProfiles {
		go func (file string, lineProfiles []*jsonprofile.LineProfile) {
			reporter.writeOneHtmlFile(file, lineProfiles)
			// fmt.Println(file)
			done <- true
		}(file, lineProfiles)
	}

	for i := 0; i < len(fileProfiles); i++ {
		<-done
	}
}

func (hw *HtmlWriter) HtmlWithCssBodyOpen(cssFile string) {
	hw.HtmlOpen()
	hw.HeadOpen()
	hw.LinkCss(cssFile)
	hw.HeadClose()
	hw.BodyOpen()
}

func (reporter *HtmlReporter) ReportFunctions(fileProfiles jsonprofile.FileProfile) {
	reporter.GenerateCssFile()
	functionCalls := fileProfiles.GetFunctionsSortedByExlusiveTime()
	reporter.GenerateHtmlFiles(fileProfiles)
	html := helper.CreateFile(reporter.ReportDir + "/functions.html")
	hw := NewHtmlWriter(html)

	hw.HtmlWithCssBodyOpen("style.css")
	hw.Html("Functions sorted by exclusive time")
	hw.TableOpen(`border="1"`, `cellpadding="0"`)
	hw.Th("Calls", "Places", "Files", "Exclusive", "Inclusive", "Function")
	for _, fc := range(functionCalls) {
		hw.TrOpen()
		hw.Td(fc.Hits,
			fc.CountCallingPlaces(),
			fc.CountCallingFiles(),
			fc.ExclusiveDuration.NonZeroMsOrNone(),
			fc.InclusiveDuration.NonZeroMsOrNone())
		hw.TdOpen(`class="s"`)
		hw.write(reporter.htmlLink(fc.Name, fc.Filename, fc.StartLine))
		hw.TdClose()
		//fmt.Println(fc.ExclusiveDuration.InMilliseconds(), fc.Name)
		hw.TrClose()
		//fmt.Println(lines)
	}
	hw.TableClose()
	hw.BodyClose()
	hw.HtmlClose()
}

func (reporter *HtmlReporter) Epilog() {
	fmt.Fprintln(reporter.ProfileFile, "</table>");
}
