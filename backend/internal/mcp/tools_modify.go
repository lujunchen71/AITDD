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
// 重要：pathName 必须与 name 保持一致，格式为 parentPath + "/" + name
// 名称会被转换为小写并用下划线替换空格
func generatePathName(parentPath, name string) string {
	// 将名称转换为小写并用下划线替换空格
	normalizedName := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	return parentPath + "/" + normalizedName
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

	// 验证路径是否属于当前项目
	if errMsg := s.configManager.ValidatePathBelongsToProject(parentPath); errMsg != "" {
		return mcp.NewToolResultText(errMsg), nil
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

	// 从嵌套的 data 结构中获取 project
	var project map[string]interface{}
	if data, ok := projectResult["data"].(map[string]interface{}); ok {
		project, _ = data["project"].(map[string]interface{})
	}
	if project == nil {
		project, _ = projectResult["project"].(map[string]interface{})
	}
	if project == nil {
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

	// 从嵌套的 data 结构中获取 module
	var moduleData map[string]interface{}
	if data, ok := moduleResult["data"].(map[string]interface{}); ok {
		moduleData, _ = data["module"].(map[string]interface{})
	}
	if moduleData == nil {
		moduleData, _ = moduleResult["module"].(map[string]interface{})
	}
	if moduleData == nil {
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

	// 验证路径是否属于当前项目
	if errMsg := s.configManager.ValidatePathBelongsToProject(path); errMsg != "" {
		return mcp.NewToolResultText(errMsg), nil
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

// ModifyOperation 修改操作结构
type ModifyOperation struct {
	PathName string
	Version  int
	Data     map[string]interface{}
}

// VersionCheckResult 版本检查结果
type VersionCheckResult struct {
	PathName        string
	ProvidedVersion int
	CurrentVersion  int
	Status          string // "ok" 或 "conflict"
}

// handleModifyModuleImpl 批量修改模块信息
// 参数:
//   - operations: 批量操作数组，每项包含 {pathName: string, version: number, data: object}
// 原子性保证：预检所有版本号，任一冲突则全部拒绝
func (s *MCPServer) handleModifyModuleImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取 operations 参数
	operationsVal, ok := getParamAny(request, "operations")
	if !ok || operationsVal == nil {
		return mcp.NewToolResultText("缺少 operations 参数"), nil
	}

	// 解析 operations 数组
	var operations []ModifyOperation
	if operationsArr, ok := operationsVal.([]interface{}); ok {
		if len(operationsArr) == 0 {
			return mcp.NewToolResultText("operations 数组不能为空"), nil
		}
		for _, op := range operationsArr {
			if opMap, ok := op.(map[string]interface{}); ok {
				operation := ModifyOperation{}
				if pathName, ok := opMap["pathName"].(string); ok {
					operation.PathName = pathName
				}
				// 支持多种数字类型
				switch v := opMap["version"].(type) {
				case float64:
					operation.Version = int(v)
				case int:
					operation.Version = v
				case int64:
					operation.Version = int(v)
				}
				if data, ok := opMap["data"].(map[string]interface{}); ok {
					operation.Data = data
				}
				if operation.PathName != "" && operation.Version > 0 && len(operation.Data) > 0 {
					operations = append(operations, operation)
				}
			}
		}
	}

	if len(operations) == 0 {
		return mcp.NewToolResultText("未能解析有效的操作项"), nil
	}

	// 验证所有 pathName 是否属于当前项目
	for _, op := range operations {
		if errMsg := s.configManager.ValidatePathBelongsToProject(op.PathName); errMsg != "" {
			return mcp.NewToolResultText(errMsg), nil
		}
	}

	// 阶段1：预检所有版本号
	var checks []VersionCheckResult
	var conflicts []VersionCheckResult

	for _, op := range operations {
		check := VersionCheckResult{
			PathName:        op.PathName,
			ProvidedVersion: op.Version,
		}

		// 获取当前模块版本
		url := fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), op.PathName)
		resp, err := http.Get(url)
		if err != nil {
			check.Status = "error: " + err.Error()
			checks = append(checks, check)
			conflicts = append(conflicts, check)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 404 {
			check.Status = "error: 模块不存在"
			checks = append(checks, check)
			conflicts = append(conflicts, check)
			continue
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			check.Status = "error: 解析失败"
			checks = append(checks, check)
			conflicts = append(conflicts, check)
			continue
		}

		// 提取当前版本
		var moduleData map[string]interface{}
		if data, ok := result["data"].(map[string]interface{}); ok {
			moduleData, _ = data["module"].(map[string]interface{})
		}
		if moduleData == nil {
			moduleData, _ = result["module"].(map[string]interface{})
		}
		if moduleData == nil {
			moduleData = result
		}

		if v, ok := moduleData["version"].(float64); ok {
			check.CurrentVersion = int(v)
			if check.CurrentVersion == check.ProvidedVersion {
				check.Status = "版本正常"
			} else {
				check.Status = "版本冲突"
				conflicts = append(conflicts, check)
			}
		} else {
			check.Status = "error: 无法获取版本"
			conflicts = append(conflicts, check)
		}
		checks = append(checks, check)
	}

	// 阶段2：决策 - 如果有任何冲突，拒绝整个批量操作
	if len(conflicts) > 0 {
		var output strings.Builder
		output.WriteString("success: false\n")
		output.WriteString("error: VERSION_CONFLICT\n")
		output.WriteString(fmt.Sprintf("message: 批量修改已取消：发现 %d 个版本冲突\n", len(conflicts)))
		output.WriteString("conflicts:\n")
		for _, c := range conflicts {
			output.WriteString(fmt.Sprintf("  - pathName: \"%s\"\n", c.PathName))
			output.WriteString(fmt.Sprintf("    providedVersion: %d\n", c.ProvidedVersion))
			output.WriteString(fmt.Sprintf("    currentVersion: %d\n", c.CurrentVersion))
			output.WriteString(fmt.Sprintf("    message: \"%s\"\n", c.Status))
		}
		output.WriteString("checks:\n")
		for _, c := range checks {
			output.WriteString(fmt.Sprintf("  - pathName: \"%s\"\n", c.PathName))
			output.WriteString(fmt.Sprintf("    providedVersion: %d\n", c.ProvidedVersion))
			output.WriteString(fmt.Sprintf("    currentVersion: %d\n", c.CurrentVersion))
			output.WriteString(fmt.Sprintf("    status: \"%s\"\n", c.Status))
		}
		return mcp.NewToolResultText(output.String()), nil
	}

	// 阶段3：执行批量修改
	type ModifyResult struct {
		PathName   string
		NewVersion int
		Error      string
	}

	results := make([]ModifyResult, 0, len(operations))
	for _, op := range operations {
		result := ModifyResult{PathName: op.PathName}

		// 构建更新数据
		updateData := make(map[string]interface{})
		for k, v := range op.Data {
			updateData[k] = v
		}
		updateData["version"] = op.Version

		// 调用 API 更新模块
		apiPath := fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), op.PathName)
		jsonData, _ := json.Marshal(updateData)
		req, _ := http.NewRequest("PUT", apiPath, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			result.Error = "请求失败：" + err.Error()
			results = append(results, result)
			continue
		}
		defer resp.Body.Close()

		var apiResult map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&apiResult); err != nil {
			result.Error = "解析失败：" + err.Error()
			results = append(results, result)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			// 获取新版本号
			var moduleData map[string]interface{}
			if data, ok := apiResult["data"].(map[string]interface{}); ok {
				moduleData, _ = data["module"].(map[string]interface{})
			}
			if moduleData == nil {
				moduleData, _ = apiResult["module"].(map[string]interface{})
			}
			if moduleData != nil {
				if v, ok := moduleData["version"].(float64); ok {
					result.NewVersion = int(v)
				}
			}
		} else {
			errorBytes, _ := json.Marshal(apiResult)
			result.Error = string(errorBytes)
		}
		results = append(results, result)
	}

	// 统计成功和失败
	successCount := 0
	failCount := 0
	for _, r := range results {
		if r.Error == "" {
			successCount++
		} else {
			failCount++
		}
	}

	// 返回结果
	var output strings.Builder
	if failCount == 0 {
		output.WriteString("success: true\n")
		output.WriteString(fmt.Sprintf("summary: 全部 %d 个模块修改成功\n", successCount))
	} else {
		output.WriteString("success: partial\n")
		output.WriteString(fmt.Sprintf("summary: %d 个成功，%d 个失败\n", successCount, failCount))
	}
	output.WriteString("results:\n")
	for _, r := range results {
		output.WriteString(fmt.Sprintf("  - pathName: \"%s\"\n", r.PathName))
		if r.Error != "" {
			output.WriteString(fmt.Sprintf("    error: \"%s\"\n", r.Error))
		} else {
			output.WriteString(fmt.Sprintf("    newVersion: %d\n", r.NewVersion))
		}
	}

	// 对每个成功修改的模块执行静态编译，并将结果追加到响应
	output.WriteString("\n--- 静态编译结果 ---\n")
	hasCompileErrors := false
	for _, r := range results {
		if r.Error != "" {
			continue // 跳过修改失败的模块
		}
		compileResult := &CompileStaticResult{
			Success:  true,
			Errors:   []CompileIssue{},
			Warnings: []CompileIssue{},
		}
		s.compileStaticModule(r.PathName, true, compileResult)
		compileResult.ErrorCount = len(compileResult.Errors)
		compileResult.WarningCount = len(compileResult.Warnings)
		compileResult.Success = compileResult.ErrorCount == 0

		// 同步写入 bugLog
		s.writeStaticBugLogToDB(r.PathName, compileResult.Errors, compileResult.Warnings)

		output.WriteString(fmt.Sprintf("module: \"%s\"\n", r.PathName))
		if compileResult.Success {
			output.WriteString("  compile: pass\n")
		} else {
			hasCompileErrors = true
			output.WriteString(fmt.Sprintf("  compile: fail (%d errors, %d warnings)\n", compileResult.ErrorCount, compileResult.WarningCount))
			for _, e := range compileResult.Errors {
				output.WriteString(fmt.Sprintf("  error [%s]: %s\n", e.RuleID, e.Message))
			}
			for _, w := range compileResult.Warnings {
				output.WriteString(fmt.Sprintf("  warning [%s]: %s\n", w.RuleID, w.Message))
			}
		}
	}
	if !hasCompileErrors {
		output.WriteString("all: 静态编译全部通过\n")
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

// handleModifyTaskImpl 批量修改任务信息
// 参数:
//   - operations: 批量操作数组，每项包含 {pathName: string, version: number, data: object}
// 原子性保证：预检所有版本号，任一冲突则全部拒绝
func (s *MCPServer) handleModifyTaskImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取 operations 参数
	operationsVal, ok := getParamAny(request, "operations")
	if !ok || operationsVal == nil {
		return mcp.NewToolResultText("缺少 operations 参数"), nil
	}

	// 解析 operations 数组
	var operations []ModifyOperation
	if operationsArr, ok := operationsVal.([]interface{}); ok {
		if len(operationsArr) == 0 {
			return mcp.NewToolResultText("operations 数组不能为空"), nil
		}
		for _, op := range operationsArr {
			if opMap, ok := op.(map[string]interface{}); ok {
				operation := ModifyOperation{}
				if pathName, ok := opMap["pathName"].(string); ok {
					operation.PathName = pathName
				}
				// 支持多种数字类型
				switch v := opMap["version"].(type) {
				case float64:
					operation.Version = int(v)
				case int:
					operation.Version = v
				case int64:
					operation.Version = int(v)
				}
				if data, ok := opMap["data"].(map[string]interface{}); ok {
					operation.Data = data
				}
				if operation.PathName != "" && operation.Version > 0 && len(operation.Data) > 0 {
					operations = append(operations, operation)
				}
			}
		}
	}

	if len(operations) == 0 {
		return mcp.NewToolResultText("未能解析有效的操作项"), nil
	}

	// 验证所有 pathName 是否属于当前项目
	for _, op := range operations {
		if errMsg := s.configManager.ValidatePathBelongsToProject(op.PathName); errMsg != "" {
			return mcp.NewToolResultText(errMsg), nil
		}
	}

	// 阶段1：预检所有版本号
	var checks []VersionCheckResult
	var conflicts []VersionCheckResult

	for _, op := range operations {
		check := VersionCheckResult{
			PathName:        op.PathName,
			ProvidedVersion: op.Version,
		}

		// 获取当前任务版本
		url := fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), op.PathName)
		resp, err := http.Get(url)
		if err != nil {
			check.Status = "error: " + err.Error()
			checks = append(checks, check)
			conflicts = append(conflicts, check)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 404 {
			check.Status = "error: 任务不存在"
			checks = append(checks, check)
			conflicts = append(conflicts, check)
			continue
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			check.Status = "error: 解析失败"
			checks = append(checks, check)
			conflicts = append(conflicts, check)
			continue
		}

		// 提取当前版本
		var taskData map[string]interface{}
		if data, ok := result["data"].(map[string]interface{}); ok {
			taskData, _ = data["task"].(map[string]interface{})
		}
		if taskData == nil {
			taskData, _ = result["task"].(map[string]interface{})
		}
		if taskData == nil {
			taskData = result
		}

		if v, ok := taskData["version"].(float64); ok {
			check.CurrentVersion = int(v)
			if check.CurrentVersion == check.ProvidedVersion {
				check.Status = "版本正常"
			} else {
				check.Status = "版本冲突"
				conflicts = append(conflicts, check)
			}
		} else {
			check.Status = "error: 无法获取版本"
			conflicts = append(conflicts, check)
		}
		checks = append(checks, check)
	}

	// 阶段2：决策 - 如果有任何冲突，拒绝整个批量操作
	if len(conflicts) > 0 {
		var output strings.Builder
		output.WriteString("success: false\n")
		output.WriteString("error: VERSION_CONFLICT\n")
		output.WriteString(fmt.Sprintf("message: 批量修改已取消：发现 %d 个版本冲突\n", len(conflicts)))
		output.WriteString("conflicts:\n")
		for _, c := range conflicts {
			output.WriteString(fmt.Sprintf("  - pathName: \"%s\"\n", c.PathName))
			output.WriteString(fmt.Sprintf("    providedVersion: %d\n", c.ProvidedVersion))
			output.WriteString(fmt.Sprintf("    currentVersion: %d\n", c.CurrentVersion))
			output.WriteString(fmt.Sprintf("    message: \"%s\"\n", c.Status))
		}
		output.WriteString("checks:\n")
		for _, c := range checks {
			output.WriteString(fmt.Sprintf("  - pathName: \"%s\"\n", c.PathName))
			output.WriteString(fmt.Sprintf("    providedVersion: %d\n", c.ProvidedVersion))
			output.WriteString(fmt.Sprintf("    currentVersion: %d\n", c.CurrentVersion))
			output.WriteString(fmt.Sprintf("    status: \"%s\"\n", c.Status))
		}
		return mcp.NewToolResultText(output.String()), nil
	}

	// 阶段3：执行批量修改
	type ModifyResult struct {
		PathName   string
		NewVersion int
		Error      string
	}

	results := make([]ModifyResult, 0, len(operations))
	for _, op := range operations {
		result := ModifyResult{PathName: op.PathName}

		// 构建更新数据，处理 JSON 字段
		updateData := make(map[string]interface{})
		for k, v := range op.Data {
			// 对需要存储为 JSON 字符串的字段进行序列化
			switch k {
			case "tests", "codePaths", "bugLog", "testResult":
				if arr, ok := v.([]interface{}); ok {
					jsonBytes, _ := json.Marshal(arr)
					updateData[k] = string(jsonBytes)
				} else if m, ok := v.(map[string]interface{}); ok {
					jsonBytes, _ := json.Marshal(m)
					updateData[k] = string(jsonBytes)
				} else if str, ok := v.(string); ok {
					updateData[k] = str
				} else {
					updateData[k] = v
				}
			case "upstreamContractDetail", "downstreamContractDetail", "humanAssistance", "issueDetails":
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
		updateData["version"] = op.Version

		// 调用 API 更新任务
		apiPath := fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), op.PathName)
		jsonData, _ := json.Marshal(updateData)
		req, _ := http.NewRequest("PUT", apiPath, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			result.Error = "请求失败：" + err.Error()
			results = append(results, result)
			continue
		}
		defer resp.Body.Close()

		var apiResult map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&apiResult); err != nil {
			result.Error = "解析失败：" + err.Error()
			results = append(results, result)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			// 获取新版本号
			var taskData map[string]interface{}
			if data, ok := apiResult["data"].(map[string]interface{}); ok {
				taskData, _ = data["task"].(map[string]interface{})
			}
			if taskData == nil {
				taskData, _ = apiResult["task"].(map[string]interface{})
			}
			if taskData != nil {
				if v, ok := taskData["version"].(float64); ok {
					result.NewVersion = int(v)
				}
			}
		} else {
			errorBytes, _ := json.Marshal(apiResult)
			result.Error = string(errorBytes)
		}
		results = append(results, result)
	}

	// 统计成功和失败
	successCount := 0
	failCount := 0
	for _, r := range results {
		if r.Error == "" {
			successCount++
		} else {
			failCount++
		}
	}

	// 返回结果
	var output strings.Builder
	if failCount == 0 {
		output.WriteString("success: true\n")
		output.WriteString(fmt.Sprintf("summary: 全部 %d 个任务修改成功\n", successCount))
	} else {
		output.WriteString("success: partial\n")
		output.WriteString(fmt.Sprintf("summary: %d 个成功，%d 个失败\n", successCount, failCount))
	}
	output.WriteString("results:\n")
	for _, r := range results {
		output.WriteString(fmt.Sprintf("  - pathName: \"%s\"\n", r.PathName))
		if r.Error != "" {
			output.WriteString(fmt.Sprintf("    error: \"%s\"\n", r.Error))
		} else {
			output.WriteString(fmt.Sprintf("    newVersion: %d\n", r.NewVersion))
		}
	}

	// 对每个成功修改的任务执行静态编译，并将结果追加到响应
	output.WriteString("\n--- 静态编译结果 ---\n")
	hasCompileErrors := false
	for _, r := range results {
		if r.Error != "" {
			continue // 跳过修改失败的任务
		}
		compileResult := &CompileStaticResult{
			Success:  true,
			Errors:   []CompileIssue{},
			Warnings: []CompileIssue{},
		}
		s.compileStaticTask(r.PathName, true, compileResult)
		compileResult.ErrorCount = len(compileResult.Errors)
		compileResult.WarningCount = len(compileResult.Warnings)
		compileResult.Success = compileResult.ErrorCount == 0

		// 同步写入 bugLog
		s.writeStaticBugLogToDB(r.PathName, compileResult.Errors, compileResult.Warnings)

		output.WriteString(fmt.Sprintf("task: \"%s\"\n", r.PathName))
		if compileResult.Success {
			output.WriteString("  compile: pass\n")
		} else {
			hasCompileErrors = true
			output.WriteString(fmt.Sprintf("  compile: fail (%d errors, %d warnings)\n", compileResult.ErrorCount, compileResult.WarningCount))
			for _, e := range compileResult.Errors {
				output.WriteString(fmt.Sprintf("  error [%s]: %s\n", e.RuleID, e.Message))
			}
			for _, w := range compileResult.Warnings {
				output.WriteString(fmt.Sprintf("  warning [%s]: %s\n", w.RuleID, w.Message))
			}
		}
	}
	if !hasCompileErrors {
		output.WriteString("all: 静态编译全部通过\n")
	}

	return mcp.NewToolResultText(output.String()), nil
}

// ==================== 批量创建操作实现 ====================

// CreateOperation 创建操作结构
type CreateOperation struct {
	ParentPath string                 // 父节点路径
	Name       string                 // 名称
	PathName   string                 // 可选的 pathName
	Data       map[string]interface{} // 创建数据
}

// CreateResult 创建操作结果
type CreateResult struct {
	PathName string `json:"pathName"`
	Name     string `json:"name"`
	Status   string `json:"status"` // "created" 或 "failed"
	Error    string `json:"error,omitempty"`
}

// handleCreateModuleImpl 批量创建模块
func (s *MCPServer) handleCreateModuleImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取 operations 参数
	operationsVal, ok := getParamAny(request, "operations")
	if !ok || operationsVal == nil {
		return mcp.NewToolResultText("缺少 operations 参数"), nil
	}

	// 解析 operations 数组
	var operations []CreateOperation
	if operationsArr, ok := operationsVal.([]interface{}); ok {
		if len(operationsArr) == 0 {
			return mcp.NewToolResultText("operations 数组不能为空"), nil
		}
		for _, op := range operationsArr {
			if opMap, ok := op.(map[string]interface{}); ok {
				operation := CreateOperation{}
				if parentPath, ok := opMap["parentPath"].(string); ok {
					operation.ParentPath = parentPath
				}
				if name, ok := opMap["name"].(string); ok {
					operation.Name = name
				}
				if pathName, ok := opMap["pathName"].(string); ok {
					operation.PathName = pathName
				}
				if data, ok := opMap["data"].(map[string]interface{}); ok {
					operation.Data = data
				}
				if operation.ParentPath != "" && operation.Name != "" {
					operations = append(operations, operation)
				}
			}
		}
	}

	if len(operations) == 0 {
		return mcp.NewToolResultText("未能解析有效的操作项，每项需要 parentPath 和 name"), nil
	}

	// 验证所有 parentPath 是否属于当前项目
	for _, op := range operations {
		if errMsg := s.configManager.ValidatePathBelongsToProject(op.ParentPath); errMsg != "" {
			return mcp.NewToolResultText(errMsg), nil
		}
	}

	// 逐个执行创建操作
	results := make([]CreateResult, len(operations))
	for i, op := range operations {
		results[i] = s.createSingleModule(op)
	}

	// 汇总结果
	return formatCreateResults(results, "模块"), nil
}

// createSingleModule 创建单个模块
func (s *MCPServer) createSingleModule(op CreateOperation) CreateResult {
	result := CreateResult{
		Name: op.Name,
	}

	// 生成或使用 pathName
	pathName := op.PathName
	expectedPathName := generatePathName(op.ParentPath, op.Name)
	if pathName == "" {
		pathName = expectedPathName
	} else if pathName != expectedPathName {
		// 校验 pathName 是否与 name 对齐
		result.PathName = pathName
		result.Status = "failed"
		result.Error = fmt.Sprintf("pathName 与 name 不对齐。pathName 应为 '%s'（格式: parentPath + \"/\" + name），但传入的是 '%s'。请修改 pathName 或省略该字段让系统自动生成", expectedPathName, pathName)
		return result
	}
	result.PathName = pathName

	// 从配置文件获取项目 pathName
	projectPathName := s.configManager.GetProjectPathName()
	if projectPathName == "" {
		result.Status = "failed"
		result.Error = "未配置项目路径名称，请先使用 init_project 设置项目"
		return result
	}

	// 校验 parentPath 是否与当前项目匹配
	// 提取 parentPath 的第一个字段（项目名称）
	parentProject := op.ParentPath
	if idx := strings.Index(parentProject, "/"); idx != -1 {
		parentProject = op.ParentPath[:idx]
	}
	if parentProject != projectPathName {
		result.Status = "failed"
		result.Error = fmt.Sprintf("parentPath '%s' 与当前项目 '%s' 不匹配。请先使用 init_project 切换到正确的项目，或修改 parentPath", op.ParentPath, projectPathName)
		return result
	}

	// 获取项目 ID
	resp, err := http.Get(fmt.Sprintf("%s/projects/by-path/%s", s.getApiURL(), projectPathName))
	if err != nil {
		result.Status = "failed"
		result.Error = "获取项目信息失败：" + err.Error()
		return result
	}
	defer resp.Body.Close()

	var projectResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&projectResult); err != nil {
		result.Status = "failed"
		result.Error = "解析项目信息失败：" + err.Error()
		return result
	}

	// 从嵌套的 data 结构中获取 project
	var project map[string]interface{}
	if data, ok := projectResult["data"].(map[string]interface{}); ok {
		project, _ = data["project"].(map[string]interface{})
	}
	if project == nil {
		project, _ = projectResult["project"].(map[string]interface{})
	}
	if project == nil {
		result.Status = "failed"
		result.Error = "项目不存在或格式错误"
		return result
	}
	projectID, ok := project["id"].(string)
	if !ok {
		result.Status = "failed"
		result.Error = "项目ID不存在"
		return result
	}

	// 构建创建数据
	createData := map[string]interface{}{
		"name":      op.Name,
		"projectId": projectID,
		"pathName":  pathName,
	}

	// 从 data 中提取模块字段
	if op.Data != nil {
		if description, ok := op.Data["description"].(string); ok {
			createData["description"] = description
		}
		if prompt, ok := op.Data["prompt"].(string); ok {
			createData["prompt"] = prompt
		}
		if upstreamContract, ok := op.Data["upstreamContractSummary"].(string); ok {
			createData["upstreamContractSummary"] = upstreamContract
		}
		if downstreamContract, ok := op.Data["downstreamContractSummary"].(string); ok {
			createData["downstreamContractSummary"] = downstreamContract
		}
		if status, ok := op.Data["status"].(string); ok {
			createData["status"] = status
		}
	}

	// 如果 parentPath 不是项目路径，则认为是子模块
	if op.ParentPath != projectPathName {
		resp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), op.ParentPath))
		if err == nil {
			defer resp.Body.Close()
			var moduleResult map[string]interface{}
			if json.NewDecoder(resp.Body).Decode(&moduleResult) == nil {
				if module, ok := moduleResult["module"].(map[string]interface{}); ok {
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
		result.Status = "failed"
		result.Error = "请求失败：" + err.Error()
		return result
	}
	defer resp2.Body.Close()

	var apiResult map[string]interface{}
	if err := json.NewDecoder(resp2.Body).Decode(&apiResult); err != nil {
		result.Status = "failed"
		result.Error = "解析失败：" + err.Error()
		return result
	}

	if resp2.StatusCode == http.StatusOK || resp2.StatusCode == http.StatusCreated {
		result.Status = "created"
		// 更新实际的 pathName
		if module, ok := apiResult["module"].(map[string]interface{}); ok {
			if pn, ok := module["pathName"].(string); ok {
				result.PathName = pn
			}
		}
	} else {
		result.Status = "failed"
		errorBytes, _ := json.Marshal(apiResult)
		result.Error = string(errorBytes)
	}

	return result
}

// handleCreateTaskImpl 批量创建任务
func (s *MCPServer) handleCreateTaskImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取 operations 参数
	operationsVal, ok := getParamAny(request, "operations")
	if !ok || operationsVal == nil {
		return mcp.NewToolResultText("缺少 operations 参数"), nil
	}

	// 解析 operations 数组
	var operations []CreateOperation
	if operationsArr, ok := operationsVal.([]interface{}); ok {
		if len(operationsArr) == 0 {
			return mcp.NewToolResultText("operations 数组不能为空"), nil
		}
		for _, op := range operationsArr {
			if opMap, ok := op.(map[string]interface{}); ok {
				operation := CreateOperation{}
				if parentPath, ok := opMap["parentPath"].(string); ok {
					operation.ParentPath = parentPath
				}
				if name, ok := opMap["name"].(string); ok {
					operation.Name = name
				}
				if pathName, ok := opMap["pathName"].(string); ok {
					operation.PathName = pathName
				}
				if data, ok := opMap["data"].(map[string]interface{}); ok {
					operation.Data = data
				}
				if operation.ParentPath != "" && operation.Name != "" {
					operations = append(operations, operation)
				}
			}
		}
	}

	if len(operations) == 0 {
		return mcp.NewToolResultText("未能解析有效的操作项，每项需要 parentPath 和 name"), nil
	}

	// 验证所有 parentPath 是否属于当前项目
	for _, op := range operations {
		if errMsg := s.configManager.ValidatePathBelongsToProject(op.ParentPath); errMsg != "" {
			return mcp.NewToolResultText(errMsg), nil
		}
	}

	// 逐个执行创建操作
	results := make([]CreateResult, len(operations))
	for i, op := range operations {
		results[i] = s.createSingleTask(op)
	}

	// 汇总结果
	return formatCreateResults(results, "任务"), nil
}

// createSingleTask 创建单个任务
func (s *MCPServer) createSingleTask(op CreateOperation) CreateResult {
	result := CreateResult{
		Name: op.Name,
	}

	// 生成或使用 pathName
	pathName := op.PathName
	expectedPathName := generatePathName(op.ParentPath, op.Name)
	if pathName == "" {
		pathName = expectedPathName
	} else if pathName != expectedPathName {
		// 校验 pathName 是否与 name 对齐
		result.PathName = pathName
		result.Status = "failed"
		result.Error = fmt.Sprintf("pathName 与 name 不对齐。pathName 应为 '%s'（格式: parentPath + \"/\" + name），但传入的是 '%s'。请修改 pathName 或省略该字段让系统自动生成", expectedPathName, pathName)
		return result
	}
	result.PathName = pathName

	// 从配置文件获取项目 pathName
	projectPathName := s.configManager.GetProjectPathName()
	if projectPathName == "" {
		result.Status = "failed"
		result.Error = "未配置项目路径名称，请先使用 init_project 设置项目"
		return result
	}

	// 校验 parentPath 是否与当前项目匹配
	// 提取 parentPath 的第一个字段（项目名称）
	parentProject := op.ParentPath
	if idx := strings.Index(parentProject, "/"); idx != -1 {
		parentProject = op.ParentPath[:idx]
	}
	if parentProject != projectPathName {
		result.Status = "failed"
		result.Error = fmt.Sprintf("parentPath '%s' 与当前项目 '%s' 不匹配。请先使用 init_project 切换到正确的项目，或修改 parentPath", op.ParentPath, projectPathName)
		return result
	}

	// parentPath 是模块路径，需要获取模块 ID
	moduleResp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), op.ParentPath))
	if err != nil {
		result.Status = "failed"
		result.Error = "获取模块信息失败：" + err.Error()
		return result
	}
	defer moduleResp.Body.Close()

	var moduleResult map[string]interface{}
	if err := json.NewDecoder(moduleResp.Body).Decode(&moduleResult); err != nil {
		result.Status = "failed"
		result.Error = "解析模块信息失败：" + err.Error()
		return result
	}

	// 从嵌套的 data 结构中获取 module
	var moduleData map[string]interface{}
	if data, ok := moduleResult["data"].(map[string]interface{}); ok {
		moduleData, _ = data["module"].(map[string]interface{})
	}
	if moduleData == nil {
		moduleData, _ = moduleResult["module"].(map[string]interface{})
	}
	if moduleData == nil {
		result.Status = "failed"
		result.Error = "模块数据格式错误"
		return result
	}
	moduleID, ok := moduleData["id"].(string)
	if !ok {
		result.Status = "failed"
		result.Error = "无法获取模块ID"
		return result
	}

	// 构建创建数据
	createData := map[string]interface{}{
		"moduleId": moduleID,
		"name":     op.Name,
		"status":   "ready",
		"pathName": pathName,
	}

	// 从 data 中提取任务字段
	if op.Data != nil {
		if description, ok := op.Data["description"].(string); ok {
			createData["description"] = description
		}
		if prompt, ok := op.Data["prompt"].(string); ok {
			createData["prompt"] = prompt
		}
		// 处理 upstreamContractDetail
		if upstreamContract, ok := op.Data["upstreamContractDetail"]; ok {
			switch v := upstreamContract.(type) {
			case map[string]interface{}:
				jsonBytes, _ := json.Marshal(v)
				createData["upstreamContractDetail"] = string(jsonBytes)
			case string:
				createData["upstreamContractDetail"] = v
			}
		}
		// 处理 downstreamContractDetail
		if downstreamContract, ok := op.Data["downstreamContractDetail"]; ok {
			switch v := downstreamContract.(type) {
			case map[string]interface{}:
				jsonBytes, _ := json.Marshal(v)
				createData["downstreamContractDetail"] = string(jsonBytes)
			case string:
				createData["downstreamContractDetail"] = v
			}
		}
		// 处理 tests
		if tests, ok := op.Data["tests"]; ok {
			switch v := tests.(type) {
			case []interface{}:
				jsonBytes, _ := json.Marshal(v)
				createData["tests"] = string(jsonBytes)
			case string:
				createData["tests"] = v
			}
		}
		// 处理 codePaths
		if codePaths, ok := op.Data["codePaths"]; ok {
			switch v := codePaths.(type) {
			case []interface{}:
				jsonBytes, _ := json.Marshal(v)
				createData["codePaths"] = string(jsonBytes)
			case string:
				createData["codePaths"] = v
			}
		}
		if status, ok := op.Data["status"].(string); ok {
			createData["status"] = status
		}
	}

	jsonData, _ := json.Marshal(createData)
	resp, err := http.Post(fmt.Sprintf("%s/tasks", s.getApiURL()), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		result.Status = "failed"
		result.Error = "请求失败：" + err.Error()
		return result
	}
	defer resp.Body.Close()

	var apiResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&apiResult); err != nil {
		result.Status = "failed"
		result.Error = "解析失败：" + err.Error()
		return result
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		result.Status = "created"
		// 更新实际的 pathName
		if task, ok := apiResult["task"].(map[string]interface{}); ok {
			if pn, ok := task["pathName"].(string); ok {
				result.PathName = pn
			}
		}
	} else {
		result.Status = "failed"
		errorBytes, _ := json.Marshal(apiResult)
		result.Error = string(errorBytes)
	}

	return result
}

// handleCreateProjectImpl 批量创建项目
func (s *MCPServer) handleCreateProjectImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取 operations 参数
	operationsVal, ok := getParamAny(request, "operations")
	if !ok || operationsVal == nil {
		return mcp.NewToolResultText("缺少 operations 参数"), nil
	}

	// 解析 operations 数组
	type ProjectCreateOperation struct {
		Name     string
		PathName string
		Data     map[string]interface{}
	}

	var operations []ProjectCreateOperation
	if operationsArr, ok := operationsVal.([]interface{}); ok {
		if len(operationsArr) == 0 {
			return mcp.NewToolResultText("operations 数组不能为空"), nil
		}
		for _, op := range operationsArr {
			if opMap, ok := op.(map[string]interface{}); ok {
				operation := ProjectCreateOperation{}
				if name, ok := opMap["name"].(string); ok {
					operation.Name = name
				}
				if pathName, ok := opMap["pathName"].(string); ok {
					operation.PathName = pathName
				}
				if data, ok := opMap["data"].(map[string]interface{}); ok {
					operation.Data = data
				}
				if operation.Name != "" && operation.PathName != "" {
					operations = append(operations, operation)
				}
			}
		}
	}

	if len(operations) == 0 {
		return mcp.NewToolResultText("未能解析有效的操作项，每项需要 name 和 pathName"), nil
	}

	// 逐个执行创建操作
	results := make([]CreateResult, len(operations))
	for i, op := range operations {
		results[i] = s.createSingleProject(op.Name, op.PathName, op.Data)
	}

	// 汇总结果
	return formatCreateResults(results, "项目"), nil
}

// createSingleProject 创建单个项目
func (s *MCPServer) createSingleProject(name, pathName string, data map[string]interface{}) CreateResult {
	result := CreateResult{
		Name:     name,
		PathName: pathName,
	}

	// 验证 pathName 不包含斜杠
	if strings.Contains(pathName, "/") {
		result.Status = "failed"
		result.Error = "pathName 不能包含斜杠"
		return result
	}

	// 构建创建数据
	createData := map[string]interface{}{
		"name":     name,
		"pathName": pathName,
	}

	// 从 data 中提取项目字段
	if data != nil {
		if description, ok := data["description"].(string); ok {
			createData["description"] = description
		}
		if repository, ok := data["repository"].(string); ok {
			createData["repository"] = repository
		}
	}

	jsonData, _ := json.Marshal(createData)
	resp, err := http.Post(fmt.Sprintf("%s/projects", s.getApiURL()), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		result.Status = "failed"
		result.Error = "请求失败：" + err.Error()
		return result
	}
	defer resp.Body.Close()

	var apiResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&apiResult); err != nil {
		result.Status = "failed"
		result.Error = "解析失败：" + err.Error()
		return result
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		result.Status = "created"
		// 更新实际的 pathName
		if project, ok := apiResult["project"].(map[string]interface{}); ok {
			if pn, ok := project["pathName"].(string); ok {
				result.PathName = pn
				// 创建成功后自动切换到新项目
				s.configManager.SetProjectPathName(pn)
			}
		}
	} else {
		result.Status = "failed"
		errorBytes, _ := json.Marshal(apiResult)
		result.Error = string(errorBytes)
	}

	return result
}

// formatCreateResults 格式化创建结果
func formatCreateResults(results []CreateResult, resourceType string) *mcp.CallToolResult {
	successCount := 0
	failedCount := 0
	for _, r := range results {
		if r.Status == "created" {
			successCount++
		} else {
			failedCount++
		}
	}

	var output strings.Builder
	total := len(results)

	if failedCount == 0 {
		output.WriteString("success: true\n")
		output.WriteString(fmt.Sprintf("summary: \"全部 %d 个%s创建成功\"\n", total, resourceType))
	} else if successCount == 0 {
		output.WriteString("success: false\n")
		output.WriteString(fmt.Sprintf("summary: \"全部 %d 个操作失败\"\n", total))
	} else {
		output.WriteString("success: partial\n")
		output.WriteString(fmt.Sprintf("summary: \"%d 个操作中 %d 个成功，%d 个失败\"\n", total, successCount, failedCount))
	}

	output.WriteString("results:\n")
	for _, r := range results {
		output.WriteString(fmt.Sprintf("  - pathName: \"%s\"\n", r.PathName))
		output.WriteString(fmt.Sprintf("    name: \"%s\"\n", r.Name))
		output.WriteString(fmt.Sprintf("    status: \"%s\"\n", r.Status))
		if r.Error != "" {
			output.WriteString(fmt.Sprintf("    error: \"%s\"\n", r.Error))
		}
	}

	return mcp.NewToolResultText(output.String())
}
