package screencap

import (
	"github.com/kbinani/screenshot"
	"image"
	"image/jpeg"
	"os"
	"time"
)

var bounds = screenshot.GetDisplayBounds(0)

func ScreenCap() (*image.RGBA, error) {
	return screenshot.CaptureRect(bounds)
}

func SaveTestCap() {
	screenCap, err := ScreenCap()
	if err != nil {
		panic(err)
	}
	file, err := os.OpenFile("./test/cap/"+time.Now().Format("2006-01-02.15-04-05")+".png", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	//png.Encode(file, screenCap)
	jpeg.Encode(file, screenCap, nil)
}
