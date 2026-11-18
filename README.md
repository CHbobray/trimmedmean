# trimmedmean

A small Go package for computing **trimmed means**, with symmetric or asymmetric trimming.
Results match R's `mean(x, trim = p)`.

A trimmed mean sorts the data, drops a proportion of the smallest and largest values,
and averages what remains. It resists outliers much better than the ordinary mean while
using more of the data than the median.

An example program that uses this package, compares it with R, and runs a bootstrap
study is in the companion repository:
[CHbobray/trimmedmean-demo](https://github.com/CHbobray/trimmedmean-demo).

## Installation

Requires Go 1.24 or later.

```bash
go get github.com/CHbobray/trimmedmean@latest
```

## Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/CHbobray/trimmedmean"
)

func main() {
	data := []float64{2, 4, 6, 8, 10, 12, 14, 16, 18, 1000}

	symmetric, err := trimmedmean.TrimmedMean(data, 0.1) // 10% from each end
	if err != nil {
		log.Fatal(err)
	}
	asymmetric, _ := trimmedmean.TrimmedMean(data, 0.0, 0.1) // top 10% only
	plain, _ := trimmedmean.TrimmedMean(data)                // no trimming

	fmt.Println(symmetric, asymmetric, plain) // 11 10 109

	counts := []int{3, 5, 4, 6, 97, 5, 4, 6, 5, 0} // integer slices work too
	m, _ := trimmedmean.TrimmedMean(counts, 0.1)
	fmt.Println(m) // 4.75
}
```

### API

```go
func TrimmedMean[T Number](data []T, trim ...float64) (float64, error)
```

| Call | Meaning |
|---|---|
| `TrimmedMean(x)` | no trimming (ordinary mean) |
| `TrimmedMean(x, p)` | symmetric: proportion `p` from each end |
| `TrimmedMean(x, lower, upper)` | asymmetric: `lower` from the low end, `upper` from the high end |

- `T` may be any integer or floating-point type (`int`, `int64`, `uint8`, `float32`, `float64`, …).
- Each proportion must be in `[0, 0.5)`.
- The number removed from an end is `floor(n * p)`, the same rule as R.
- The input slice is never modified (the function sorts a copy).
- Sums use Kahan compensation, so long slices keep full precision.

### Errors

Errors can be tested with `errors.Is`:

| Error | Cause |
|---|---|
| `ErrEmptySlice` | `data` has no elements |
| `ErrNaN` | `data` contains NaN |
| `ErrTrimArgCount` | more than two trimming arguments |
| `ErrTrimOutOfRange` | a proportion is negative, NaN, or 0.5 or more |
| `ErrNothingRemains` | trimming would remove every observation (cannot happen with valid proportions; kept as a safeguard) |

## Testing

```bash
go test -v -cover ./...        # unit tests and runnable examples (97% coverage)
go test -run=^$ -bench=. ./... # benchmarks
```

The tests check:

- **Agreement with R.** Expected values come from `mean(x, trim = p)` in R
  (the script that produced them is `testdata/reference.R`).
- Integer slices, including small types such as `uint8`, which must be converted before summing.
- Asymmetric trimming, and that `TrimmedMean(x, p, p)` equals `TrimmedMean(x, p)`.
- Every error case, and that valid proportions always leave at least one value (n = 1 to 200).
- That the input slice is unchanged.
- Robustness: two gross outliers in 200 values move the trimmed mean by less than 1,
  while the ordinary mean moves by more than 50.

Benchmark on a 2-core Intel Xeon @ 2.10GHz (`trim = 0.05`, one allocation per call):

| n | time per call | memory per call |
|---:|---:|---:|
| 100 | 1.7 µs | 896 B |
| 10,000 | 0.8 ms | 80 KB |
| 1,000,000 | 113 ms | 8 MB |

Time is dominated by the O(n log n) sort.

## Files

| File | Role |
|---|---|
| `trimmedmean.go` | the package |
| `trimmedmean_test.go` | unit tests and benchmark |
| `example_test.go` | runnable examples (shown on pkg.go.dev) |
| `testdata/reference.R` | R script that produced the expected test values |

## GenAI Tools

This package was written with the help of Claude (Anthropic), an AI assistant. I gave it
the assignment description and the instructor's R bootstrap example. Claude drafted the
package design (a generic function with variadic trimming arguments to support both
symmetric and asymmetric trimming), the code, the unit tests, and this README. It
generated the expected test values by running R's `mean(x, trim = p)` and checked that
the Go results match. I reviewed the code and the results. The full conversation log is
available at [https://claude.ai/share/9448bf64-cd2d-49c9-b7c8-85bf1a67f41c](https://claude.ai/share/9448bf64-cd2d-49c9-b7c8-85bf1a67f41c).
