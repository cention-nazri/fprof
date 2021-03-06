package html

import (
	"bufio"
	"bytes"
	"fmt"
	"fprof/log"
	"html"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
)

import "fprof/report"
import "fprof/osutil"
import "fprof/stats"
import "fprof/json"

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
		osutil.CreateFile(htmlfile),
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
		fmt.Fprint(hw.realw, hw.w.String())
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
func (hw *HtmlWriter) ThOpen(attrs ...string)    { hw.beginln("th", attrs...) }
func (hw *HtmlWriter) ThClose()                  { hw.end("th") }
func (hw *HtmlWriter) TdOpen(attrs ...string)    { hw.begin("td", attrs...) }
func (hw *HtmlWriter) TdClose()                  { hw.end("td") }
func (hw *HtmlWriter) TdTitled(title string, content interface{}) {
	hw.begin("td", `title="`+title+`"`)
	hw.write(content)
	hw.TdCloseNoIndent()
}
func (hw *HtmlWriter) TdCloseNoIndent() {
	hw.Html("</td>\n")
	hw.indent--
}
func (hw *HtmlWriter) TdTitledWithClassOrEmpty(title, class, content string) {
	if len(content) == 0 {
		hw.Td("")
		return
	}

	if len(title) > 0 {
		hw.TdOpen(`class="`+class+`"`, `title="`+title+`"`)
	} else {
		hw.TdOpen(`class="` + class + `"`)
	}
	hw.write(content)
	hw.TdCloseNoIndent()
}
func (hw *HtmlWriter) TdWithClassOrEmpty(class, content string) {
	hw.TdTitledWithClassOrEmpty("", class, content)
}
func (hw *HtmlWriter) DivOpen(attrs ...string) { hw.begin("div", attrs...) }
func (hw *HtmlWriter) DivClose()               { hw.end("div") }
func (hw *HtmlWriter) Th(v ...interface{})     { hw.repeatIn("th", v...) }
func (hw *HtmlWriter) Td(v ...interface{})     { hw.repeatIn("td", v...) }
func (hw *HtmlWriter) Tr(v ...interface{})     { hw.repeatIn("tr", v...) }
func (hw *HtmlWriter) Div(v ...interface{})    { hw.repeatIn("div", v...) }

func New(reportDir string) *HtmlReporter {
	osutil.CreateDir(reportDir)
	r := HtmlReporter{}
	r.ReportDir = reportDir
	return &r
}

func (r *HtmlReporter) PathTo(file string) string {
	return r.ReportDir + "/" + file
}

func (r *HtmlReporter) htmlLineFilename(file string) string {
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

func htmlLink(fromFile, funcName, toFile string, lineNo json.Counter) string {
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
	if path.Dir(rfrom) == "." {
		r = rto
	} else if path.Dir(rto) != "." {
		r = "../" + pathToRoot(path.Dir(rfrom)) + rto
	} else {
		r = pathToRoot(path.Dir(rfrom)) + rto
	}
	return r
}

func (r *HtmlReporter) showCallers(hw *HtmlWriter, fp *json.FunctionProfile, indent string) {
	hideThreshold := 10

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
	if path.IsAbs(calleeFile) {
		calleeFile = r.htmlLineFilename(calleeFile)
	}
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
		if path.IsAbs(callerFile) {
			callerFile = r.htmlLineFilename(callerFile)
		}
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
}

func (r *HtmlReporter) showCallsMade(hw *HtmlWriter, lp *json.LineProfile, indent string) {
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

			link := calleeFQN
			if path.IsAbs(calleeFile) {
				calleeFile = r.htmlLineFilename(calleeFile)
				link = htmlLink(r.htmlLineFilename(hw.SourceFile), calleeFQN, calleeFile, calleeAt)
			}
			hw.commentln(indent, "Spent %vms %s %s()%s",
				c.TimeInFunctions.InMillisecondsStr(),
				callTxt,
				link,
				avgTxt)
		}
		if startHideAt > 0 {
			hw.Html(`</div>`)
		}
	}
}

type CodeTableHeader struct {
	line            string
	hits            string
	timeOnLine      string
	callsMade       string
	timeInFunctions string
}

var cth = CodeTableHeader{
	line:            "Line",
	hits:            "Hits",
	timeOnLine:      "Time on line (ms)",
	callsMade:       "Calls Made",
	timeInFunctions: "Time in functions",
}

func (r *HtmlReporter) writeOneSourceCodeLine(hw *HtmlWriter, lineNo int, lp *json.LineProfile, sourceLine *string, ownTimeStats, otherTimeStats *stats.Stats) {
	indent := ""
	hw.TrOpen()
	hw.TdOpen(`title="Line number"`)
	hw.Html(fmt.Sprintf(`<a id="%d">%d</a>`, lineNo, lineNo))
	hw.TdCloseNoIndent()

	if sourceLine != nil {
		indent = getFirstWhiteSpaces(*sourceLine)
	}

	if lp == nil {
		hw.TdTitled(cth.hits, "")
		hw.TdTitled(cth.timeOnLine, "")
		hw.TdTitled(cth.callsMade, "")
		hw.TdTitled(cth.timeInFunctions, "")
		hw.TdOpen(`class="s"`)
	} else {
		ownTime := lp.TotalDuration
		ownTime.Subtract(lp.TimeInFunctions)
		if lp.Hits > 0 {
			hw.TdTitled(cth.hits, lp.Hits)
		} else {
			hw.TdTitled(cth.hits, "")
		}

		hw.TdTitledWithClassOrEmpty(
			cth.timeOnLine,
			getSeverityClass(ownTime.InMilliseconds(), ownTimeStats),
			ownTime.NonZeroMsOrNone(),
		)

		hw.TdTitled(cth.callsMade, lp.CallsMade.EmptyIfZero())
		hw.TdTitledWithClassOrEmpty(cth.timeInFunctions, getSeverityClass(lp.TimeInFunctions.InMilliseconds(), otherTimeStats), lp.TimeInFunctions.NonZeroMsOrNone())

		hw.TdOpen(`class="s"`)
		if lp.Functions != nil {
			functions := *lp.Functions
			for _, f := range functions {
				r.showCallers(hw, f, indent)
			}
		}
	}

	if sourceLine != nil {
		hw.Html(html.EscapeString(*sourceLine))
		r.showCallsMade(hw, lp, indent)
	}
	hw.TdCloseNoIndent()
	hw.TrClose()
}

func makeEmptyLineProfiles(file string) []*json.LineProfile {
	return make([]*json.LineProfile, osutil.CountLine(file))
}

func fileExists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func (r *HtmlReporter) writeOneSourceCodeHtmlFile(file string, fileProfiles json.FileProfile, rootJsFiles []string, done chan bool) {
	htmlfile := r.ReportDir + "/" + r.htmlLineFilename(file)
	osutil.CreateDir(path.Dir(htmlfile))
	hw := NewHtmlWriter(file, htmlfile)
	defer hw.writeToDiskAsync(done)

	rootPath := pathToRoot(file)
	jsFiles := []string{}
	for _, file := range rootJsFiles {
		jsFiles = append(jsFiles, rootPath+"../"+file)
	}
	hw.HtmlWithCssBodyOpen(rootPath+"../css/style.css", jsFiles)
	hw.DivOpen(`class="left"`)
	hw.Html(file)
	hw.DivClose()
	writeSeverityLegend(hw)
	hw.TableOpen(`id="function_table"`, `border="1"`, `cellpadding="0"`, `class="sortable clear"`)
	hw.TheadOpen()
	hw.Th(cth.line, cth.hits, cth.timeOnLine, cth.callsMade, cth.timeInFunctions)
	hw.ThOpen(`style="text-align:left"`)
	hw.Html("Code")
	hw.ThClose()
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

	var sourceLine *string
	for i, lp := range lineProfiles {
		lineNo := i + 1
		sourceLine = nil
		if scanner.Scan() {
			line := scanner.Text()
			sourceLine = &line
		}
		r.writeOneSourceCodeLine(hw, lineNo, lp, sourceLine, ownTimeStats, otherTimeStats)
	}
	hw.TbodyClose()
	hw.TableClose()
	for i := 0; i < 50; i++ {
		hw.writeln("<br>")
	}
	hw.BodyClose()
	hw.HtmlClose()
}

func (r *HtmlReporter) GenerateJsFiles() {
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

	d := r.PathTo("js")
	jsFiles := map[string]string{
		path.Join(d, "jquery-min.js"):             JQuery,
		path.Join(d, "jquery-tablesorter-min.js"): JQueryTableSorter,
		path.Join(d, "fprof.js"):                  fprofJs,
		path.Join(d, "tablesorter.js"):            tableSorterJs,
		path.Join(d, "functions.js"):              functionsJs,
		path.Join(d, "function.js"):               functionJs,
	}

	osutil.CreateFiles(jsFiles)
}

func (r *HtmlReporter) GenerateCssFile() {
	cssFile := r.ReportDir + "/css/style.css"
	osutil.CreateDir(path.Dir(cssFile))
	css := osutil.CreateFile(cssFile)
	fmt.Fprint(css, `body {
	font-family: sans-serif;
}
.clear {
	clear: both;
}
.left {
	float: left;
}
div.legend {
	float: right;
	margin-bottom: 0.5em;
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

func (r *HtmlReporter) generateHtmlFilesParallerWorkers(exists map[string]bool, fileProfiles json.FileProfile, jsFiles []string, nWorkers int) {
	nFiles := 0
	for _, exist := range exists {
		if exist {
			nFiles++
		}
	}

	type Job struct {
		file string
		// TODO use lineprofiles instead
		fileProfiles json.FileProfile
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
				r.writeOneSourceCodeHtmlFile(j.file, j.fileProfiles, jsFiles, done)
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
	fmt.Printf("")
}

func (r *HtmlReporter) GenerateSourceCodeHtmlFiles(fileProfiles json.FileProfile, jsFiles []string) map[string]bool {
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

	r.generateHtmlFilesParallerWorkers(exists, fileProfiles, jsFiles, 8)
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

func getMADStats(functionCalls json.FunctionProfileSlice) (*stats.Stats, *stats.Stats) {
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

func (r *HtmlReporter) writeOneFunctionMetric(hw *HtmlWriter, fc *json.FunctionProfile, exists map[string]bool, ownTimeStat *stats.Stats, incTimeStat *stats.Stats) {
	ieRatio := ""
	inclMS := fc.InclusiveDuration.InMilliseconds()
	exclMS := fc.OwnTime.InMilliseconds()
	if inclMS > 0 {
		ieRatio = fmt.Sprintf("%3.1f", exclMS*100/inclMS)
	}
	hw.TrOpen()
	hw.TdTitled(fth.calls, fc.Hits)
	hw.TdTitled(fth.places, fc.CountCallingPlaces())
	hw.TdTitled(fth.files, fc.CountCallingFiles())

	hw.TdTitledWithClassOrEmpty(
		fth.selfMs,
		getSeverityClass(exclMS, ownTimeStat),
		fc.OwnTime.NonZeroMsOrNone(),
	)

	hw.TdTitledWithClassOrEmpty(
		fth.inclusiveMs,
		getSeverityClass(inclMS, incTimeStat),
		fc.InclusiveDuration.NonZeroMsOrNone(),
	)

	hw.TdTitled(fth.ratio, ieRatio)

	hw.TdOpen(`class="s"`)
	if exists[fc.Filename] {
		hw.write(htmlLink(".", fc.FullName(), r.htmlLineFilename(fc.Filename), fc.StartLine))
	} else {
		r.showCallers(hw, fc, "")
		hw.write(fc.FullName())
	}
	hw.TdCloseNoIndent()
	hw.TrClose()
}

var tableAttrs = []string{`border="1"`, `cellpadding="0"`}

func writeSeverityLegend(hw *HtmlWriter) {
	legend := []map[string]string{
		{"s_bad": "Bad"},
		{"s_high": "High"},
		{"s_medium": "Medium"},
		{"s_low": "Low"},
	}
	hw.DivOpen(`class="legend"`)
	hw.Html("Severity:")
	hw.TableOpen(tableAttrs...)
	hw.TrOpen()
	for _, u := range legend {
		for k, v := range u {
			hw.TdWithClassOrEmpty(k, " ")
			hw.Td(v)
		}
	}
	hw.TrClose()
	hw.TableClose()
	hw.DivClose()
}

type FunctionTableHeader struct {
	calls       string
	places      string
	files       string
	selfMs      string
	inclusiveMs string
	ratio       string
}

var fth = FunctionTableHeader{
	calls:       "Calls",
	places:      "Places",
	files:       "Files",
	selfMs:      "Self (ms)",
	inclusiveMs: "Inclusive (ms)",
	ratio:       "Incl/Excl %",
}

func (r *HtmlReporter) GenerateFunctionsHtmlFile(p *json.Profile, jsFiles []string, exists map[string]bool, functionCalls json.FunctionProfileSlice) {
	hw := NewHtmlWriter("", r.ReportDir+"/functions.html")
	defer hw.writeToDiskAsync(nil)

	ownTimeStat, incTimeStat := getMADStats(functionCalls)

	hw.HtmlWithCssBodyOpen("css/style.css", jsFiles)
	hw.DivOpen(`class="left"`)
	hw.Div("Start: " + p.Start.Time())
	hw.Div("Stop: " + p.Stop.Time())
	hw.Div("Duration: " + p.Duration.InMillisecondsStr() + "ms")
	hw.DivClose()
	writeSeverityLegend(hw)
	attrs := []string{`id="functions_table"`, `class="sortable clear"`}
	attrs = append(attrs, tableAttrs...)
	hw.TableOpen(attrs...)
	hw.TheadOpen()
	hw.Th(fth.calls, fth.places, fth.files, fth.selfMs, fth.inclusiveMs, fth.ratio)
	hw.ThOpen(`style="text-align:left"`)
	hw.Html("Function")
	hw.ThClose()
	hw.TheadClose()
	hw.TbodyOpen()
	for _, fc := range functionCalls {
		if fc == nil {
			continue
		}
		r.writeOneFunctionMetric(hw, fc, exists, ownTimeStat, incTimeStat)
	}
	hw.TbodyClose()
	hw.TableClose()
	hw.BodyClose()
	hw.HtmlClose()
}

func (r *HtmlReporter) ReportFunctions(p *json.Profile) {
	fileProfiles := p.FileProfileMap
	r.GenerateCssFile()
	r.GenerateJsFiles()
	log.Println("Cross referencing function call metrics...")
	functionCalls := fileProfiles.GetFunctionsSortedByExlusiveTime()

	jsFiles := []string{
		"js/jquery-min.js",
		"js/jquery-tablesorter-min.js",
		"js/tablesorter.js",
		"js/fprof.js",
		"js/function.js",
	}

	exists := r.GenerateSourceCodeHtmlFiles(fileProfiles, jsFiles)
	jsFiles[4] = "js/functions.js"
	r.GenerateFunctionsHtmlFile(p, jsFiles, exists, functionCalls)
}
