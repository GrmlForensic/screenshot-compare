package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strconv"
	"time"
)

// WIDTH defines the width of the created image
const WIDTH = 640

// HEIGHT defines the height of the created image
const HEIGHT = 400

func euclideanDistance(x1, y1, x2, y2 int) float64 {
	return math.Sqrt(math.Pow(float64(x2-x1), 2) + math.Pow(float64(y2-y1), 2))
}

func fivePoints(randNum int64) [5][2]int {
	var result [5][2]int
	for i, d := range [5]int64{7, 11, 13, 17, 19} {
		x := (int64(randNum/d) + (randNum % (135 * d))) % WIDTH
		y := (int64(3*randNum/d) + (randNum % (287 * d))) % HEIGHT
		result[i][0] = int(x)
		result[i][1] = int(y)
	}
	return result
}

func drawRandom(img *image.RGBA, randNum int64) {
	five := fivePoints(randNum)
	moreWhite := func(v int64) int64 {
		return int64((220*v)/256) + 36
	}

	for x := 0; x < WIDTH; x++ {
		for y := 0; y < HEIGHT; y++ {
			d1 := euclideanDistance(x, y, five[0][0], five[0][1]) + 2*euclideanDistance(x, y, five[1][0], five[1][1])
			d2 := euclideanDistance(x, y, five[2][0], five[2][1]) + d1 - 5*euclideanDistance(x, y, five[3][0], five[3][1])
			d3 := euclideanDistance(x, y, five[4][0], five[4][1])

			r := moreWhite(int64(d1) % 256)
			g := moreWhite(int64(d2) % 256)
			b := moreWhite(int64(d3) % 256)

			c := color.RGBA{uint8(r), uint8(g), uint8(b), 255}
			img.Set(x, y, c)
		}
	}
}

// Draw actually draws an image based on `randNum` and stores the result at `filepath`
func Draw(filepath string, randNum int64) error {
	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{WIDTH, HEIGHT}})

	drawRandom(img, randNum)

	fd, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer fd.Close()

	return png.Encode(fd, img)
}

func main() {
	var num int64
	var filepath string

	switch len(os.Args) {
	case 1:
		num = time.Now().Unix()
		fmt.Printf("Using current time as random seed: %d\n", num)
	case 2:
		filepath = "randimg.png"
		fallthrough
	case 3:
		n, err := strconv.Atoi(os.Args[1])
		if err != nil {
			panic(fmt.Errorf("Expected integer as positional argument; got '%s'", os.Args[1]))
		}
		num = int64(n)
		if len(os.Args) > 2 {
			filepath = os.Args[2]
		}
	default:
		panic(fmt.Errorf("usage: ./randimg <integer> <output.png>"))
	}

	if err := Draw(filepath, num); err != nil {
		panic(err)
	}
}
