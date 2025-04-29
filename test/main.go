package main

import (
	"fmt"
	"monitor-desktop-client/devices"
	"monitor-desktop-client/foreground"
	"syscall"
	"time"
)

func main() {
	fmt.Println("程序启动，测试窗口监控功能...")

	// 窗口监控测试
	MonitorForegroundWindow()

	// 硬件信息可选测试
	// GetHardwareInfo()
}

// GetHardwareInfo 获取设备硬件信息
func GetHardwareInfo() {
	deviceInfo, err := devices.GetDeviceInfo()
	if err != nil {
		fmt.Println("获取设备信息失败:", err)
		return
	}

	// 打印格式化的设备信息到控制台
	fmt.Println("设备信息:")
	fmt.Println(devices.FormatDeviceInfo(deviceInfo))
}

// MonitorForegroundWindow 监控前台窗口变化
func MonitorForegroundWindow() {
	var prev syscall.Handle = 0

	for {
		info := foreground.GetWindowInfo()
		if info == nil {
			continue
		}
		if info.Handle != prev {
			fmt.Printf("焦点切换 -> 进程: %-20s PID: %-6d 窗口: %-50s 路径: %s\n", info.ProcessName, info.ProcessID, info.Title, info.ProcessPath)
			prev = info.Handle
		}

		time.Sleep(500 * time.Millisecond)
	}
}
