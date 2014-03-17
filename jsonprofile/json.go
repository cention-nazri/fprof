package jsonprofile

import (
	"log"
	"fmt"
	"io"
	"sort"
)

import "encoding/json"

const ONE_BILLION = 1000000000
const ONE_MILLION = 1000000

type FileProfile map[string][]*LineProfile

type FunctionProfileSlice []*FunctionProfile
type FunctionCallerSlice []*FunctionCaller
type FunctionCall struct {
	To *FunctionProfile
	CallsMade Counter
	TimeInFunctions TimeSpec
}
type FunctionCallSlice []FunctionCall

type NameSpacedEntity struct {
	NameSpace	  string	   `json:"namespace"`
	Name              string           `json:"name"`
}

type HitCount struct {
	Hits          Counter          `json:"hits"`
}

type Counter uint64
type LineProfile struct {
	Function      *FunctionProfile `json:"function"`
	HitCount
	TotalDuration TimeSpec         `json:"total_duration"`
	FunctionCalls FunctionCallSlice
	/* Cumulative of calls in FunctionCalls: */
	CallsMade Counter
	TimeInFunctions TimeSpec
}

type FunctionProfile struct {
	NameSpacedEntity
	Filename          string           `json:"filename"`
	StartLine         Counter           `json:"start_line"`
	HitCount
	ExclusiveDuration TimeSpec         `json:"exclusive_duration"`
	InclusiveDuration TimeSpec         `json:"inclusive_duration"`
	IsNative          bool             `json:"is_native"`
	Callers           FunctionCallerSlice `json:"callers"`
}

func (f *NameSpacedEntity) FullName() string {
	if len(f.NameSpace) > 0 {
		return f.NameSpace + "." + f.Name
	}
	return f.Name
}

func (fc *FunctionProfile) CountCallingPlaces() int {
	if fc.Callers == nil {
		return 0;
	}
	return len(fc.Callers)
}

func (fc *FunctionProfile) CountCallingFiles() int {
	files := make(map[string]int)
	log.Println("callers for CallingFiles", fc.Callers)
	for _, v := range(fc.Callers) {
		files[v.Filename]++
	}
	return len(files)
}

type TimeSpec struct {
	Sec  int64
	Nsec int64
}

type FunctionCaller struct {
	At            Counter   `json:"at"`
	Filename      string   `json:"file"`
	Frequency     Counter   `json:"frequency"`
	NameSpacedEntity
	TotalDuration TimeSpec `json:"total_duration"`
}

func (ts *TimeSpec) AverageInMilliseconds(n Counter) float64 {
	var total float64
	total = float64(ts.Sec * ONE_BILLION + ts.Nsec) / ONE_MILLION
	return float64(total) / float64(n)
}

func (ts *TimeSpec) Add(other TimeSpec) {
	ts.Nsec += other.Nsec
	if (ts.Nsec > ONE_BILLION) {
		ts.Sec++
	}
	ts.Sec += other.Sec
}

func (ts TimeSpec) InMillisecondsStr() string {
	return fmt.Sprintf("%.3f", ts.InMilliseconds())
}

func (ts TimeSpec) InMilliseconds() float64 {
	return float64(ts.Sec) * 1000 + float64(ts.Nsec) / 1000000
}

func (n Counter) EmptyIfZero() interface{} {
	if (n > 0) {
		return n
	}
	return ""
}

func (ts TimeSpec) NonZeroMsOrNone() string {
	if (ts.Sec != 0 || ts.Nsec != 0) {
		return ts.InMillisecondsStr()
	}
	return ""
}

func (ts TimeSpec) InSeconds() float64 {
	return float64(ts.Sec) + float64(ts.Nsec) / 1000000000
}

func DecodeFromBytes(b []byte) FileProfile {
	var o FileProfile

	err := json.Unmarshal(b, &o)
	if err != nil {
		return nil
	}
	return o
}

func From(stream io.Reader) FileProfile {
	var o FileProfile

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
	return o

}

func (p FunctionCallerSlice) Len() int { return len(p) }
func (p FunctionCallerSlice) Less(j, i int) bool {
	if p[i].Frequency < p[j].Frequency {
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

	if p[i].ExclusiveDuration.Sec < p[j].ExclusiveDuration.Sec {
		return true
	} else if p[i].ExclusiveDuration.Sec > p[j].ExclusiveDuration.Sec {
		return false
	}
	if p[i].ExclusiveDuration.Nsec < p[j].ExclusiveDuration.Nsec {
		return true
	} else if p[i].ExclusiveDuration.Nsec > p[j].ExclusiveDuration.Nsec {
		return false
	}
	return true
}

func (p FunctionProfileSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (profile FileProfile )GetFunctionsSortedByExlusiveTime() FunctionProfileSlice {

	calls := profile.getFunctionCalls()
	sort.Sort(calls)
	return calls
}

func (fp FileProfile) injectCallerDurations(function *FunctionProfile) {
	callers := function.Callers
	for _, caller := range(callers) {

		//fmt.Printf("%s:%d frequency: %d\n", caller.Filename, caller.At, caller.Frequency);
		lines := fp[caller.Filename]
		if lines != nil {
			log.Println("file", caller.Filename)
			log.Println("len lines", len(lines))
			log.Println("caller.At", caller.At)
			if lines[caller.At-1] == nil {
				lines[caller.At-1] = &LineProfile{}
			}

			lp := lines[caller.At-1]
			lp.FunctionCalls = append(lp.FunctionCalls, FunctionCall{function, caller.Frequency, caller.TotalDuration})
			lp.CallsMade += caller.Frequency
			lp.TimeInFunctions.Add(caller.TotalDuration)
		} else {
			log.Printf("?? No line profiles for [%s] ??", caller.Filename)
		}
	}
}

func (fileProfiles FileProfile) getFunctionCalls() FunctionProfileSlice {
	calls := make(FunctionProfileSlice,50)

	for file, lineProfiles := range fileProfiles {
		for lineNo, lineProfile := range lineProfiles {
			if lineProfile == nil || lineProfile.Function == nil {
				continue
			}
			lineProfile.Function.Filename = file
			lineProfile.Function.StartLine = Counter(lineNo)
			calls = append(calls, lineProfile.Function)
			//sort.Sort(lineProfile.Function.Callers)
			fileProfiles.injectCallerDurations(lineProfile.Function)
		}
	}
	return calls
}
