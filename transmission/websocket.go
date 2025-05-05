package wsc

import (
	"encoding/json"
	"fmt"
	"log"
	"monitor-desktop-client/utils"
	"os"
	"time"
)

// WebSocketMessage 定义与后端一致的消息结构
type WebSocketMessage struct {
	Type         string      `json:"type"`
	Message      string      `json:"message"`
	TargetUserId string      `json:"targetUserId,omitempty"`
	FromUserId   string      `json:"fromUserId,omitempty"`
	Data         interface{} `json:"data,omitempty"`
	Timestamp    int64       `json:"timestamp"`
}

// WebSocketClient 管理WebSocket连接的结构体
type WebSocketClient struct {
	Conn         *Wsc
	UserId       string
	IsConnected  bool
	Handlers     map[string]func(message WebSocketMessage)
	OnConnected  func()
	OnDisconnect func()
}

var (
	// 全局WebSocket客户端实例
	wsClient *WebSocketClient
	// 服务器地址
	serverUrl string
	// 心跳定时器
	heartbeatTicker *time.Ticker
)

// SetupWebsocket 设置并连接到后端WebSocket服务
func SetupWebsocket(serverAddress string, userId string) *WebSocketClient {
	// 如果已经有连接，先关闭
	if wsClient != nil && wsClient.IsConnected {
		if wsClient.Conn != nil {
			// 使用CloseErr变量来关闭连接
			wsClient.Conn.Close()
		}
	}

	// 构建WebSocket URL
	serverUrl = fmt.Sprintf("ws://%s/ws/monitor/%s", serverAddress, userId)

	// 初始化客户端
	wsClient = &WebSocketClient{
		UserId:      userId,
		IsConnected: false,
		Handlers:    make(map[string]func(message WebSocketMessage)),
	}

	// 创建WebSocket连接
	ws := New(serverUrl)

	// 配置WebSocket
	ws.SetConfig(&Config{
		WriteWait:         10 * time.Second,
		MaxMessageSize:    4096,
		MinRecTime:        2 * time.Second,
		MaxRecTime:        60 * time.Second,
		RecFactor:         1.5,
		MessageBufferSize: 1024,
	})

	// 保存连接
	wsClient.Conn = ws

	// 设置连接成功回调
	ws.OnConnected(func() {
		log.Printf("WebSocket连接成功: %s", serverUrl)
		wsClient.IsConnected = true

		// 启动心跳
		startHeartbeat()

		// 如果有设置连接成功回调，则执行
		if wsClient.OnConnected != nil {
			wsClient.OnConnected()
		}
	})

	// 设置连接错误回调
	ws.OnConnectError(func(err error) {
		log.Printf("WebSocket连接错误: %s", err.Error())
		wsClient.IsConnected = false
	})

	// 设置断开连接回调
	ws.OnDisconnected(func(err error) {
		log.Printf("WebSocket断开连接: %s", err.Error())
		wsClient.IsConnected = false

		// 停止心跳
		stopHeartbeat()

		// 如果有设置断开连接回调，则执行
		if wsClient.OnDisconnect != nil {
			wsClient.OnDisconnect()
		}
	})

	// 设置关闭回调
	ws.OnClose(func(code int, text string) {
		log.Printf("WebSocket关闭: %d %s", code, text)
		wsClient.IsConnected = false

		// 停止心跳
		stopHeartbeat()
	})

	// 设置接收文本消息回调
	ws.OnTextMessageReceived(func(message string) {
		log.Printf("收到WebSocket消息: %s", message)

		// 解析消息
		var wsMessage WebSocketMessage
		err := json.Unmarshal([]byte(message), &wsMessage)
		if err != nil {
			log.Printf("解析WebSocket消息失败: %s", err.Error())
			return
		}

		// 根据消息类型调用对应的处理函数
		if handler, ok := wsClient.Handlers[wsMessage.Type]; ok {
			handler(wsMessage)
		} else {
			log.Printf("未处理的消息类型: %s", wsMessage.Type)
		}
	})

	// 设置发送文本消息回调
	ws.OnTextMessageSent(func(message string) {
		log.Printf("发送WebSocket消息: %s", message)
	})

	// 设置发送错误回调
	ws.OnSentError(func(err error) {
		log.Printf("发送WebSocket消息错误: %s", err.Error())
	})

	// 开始连接
	utils.Go(func() {
		ws.Connect()
	})

	// 注册默认消息处理器
	registerDefaultHandlers()

	return wsClient
}

// 启动心跳
func startHeartbeat() {
	// 先停止现有的心跳
	stopHeartbeat()

	// 创建新的心跳定时器，每30秒发送一次
	heartbeatTicker = time.NewTicker(30 * time.Second)

	utils.Go(func() {
		for {
			select {
			case <-heartbeatTicker.C:
				if wsClient != nil && wsClient.IsConnected {
					sendHeartbeat()
				} else {
					// 如果连接已断开，停止心跳
					stopHeartbeat()
					return
				}
			}
		}
	})
}

// 停止心跳
func stopHeartbeat() {
	if heartbeatTicker != nil {
		heartbeatTicker.Stop()
		heartbeatTicker = nil
	}
}

// 发送心跳消息
func sendHeartbeat() {
	heartbeat := WebSocketMessage{
		Type:       "HEARTBEAT",
		Message:    "ping",
		FromUserId: wsClient.UserId,
		Timestamp:  time.Now().UnixMilli(),
	}

	wsClient.SendMessage(heartbeat)
}

// 注册默认消息处理器
func registerDefaultHandlers() {
	// 处理连接成功消息
	wsClient.RegisterHandler("CONNECT", func(message WebSocketMessage) {
		log.Printf("连接成功: %s", message.Message)
	})

	// 处理系统消息
	wsClient.RegisterHandler("SYSTEM", func(message WebSocketMessage) {
		log.Printf("系统消息: %s", message.Message)

		// 如果数据中包含在线人数，可以在这里处理
		if count, ok := message.Data.(float64); ok {
			log.Printf("当前在线人数: %d", int(count))
		}
	})

	// 处理心跳响应
	wsClient.RegisterHandler("HEARTBEAT", func(message WebSocketMessage) {
		log.Printf("心跳响应: %s", message.Message)
	})

	// 处理通知消息
	wsClient.RegisterHandler("NOTIFICATION", func(message WebSocketMessage) {
		log.Printf("收到通知: %s", message.Message)
	})
}

// RegisterHandler 注册消息处理器
func (c *WebSocketClient) RegisterHandler(messageType string, handler func(message WebSocketMessage)) {
	c.Handlers[messageType] = handler
}

// SendMessage 发送消息
func (c *WebSocketClient) SendMessage(message WebSocketMessage) error {
	if !c.IsConnected {
		return fmt.Errorf("WebSocket未连接")
	}

	// 序列化消息
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 发送消息
	return c.Conn.SendTextMessage(string(messageBytes))
}

// SendTextToServer 发送文本消息给服务器
func (c *WebSocketClient) SendTextToServer(messageType string, text string) error {
	message := WebSocketMessage{
		Type:       messageType,
		Message:    text,
		FromUserId: c.UserId,
		Timestamp:  time.Now().UnixMilli(),
	}

	return c.SendMessage(message)
}

// SendBroadcast 发送广播消息
func (c *WebSocketClient) SendBroadcast(text string) error {
	return c.SendTextToServer("BROADCAST", text)
}

// SendPrivateMessage 发送私聊消息
func (c *WebSocketClient) SendPrivateMessage(targetUserId string, text string) error {
	message := WebSocketMessage{
		Type:         "PRIVATE",
		Message:      text,
		FromUserId:   c.UserId,
		TargetUserId: targetUserId,
		Timestamp:    time.Now().UnixMilli(),
	}

	return c.SendMessage(message)
}

// IsClientConnected 检查WebSocket是否已连接
func IsClientConnected() bool {
	return wsClient != nil && wsClient.IsConnected
}

// DisconnectWebsocket 断开WebSocket连接
func DisconnectWebsocket() {
	if wsClient != nil && wsClient.IsConnected && wsClient.Conn != nil {
		wsClient.Conn.Close()
		wsClient.IsConnected = false
		stopHeartbeat()
	}
}

// GetWebSocketClient 获取WebSocket客户端实例
func GetWebSocketClient() *WebSocketClient {
	return wsClient
}

// InitWebsocketFromEnv 从环境变量初始化WebSocket连接
func InitWebsocketFromEnv() *WebSocketClient {
	// 从环境变量或配置文件获取服务器地址和用户ID
	serverAddress := os.Getenv("WS_SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = "localhost:8080" // 默认地址
	}

	userId := os.Getenv("WS_USER_ID")
	if userId == "" {
		// 使用机器标识或随机生成的ID作为用户ID
		// 这里简化为使用主机名和时间戳的组合
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown-host"
		}
		userId = fmt.Sprintf("%s-%d", hostname, time.Now().Unix())
	}

	return SetupWebsocket(serverAddress, userId)
}
