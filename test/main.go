package main

import (
	"fmt"
	"github.com/google/gopacket/pcap"
	"log"
	"monitor-desktop-client/devices"
	"monitor-desktop-client/ffmpeg"
	"monitor-desktop-client/foreground"
	"monitor-desktop-client/netcap"
	"monitor-desktop-client/screencap"
	"monitor-desktop-client/utils"
	"strings"
	"syscall"
	"time"
)

func main() {
	//fmt.Println("程序启动，测试窗口监控功能...")

	UnPackFfmpeg()

	//// 硬件信息可选测试
	//GetHardwareInfo()
	//
	//WatchNetworkInfo()
	//
	//// 窗口监控测试
	//MonitorForegroundWindow()

}

func UnPackFfmpeg() {
	err := ffmpeg.UnPack()
	if err != nil {
		log.Println(err)
	}
	err = ffmpeg.Version()
	if err != nil {
		log.Println(err)
	}
	go func() {
		err := ffmpeg.RtmpPushScreen("rtmp://localhost:1935/live/test")
		if err != nil {
			log.Println(err)
		}
	}()
}

func WatchNetworkInfo() {

	ifs, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
		return
	}
	var ds []string

	for _, iface := range ifs {
		if !strings.Contains(strings.ToLower(iface.Name), "loopback") &&
			!devices.IsVirtualInterface(iface.Name) &&
			!devices.IsVirtualInterface(iface.Description) &&
			len(iface.Addresses) > 0 {
			fmt.Println(iface.Description)
			ds = append(ds, iface.Name)
		}
	}
	for _, d := range ds {
		utils.Go(func() {
			netcap.OpenLive(d)
		})
	}

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

	throttler := utils.NewAdvancedThrottler(time.Second * 3)
	for {
		info := foreground.GetWindowInfo()
		if info == nil {
			continue
		}
		if info.Handle != prev {
			fmt.Printf("焦点切换 -> 进程: %-20s PID: %-6d 窗口: %-50s 路径: %s\n", info.ProcessName, info.ProcessID, info.Title, info.ProcessPath)
			throttler.Do(func() {
				screencap.SaveTestCap()
			})
			prev = info.Handle
		}

		time.Sleep(500 * time.Millisecond)
	}
}

type ctrl struct {
	done chan struct{}
}

func test() {
	c := ctrl{make(chan struct{})}
	go func() {
		time.After(time.Second * 10)
		c.done <- struct{}{}
	}()

	select {
	case <-c.done:
		fmt.Println("done")
	case <-time.After(time.Second * 5):
		fmt.Println("timeout")
	}

}
