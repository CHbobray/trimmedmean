# Produces the expected values used in trimmedmean_test.go.
# Run with: Rscript testdata/reference.R
cases <- list(
  odd   = c(5, 1, 9, 3, 7, 100, -40, 2, 8, 6, 4),
  ties  = c(rep(3, 10), 1, 50),
  ints  = 1:20,
  big   = seq(0.5, 99.5, by = 1)^2
)
trims <- c(0, 0.05, 0.1, 0.25, 0.49)
for (nm in names(cases)) for (p in trims)
  cat(sprintf('{"%s", %s, %.17g},\n', nm, format(p), mean(cases[[nm]], trim = p)))
