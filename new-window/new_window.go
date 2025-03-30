//----------------------------------------
//
// Copyright © yanghy. All Rights Reserved.
//
// Licensed under Apache License Version 2.0, January 2004
//
// https://www.apache.org/licenses/LICENSE-2.0
//
//----------------------------------------

package main

import (
	"embed"
	"fmt"
	"github.com/energye/energy/v2/cef"
	"github.com/energye/energy/v2/cef/ipc"
	"github.com/energye/energy/v2/cef/ipc/callback"
	"github.com/energye/energy/v2/consts"
	_ "github.com/energye/energy/v2/examples/syso"
	"github.com/energye/golcl/lcl"
	"github.com/energye/golcl/lcl/types"
)

//go:embed resources
var resources embed.FS
var (
	chromium cef.IChromium
)

func main() {
	//全局初始化 每个应用都必须调用的
	cef.GlobalInit(nil, resources)
	//创建应用
	app := cef.NewApplication()

	//指定一个URL地址，或本地html文件目录
	cef.BrowserWindow.Config.Url = "fs://energy"
	cef.BrowserWindow.Config.LocalResource(cef.LocalLoadConfig{
		ResRootDir: "resources",
		FS:         resources,
	}.Build())
	//cef.BrowserWindow.Config.EnableClose = false
	cef.BrowserWindow.SetBrowserInit(func(event *cef.BrowserEvent, window cef.IBrowserWindow) {
		//浏览器窗口之后回调，在这里获取创建的浏览器ID
		event.SetOnAfterCreated(func(sender lcl.IObject, browser *cef.ICefBrowser, window cef.IBrowserWindow) bool {
			// 创建完之后再拿浏览器id
			fmt.Println("on-create-window-ok", browser.Identifier(), window.Id())
			ipc.Emit("on-create-window-ok", browser.Identifier(), window.Id())
			return false // 什么都不做
		})
		//浏览器窗口关闭回调, 在这里触发ipc事件通知主窗口
		event.SetOnClose(func(sender lcl.IObject, browser *cef.ICefBrowser, aAction *consts.TCefCloseBrowserAction, window cef.IBrowserWindow) bool {
			ipc.Emit("on-close-window", window.Id())
			return false
		})
		//---- ipc 监听事件
		// 监听事件, 创建新窗口
		ipc.On("create-window", func(name string) {
			handle := cef.InitializeWindowHandle()
			rect := types.TRect{}
			chromium = cef.NewChromium(nil, nil)
			//chromium.SetDefaultURL("https://www.baidu.com")
			chromium.SetOnBeforeClose(func(sender lcl.IObject, browser *cef.ICefBrowser) {
				app.QuitMessageLoop()
			})
			var tabURL string
			chromium.SetOnLoadStart(func(sender lcl.IObject, browser *cef.ICefBrowser, frame *cef.ICefFrame, transitionType consts.TCefTransitionType) {
				fmt.Println("OnLoadStart", browser.Identifier())
				if tabURL != "" {
					frame.LoadUrl(tabURL)
					tabURL = ""
				}
			})
			chromium.SetOnBeforePopup(func(sender lcl.IObject, browser *cef.ICefBrowser, frame *cef.ICefFrame, beforePopupInfo *cef.BeforePopupInfo, popupFeatures *cef.TCefPopupFeatures, windowInfo *cef.TCefWindowInfo, resultClient *cef.ICefClient, settings *cef.TCefBrowserSettings, resultExtraInfo *cef.ICefDictionaryValue, noJavascriptAccess *bool) bool {
				browser.ExecuteChromeCommand(consts.IDC_NEW_TAB, consts.CEF_WOD_CURRENT_TAB)
				tabURL = beforePopupInfo.TargetUrl
				fmt.Println("OnBeforePopup", tabURL)
				return true
			})
			chromium.SetOnOpenUrlFromTab(func(sender lcl.IObject, browser *cef.ICefBrowser, frame *cef.ICefFrame, targetUrl string, targetDisposition consts.TCefWindowOpenDisposition, userGesture bool) bool {
				fmt.Println("OpenUrlFromTab", tabURL)
				return false
			})
			chromium.CreateBrowserByWindowHandle(handle, rect, "tiny browser", nil, nil, true)

		})
		// 改变当前窗口大小
		ipc.On("resize", func(_type int, channel callback.IChannel) {
			println("resize type", _type, "channel", channel.ChannelId(), channel.BrowserId())
			win := cef.BrowserWindow.GetWindowInfo(channel.BrowserId())
			if win == nil {
				return
			}
			window.RunOnMainThread(func() {
				switch _type {
				case 1:
					win.SetSize(400, 200)
				case 2:
					win.SetSize(600, 400)
				case 3:
					win.SetSize(1024, 768)
				}
				win.SetCenterWindow(true)
			})
		})

	})
	//运行应用
	cef.Run(app)
}
