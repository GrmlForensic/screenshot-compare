package v1

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"time"
)

// Compare applies the two images available in Config and compares them pixel-by-pixel.
// The result will be stored in the Result argument. If the score cannot be computed,
// then error will be non-nil and give a reason.
func Compare(c *Config, r *Result) error {
	if c.BaseImg.Width != c.RefImg.Width || c.BaseImg.Height != c.RefImg.Height {
		if !c.NoDimensionError {
			msg := "image dimensions do not correspond; got %d×%d (base) and %d×%d (ref)\n"
			return fmt.Errorf(msg, c.BaseImg.Width, c.BaseImg.Height, c.RefImg.Width, c.RefImg.Height)
		} else {
			r.Timeout = false
			r.Runtime = time.Duration(0)
			r.Config = c.String()
			r.Score = 1.0
			return nil
		}
	}

	if err := c.Valid(); err != nil {
		return err
	}

	beforeTime := time.Now()

	// this goroutine waits for waiting time to pass
	if c.PreWait > time.Duration(0) {
		time.Sleep(c.PreWait)
	}

	if c.Timeout > time.Duration(0) {
		// use two goroutines
		type result struct {
			timeout bool
			err     error
		}
		timedOut := make(chan result, 1)

		go func() {
			time.Sleep(c.Timeout)
			timedOut <- result{err: nil, timeout: true}
		}()

		go func() {
			// processing
			err := compareImages(c, r, 0, c.BaseImg.Height)
			timedOut <- result{err: err, timeout: false}
		}()

		r.Config = c.String()

		t := <-timedOut
		if t.err != nil {
			return t.err
		}
		if t.timeout {
			r.Timeout = true
			r.Runtime = c.PreWait + c.Timeout
			return fmt.Errorf(`timeout %s exceeded`, c.Timeout)
		} else {
			r.Timeout = false
			r.Runtime = time.Now().Sub(beforeTime)
			return nil
		}
	} else {
		// use only main goroutine
		r.Config = c.String()
		err := compareImages(c, r, 0, c.BaseImg.Height)
		r.Runtime = time.Now().Sub(beforeTime)
		return err
	}
}

// compareImages corresponds to Compare, but begins comparison
// at y=yOffset and ends at y=yOffset+yCount
func compareImages(c *Config, r *Result, yOffset, yCount int) error {
	roundingErrorFactor := 1.25
	cumul := 0.0
	r.PixelsDifferent = 0

	for y := yOffset; y < yCount; y++ {
		for x := 0; x < c.BaseImg.Width; x++ {
			var d float64
			r1, g1, b1, _ := toNRGBA(c.BaseImg.Image.At(c.BaseImg.MinX+x, c.BaseImg.MinY+y).RGBA())
			r2, g2, b2, a2 := toNRGBA(c.RefImg.Image.At(c.RefImg.MinX+x, c.RefImg.MinY+y).RGBA())
			//log.Println(y, x, ":", "(1)", r1, g1, b1, a1, "(2)", r2, g2, b2, a2)

			switch c.ColorSpace {
			case "RGB":
				d = euclideanDistance(r1, r2, g1, g2, b1, b2) / 113510.0
			case "Y'UV":
				yPrime1, u1, v1 := toYUV(r1, g1, b1)
				yPrime2, u2, v2 := toYUV(r2, g2, b2)
				d = euclideanDistance(yPrime1, yPrime2, u1, u2, v1, v2) / 113510.0
			}

			if d != 0.0 {
				r.PixelsDifferent += 1
			}

			// NOTE only alpha channel of c.RefImg is considered
			alpha := a2 / 65535
			if alpha < 0.0 || alpha > 1.0 {
				panic(alpha) // should not occur
			}
			cumul += d * alpha
		}
	}

	// r.Runtime will be set from the outside
	r.Timeout = false
	r.Config = c.String()
	r.Score = cumul / float64(yCount*c.BaseImg.Width) * roundingErrorFactor
	if r.Score > 1.0 {
		r.Score = 1.0
	}
	return nil
}
