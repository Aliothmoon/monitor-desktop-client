package main

import (
	"embed"
	"fmt"
	"time"

	"github.com/energye/energy/v2/cef"
	"github.com/energye/energy/v2/cef/ipc"
	"github.com/energye/energy/v2/consts"
	"github.com/energye/golcl/lcl"
	"github.com/energye/golcl/lcl/types"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/process"
)

//go:embed resources
//go:embed web/dist
var resources embed.FS

func main() {
	// 全局初始化
	cef.GlobalInit(nil, resources)

	// 创建应用程序
	app := cef.NewApplication()

	// 本地资源配置
	cef.BrowserWindow.Config.Title = "考试客户端系统"
	cef.BrowserWindow.Config.Width = 1024
	cef.BrowserWindow.Config.Height = 768
	cef.BrowserWindow.Config.Url = "fs://energy"
	cef.BrowserWindow.Config.LocalResource(cef.LocalLoadConfig{
		ResRootDir: "web/dist",
		FS:         resources,
	}.Build())

	// 初始化浏览器事件
	cef.BrowserWindow.SetBrowserInit(func(event *cef.BrowserEvent, window cef.IBrowserWindow) {
		// 获取系统信息并发送给前端
		go func() {
			// 等待浏览器完全加载
			time.Sleep(time.Second)

			// 获取操作系统信息
			info, err := host.Info()
			if err == nil {
				osInfo := fmt.Sprintf("系统: %s %s, 主机名: %s",
					info.Platform, info.PlatformVersion, info.Hostname)
				ipc.Emit("systemInfo", osInfo)
			}

			// 定时更新进程信息
			go updateProcessInfo()
		}()

		// 登录处理
		ipc.On("login", func(username, password string) {
			// 这里应该调用实际的登录API
			// 模拟登录过程
			//time.Sleep(time.Second * 2)

			// 模拟登录成功
			examInfo := map[string]interface{}{
				"examId":      "EX-2023-001",
				"title":       "2023年度技能测试",
				"startTime":   time.Now().Format("2006-01-02 15:04:05"),
				"endTime":     time.Now().Add(time.Hour * 2).Format("2006-01-02 15:04:05"),
				"duration":    120, // 分钟
				"studentName": username,
			}

			ipc.Emit("loginResult", true, examInfo)
		})

		// 打开内嵌浏览器
		ipc.On("openBrowser", func(url string) {
			handle := cef.InitializeWindowHandle()
			rect := types.TRect{}
			chromium := cef.NewChromium(nil, nil)

			// 设置关闭事件
			chromium.SetOnBeforeClose(func(sender lcl.IObject, browser *cef.ICefBrowser) {
				// 不做任何处理，不关闭主程序
			})

			// 设置HTTPS网站访问信息监控
			chromium.SetOnLoadStart(func(sender lcl.IObject, browser *cef.ICefBrowser, frame *cef.ICefFrame, transitionType consts.TCefTransitionType) {
				currentUrl := frame.Url()
				if len(currentUrl) > 0 {
					fmt.Println("访问网站:", currentUrl)
					ipc.Emit("browserVisit", currentUrl)
				}
			})

			// 创建浏览器
			chromium.CreateBrowserByWindowHandle(handle, rect, "考试浏览器", nil, nil, true)
		})
	})

	// 运行应用
	cef.Run(app)
}

// 更新进程信息
func updateProcessInfo() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for range ticker.C {
		processes, err := process.Processes()
		if err != nil {
			continue
		}

		// 只获取前10个进程信息
		var processInfo []map[string]interface{}
		count := 0

		for _, p := range processes {
			if count >= 10 {
				break
			}

			name, err := p.Name()
			if err != nil {
				continue
			}

			// 内存使用
			memInfo, err := p.MemoryInfo()
			var memUsage uint64
			if err == nil && memInfo != nil {
				memUsage = memInfo.RSS / 1024 / 1024 // MB
			}

			// CPU使用
			cpuPercent, _ := p.CPUPercent()

			processInfo = append(processInfo, map[string]interface{}{
				"pid":    p.Pid,
				"name":   name,
				"memory": memUsage,
				"cpu":    cpuPercent,
			})

			count++
		}

		if len(processInfo) > 0 {
			ipc.Emit("processInfo", processInfo)
		}
	}
}
