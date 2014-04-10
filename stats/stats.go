package stats

import (
	"math"
	"sort"
)

type Stats struct {
	MAD    float64 // Median Absolute Deviation
	Median float64
}

func calculateAbsoluteDeviation(values []float64, median float64) []float64 {
	deviations := make([]float64, len(values))
	for i, v := range values {
		deviations[i] = math.Abs(v - median)
	}
	return deviations
}

func calculateMedian(values []float64) float64 {
	m := len(values) / 2
	sort.Float64s(values)
	return values[m]
}

func MedianAbsoluteDeviation(values []float64) (float64, float64) {
	l := len(values)
	if l == 0 {
		return 0, 0
	}
	if l == 1 {
		return 0, values[0]
	}
	median := calculateMedian(values)
	deviations := calculateAbsoluteDeviation(values, median)

	sort.Float64s(deviations)
	median_dev := deviations[len(deviations)/2]
	if median_dev < 0 {
		median_dev = -median_dev
	}
	return median_dev, median
}

func MadMedian(values []float64) *Stats {
	MAD, Median := MedianAbsoluteDeviation(values)
	s := &Stats{MAD, Median}
	return s
}
