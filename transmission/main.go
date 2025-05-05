package wsc

import (
	"errors"
	"log"
	"monitor-desktop-client/utils"
	"time"
)

// StartWebSocketMonitor  可导出的启动WebSocket监控的函数
func StartWebSocketMonitor(serverAddress string, userId string) {
	log.Printf("正在启动WebSocket监控，连接到 %s，用户ID: %s", serverAddress, userId)

	// 设置WebSocket连接
	client := SetupWebsocket(serverAddress, userId)

	// 设置连接成功回调
	client.OnConnected = func() {
		log.Printf("WebSocket连接成功，正在监听服务器消息...")

		// 发送连接成功通知
		deviceInfo := map[string]interface{}{
			"os":         "Windows", // 这里可以改为动态获取操作系统信息
			"clientType": "ExamClient",
			"version":    "1.0.0",
		}

		// 报告客户端状态
		err := SendUserStatusUpdate(userId, true, deviceInfo)
		if err != nil {
			log.Printf("发送用户状态更新失败: %s", err.Error())
		}
	}

	// 设置断开连接回调
	client.OnDisconnect = func() {
		log.Printf("WebSocket连接已断开，将在稍后尝试重新连接...")
	}

	// 注册自定义消息处理器
	registerMonitorHandlers(client)
}

// 注册监控相关的消息处理器
func registerMonitorHandlers(client *WebSocketClient) {
	// 处理服务器端命令
	client.RegisterHandler("COMMAND", func(message WebSocketMessage) {
		log.Printf("收到服务器命令: %s", message.Message)

		// 解析命令内容
		if data, ok := message.Data.(map[string]interface{}); ok {
			if cmd, ok := data["command"].(string); ok {
				executeCommand(cmd, data)
			}
		}
	})

	// 处理服务器请求状态更新
	client.RegisterHandler("REQUEST_STATUS", func(message WebSocketMessage) {
		log.Printf("收到状态更新请求")

		// 收集并发送当前状态
		cpuUsage := 30.5    // 示例值，实际应从系统获取
		memoryUsage := 45.2 // 示例值，实际应从系统获取

		// 模拟网络活动数据
		networkActivity := map[string]interface{}{
			"bytesReceived": 1024,
			"bytesSent":     512,
			"connections":   2,
		}

		// 报告硬件活动
		err := ReportHardwareActivity(cpuUsage, memoryUsage, networkActivity)
		if err != nil {
			log.Printf("发送硬件活动数据失败: %s", err.Error())
		}
	})
}

// 执行服务器下发的命令
func executeCommand(command string, params map[string]interface{}) {
	log.Printf("执行命令: %s, 参数: %v", command, params)

	// 根据命令类型执行不同操作
	switch command {
	case "RESTART_CLIENT":
		log.Println("收到重启客户端命令，准备重启...")
		// 实际重启逻辑...

	case "LOCK_SCREEN":
		log.Println("收到锁定屏幕命令，准备锁定...")
		// 实际锁定逻辑...

	case "TAKE_SCREENSHOT":
		log.Println("收到截图命令，准备截图...")
		// 实际截图逻辑...

	case "UPDATE_CONFIG":
		log.Println("收到更新配置命令，准备更新配置...")
		// 实际更新配置逻辑...

	default:
		log.Printf("未知命令: %s", command)
	}
}

func TestWs() {
	done := make(chan bool)
	ws := New("ws://127.0.0.1:7777/ws")
	// 可自定义配置，不使用默认配置
	//ws.SetConfig(&wsc.Config{
	//	// 写超时
	//	WriteWait: 10 * time.Second,
	//	// 支持接受的消息最大长度，默认512字节
	//	MaxMessageSize: 2048,
	//	// 最小重连时间间隔
	//	MinRecTime: 2 * time.Second,
	//	// 最大重连时间间隔
	//	MaxRecTime: 60 * time.Second,
	//	// 每次重连失败继续重连的时间间隔递增的乘数因子，递增到最大重连时间间隔为止
	//	RecFactor: 1.5,
	//	// 消息发送缓冲池大小，默认256
	//	MessageBufferSize: 1024,
	//})
	// 设置回调处理
	ws.OnConnected(func() {
		log.Println("OnConnected: 连接成功")
		// 连接成功后，测试每5秒发送消息
		utils.Go(func() {
			t := time.NewTicker(5 * time.Second)
			for {
				select {
				case <-t.C:
					err := ws.SendTextMessage("hello")
					if errors.Is(err, CloseErr) {
						return
					}
				}
			}
		})
	})
	ws.OnConnectError(func(err error) {
		log.Println("OnConnectError: ", err.Error())
	})
	ws.OnDisconnected(func(err error) {
		log.Println("OnDisconnected: ", err.Error())
	})
	ws.OnClose(func(code int, text string) {
		log.Println("OnClose: ", code, text)
		done <- true
	})
	ws.OnTextMessageSent(func(message string) {
		log.Println("OnTextMessageSent: ", message)
	})
	ws.OnBinaryMessageSent(func(data []byte) {
		log.Println("OnBinaryMessageSent: ", string(data))
	})
	ws.OnSentError(func(err error) {
		log.Println("OnSentError: ", err.Error())
	})
	ws.OnPingReceived(func(appData string) {
		log.Println("OnPingReceived: ", appData)
	})
	ws.OnPongReceived(func(appData string) {
		log.Println("OnPongReceived: ", appData)
	})
	ws.OnTextMessageReceived(func(message string) {
		log.Println("OnTextMessageReceived: ", message)
	})
	ws.OnBinaryMessageReceived(func(data []byte) {
		log.Println("OnBinaryMessageReceived: ", string(data))
	})
	// 开始连接
	utils.Go(func() {
		ws.Connect()
	})
	for {
		select {
		case <-done:
			return
		}
	}
}
