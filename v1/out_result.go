package v1

import "time"

// Result is the result of an image comparison
type Result struct {
	// Config provides a representation of the values used to run this program.
	// Useful for debugging
	Config string
	// Runtime gives the duration of the comparison algorithm plus waiting times
	Runtime time.Duration
	// Gives the number of pixels with _any_ difference between the two images.
	// If PixelsDifferent is smaller or equal to AdmissibleDiffPixel, the Score must be necessarily zero.
	PixelsDifferent uint
	// True, if the program did not finish within the timeframe given by Timeout
	Timeout bool
	// Score gives the percentage of pixels with difference (minus AdmissibleDiffPixel) between two images.
	// Is a value between 0 (inclusively) and 1 (inclusively)
	Score float64

	config Config
}
