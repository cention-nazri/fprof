package json

import (
	"encoding/json"
	"fmt"
	"fprof/log"
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
	"files": {
		"/some/file" : [
			null,
			{
				"functions": [
					{
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
					}
				],
				"hits": 1,
				"total_duration": {
					"nsec": 1668,
					"sec": 0
				}
			}
			]
	},
	"start": {
		"nsec": 22,
		"sec": 42
	},
	"stop": {
		"nsec": 22,
		"sec": 52
	},
	"duration": {
		"nsec": 0,
		"sec": 10
	}
	}`)
	p := DecodeFromBytes(bytes)
	logFailIf(p == nil, "DecodeFromBytes must return profile pointer")
	logFailIf(p.Start.Nsec != 22 && p.Start.Sec != 42, "Start time")
	logFailIf(p.Stop.Nsec != 22 && p.Stop.Sec != 52, "Stop time")
	logFailIf(p.Duration.Nsec != 0 && p.Duration.Sec != 10, "Stop time")

	fp := p.FileProfileMap
	Lines := fp["/some/file"]
	assertEqual(len(Lines), 2, "Must found 2 line profile records")
	logFailIf(Lines[0] != nil, "Profile for line 0 must not exist")
	logFailIf(Lines[1].Hits == 0, "Profile record for line 1 must exist")
	failIf(Lines[1].TotalDuration.Sec != 0)
	failIf(Lines[1].TotalDuration.Nsec != 1668)
	failIf(Lines[1].Hits != 1)
	f := (*Lines[1].Functions)[0]
	failIf(f.Name != "FunctionName")
	failIf(f.Hits != 1)
	failIf(f.IsNative == true)
	failIf(f.InclusiveDuration.Sec != 0)
	failIf(f.InclusiveDuration.Nsec != 732324)
	failIf(f.ExclusiveDuration.Sec != 0)
	failIf(f.ExclusiveDuration.Nsec != 0)
	failIf(f.Filename != "/filename")
	failIf(f.StartLine != 88)

}

func TestTimeSpecAdd(tt *testing.T) {
	t = tt
	t1 := TimeSpec{0, ONE_BILLION - 1}
	t2 := TimeSpec{0, 1}
	t1.Add(t2)
	logFailIf(t1.Sec != 1 && t1.Nsec != 0, "Overflow Nsec sum must add 1 second")

	t1 = TimeSpec{0, 0}
	t2 = TimeSpec{0, 0}
	t1.Add(t2)
	logFailIf(t1.Sec != 0 && t1.Nsec != 0, "Adding 0.0 with 0.0 must produce 0.0")

}

func TestTimeSpecAverageInMilliseconds(tt *testing.T) {
	t = tt

	t1 := TimeSpec{1, 0}
	avg := t1.AverageInMilliseconds(1000)
	logFailIf(avg != float64(1), "AverageInMilliseconds() fail")

	t1 = TimeSpec{1, 999000000}
	avg = t1.AverageInMilliseconds(1000)
	assertEqual(avg, float64(1999/1000.0), "AverageInMilliseconds() failed")
}
