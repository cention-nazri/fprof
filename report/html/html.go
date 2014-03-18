package html

import (
	"fmt"
	"log"
	"io"
	"os"
	"bufio"
	"strings"
	"path"
	"html"
	_ "sort"
)

import "fprof/report"
import "fprof/helper"
import "fprof/jsonprofile"

var indent = 0

type HtmlReporter struct {
	report.Report
}

type HtmlWriter struct {
	SourceFile string
	HtmlFilename string
	w io.Writer
	indent int
}

func NewHtmlWriter(sourceFile, htmlfile string) *HtmlWriter {
	hw := HtmlWriter{sourceFile, htmlfile, helper.CreateFile(htmlfile), 0}
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
	hw.comment(indent, format + "\n", args...)
}

func (hw *HtmlWriter) comment(indent, format string, args ...interface{}) {
	comment := fmt.Sprintf(format, args...)
	fmt.Fprintf(hw.w, `<div class="profile_note">%s// %s</div>`, indent, comment)
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
	//reporter.ProfileFile = helper.CreateFile(reportDir + "/profile.html")
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

func htmlLink(fromFile, funcName, toFile string, lineNo jsonprofile.Counter) string {
	return fmt.Sprintf(`<a href="%s#%d">%s</a>`,  getRelativePathTo(toFile, fromFile), lineNo, funcName)
}

func pathToRoot(file string) string {
	if file[0] != '/' {
		file = "/" + file
	}
	up := ""
	for d := path.Dir(file); len(d) > 1; d = path.Dir(d) {
		up += "../"
	}
	return up
}

func explodePath(path string) []string {
	return strings.FieldsFunc(path, func(ch rune) bool {
		return ch == '/'
	})
}

func stripCommonPath(file, from string) (string, string) {
	branchPoint := -1
	i := 0
	for i < len(file) && i < len(from) && file[i] == from[i] {
		if file[i] == '/' {
			branchPoint = i
		}
		i++
	}
	if branchPoint > 0 {
		file = file[branchPoint+1:]
		from = from[branchPoint+1:]
	}
	return file, from
}

func getRelativePathTo(to, from string) string {
	to, from = path.Clean(to), path.Clean(from)
	if to == from {
		return ""
	}
	if from == "." {
		return to
	}

	//log.Printf("getRelativePathTo(%s, %s)\n", to, from)
	rto, rfrom := stripCommonPath(to, from)
	r := ""
	if path.Dir(rto) != "." {
		r = "../" + pathToRoot(path.Dir(rfrom)) + rto
	} else {
		r = pathToRoot(path.Dir(rfrom)) + rto
	}
	//log.Printf(" => %s\n", r)
	return r
}

func (reporter *HtmlReporter) showCallers(hw *HtmlWriter, fp *jsonprofile.FunctionProfile, indent string) {
	nCallers := len(fp.Callers)
	if nCallers == 0 {
		hw.comment(indent, "Spent %vms within %v()", fp.InclusiveDuration.InMillisecondsStr(), fp.FullName())
	} else {
		hw.commentln(indent, "Spent %vms within %v() which was called:", fp.InclusiveDuration.InMillisecondsStr(), fp.FullName())
		calleeFile := reporter.htmlLineFilename(fp.Filename)
		for _, c := range(fp.Callers) {
			callerFile := reporter.htmlLineFilename(c.Filename)
			hw.commentln(indent, "%d time(s) (%vms) by %s() at %s, avg %.3fms/call",
				c.Frequency, c.TotalDuration.InMillisecondsStr(),
				c.FullName(),
				htmlLink(calleeFile, fmt.Sprintf("line %d", c.At), callerFile, c.At),
				c.TotalDuration.AverageInMilliseconds(c.Frequency))
		}
	}
}

func (reporter *HtmlReporter) showCallsMade(hw *HtmlWriter, lp *jsonprofile.LineProfile, indent string) {
	/* Time spent calling functions */
	/* FIXME populate function call metric from lp.Function.Callers */
	if lp != nil && len(lp.FunctionCalls) > 0 {
		for _, c := range(lp.FunctionCalls) {
			callTxt := "call" // i18n unfriendly
			if c.CallsMade > 1 { callTxt = "calls" }
			hw.commentln(indent, "Spent %vms making %d %s to %s(), avg %.3fms/call",
				c.TimeInFunctions.InMillisecondsStr(),
				c.CallsMade,
				callTxt,
				htmlLink(reporter.htmlLineFilename(hw.SourceFile), c.To.FullName(), reporter.htmlLineFilename(c.To.Filename), c.To.StartLine-1),
				c.TimeInFunctions.AverageInMilliseconds(c.CallsMade))
		}
	}
}

func (reporter *HtmlReporter) writeOneTableRow(hw *HtmlWriter, lineNo int, lp *jsonprofile.LineProfile, fp *jsonprofile.FunctionProfile, scanner *bufio.Scanner) {
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
	}

	/* Function definition */
	if fp != nil  {
		reporter.showCallers(hw, fp, indent)
	}
	if hasSourceLine {
		hw.writeln(html.EscapeString(sourceLine))
		reporter.showCallsMade(hw, lp, indent)
	}
	hw.TdClose()
	hw.TrClose()
}

func makeEmptyLineProfiles(file string) []*jsonprofile.LineProfile {
	return make([]*jsonprofile.LineProfile, helper.GetLineCount(file))
}

func fileExists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func (reporter *HtmlReporter) writeOneHtmlFile(file string, fileProfiles jsonprofile.FileProfile) {
	htmlfile := reporter.ReportDir +"/"+ reporter.htmlLineFilename(file)
	helper.CreateDir(path.Dir(htmlfile))
	hw := NewHtmlWriter(file, htmlfile)
	hw.HtmlWithCssBodyOpen(pathToRoot(file) + "../style.css")
	hw.Html(file)
	hw.TableOpen(`border="1"`, `cellpadding="0"`)
	hw.Th("Line", "Hits", "Time on line (ms)", "Calls Made", "Time in functions", "Code")

	if ! fileExists(file) {
		log.Printf("Skipped (file does not exist): %s\n", file);
		return;
	}
	sourceFile, err := os.Open(file)
	if err != nil {
		log.Printf("Error reading %v:%v\n", file, err)
		return
	}
	scanner := bufio.NewScanner(sourceFile)
	lineProfiles := fileProfiles[file]
	if lineProfiles == nil {
		lineProfiles = makeEmptyLineProfiles(file)
	}

	var fp *jsonprofile.FunctionProfile
	for i, lp := range lineProfiles {
		lineNo := i + 1
		fp = nil
		if i + 1 < len(lineProfiles) {
			if lineProfiles[i+1] != nil {
				fp = lineProfiles[i+1].Function
				//log.Println("function:",fp)
				//log.Println("ncallers", len(fp.Callers))
			}
		}
		reporter.writeOneTableRow(hw, lineNo, lp, fp, scanner)
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
.profile_note {
	color: gray;
}
.profile_note:hover {
	color: black;
	background-color: gray;
}
`)
}

func (reporter *HtmlReporter) GenerateHtmlFiles(fileProfiles jsonprofile.FileProfile) {
	//helper.CreateFile(reporter.ReportDir +"/"+ report.FilesDir)

	fileMap := make(map[string]bool)
	done := make(chan bool)
	for file, lineProfiles := range fileProfiles {
		fileMap[file] = true
		for _, v := range(lineProfiles) {
			if v == nil { continue }
			if v.Function == nil { continue }
			fp := v.Function
			if len(fp.Filename) == 0 {
				log.Println("Got empty filename from func profile")
			} else {
				fileMap[fp.Filename] = true
			}
			for _, caller := range(fp.Callers) {
				if caller == nil { continue }
				if len(caller.Filename) == 0 {
					log.Println("Got empty filename from caller profile")
				} else {
					fileMap[caller.Filename] = true
				}
			}
		}
	}
	log.Println("Generating source html files")
	for filename, _ := range(fileMap) {
		go func (file string, fileProfile jsonprofile.FileProfile) {
			reporter.writeOneHtmlFile(file, fileProfiles)
			// fmt.Println(file)
			done <- true
		}(filename, fileProfiles)
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
	log.Println("Cross referencing function call metrics...")
	functionCalls := fileProfiles.GetFunctionsSortedByExlusiveTime()
	reporter.GenerateHtmlFiles(fileProfiles)
	log.Println("Generating functions html report")
	hw := NewHtmlWriter("", reporter.ReportDir + "/functions.html")

	hw.HtmlWithCssBodyOpen("style.css")
	hw.Html("Functions sorted by exclusive time")
	hw.TableOpen(`border="1"`, `cellpadding="0"`)
	hw.Th("Calls", "Places", "Files", "Exclusive", "Inclusive", "Function")
	na := "n/a"
	//done := make(chan bool)
	//nthreads := 0
	for _, fc := range(functionCalls) {
		hw.TrOpen()
		if fc == nil {
			hw.Td(na, na, na, na, na)
			hw.TdOpen(`class="s"`)
			hw.write(na)
		} else {
			hw.Td(fc.Hits,
				fc.CountCallingPlaces(),
				fc.CountCallingFiles(),
				fc.ExclusiveDuration.NonZeroMsOrNone(),
				fc.InclusiveDuration.NonZeroMsOrNone())
			hw.TdOpen(`class="s"`)
			//if fileProfiles[fc.Filename] == nil {
			//	go func (file string, lineProfiles []*jsonprofile.LineProfile) {
			//		reporter.writeOneHtmlFile(file, lineProfiles)
			//		done <- true
			//	}(fc.Filename, make([]*jsonprofile.LineProfile, helper.GetLineCount(fc.Filename)))
			//	nthreads++
			//}
			hw.write(htmlLink(".", fc.FullName(), reporter.htmlLineFilename(fc.Filename), fc.StartLine))
		}
		hw.TdClose()
		//fmt.Println(fc.ExclusiveDuration.InMilliseconds(), fc.FullName())
		hw.TrClose()
		//fmt.Println(lines)
	}
	hw.TableClose()
	hw.BodyClose()
	hw.HtmlClose()
	//for i:=0; i<nthreads; i++ {
	//	<-done
	//}
}

func (reporter *HtmlReporter) Epilog() {
	fmt.Fprintln(reporter.ProfileFile, "</table>");
}
