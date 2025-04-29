package foreground

import (
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	psapi    = syscall.NewLazyDLL("psapi.dll")

	procGetForeground     = user32.NewProc("GetForegroundWindow")
	procGetWindowText     = user32.NewProc("GetWindowTextW")
	procGetWindowProcess  = user32.NewProc("GetWindowThreadProcessId")
	procOpenProcess       = kernel32.NewProc("OpenProcess")
	procCloseHandle       = kernel32.NewProc("CloseHandle")
	procGetModuleName     = psapi.NewProc("GetModuleFileNameExW")
	procQueryFullProcess  = kernel32.NewProc("QueryFullProcessImageNameW")
	procGetModuleBaseName = psapi.NewProc("GetModuleBaseNameW")
)

const (
	PROCESS_QUERY_INFORMATION         = 0x0400
	PROCESS_VM_READ                   = 0x0010
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
)

// WindowInfo 窗口信息结构体
type WindowInfo struct {
	Handle      syscall.Handle // 窗口句柄
	Title       string         // 窗口标题
	ProcessID   uint32         // 进程ID
	ProcessName string         // 进程名称
	ProcessPath string         // 进程完整路径
}

// GetForegroundWindow 获取当前焦点窗口句柄
func GetForegroundWindow() syscall.Handle {
	ret, _, _ := procGetForeground.Call()
	return syscall.Handle(ret)
}

// GetWindowText 获取窗口标题
func GetWindowText(hwnd syscall.Handle) string {
	// 防止无效句柄
	if hwnd == 0 {
		return ""
	}

	var text [512]uint16
	ret, _, _ := procGetWindowText.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&text[0])),
		uintptr(len(text)),
	)

	if ret == 0 {
		return ""
	}

	return syscall.UTF16ToString(text[:])
}

// GetProcessID 获取窗口关联进程ID
func GetProcessID(hwnd syscall.Handle) uint32 {
	// 防止无效句柄
	if hwnd == 0 {
		return 0
	}

	var pid uint32
	// GetWindowThreadProcessId 返回值是线程ID，通过第二个参数输出进程ID
	// 第一个返回值不是错误代码，所以不能用它判断是否成功
	_, _, _ = procGetWindowProcess.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&pid)),
	)

	return pid
}

// GetProcessName 获取进程可执行文件名称（仅文件名）
func GetProcessName(pid uint32) string {
	// 防止无效进程ID
	if pid == 0 {
		return ""
	}

	// 打开进程获取句柄，使用更宽松的访问权限
	hProcess, _, _ := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ|PROCESS_QUERY_LIMITED_INFORMATION),
		0,
		uintptr(pid),
	)

	if hProcess == 0 {
		// 如果失败，尝试使用最低的访问权限
		hProcess, _, _ = procOpenProcess.Call(
			uintptr(PROCESS_QUERY_LIMITED_INFORMATION),
			0,
			uintptr(pid),
		)
		if hProcess == 0 {
			return "无法访问"
		}
	}

	defer procCloseHandle.Call(hProcess)

	// 尝试方法1：使用GetModuleBaseNameW直接获取基本名称
	var baseName [260]uint16
	ret, _, _ := procGetModuleBaseName.Call(
		hProcess,
		0,
		uintptr(unsafe.Pointer(&baseName[0])),
		uintptr(len(baseName)),
	)

	if ret > 0 {
		return syscall.UTF16ToString(baseName[:ret])
	}

	// 尝试方法2：获取完整路径再提取文件名
	if processPath := GetProcessPath(pid); processPath != "" {
		return filepath.Base(processPath)
	}

	return "未知进程"
}

// GetProcessPath 获取进程完整路径
func GetProcessPath(pid uint32) string {
	// 防止无效进程ID
	if pid == 0 {
		return ""
	}

	// 打开进程获取句柄
	hProcess, _, _ := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_LIMITED_INFORMATION),
		0,
		uintptr(pid),
	)

	if hProcess == 0 {
		return ""
	}

	defer procCloseHandle.Call(hProcess)

	// 尝试方法1：使用QueryFullProcessImageNameW(Windows Vista及以上)
	var pathLen uint32 = syscall.MAX_PATH
	var path [syscall.MAX_PATH]uint16
	ret, _, _ := procQueryFullProcess.Call(
		hProcess,
		0,
		uintptr(unsafe.Pointer(&path[0])),
		uintptr(unsafe.Pointer(&pathLen)),
	)

	if ret != 0 {
		return syscall.UTF16ToString(path[:pathLen])
	}

	// 尝试方法2：使用GetModuleFileNameExW
	ret, _, _ = procGetModuleName.Call(
		hProcess,
		0,
		uintptr(unsafe.Pointer(&path[0])),
		uintptr(syscall.MAX_PATH),
	)

	if ret > 0 {
		return syscall.UTF16ToString(path[:ret])
	}

	return ""
}

// GetWindowInfo 获取当前焦点窗口完整信息
func GetWindowInfo() *WindowInfo {
	// 获取当前焦点窗口
	hwnd := GetForegroundWindow()
	if hwnd == 0 {
		return nil
	}

	// 获取窗口标题
	title := GetWindowText(hwnd)

	// 获取进程ID
	pid := GetProcessID(hwnd)
	if pid == 0 {
		return &WindowInfo{
			Handle: hwnd,
			Title:  title,
		}
	}

	// 获取进程路径
	processPath := GetProcessPath(pid)

	// 获取进程名称
	var processName string
	if processPath != "" {
		processName = filepath.Base(processPath)
	} else {
		processName = GetProcessName(pid)
	}

	return &WindowInfo{
		Handle:      hwnd,
		Title:       title,
		ProcessID:   pid,
		ProcessName: processName,
		ProcessPath: processPath,
	}
}
