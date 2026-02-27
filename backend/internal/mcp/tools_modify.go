package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleDeleteModuleImpl 删除模块
func (s *MCPServer) handleDeleteModuleImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), pathName), nil)
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

// handleCreateModuleImpl 创建模块
func (s *MCPServer) handleCreateModuleImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, ok := getParam(request, "name")
	if !ok {
		return mcp.NewToolResultText("缺少 name 参数"), nil
	}

	// 从配置文件获取项目 pathName
	projectPathName := s.configManager.GetProjectPathName()
	if projectPathName == "" {
		return mcp.NewToolResultText("未配置项目路径名称，请先使用 init_project 或 set_project 设置项目"), nil
	}

	// 通过 pathName 获取项目信息（包含 ID）
	resp, err := http.Get(fmt.Sprintf("%s/projects/by-path/%s", s.getApiURL(), projectPathName))
	if err != nil {
		return mcp.NewToolResultText("获取项目信息失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var projectResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&projectResult); err != nil {
		return mcp.NewToolResultText("解析项目信息失败：" + err.Error()), nil
	}

	project, ok := projectResult["project"].(map[string]interface{})
	if !ok {
		return mcp.NewToolResultText("项目不存在或格式错误"), nil
	}
	projectID, ok := project["id"].(string)
	if !ok {
		return mcp.NewToolResultText("项目ID不存在"), nil
	}

	// 构建创建数据
	createData := map[string]interface{}{
		"name":      name,
		"projectId": projectID,
	}
	if description, ok := getParamAny(request, "description"); ok {
		createData["description"] = description
	}
	if prompt, ok := getParamAny(request, "prompt"); ok {
		createData["prompt"] = prompt
	}
	// pathName 参数用于指定新模块的路径名称
	if pathName, ok := getParam(request, "pathName"); ok && pathName != "" {
		createData["pathName"] = pathName
	}
	// parentPathName 参数用于指定父模块
	if parentPathName, ok := getParam(request, "parentPathName"); ok && parentPathName != "" {
		// 需要先通过 pathName 获取父模块 ID
		resp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), parentPathName))
		if err == nil {
			defer resp.Body.Close()
			var result map[string]interface{}
			if json.NewDecoder(resp.Body).Decode(&result) == nil {
				if module, ok := result["module"].(map[string]interface{}); ok {
					if parentID, ok := module["id"].(string); ok {
						createData["parentId"] = parentID
					}
				}
			}
		}
	}

	jsonData, _ := json.Marshal(createData)
	resp2, err := http.Post(fmt.Sprintf("%s/modules", s.getApiURL()), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp2.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp2.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleUpdateModuleImpl 更新模块
func (s *MCPServer) handleUpdateModuleImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	// 构建更新数据
	updateData := make(map[string]interface{})
	if name, ok := getParamAny(request, "name"); ok {
		updateData["name"] = name
	}
	if description, ok := getParamAny(request, "description"); ok {
		updateData["description"] = description
	}
	if prompt, ok := getParamAny(request, "prompt"); ok {
		updateData["prompt"] = prompt
	}
	if status, ok := getParamAny(request, "status"); ok {
		updateData["status"] = status
	}
	if version, ok := getParamInt(request, "version"); ok {
		updateData["version"] = version
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), pathName), bytes.NewBuffer(jsonData))
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

// handleDeleteModuleTasksImpl 删除模块所有任务
func (s *MCPServer) handleDeleteModuleTasksImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	// 获取模块下的所有任务
	resp, err := http.Get(fmt.Sprintf("%s/tasks?modulePathName=%s", s.getApiURL(), pathName))
	if err != nil {
		return mcp.NewToolResultText("获取任务列表失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析任务列表失败：" + err.Error()), nil
	}

	// 获取任务列表
	tasks, ok := result["data"].([]interface{})
	if !ok {
		return mcp.NewToolResultText("任务列表格式错误"), nil
	}

	// 逐个删除任务
	deletedCount := 0
	var deletedTasks []map[string]interface{}
	var errors []string

	for _, task := range tasks {
		if taskMap, ok := task.(map[string]interface{}); ok {
			taskPathName, ok := taskMap["pathName"].(string)
			if !ok {
				// 尝试使用 ID
				taskID, okID := taskMap["id"].(string)
				if !okID {
					continue
				}
				taskPathName = taskID
			}

			req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), taskPathName), nil)
			delResp, err := http.DefaultClient.Do(req)
			if err != nil {
				errors = append(errors, fmt.Sprintf("删除任务 %s 失败: %s", taskPathName, err.Error()))
				continue
			}
			delResp.Body.Close()
			deletedCount++
			deletedTasks = append(deletedTasks, map[string]interface{}{
				"pathName": taskPathName,
				"name":     taskMap["name"],
			})
		}
	}

	resultData := map[string]interface{}{
		"success":      len(errors) == 0,
		"message":      fmt.Sprintf("删除完成，共删除 %d 个任务", deletedCount),
		"deletedCount": deletedCount,
		"deletedTasks": deletedTasks,
	}
	if len(errors) > 0 {
		resultData["errors"] = errors
	}

	data, _ := json.MarshalIndent(resultData, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleCreateTaskImpl 创建任务
func (s *MCPServer) handleCreateTaskImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	modulePathName, ok := getParam(request, "pathName")
	if !ok || modulePathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数（模块路径名称）"), nil
	}

	name, ok := getParam(request, "name")
	if !ok {
		return mcp.NewToolResultText("缺少 name 参数"), nil
	}

	// 先通过 pathName 获取模块 ID
	moduleResp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), modulePathName))
	if err != nil {
		return mcp.NewToolResultText("获取模块信息失败：" + err.Error()), nil
	}
	defer moduleResp.Body.Close()

	var moduleResult map[string]interface{}
	if err := json.NewDecoder(moduleResp.Body).Decode(&moduleResult); err != nil {
		return mcp.NewToolResultText("解析模块信息失败：" + err.Error()), nil
	}

	// API 返回格式是 {"module": {...}}
	moduleData, ok := moduleResult["module"].(map[string]interface{})
	if !ok {
		return mcp.NewToolResultText("模块数据格式错误"), nil
	}
	moduleID, ok := moduleData["id"].(string)
	if !ok {
		return mcp.NewToolResultText("无法获取模块ID"), nil
	}

	// 构建创建数据
	createData := map[string]interface{}{
		"moduleId": moduleID,
		"name":     name,
		"status":   "ready",
	}
	if description, ok := getParamAny(request, "description"); ok {
		createData["description"] = description
	}
	if prompt, ok := getParamAny(request, "prompt"); ok {
		createData["prompt"] = prompt
	}
	if upstreamContract, ok := getParamAny(request, "upstreamContractDetail"); ok {
		createData["upstreamContractDetail"] = upstreamContract
	}
	if downstreamContract, ok := getParamAny(request, "downstreamContractDetail"); ok {
		createData["downstreamContractDetail"] = downstreamContract
	}
	if tests, ok := getParamAny(request, "tests"); ok {
		createData["tests"] = tests
	}
	if codePaths, ok := getParamAny(request, "codePaths"); ok {
		createData["codePaths"] = codePaths
	}

	jsonData, _ := json.Marshal(createData)
	resp, err := http.Post(fmt.Sprintf("%s/tasks", s.getApiURL()), "application/json", bytes.NewBuffer(jsonData))
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

// handleUpdateModuleFullImpl 完整更新模块
func (s *MCPServer) handleUpdateModuleFullImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("必须提供 pathName 参数"), nil
	}

	moduleJson, ok := getParam(request, "moduleJson")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleJson 参数"), nil
	}

	// 解析JSON
	var updateData map[string]interface{}
	if err := json.Unmarshal([]byte(moduleJson), &updateData); err != nil {
		return mcp.NewToolResultText("解析 moduleJson 失败：" + err.Error()), nil
	}

	url := fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), pathName)

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
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

// handleUpdateTaskFullImpl 完整更新任务
func (s *MCPServer) handleUpdateTaskFullImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("必须提供 pathName 参数"), nil
	}

	taskJson, ok := getParam(request, "taskJson")
	if !ok {
		return mcp.NewToolResultText("缺少 taskJson 参数"), nil
	}

	// 解析JSON
	var updateData map[string]interface{}
	if err := json.Unmarshal([]byte(taskJson), &updateData); err != nil {
		return mcp.NewToolResultText("解析 taskJson 失败：" + err.Error()), nil
	}

	url := fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), pathName)

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
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

// handleUpdateTaskImpl 更新任务
func (s *MCPServer) handleUpdateTaskImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	// 构建更新数据
	updateData := make(map[string]interface{})
	if name, ok := getParamAny(request, "name"); ok {
		updateData["name"] = name
	}
	if description, ok := getParamAny(request, "description"); ok {
		updateData["description"] = description
	}
	if status, ok := getParamAny(request, "status"); ok {
		updateData["status"] = status
	}
	if prompt, ok := getParamAny(request, "prompt"); ok {
		updateData["prompt"] = prompt
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), pathName), bytes.NewBuffer(jsonData))
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

// handleDeleteTaskImpl 删除任务
func (s *MCPServer) handleDeleteTaskImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), pathName), nil)
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
