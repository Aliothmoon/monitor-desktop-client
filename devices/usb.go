package devices

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// USBDevice 表示USB设备信息
type USBDevice struct {
	DeviceID      string
	DeviceName    string
	Description   string
	Manufacturer  string
	DeviceType    string
	IsStorageType bool
	IsExternal    bool
	DriveLetters  []string
}

// 定义Windows API常量
const (
	DIGCF_PRESENT         = 0x00000002
	DIGCF_DEVICEINTERFACE = 0x00000010
)

// Setup API 函数
var (
	setupapi                          = syscall.NewLazyDLL("setupapi.dll")
	setupDiGetClassDevsW              = setupapi.NewProc("SetupDiGetClassDevsW")
	setupDiEnumDeviceInfo             = setupapi.NewProc("SetupDiEnumDeviceInfo")
	setupDiGetDeviceRegistryPropertyW = setupapi.NewProc("SetupDiGetDeviceRegistryPropertyW")
	setupDiDestroyDeviceInfoList      = setupapi.NewProc("SetupDiDestroyDeviceInfoList")
)

// GUID 结构体
type _GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// SP_DEVINFO_DATA 结构体
type _SP_DEVINFO_DATA struct {
	cbSize    uint32
	ClassGuid _GUID
	DevInst   uint32
	Reserved  uintptr
}

// GetUSBDevices 获取所有当前连接的外接USB设备信息
func GetUSBDevices() ([]USBDevice, error) {
	var devices []USBDevice

	// 获取当前连接的USB存储设备
	storageDevices, err := getConnectedUsbStorageDevices()
	if err != nil {
		return nil, fmt.Errorf("获取USB存储设备失败: %v", err)
	}
	fmt.Println("storageDevices:", len(storageDevices))
	devices = append(devices, storageDevices...)

	// 获取其他USB外设
	peripheralDevices, err := getConnectedUsbPeripherals()
	if err != nil {
		return nil, fmt.Errorf("获取USB外设失败: %v", err)
	}
	fmt.Println("peripheralDevices:", len(peripheralDevices))
	devices = append(devices, peripheralDevices...)

	return devices, nil
}

// getConnectedUsbStorageDevices 获取当前连接的USB存储设备
func getConnectedUsbStorageDevices() ([]USBDevice, error) {
	var devices []USBDevice

	// 获取活动的可移动驱动器
	drives, err := getActiveRemovableDrives()
	if err != nil {
		return nil, err
	}

	for _, drive := range drives {
		devices = append(devices, USBDevice{
			DeviceID:      drive.deviceID,
			DeviceName:    drive.name,
			Description:   "USB存储设备",
			Manufacturer:  drive.manufacturer,
			DeviceType:    "存储设备",
			IsStorageType: true,
			IsExternal:    true,
			DriveLetters:  []string{drive.letter},
		})
	}

	return devices, nil
}

// Drive 表示驱动器信息
type Drive struct {
	letter       string
	deviceID     string
	name         string
	manufacturer string
}

// getActiveRemovableDrives 获取当前活动的可移动驱动器
func getActiveRemovableDrives() ([]Drive, error) {
	var drives []Drive

	// 获取所有逻辑驱动器
	driveLetters, err := getLogicalDrives()
	if err != nil {
		return nil, err
	}

	for _, letter := range driveLetters {
		// 检查驱动器类型是否为可移动设备
		driveType := windows.GetDriveType(windows.StringToUTF16Ptr(letter + ":\\"))
		if driveType == windows.DRIVE_REMOVABLE {
			// 找到这个驱动器的设备信息
			deviceID, name, manufacturer, err := getDeviceInfoForDrive(letter)
			if err != nil {
				// 如果获取不到详细信息，使用默认值
				deviceID = letter
				name = "可移动存储设备"
				manufacturer = "未知厂商"
			}

			drives = append(drives, Drive{
				letter:       letter,
				deviceID:     deviceID,
				name:         name,
				manufacturer: manufacturer,
			})
		}
	}

	return drives, nil
}

// getDeviceInfoForDrive 获取驱动器对应的设备信息
func getDeviceInfoForDrive(letter string) (deviceID string, name string, manufacturer string, err error) {
	// 获取卷名
	volumeName, err := getVolumeNameForDriveLetter(letter)
	if err != nil {
		return "", "", "", err
	}

	// 从注册表查询设备信息
	deviceID, name, manufacturer = queryDeviceInfoFromRegistry(volumeName)
	if deviceID == "" {
		return volumeName, "可移动存储设备", "未知厂商", nil
	}

	return deviceID, name, manufacturer, nil
}

// getVolumeNameForDriveLetter 获取驱动器对应的卷名
func getVolumeNameForDriveLetter(letter string) (string, error) {
	drivePath := letter + ":\\"
	volumePath := make([]uint16, 256)

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getVolumeNameForVolumeMountPoint := kernel32.NewProc("GetVolumeNameForVolumeMountPointW")

	mountPoint := syscall.StringToUTF16Ptr(drivePath)
	ret, _, err := getVolumeNameForVolumeMountPoint.Call(
		uintptr(unsafe.Pointer(mountPoint)),
		uintptr(unsafe.Pointer(&volumePath[0])),
		uintptr(len(volumePath)),
	)

	if ret == 0 {
		return "", err
	}

	return windows.UTF16ToString(volumePath[:]), nil
}

// queryDeviceInfoFromRegistry 从注册表查询设备信息
func queryDeviceInfoFromRegistry(volumeName string) (deviceID string, name string, manufacturer string) {
	// 尝试查找存储设备的注册表信息
	diskKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services\disk\Enum`, registry.READ)
	if err != nil {
		return "", "", ""
	}
	defer diskKey.Close()

	// 读取所有设备编号
	count, _, _ := diskKey.GetIntegerValue("Count")
	for i := 0; i < int(count); i++ {
		key := fmt.Sprintf("%d", i)
		value, _, err := diskKey.GetStringValue(key)
		if err != nil {
			continue
		}

		// 如果找到对应的卷设备
		if strings.Contains(value, volumeName) {
			deviceID = value

			// 尝试从对应设备ID获取厂商和名称
			devPath := fmt.Sprintf(`SYSTEM\CurrentControlSet\Enum\%s`, value)
			devKey, err := registry.OpenKey(registry.LOCAL_MACHINE, devPath, registry.READ)
			if err == nil {
				defer devKey.Close()
				name, _, _ = devKey.GetStringValue("FriendlyName")
				if name == "" {
					name, _, _ = devKey.GetStringValue("DeviceDesc")
				}
				manufacturer, _, _ = devKey.GetStringValue("Mfg")
			}

			return deviceID, name, manufacturer
		}
	}

	return "", "", ""
}

// getConnectedUsbPeripherals 获取当前连接的USB外设
func getConnectedUsbPeripherals() ([]USBDevice, error) {
	var devices []USBDevice

	// 使用SetupAPI直接查询当前连接的USB设备
	hDevInfo, err := getUsbDeviceInfoSet()
	if err != nil {
		// 如果SetupAPI失败，尝试从注册表获取
		return getConnectedUsbDevicesFromRegistry()
	}
	defer setupDiDestroyDeviceInfoList.Call(hDevInfo)

	// 枚举设备
	var index uint32 = 0
	for {
		var devInfo _SP_DEVINFO_DATA
		devInfo.cbSize = uint32(unsafe.Sizeof(devInfo))

		ret, _, _ := setupDiEnumDeviceInfo.Call(
			hDevInfo,
			uintptr(index),
			uintptr(unsafe.Pointer(&devInfo)),
		)

		// 如果没有更多设备，退出循环
		if ret == 0 {
			break
		}

		// 获取设备信息
		deviceID := getDeviceID(hDevInfo, &devInfo)
		description := getDeviceProperty(hDevInfo, &devInfo, 0x00000000)  // SPDRP_DEVICEDESC
		manufacturer := getDeviceProperty(hDevInfo, &devInfo, 0x0000000B) // SPDRP_MFG

		// 判断设备类型
		deviceType := "外设"
		if strings.Contains(strings.ToLower(description), "keyboard") {
			deviceType = "键盘"
		} else if strings.Contains(strings.ToLower(description), "mouse") {
			deviceType = "鼠标"
		} else if strings.Contains(strings.ToLower(description), "camera") ||
			strings.Contains(strings.ToLower(description), "webcam") {
			deviceType = "摄像头"
		} else if strings.Contains(strings.ToLower(description), "mass storage") ||
			strings.Contains(strings.ToLower(description), "disk drive") {
			deviceType = "存储设备"
		}

		// 添加到设备列表
		device := USBDevice{
			DeviceID:      deviceID,
			DeviceName:    description,
			Description:   description,
			Manufacturer:  manufacturer,
			DeviceType:    deviceType,
			IsExternal:    true,
			IsStorageType: deviceType == "存储设备",
		}

		devices = append(devices, device)
		index++
	}

	return devices, nil
}

// getUsbDeviceInfoSet 获取USB设备集合句柄
func getUsbDeviceInfoSet() (uintptr, error) {
	// USB类GUID
	guid := _GUID{
		Data1: 0x36FC9E60,
		Data2: 0xC465,
		Data3: 0x11CF,
		Data4: [8]byte{0x80, 0x56, 0x44, 0x45, 0x53, 0x54, 0x00, 0x00},
	}

	// 获取当前连接的USB设备集合
	handle, _, err := setupDiGetClassDevsW.Call(
		uintptr(unsafe.Pointer(&guid)),
		0,
		0,
		uintptr(DIGCF_PRESENT),
	)

	if handle == 0 {
		return 0, err
	}

	return handle, nil
}

// getDeviceID 获取设备ID
func getDeviceID(hDevInfo uintptr, devInfo *_SP_DEVINFO_DATA) string {
	instanceID := getDeviceProperty(hDevInfo, devInfo, 0x00000001) // SPDRP_HARDWAREID
	return instanceID
}

// getDeviceProperty 获取设备属性
func getDeviceProperty(hDevInfo uintptr, devInfo *_SP_DEVINFO_DATA, property uint32) string {
	var dataType, bufferSize uint32

	// 第一次调用获取需要的缓冲区大小
	setupDiGetDeviceRegistryPropertyW.Call(
		hDevInfo,
		uintptr(unsafe.Pointer(devInfo)),
		uintptr(property),
		uintptr(unsafe.Pointer(&dataType)),
		0,
		0,
		uintptr(unsafe.Pointer(&bufferSize)),
	)

	// 如果不需要缓冲区，表示没有该属性
	if bufferSize == 0 {
		return ""
	}

	// 分配缓冲区
	buffer := make([]uint16, bufferSize/2)

	// 第二次调用获取实际数据
	ret, _, _ := setupDiGetDeviceRegistryPropertyW.Call(
		hDevInfo,
		uintptr(unsafe.Pointer(devInfo)),
		uintptr(property),
		uintptr(unsafe.Pointer(&dataType)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(&bufferSize)),
	)

	if ret == 0 {
		return ""
	}

	return windows.UTF16ToString(buffer)
}

// getConnectedUsbDevicesFromRegistry 从注册表获取当前连接的USB设备
func getConnectedUsbDevicesFromRegistry() ([]USBDevice, error) {
	var devices []USBDevice

	// 打开USB枚举注册表
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Enum\USB`, registry.READ)
	if err != nil {
		return nil, fmt.Errorf("打开USB注册表失败: %v", err)
	}
	defer key.Close()

	// 读取所有USB设备ID
	deviceIDs, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return nil, fmt.Errorf("读取USB设备ID失败: %v", err)
	}
	for _, deviceID := range deviceIDs {
		// 打开设备子键
		deviceKey, err := registry.OpenKey(key, deviceID, registry.READ)
		if err != nil {
			continue
		}
		defer deviceKey.Close()

		// 读取子设备
		subDevices, err := deviceKey.ReadSubKeyNames(-1)
		if err != nil {
			continue
		}

		for _, subDevice := range subDevices[:1] {
			subKey, err := registry.OpenKey(deviceKey, subDevice, registry.READ)
			if err != nil {
				continue
			}
			defer subKey.Close()

			// 关键判断：检查设备是否真的连接
			// 查看设备是否有一个"Device Parameters"子键，表示设备连接
			if isDeviceConnected(subKey) {
				// 读取设备信息
				deviceDesc, _, _ := subKey.GetStringValue("DeviceDesc")
				manufacturer, _, _ := subKey.GetStringValue("Mfg")
				friendlyName, _, _ := subKey.GetStringValue("FriendlyName")

				if friendlyName != "" {
					deviceDesc = friendlyName
				}

				// 判断设备类型
				deviceType := "外设"
				if strings.Contains(strings.ToLower(deviceDesc), "mass storage") ||
					strings.Contains(strings.ToLower(deviceDesc), "disk drive") {
					deviceType = "存储设备"
				} else if strings.Contains(strings.ToLower(deviceDesc), "keyboard") {
					deviceType = "键盘"
				} else if strings.Contains(strings.ToLower(deviceDesc), "mouse") {
					deviceType = "鼠标"
				} else if strings.Contains(strings.ToLower(deviceDesc), "camera") ||
					strings.Contains(strings.ToLower(deviceDesc), "webcam") {
					deviceType = "摄像头"
				}

				// 添加设备
				device := USBDevice{
					DeviceID:      deviceID + "\\" + subDevice,
					DeviceName:    deviceDesc,
					Description:   deviceDesc,
					Manufacturer:  manufacturer,
					DeviceType:    deviceType,
					IsExternal:    true,
					IsStorageType: deviceType == "存储设备",
				}

				devices = append(devices, device)
			}
		}
	}

	return devices, nil
}

// isDeviceConnected 检查设备是否真的连接着
func isDeviceConnected(deviceKey registry.Key) bool {
	// 检查方法1: 检查是否有设备参数子键
	_, err := registry.OpenKey(deviceKey, "Device Parameters", registry.READ)
	if err == nil {
		return true
	}

	// 检查方法2: 检查服务状态
	_, err = registry.OpenKey(deviceKey, "Control", registry.READ)
	if err == nil {
		return true
	}

	// 检查方法3: 检查设备状态
	// ConfigFlags为4表示设备被禁用，0表示正常
	configFlags, _, err := deviceKey.GetIntegerValue("ConfigFlags")
	if err == nil && configFlags == 0 {
		return true
	}

	// 检查标记为已移除的设备
	removed, _, err := deviceKey.GetIntegerValue("Removed")
	if err == nil && removed != 0 {
		return false
	}

	// 检查状态标志
	status, _, err := deviceKey.GetStringValue("Status")
	if err == nil && status != "" {
		// 状态字符串为空或包含错误表示设备不可用
		if strings.Contains(strings.ToLower(status), "error") {
			return false
		}
		return true
	}

	// 无法确定，保守起见返回false
	return false
}

// getLogicalDrives 获取所有逻辑驱动器
func getLogicalDrives() ([]string, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getLogicalDrives := kernel32.NewProc("GetLogicalDrives")

	drives, _, _ := getLogicalDrives.Call()
	var driveLetters []string

	for i := 0; i < 26; i++ {
		mask := 1 << uint(i)
		if int(drives)&mask != 0 {
			driveLetters = append(driveLetters, string('A'+i))
		}
	}

	return driveLetters, nil
}

// PrintDeviceInfo 打印设备信息
func PrintDeviceInfo(device USBDevice) string {
	var info strings.Builder
	info.WriteString(fmt.Sprintf("设备名称: %s\n", device.DeviceName))
	info.WriteString(fmt.Sprintf("设备类型: %s\n", device.DeviceType))
	info.WriteString(fmt.Sprintf("厂商: %s\n", device.Manufacturer))

	if device.IsStorageType {
		info.WriteString(fmt.Sprintf("存储设备: 是\n"))
		info.WriteString(fmt.Sprintf("盘符: %s\n", strings.Join(device.DriveLetters, ", ")))
	} else {
		info.WriteString(fmt.Sprintf("存储设备: 否\n"))
	}

	return info.String()
}
