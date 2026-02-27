package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleDeleteModule 删除模块
func (s *MCPServer) handleDeleteModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/modules/%s", s.getApiURL(), moduleID), nil)
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

// handleCreateModule 创建模块
func (s *MCPServer) handleCreateModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, ok := getParam(request, "name")
	if !ok {
		return mcp.NewToolResultText("缺少 name 参数"), nil
	}

	projectID, _ := getParam(request, "projectId")
	if projectID == "" {
		projectID = s.configManager.GetProjectID()
	}
	if projectID == "" {
		return mcp.NewToolResultText("未配置项目ID，请先使用 init_project 或 set_project 设置项目"), nil
	}

	// 构建创建数据
	createData := map[string]interface{}{
		"name":      name,
		"projectId": projectID,
		"status":    "designing",
	}
	if description, ok := getParamAny(request, "description"); ok {
		createData["description"] = description
	}
	if prompt, ok := getParamAny(request, "prompt"); ok {
		createData["prompt"] = prompt
	}
	if parentId, ok := getParamAny(request, "parentId"); ok {
		createData["parentId"] = parentId
	}

	jsonData, _ := json.Marshal(createData)
	resp, err := http.Post(fmt.Sprintf("%s/modules", s.getApiURL()), "application/json", bytes.NewBuffer(jsonData))
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

// handleUpdateModule 更新模块
func (s *MCPServer) handleUpdateModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
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
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/modules/%s", s.getApiURL(), moduleID), bytes.NewBuffer(jsonData))
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

// handleDeleteModuleTasks 删除模块所有任务
func (s *MCPServer) handleDeleteModuleTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	// 获取模块下的所有任务
	resp, err := http.Get(fmt.Sprintf("%s/tasks?moduleId=%s", s.getApiURL(), moduleID))
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
			taskID, ok := taskMap["id"].(string)
			if !ok {
				continue
			}

			req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/tasks/%s", s.getApiURL(), taskID), nil)
			delResp, err := http.DefaultClient.Do(req)
			if err != nil {
				errors = append(errors, fmt.Sprintf("删除任务 %s 失败: %s", taskID, err.Error()))
				continue
			}
			delResp.Body.Close()
			deletedCount++
			deletedTasks = append(deletedTasks, map[string]interface{}{
				"id":   taskID,
				"name": taskMap["name"],
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

// handleCreateTask 创建任务
func (s *MCPServer) handleCreateTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	name, ok := getParam(request, "name")
	if !ok {
		return mcp.NewToolResultText("缺少 name 参数"), nil
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

// handleUpdateModuleFull 完整更新模块
func (s *MCPServer) handleUpdateModuleFull(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
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

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/modules/%s", s.getApiURL(), moduleID), bytes.NewBuffer(jsonData))
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

// handleUpdateTaskFull 完整更新任务
func (s *MCPServer) handleUpdateTaskFull(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, ok := getParam(request, "taskId")
	if !ok {
		return mcp.NewToolResultText("缺少 taskId 参数"), nil
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

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/tasks/%s", s.getApiURL(), taskID), bytes.NewBuffer(jsonData))
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

// handleUpdateTask 更新任务
func (s *MCPServer) handleUpdateTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, ok := getParam(request, "taskId")
	if !ok {
		return mcp.NewToolResultText("缺少 taskId 参数"), nil
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
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/tasks/%s", s.getApiURL(), taskID), bytes.NewBuffer(jsonData))
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

// handleDeleteTask 删除任务
func (s *MCPServer) handleDeleteTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, ok := getParam(request, "taskId")
	if !ok {
		return mcp.NewToolResultText("缺少 taskId 参数"), nil
	}

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/tasks/%s", s.getApiURL(), taskID), nil)
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
