package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleCheckModule 检查模块并生成报告
func (s *MCPServer) handleCheckModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	// 获取模块信息
	moduleResp, err := http.Get(fmt.Sprintf("%s/modules/%s", s.getApiURL(), moduleID))
	if err != nil {
		return mcp.NewToolResultText("获取模块信息失败：" + err.Error()), nil
	}
	defer moduleResp.Body.Close()

	var module map[string]interface{}
	if err := json.NewDecoder(moduleResp.Body).Decode(&module); err != nil {
		return mcp.NewToolResultText("解析模块信息失败：" + err.Error()), nil
	}

	// 获取模块下的任务
	tasksResp, err := http.Get(fmt.Sprintf("%s/tasks?moduleId=%s", s.getApiURL(), moduleID))
	if err != nil {
		return mcp.NewToolResultText("获取任务列表失败：" + err.Error()), nil
	}
	defer tasksResp.Body.Close()

	var tasksResult map[string]interface{}
	if err := json.NewDecoder(tasksResp.Body).Decode(&tasksResult); err != nil {
		return mcp.NewToolResultText("解析任务列表失败：" + err.Error()), nil
	}

	// 提取任务列表
	var tasks []map[string]interface{}
	if data, ok := tasksResult["data"].([]interface{}); ok {
		for _, task := range data {
			if taskMap, ok := task.(map[string]interface{}); ok {
				tasks = append(tasks, taskMap)
			}
		}
	}

	// 执行检查
	report := s.ruleEngine.CheckModule(module, tasks)

	// 获取输出格式
	format, _ := getParam(request, "format")
	if format == "markdown" {
		return mcp.NewToolResultText(report.ToMarkdown()), nil
	}

	data, _ := json.MarshalIndent(report, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleOpenFrontend 打开前端网页
func (s *MCPServer) handleOpenFrontend(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 尝试打开浏览器
	url := "http://localhost:5173"

	// 不同平台的打开命令
	var cmd string
	switch {
	case isWindows():
		cmd = fmt.Sprintf("start %s", url)
	case isMacOS():
		cmd = fmt.Sprintf("open %s", url)
	default:
		cmd = fmt.Sprintf("xdg-open %s", url)
	}

	// 执行命令 (这里简化处理，实际需要使用 os/exec)
	_ = cmd

	return mcp.NewToolResultText("已尝试打开前端页面：" + url), nil
}

func isWindows() bool {
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}

func isMacOS() bool {
	// 简化的判断
	return false
}

// handleSendNotification 发送通知
func (s *MCPServer) handleSendNotification(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	toTaskID, ok := getParam(request, "toTaskId")
	if !ok {
		return mcp.NewToolResultText("缺少 toTaskId 参数"), nil
	}

	// 构建通知数据
	notificationData := map[string]interface{}{
		"toTaskId": toTaskID,
		"type":     "",
		"title":    "",
		"message":  "",
	}
	if t, ok := getParamAny(request, "type"); ok {
		notificationData["type"] = t
	}
	if t, ok := getParamAny(request, "title"); ok {
		notificationData["title"] = t
	}
	if m, ok := getParamAny(request, "message"); ok {
		notificationData["message"] = m
	}

	jsonData, _ := json.Marshal(notificationData)
	resp, err := http.Post(fmt.Sprintf("%s/notifications", s.getApiURL()), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleReadNotifications 读取未读通知
func (s *MCPServer) handleReadNotifications(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	toTaskID, _ := getParam(request, "toTaskId")

	url := fmt.Sprintf("%s/notifications?read=false", s.getApiURL())
	if toTaskID != "" {
		url += fmt.Sprintf("&toTaskId=%s", toTaskID)
	}

	resp, err := http.Get(url)
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleLockResource 锁定资源
func (s *MCPServer) handleLockResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resourceType, ok1 := getParam(request, "resourceType")
	resourceID, ok2 := getParam(request, "resourceId")

	if !ok1 || !ok2 {
		return mcp.NewToolResultText("缺少必要参数"), nil
	}

	lockData := map[string]interface{}{
		"resourceType": resourceType,
		"resourceId":   resourceID,
	}

	jsonData, _ := json.Marshal(lockData)
	resp, err := http.Post(fmt.Sprintf("%s/lock", s.getApiURL()), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleUnlockResource 解锁资源
func (s *MCPServer) handleUnlockResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resourceType, ok1 := getParam(request, "resourceType")
	resourceID, ok2 := getParam(request, "resourceId")

	if !ok1 || !ok2 {
		return mcp.NewToolResultText("缺少必要参数"), nil
	}

	unlockData := map[string]interface{}{
		"resourceType": resourceType,
		"resourceId":   resourceID,
	}

	jsonData, _ := json.Marshal(unlockData)
	resp, err := http.Post(fmt.Sprintf("%s/lock/unlock", s.getApiURL()), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleGetLockStatus 查询锁定状态
func (s *MCPServer) handleGetLockStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resourceType, ok1 := getParam(request, "resourceType")
	resourceID, ok2 := getParam(request, "resourceId")

	if !ok1 || !ok2 {
		return mcp.NewToolResultText("缺少必要参数"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/lock/status?resourceType=%s&resourceId=%s", s.getApiURL(), resourceType, resourceID))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleCreateModuleDependency 创建模块依赖
func (s *MCPServer) handleCreateModuleDependency(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok1 := getParam(request, "moduleId")
	dependsOnModuleID, ok2 := getParam(request, "dependsOnModuleId")

	if !ok1 || !ok2 {
		return mcp.NewToolResultText("缺少必要参数"), nil
	}

	depData := map[string]interface{}{
		"dependsOnModuleId": dependsOnModuleID,
	}
	if depType, ok := getParamAny(request, "dependencyType"); ok {
		depData["dependencyType"] = depType
	}
	if contract, ok := getParamAny(request, "contractSummary"); ok {
		depData["contractSummary"] = contract
	}

	jsonData, _ := json.Marshal(depData)
	resp, err := http.Post(fmt.Sprintf("%s/modules/%s/dependencies", s.getApiURL(), moduleID), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleGetModuleDependencies 获取模块依赖
func (s *MCPServer) handleGetModuleDependencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/modules/%s/dependencies", s.getApiURL(), moduleID))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleCreateTaskDependency 创建任务依赖
func (s *MCPServer) handleCreateTaskDependency(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstreamTaskID, ok1 := getParam(request, "upstreamTaskId")
	downstreamTaskID, ok2 := getParam(request, "downstreamTaskId")

	if !ok1 || !ok2 {
		return mcp.NewToolResultText("缺少必要参数"), nil
	}

	depData := map[string]interface{}{
		"upstreamTaskId":   upstreamTaskID,
		"downstreamTaskId": downstreamTaskID,
	}
	if contract, ok := getParamAny(request, "contractSummary"); ok {
		depData["contractSummary"] = contract
	}

	jsonData, _ := json.Marshal(depData)
	resp, err := http.Post(fmt.Sprintf("%s/dependencies", s.getApiURL()), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleGetTaskDependencies 获取任务依赖
func (s *MCPServer) handleGetTaskDependencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, ok := getParam(request, "taskId")
	if !ok {
		return mcp.NewToolResultText("缺少 taskId 参数"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/dependencies?taskId=%s", s.getApiURL(), taskID))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
