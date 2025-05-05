package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// MonitorDataCollector 监控数据收集器
type MonitorDataCollector struct {
	ServerURL string
	Token     string
	AccountID int
	ExamID    int

	// 控制标志
	IsRunning         bool
	ScreenshotEnabled bool
	ProcessEnabled    bool
	WebsiteEnabled    bool
	BehaviorEnabled   bool

	// 采集间隔时间(秒)
	ScreenshotInterval int
	ProcessInterval    int
}

// NewMonitorDataCollector 创建监控数据收集器
func NewMonitorDataCollector(serverURL string, token string, accountID int, examID int) *MonitorDataCollector {
	return &MonitorDataCollector{
		ServerURL:         serverURL,
		Token:             token,
		AccountID:         accountID,
		ExamID:            examID,
		IsRunning:         false,
		ScreenshotEnabled: true,
		ProcessEnabled:    true,
		WebsiteEnabled:    true,
		BehaviorEnabled:   true,
		ProcessInterval:   60, // 默认15秒一次进程检查
	}
}

// Start 开始数据收集和上报
func (m *MonitorDataCollector) Start() {
	if m.IsRunning {
		return
	}

	m.IsRunning = true

	// 启动进程信息收集
	if m.ProcessEnabled {
		go m.startProcessCollection()
	}

	// 网站访问记录不需要定期收集，会在访问时即时上报

	fmt.Println("监控数据收集已启动")
}

// Stop 停止数据收集
func (m *MonitorDataCollector) Stop() {
	m.IsRunning = false
	fmt.Println("监控数据收集已停止")
}

func (m *MonitorDataCollector) startProcessCollection() {
	var reportedProcesses = make(map[string]bool)

	for m.IsRunning {
		processes, err := GetProcesses()
		if err != nil {
			fmt.Printf("获取进程信息失败: %v\n", err)
			time.Sleep(time.Duration(m.ProcessInterval) * time.Second)
			continue
		}

		var toReport []map[string]string
		for _, val := range processes {
			if reportedProcesses[val["name"]] {
				continue
			}
			toReport = append(toReport, val)

		}

		if len(toReport) > 0 {
			m.uploadProcesses(toReport)
		}

		time.Sleep(time.Duration(m.ProcessInterval) * time.Second)
	}
}

// ReportWebsiteVisit 上报网站访问记录
func (m *MonitorDataCollector) ReportWebsiteVisit(url string, title string) {
	if !m.IsRunning || !m.WebsiteEnabled {
		return
	}

	visitData := map[string]interface{}{
		"examId":            m.ExamID,
		"examineeAccountId": m.AccountID,
		"url":               url,
		"title":             title,
		"visitTime":         time.Now().Format("2006-01-02T15:04:05"),
	}

	jsonData, err := json.Marshal(visitData)
	if err != nil {
		fmt.Printf("序列化网站访问数据失败: %v\n", err)
		return
	}

	// 构建请求头
	headers := map[string]string{
		"Authorization": "Bearer " + m.Token,
		"Content-Type":  "application/json",
	}

	// 发送数据
	url = fmt.Sprintf("%s/monitor/data/website-visit", m.ServerURL)
	_, err = HttpPostWithHeaders(url, jsonData, headers)
	if err != nil {
		fmt.Printf("上报网站访问记录失败: %v\n", err)
	}
}

// ReportBehavior 上报行为数据
func (m *MonitorDataCollector) ReportBehavior(eventType int, content string, level string) {
	if !m.IsRunning || !m.BehaviorEnabled {
		return
	}

	behaviorData := map[string]interface{}{
		"examId":            m.ExamID,
		"examineeAccountId": m.AccountID,
		"eventType":         eventType,
		"content":           content,
		"eventTime":         time.Now().Format("2006-01-02T15:04:05"),
		"level":             level,
	}

	jsonData, err := json.Marshal(behaviorData)
	if err != nil {
		fmt.Printf("序列化行为数据失败: %v\n", err)
		return
	}

	// 构建请求头
	headers := map[string]string{
		"Authorization": "Bearer " + m.Token,
		"Content-Type":  "application/json",
	}

	// 发送数据
	url := fmt.Sprintf("%s/monitor/data/behavior", m.ServerURL)
	_, err = HttpPostWithHeaders(url, jsonData, headers)
	if err != nil {
		fmt.Printf("上报行为记录失败: %v\n", err)
	}
}

// UploadFile 上传文件到服务器
func (m *MonitorDataCollector) UploadFile(fileBytes []byte, filename string) (string, error) {
	// 创建一个缓冲区，用于存储multipart表单数据
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建文件表单字段
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("创建文件表单字段失败: %w", err)
	}

	// 写入文件数据
	_, err = io.Copy(part, bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("写入文件数据失败: %w", err)
	}

	// 添加文件名表单字段
	err = writer.WriteField("filename", filename)
	if err != nil {
		return "", fmt.Errorf("写入文件名失败: %w", err)
	}

	// 关闭writer，完成表单
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("关闭writer失败: %w", err)
	}

	// 构建请求
	uploadURL := fmt.Sprintf("%s/common/upload", m.ServerURL)
	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return "", fmt.Errorf("创建上传请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+m.Token)

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送上传请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取上传响应失败: %w", err)
	}
	// 解析响应
	var uploadResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data any    `json:"data"` // 假设返回的是文件相对路径
	}

	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return "", fmt.Errorf("解析上传响应失败: %w", err)
	}

	// 检查是否上传成功
	if uploadResp.Code != 0 {
		return "", fmt.Errorf("上传失败: %s", uploadResp.Msg)
	}

	// 返回文件路径
	return filename, nil
}

// 上报屏幕截图
func (m *MonitorDataCollector) uploadScreenshot(imageBuffer *bytes.Buffer) {
	if !m.IsRunning || !m.ScreenshotEnabled {
		return
	}

	// 创建唯一的文件名
	timestamp := time.Now().Format("20060102150405")
	studentId := fmt.Sprintf("%d", m.AccountID)
	examId := fmt.Sprintf("%d", m.ExamID)
	filename := fmt.Sprintf("screenshot/screenshot_%s_%s_%s.jpg", examId, studentId, timestamp)

	// 上传文件
	_, err := m.UploadFile(imageBuffer.Bytes(), filename)
	if err != nil {
		fmt.Printf("上传截图文件失败: %v\n", err)
		return
	}

	// 构建截图记录数据
	screenshotData := map[string]interface{}{
		"examId":            m.ExamID,
		"examineeAccountId": m.AccountID,
		"captureTime":       time.Now().Format("2006-01-02T15:04:05"),
		"screenshotUrl":     filename,
	}

	jsonData, err := json.Marshal(screenshotData)
	if err != nil {
		fmt.Printf("序列化截图数据失败: %v\n", err)
		return
	}

	// 构建请求头
	headers := map[string]string{
		"Authorization": "Bearer " + m.Token,
		"Content-Type":  "application/json",
	}

	// 发送数据
	url := fmt.Sprintf("%s/monitor/data/screenshot", m.ServerURL)
	_, err = HttpPostWithHeaders(url, jsonData, headers)
	if err != nil {
		fmt.Printf("上报截图失败: %v\n", err)
	} else {
		fmt.Printf("成功上传截图: %s\n", filename)
	}
}

// UploadScreenshotData 公开方法，用于上传截图数据
func (m *MonitorDataCollector) UploadScreenshotData(imageBuffer *bytes.Buffer) {
	m.uploadScreenshot(imageBuffer)
}

// 上报进程信息
func (m *MonitorDataCollector) uploadProcesses(processes []map[string]string) {
	if !m.IsRunning || !m.ProcessEnabled {
		return
	}

	processData := map[string]interface{}{
		"examId":            m.ExamID,
		"examineeAccountId": m.AccountID,
		"recordTime":        time.Now().Format("2006-01-02T15:04:05"),
		"processes":         processes,
	}

	jsonData, err := json.Marshal(processData)
	if err != nil {
		fmt.Printf("序列化进程数据失败: %v\n", err)
		return
	}

	// 构建请求头
	headers := map[string]string{
		"Authorization": "Bearer " + m.Token,
		"Content-Type":  "application/json",
	}

	// 发送数据
	url := fmt.Sprintf("%s/monitor/data/process", m.ServerURL)
	_, err = HttpPostWithHeaders(url, jsonData, headers)
	if err != nil {
		fmt.Printf("上报进程信息失败: %v\n", err)
	}
}

// GetProcesses 获取系统进程信息
func GetProcesses() ([]map[string]string, error) {
	// 获取进程信息（复用现有代码）
	processes, err := getProcessInfo()
	if err != nil {
		return nil, err
	}

	return processes, nil
}

func collectProcessInfo() []map[string]string {
	processes, err := process.Processes()
	if err != nil {
		return nil
	}

	// 只获取前10个进程信息
	var processInfo []map[string]string
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

		processInfo = append(processInfo, map[string]string{
			"pid":    strconv.Itoa(int(p.Pid)),
			"name":   name,
			"memory": strconv.Itoa(int(memUsage)),
			"cpu":    strconv.FormatFloat(cpuPercent, 'f', 2, 64),
		})

		count++
	}

	return processInfo
}

// 模拟获取进程信息
func getProcessInfo() ([]map[string]string, error) {
	return collectProcessInfo(), nil
}

// HttpPostWithHeaders 发送带头信息的POST请求
func HttpPostWithHeaders(url string, jsonData []byte, headers map[string]string) ([]byte, error) {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}
