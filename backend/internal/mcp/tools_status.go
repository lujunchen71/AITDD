package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleGetAllTaskStatusImpl 获取所有任务状态
func (s *MCPServer) handleGetAllTaskStatusImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, _ := getParam(request, "pathName")
	if pathName == "" {
		pathName = s.configManager.GetProjectPathName()
	}
	if pathName == "" {
		return mcp.NewToolResultText("未配置项目路径名称，请先使用 init_project 或 set_project 设置项目"), nil
	}

	// 先获取项目 ID
	projectResp, err := http.Get(fmt.Sprintf("%s/projects/by-path/%s", s.getApiURL(), pathName))
	if err != nil {
		return mcp.NewToolResultText("获取项目信息失败：" + err.Error()), nil
	}
	defer projectResp.Body.Close()

	var project map[string]interface{}
	if err := json.NewDecoder(projectResp.Body).Decode(&project); err != nil {
		return mcp.NewToolResultText("解析项目信息失败：" + err.Error()), nil
	}

	projectID, ok := project["id"].(string)
	if !ok {
		return mcp.NewToolResultText("无法获取项目ID"), nil
	}

	url := fmt.Sprintf("%s/tasks?projectId=%s&fields=id,name,moduleId,status,pathName", s.getApiURL(), projectID)
	if status, ok := getParam(request, "status"); ok && status != "" {
		url += fmt.Sprintf("&status=%s", status)
	}
	if includeLockInfo, ok := getParamBool(request, "includeLockInfo"); ok && includeLockInfo {
		url += "&includeLockInfo=true"
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

// handleGetModuleTaskStatusImpl 获取模块任务状态
func (s *MCPServer) handleGetModuleTaskStatusImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	// 先获取模块 ID
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

	url := fmt.Sprintf("%s/tasks?moduleId=%s&fields=id,name,status,pathName", s.getApiURL(), moduleID)
	if includeLockInfo, ok := getParamBool(request, "includeLockInfo"); ok && includeLockInfo {
		url += "&includeLockInfo=true"
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

// handleGetProjectErrorsImpl 获取项目错误列表
func (s *MCPServer) handleGetProjectErrorsImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, _ := getParam(request, "pathName")
	if pathName == "" {
		pathName = s.configManager.GetProjectPathName()
	}
	if pathName == "" {
		return mcp.NewToolResultText("未配置项目路径名称，请先使用 init_project 或 set_project 设置项目"), nil
	}

	// 先获取项目 ID
	projectResp, err := http.Get(fmt.Sprintf("%s/projects/by-path/%s", s.getApiURL(), pathName))
	if err != nil {
		return mcp.NewToolResultText("获取项目信息失败：" + err.Error()), nil
	}
	defer projectResp.Body.Close()

	var project map[string]interface{}
	if err := json.NewDecoder(projectResp.Body).Decode(&project); err != nil {
		return mcp.NewToolResultText("解析项目信息失败：" + err.Error()), nil
	}

	projectID, ok := project["id"].(string)
	if !ok {
		return mcp.NewToolResultText("无法获取项目ID"), nil
	}

	// 获取所有任务
	resp, err := http.Get(fmt.Sprintf("%s/tasks?projectId=%s", s.getApiURL(), projectID))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	// 收集错误
	var errors []map[string]interface{}
	var totalTasks int
	if tasks, ok := result["data"].([]interface{}); ok {
		totalTasks = len(tasks)
		for _, task := range tasks {
			if taskMap, ok := task.(map[string]interface{}); ok {
				var taskErrors []map[string]interface{}

				// 检查 issue_details
				if issueDetails, ok := taskMap["issueDetails"].(string); ok && issueDetails != "" {
					taskErrors = append(taskErrors, map[string]interface{}{
						"type":    "issue",
						"message": issueDetails,
					})
				}

				// 检查 bug_log
				if bugLog, ok := taskMap["bugLog"].(string); ok && bugLog != "" && bugLog != "[]" {
					taskErrors = append(taskErrors, map[string]interface{}{
						"type":    "bug",
						"message": bugLog,
					})
				}

				if len(taskErrors) > 0 {
					errors = append(errors, map[string]interface{}{
						"taskPathName": taskMap["pathName"],
						"taskName":     taskMap["name"],
						"errors":       taskErrors,
					})
				}
			}
		}
	}

	resultData := map[string]interface{}{
		"projectPathName": pathName,
		"scanTime":        time.Now().Format(time.RFC3339),
		"totalTasks":      totalTasks,
		"errorCount":      len(errors),
		"errors":          errors,
	}

	data, _ := json.MarshalIndent(resultData, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleGetModuleErrorsImpl 获取模块错误列表
func (s *MCPServer) handleGetModuleErrorsImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	// 先获取模块 ID
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

	// 获取模块下的所有任务
	resp, err := http.Get(fmt.Sprintf("%s/tasks?moduleId=%s", s.getApiURL(), moduleID))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	// 收集错误
	var errors []map[string]interface{}
	var totalTasks int
	if tasks, ok := result["data"].([]interface{}); ok {
		totalTasks = len(tasks)
		for _, task := range tasks {
			if taskMap, ok := task.(map[string]interface{}); ok {
				var taskErrors []map[string]interface{}

				// 检查 issue_details
				if issueDetails, ok := taskMap["issueDetails"].(string); ok && issueDetails != "" {
					taskErrors = append(taskErrors, map[string]interface{}{
						"type":    "issue",
						"message": issueDetails,
					})
				}

				// 检查 bug_log
				if bugLog, ok := taskMap["bugLog"].(string); ok && bugLog != "" && bugLog != "[]" {
					taskErrors = append(taskErrors, map[string]interface{}{
						"type":    "bug",
						"message": bugLog,
					})
				}

				if len(taskErrors) > 0 {
					errors = append(errors, map[string]interface{}{
						"taskPathName": taskMap["pathName"],
						"taskName":     taskMap["name"],
						"errors":       taskErrors,
					})
				}
			}
		}
	}

	resultData := map[string]interface{}{
		"modulePathName": pathName,
		"scanTime":       time.Now().Format(time.RFC3339),
		"totalTasks":     totalTasks,
		"errorCount":     len(errors),
		"errors":         errors,
	}

	data, _ := json.MarshalIndent(resultData, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
