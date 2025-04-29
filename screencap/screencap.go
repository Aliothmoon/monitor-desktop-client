package main

import (
	"github.com/kbinani/screenshot"
	"image"
	"image/png"
	"os"
)

var bounds = screenshot.GetDisplayBounds(0)

func ScreenCap() (*image.RGBA, error) {
	return screenshot.CaptureRect(bounds)
}

func main() {
	screenCap, err := ScreenCap()
	if err != nil {
		panic(err)
	}
	file, err := os.OpenFile("screen.png", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	png.Encode(file, screenCap)
}
