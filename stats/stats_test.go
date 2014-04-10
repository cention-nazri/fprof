package stats

import (
	"testing"
)

func TestMedianAbsoluteDeviation(t *testing.T) {
	type Tests struct {
		values       []float64
		expectedStat Stats
	}

	tests := []Tests{
		{
			[]float64{},
			Stats{0.0, 0},
		},
		{
			[]float64{1},
			Stats{0.0, 1},
		},
		{
			[]float64{1, 1},
			Stats{0.0, 1},
		},
		{
			[]float64{1, 2},
			Stats{1.0, 2},
		},
		{
			[]float64{1, 2, 3},
			Stats{1.0, 2},
		},
		{
			[]float64{1, 2, 3, 3, 3},
			Stats{0.0, 3},
		},
	}

	for i, v := range tests {
		s := MadMedian(v.values)
		if s.MAD != v.expectedStat.MAD {
			t.Errorf("idx %v: s.MAD, got = %v, expected = %v", i, s.MAD, v.expectedStat.MAD)
		}
		if s.Median != v.expectedStat.Median {
			t.Errorf("idx %v: s.Median, got = %v, expected = %v", i, s.Median, v.expectedStat.Median)
		}
	}

}

func TestCalculateDeviationsAndMedian(t *testing.T) {
	values := []float64{1, 1, 1, 2, 2, 2, 3}
	med := calculateMedian(values)
	devs := calculateAbsoluteDeviation(values, med)
	expectedMedian := 2.0
	if med != expectedMedian {
		t.Errorf("med must be 2")
	}
	if len(devs) != len(values) {
		t.Errorf("length of devs(%v) must match length of values(%v)", len(devs), len(values))
	}
	expectedDevs := make([]float64, len(values))
	for i, v := range values {
		expectedDevs[i] = v - expectedMedian
		if expectedDevs[i] < 0 {
			expectedDevs[i] = -expectedDevs[i]
		}
	}
	for i, v := range expectedDevs {
		if devs[i] != v {
			t.Errorf("dev[%v] != expectedDevs[%v], got %v , expected %v", i, i, devs[i], v)
		}
	}
}

func TestMadMedian(t *testing.T) {
	s := MadMedian([]float64{1, 2, 3})
	if s.MAD != 1.0 {
		t.Errorf("s.MAD must be 1, got %v", s.MAD)
	}
	if s.Median != 2.0 {
		t.Errorf("s.Median must be 2, got %v", s.Median)
	}
}
