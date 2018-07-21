package v1

import "math"

// WR as defined by standard BT.601 by CCIR
const WR = float64(0.299)

// WG as defined by standard BT.601 by CCIR
const WG = float64(0.587)

// WB as defined by standard BT.601 by CCIR
const WB = float64(0.114)

// toNRGBA converts a RGBA color to un-alpha-scaled NRGBA
// based on https://golang.org/src/image/color/color.go?s=4600:4767
func toNRGBA(r, g, b, a uint32) (float64, float64, float64, float64) {
	d := float64(a)
	if d == 0.0 {
		// if transparent, return black, not NaN
		return 0.0, 0.0, 0.0, 0.0
	}
	return float64(r*0xFFFF) / d, float64(g*0xFFFF) / d, float64(b*0xFFFF) / d, d
}

// toYUV converts a RGB color to the Y'UV color space
func toYUV(r, g, b float64) (float64, float64, float64) {
	// https://en.wikipedia.org/wiki/YUV#SDTV_with_BT.601
	yPrime := WR*r + WG*g + WB*b
	return yPrime, 0.492 * (b - yPrime), 0.877 * (r - yPrime)
}

func euclideanDistance(a, x, b, y, c, z float64) float64 {
	return math.Sqrt(math.Pow(a-x, 2) + math.Pow(b-y, 2) + math.Pow(c-z, 2))
}
