package wsc

import (
	"log"
	"monitor-desktop-client/utils"
	"time"
)

// ExampleWebSocket 演示如何使用WebSocket客户端
func ExampleWebSocket() {
	// 创建一个通道，用于等待程序结束
	done := make(chan bool)

	// 设置WebSocket连接
	client := SetupWebsocket("localhost:8080", "example-client-001")

	// 设置自定义消息处理器
	client.RegisterHandler("NOTIFICATION", func(message WebSocketMessage) {
		log.Printf("收到自定义通知: %s", message.Message)
		if data, ok := message.Data.(map[string]interface{}); ok {
			log.Printf("通知数据: %v", data)
		}
	})

	// 设置连接成功回调
	client.OnConnected = func() {
		log.Println("连接成功回调: 开始发送测试消息")

		// 连接成功后，每5秒发送一条测试消息
		utils.Go(func() {
			ticker := time.NewTicker(5 * time.Second)
			counter := 0

			for {
				select {
				case <-ticker.C:
					counter++
					// 发送广播消息
					err := client.SendBroadcast("这是一条测试广播消息 #" + string(counter))
					if err != nil {
						log.Printf("发送广播消息失败: %s", err.Error())
					}

					// 每10秒发送一次私聊消息
					if counter%2 == 0 {
						err := client.SendPrivateMessage("admin", "这是一条发送给管理员的私聊消息")
						if err != nil {
							log.Printf("发送私聊消息失败: %s", err.Error())
						}
					}

					// 20秒后断开连接
					if counter >= 4 {
						log.Println("测试完成，准备断开连接")
						DisconnectWebsocket()
						done <- true
						return
					}
				}
			}
		})
	}

	// 设置断开连接回调
	client.OnDisconnect = func() {
		log.Println("连接已断开")
	}

	// 等待程序结束
	<-done
}

// 向前端发送用户状态更新
func SendUserStatusUpdate(userId string, isOnline bool, deviceInfo map[string]interface{}) error {
	client := GetWebSocketClient()
	if client == nil {
		return nil // 客户端未初始化，无需发送
	}

	// 构建状态更新消息
	message := WebSocketMessage{
		Type:       "STATUS_UPDATE",
		Message:    "",
		FromUserId: client.UserId,
		Timestamp:  time.Now().UnixMilli(),
	}

	// 设置状态数据
	statusData := map[string]interface{}{
		"userId":   userId,
		"isOnline": isOnline,
		"device":   deviceInfo,
		"time":     time.Now().Format(time.RFC3339),
	}
	message.Data = statusData

	// 发送消息
	return client.SendMessage(message)
}

// 向服务器报告硬件活动情况
func ReportHardwareActivity(cpuUsage float64, memoryUsage float64, networkActivity map[string]interface{}) error {
	client := GetWebSocketClient()
	if client == nil || !client.IsConnected {
		return nil // 客户端未初始化或未连接，无需发送
	}

	// 构建硬件监控消息
	message := WebSocketMessage{
		Type:       "HARDWARE_STATS",
		Message:    "硬件活动数据",
		FromUserId: client.UserId,
		Timestamp:  time.Now().UnixMilli(),
	}

	// 设置硬件数据
	hardwareData := map[string]interface{}{
		"cpuUsage":        cpuUsage,
		"memoryUsage":     memoryUsage,
		"networkActivity": networkActivity,
		"timestamp":       time.Now().UnixMilli(),
	}
	message.Data = hardwareData

	// 发送消息
	return client.SendMessage(message)
}

// 报告考试客户端状态
func ReportExamClientStatus(examId string, status string, details map[string]interface{}) error {
	client := GetWebSocketClient()
	if client == nil || !client.IsConnected {
		return nil
	}

	// 构建考试状态消息
	message := WebSocketMessage{
		Type:       "EXAM_STATUS",
		Message:    status,
		FromUserId: client.UserId,
		Timestamp:  time.Now().UnixMilli(),
	}

	// 设置状态数据
	statusData := map[string]interface{}{
		"examId":  examId,
		"status":  status,
		"details": details,
		"time":    time.Now().Format(time.RFC3339),
	}
	message.Data = statusData

	// 发送消息
	return client.SendMessage(message)
}
