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

// handleCheckModuleImpl 检查模块并生成报告
func (s *MCPServer) handleCheckModuleImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	// 获取模块信息
	moduleResp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), pathName))
	if err != nil {
		return mcp.NewToolResultText("获取模块信息失败：" + err.Error()), nil
	}
	defer moduleResp.Body.Close()

	var module map[string]interface{}
	if err := json.NewDecoder(moduleResp.Body).Decode(&module); err != nil {
		return mcp.NewToolResultText("解析模块信息失败：" + err.Error()), nil
	}

	// 获取模块ID用于查询任务
	moduleID, _ := module["id"].(string)

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

// handleOpenFrontendImpl 打开前端网页
func (s *MCPServer) handleOpenFrontendImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

// handleSendNotificationImpl 发送通知
func (s *MCPServer) handleSendNotificationImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数（目标任务路径名称）"), nil
	}

	// 先通过 pathName 获取任务 ID
	taskResp, err := http.Get(fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), pathName))
	if err != nil {
		return mcp.NewToolResultText("获取任务信息失败：" + err.Error()), nil
	}
	defer taskResp.Body.Close()

	var taskResult map[string]interface{}
	if err := json.NewDecoder(taskResp.Body).Decode(&taskResult); err != nil {
		return mcp.NewToolResultText("解析任务信息失败：" + err.Error()), nil
	}

	toTaskID, ok := taskResult["id"].(string)
	if !ok {
		return mcp.NewToolResultText("无法获取任务ID"), nil
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

// handleReadNotificationsImpl 读取未读通知
func (s *MCPServer) handleReadNotificationsImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, _ := getParam(request, "pathName")

	url := fmt.Sprintf("%s/notifications?read=false", s.getApiURL())
	if pathName != "" {
		// 先通过 pathName 获取任务 ID
		taskResp, err := http.Get(fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), pathName))
		if err == nil {
			defer taskResp.Body.Close()
			var taskResult map[string]interface{}
			if json.NewDecoder(taskResp.Body).Decode(&taskResult) == nil {
				if toTaskID, ok := taskResult["id"].(string); ok {
					url += fmt.Sprintf("&toTaskId=%s", toTaskID)
				}
			}
		}
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

// handleLockResourceImpl 锁定资源
func (s *MCPServer) handleLockResourceImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resourceType, ok1 := getParam(request, "resourceType")
	pathName, ok2 := getParam(request, "pathName")

	if !ok1 || !ok2 {
		return mcp.NewToolResultText("缺少必要参数"), nil
	}

	// 使用 by-path API 锁定资源
	lockData := map[string]interface{}{
		"resourceType": resourceType,
	}

	jsonData, _ := json.Marshal(lockData)
	resp, err := http.Post(fmt.Sprintf("%s/lock/by-path/%s", s.getApiURL(), pathName), "application/json", bytes.NewBuffer(jsonData))
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

// handleUnlockResourceImpl 解锁资源
func (s *MCPServer) handleUnlockResourceImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resourceType, ok1 := getParam(request, "resourceType")
	pathName, ok2 := getParam(request, "pathName")

	if !ok1 || !ok2 {
		return mcp.NewToolResultText("缺少必要参数"), nil
	}

	unlockData := map[string]interface{}{
		"resourceType": resourceType,
	}

	jsonData, _ := json.Marshal(unlockData)
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/lock/by-path/%s", s.getApiURL(), pathName), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
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

// handleGetLockStatusImpl 查询锁定状态
func (s *MCPServer) handleGetLockStatusImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resourceType, ok1 := getParam(request, "resourceType")
	pathName, ok2 := getParam(request, "pathName")

	if !ok1 || !ok2 {
		return mcp.NewToolResultText("缺少必要参数"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/lock/by-path/%s?resourceType=%s", s.getApiURL(), pathName, resourceType))
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

// handleCreateModuleDependencyImpl 创建模块依赖
func (s *MCPServer) handleCreateModuleDependencyImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok1 := getParam(request, "pathName")
	dependsOnPathName, ok2 := getParam(request, "dependsOnPathName")

	if !ok1 || !ok2 {
		return mcp.NewToolResultText("缺少必要参数"), nil
	}

	// 先获取被依赖模块的 ID
	dependsOnResp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), dependsOnPathName))
	if err != nil {
		return mcp.NewToolResultText("获取被依赖模块信息失败：" + err.Error()), nil
	}
	defer dependsOnResp.Body.Close()

	var dependsOnModule map[string]interface{}
	if err := json.NewDecoder(dependsOnResp.Body).Decode(&dependsOnModule); err != nil {
		return mcp.NewToolResultText("解析被依赖模块信息失败：" + err.Error()), nil
	}

	dependsOnModuleID, ok := dependsOnModule["id"].(string)
	if !ok {
		return mcp.NewToolResultText("无法获取被依赖模块ID"), nil
	}

	// 获取当前模块的 ID
	moduleResp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), pathName))
	if err != nil {
		return mcp.NewToolResultText("获取模块信息失败：" + err.Error()), nil
	}
	defer moduleResp.Body.Close()

	var module map[string]interface{}
	if err := json.NewDecoder(moduleResp.Body).Decode(&module); err != nil {
		return mcp.NewToolResultText("解析模块信息失败：" + err.Error()), nil
	}

	moduleID, ok := module["id"].(string)
	if !ok {
		return mcp.NewToolResultText("无法获取模块ID"), nil
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

// handleGetModuleDependenciesImpl 获取模块依赖
func (s *MCPServer) handleGetModuleDependenciesImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	// 使用 action=dependencies 参数
	resp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s?action=dependencies", s.getApiURL(), pathName))
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

// handleCreateTaskDependencyImpl 创建任务依赖
func (s *MCPServer) handleCreateTaskDependencyImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstreamPathName, ok1 := getParam(request, "upstreamPathName")
	downstreamPathName, ok2 := getParam(request, "downstreamPathName")

	if !ok1 || !ok2 {
		return mcp.NewToolResultText("缺少必要参数"), nil
	}

	// 获取上游任务 ID
	upstreamResp, err := http.Get(fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), upstreamPathName))
	if err != nil {
		return mcp.NewToolResultText("获取上游任务信息失败：" + err.Error()), nil
	}
	defer upstreamResp.Body.Close()

	var upstreamTask map[string]interface{}
	if err := json.NewDecoder(upstreamResp.Body).Decode(&upstreamTask); err != nil {
		return mcp.NewToolResultText("解析上游任务信息失败：" + err.Error()), nil
	}

	upstreamTaskID, ok := upstreamTask["id"].(string)
	if !ok {
		return mcp.NewToolResultText("无法获取上游任务ID"), nil
	}

	// 获取下游任务 ID
	downstreamResp, err := http.Get(fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), downstreamPathName))
	if err != nil {
		return mcp.NewToolResultText("获取下游任务信息失败：" + err.Error()), nil
	}
	defer downstreamResp.Body.Close()

	var downstreamTask map[string]interface{}
	if err := json.NewDecoder(downstreamResp.Body).Decode(&downstreamTask); err != nil {
		return mcp.NewToolResultText("解析下游任务信息失败：" + err.Error()), nil
	}

	downstreamTaskID, ok := downstreamTask["id"].(string)
	if !ok {
		return mcp.NewToolResultText("无法获取下游任务ID"), nil
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

// handleGetTaskDependenciesImpl 获取任务依赖
func (s *MCPServer) handleGetTaskDependenciesImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	// 先获取任务 ID
	taskResp, err := http.Get(fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), pathName))
	if err != nil {
		return mcp.NewToolResultText("获取任务信息失败：" + err.Error()), nil
	}
	defer taskResp.Body.Close()

	var task map[string]interface{}
	if err := json.NewDecoder(taskResp.Body).Decode(&task); err != nil {
		return mcp.NewToolResultText("解析任务信息失败：" + err.Error()), nil
	}

	taskID, ok := task["id"].(string)
	if !ok {
		return mcp.NewToolResultText("无法获取任务ID"), nil
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
