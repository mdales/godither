package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"

	"github.com/nfnt/resize"
)

func naiveMono(oldImage image.Image) (image.Image, error) {

	bounds := oldImage.Bounds()
	newImage := image.NewGray(bounds)

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			r, g, b, _ := oldImage.At(x, y).RGBA()
			avg := ((r + g + b) / 3) >> 8
			new := color.Gray{Y: uint8(avg)}
			newImage.SetGray(x, y, new)
		}
	}

	return newImage, nil
}

func naiveDither(oldImage image.Image) (image.Image, error) {

	bounds := oldImage.Bounds()
	newImage := image.NewGray(bounds)

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			r, g, b, _ := oldImage.At(x, y).RGBA()
			avg := (r + g + b) / 3
			new := color.Gray{Y: uint8(avg>>15) * 255}
			newImage.SetGray(x, y, new)
		}
	}

	return newImage, nil
}

func addErrorValue(image *image.Gray, errorValue uint8, x int, y int) {
	bounds := image.Bounds()
	if (x >= 0) && (x < bounds.Dx()) && (y >= 0) && (y < bounds.Dy()) {
		current := image.GrayAt(x, y).Y
		image.SetGray(x, y, color.Gray{Y: errorValue + current})
	}
}

func atkinsonDither(oldImage image.Image) (image.Image, error) {

	bounds := oldImage.Bounds()
	newImage := image.NewGray(bounds)

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			// Source is 16 bit, turn into 8 bit mono
			// r, g, b, _ := oldImage.At(x, y).RGBA()
			// avg := uint8(((r + g + b) / 3) >> 8)
			_, g, _, _ := oldImage.At(x, y).RGBA()
			avg := uint8(g >> 8)

			current := newImage.GrayAt(x, y).Y

			thisValue := ((avg + current) >> 7) * 255
			newImage.SetGray(x, y, color.Gray{Y: thisValue})

			errorValue := uint8(int8((avg+current)-thisValue) >> 3)
			addErrorValue(newImage, errorValue, x+1, y)
			addErrorValue(newImage, errorValue, x, y+1)
			addErrorValue(newImage, errorValue, x+1, y+1)
			addErrorValue(newImage, errorValue, x+2, y)
			addErrorValue(newImage, errorValue, x, y+2)
			addErrorValue(newImage, errorValue, x-1, y+1)
		}
	}
	return newImage, nil
}

func main() {

	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s [INPUT JPG] [OUTPUT PNG]\n", os.Args[0])
	}

	input := os.Args[1]

	buf, err := os.Open(input)
	if err != nil {
		log.Fatalf("Failed to open input file: %v", err)
	}
	defer buf.Close()

	reader := bufio.NewReader(buf)
	img, _, err := image.Decode(reader)
	if err != nil {
		log.Fatalf("Failed to decode input file: %v", err)
	}

	resizedImage := resize.Resize(512, 0, img, resize.Lanczos3)

	newImage, err := atkinsonDither(resizedImage)
	if err != nil {
		log.Fatalf("Failed to process image: %v", err)
	}

	output := os.Args[2]
	buf2, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Failed to open file for output: %v", err)
	}
	defer buf2.Close()
	err = png.Encode(buf2, newImage)
	if err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}
}
