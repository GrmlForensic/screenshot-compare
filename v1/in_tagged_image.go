package v1

import (
	"fmt"
	"image"
	"os"
)

// TaggedImage represents an image with explicit width, height, format and source values
type TaggedImage struct {
	// Image is an in-memory Go image.Image instance
	Image image.Image
	// Width gives the width of Image
	Width int
	// Height gives the height of Image
	Height int
	// MinX defines the smallest X coordinate to start comparison with
	MinX int
	// MinY defines the smallest Y coordinate to start comparison with
	MinY int
	// Format is the file format as returned by Go's image.Decode
	Format string
	// Source is a simple description for the source of this image (for example, its filepath)
	Source string
}

// FromFilepath reads an image from the given filepath
// and fills TaggedImage with its data
func (i *TaggedImage) FromFilepath(fp string) error {
	reader, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer reader.Close()
	decoded, format, err := image.Decode(reader)
	if err != nil {
		return err
	}

	i.Image = decoded
	i.Width = decoded.Bounds().Max.X - decoded.Bounds().Min.X
	i.Height = decoded.Bounds().Max.Y - decoded.Bounds().Min.Y
	i.MinX = decoded.Bounds().Min.X
	i.MinX = decoded.Bounds().Min.Y
	i.Format = format
	i.Source = fp

	return nil
}

// String returns the human-readable representation of TaggedImage
func (i *TaggedImage) String() string {
	tmpl := `{Width: %d, Height: %d, MinX: %d, MinY: %d, Format: '%s', Source: '%s', Image: %s}`
	if i.Image == nil {
		return fmt.Sprintf(tmpl, i.Width, i.Height, i.MinX, i.MinY, i.Format, i.Source, `nil`)
	} else {
		return fmt.Sprintf(tmpl, i.Width, i.Height, i.MinX, i.MinY, i.Format, i.Source, `<ready>`)

	}
}
