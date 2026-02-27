package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleGetProjectInfo 获取项目简介
func (s *MCPServer) handleGetProjectInfo(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID, _ := getParam(request, "projectId")
	if projectID == "" {
		projectID = s.configManager.GetProjectID()
	}
	if projectID == "" {
		return mcp.NewToolResultText("未配置项目ID，请先使用 init_project 或 set_project 设置项目"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/projects/%s", s.getApiURL(), projectID))
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

// handleGetConstitution 获取项目公约
func (s *MCPServer) handleGetConstitution(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := s.configManager.GetProjectID()
	if projectID == "" {
		return mcp.NewToolResultText("未配置项目ID，请先使用 init_project 或 set_project 设置项目"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/projects/%s/constitution", s.getApiURL(), projectID))
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

// handleGetAllTaskCodePaths 获取所有task代码路径
func (s *MCPServer) handleGetAllTaskCodePaths(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID, _ := getParam(request, "projectId")
	if projectID == "" {
		projectID = s.configManager.GetProjectID()
	}
	if projectID == "" {
		return mcp.NewToolResultText("未配置项目ID，请先使用 init_project 或 set_project 设置项目"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/tasks?projectId=%s&fields=id,name,moduleId,codePaths", s.getApiURL(), projectID))
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

// handleGetAllModules 获取所有模块概要
func (s *MCPServer) handleGetAllModules(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID, _ := getParam(request, "projectId")
	if projectID == "" {
		projectID = s.configManager.GetProjectID()
	}
	if projectID == "" {
		return mcp.NewToolResultText("未配置项目ID，请先使用 init_project 或 set_project 设置项目"), nil
	}

	url := fmt.Sprintf("%s/modules?projectId=%s", s.getApiURL(), projectID)
	if includeStats, ok := getParamBool(request, "includeStats"); ok && includeStats {
		url += "&includeStats=true"
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

// handleGetModuleOverview 获取模块概览
func (s *MCPServer) handleGetModuleOverview(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/modules/%s", s.getApiURL(), moduleID))
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

// handleGetModuleTaskIDs 获取模块中所有任务 ID 列表
func (s *MCPServer) handleGetModuleTaskIDs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/modules/%s/tasks?fields=id", s.getApiURL(), moduleID))
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

// handleGetModuleTasks 获取模块任务列表
func (s *MCPServer) handleGetModuleTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	url := fmt.Sprintf("%s/tasks?moduleId=%s", s.getApiURL(), moduleID)
	if includeContracts, ok := getParamBool(request, "includeContracts"); ok && includeContracts {
		url += "&includeContracts=true"
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

// handleGetTaskDetail 获取任务详情
func (s *MCPServer) handleGetTaskDetail(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, ok := getParam(request, "taskId")
	if !ok {
		return mcp.NewToolResultText("缺少 taskId 参数"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/tasks/%s", s.getApiURL(), taskID))
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

// handleGetTaskContracts 获取任务上下游契约
func (s *MCPServer) handleGetTaskContracts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, ok := getParam(request, "taskId")
	if !ok {
		return mcp.NewToolResultText("缺少 taskId 参数"), nil
	}

	direction, _ := getParam(request, "direction")
	if direction == "" {
		direction = "both"
	}

	// 获取任务详情
	resp, err := http.Get(fmt.Sprintf("%s/tasks/%s", s.getApiURL(), taskID))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var task map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	// 获取任务依赖
	depResp, err := http.Get(fmt.Sprintf("%s/dependencies?taskId=%s", s.getApiURL(), taskID))
	if err != nil {
		return mcp.NewToolResultText("获取依赖失败：" + err.Error()), nil
	}
	defer depResp.Body.Close()

	var deps map[string]interface{}
	if err := json.NewDecoder(depResp.Body).Decode(&deps); err != nil {
		return mcp.NewToolResultText("解析依赖失败：" + err.Error()), nil
	}

	// 构建结果
	result := map[string]interface{}{
		"task": map[string]interface{}{
			"id":   task["id"],
			"name": task["name"],
		},
		"upstream": map[string]interface{}{
			"title":     "依赖上游任务提供",
			"contracts": task["upstreamContractDetail"],
		},
		"downstream": map[string]interface{}{
			"title":     "为下游任务提供以下接口",
			"contracts": task["downstreamContractDetail"],
		},
		"dependencies": deps,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
