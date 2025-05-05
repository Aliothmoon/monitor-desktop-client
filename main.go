package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"monitor-desktop-client/compose"
	"monitor-desktop-client/utils"
	"time"

	"github.com/energye/energy/v2/cef"
	"github.com/energye/energy/v2/cef/ipc"
	"github.com/energye/energy/v2/consts"
	"github.com/energye/golcl/lcl"
	"github.com/energye/golcl/lcl/types"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/process"

	"monitor-desktop-client/devices"
)

//go:embed web/dist
var resources embed.FS

// Config 应用配置
type Config struct {
	ServerURL  string // 服务器URL
	AccountID  int    // 考生账号ID
	ExamID     int    // 考试ID
	Token      string // 认证令牌
	WSEndpoint string // WebSocket端点
}

// 全局配置
var appConfig = &Config{
	//ServerURL:  "http://localhost:8777", // 默认服务器地址
	//WSEndpoint: "ws://localhost:8777/ws/monitor",
	ServerURL:  "https://monitor.ivresse.top/api", // 默认服务器地址
	WSEndpoint: "ws://localhost:8777/ws/monitor",
}

// 全局数据收集器
var monitorCollector *utils.MonitorDataCollector

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
	cef.BrowserWindow.Config.IconFS = "web/dist/icon.ico"
	// 本地资源配置
	cef.BrowserWindow.Config.Title = "考试客户端系统"
	cef.BrowserWindow.Config.Width = 1700
	cef.BrowserWindow.Config.Height = 1000
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

	// 初始化回调函数
	initReportCallbacks()
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

// 初始化数据上报回调函数
func initReportCallbacks() {
	// 初始化全局回调函数，将网站访问和截图上报连接到数据收集器
	compose.SetReportCallbacks(reportNetworkInfo, reportScreenCap)
}

// 网站访问上报回调
func reportNetworkInfo(domain string) {
	if monitorCollector != nil && monitorCollector.IsRunning {
		fmt.Printf("检测到网站访问: %s，准备上报\n", domain)
		monitorCollector.ReportWebsiteVisit(domain, domain)

		// 通知前端显示
		ipc.Emit("websiteVisit", domain)
	} else {
		fmt.Printf("收到网站访问: %s，但监控收集器未运行\n", domain)
	}
}

// 截图上报回调
func reportScreenCap(buffer *bytes.Buffer) {
	if monitorCollector != nil && monitorCollector.IsRunning {
		fmt.Println("上报屏幕截图")
		monitorCollector.UploadScreenshotData(buffer)
	}
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
		fmt.Println("用户尝试登录:", username)

		// 构建登录请求
		loginURL := appConfig.ServerURL + "/examinee/account/login"
		loginData := map[string]string{
			"account":  username,
			"password": password,
		}

		jsonData, err := json.Marshal(loginData)
		if err != nil {
			fmt.Println("登录数据序列化失败:", err)
			ipc.Emit("loginResult", false, nil, "系统错误: 无法处理登录请求")
			return
		}

		// 发送登录请求
		resp, err := utils.HttpPost(loginURL, jsonData)
		if err != nil {
			fmt.Println("登录请求失败:", err)
			ipc.Emit("loginResult", false, nil, "连接服务器失败: "+err.Error())
			return
		}

		// 解析登录响应
		var loginResp struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
			Data struct {
				Token     string `json:"token"`
				AccountId int    `json:"accountId"`
				ExamId    int    `json:"examId"`
			} `json:"data"`
		}

		if err := json.Unmarshal(resp, &loginResp); err != nil {
			fmt.Println("解析登录响应失败:", err)
			ipc.Emit("loginResult", false, nil, "解析服务器响应失败")
			return
		}

		fmt.Println("登录响应:", loginResp)
		// 检查登录结果
		if loginResp.Code != 0 {
			fmt.Println("登录失败:", loginResp.Msg)
			ipc.Emit("loginResult", false, nil, loginResp.Msg)
			return
		}

		// 保存登录信息
		appConfig.Token = loginResp.Data.Token
		appConfig.AccountID = loginResp.Data.AccountId
		appConfig.ExamID = loginResp.Data.ExamId

		// 获取考生和考试信息
		examInfo, err := getExamineeInfo()
		if err != nil {
			fmt.Println("获取考试信息失败:", err)
			ipc.Emit("loginResult", false, nil, "登录成功但获取考试信息失败: "+err.Error())
			return
		}

		// 创建并启动监控数据收集器
		monitorCollector = utils.NewMonitorDataCollector(
			appConfig.ServerURL,
			appConfig.Token,
			appConfig.AccountID,
			appConfig.ExamID,
		)
		monitorCollector.Start()

		// 启动网络监控
		go compose.WatchNetworkInfo()

		// 启动窗口前台监控
		go compose.MonitorForegroundWindow()

		// 发送登录成功事件
		ipc.Emit("loginResult", true, examInfo, "")
	})

	ipc.On("logout", func() {
		fmt.Println("用户登出")
		// 停止监控数据收集
		if monitorCollector != nil {
			monitorCollector.Stop()
			monitorCollector = nil
		}

		// 清除登录信息
		appConfig.Token = ""
		appConfig.AccountID = 0
		appConfig.ExamID = 0
	})
}

// 获取考生信息和考试信息
func getExamineeInfo() (map[string]interface{}, error) {
	// 构建请求
	infoURL := appConfig.ServerURL + "/examinee/account/info"
	headers := map[string]string{
		"Authorization": "Bearer " + appConfig.Token,
	}

	// 发送请求
	resp, err := utils.HttpGetWithHeaders(infoURL, headers)
	if err != nil {
		return nil, fmt.Errorf("请求考生信息失败: %w", err)
	}

	// 解析响应
	var infoResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Account      map[string]interface{} `json:"account"`
			ExamId       int                    `json:"examId"`
			ExamineeInfo map[string]interface{} `json:"examineeInfo"`
			ExamDetails  struct {
				Id            int    `json:"id"`
				Name          string `json:"name"`
				Description   string `json:"description"`
				StartTime     string `json:"startTime"`
				EndTime       string `json:"endTime"`
				Duration      int    `json:"duration"`
				Location      string `json:"location"`
				Status        int    `json:"status"`
				ServerTime    string `json:"serverTime"`
				RemainingTime int64  `json:"remainingTime"`
			} `json:"examDetails"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &infoResp); err != nil {
		return nil, fmt.Errorf("解析考生信息失败: %w", err)
	}

	if infoResp.Code != 0 {
		return nil, fmt.Errorf("获取考生信息失败: %s", infoResp.Msg)
	}

	// 构造前端需要的考试信息
	result := map[string]interface{}{
		"examId":        infoResp.Data.ExamDetails.Id,
		"title":         infoResp.Data.ExamDetails.Name,
		"description":   infoResp.Data.ExamDetails.Description,
		"startTime":     infoResp.Data.ExamDetails.StartTime,
		"endTime":       infoResp.Data.ExamDetails.EndTime,
		"duration":      infoResp.Data.ExamDetails.Duration,
		"location":      infoResp.Data.ExamDetails.Location,
		"status":        infoResp.Data.ExamDetails.Status,
		"remainingTime": infoResp.Data.ExamDetails.RemainingTime,
		"studentName":   getStringValue(infoResp.Data.ExamineeInfo, "name"),
		"studentId":     getStringValue(infoResp.Data.ExamineeInfo, "studentId"),
		"college":       getStringValue(infoResp.Data.ExamineeInfo, "college"),
		"className":     getStringValue(infoResp.Data.ExamineeInfo, "className"),
	}

	return result, nil
}

// 安全获取字符串值
func getStringValue(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return ""
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
		processInfo := CollectProcessInfo()
		if len(processInfo) > 0 {
			ipc.Emit("processInfo", processInfo)
		}
	}
}

// 收集进程信息
func CollectProcessInfo() []map[string]any {
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
	utils.Go(monitorUSBDevices)
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

// 加载配置
func loadConfig() *Config {
	// 这里可以从配置文件或环境变量加载
	// 此处使用默认配置作为示例
	return &Config{
		ServerURL:  "http://localhost:8777",
		WSEndpoint: "ws://localhost:8777/ws/monitor",
		AccountID:  0, // 这里应当从登录结果中获取
		ExamID:     0, // 这里应当从登录结果中获取
	}
}

// 初始化WebSocket客户端
func initWebSocket(config *Config) *websocket.Conn {
	// 连接WebSocket服务器
	c, _, err := websocket.DefaultDialer.Dial(config.WSEndpoint, nil)
	if err != nil {
		fmt.Printf("连接WebSocket服务器失败: %v\n", err)
		return nil
	}

	// 启动接收消息的goroutine
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				fmt.Printf("读取WebSocket消息失败: %v\n", err)
				return
			}
			fmt.Printf("收到消息: %s\n", message)
		}
	}()

	// 发送连接成功消息
	connectMsg := map[string]interface{}{
		"type":       "CONNECT",
		"message":    "客户端已连接",
		"fromUserId": fmt.Sprintf("%d", config.AccountID),
		"timestamp":  time.Now().UnixNano() / int64(time.Millisecond),
	}

	msgBytes, err := json.Marshal(connectMsg)
	if err != nil {
		fmt.Printf("序列化消息失败: %v\n", err)
	} else {
		if err := c.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
			fmt.Printf("发送连接消息失败: %v\n", err)
		}
	}

	return c
}
