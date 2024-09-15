package tools

import (
	"image"
	"image/jpeg"
	"image/png"
	"os"

	"golang.org/x/image/webp"
)

// DecodeWebP decodes a WebP image from a file
func DecodeWebP(filepath string) (image.Image, error) {
	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode the WebP image
	img, err := webp.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

/*// EncodeWebP encodes an image into WebP format and saves it to a file
func EncodeWebP(img image.Image, filepath string) error {
	// Create the output file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Encode the image to WebP format
	err = webp.Encode(out, img, &webp.Options{Lossless: true})
	if err != nil {
		return err
	}
	return nil
}*/

// Helper functions for saving images as JPEG or PNG for testing
func saveJPEG(filepath string, img image.Image) {
	file, _ := os.Create(filepath)
	defer file.Close()
	jpeg.Encode(file, img, nil)
}

func savePNG(filepath string, img image.Image) {
	file, _ := os.Create(filepath)
	defer file.Close()
	png.Encode(file, img)
}
