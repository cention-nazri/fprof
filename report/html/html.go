package html

import (
	"bufio"
	"bytes"
	"fmt"
	"html"
	"io"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
)

import "fprof/report"
import "fprof/helper"
import "fprof/stats"
import "fprof/jsonprofile"

var SEVERITY_LOW = .5
var SEVERITY_MEDIUM = 1.0
var SEVERITY_HIGH = 2.0

type HtmlReporter struct {
	report.Report
}

type HtmlWriter struct {
	SourceFile   string
	HtmlFilename string
	realw        io.Writer
	indent       int
	w            *bytes.Buffer
}

func NewHtmlWriter(sourceFile, htmlfile string) *HtmlWriter {
	hw := HtmlWriter{sourceFile,
		htmlfile,
		helper.CreateFile(htmlfile),
		0,
		new(bytes.Buffer),
	}
	return &hw
}

func (hw *HtmlWriter) HiderLink(indent string, nHidden int) {
	hw.write(`<div class="toggleHide">`)
	hw.write(indent)
	hw.write(fmt.Sprintf(`<a href="javascript:" onclick="toggleHide();return false;">Show %d more ...</a>`, nHidden))
	hw.write(`</div>`)
	hw.write(`<div class="hide">`)
}

func (hw *HtmlWriter) writeToDiskAsync(done chan bool) {
	work := func() {
		fmt.Fprintf(hw.realw, hw.w.String())
	}
	if done == nil {
		work()
	} else {
		go func() {
			work()
			done <- true
		}()
	}
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

func getFirstWhiteSpaces(str string) string {

	for i, v := range str {
		if v != ' ' && v != '\t' {
			return str[0:i]
		}
	}
	return ""
}

func (hw *HtmlWriter) commentln(indent, format string, args ...interface{}) {
	hw.comment(indent, format+"\n", args...)
}

func (hw *HtmlWriter) comment(indent, format string, args ...interface{}) {
	comment := fmt.Sprintf(format, args...)
	fmt.Fprintf(hw.w, `<div class="profile_note">%s// %s</div>`, indent, comment)
}

func (hw *HtmlWriter) begin(el string, attrs ...string) {
	hw.spaces()
	hw.write("<" + el)
	for _, v := range attrs {
		hw.write(" " + v)
	}
	hw.write(">")
	hw.indent++
}

func (hw *HtmlWriter) beginln(el string, attrs ...string) {
	hw.begin(el, attrs...)
	fmt.Fprintln(hw.w, "")
}

func (hw *HtmlWriter) end(el string) {
	hw.indent--
	hw.writeln("</" + el + ">")
}

func (hw *HtmlWriter) Html(v ...interface{}) {
	for _, e := range v {
		hw.write(e)
	}
}

func (hw *HtmlWriter) in(el string, v interface{}) {
	hw.spaces()
	hw.Html("<", el, ">")
	hw.Html(v)
	hw.Html("</", el, ">\n")
}

func (hw *HtmlWriter) repeatIn(el string, items ...interface{}) {
	for _, v := range items {
		hw.in(el, v)
	}
}

func (hw *HtmlWriter) HtmlOpen()  { hw.beginln("html") }
func (hw *HtmlWriter) HtmlClose() { hw.end("html") }
func (hw *HtmlWriter) HeadOpen()  { hw.beginln("head") }
func (hw *HtmlWriter) LinkJs(jsFile string) {
	hw.spaces()
	hw.Html(fmt.Sprintf(`<script src="%s" type="text/javascript"></script>`+"\n", jsFile))
}
func (hw *HtmlWriter) LinkCss(cssFile string) {
	hw.spaces()
	hw.Html(fmt.Sprintf(`<link rel="stylesheet" type="text/css" href="%s">`+"\n", cssFile))
}
func (hw *HtmlWriter) HeadClose()                { hw.end("head") }
func (hw *HtmlWriter) BodyOpen()                 { hw.beginln("body") }
func (hw *HtmlWriter) BodyClose()                { hw.end("body") }
func (hw *HtmlWriter) TableOpen(attrs ...string) { hw.beginln("table", attrs...) }
func (hw *HtmlWriter) TableClose()               { hw.end("table") }
func (hw *HtmlWriter) TheadOpen()                { hw.beginln("thead") }
func (hw *HtmlWriter) TheadClose()               { hw.end("thead") }
func (hw *HtmlWriter) TbodyOpen()                { hw.beginln("tbody") }
func (hw *HtmlWriter) TbodyClose()               { hw.end("tbody") }
func (hw *HtmlWriter) TrOpen()                   { hw.beginln("tr") }
func (hw *HtmlWriter) TrClose()                  { hw.end("tr") }
func (hw *HtmlWriter) ThOpen()                   { hw.beginln("th") }
func (hw *HtmlWriter) ThClose()                  { hw.end("th") }
func (hw *HtmlWriter) TdOpen(attrs ...string)    { hw.begin("td", attrs...) }
func (hw *HtmlWriter) TdClose()                  { hw.end("td") }
func (hw *HtmlWriter) TdCloseNoIndent() {
	hw.Html("</td>\n")
	hw.indent--
}
func (hw *HtmlWriter) TdWithClassOrEmpty(class string, content string) {
	if len(content) == 0 {
		hw.Td("")
		return
	}

	hw.TdOpen(`class="` + class + `"`)
	hw.write(content)
	hw.TdCloseNoIndent()
}
func (hw *HtmlWriter) DivOpen(attrs ...string) { hw.begin("div", attrs...) }
func (hw *HtmlWriter) DivClose()               { hw.end("div") }
func (hw *HtmlWriter) Th(v ...interface{})     { hw.repeatIn("th", v...) }
func (hw *HtmlWriter) Td(v ...interface{})     { hw.repeatIn("td", v...) }
func (hw *HtmlWriter) Tr(v ...interface{})     { hw.repeatIn("tr", v...) }
func (hw *HtmlWriter) Div(v ...interface{})    { hw.repeatIn("div", v...) }

func New(reportDir string) *HtmlReporter {
	helper.CreateDir(reportDir)
	reporter := HtmlReporter{}
	reporter.ReportDir = reportDir
	return &reporter
}

func (reporter *HtmlReporter) GetPathTo(file string) string {
	return reporter.ReportDir + "/" + file
}
func (reporter *HtmlReporter) Prolog(header string) {
	fmt.Fprint(reporter.ProfileFile, "<table><tr>")
	for _, head := range strings.Fields(header) {
		fmt.Fprintf(reporter.ProfileFile, "<th>%v</th>", head)
	}
	fmt.Fprintln(reporter.ProfileFile, "<th></th></tr>")
}

func (reporter *HtmlReporter) PrintMetrics(filesDir string, timings report.LineMetric, filenameAndLine string) {
	fmt.Fprint(reporter.ProfileFile, "<tr>")
	nPrinted := 0
	for _, metric := range strings.Fields(string(timings)) {
		fmt.Fprintf(reporter.ProfileFile, "<td>%v</td>", metric)
		nPrinted++
	}
	for i := nPrinted; i < 5; i++ {
		fmt.Fprint(reporter.ProfileFile, "<td></td>")
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
	profileFor[filename][line-1] = timings
	reporter.PrintMetrics(report.FilesDir, timings, filenameAndLine)
}

func (reporter *HtmlReporter) htmlLineFilename(file string) string {
	return report.FilesDir + "/" + file + ".html"
}

func isEval(f string) bool {
	b := path.Base(f)
	if len(b) >= 6 {
		if b[0:6] == "eval()" {
			return true
		}
	}
	return false
}

func htmlLink(fromFile, funcName, toFile string, lineNo jsonprofile.Counter) string {
	if isEval(toFile) {
		return `<span title="Called from eval()"><i>` + funcName + `</i></span>`
	}
	return fmt.Sprintf(`<a href="%s#%d">%s</a>`, getRelativePathTo(toFile, fromFile), lineNo, funcName)
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

	rto, rfrom := stripCommonPath(to, from)
	r := ""
	if path.Dir(rto) != "." {
		r = "../" + pathToRoot(path.Dir(rfrom)) + rto
	} else {
		r = pathToRoot(path.Dir(rfrom)) + rto
	}
	return r
}

func (reporter *HtmlReporter) showCallers(hw *HtmlWriter, fp *jsonprofile.FunctionProfile, indent string) {
	hideThreshold := 10
	//nCallers := len(fp.Callers)
	//if nCallers == 0 {
	//	avgStr := ""
	//	freqStr := "once"
	//	if fp.Hits > 1 {
	//		freqStr = fmt.Sprintf("%d times", fp.Hits)
	//		avgStr = fmt.Sprintf(", avg %vms/call", fp.InclusiveDuration.AverageInMilliseconds(fp.Hits))
	//	}
	//	hw.comment(indent, "Spent %vms within %v() which was called %s (caller data not available)%s", fp.InclusiveDuration.InMillisecondsStr(), fp.FullName(), freqStr, avgStr)
	//} else {

	freqStr := ":"
	nCalls := fp.Callers.Total()
	nilCallerStr := ""
	diff := fp.Hits - nCalls
	if diff > 0 {
		if diff == 1 {
			nilCallerStr = fmt.Sprintf("once by an unknown caller")
		} else {
			nilCallerStr = fmt.Sprintf("%d times by unknown callers, avg %.3fms/call", diff, fp.GetTimeSpentByUnknownCallers().AverageInMilliseconds(diff))
		}
	}
	if fp.Hits > 1 {
		freqStr = fmt.Sprintf(" %d times:", fp.Hits)
	}
	hw.comment(indent, "Spent %vms within %v() which was called%s", fp.InclusiveDuration.InMillisecondsStr(), fp.FullName(), freqStr)
	if diff > 0 {
		hw.commentln(indent, "%s", nilCallerStr)
	}
	calleeFile := fp.Filename
	calleeFile = reporter.htmlLineFilename(calleeFile)
	startHideAt := 0
	if len(fp.Callers) > hideThreshold {
		startHideAt = 5
	}
	for i, c := range fp.Callers {
		if startHideAt > 0 {
			if i == startHideAt {
				hw.HiderLink(indent, len(fp.Callers)-startHideAt)
			}
		}
		callerFile, callerAt := c.Filename, c.At
		callerFile = reporter.htmlLineFilename(callerFile)
		freqStr = "once"
		if c.Frequency > 1 {
			freqStr = fmt.Sprintf("%d times", c.Frequency)
		}
		hw.commentln(indent, "%s (%vms) by %s() at %s, avg %.3fms/call",
			freqStr, c.TotalDuration.InMillisecondsStr(),
			c.FullName(),
			htmlLink(calleeFile, fmt.Sprintf("line %d", callerAt), callerFile, callerAt),
			c.TotalDuration.AverageInMilliseconds(c.Frequency))
	}
	if startHideAt > 0 {
		hw.Html(`</div>`)
	}
	//}
}

func (reporter *HtmlReporter) showCallsMade(hw *HtmlWriter, lp *jsonprofile.LineProfile, indent string) {
	/* Time spent calling functions */
	/* FIXME populate function call metric from lp.Function.Callers */
	var callTxt string
	var avgTxt string
	hideThreshold := 10
	startHideAt := 0
	if lp != nil && len(lp.FunctionCalls) > 0 {
		sort.Stable(lp.FunctionCalls)

		if len(lp.FunctionCalls) > hideThreshold {
			startHideAt = 5
		}
		for i, c := range lp.FunctionCalls {
			if startHideAt > 0 {
				if i == startHideAt {
					// TODO check for off by one
					hw.HiderLink(indent, len(lp.FunctionCalls)-startHideAt)
				}
			}
			callTxt = "in" // i18n unfriendly
			avgTxt = ""
			if c.CallsMade > 1 {
				callTxt = fmt.Sprintf("making %d calls to", c.CallsMade)
				avgTxt = fmt.Sprintf(", avg %.3fms/call", c.TimeInFunctions.AverageInMilliseconds(c.CallsMade))
			}

			calleeFQN := c.To.FullName()
			calleeFile, calleeAt := c.To.Filename, c.To.StartLine-1
			calleeFile = reporter.htmlLineFilename(calleeFile)
			hw.commentln(indent, "Spent %vms %s %s()%s",
				c.TimeInFunctions.InMillisecondsStr(),
				callTxt,
				htmlLink(reporter.htmlLineFilename(hw.SourceFile), calleeFQN, calleeFile, calleeAt),
				avgTxt)
		}
		if startHideAt > 0 {
			hw.Html(`</div>`)
		}
	}
}

func (reporter *HtmlReporter) writeOneSourceCodeLine(hw *HtmlWriter, lineNo int, lp *jsonprofile.LineProfile, scanner *bufio.Scanner, ownTimeStats, otherTimeStats *stats.Stats) {
	hasSourceLine := false
	sourceLine := ""
	indent := ""
	hw.TrOpen()
	hw.TdOpen()
	hw.Html(fmt.Sprintf(`<a id="%d">%d</a>`, lineNo, lineNo))
	hw.TdCloseNoIndent()
	if scanner.Scan() {
		hasSourceLine = true
		sourceLine = scanner.Text()
		indent = getFirstWhiteSpaces(sourceLine)
	}

	if lp == nil {
		hw.Td("", "", "", "")
		hw.TdOpen(`class="s"`)
	} else {
		ownTime := lp.TotalDuration
		ownTime.Subtract(lp.TimeInFunctions)
		if lp.Hits > 0 {
			hw.Td(lp.Hits)
		} else {
			hw.Td("")
		}

		hw.TdWithClassOrEmpty(
			getSeverityClass(ownTime.InMilliseconds(), ownTimeStats),
			ownTime.NonZeroMsOrNone(),
		)

		hw.Td(lp.CallsMade.EmptyIfZero())
		hw.TdWithClassOrEmpty(getSeverityClass(lp.TimeInFunctions.InMilliseconds(), otherTimeStats), lp.TimeInFunctions.NonZeroMsOrNone())

		hw.TdOpen(`class="s"`)
		if lp.Functions != nil {
			functions := *lp.Functions
			for _, f := range functions {
				reporter.showCallers(hw, f, indent)
			}
		}
	}

	if hasSourceLine {
		hw.Html(html.EscapeString(sourceLine))
		reporter.showCallsMade(hw, lp, indent)
	}
	hw.TdCloseNoIndent()
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

func (reporter *HtmlReporter) writeOneSourceCodeHtmlFile(file string, fileProfiles jsonprofile.FileProfile, rootJsFiles []string, done chan bool) {
	htmlfile := reporter.ReportDir + "/" + reporter.htmlLineFilename(file)
	helper.CreateDir(path.Dir(htmlfile))
	hw := NewHtmlWriter(file, htmlfile)
	defer hw.writeToDiskAsync(done)

	rootPath := pathToRoot(file)
	jsFiles := []string{}
	for _, file := range rootJsFiles {
		jsFiles = append(jsFiles, rootPath+"../"+file)
	}
	hw.HtmlWithCssBodyOpen(rootPath+"../style.css", jsFiles)
	hw.Html(file)
	hw.TableOpen(`id="function_table"`, `border="1"`, `cellpadding="0"`, `class="sortable"`)
	hw.TheadOpen()
	hw.Th("Line", "Hits", "Time on line (ms)", "Calls Made", "Time in functions", "Code")
	hw.TheadClose()

	if !fileExists(file) {
		log.Printf("FIXME We should not reach here, file %s should exist\n", file)
		return
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

	timesOnLine := make([]float64, 0, len(lineProfiles))
	timesInFunction := make([]float64, 0, len(lineProfiles))
	for _, lp := range lineProfiles {
		if lp == nil {
			continue
		}
		ownTime := lp.TotalDuration
		ownTime.Subtract(lp.TimeInFunctions)
		d := ownTime.InMilliseconds()
		if d > 0 {
			timesOnLine = append(timesOnLine, d)
		}
		d = lp.TimeInFunctions.InMilliseconds()
		if d > 0 {
			timesInFunction = append(timesInFunction, d)
		}
	}
	ownTimeStats := stats.MadMedian(timesOnLine)
	otherTimeStats := stats.MadMedian(timesInFunction)

	hw.TbodyOpen()
	for i, lp := range lineProfiles {
		lineNo := i + 1
		// TODO refactor: don't pass scanner, pass the line
		reporter.writeOneSourceCodeLine(hw, lineNo, lp, scanner, ownTimeStats, otherTimeStats)
	}
	hw.TbodyClose()
	hw.TableClose()
	for i := 0; i < 50; i++ {
		hw.writeln("<br>")
	}
	hw.BodyClose()
	hw.HtmlClose()
}

func (reporter *HtmlReporter) GenerateJsFiles() {
	tableSorterJs := `$.tablesorter.defaults.sortInitialOrder = "desc";`
	fprofJs := `function srcElement(e) {
	e = e || window.event;
	var targ = e.target || e.srcElement;
	if (targ.nodeType == 3) targ = targ.parentNode; // defeat Safari bug
	return targ;
}
function toggleHide(e) {
	var el = srcElement(e);
	var div = el.parentNode;
	var hidden = div.nextSibling;
	if (! hidden.style.display || hidden.style.display === 'none') {
		el.origHTML = el.innerHTML;
		hidden.style.display = 'block';
		el.innerHTML = 'Hide'
	} else {
		hidden.style.display = 'none';
		el.innerHTML = el.origHTML;
	}
}
`
	functionsJs := `$(document).ready(function(){
	$("#functions_table").tablesorter({
		sortList: [[3,1]]
	});
});`
	functionJs := `$(document).ready(function(){
	$("#function_table").tablesorter();
});`

	jsFiles := map[string]string{
		"jquery-min.js":             JQuery,
		"jquery-tablesorter-min.js": JQueryTableSorter,
		"fprof.js":                  fprofJs,
		"tablesorter.js":            tableSorterJs,
		"functions.js":              functionsJs,
		"function.js":               functionJs,
	}

	for filename, content := range jsFiles {
		file := helper.CreateFile(reporter.GetPathTo(filename))
		fmt.Fprint(file, content)
	}
}

func (reporter *HtmlReporter) GenerateCssFile() {
	css := helper.CreateFile(reporter.ReportDir + "/style.css")
	fmt.Fprint(css, `body {
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
.hide {
	display: none;
}
td.s_low {
	background: limegreen;
}
td.s_medium {
	background: darkorange;
}
td.s_high {
	background: lightsalmon;
}
td.s_bad {
	background: salmon;
}

table.sortable thead tr .header {
	background-repeat: no-repeat;
	background-position: 0% 80%;
	cursor: pointer;
}
table.sortable thead tr .headerSortUp   { background-image: url(data:image/png;base64,`)
	fmt.Fprint(css, ImgAscending)

	fmt.Fprint(css, `); }
table.sortable thead tr .headerSortDown { background-image: url(data:image/png;base64,`)
	fmt.Fprint(css, ImgDescending)
	fmt.Fprint(css, `); }
`)
}

func (reporter *HtmlReporter) generateHtmlFilesParallerWorkers(exists map[string]bool, fileProfiles jsonprofile.FileProfile, jsFiles []string, nWorkers int) {
	nFiles := 0
	for _, exist := range exists {
		if exist {
			nFiles++
		}
	}

	type Job struct {
		file string
		// TODO use lineprofiles instead
		fileProfiles jsonprofile.FileProfile
	}

	tasks := make(chan *Job, nFiles)
	defer close(tasks)
	done := make(chan bool, nFiles)
	defer close(done)

	log.Printf("Generating %d source html files\n", nFiles)
	var wg sync.WaitGroup

	for i := 0; i < nWorkers; i++ {
		wg.Add(1)
		go func() {
			for j := range tasks {
				reporter.writeOneSourceCodeHtmlFile(j.file, j.fileProfiles, jsFiles, done)
			}
			wg.Done()
		}()
	}

	for file, exist := range exists {
		if !exist {
			log.Printf("Skipped (file does not exist): %s\n", file)
			continue
		}
		tasks <- &Job{file, fileProfiles}
	}

	for i := 1; i <= nFiles; i++ {
		<-done
		percent := i * 100 / nFiles
		fmt.Printf("%3d%%\r", percent)
	}
	fmt.Println("")
}

func (reporter *HtmlReporter) GenerateSourceCodeHtmlFiles(fileProfiles jsonprofile.FileProfile, jsFiles []string) map[string]bool {
	exists := make(map[string]bool)
	for file, lineProfiles := range fileProfiles {
		exists[file] = fileExists(file)
		for _, v := range lineProfiles {
			if v == nil {
				continue
			}
			if v.Functions == nil {
				continue
			}
			for _, fp := range *v.Functions {
				if len(fp.Filename) == 0 {
					log.Println("Got empty filename from func profile")
				} else {
					exists[fp.Filename] = fileExists(fp.Filename)
				}
				for _, caller := range fp.Callers {
					if caller == nil {
						continue
					}
					if len(caller.Filename) == 0 {
						log.Println("Got empty filename from caller profile")
					} else {
						exists[caller.Filename] = fileExists(caller.Filename)
					}
				}
			}
		}
	}

	reporter.generateHtmlFilesParallerWorkers(exists, fileProfiles, jsFiles, 8)
	return exists
}

func (hw *HtmlWriter) HtmlWithCssBodyOpen(cssFile string, jsFiles []string) {
	hw.HtmlOpen()
	hw.HeadOpen()
	hw.LinkCss(cssFile)
	for _, jsFile := range jsFiles {
		hw.LinkJs(jsFile)
	}
	hw.HeadClose()
	hw.BodyOpen()
}

func getSeverityClass(v float64, stat *stats.Stats) string {
	if stat.MAD == 0 {
		return "s_low"
	}
	d := v - stat.Median
	severity := d / stat.MAD
	if severity < SEVERITY_LOW {
		return "s_low"
	}
	if severity < SEVERITY_MEDIUM {
		return "s_medium"
	}
	if severity < SEVERITY_HIGH {
		return "s_high"
	}
	return "s_bad"
}

func getMADStats(functionCalls jsonprofile.FunctionProfileSlice) (*stats.Stats, *stats.Stats) {
	ownTimes := make([]float64, 0, len(functionCalls))
	incTimes := make([]float64, 0, len(functionCalls))
	for _, fc := range functionCalls {
		if fc == nil {
			continue
		}
		d := fc.OwnTime.InMilliseconds()
		if d > 0 {
			ownTimes = append(ownTimes, d)
		}
		d = fc.InclusiveDuration.InMilliseconds()
		if d > 0 {
			incTimes = append(incTimes, d)
		}
	}
	ownTimeStat := stats.MadMedian(ownTimes)
	incTimeStat := stats.MadMedian(incTimes)

	return ownTimeStat, incTimeStat
}

func (reporter *HtmlReporter) writeOneFunctionMetric(hw *HtmlWriter, fc *jsonprofile.FunctionProfile, exists map[string]bool, ownTimeStat *stats.Stats, incTimeStat *stats.Stats) {
	ieRatio := ""
	inclMS := fc.InclusiveDuration.InMilliseconds()
	exclMS := fc.OwnTime.InMilliseconds()
	if inclMS > 0 {
		ieRatio = fmt.Sprintf("%3.1f", exclMS*100/inclMS)
	}
	hw.TrOpen()
	hw.Td(fc.Hits,
		fc.CountCallingPlaces(),
		fc.CountCallingFiles())

	hw.TdWithClassOrEmpty(
		getSeverityClass(exclMS, ownTimeStat),
		fc.OwnTime.NonZeroMsOrNone(),
	)

	hw.TdWithClassOrEmpty(
		getSeverityClass(inclMS, incTimeStat),
		fc.InclusiveDuration.NonZeroMsOrNone(),
	)

	hw.Td(ieRatio)

	hw.TdOpen(`class="s"`)
	if exists[fc.Filename] {
		hw.write(htmlLink(".", fc.FullName(), reporter.htmlLineFilename(fc.Filename), fc.StartLine))
	} else {
		hw.write(fc.FullName())
	}
	hw.TdCloseNoIndent()
	hw.TrClose()
}

func (reporter *HtmlReporter) GenerateFunctionsHtmlFile(p *jsonprofile.Profile, jsFiles []string, exists map[string]bool, functionCalls jsonprofile.FunctionProfileSlice) {
	hw := NewHtmlWriter("", reporter.ReportDir+"/functions.html")
	defer hw.writeToDiskAsync(nil)

	ownTimeStat, incTimeStat := getMADStats(functionCalls)

	hw.HtmlWithCssBodyOpen("style.css", jsFiles)
	hw.Div("Start: " + p.Start.Time())
	hw.Div("Stop: " + p.Stop.Time())
	hw.Div("Duration: " + p.Duration.InMillisecondsStr() + "ms")
	hw.TableOpen(`border="1"`, `cellpadding="0"`, `id="functions_table"`, `class="sortable"`)
	hw.TheadOpen()
	hw.Th("Calls", "Places", "Files", "Self (ms)", "Inclusive (ms)", "Incl/Excl %%", "Function")
	hw.TheadClose()
	hw.TbodyOpen()
	for _, fc := range functionCalls {
		if fc == nil {
			continue
		}
		reporter.writeOneFunctionMetric(hw, fc, exists, ownTimeStat, incTimeStat)
	}
	hw.TbodyClose()
	hw.TableClose()
	hw.BodyClose()
	hw.HtmlClose()
}

func (reporter *HtmlReporter) ReportFunctions(p *jsonprofile.Profile) {
	fileProfiles := p.FileProfileMap
	reporter.GenerateCssFile()
	reporter.GenerateJsFiles()
	log.Println("Cross referencing function call metrics...")
	functionCalls := fileProfiles.GetFunctionsSortedByExlusiveTime()

	jsFiles := []string{
		"jquery-min.js",
		"jquery-tablesorter-min.js",
		"tablesorter.js",
		"fprof.js",
		"function.js",
	}

	exists := reporter.GenerateSourceCodeHtmlFiles(fileProfiles, jsFiles)
	jsFiles[4] = "functions.js"
	reporter.GenerateFunctionsHtmlFile(p, jsFiles, exists, functionCalls)
}

func (reporter *HtmlReporter) Epilog() {
	fmt.Fprintln(reporter.ProfileFile, "</table>")
}
