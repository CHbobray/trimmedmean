package trimmedmean_test

import (
	"fmt"

	"github.com/CHbobray/trimmedmean"
)

func ExampleTrimmedMean() {
	data := []float64{2, 4, 6, 8, 10, 12, 14, 16, 18, 1000}

	plain, _ := trimmedmean.TrimmedMean(data)               // ordinary mean
	symmetric, _ := trimmedmean.TrimmedMean(data, 0.1)      // 10% from each end
	upperOnly, _ := trimmedmean.TrimmedMean(data, 0.0, 0.1) // drop only the top 10%

	fmt.Printf("mean: %.1f\n", plain)
	fmt.Printf("symmetric 10%%: %.1f\n", symmetric)
	fmt.Printf("upper 10%% only: %.1f\n", upperOnly)
	// Output:
	// mean: 109.0
	// symmetric 10%: 11.0
	// upper 10% only: 10.0
}

func ExampleTrimmedMean_integers() {
	counts := []int{3, 5, 4, 6, 5, 4, 97, 5, 4, 6, 5, 0, 4, 5, 6, 5, 4, 5, 6, 5}
	m, err := trimmedmean.TrimmedMean(counts, 0.05)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%.4f\n", m)
	// Output: 4.8333
}
