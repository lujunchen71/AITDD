package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// ==================== 4. 节点操作实现 ====================

// generatePathName 根据 parentPath 和 name 自动生成 pathName
func generatePathName(parentPath, name string) string {
	// 转换为小写，替换空格为下划线
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	return parentPath + "/" + slug
}

// handleCreateNodeImpl 统一创建节点
func (s *MCPServer) handleCreateNodeImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeType, ok := getParam(request, "type")
	if !ok || nodeType == "" {
		return mcp.NewToolResultText("缺少 type 参数（module 或 task）"), nil
	}

	parentPath, ok := getParam(request, "parentPath")
	if !ok || parentPath == "" {
		return mcp.NewToolResultText("缺少 parentPath 参数"), nil
	}

	name, ok := getParam(request, "name")
	if !ok || name == "" {
		return mcp.NewToolResultText("缺少 name 参数"), nil
	}

	// 获取可选的 pathName，不传则自动生成
	pathName, _ := getParam(request, "pathName")
	if pathName == "" {
		pathName = generatePathName(parentPath, name)
	}

	// 获取 data 参数
	data, _ := request.Params.Arguments.(map[string]any)["data"].(map[string]interface{})

	switch nodeType {
	case "module":
		return s.createModuleFromNode(pathName, parentPath, name, data)
	case "task":
		return s.createTaskFromNode(pathName, parentPath, name, data)
	default:
		return mcp.NewToolResultText(fmt.Sprintf("不支持的节点类型: %s（支持: module, task）", nodeType)), nil
	}
}

// createModuleFromNode 从统一接口创建模块
func (s *MCPServer) createModuleFromNode(pathName, parentPath, name string, data map[string]interface{}) (*mcp.CallToolResult, error) {
	// 从配置文件获取项目 pathName
	projectPathName := s.configManager.GetProjectPathName()
	if projectPathName == "" {
		return mcp.NewToolResultText("未配置项目路径名称，请先使用 init_project 设置项目"), nil
	}

	// 获取项目 ID
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
		"pathName":  pathName,
	}

	// 从 data 中提取模块字段
	if data != nil {
		if description, ok := data["description"].(string); ok {
			createData["description"] = description
		}
		if prompt, ok := data["prompt"].(string); ok {
			createData["prompt"] = prompt
		}
		if upstreamContract, ok := data["upstreamContractSummary"].(string); ok {
			createData["upstreamContractSummary"] = upstreamContract
		}
		if downstreamContract, ok := data["downstreamContractSummary"].(string); ok {
			createData["downstreamContractSummary"] = downstreamContract
		}
	}

	// 如果 parentPath 不是项目路径，则认为是子模块
	if parentPath != projectPathName {
		resp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), parentPath))
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

	// 返回统一格式 (YAML 风格)
	success := resp2.StatusCode == http.StatusOK || resp2.StatusCode == http.StatusCreated
	actualPathName := pathName
	if module, ok := result["module"].(map[string]interface{}); ok {
		if pn, ok := module["pathName"].(string); ok {
			actualPathName = pn
		}
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("success: %v\n", success))
	output.WriteString(fmt.Sprintf("pathName: \"%s\"", actualPathName))

	if !success {
		errorBytes, _ := json.Marshal(result)
		output.WriteString(fmt.Sprintf("\nerror: %s", string(errorBytes)))
	}

	return mcp.NewToolResultText(output.String()), nil
}

// createTaskFromNode 从统一接口创建任务
func (s *MCPServer) createTaskFromNode(pathName, parentPath, name string, data map[string]interface{}) (*mcp.CallToolResult, error) {
	// parentPath 是模块路径，需要获取模块 ID
	moduleResp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), parentPath))
	if err != nil {
		return mcp.NewToolResultText("获取模块信息失败：" + err.Error()), nil
	}
	defer moduleResp.Body.Close()

	var moduleResult map[string]interface{}
	if err := json.NewDecoder(moduleResp.Body).Decode(&moduleResult); err != nil {
		return mcp.NewToolResultText("解析模块信息失败：" + err.Error()), nil
	}

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
		"pathName": pathName,
	}

	// 从 data 中提取任务字段
	if data != nil {
		if description, ok := data["description"].(string); ok {
			createData["description"] = description
		}
		if prompt, ok := data["prompt"].(string); ok {
			createData["prompt"] = prompt
		}
		if upstreamContract, ok := data["upstreamContractDetail"].(string); ok {
			createData["upstreamContractDetail"] = upstreamContract
		}
		if downstreamContract, ok := data["downstreamContractDetail"].(string); ok {
			createData["downstreamContractDetail"] = downstreamContract
		}
		if tests, ok := data["tests"].(string); ok {
			createData["tests"] = tests
		}
		if codePaths, ok := data["codePaths"].([]interface{}); ok {
			createData["codePaths"] = codePaths
		} else if codePathsStr, ok := data["codePaths"].(string); ok {
			// 支持字符串格式的 JSON 数组
			var codePathsArr []interface{}
			if json.Unmarshal([]byte(codePathsStr), &codePathsArr) == nil {
				createData["codePaths"] = codePathsArr
			}
		}
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

	// 返回统一格式 (YAML 风格)
	success := resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated
	actualPathName := pathName
	if task, ok := result["task"].(map[string]interface{}); ok {
		if pn, ok := task["pathName"].(string); ok {
			actualPathName = pn
		}
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("success: %v\n", success))
	output.WriteString(fmt.Sprintf("pathName: \"%s\"", actualPathName))

	if !success {
		errorBytes, _ := json.Marshal(result)
		output.WriteString(fmt.Sprintf("\nerror: %s", string(errorBytes)))
	}

	return mcp.NewToolResultText(output.String()), nil
}

// handleModifyNodeImpl 统一修改节点
func (s *MCPServer) handleModifyNodeImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, ok := getParam(request, "path")
	if !ok || path == "" {
		return mcp.NewToolResultText("缺少 path 参数"), nil
	}

	version, ok := getParamFloat(request, "version")
	if !ok {
		return mcp.NewToolResultText("缺少 version 参数"), nil
	}

	data, ok := request.Params.Arguments.(map[string]any)["data"].(map[string]interface{})
	if !ok || len(data) == 0 {
		return mcp.NewToolResultText("缺少 data 参数或 data 为空"), nil
	}

	// 检测节点类型
	nodeType := s.detectNodeType(path)

	// 添加版本号到更新数据
	updateData := make(map[string]interface{})
	for k, v := range data {
		updateData[k] = v
	}
	updateData["version"] = int(version)

	var apiPath string
	switch nodeType {
	case "project":
		apiPath = fmt.Sprintf("%s/projects/by-path/%s", s.getApiURL(), path)
	case "module":
		apiPath = fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), path)
	case "task":
		apiPath = fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), path)
	default:
		return mcp.NewToolResultText(fmt.Sprintf("无法识别的节点类型: %s", nodeType)), nil
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", apiPath, bytes.NewBuffer(jsonData))
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

	// 返回统一格式 (YAML 风格)
	success := resp.StatusCode == http.StatusOK
	var newVersion int

	if success {
		// 尝试从响应中获取新版本号
		var nodeData map[string]interface{}
		switch nodeType {
		case "project":
			nodeData, _ = result["project"].(map[string]interface{})
		case "module":
			nodeData, _ = result["module"].(map[string]interface{})
		case "task":
			nodeData, _ = result["task"].(map[string]interface{})
		}
		if nodeData != nil {
			if v, ok := nodeData["version"].(float64); ok {
				newVersion = int(v)
			}
		}
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("success: %v\n", success))
	output.WriteString(fmt.Sprintf("version: %d", newVersion))

	if !success {
		errorBytes, _ := json.Marshal(result)
		output.WriteString(fmt.Sprintf("\nerror: %s", string(errorBytes)))
	}

	return mcp.NewToolResultText(output.String()), nil
}

// handleDeleteNodeImpl 统一删除节点
func (s *MCPServer) handleDeleteNodeImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, ok := getParam(request, "path")
	if !ok || path == "" {
		return mcp.NewToolResultText("缺少 path 参数"), nil
	}

	force, _ := getParamBool(request, "force")

	// 检测节点类型
	nodeType := s.detectNodeType(path)

	var apiPath string
	deletedChildren := 0

	switch nodeType {
	case "module":
		// 删除模块时，先统计子任务数量
		resp, err := http.Get(fmt.Sprintf("%s/tasks?modulePathName=%s", s.getApiURL(), path))
		if err == nil {
			var result map[string]interface{}
			if json.NewDecoder(resp.Body).Decode(&result) == nil {
				if tasks, ok := result["data"].([]interface{}); ok {
					deletedChildren = len(tasks)
				}
			}
			resp.Body.Close()
		}

		apiPath = fmt.Sprintf("%s/modules/by-path/%s?force=%v", s.getApiURL(), path, force)
	case "task":
		apiPath = fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), path)
	case "project":
		return mcp.NewToolResultText("不支持删除项目，请使用 delete_node 删除模块"), nil
	default:
		return mcp.NewToolResultText(fmt.Sprintf("无法识别的节点类型: %s", nodeType)), nil
	}

	req, _ := http.NewRequest("DELETE", apiPath, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	// 返回统一格式 (YAML 风格)
	success := resp.StatusCode == http.StatusOK

	var output strings.Builder
	output.WriteString(fmt.Sprintf("success: %v\n", success))
	output.WriteString(fmt.Sprintf("deletedChildren: %d", deletedChildren))

	if !success {
		errorBytes, _ := json.Marshal(result)
		output.WriteString(fmt.Sprintf("\nerror: %s", string(errorBytes)))
	}

	return mcp.NewToolResultText(output.String()), nil
}

// getParamFloat 从请求参数中获取浮点数值
func getParamFloat(request mcp.CallToolRequest, key string) (float64, bool) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return 0, false
	}
	val, exists := args[key]
	if !exists {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	default:
		return 0, false
	}
}

// ==================== modify_module 和 modify_task 实现 ====================

// Module 可修改字段:
// - name: 模块名称 (string)
// - description: 模块描述 (string)
// - status: 状态 (string: pending/in_progress/completed/blocked)
// - prompt: 提示词内容 (string)
// - upstreamContractSummary: 上游契约摘要 (string)
// - downstreamContractSummary: 下游契约摘要 (string)
// - testCoverage: 测试覆盖率 (number)

// handleModifyModuleImpl 修改模块信息
// 参数:
//   - pathName: 模块路径名称 (必填)，格式: "项目名/模块名"
//   - version: 当前版本号 (必填)，用于乐观锁
//   - data: 要修改的字段键值对 (必填)
// 可修改字段: name, description, status, prompt, upstreamContractSummary, downstreamContractSummary, testCoverage
func (s *MCPServer) handleModifyModuleImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	version, ok := getParamFloat(request, "version")
	if !ok {
		return mcp.NewToolResultText("缺少 version 参数"), nil
	}

	data, ok := request.Params.Arguments.(map[string]any)["data"].(map[string]interface{})
	if !ok || len(data) == 0 {
		return mcp.NewToolResultText("缺少 data 参数或 data 为空"), nil
	}

	// 构建更新数据
	updateData := make(map[string]interface{})
	for k, v := range data {
		updateData[k] = v
	}
	updateData["version"] = int(version)

	// 调用 API 更新模块
	apiPath := fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), pathName)
	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", apiPath, bytes.NewBuffer(jsonData))
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

	// 返回统一格式 (YAML 风格)
	success := resp.StatusCode == http.StatusOK
	var newVersion int

	if success {
		// 尝试从响应中获取新版本号
		var moduleData map[string]interface{}
		if data, ok := result["data"].(map[string]interface{}); ok {
			moduleData, _ = data["module"].(map[string]interface{})
		}
		if moduleData == nil {
			moduleData, _ = result["module"].(map[string]interface{})
		}
		if moduleData != nil {
			if v, ok := moduleData["version"].(float64); ok {
				newVersion = int(v)
			}
		}
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("success: %v\n", success))
	output.WriteString(fmt.Sprintf("version: %d", newVersion))

	if !success {
		errorBytes, _ := json.Marshal(result)
		output.WriteString(fmt.Sprintf("\nerror: %s", string(errorBytes)))
	}

	return mcp.NewToolResultText(output.String()), nil
}

// Task 可修改字段:
// - name: 任务名称 (string)
// - description: 任务描述 (string)
// - status: 状态 (string: pending/in_progress/completed/blocked)
// - prompt: 提示词内容 (string)
// - upstreamContractDetail: 上游契约详情 (object: {title:string, list:[{label:string, contract_api:string, from:string}]})
// - downstreamContractDetail: 下游契约详情 (object: {title:string, list:[{label:string, contract_api:string, from:string}]})
// - tests: 测试用例列表 (array: [{target:string, api:string}])
// - testResult: 测试结果 (string: pass/fail/pending)
// - codePaths: 代码文件路径列表 (array: ["path/to/file1.ts", "path/to/file2.ts"])
// - bugLog: bug 日志 (string)
// - humanAssistance: 是否需要人工协助 (boolean)
// - issueDetails: 问题详情 (string)

// handleModifyTaskImpl 修改任务信息
// 参数:
//   - pathName: 任务路径名称 (必填)，格式: "项目名/模块名/任务名"
//   - version: 当前版本号 (必填)，用于乐观锁
//   - data: 要修改的字段键值对 (必填)
// 可修改字段: name, description, status, prompt, upstreamContractDetail, downstreamContractDetail, tests, testResult, codePaths, bugLog, humanAssistance, issueDetails
func (s *MCPServer) handleModifyTaskImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	version, ok := getParamFloat(request, "version")
	if !ok {
		return mcp.NewToolResultText("缺少 version 参数"), nil
	}

	data, ok := request.Params.Arguments.(map[string]any)["data"].(map[string]interface{})
	if !ok || len(data) == 0 {
		return mcp.NewToolResultText("缺少 data 参数或 data 为空"), nil
	}

	// 构建更新数据
	updateData := make(map[string]interface{})
	for k, v := range data {
		// 对需要存储为 JSON 字符串的字段进行序列化
		// 这些字段在数据库中是 TEXT 类型，存储 JSON 字符串
		switch k {
		case "tests", "codePaths", "bugLog", "testResult":
			// 处理数组类型
			if arr, ok := v.([]interface{}); ok {
				jsonBytes, _ := json.Marshal(arr)
				updateData[k] = string(jsonBytes)
			} else if m, ok := v.(map[string]interface{}); ok {
				// 处理对象类型
				jsonBytes, _ := json.Marshal(m)
				updateData[k] = string(jsonBytes)
			} else if str, ok := v.(string); ok {
				// 已经是字符串，直接使用
				updateData[k] = str
			} else {
				updateData[k] = v
			}
		case "upstreamContractDetail", "downstreamContractDetail", "humanAssistance", "issueDetails":
			// 这些字段也可能是对象类型，需要序列化为 JSON 字符串
			if m, ok := v.(map[string]interface{}); ok {
				jsonBytes, _ := json.Marshal(m)
				updateData[k] = string(jsonBytes)
			} else if arr, ok := v.([]interface{}); ok {
				jsonBytes, _ := json.Marshal(arr)
				updateData[k] = string(jsonBytes)
			} else if str, ok := v.(string); ok {
				updateData[k] = str
			} else {
				updateData[k] = v
			}
		default:
			updateData[k] = v
		}
	}
	updateData["version"] = int(version)

	// 调用 API 更新任务
	apiPath := fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), pathName)
	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", apiPath, bytes.NewBuffer(jsonData))
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

	// 返回统一格式 (YAML 风格)
	success := resp.StatusCode == http.StatusOK
	var newVersion int

	if success {
		// 尝试从响应中获取新版本号
		var taskData map[string]interface{}
		if data, ok := result["data"].(map[string]interface{}); ok {
			taskData, _ = data["task"].(map[string]interface{})
		}
		if taskData == nil {
			taskData, _ = result["task"].(map[string]interface{})
		}
		if taskData != nil {
			if v, ok := taskData["version"].(float64); ok {
				newVersion = int(v)
			}
		}
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("success: %v\n", success))
	output.WriteString(fmt.Sprintf("version: %d", newVersion))

	if !success {
		errorBytes, _ := json.Marshal(result)
		output.WriteString(fmt.Sprintf("\nerror: %s", string(errorBytes)))
	}

	return mcp.NewToolResultText(output.String()), nil
}
