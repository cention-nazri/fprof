package jsonprofile

import (
	"log"
	"io"
)

import "encoding/json"

type FileProfile map[string][]*LineProfile

type LineProfile struct {
	Function      *FunctionProfile `json:"function"`
	Hits          uint64           `json:"hits"`
	TotalDuration TimeSpec         `json:"total_duration"`
}

type FunctionProfile struct {
	Name              string           `json:"name"`
	Hits              uint64           `json:"hits"`
	ExclusiveDuration TimeSpec         `json:"exclusive_duration"`
	InclusiveDuration TimeSpec         `json:"inclusive_duration"`
	IsNative          bool             `json:"is_native"`
	Callers           []FunctionCaller `json:"callers"`
}

type TimeSpec struct {
	Sec  uint64
	Nsec uint64
}

type FunctionCaller struct {
	At            uint64   `json:"at"`
	File          string   `json:"file"`
	Frequency     uint64   `json:"frequency"`
	Name          string   `json:"name"`
	TotalDuration TimeSpec `json:"total_duration"`
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
