package devices

import (
	"fmt"
	"net"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// DeviceInfo 包含设备软硬件信息
type DeviceInfo struct {
	// 系统信息
	Hostname       string `json:"hostname"`
	OS             string `json:"os"`
	Platform       string `json:"platform"`
	PlatformFamily string `json:"platformFamily"`
	PlatformVer    string `json:"platformVer"`
	KernelVer      string `json:"kernelVer"`
	KernelArch     string `json:"kernelArch"`

	// 设备标识
	UniqueID   string `json:"uniqueId"`
	MachineID  string `json:"machineId"`
	BIOSUUID   string `json:"biosUuid"`
	ProductID  string `json:"productId"`
	HardwareID string `json:"hardwareId"`

	// CPU信息
	CPUModel     string  `json:"cpuModel"`
	CPUCores     int     `json:"cpuCores"`
	CPUFrequency float64 `json:"cpuFrequency"`
	CPUUsage     float64 `json:"cpuUsage"`

	// 内存信息
	MemTotal       uint64  `json:"memTotal"`
	MemAvailable   uint64  `json:"memAvailable"`
	MemUsed        uint64  `json:"memUsed"`
	MemUsedPercent float64 `json:"memUsedPercent"`

	// 磁盘信息
	DiskInfo []DiskInfo `json:"diskInfo"`

	// 网络信息
	NetworkInfo []NetworkInfo `json:"networkInfo"`

	// 显卡信息
	GPUInfo []GPUInfo `json:"gpuInfo"`

	// BIOS信息
	BIOSVendor  string `json:"biosVendor"`
	BIOSVersion string `json:"biosVersion"`
	BIOSDate    string `json:"biosDate"`

	// 主板信息
	MotherboardInfo MotherboardInfo `json:"motherboardInfo"`

	// 产品信息
	ProductName    string `json:"productName"`
	ProductVendor  string `json:"productVendor"`
	ProductVersion string `json:"productVersion"`
	ProductSerial  string `json:"productSerial"`

	// 安全信息
	IsVirtualMachine bool     `json:"isVirtualMachine"`
	SecuritySoftware []string `json:"securitySoftware"`
}

// DiskInfo 磁盘信息
type DiskInfo struct {
	Device      string  `json:"device"`
	MountPoint  string  `json:"mountPoint"`
	FSType      string  `json:"fsType"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
	IsRemovable bool    `json:"isRemovable"`
}

// NetworkInfo 网络接口信息
type NetworkInfo struct {
	Name        string   `json:"name"`
	MTU         int      `json:"mtu"`
	MACAddr     string   `json:"macAddr"`
	IPAddresses []string `json:"ipAddresses"`
	IsUp        bool     `json:"isUp"`
	Type        string   `json:"type"`
	Speed       uint64   `json:"speed"`
}

// GPUInfo 显卡信息
type GPUInfo struct {
	Name       string `json:"name"`
	Vendor     string `json:"vendor"`
	DriverVer  string `json:"driverVersion"`
	DriverDate string `json:"driverDate"`
	Memory     uint64 `json:"memory"`
}

// MotherboardInfo 主板信息
type MotherboardInfo struct {
	Manufacturer string `json:"manufacturer"`
	Product      string `json:"product"`
	SerialNumber string `json:"serialNumber"`
	Version      string `json:"version"`
}

// GetDeviceInfo 获取当前设备的软硬件信息
func GetDeviceInfo() (*DeviceInfo, error) {
	deviceInfo := &DeviceInfo{}

	// 获取系统信息
	if err := deviceInfo.collectSystemInfo(); err != nil {
		return nil, fmt.Errorf("获取系统信息失败: %v", err)
	}

	// 获取CPU信息
	if err := deviceInfo.collectCPUInfo(); err != nil {
		return nil, fmt.Errorf("获取CPU信息失败: %v", err)
	}

	// 获取内存信息
	if err := deviceInfo.collectMemoryInfo(); err != nil {
		return nil, fmt.Errorf("获取内存信息失败: %v", err)
	}

	// 获取磁盘信息
	if err := deviceInfo.collectDiskInfo(); err != nil {
		return nil, fmt.Errorf("获取磁盘信息失败: %v", err)
	}

	// 获取网络信息
	if err := deviceInfo.collectNetworkInfo(); err != nil {
		return nil, fmt.Errorf("获取网络信息失败: %v", err)
	}

	// 获取显卡信息
	if err := deviceInfo.collectGPUInfo(); err != nil {
		return nil, fmt.Errorf("获取显卡信息失败: %v", err)
	}

	// 获取BIOS和主板信息
	if err := deviceInfo.collectHardwareInfo(); err != nil {
		return nil, fmt.Errorf("获取硬件信息失败: %v", err)
	}

	// 获取安全信息
	if err := deviceInfo.collectSecurityInfo(); err != nil {
		return nil, fmt.Errorf("获取安全信息失败: %v", err)
	}

	return deviceInfo, nil
}

// collectSystemInfo 收集系统信息
func (d *DeviceInfo) collectSystemInfo() error {
	// 获取主机信息
	hostInfo, err := host.Info()
	if err != nil {
		return err
	}

	d.Hostname = hostInfo.Hostname
	d.OS = hostInfo.OS
	d.Platform = hostInfo.Platform
	d.PlatformFamily = hostInfo.PlatformFamily
	d.PlatformVer = hostInfo.PlatformVersion
	d.KernelVer = hostInfo.KernelVersion
	d.KernelArch = hostInfo.KernelArch

	// 获取产品信息
	d.collectProductInfo()

	// 获取设备唯一标识
	d.collectDeviceIdentifiers()

	return nil
}

// collectDeviceIdentifiers 收集设备唯一标识信息
func (d *DeviceInfo) collectDeviceIdentifiers() error {
	// 1. 收集 Windows 产品ID
	d.collectWindowsProductID()

	// 2. 收集 BIOS UUID
	d.collectBIOSUUID()

	// 3. 收集硬件ID (使用主板、CPU和硬盘信息组合)
	d.collectHardwareID()

	// 4. 生成设备唯一ID
	d.generateUniqueID()

	return nil
}

// collectWindowsProductID 获取Windows产品ID
func (d *DeviceInfo) collectWindowsProductID() {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.READ)
	if err != nil {
		return
	}
	defer key.Close()

	// 尝试读取产品ID
	productID, _, _ := key.GetStringValue("ProductId")
	if productID == "" {
		// 备用方案：尝试读取数字产品ID
		digitalID, _, _ := key.GetBinaryValue("DigitalProductId")
		if len(digitalID) > 0 {
			// 将数字产品ID转换为字符串表示
			d.ProductID = fmt.Sprintf("%x", digitalID)
		}
	} else {
		d.ProductID = productID
	}
}

// collectBIOSUUID 获取BIOS UUID
func (d *DeviceInfo) collectBIOSUUID() {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\BIOS`, registry.READ)
	if err != nil {
		return
	}
	defer key.Close()

	// 读取BIOS UUID
	uuid, _, _ := key.GetStringValue("SystemUUID")
	if uuid == "" {
		// 备用方案：尝试读取系统产品UUID
		uuid, _, _ = key.GetStringValue("UUID")
	}

	d.BIOSUUID = uuid
}

// collectHardwareID 收集硬件ID
func (d *DeviceInfo) collectHardwareID() {
	var hardwareComponents []string

	// 1. 使用主板序列号
	if d.MotherboardInfo.SerialNumber != "" {
		hardwareComponents = append(hardwareComponents, "MB:"+d.MotherboardInfo.SerialNumber)
	}

	// 2. 使用CPU ID
	cpuID := getCPUID()
	if cpuID != "" {
		hardwareComponents = append(hardwareComponents, "CPU:"+cpuID)
	}

	// 3. 使用主磁盘序列号
	diskSerials := getPhysicalDiskSerials()
	if len(diskSerials) > 0 {
		for i, serial := range diskSerials {
			if i > 1 { // 只使用前两个磁盘
				break
			}
			hardwareComponents = append(hardwareComponents, fmt.Sprintf("DISK%d:%s", i, serial))
		}
	}

	// 4. 使用网卡MAC地址
	macAddresses := getMACAddresses()
	if len(macAddresses) > 0 {
		for i, mac := range macAddresses {
			if i > 1 { // 只使用前两个MAC地址
				break
			}
			if mac != "" {
				hardwareComponents = append(hardwareComponents, fmt.Sprintf("MAC%d:%s", i, mac))
			}
		}
	}

	// 组合硬件ID
	d.HardwareID = strings.Join(hardwareComponents, ";")
}

// getMACAddresses 获取所有物理网卡的MAC地址
func getMACAddresses() []string {
	var macAddresses []string

	interfaces, err := net.Interfaces()
	if err != nil {
		return macAddresses
	}

	for _, iface := range interfaces {
		// 排除虚拟接口和回环接口
		if iface.Flags&net.FlagLoopback == 0 && !isVirtualInterface(iface.Name) && len(iface.HardwareAddr) > 0 {
			macAddresses = append(macAddresses, iface.HardwareAddr.String())
		}
	}

	return macAddresses
}

// isVirtualInterface 检查是否是虚拟网络接口
func isVirtualInterface(name string) bool {
	// 常见虚拟接口名称标识
	virtualPrefixes := []string{
		"vethernet", "veth", "vmnet", "vboxnet",
		"docker", "virbr", "br-", "vnet", "virtual",
	}

	nameLower := strings.ToLower(name)
	for _, prefix := range virtualPrefixes {
		if strings.Contains(nameLower, prefix) {
			return true
		}
	}

	return false
}

// getCPUID 获取CPU ID
func getCPUID() string {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\CentralProcessor\0`, registry.READ)
	if err != nil {
		return ""
	}
	defer key.Close()

	// 读取处理器ID
	processorID, _, _ := key.GetStringValue("ProcessorNameString")
	if processorID == "" {
		processorID, _, _ = key.GetStringValue("Identifier")
	}

	// 读取特征码
	featureID, _, _ := key.GetIntegerValue("FeatureSet")

	return fmt.Sprintf("%s-%d", processorID, featureID)
}

// getPhysicalDiskSerials 获取物理磁盘序列号
func getPhysicalDiskSerials() []string {
	var serials []string

	// 使用WMI查询物理磁盘序列号
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DEVICEMAP\Scsi`, registry.READ)
	if err != nil {
		return serials
	}
	defer key.Close()

	// 获取所有SCSI控制器
	controllers, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return serials
	}

	// 遍历所有控制器和设备
	for _, controller := range controllers {
		controllerKey, err := registry.OpenKey(key, controller, registry.READ)
		if err != nil {
			continue
		}
		defer controllerKey.Close()

		// 获取设备
		devices, err := controllerKey.ReadSubKeyNames(-1)
		if err != nil {
			continue
		}

		for _, device := range devices {
			deviceKey, err := registry.OpenKey(controllerKey, device, registry.READ)
			if err != nil {
				continue
			}
			defer deviceKey.Close()

			// 读取设备标识符和序列号
			identifier, _, _ := deviceKey.GetStringValue("Identifier")
			serialNumber, _, _ := deviceKey.GetStringValue("SerialNumber")

			if serialNumber != "" {
				serials = append(serials, serialNumber)
			} else if identifier != "" {
				serials = append(serials, identifier)
			}
		}
	}

	// 如果使用注册表方法没有找到，尝试使用WMIC命令获取
	if len(serials) == 0 {
		// 尝试使用固定一些已知信息(由于WMIC命令执行比较复杂，这里简化处理)
		key, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\BIOS`, registry.READ)
		if err == nil {
			defer key.Close()
			diskSerial, _, _ := key.GetStringValue("DiskSerialNumber")
			if diskSerial != "" {
				serials = append(serials, diskSerial)
			}
		}
	}

	return serials
}

// generateUniqueID 生成设备唯一标识
func (d *DeviceInfo) generateUniqueID() {
	// 收集所有可用的标识符
	var identifiers []string

	// 添加BIOS UUID (如果有)
	if d.BIOSUUID != "" {
		identifiers = append(identifiers, d.BIOSUUID)
	}

	// 添加产品ID (如果有)
	if d.ProductID != "" {
		identifiers = append(identifiers, d.ProductID)
	}

	// 添加硬件ID (如果有)
	if d.HardwareID != "" {
		identifiers = append(identifiers, d.HardwareID)
	}

	// 添加产品序列号 (如果有)
	if d.ProductSerial != "" {
		identifiers = append(identifiers, d.ProductSerial)
	}

	// 添加主板序列号 (如果有)
	if d.MotherboardInfo.SerialNumber != "" {
		identifiers = append(identifiers, d.MotherboardInfo.SerialNumber)
	}

	// 生成唯一ID
	machineIdSource := strings.Join(identifiers, "-")

	// 设置机器ID (简单的哈希算法)
	d.MachineID = generateHash(machineIdSource)

	// 设置最终的唯一ID
	// 使用产品ID + 主板信息 + MAC地址的组合作为唯一标识
	d.UniqueID = d.MachineID
}

// generateHash 生成字符串的哈希值
func generateHash(input string) string {
	if input == "" {
		return ""
	}

	// 一种简单的哈希算法
	var h uint32
	for i := 0; i < len(input); i++ {
		h = 31*h + uint32(input[i])
	}

	return fmt.Sprintf("%x", h)
}

// collectCPUInfo 收集CPU信息
func (d *DeviceInfo) collectCPUInfo() error {
	// 获取CPU信息
	cpuInfo, err := cpu.Info()
	if err != nil {
		return err
	}

	if len(cpuInfo) > 0 {
		d.CPUModel = cpuInfo[0].ModelName
		d.CPUCores = len(cpuInfo)
		d.CPUFrequency = cpuInfo[0].Mhz
	}

	// 获取CPU使用率
	percentage, err := cpu.Percent(0, false)
	if err == nil && len(percentage) > 0 {
		d.CPUUsage = percentage[0]
	}

	return nil
}

// collectMemoryInfo 收集内存信息
func (d *DeviceInfo) collectMemoryInfo() error {
	// 获取内存信息
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	d.MemTotal = memInfo.Total
	d.MemAvailable = memInfo.Available
	d.MemUsed = memInfo.Used
	d.MemUsedPercent = memInfo.UsedPercent

	return nil
}

// collectDiskInfo 收集磁盘信息
func (d *DeviceInfo) collectDiskInfo() error {
	// 获取磁盘分区信息
	partitions, err := disk.Partitions(false)
	if err != nil {
		return err
	}

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		// 判断是否为可移动设备
		isRemovable := false
		if strings.HasPrefix(partition.Device, "\\\\.\\") {
			drive := partition.Mountpoint
			if len(drive) >= 2 && drive[1] == ':' {
				driveType := windows.GetDriveType(windows.StringToUTF16Ptr(drive + "\\"))
				isRemovable = driveType == windows.DRIVE_REMOVABLE
			}
		}

		diskInfo := DiskInfo{
			Device:      partition.Device,
			MountPoint:  partition.Mountpoint,
			FSType:      partition.Fstype,
			Total:       usage.Total,
			Used:        usage.Used,
			UsedPercent: usage.UsedPercent,
			IsRemovable: isRemovable,
		}

		d.DiskInfo = append(d.DiskInfo, diskInfo)
	}

	return nil
}

// collectNetworkInfo 收集网络信息
func (d *DeviceInfo) collectNetworkInfo() error {
	// 获取网络接口信息
	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	for _, iface := range interfaces {
		// 过滤掉回环接口
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		netInfo := NetworkInfo{
			Name:    iface.Name,
			MTU:     iface.MTU,
			MACAddr: iface.HardwareAddr.String(),
			IsUp:    iface.Flags&net.FlagUp != 0,
		}

		// 获取IP地址
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			netInfo.IPAddresses = append(netInfo.IPAddresses, addr.String())
		}

		// 获取网络接口类型和速度
		netInfo.Type, netInfo.Speed = getNetworkInterfaceTypeAndSpeed(iface.Name)

		d.NetworkInfo = append(d.NetworkInfo, netInfo)
	}

	return nil
}

// getNetworkInterfaceTypeAndSpeed 获取网络接口类型和速度
func getNetworkInterfaceTypeAndSpeed(interfaceName string) (string, uint64) {
	// 尝试从注册表获取接口类型和速度
	keyPath := fmt.Sprintf(`SYSTEM\CurrentControlSet\Control\Network\{4D36E972-E325-11CE-BFC1-08002BE10318}`)
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.READ)
	if err != nil {
		return "未知", 0
	}
	defer key.Close()

	// 遍历子键查找匹配的接口
	subKeys, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return "未知", 0
	}

	for _, subKey := range subKeys {
		connectionKey, err := registry.OpenKey(key, subKey+`\Connection`, registry.READ)
		if err != nil {
			continue
		}
		defer connectionKey.Close()

		// 检查接口名称
		name, _, err := connectionKey.GetStringValue("Name")
		if err != nil || name != interfaceName {
			continue
		}

		// 获取接口类型
		mediaType := "有线"
		if strings.Contains(strings.ToLower(name), "wi-fi") || strings.Contains(strings.ToLower(name), "wireless") {
			mediaType = "无线"
		} else if strings.Contains(strings.ToLower(name), "bluetooth") {
			mediaType = "蓝牙"
		}

		// 获取接口速度
		var speed uint64 = 0
		speedValue, _, err := connectionKey.GetIntegerValue("Speed")
		if err == nil {
			speed = speedValue
		}

		return mediaType, speed
	}

	return "未知", 0
}

// collectGPUInfo 收集显卡信息
func (d *DeviceInfo) collectGPUInfo() error {
	// 从注册表获取显卡信息
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Class\{4d36e968-e325-11ce-bfc1-08002be10318}`, registry.READ)
	if err != nil {
		return err
	}
	defer key.Close()

	// 遍历子键查找显卡
	subKeys, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return err
	}

	for _, subKey := range subKeys {
		if subKey == "Properties" {
			continue
		}

		gpuKey, err := registry.OpenKey(key, subKey, registry.READ)
		if err != nil {
			continue
		}
		defer gpuKey.Close()

		// 读取显卡信息
		driverDesc, _, err := gpuKey.GetStringValue("DriverDesc")
		if err != nil {
			continue
		}

		// 如果能找到驱动描述，说明是显卡
		gpuInfo := GPUInfo{
			Name: driverDesc,
		}

		// 尝试获取其他信息
		gpuInfo.Vendor, _, _ = gpuKey.GetStringValue("ProviderName")
		gpuInfo.DriverVer, _, _ = gpuKey.GetStringValue("DriverVersion")
		gpuInfo.DriverDate, _, _ = gpuKey.GetStringValue("DriverDate")

		// 获取显存
		memoryValue, _, err := gpuKey.GetIntegerValue("HardwareInformation.MemorySize")
		if err == nil {
			gpuInfo.Memory = memoryValue
		}

		d.GPUInfo = append(d.GPUInfo, gpuInfo)
	}

	return nil
}

// collectHardwareInfo 收集BIOS和主板信息
func (d *DeviceInfo) collectHardwareInfo() error {
	// 获取BIOS信息
	err := d.collectBIOSInfo()
	if err != nil {
		return err
	}

	// 获取主板信息
	err = d.collectMotherboardInfo()
	if err != nil {
		return err
	}

	return nil
}

// collectBIOSInfo 收集BIOS信息
func (d *DeviceInfo) collectBIOSInfo() error {
	// 使用WMI查询BIOS信息
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\BIOS`, registry.READ)
	if err != nil {
		return err
	}
	defer key.Close()

	// 读取BIOS信息
	d.BIOSVendor, _, _ = key.GetStringValue("BIOSVendor")
	d.BIOSVersion, _, _ = key.GetStringValue("BIOSVersion")
	d.BIOSDate, _, _ = key.GetStringValue("BIOSReleaseDate")

	return nil
}

// collectMotherboardInfo 收集主板信息
func (d *DeviceInfo) collectMotherboardInfo() error {
	// 从注册表获取主板信息
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\BIOS`, registry.READ)
	if err != nil {
		return err
	}
	defer key.Close()

	d.MotherboardInfo.Manufacturer, _, _ = key.GetStringValue("BaseBoardManufacturer")
	d.MotherboardInfo.Product, _, _ = key.GetStringValue("BaseBoardProduct")
	d.MotherboardInfo.SerialNumber, _, _ = key.GetStringValue("BaseBoardSerialNumber")
	d.MotherboardInfo.Version, _, _ = key.GetStringValue("BaseBoardVersion")

	return nil
}

// collectProductInfo 收集产品信息
func (d *DeviceInfo) collectProductInfo() error {
	// 从注册表获取产品信息
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\BIOS`, registry.READ)
	if err != nil {
		return err
	}
	defer key.Close()

	d.ProductName, _, _ = key.GetStringValue("SystemProductName")
	d.ProductVendor, _, _ = key.GetStringValue("SystemManufacturer")
	d.ProductVersion, _, _ = key.GetStringValue("SystemVersion")
	d.ProductSerial, _, _ = key.GetStringValue("SystemSerialNumber")

	return nil
}

// collectSecurityInfo 收集安全相关信息
func (d *DeviceInfo) collectSecurityInfo() error {
	// 检测是否为虚拟机
	d.detectVirtualMachine()

	return nil
}

// detectVirtualMachine 检测是否为虚拟机
func (d *DeviceInfo) detectVirtualMachine() {
	// 检查常见虚拟机特征

	// 1. 检查制造商信息
	vmManufacturers := []string{"VMware", "VirtualBox", "KVM", "QEMU", "Xen", "Parallels", "Virtual Machine"}
	for _, vm := range vmManufacturers {
		if strings.Contains(strings.ToLower(d.MotherboardInfo.Manufacturer), strings.ToLower(vm)) ||
			strings.Contains(strings.ToLower(d.ProductVendor), strings.ToLower(vm)) {
			d.IsVirtualMachine = true
			return
		}
	}

	// 2. 检查产品名称
	if strings.Contains(strings.ToLower(d.ProductName), "virtual") {
		d.IsVirtualMachine = true
		return
	}

	// 3. 通过特定设备检测
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Enum\PCI`, registry.READ)
	if err == nil {
		defer key.Close()
		subKeys, _ := key.ReadSubKeyNames(-1)
		for _, subKey := range subKeys {
			if strings.Contains(strings.ToLower(subKey), "vmware") ||
				strings.Contains(strings.ToLower(subKey), "virtualbox") {
				d.IsVirtualMachine = true
				return
			}
		}
	}

	// 默认为物理机
	d.IsVirtualMachine = false
}

// FormatDeviceInfo 格式化设备信息为可读字符串
func FormatDeviceInfo(info *DeviceInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("计算机名: %s\n", info.Hostname))
	sb.WriteString(fmt.Sprintf("操作系统: %s %s (%s)\n", info.Platform, info.PlatformVer, info.OS))
	sb.WriteString(fmt.Sprintf("内核版本: %s\n", info.KernelVer))
	sb.WriteString(fmt.Sprintf("架构: %s\n", info.KernelArch))
	sb.WriteString("\n")

	// 设备唯一标识
	sb.WriteString(fmt.Sprintf("设备唯一标识: %s\n", info.UniqueID))
	sb.WriteString(fmt.Sprintf("机器ID: %s\n", info.MachineID))
	if info.BIOSUUID != "" {
		sb.WriteString(fmt.Sprintf("BIOS UUID: %s\n", info.BIOSUUID))
	}
	if info.ProductID != "" {
		sb.WriteString(fmt.Sprintf("产品ID: %s\n", info.ProductID))
	}
	sb.WriteString("\n")

	// 硬件信息
	sb.WriteString(fmt.Sprintf("产品型号: %s\n", info.ProductName))
	sb.WriteString(fmt.Sprintf("制造商: %s\n", info.ProductVendor))
	sb.WriteString(fmt.Sprintf("序列号: %s\n", info.ProductSerial))
	sb.WriteString("\n")

	// CPU信息
	sb.WriteString(fmt.Sprintf("CPU: %s\n", info.CPUModel))
	sb.WriteString(fmt.Sprintf("CPU核心数: %d\n", info.CPUCores))
	sb.WriteString(fmt.Sprintf("CPU频率: %.2f MHz\n", info.CPUFrequency))
	sb.WriteString(fmt.Sprintf("CPU使用率: %.2f%%\n", info.CPUUsage))
	sb.WriteString("\n")

	// 内存信息
	sb.WriteString(fmt.Sprintf("内存总量: %s\n", formatBytes(info.MemTotal)))
	sb.WriteString(fmt.Sprintf("可用内存: %s\n", formatBytes(info.MemAvailable)))
	sb.WriteString(fmt.Sprintf("内存使用率: %.2f%%\n", info.MemUsedPercent))
	sb.WriteString("\n")

	// 磁盘信息
	sb.WriteString("磁盘信息:\n")
	for _, diskInfo := range info.DiskInfo {
		sb.WriteString(fmt.Sprintf("  %s (%s):\n", diskInfo.MountPoint, diskInfo.FSType))
		sb.WriteString(fmt.Sprintf("    总容量: %s\n", formatBytes(diskInfo.Total)))
		sb.WriteString(fmt.Sprintf("    使用率: %.2f%%\n", diskInfo.UsedPercent))
		if diskInfo.IsRemovable {
			sb.WriteString("    类型: 可移动存储设备\n")
		} else {
			sb.WriteString("    类型: 固定磁盘\n")
		}
	}
	sb.WriteString("\n")

	// 网络信息
	sb.WriteString("网络适配器:\n")
	for _, networkInfo := range info.NetworkInfo {
		sb.WriteString(fmt.Sprintf("  %s (%s):\n", networkInfo.Name, networkInfo.Type))
		sb.WriteString(fmt.Sprintf("    MAC地址: %s\n", networkInfo.MACAddr))
		sb.WriteString(fmt.Sprintf("    状态: %s\n", formatNetStatus(networkInfo.IsUp)))
		if len(networkInfo.IPAddresses) > 0 {
			sb.WriteString("    IP地址:\n")
			for _, ip := range networkInfo.IPAddresses {
				sb.WriteString(fmt.Sprintf("      %s\n", ip))
			}
		}
	}
	sb.WriteString("\n")

	// 显卡信息
	sb.WriteString("显卡信息:\n")
	for _, gpu := range info.GPUInfo {
		sb.WriteString(fmt.Sprintf("  %s\n", gpu.Name))
		sb.WriteString(fmt.Sprintf("    厂商: %s\n", gpu.Vendor))
		sb.WriteString(fmt.Sprintf("    驱动版本: %s\n", gpu.DriverVer))
		if gpu.Memory > 0 {
			sb.WriteString(fmt.Sprintf("    显存: %s\n", formatBytes(gpu.Memory)))
		}
	}
	sb.WriteString("\n")

	// BIOS信息
	sb.WriteString(fmt.Sprintf("BIOS信息:\n"))
	sb.WriteString(fmt.Sprintf("  厂商: %s\n", info.BIOSVendor))
	sb.WriteString(fmt.Sprintf("  版本: %s\n", info.BIOSVersion))
	sb.WriteString(fmt.Sprintf("  日期: %s\n", info.BIOSDate))
	sb.WriteString("\n")

	// 安全信息
	sb.WriteString("安全信息:\n")
	sb.WriteString(fmt.Sprintf("  设备类型: %s\n", formatVirtualStatus(info.IsVirtualMachine)))

	return sb.String()
}

// formatBytes 格式化字节为可读格式
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatNetStatus 格式化网络状态
func formatNetStatus(isUp bool) string {
	if isUp {
		return "已连接"
	}
	return "已断开"
}

// formatVirtualStatus 格式化虚拟机状态
func formatVirtualStatus(isVirtual bool) string {
	if isVirtual {
		return "虚拟机"
	}
	return "物理机"
}
