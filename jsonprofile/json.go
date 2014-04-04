package jsonprofile

import (
	"fmt"
	"io"
	"log"
	"sort"
	"time"
)

import "encoding/json"

const ONE_BILLION = 1000000000
const ONE_MILLION = 1000000

type Profile struct {
	Start          TimeSpec    `json:"start"`
	Stop           TimeSpec    `json:"stop"`
	Duration       TimeSpec    `json:"duration"`
	FileProfileMap FileProfile `json:"files"`
}

type FileProfile map[string][]*LineProfile

type FunctionProfileSlice []*FunctionProfile
type FunctionCallerSlice []*FunctionCaller
type FunctionCall struct {
	To              *FunctionProfile
	CallsMade       Counter
	TimeInFunctions TimeSpec
}
type FunctionCallSlice []*FunctionCall

type NameSpacedEntity struct {
	NameSpace string `json:"namespace"`
	Name      string `json:"name"`
}

type HitCount struct {
	Hits Counter `json:"hits"`
}

type Counter uint64
type LineProfile struct {
	Function *FunctionProfile `json:"function"`
	HitCount
	TotalDuration TimeSpec `json:"total_duration"`
	FunctionCalls FunctionCallSlice
	/* Cumulative of calls in FunctionCalls: */
	CallsMade       Counter
	TimeInFunctions TimeSpec
}

type FunctionProfile struct {
	NameSpacedEntity
	Filename          string              `json:"filename"`
	StartLine         Counter             `json:"start_line"`
	ExclusiveDuration TimeSpec            `json:"exclusive_duration"`
	InclusiveDuration TimeSpec            `json:"inclusive_duration"`
	IsNative          bool                `json:"is_native"`
	Callers           FunctionCallerSlice `json:"callers"`
	HitCount
	OwnTime TimeSpec
}

func removeParenthesis(name string) string {
	l := len(name)
	if l > 2 && name[l-2:] == "()" {
		return name[:l-2]
	}
	return name
}
func (f *NameSpacedEntity) FullName() string {
	if len(f.NameSpace) > 0 {
		return f.NameSpace + "." + f.Name
	}
	return removeParenthesis(f.Name)
}

func (fc *FunctionProfile) GetTimeSpentByUnknownCallers() *TimeSpec {
	known := TimeSpec{0,0}
	for _, c := range fc.Callers {
		known.Add(c.TotalDuration)
	}
	unknown := fc.InclusiveDuration
	unknown.Subtract(known)
	return &unknown
}

func (fc *FunctionProfile) CountCallingPlaces() int {
	if fc.Callers == nil {
		return 0
	}
	return len(fc.Callers)
}

func (fc *FunctionProfile) CountCallingFiles() int {
	files := make(map[string]int)
	//log.Println("callers for CallingFiles", fc.Callers)
	for _, v := range fc.Callers {
		files[v.Filename]++
	}
	return len(files)
}

type TimeSpec struct {
	Sec  int64
	Nsec int64
}

type FunctionCaller struct {
	At        Counter `json:"at"`
	Filename  string  `json:"file"`
	Frequency Counter `json:"frequency"`
	NameSpacedEntity
	TotalDuration TimeSpec `json:"total_duration"`
}

func (ts *TimeSpec) AverageInMilliseconds(n Counter) float64 {
	var total float64
	total = float64(ts.Sec*ONE_BILLION+ts.Nsec) / ONE_MILLION
	return float64(total) / float64(n)
}

func (f *FunctionProfile) CalculateOwnTime() {
	f.OwnTime = f.InclusiveDuration
	f.OwnTime.Subtract(f.ExclusiveDuration)
}

func (ts *TimeSpec) Subtract(other TimeSpec) {
	if ts.Nsec < other.Nsec {
		ts.Sec--
		ts.Nsec += ONE_BILLION
	}
	ts.Nsec = ts.Nsec - other.Nsec
	ts.Sec = ts.Sec - other.Sec
}

func (ts *TimeSpec) Add(other TimeSpec) {
	ts.Nsec += other.Nsec
	if ts.Nsec >= ONE_BILLION {
		ts.Sec++
		ts.Nsec = ts.Nsec % ONE_BILLION
	}
	ts.Sec += other.Sec
}

func (ts TimeSpec) InMillisecondsStr() string {
	return fmt.Sprintf("%.3f", ts.InMilliseconds())
}

func (ts TimeSpec) InMilliseconds() float64 {
	return float64(ts.Sec)*1000 + float64(ts.Nsec)/1000000
}

func (ts TimeSpec) Time() string {
	return ts.String()
}

func (ts TimeSpec) String() string {
	return time.Unix(ts.Sec, ts.Nsec).String()
}

func (ts *TimeSpec) IsLessThan(other *TimeSpec) bool {
	if ts.Sec < other.Sec {
		return true
	}
	if ts.Sec > other.Sec {
		return false
	}
	return ts.Nsec < other.Nsec

}

func (n Counter) EmptyIfZero() interface{} {
	if n > 0 {
		return n
	}
	return ""
}

func (ts TimeSpec) NonZeroMsOrNone() string {
	if ts.Sec != 0 || ts.Nsec != 0 {
		return ts.InMillisecondsStr()
	}
	return ""
}

func (ts TimeSpec) InSeconds() float64 {
	return float64(ts.Sec) + float64(ts.Nsec)/1000000000
}

func DecodeFromBytes(b []byte) *Profile {
	var o Profile

	err := json.Unmarshal(b, &o)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return &o
}

func From(stream io.Reader) *Profile {
	var o Profile

	r := json.NewDecoder(stream)
	if r == nil {
		log.Print("Error creating decoder from stream")
		return nil
	}

	err := r.Decode(&o)
	if err != nil {
		log.Print(err)
		return nil
	}
	return &o

}

/* Note we use j, i to sort descending in all Less() implementations */
func (p FunctionCallSlice) Len() int { return len(p) }
func (p FunctionCallSlice) Less(j, i int) bool {
	return p[i].TimeInFunctions.IsLessThan(&p[j].TimeInFunctions)
}
func (p FunctionCallSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p FunctionCallerSlice) Total() Counter {
	var nCalls Counter
	for _, caller := range p {
		nCalls += caller.Frequency
	}
	return nCalls
}
func (p FunctionCallerSlice) Len() int { return len(p) }
func (p FunctionCallerSlice) Less(j, i int) bool {
	if p[i].TotalDuration.IsLessThan(&p[j].TotalDuration) {
		return true
	}
	return false
}
func (p FunctionCallerSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (p FunctionProfileSlice) Len() int {
	return len(p)
}

func (p FunctionProfileSlice) Less(j, i int) bool {
	if p[i] == nil && p[j] == nil {
		return false
	}
	if p[i] != nil && p[j] == nil {
		return false
	}
	if p[i] == nil && p[j] != nil {
		return true
	}

	if p[i].OwnTime.Sec < p[j].OwnTime.Sec {
		return true
	} else if p[i].OwnTime.Sec > p[j].OwnTime.Sec {
		return false
	}
	if p[i].OwnTime.Nsec < p[j].OwnTime.Nsec {
		return true
	} else if p[i].OwnTime.Nsec > p[j].OwnTime.Nsec {
		return false
	}
	return true
}

func (p FunctionProfileSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (profile FileProfile) GetFunctionsSortedByExlusiveTime() FunctionProfileSlice {

	calls := profile.getFunctionCalls()
	sort.Stable(calls)
	return calls
}

func (fp FileProfile) injectCallerDurations(function *FunctionProfile) {
	callers := function.Callers
	for _, caller := range callers {
		//if caller.Filename == "eval()" {
		//	continue
		//}

		//fmt.Printf("%s:%d frequency: %d\n", caller.Filename, caller.At, caller.Frequency);
		lines := fp[caller.Filename]
		if lines != nil {
			//log.Println("file", caller.Filename)
			//log.Println("len lines", len(lines))
			//log.Println("caller.At", caller.At)
			if lines[caller.At-1] == nil {
				lines[caller.At-1] = &LineProfile{}
			}

			lp := lines[caller.At-1]
			lp.FunctionCalls = append(lp.FunctionCalls, &FunctionCall{function, caller.Frequency, caller.TotalDuration})
			lp.CallsMade += caller.Frequency
			lp.TimeInFunctions.Add(caller.TotalDuration)
		} else {
			log.Printf("?? No line profiles for [%s] ??", caller.Filename)
		}
	}
}

func (fileProfiles FileProfile) getFunctionCalls() FunctionProfileSlice {
	calls := make(FunctionProfileSlice, 50)

	for file, lineProfiles := range fileProfiles {
		//if file == "eval()" {
		//	continue
		//}
		for lineNo, lineProfile := range lineProfiles {
			if lineProfile == nil || lineProfile.Function == nil {
				continue
			}
			f := lineProfile.Function
			f.Filename = file
			f.StartLine = Counter(lineNo)
			f.CalculateOwnTime()
			calls = append(calls, f)
			sort.Stable(f.Callers)
			fileProfiles.injectCallerDurations(f)
		}
	}
	return calls
}
