package v1

import (
	"fmt"
	"os"
	"time"
)

// Config defines the runtime configuration.
// Two runs of the executable with the same Config must yield the same result.
type Config struct {
	// ColorSpace to use for comparison.
	// Currently supported: {Y'UV, RGB}
	ColorSpace string
	// Timeout defines a duration threshold. Timeout does not consider PreWait time.
	// If comparison exceeds this duration threshold, it will terminate prematurely.
	Timeout time.Duration
	// PreWait defines how long to wait before comparison starts
	PreWait time.Duration
	// AdmissibleDiffPixel is a fixed number N of pixels that are
	// allowed to be different. The comparison score will ignore the
	// the first N pixels yielding _any_ difference
	AdmissibleDiffPixel uint
	// NoDimensionError returns the maximum difference value as Score if
	// dimensions do not match instead of returning an error
	NoDimensionError bool
	// BaseImg is the image to compare in memory
	BaseImg TaggedImage
	// RefImg is the image to compare with ("expected image").
	// RImg is the parsed image to compare with in memory
	RefImg TaggedImage
}

func (c *Config) Valid() error {
	if c.ColorSpace != "Y'UV" && c.ColorSpace != "RGB" {
		return fmt.Errorf(`color space is invalid`)
	}
	if c.BaseImg.Image == nil {
		return fmt.Errorf(`base image required`)
	}
	if c.RefImg.Image == nil {
		return fmt.Errorf(`reference image required`)
	}
	if c.BaseImg.Width == 0 {
		return fmt.Errorf(`width of base image to compare is zero`)
	}
	if c.BaseImg.Height == 0 {
		return fmt.Errorf(`height of base image to compare is zero`)
	}
	if c.RefImg.Width == 0 {
		return fmt.Errorf(`width of reference image to compare is zero`)
	}
	if c.RefImg.Height == 0 {
		return fmt.Errorf(`height of reference image to compare is zero`)
	}
	if c.BaseImg.MinX < 0 {
		return fmt.Errorf(`base image minimum x coordinate is smaller than 0`)
	}
	if c.BaseImg.MinY < 0 {
		return fmt.Errorf(`base image minimum y coordinate is smaller than 0`)
	}
	if c.RefImg.MinX < 0 {
		return fmt.Errorf(`reference image minimum x coordinate is smaller than 0`)
	}
	if c.RefImg.MinY < 0 {
		return fmt.Errorf(`reference image minimum y coordinate is smaller than 0`)
	}
	if c.AdmissibleDiffPixel > uint(c.BaseImg.Width*c.BaseImg.Height) {
		fmt.Fprintf(os.Stderr, "warning: admissible diff pixel > baseimage(width * height)\n")
	}
	if c.AdmissibleDiffPixel > uint(c.RefImg.Width*c.RefImg.Height) {
		fmt.Fprintf(os.Stderr, "warning: admissible diff pixel > reference-image(width * height)\n")
	}

	actualBWidth := c.BaseImg.Image.Bounds().Max.X - c.BaseImg.Image.Bounds().Min.X
	actualBHeight := c.BaseImg.Image.Bounds().Max.Y - c.BaseImg.Image.Bounds().Min.Y
	actualRWidth := c.RefImg.Image.Bounds().Max.X - c.RefImg.Image.Bounds().Min.X
	actualRHeight := c.RefImg.Image.Bounds().Max.Y - c.RefImg.Image.Bounds().Min.Y
	if c.BaseImg.MinX+c.BaseImg.Width > actualBWidth {
		return fmt.Errorf(`width of base image is smaller than MinX + comparison Width`)
	}
	if c.BaseImg.MinY+c.BaseImg.Height > actualBHeight {
		return fmt.Errorf(`height of base image is smaller than MinY + comparison Height`)
	}
	if c.RefImg.MinX+c.RefImg.Width > actualRWidth {
		return fmt.Errorf(`width of base image is smaller than MinX + comparison Width`)
	}
	if c.RefImg.MinY+c.RefImg.Height > actualRHeight {
		return fmt.Errorf(`height of base image is smaller than MinY + comparison Height`)
	}

	return nil
}

func (c *Config) String() string {
	return fmt.Sprintf(`{colors: %v, timeout: %s, wait: %s, diffpixel: %d, nodimerr: %t, baseimg: %s, refimg: %s}`,
		c.ColorSpace, c.Timeout, c.PreWait, c.AdmissibleDiffPixel, c.NoDimensionError, c.BaseImg.String(), c.RefImg.String())
}
