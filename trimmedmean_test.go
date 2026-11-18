package trimmedmean

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"testing"
)

const tolerance = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= tolerance*math.Max(1, math.Abs(b))
}

// Test data shared with the R script that produced the expected values
// (see testdata/reference.R). Expected values come from mean(x, trim = p).
var referenceInputs = map[string][]float64{
	"odd":  {5, 1, 9, 3, 7, 100, -40, 2, 8, 6, 4},
	"ties": {3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 50},
	"big":  squaresOfHalves(100),
}

// squaresOfHalves returns (0.5)^2, (1.5)^2, ..., matching R's seq(0.5, n-0.5, by = 1)^2.
func squaresOfHalves(n int) []float64 {
	values := make([]float64, n)
	for i := range values {
		x := float64(i) + 0.5
		values[i] = x * x
	}
	return values
}

func TestSymmetricTrimMatchesR(t *testing.T) {
	cases := []struct {
		name string
		trim float64
		want float64
	}{
		{"odd", 0, 9.545454545454545},
		{"odd", 0.05, 9.545454545454545},
		{"odd", 0.1, 5},
		{"odd", 0.25, 5},
		{"odd", 0.49, 5},
		{"ties", 0, 6.75},
		{"ties", 0.05, 6.75},
		{"ties", 0.1, 3},
		{"ties", 0.49, 3},
		{"big", 0, 3333.25},
		{"big", 0.05, 3174.9166666666665},
		{"big", 0.1, 3033.25},
		{"big", 0.25, 2708.25},
		{"big", 0.49, 2500.25},
	}
	for _, tc := range cases {
		got, err := TrimmedMean(referenceInputs[tc.name], tc.trim)
		if err != nil {
			t.Fatalf("%s trim=%v: unexpected error %v", tc.name, tc.trim, err)
		}
		if !almostEqual(got, tc.want) {
			t.Errorf("%s trim=%v: got %.15g, want %.15g", tc.name, tc.trim, got, tc.want)
		}
	}
}

func TestIntegerSlices(t *testing.T) {
	ints := make([]int, 20) // R: mean(1:20, trim = p) is 10.5 for every p
	for i := range ints {
		ints[i] = i + 1
	}
	for _, p := range []float64{0, 0.05, 0.1, 0.25, 0.49} {
		got, err := TrimmedMean(ints, p)
		if err != nil || got != 10.5 {
			t.Errorf("ints trim=%v: got %v, %v; want 10.5", p, got, err)
		}
	}

	small := []uint8{200, 1, 2, 3, 4} // sum would overflow uint8; conversion must happen first
	if got, _ := TrimmedMean(small); !almostEqual(got, 42) {
		t.Errorf("uint8 mean: got %v, want 42", got)
	}
}

func TestNoTrimIsOrdinaryMean(t *testing.T) {
	data := []float64{2, 4, 6, 8, 1000}
	got, err := TrimmedMean(data)
	if err != nil || got != 204 {
		t.Errorf("got %v, %v; want 204", got, err)
	}
}

func TestAsymmetricTrim(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} // n = 10
	cases := []struct {
		lower, upper float64
		want         float64
	}{
		{0.2, 0, 6.5},     // drop 1,2               -> mean(3..10)
		{0, 0.2, 4.5},     // drop 9,10              -> mean(1..8)
		{0.1, 0.3, 4.5},   // drop 1 and 8,9,10      -> mean(2..7)
		{0.15, 0.15, 5.5}, // floor(1.5)=1 each end  -> mean(2..9)
	}
	for _, tc := range cases {
		got, err := TrimmedMean(data, tc.lower, tc.upper)
		if err != nil || !almostEqual(got, tc.want) {
			t.Errorf("trim(%v,%v): got %v, %v; want %v", tc.lower, tc.upper, got, err, tc.want)
		}
	}

	// Two equal proportions must give the same result as one symmetric proportion.
	sym, _ := TrimmedMean(referenceInputs["big"], 0.05)
	asym, _ := TrimmedMean(referenceInputs["big"], 0.05, 0.05)
	if sym != asym {
		t.Errorf("symmetric %v != asymmetric-equal %v", sym, asym)
	}
}

func TestInputIsNotModified(t *testing.T) {
	data := []float64{9, 1, 8, 2, 7}
	original := slices.Clone(data)
	if _, err := TrimmedMean(data, 0.2); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(data, original) {
		t.Errorf("input was modified: %v", data)
	}
}

func TestErrors(t *testing.T) {
	cases := []struct {
		name string
		data []float64
		trim []float64
		want error
	}{
		{"empty", nil, []float64{0.1}, ErrEmptySlice},
		{"NaN in data", []float64{1, math.NaN()}, nil, ErrNaN},
		{"negative trim", []float64{1, 2}, []float64{-0.1}, ErrTrimOutOfRange},
		{"trim of one half", []float64{1, 2}, []float64{0.5}, ErrTrimOutOfRange},
		{"NaN trim", []float64{1, 2}, []float64{math.NaN()}, ErrTrimOutOfRange},
		{"bad upper trim", []float64{1, 2}, []float64{0.1, 0.7}, ErrTrimOutOfRange},
		{"three trims", []float64{1, 2}, []float64{0.1, 0.1, 0.1}, ErrTrimArgCount},
	}
	for _, tc := range cases {
		_, err := TrimmedMean(tc.data, tc.trim...)
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: got error %v, want %v", tc.name, err, tc.want)
		}
	}

	// floor(n*p) per end with p < 0.5 always leaves at least one value.
	for n := 1; n <= 200; n++ {
		data := make([]float64, n)
		if _, err := TrimmedMean(data, 0.4999, 0.4999); err != nil {
			t.Fatalf("n=%d: unexpected error %v", n, err)
		}
	}
}

// TestRobustToOutliers checks the property that motivates trimming:
// a few gross outliers move the ordinary mean far more than the trimmed mean.
func TestRobustToOutliers(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	clean := make([]float64, 200)
	for i := range clean {
		clean[i] = 100 + 10*rng.NormFloat64()
	}
	contaminated := slices.Clone(clean)
	contaminated[0], contaminated[1] = 10_000, 10_000

	plain, _ := TrimmedMean(contaminated)
	trimmed, _ := TrimmedMean(contaminated, 0.05)
	cleanTrimmed, _ := TrimmedMean(clean, 0.05)

	if math.Abs(trimmed-cleanTrimmed) > 1 {
		t.Errorf("trimmed mean shifted by %v", trimmed-cleanTrimmed)
	}
	if math.Abs(plain-cleanTrimmed) < 50 {
		t.Errorf("expected ordinary mean to be pulled by outliers, got %v", plain)
	}
}

func BenchmarkTrimmedMean(b *testing.B) {
	for _, n := range []int{100, 10_000, 1_000_000} {
		rng := rand.New(rand.NewPCG(9999, uint64(n)))
		data := make([]float64, n)
		for i := range data {
			data[i] = rng.NormFloat64()
		}
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_, _ = TrimmedMean(data, 0.05)
			}
		})
	}
}
