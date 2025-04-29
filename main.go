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

	"monitor-desktop-client/devices"
)

//go:embed web/dist
var resources embed.FS

func main() {
	// 全局初始化
	cef.GlobalInit(nil, resources)

	// 创建应用程序
	app := cef.NewApplication()

	// 配置应用
	setupAppConfig()

	// 初始化浏览器事件
	cef.BrowserWindow.SetBrowserInit(setupBrowserEvents)

	// 运行应用
	cef.Run(app)
}

// 配置应用程序
func setupAppConfig() {
	// 本地资源配置
	cef.BrowserWindow.Config.Title = "考试客户端系统"
	cef.BrowserWindow.Config.Width = 1700
	cef.BrowserWindow.Config.Height = 1000
	//cef.BrowserWindow.Config.Url = " http://localhost:5173/"
	cef.BrowserWindow.Config.Url = "fs://energy"
	cef.BrowserWindow.Config.LocalResource(cef.LocalLoadConfig{
		ResRootDir: "web/dist",
		FS:         resources,
	}.Build())

	// 禁用开发者工具
	//cef.BrowserWindow.Config.ChromiumConfig().SetEnableDevTools(false)

}

// 设置浏览器事件处理
func setupBrowserEvents(event *cef.BrowserEvent, window cef.IBrowserWindow) {
	// 获取系统信息并发送给前端
	go collectSystemInfo()

	// 注册IPC事件处理
	registerIPCEvents()
}

// 收集系统信息
func collectSystemInfo() {
	// 等待浏览器完全加载
	time.Sleep(time.Second)

	// 获取操作系统信息
	info, err := host.Info()
	if err == nil {
		ipc.Emit("systemInfo", info.String())
	}

	// 定时更新进程信息
	go updateProcessInfo()
}

// 注册IPC事件
func registerIPCEvents() {
	// 登录处理
	registerLoginEvent()

	registerLoadEvent()

	// 打开内嵌浏览器
	registerOpenBrowserEvent()

	// 注册USB设备监控事件
	registerUSBMonitorEvent()

	// 注册设备信息获取事件
	registerDeviceInfoEvent()
}

func registerLoadEvent() {
	ipc.On("load", func() {
		fmt.Println("加载完成")
		info, err := host.Info()
		if err == nil {
			ipc.Emit("systemInfo", info.String())
		}
	})
}

// 注册登录事件处理
func registerLoginEvent() {
	ipc.On("login", func(username, password string) {
		// 这里应该调用实际的登录API
		// 模拟登录过程
		//time.Sleep(time.Second * 2)
		fmt.Println("用户登录")

		// 模拟登录成功
		examInfo := map[string]any{
			"examId":      "EX-2023-001",
			"title":       "2023年度技能测试",
			"startTime":   time.Now().Format("2006-01-02 15:04:05"),
			"endTime":     time.Now().Add(time.Hour * 2).Format("2006-01-02 15:04:05"),
			"duration":    120, // 分钟
			"studentName": username,
		}

		ipc.Emit("loginResult", true, examInfo)
	})

	ipc.On("logout", func() {
		fmt.Println("用户登出")
	})
}

// 注册打开浏览器事件处理
func registerOpenBrowserEvent() {
	ipc.On("openBrowser", func() {
		createEmbeddedBrowser()
	})
}

// 创建嵌入式浏览器
func createEmbeddedBrowser() {
	handle := cef.InitializeWindowHandle()
	rect := types.TRect{}

	// 创建浏览器配置，禁用开发者工具和菜单
	config := &cef.TCefChromiumConfig{}
	config.SetEnableMenu(false)
	config.SetEnableDevTools(false)

	chromium := cef.NewChromium(nil, config)

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
}

// 更新进程信息
func updateProcessInfo() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for range ticker.C {
		processInfo := collectProcessInfo()
		if len(processInfo) > 0 {
			ipc.Emit("processInfo", processInfo)
		}
	}
}

// 收集进程信息
func collectProcessInfo() []map[string]any {
	processes, err := process.Processes()
	if err != nil {
		return nil
	}

	// 只获取前10个进程信息
	var processInfo []map[string]any
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

		processInfo = append(processInfo, map[string]any{
			"pid":    p.Pid,
			"name":   name,
			"memory": memUsage,
			"cpu":    cpuPercent,
		})

		count++
	}

	return processInfo
}

// 注册USB设备监控事件
func registerUSBMonitorEvent() {
	ipc.On("getUSBDevices", func() {
		usbDevices, err := devices.GetUSBDevices()
		if err != nil {
			fmt.Println("获取USB设备失败:", err)
			ipc.Emit("usbDevicesResult", []string{}, err.Error())
			return
		}

		// 将设备信息格式化为前端可用的格式
		var result []map[string]interface{}
		for _, device := range usbDevices {
			deviceInfo := map[string]interface{}{
				"name":         device.DeviceName,
				"type":         device.DeviceType,
				"manufacturer": device.Manufacturer,
				"isStorage":    device.IsStorageType,
				"driveLetters": device.DriveLetters,
			}
			result = append(result, deviceInfo)
		}

		ipc.Emit("usbDevicesResult", result, "")
	})

	// 启动定时监控USB设备
	go monitorUSBDevices()
}

// 监控USB设备变化
func monitorUSBDevices() {
	var lastDevices []devices.USBDevice
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for range ticker.C {
		currentDevices, err := devices.GetUSBDevices()
		if err != nil {
			continue
		}

		// 检测设备变化
		if hasDeviceChanges(lastDevices, currentDevices) {
			// 将设备信息发送到前端
			var result []map[string]interface{}
			for _, device := range currentDevices {
				deviceInfo := map[string]interface{}{
					"name":         device.DeviceName,
					"type":         device.DeviceType,
					"manufacturer": device.Manufacturer,
					"isStorage":    device.IsStorageType,
					"driveLetters": device.DriveLetters,
				}
				result = append(result, deviceInfo)
			}

			ipc.Emit("usbDevicesChanged", result)
		}

		lastDevices = currentDevices
	}
}

// 检测设备是否有变化
func hasDeviceChanges(old, new []devices.USBDevice) bool {
	if len(old) != len(new) {
		return true
	}

	// 简单比较设备ID
	oldIDs := make(map[string]bool)
	for _, device := range old {
		oldIDs[device.DeviceID] = true
	}

	for _, device := range new {
		if !oldIDs[device.DeviceID] {
			return true
		}
	}

	return false
}

// 注册设备信息获取事件
func registerDeviceInfoEvent() {
	ipc.On("getDeviceInfo", func() {
		// 获取设备信息
		deviceInfo, err := devices.GetDeviceInfo()
		if err != nil {
			fmt.Println("获取设备信息失败:", err)
			ipc.Emit("deviceInfoResult", nil, err.Error())
			return
		}

		// 发送设备信息到前端
		ipc.Emit("deviceInfoResult", deviceInfo, "")

		// 打印格式化的设备信息到控制台
		fmt.Println("设备信息:")
		fmt.Println(devices.FormatDeviceInfo(deviceInfo))
	})
}
