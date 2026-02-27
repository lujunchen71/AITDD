package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleGetProjectInfoImpl 获取项目简介
func (s *MCPServer) handleGetProjectInfoImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, _ := getParam(request, "pathName")
	
	if pathName == "" {
		pathName = s.configManager.GetProjectPathName()
	}
	if pathName == "" {
		return mcp.NewToolResultText("未配置项目路径名称，请先使用 init_project 或 set_project 设置项目，或提供 pathName"), nil
	}

	url := fmt.Sprintf("%s/projects/by-path/%s", s.getApiURL(), pathName)

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

// handleGetConstitutionImpl 获取项目公约
func (s *MCPServer) handleGetConstitutionImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName := s.configManager.GetProjectPathName()
	if pathName == "" {
		return mcp.NewToolResultText("未配置项目路径名称，请先使用 init_project 或 set_project 设置项目"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/projects/by-path/%s/constitution", s.getApiURL(), pathName))
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

// handleGetAllTaskCodePathsImpl 获取所有task代码路径
func (s *MCPServer) handleGetAllTaskCodePathsImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, _ := getParam(request, "pathName")
	if pathName == "" {
		pathName = s.configManager.GetProjectPathName()
	}
	if pathName == "" {
		return mcp.NewToolResultText("未配置项目路径名称，请先使用 init_project 或 set_project 设置项目"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/tasks?projectPathName=%s&fields=id,name,moduleId,codePaths,pathName", s.getApiURL(), pathName))
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

// handleGetAllModulesImpl 获取所有模块概要
func (s *MCPServer) handleGetAllModulesImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, _ := getParam(request, "pathName")
	if pathName == "" {
		pathName = s.configManager.GetProjectPathName()
	}
	if pathName == "" {
		return mcp.NewToolResultText("未配置项目路径名称，请先使用 init_project 或 set_project 设置项目"), nil
	}

	url := fmt.Sprintf("%s/modules?projectPathName=%s", s.getApiURL(), pathName)
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

// handleGetModuleOverviewImpl 获取模块概览
func (s *MCPServer) handleGetModuleOverviewImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	url := fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), pathName)

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

// handleGetModuleTaskPathNamesImpl 获取模块中所有任务的 pathName 列表
func (s *MCPServer) handleGetModuleTaskPathNamesImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	// 使用 action=tasks 参数
	resp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s?action=tasks", s.getApiURL(), pathName))
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

// handleGetModuleTasksImpl 获取模块任务列表
func (s *MCPServer) handleGetModuleTasksImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	// 使用 action=tasks 参数
	url := fmt.Sprintf("%s/modules/by-path/%s?action=tasks", s.getApiURL(), pathName)
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

// handleGetTaskDetailImpl 获取任务详情
func (s *MCPServer) handleGetTaskDetailImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	url := fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), pathName)

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

// handleGetTaskContractsImpl 获取任务上下游契约
func (s *MCPServer) handleGetTaskContractsImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	direction, _ := getParam(request, "direction")
	if direction == "" {
		direction = "both"
	}

	// 获取任务详情
	resp, err := http.Get(fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), pathName))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var task map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	// 获取任务ID用于查询依赖
	taskID, _ := task["id"].(string)
	
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
			"pathName": pathName,
			"name":     task["name"],
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
