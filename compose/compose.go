package compose

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"log"
	"monitor-desktop-client/devices"
	"monitor-desktop-client/foreground"
	"monitor-desktop-client/netcap"
	"monitor-desktop-client/screencap"
	"monitor-desktop-client/utils"
	"strings"
	"syscall"
	"time"

	"github.com/google/gopacket/pcap"
)

// 回调函数类型定义
type NetworkInfoCallback func(domain string)
type ScreenCapCallback func(buffer *bytes.Buffer)

// 全局回调函数
var networkInfoCallback NetworkInfoCallback
var screenCapCallback ScreenCapCallback

// SetReportCallbacks 设置回调函数
func SetReportCallbacks(netCallback NetworkInfoCallback, screenCallback ScreenCapCallback) {
	networkInfoCallback = netCallback
	screenCapCallback = screenCallback
}

// ReportNetworkInfo 上报网络访问信息
func ReportNetworkInfo(domain string) {
	if networkInfoCallback != nil {
		networkInfoCallback(domain)
	}
}

// ReportScreenCap 上报屏幕截图
func ReportScreenCap(buffer *bytes.Buffer) {
	if screenCapCallback != nil {
		screenCapCallback(buffer)
	}
}

func GetUsbDeviceInfo() ([]devices.USBDevice, error) {
	ds, err := devices.GetUSBDevices()
	return ds, err
}

func WatchNetworkInfo() {

	ifs, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
		return
	}
	var ds []string

	log.Println("Watching network info...")

	for _, iface := range ifs {
		if !strings.Contains(strings.ToLower(iface.Name), "loopback") &&
			!devices.IsVirtualInterface(iface.Name) &&
			!devices.IsVirtualInterface(iface.Description) &&
			len(iface.Addresses) > 0 {
			log.Println(iface.Description, iface.Name)
			ds = append(ds, iface.Name)
		}
	}
	for _, d := range ds {
		utils.Go(func() {
			live := netcap.OpenLive(d)
			if live != nil {
				log.Printf("开始监控网络设备: %s", d)
				for domain := range live.Ch {
					log.Printf("检测到域名访问: %s (设备: %s)", domain, d)
					ReportNetworkInfo(domain)
				}
			}
		})
	}

}

// GetHardwareInfo 获取设备硬件信息
func GetHardwareInfo() {
	deviceInfo, err := devices.GetDeviceInfo()
	if err != nil {
		log.Println("GetHardwareInfo", err)
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
			log.Printf("焦点切换 -> 进程: %-20s PID: %-6d 窗口: %-50s 路径: %s\n", info.ProcessName, info.ProcessID, info.Title, info.ProcessPath)
			throttler.Do(func() {
				img, err := screencap.ScreenCap()
				if err != nil {
					log.Println(err)
					return
				}
				buffer := bytes.NewBuffer(nil)
				err = jpeg.Encode(buffer, img, nil)
				if err != nil {
					log.Println(err)
					return
				}
				ReportScreenCap(buffer)
			})
			prev = info.Handle
		}

		time.Sleep(500 * time.Millisecond)
	}
}
