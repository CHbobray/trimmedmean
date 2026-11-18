// Package trimmedmean computes trimmed (truncated) means of numeric slices.
//
// A trimmed mean discards a proportion of the smallest and largest
// observations before averaging the rest. It is a robust estimator of
// central tendency: a few extreme values have limited influence on it,
// yet it keeps more information than the median.
//
// Trimming follows the convention of R's mean(x, trim = p): for a sample of
// size n, floor(n*p) observations are removed from each end of the sorted
// data. With one trimming argument the trimming is symmetric; with two
// arguments the first applies to the lower end and the second to the upper
// end (asymmetric trimming).
package trimmedmean

import (
	"errors"
	"fmt"
	"math"
	"slices"
)

// Number is the set of numeric types accepted by TrimmedMean.
// Both integer and floating-point slices are supported.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Errors returned by TrimmedMean. Use errors.Is to test for them.
var (
	ErrEmptySlice     = errors.New("trimmedmean: input slice is empty")
	ErrNaN            = errors.New("trimmedmean: input contains NaN")
	ErrTrimArgCount   = errors.New("trimmedmean: provide one (symmetric) or two (lower, upper) trimming proportions")
	ErrTrimOutOfRange = errors.New("trimmedmean: trimming proportion must be in [0, 0.5)")
	ErrNothingRemains = errors.New("trimmedmean: trimming removes every observation")
)

// TrimmedMean returns the mean of data after removing a proportion of the
// smallest and largest values.
//
//	TrimmedMean(x)             // no trimming: the ordinary mean
//	TrimmedMean(x, 0.05)       // symmetric: 5% from each end
//	TrimmedMean(x, 0.10, 0.02) // asymmetric: 10% from the low end, 2% from the high end
//
// Each proportion must lie in [0, 0.5). The number of observations removed
// from an end is floor(len(data) * proportion), matching R's mean(x, trim).
// The input slice is not modified.
func TrimmedMean[T Number](data []T, trim ...float64) (float64, error) {
	lowerTrim, upperTrim, err := resolveTrims(trim)
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, ErrEmptySlice
	}

	sorted, err := sortedFloats(data)
	if err != nil {
		return 0, err
	}

	n := len(sorted)
	lowerCount := int(math.Floor(float64(n) * lowerTrim))
	upperCount := int(math.Floor(float64(n) * upperTrim))
	kept := sorted[lowerCount : n-upperCount]
	if len(kept) == 0 {
		return 0, ErrNothingRemains
	}
	return mean(kept), nil
}

// resolveTrims turns the variadic trimming arguments into a lower and an
// upper proportion, validating each one.
func resolveTrims(trim []float64) (lower, upper float64, err error) {
	switch len(trim) {
	case 0:
		return 0, 0, nil
	case 1:
		lower, upper = trim[0], trim[0]
	case 2:
		lower, upper = trim[0], trim[1]
	default:
		return 0, 0, fmt.Errorf("%w: got %d", ErrTrimArgCount, len(trim))
	}
	for _, p := range []float64{lower, upper} {
		if !(p >= 0 && p < 0.5) { // also rejects NaN
			return 0, 0, fmt.Errorf("%w: got %v", ErrTrimOutOfRange, p)
		}
	}
	return lower, upper, nil
}

// sortedFloats copies data into a new float64 slice sorted ascending,
// rejecting NaN values (which have no place in an ordering).
func sortedFloats[T Number](data []T) ([]float64, error) {
	values := make([]float64, len(data))
	for i, v := range data {
		f := float64(v)
		if math.IsNaN(f) {
			return nil, fmt.Errorf("%w (index %d)", ErrNaN, i)
		}
		values[i] = f
	}
	slices.Sort(values)
	return values, nil
}

// mean returns the arithmetic mean of a non-empty slice. Values are summed
// with Kahan compensation so that long slices keep full precision.
func mean(values []float64) float64 {
	var sum, compensation float64
	for _, v := range values {
		y := v - compensation
		t := sum + y
		compensation = (t - sum) - y
		sum = t
	}
	return sum / float64(len(values))
}
