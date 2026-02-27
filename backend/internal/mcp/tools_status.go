package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleGetAllTaskStatus 获取所有任务状态
func (s *MCPServer) handleGetAllTaskStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID, _ := getParam(request, "projectId")
	if projectID == "" {
		projectID = s.configManager.GetProjectID()
	}
	if projectID == "" {
		return mcp.NewToolResultText("未配置项目ID，请先使用 init_project 或 set_project 设置项目"), nil
	}

	url := fmt.Sprintf("%s/tasks?projectId=%s&fields=id,name,moduleId,status", s.getApiURL(), projectID)
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

// handleGetModuleTaskStatus 获取模块任务状态
func (s *MCPServer) handleGetModuleTaskStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	url := fmt.Sprintf("%s/tasks?moduleId=%s&fields=id,name,status", s.getApiURL(), moduleID)
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

// handleGetProjectErrors 获取项目错误列表
func (s *MCPServer) handleGetProjectErrors(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID, _ := getParam(request, "projectId")
	if projectID == "" {
		projectID = s.configManager.GetProjectID()
	}
	if projectID == "" {
		return mcp.NewToolResultText("未配置项目ID，请先使用 init_project 或 set_project 设置项目"), nil
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
	if tasks, ok := result["data"].([]interface{}); ok {
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
						"taskId":   taskMap["id"],
						"taskName": taskMap["name"],
						"errors":   taskErrors,
					})
				}
			}
		}
	}

	resultData := map[string]interface{}{
		"projectId":  projectID,
		"scanTime":   time.Now().Format(time.RFC3339),
		"totalTasks": len(result["data"].([]interface{})),
		"errorCount": len(errors),
		"errors":     errors,
	}

	data, _ := json.MarshalIndent(resultData, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleGetModuleErrors 获取模块错误列表
func (s *MCPServer) handleGetModuleErrors(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
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
	if tasks, ok := result["data"].([]interface{}); ok {
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
						"taskId":   taskMap["id"],
						"taskName": taskMap["name"],
						"errors":   taskErrors,
					})
				}
			}
		}
	}

	resultData := map[string]interface{}{
		"moduleId":   moduleID,
		"scanTime":   time.Now().Format(time.RFC3339),
		"totalTasks": len(result["data"].([]interface{})),
		"errorCount": len(errors),
		"errors":     errors,
	}

	data, _ := json.MarshalIndent(resultData, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
