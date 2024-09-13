package image_tools

import (
	"bytes"
	"image"
	_ "image/gif"  // Import to support GIF
	_ "image/jpeg" // Import to support JPEG
	_ "image/png"  // Import to support PNG
)

func GetImageFormat(data []byte) (string, error) {
	// Create a bytes.Reader from the image data
	reader := bytes.NewReader(data)

	// Decode the image
	_, format, err := image.Decode(reader)
	if err != nil {
		return "", err
	}

	// Return the image format
	return format, nil
}
