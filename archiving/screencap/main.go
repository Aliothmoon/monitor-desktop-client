package main

import (
	"github.com/energye/energy/v2/cef"
	"github.com/energye/energy/v2/cef/ipc"
	demoCommon "github.com/energye/energy/v2/examples/common"
	_ "github.com/energye/energy/v2/examples/syso"
	"github.com/energye/golcl/lcl"
	"github.com/energye/golcl/lcl/rtl"
	"github.com/energye/golcl/lcl/types"
	"github.com/segmentio/ksuid"
	"golang.org/x/image/bmp"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	//Global initialization must be called
	cef.GlobalInit(nil, demoCommon.ResourcesFS())
	//Create an application
	app := cef.NewApplication()
	//Local load resources

	cef.BrowserWindow.Config.LocalResource(cef.LocalLoadConfig{
		ResRootDir: "resources",
		FS:         demoCommon.ResourcesFS(),
		Home:       "screenshot.html",
	}.Build())
	cef.BrowserWindow.Config.Width = 600
	cef.BrowserWindow.Config.Height = 400
	// run main process and main thread
	cef.BrowserWindow.SetBrowserInit(browserInit)
	//run app
	cef.Run(app)
}

// run main process and main thread
func browserInit(event *cef.BrowserEvent, window cef.IBrowserWindow) {
	var (
		schotForm *lcl.TForm
		image     *lcl.TImage
	)
	if window.IsLCL() {
		// 创建一个窗口显示截屏图片
		schotForm = lcl.NewForm(window.AsLCLBrowserWindow().BrowserWindow())
		// 窗口透明
		schotForm.SetAlphaBlend(true)
		// 无边框窗口
		//schotForm.SetBorderStyle(types.BsNone)
		// 窗口透明度
		//schotForm.SetAlphaBlendValue(155)
		// 窗口大小是整个显示器大小
		//schotForm.SetBoundsRect(window.AsLCLBrowserWindow().BrowserWindow().Monitor().BoundsRect())
		// 显示截屏图片
		image = lcl.NewImage(schotForm)
		image.SetParent(schotForm)
		image.SetAlign(types.AlClient)
		// 可以使用一些事件来处理截图.
		image.SetOnMouseMove(func(sender lcl.IObject, shift types.TShiftState, x, y int32) {
			println("MouseMove")
		})
		image.SetOnMouseDown(func(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {

		})
		image.SetOnMouseUp(func(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {

		})
	}
	// 屏幕截图
	ipc.On("screenshot", func() {
		log.Println("screenshot")
		future := ScreenCap(window)
		if future.Err != nil {
			log.Println(future.Err)
			return
		}
		log.Println("图片已保存到", future.File)
	})
}

type FileFuture struct {
	File string
	Err  error
}

func ScreenCap(window cef.IBrowserWindow) FileFuture {
	var c = make(chan FileFuture)
	defer close(c)
	start := time.Now()
	window.RunOnMainThread(func() {

		dc := rtl.GetDC(0)
		// 一定要释放
		defer rtl.ReleaseDC(0, dc)

		// 位图
		b := lcl.NewBitmap()
		defer b.Free()
		b.LoadFromDevice(dc)

		log.Println("屏幕截图耗时", time.Since(start))

		tmp := filepath.Join(os.TempDir(), ksuid.New().String())
		b.SaveToFile(tmp)
		log.Println("屏幕截图耗时", time.Since(start))
		go func() {
			defer func() {
				if err := recover(); err != nil {
					c <- FileFuture{Err: err.(error)}
				}
			}()
			defer os.Remove(tmp)
			f, err := os.Open(tmp)
			if err != nil {
				c <- FileFuture{Err: err}
				return
			}
			defer f.Close()
			pic, err := bmp.Decode(f)
			if err != nil {
				c <- FileFuture{Err: err}
				return
			}
			var file = filepath.Join(os.TempDir(), ksuid.New().String())

			output, err := os.Create(file)
			if err != nil {
				c <- FileFuture{Err: err}
				return
			}
			defer output.Close()

			encoder := png.Encoder{
				CompressionLevel: png.BestSpeed,
			}
			if err = encoder.Encode(output, pic); err != nil {
				c <- FileFuture{Err: err}
				return
			}
			c <- FileFuture{File: file}
		}()

	})
	log.Println("屏幕截图耗时", time.Since(start))
	return <-c
}
