package jsonprofile

import (
	"log"
	"io"
	"sort"
)

import "encoding/json"

type FileProfile map[string][]*LineProfile

type FunctionProfileSlice []*FunctionProfile

type LineProfile struct {
	Function      *FunctionProfile `json:"function"`
	Hits          uint64           `json:"hits"`
	TotalDuration TimeSpec         `json:"total_duration"`
}

type FunctionProfile struct {
	Name              string           `json:"name"`
	Filename          string           `json:"filename"`
	StartLine         uint64           `json:"start_line"`
	Hits              uint64           `json:"hits"`
	ExclusiveDuration TimeSpec         `json:"exclusive_duration"`
	InclusiveDuration TimeSpec         `json:"inclusive_duration"`
	IsNative          bool             `json:"is_native"`
	Callers           []FunctionCaller `json:"callers"`
}

type TimeSpec struct {
	Sec  int64
	Nsec int64
}

type FunctionCaller struct {
	At            uint64   `json:"at"`
	File          string   `json:"file"`
	Frequency     uint64   `json:"frequency"`
	Name          string   `json:"name"`
	TotalDuration TimeSpec `json:"total_duration"`
}

func (ts TimeSpec) InMilliseconds() float64 {
	return float64(ts.Sec) * 1000 + float64(ts.Nsec) / 1000000
}

func (ts TimeSpec) NonZeroMsOrNone() interface{} {
	v := ts.InMilliseconds()
	if (v > 0) {
		return v
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

func (p FunctionProfileSlice) Len() int {
	return len(p)
}

func (p FunctionProfileSlice) Less(j, i int) bool {
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

func (profile *FileProfile )GetFunctionsSortedByExlusiveTime() FunctionProfileSlice {

	calls := profile.getFunctionCalls()
	sort.Sort(calls)
	return calls
}

func (fileProfiles *FileProfile) getFunctionCalls() FunctionProfileSlice {
	var calls FunctionProfileSlice

	for file, lineProfiles := range *fileProfiles {
		for lineNo, lineProfile := range lineProfiles {
			if lineProfile == nil || lineProfile.Function == nil {
				continue
			}
			lineProfile.Function.Filename = file
			lineProfile.Function.StartLine = uint64(lineNo)
			calls = append(calls, lineProfile.Function)
		}
	}
	return calls
}
