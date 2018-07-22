package main

import (
	"fmt"
	"os"

	scmp "github.com/GrmlForensic/screenshot-compare/v1"
)

// USAGE for CLI
const USAGE = `PARAMETERS

  [--colors <colorspace> | --timeout <duration> | --wait <duration>
  | --diffpixel <count> | --nodimerror] <base> <ref>

DESCRIPTION

  Compare two images and quantify their difference.

DURATION

  <duration> matches '\d+[ismh]'
    is a duration specifier. The prefix defines the value.
    The suffix defines the unit. Examples:
      '600i'   600 milliseconds       '2s'    2 seconds
      '1m'     1 minute               '24h'   24 hours

OPTIONS

  --colors <colorspace> ∈ {"RGB", "Y'UV"} with default value "RGB"
    RGB is the standard color model.
    "Y'UV" resembles the perception of the colors by the eye better.
    Hence the differences better quantify the visual differences.

  --timeout <duration> with default value "0s"
    Assigns a maximum runtime for the comparison algorithm.
    "0s" has the special meaning, that no runtime limit is imposed.

  --wait <duration> with default value "0s"
    Defines how long the program should wait before starting comparison

  --diffpixel <count> with default value "0"
    An integer specifying how many pixels with any differences shall
    be ignored in the score

  --nodimerror with default value false
    if true and dimensions of the images do not match, returns difference
    set to maximum. if false, return error with exit code 101.

  <base> is a required positional argument
    is a filepath to the base image (alpha channel is ignored)

  <ref> is a required positional argument
    is a filepath to the reference image (alpha channel represents transparency)

REMARKS

  Scoring uses a 64-bit floating point number.
  So this program is subject to floating point rounding errors.

EXIT CODE

  The return code is an integer with min. 0 and max. 102:
    0     no differences (every pixel has same RGB value)
    100   high difference
    101   any runtime error
    102   timeout reached
`

func showPotentialCLIError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, USAGE)
		fmt.Fprintf(os.Stderr, "\n\033[1merror:\033[0m "+err.Error()+"\n")
		os.Exit(101)
	}
}

func main() {
	conf := scmp.NewConfig()
	result := scmp.Result{}

	_, errEnv := conf.FromEnv(2)
	showPotentialCLIError(errEnv)
	_, errJSON := conf.FromJSON("", true, 2)
	showPotentialCLIError(errJSON)
	_, errArgs := conf.FromArgs(os.Args, USAGE, 2)
	showPotentialCLIError(errArgs)

	// even though Valid() is called within Compare, we want
	// to ensure it is represented as CLI error
	showPotentialCLIError(conf.Valid())

	// image comparison
	err := scmp.Compare(conf, &result)
	if err != nil && !result.Timeout {
		fmt.Fprintf(os.Stderr, "\n\033[1merror:\033[0m "+err.Error()+"\n")
		os.Exit(101)
	}

	// wait for result (either timeout or result)
	percent := float64(100 * result.Score)
	fmt.Printf("runtime:                %s\n", result.Runtime)
	fmt.Printf("timeout:                %t\n", result.Timeout)
	fmt.Printf("pixels different:       %d\n", result.PixelsDifferent)
	fmt.Printf("difference percentage:  %.3f %%\n", percent)

	if result.Timeout {
		os.Exit(102)
	} else {
		os.Exit(int(percent))
	}
}
