package jsonprofile

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"testing"
)

var t *testing.T
var c = 0

func ok(r bool) {
	c++
	fmt.Print(c, " - ")
	if !r {
		fmt.Print("not ")
		t.Fail()
	}
	fmt.Println("ok")
}

func failIf(failed bool) {
	if failed {
		t.Fail()
	}
}

func logFailIf(failed bool, fmt string, args ...interface{}) {
	if failed {
		t.Errorf(fmt, args...)
	}
}

func isEqual(expected, got interface{}) bool {
	if reflect.DeepEqual(expected, got) {
		return true
	}
	if reflect.ValueOf(expected) == reflect.ValueOf(got) {
		return true
	}

	if fmt.Sprintf("%v", expected) == fmt.Sprintf("%v", got) {
		return true
	}

	return false
}

func assertEqual(got, expected interface{}, fmt string, args ...interface{}) {
	if !isEqual(expected, got) {
		t.Fail()
		t.Logf(fmt, args...)
		t.Logf("  Expected: %v", expected)
		t.Logf("       Got: %v", got)
	}
}

func TestDecodeFunctionCallers(tt *testing.T) {
	t = tt
	bytes := []byte(`{
		"at": 66,
		"file": "/some/file",
		"frequency": 1,
		"name": "printf",
		"total_duration": {
			"nsec": 23989,
			"sec": 42
		}
	}`)

	var caller FunctionCaller
	err := json.Unmarshal(bytes, &caller)
	if err != nil {
		log.Fatal(err)
	}
	assertEqual(caller.At, 66, "caller.At")
	assertEqual(caller.Filename, "/some/file", "caller.Filename")
	assertEqual(caller.Frequency, 1, "caller.Frequency")
	assertEqual(caller.Name, "printf", "caller.Name")
	assertEqual(caller.TotalDuration, TimeSpec{42, 23989}, "caller.TotalDuration")
}

func TestDecodeFromBytes(tt *testing.T) {
	t = tt
	bytes := []byte(`{
		"/some/file" : [
			null,
			{
				"function": {
					"filename" : "/filename",
					"start_line" : 88,
					"callers": null,
					"exclusive_duration": {
						"nsec": 0,
						"sec": 0
					},
					"hits": 1,
					"inclusive_duration": {
						"nsec": 732324,
						"sec": 0
					},
					"is_native": false,
					"name": "FunctionName"
				},
				"hits": 1,
				"total_duration": {
					"nsec": 1668,
					"sec": 0
				}
			}
			]
	}
	`)
	fp := DecodeFromBytes(bytes)

	logFailIf(fp == nil, "DecodeFromBytes must not return non nil")
	Lines := fp["/some/file"]
	assertEqual(len(Lines), 2, "Must found 2 line profile records")
	logFailIf(Lines[0] != nil, "Profile for line 0 must not exist")
	logFailIf(Lines[1].Hits == 0, "Profile record for line 1 must exist")
	failIf(Lines[1].TotalDuration.Sec != 0)
	failIf(Lines[1].TotalDuration.Nsec != 1668)
	failIf(Lines[1].Hits != 1)
	failIf(Lines[1].Function.Name != "FunctionName")
	failIf(Lines[1].Function.Hits != 1)
	failIf(Lines[1].Function.IsNative == true)
	failIf(Lines[1].Function.InclusiveDuration.Sec != 0)
	failIf(Lines[1].Function.InclusiveDuration.Nsec != 732324)
	failIf(Lines[1].Function.ExclusiveDuration.Sec != 0)
	failIf(Lines[1].Function.ExclusiveDuration.Nsec != 0)
	failIf(Lines[1].Function.Filename != "/filename")
	failIf(Lines[1].Function.StartLine != 88)

}
