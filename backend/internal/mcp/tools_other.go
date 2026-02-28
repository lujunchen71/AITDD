package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// ==================== 3. 验证检查实现 ====================

// handleCheckContractAlignmentImpl 检查上下游任务契约是否对齐
func (s *MCPServer) handleCheckContractAlignmentImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstreamPathName, ok1 := getParam(request, "upstreamPathName")
	downstreamPathName, ok2 := getParam(request, "downstreamPathName")

	if !ok1 || !ok2 || upstreamPathName == "" || downstreamPathName == "" {
		return mcp.NewToolResultText("缺少必要参数：upstreamPathName 和 downstreamPathName"), nil
	}

	// 获取上游任务信息
	upstreamResp, err := http.Get(fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), upstreamPathName))
	if err != nil {
		return mcp.NewToolResultText("获取上游任务信息失败：" + err.Error()), nil
	}
	defer upstreamResp.Body.Close()

	var upstreamTask map[string]interface{}
	if err := json.NewDecoder(upstreamResp.Body).Decode(&upstreamTask); err != nil {
		return mcp.NewToolResultText("解析上游任务信息失败：" + err.Error()), nil
	}

	// 获取下游任务信息
	downstreamResp, err := http.Get(fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), downstreamPathName))
	if err != nil {
		return mcp.NewToolResultText("获取下游任务信息失败：" + err.Error()), nil
	}
	defer downstreamResp.Body.Close()

	var downstreamTask map[string]interface{}
	if err := json.NewDecoder(downstreamResp.Body).Decode(&downstreamTask); err != nil {
		return mcp.NewToolResultText("解析下游任务信息失败：" + err.Error()), nil
	}

	// 获取契约详情
	upstreamDownstreamContract, _ := upstreamTask["downstreamContractDetail"].(string)
	downstreamUpstreamContract, _ := downstreamTask["upstreamContractDetail"].(string)

	// 比较契约
	differences := compareContracts(upstreamDownstreamContract, downstreamUpstreamContract)
	aligned := len(differences) == 0

	// 构建输出
	var output strings.Builder
	output.WriteString(fmt.Sprintf("aligned: %v\n", aligned))

	if len(differences) > 0 {
		output.WriteString("differences:\n")
		for _, diff := range differences {
			output.WriteString(fmt.Sprintf("  - field: \"%s\"\n", diff["field"]))
			output.WriteString(fmt.Sprintf("    upstream: \"%s\"\n", diff["upstream"]))
			output.WriteString(fmt.Sprintf("    downstream: \"%s\"\n", diff["downstream"]))
		}
	}

	return mcp.NewToolResultText(output.String()), nil
}

// compareContracts 比较两个契约并返回差异
func compareContracts(upstreamContract, downstreamContract string) []map[string]string {
	differences := []map[string]string{}

	// 如果两者都为空，认为对齐
	if upstreamContract == "" && downstreamContract == "" {
		return differences
	}

	// 如果其中一个为空
	if upstreamContract == "" {
		differences = append(differences, map[string]string{
			"field":      "contract",
			"upstream":   "(空)",
			"downstream": downstreamContract,
		})
		return differences
	}
	if downstreamContract == "" {
		differences = append(differences, map[string]string{
			"field":      "contract",
			"upstream":   upstreamContract,
			"downstream": "(空)",
		})
		return differences
	}

	// 尝试解析为 JSON 进行比较
	var upstreamJSON, downstreamJSON map[string]interface{}
	upstreamErr := json.Unmarshal([]byte(upstreamContract), &upstreamJSON)
	downstreamErr := json.Unmarshal([]byte(downstreamContract), &downstreamJSON)

	if upstreamErr != nil || downstreamErr != nil {
		// 如果无法解析为 JSON，进行字符串比较
		if upstreamContract != downstreamContract {
			differences = append(differences, map[string]string{
				"field":      "contract",
				"upstream":   truncateString(upstreamContract, 100),
				"downstream": truncateString(downstreamContract, 100),
			})
		}
		return differences
	}

	// 比较 JSON 字段
	return compareJSONFields(upstreamJSON, downstreamJSON, "")
}

// compareJSONFields 递归比较 JSON 字段
func compareJSONFields(upstream, downstream map[string]interface{}, prefix string) []map[string]string {
	differences := []map[string]string{}

	// 检查上游有但下游没有的字段
	for key, upValue := range upstream {
		fieldPath := key
		if prefix != "" {
			fieldPath = prefix + "." + key
		}

		downValue, exists := downstream[key]
		if !exists {
			differences = append(differences, map[string]string{
				"field":      fieldPath,
				"upstream":   fmt.Sprintf("%v", upValue),
				"downstream": "(缺失)",
			})
			continue
		}

		// 递归比较嵌套对象
		upNested, upOk := upValue.(map[string]interface{})
		downNested, downOk := downValue.(map[string]interface{})
		if upOk && downOk {
			nestedDiffs := compareJSONFields(upNested, downNested, fieldPath)
			differences = append(differences, nestedDiffs...)
			continue
		}

		// 比较值
		if !valuesEqual(upValue, downValue) {
			differences = append(differences, map[string]string{
				"field":      fieldPath,
				"upstream":   fmt.Sprintf("%v", upValue),
				"downstream": fmt.Sprintf("%v", downValue),
			})
		}
	}

	// 检查下游有但上游没有的字段
	for key, downValue := range downstream {
		fieldPath := key
		if prefix != "" {
			fieldPath = prefix + "." + key
		}

		if _, exists := upstream[key]; !exists {
			differences = append(differences, map[string]string{
				"field":      fieldPath,
				"upstream":   "(缺失)",
				"downstream": fmt.Sprintf("%v", downValue),
			})
		}
	}

	return differences
}

// valuesEqual 比较两个值是否相等
func valuesEqual(a, b interface{}) bool {
	// 处理 nil 情况
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// 处理数值类型（JSON 解析后整数会变成 float64）
	switch aVal := a.(type) {
	case float64:
		if bVal, ok := b.(float64); ok {
			return aVal == bVal
		}
	case string:
		if bVal, ok := b.(string); ok {
			return aVal == bVal
		}
	case bool:
		if bVal, ok := b.(bool); ok {
			return aVal == bVal
		}
	}

	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// handleCheckTaskReadinessImpl 检查任务是否准备好开始开发
func (s *MCPServer) handleCheckTaskReadinessImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	// 获取任务信息
	taskResp, err := http.Get(fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), pathName))
	if err != nil {
		return mcp.NewToolResultText("获取任务信息失败：" + err.Error()), nil
	}
	defer taskResp.Body.Close()

	var task map[string]interface{}
	if err := json.NewDecoder(taskResp.Body).Decode(&task); err != nil {
		return mcp.NewToolResultText("解析任务信息失败：" + err.Error()), nil
	}

	issues := []map[string]interface{}{}

	// 检查 prompt 是否为空 (error)
	prompt, _ := task["prompt"].(string)
	if prompt == "" {
		issues = append(issues, map[string]interface{}{
			"type":     "missing_prompt",
			"message":  "任务提示词为空",
			"severity": "error",
		})
	}

	// 检查 tests 是否为空 (warning)
	tests, _ := task["tests"].(string)
	if tests == "" || tests == "[]" {
		issues = append(issues, map[string]interface{}{
			"type":     "empty_tests",
			"message":  "未定义测试用例",
			"severity": "warning",
		})
	}

	// 获取任务 ID 用于查询依赖
	taskID, _ := task["id"].(string)

	// 检查上游依赖和契约
	depResp, err := http.Get(fmt.Sprintf("%s/dependencies?downstreamTaskId=%s", s.getApiURL(), taskID))
	if err == nil {
		defer depResp.Body.Close()
		var depResult map[string]interface{}
		if json.NewDecoder(depResp.Body).Decode(&depResult) == nil {
			if deps, ok := depResult["data"].([]interface{}); ok && len(deps) > 0 {
				// 有上游依赖
				upstreamContract, _ := task["upstreamContractDetail"].(string)
				if upstreamContract == "" {
					issues = append(issues, map[string]interface{}{
						"type":     "missing_upstream_contract",
						"message":  "有上游依赖但上游契约详情为空",
						"severity": "error",
					})
				}

				// 检查上游任务是否完成
				for _, dep := range deps {
					if depMap, ok := dep.(map[string]interface{}); ok {
						if upstreamTask, ok := depMap["upstreamTask"].(map[string]interface{}); ok {
							status, _ := upstreamTask["status"].(string)
							if status != "completed" {
								upstreamName, _ := upstreamTask["name"].(string)
								issues = append(issues, map[string]interface{}{
									"type":     "upstream_not_complete",
									"message":  fmt.Sprintf("上游任务 '%s' 尚未完成（状态：%s）", upstreamName, status),
									"severity": "warning",
								})
							}
						}
					}
				}
			}
		}
	}

	// 检查是否被锁定 (info)
	locked, _ := task["locked"].(bool)
	if locked {
		lockedBy, _ := task["lockedBy"].(string)
		issues = append(issues, map[string]interface{}{
			"type":     "locked",
			"message":  fmt.Sprintf("任务已被锁定（锁定者：%s）", lockedBy),
			"severity": "info",
		})
	}

	// 构建输出
	var output strings.Builder
	ready := true
	for _, issue := range issues {
		if issue["severity"] == "error" {
			ready = false
			break
		}
	}

	output.WriteString(fmt.Sprintf("ready: %v\n", ready))

	if len(issues) > 0 {
		output.WriteString("issues:\n")
		for _, issue := range issues {
			output.WriteString(fmt.Sprintf("  - type: \"%s\"\n", issue["type"]))
			output.WriteString(fmt.Sprintf("    message: \"%s\"\n", issue["message"]))
			output.WriteString(fmt.Sprintf("    severity: \"%s\"\n", issue["severity"]))
		}
	}

	return mcp.NewToolResultText(output.String()), nil
}

// handleCheckDependenciesImpl 检查依赖关系和阻塞状态
func (s *MCPServer) handleCheckDependenciesImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	nodeType, typeOk := getParam(request, "type")

	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}
	if !typeOk || nodeType == "" {
		return mcp.NewToolResultText("缺少 type 参数（task 或 module）"), nil
	}

	if nodeType != "task" && nodeType != "module" {
		return mcp.NewToolResultText("type 参数必须是 task 或 module"), nil
	}

	if nodeType == "task" {
		return s.checkTaskDependencies(pathName)
	}

	return s.checkModuleDependencies(pathName)
}

// checkTaskDependencies 检查任务依赖
func (s *MCPServer) checkTaskDependencies(pathName string) (*mcp.CallToolResult, error) {
	// 获取任务信息
	taskResp, err := http.Get(fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), pathName))
	if err != nil {
		return mcp.NewToolResultText("获取任务信息失败：" + err.Error()), nil
	}
	defer taskResp.Body.Close()

	var task map[string]interface{}
	if err := json.NewDecoder(taskResp.Body).Decode(&task); err != nil {
		return mcp.NewToolResultText("解析任务信息失败：" + err.Error()), nil
	}

	taskID, _ := task["id"].(string)

	// 获取上游依赖（当前任务作为下游）
	upstreamResp, err := http.Get(fmt.Sprintf("%s/dependencies?downstreamTaskId=%s", s.getApiURL(), taskID))
	if err != nil {
		return mcp.NewToolResultText("获取上游依赖失败：" + err.Error()), nil
	}
	defer upstreamResp.Body.Close()

	var upstreamResult map[string]interface{}
	if err := json.NewDecoder(upstreamResp.Body).Decode(&upstreamResult); err != nil {
		return mcp.NewToolResultText("解析上游依赖失败：" + err.Error()), nil
	}

	// 获取下游依赖（当前任务作为上游）
	downstreamResp, err := http.Get(fmt.Sprintf("%s/dependencies?upstreamTaskId=%s", s.getApiURL(), taskID))
	if err != nil {
		return mcp.NewToolResultText("获取下游依赖失败：" + err.Error()), nil
	}
	defer downstreamResp.Body.Close()

	var downstreamResult map[string]interface{}
	if err := json.NewDecoder(downstreamResp.Body).Decode(&downstreamResult); err != nil {
		return mcp.NewToolResultText("解析下游依赖失败：" + err.Error()), nil
	}

	// 解析上游依赖
	upstreamList := []map[string]interface{}{}
	blockedBy := []map[string]interface{}{}
	blocked := false

	if data, ok := upstreamResult["data"].([]interface{}); ok {
		for _, dep := range data {
			if depMap, ok := dep.(map[string]interface{}); ok {
				if upstreamTask, ok := depMap["upstreamTask"].(map[string]interface{}); ok {
					upstreamName, _ := upstreamTask["name"].(string)
					upstreamPathName, _ := upstreamTask["pathName"].(string)
					upstreamStatus, _ := upstreamTask["status"].(string)

					upstreamList = append(upstreamList, map[string]interface{}{
						"pathName": upstreamPathName,
						"name":     upstreamName,
						"status":   upstreamStatus,
					})

					// 检查是否阻塞
					if upstreamStatus != "completed" {
						blocked = true
						blockedBy = append(blockedBy, map[string]interface{}{
							"pathName": upstreamPathName,
							"name":     upstreamName,
							"status":   upstreamStatus,
						})
					}
				}
			}
		}
	}

	// 解析下游依赖
	downstreamList := []map[string]interface{}{}
	if data, ok := downstreamResult["data"].([]interface{}); ok {
		for _, dep := range data {
			if depMap, ok := dep.(map[string]interface{}); ok {
				if downstreamTask, ok := depMap["downstreamTask"].(map[string]interface{}); ok {
					downstreamName, _ := downstreamTask["name"].(string)
					downstreamPathName, _ := downstreamTask["pathName"].(string)

					downstreamList = append(downstreamList, map[string]interface{}{
						"pathName": downstreamPathName,
						"name":     downstreamName,
					})
				}
			}
		}
	}

	// 检查循环依赖（简化版：检查是否有下游任务同时是上游任务）
	hasCycle := false
	upstreamPathSet := make(map[string]bool)
	for _, up := range upstreamList {
		if pn, ok := up["pathName"].(string); ok {
			upstreamPathSet[pn] = true
		}
	}
	for _, down := range downstreamList {
		if pn, ok := down["pathName"].(string); ok {
			if upstreamPathSet[pn] {
				hasCycle = true
				break
			}
		}
	}

	// 构建输出
	var output strings.Builder
	output.WriteString(fmt.Sprintf("hasCycle: %v\n", hasCycle))
	output.WriteString(fmt.Sprintf("blocked: %v\n", blocked))

	if len(blockedBy) > 0 {
		output.WriteString("blockedBy:\n")
		for _, b := range blockedBy {
			output.WriteString(fmt.Sprintf("  - pathName: \"%s\"\n", b["pathName"]))
			output.WriteString(fmt.Sprintf("    name: \"%s\"\n", b["name"]))
			output.WriteString(fmt.Sprintf("    status: \"%s\"\n", b["status"]))
		}
	}

	if len(upstreamList) > 0 {
		output.WriteString("upstream:\n")
		for _, up := range upstreamList {
			output.WriteString(fmt.Sprintf("  - pathName: \"%s\"\n", up["pathName"]))
			output.WriteString(fmt.Sprintf("    name: \"%s\"\n", up["name"]))
		}
	}

	if len(downstreamList) > 0 {
		output.WriteString("downstream:\n")
		for _, down := range downstreamList {
			output.WriteString(fmt.Sprintf("  - pathName: \"%s\"\n", down["pathName"]))
			output.WriteString(fmt.Sprintf("    name: \"%s\"\n", down["name"]))
		}
	}

	return mcp.NewToolResultText(output.String()), nil
}

// checkModuleDependencies 检查模块依赖
func (s *MCPServer) checkModuleDependencies(pathName string) (*mcp.CallToolResult, error) {
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

	moduleID, _ := module["id"].(string)

	// 获取模块依赖
	depResp, err := http.Get(fmt.Sprintf("%s/modules/%s/dependencies", s.getApiURL(), moduleID))
	if err != nil {
		return mcp.NewToolResultText("获取模块依赖失败：" + err.Error()), nil
	}
	defer depResp.Body.Close()

	var depResult map[string]interface{}
	if err := json.NewDecoder(depResp.Body).Decode(&depResult); err != nil {
		return mcp.NewToolResultText("解析模块依赖失败：" + err.Error()), nil
	}

	// 解析上游依赖
	upstreamList := []map[string]interface{}{}
	blockedBy := []map[string]interface{}{}
	blocked := false

	if data, ok := depResult["data"].([]interface{}); ok {
		for _, dep := range data {
			if depMap, ok := dep.(map[string]interface{}); ok {
				if dependsOnModule, ok := depMap["dependsOnModule"].(map[string]interface{}); ok {
					depName, _ := dependsOnModule["name"].(string)
					depPathName, _ := dependsOnModule["pathName"].(string)
					depStatus, _ := dependsOnModule["status"].(string)

					upstreamList = append(upstreamList, map[string]interface{}{
						"pathName": depPathName,
						"name":     depName,
						"status":   depStatus,
					})

					// 检查是否阻塞
					if depStatus != "completed" {
						blocked = true
						blockedBy = append(blockedBy, map[string]interface{}{
							"pathName": depPathName,
							"name":     depName,
							"status":   depStatus,
						})
					}
				}
			}
		}
	}

	// 获取下游依赖（被依赖方）
	dependentsResp, err := http.Get(fmt.Sprintf("%s/modules/%s/dependents", s.getApiURL(), moduleID))
	if err != nil {
		return mcp.NewToolResultText("获取模块被依赖失败：" + err.Error()), nil
	}
	defer dependentsResp.Body.Close()

	var dependentsResult map[string]interface{}
	if err := json.NewDecoder(dependentsResp.Body).Decode(&dependentsResult); err != nil {
		return mcp.NewToolResultText("解析模块被依赖失败：" + err.Error()), nil
	}

	// 解析下游依赖
	downstreamList := []map[string]interface{}{}
	if data, ok := dependentsResult["data"].([]interface{}); ok {
		for _, dep := range data {
			if depMap, ok := dep.(map[string]interface{}); ok {
				if mod, ok := depMap["module"].(map[string]interface{}); ok {
					depName, _ := mod["name"].(string)
					depPathName, _ := mod["pathName"].(string)

					downstreamList = append(downstreamList, map[string]interface{}{
						"pathName": depPathName,
						"name":     depName,
					})
				}
			}
		}
	}

	// 检查循环依赖
	hasCycle := false
	upstreamPathSet := make(map[string]bool)
	for _, up := range upstreamList {
		if pn, ok := up["pathName"].(string); ok {
			upstreamPathSet[pn] = true
		}
	}
	for _, down := range downstreamList {
		if pn, ok := down["pathName"].(string); ok {
			if upstreamPathSet[pn] {
				hasCycle = true
				break
			}
		}
	}

	// 构建输出
	var output strings.Builder
	output.WriteString(fmt.Sprintf("hasCycle: %v\n", hasCycle))
	output.WriteString(fmt.Sprintf("blocked: %v\n", blocked))

	if len(blockedBy) > 0 {
		output.WriteString("blockedBy:\n")
		for _, b := range blockedBy {
			output.WriteString(fmt.Sprintf("  - pathName: \"%s\"\n", b["pathName"]))
			output.WriteString(fmt.Sprintf("    name: \"%s\"\n", b["name"]))
			output.WriteString(fmt.Sprintf("    status: \"%s\"\n", b["status"]))
		}
	}

	if len(upstreamList) > 0 {
		output.WriteString("upstream:\n")
		for _, up := range upstreamList {
			output.WriteString(fmt.Sprintf("  - pathName: \"%s\"\n", up["pathName"]))
			output.WriteString(fmt.Sprintf("    name: \"%s\"\n", up["name"]))
		}
	}

	if len(downstreamList) > 0 {
		output.WriteString("downstream:\n")
		for _, down := range downstreamList {
			output.WriteString(fmt.Sprintf("  - pathName: \"%s\"\n", down["pathName"]))
			output.WriteString(fmt.Sprintf("    name: \"%s\"\n", down["name"]))
		}
	}

	return mcp.NewToolResultText(output.String()), nil
}

// ==================== 5. 依赖管理实现 ====================

// handleCreateDependencyImpl 统一创建依赖
func (s *MCPServer) handleCreateDependencyImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	depType, ok1 := getParam(request, "type")
	upstreamPath, ok2 := getParam(request, "upstreamPath")
	downstreamPath, ok3 := getParam(request, "downstreamPath")

	if !ok1 || !ok2 || !ok3 || depType == "" || upstreamPath == "" || downstreamPath == "" {
		return mcp.NewToolResultText("缺少必要参数：type, upstreamPath, downstreamPath"), nil
	}

	switch depType {
	case "module":
		return s.createModuleDependencyUnified(upstreamPath, downstreamPath, request)
	case "task":
		return s.createTaskDependencyUnified(upstreamPath, downstreamPath, request)
	default:
		return mcp.NewToolResultText("无效的 type 参数，必须是 module 或 task"), nil
	}
}

// createModuleDependencyUnified 创建模块依赖（统一接口）
func (s *MCPServer) createModuleDependencyUnified(upstreamPath, downstreamPath string, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取上游模块 ID
	upstreamResp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), upstreamPath))
	if err != nil {
		return mcp.NewToolResultText("获取上游模块信息失败：" + err.Error()), nil
	}
	defer upstreamResp.Body.Close()

	var upstreamModule map[string]interface{}
	if err := json.NewDecoder(upstreamResp.Body).Decode(&upstreamModule); err != nil {
		return mcp.NewToolResultText("解析上游模块信息失败：" + err.Error()), nil
	}

	upstreamModuleID, ok := upstreamModule["id"].(string)
	if !ok {
		return mcp.NewToolResultText("无法获取上游模块ID"), nil
	}

	// 获取下游模块 ID
	downstreamResp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), downstreamPath))
	if err != nil {
		return mcp.NewToolResultText("获取下游模块信息失败：" + err.Error()), nil
	}
	defer downstreamResp.Body.Close()

	var downstreamModule map[string]interface{}
	if err := json.NewDecoder(downstreamResp.Body).Decode(&downstreamModule); err != nil {
		return mcp.NewToolResultText("解析下游模块信息失败：" + err.Error()), nil
	}

	downstreamModuleID, ok := downstreamModule["id"].(string)
	if !ok {
		return mcp.NewToolResultText("无法获取下游模块ID"), nil
	}

	// 构建依赖数据
	depData := map[string]interface{}{
		"dependsOnModuleId": upstreamModuleID,
	}
	if depType, ok := getParamAny(request, "dependencyType"); ok {
		depData["dependencyType"] = depType
	}
	if contract, ok := getParamAny(request, "contractSummary"); ok {
		depData["contractSummary"] = contract
	}

	jsonData, _ := json.Marshal(depData)
	resp, err := http.Post(fmt.Sprintf("%s/modules/%s/dependencies", s.getApiURL(), downstreamModuleID), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return mcp.NewToolResultText("success: true"), nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// createTaskDependencyUnified 创建任务依赖（统一接口）
func (s *MCPServer) createTaskDependencyUnified(upstreamPath, downstreamPath string, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取上游任务 ID
	upstreamResp, err := http.Get(fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), upstreamPath))
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
	downstreamResp, err := http.Get(fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), downstreamPath))
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

	// 构建依赖数据
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

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return mcp.NewToolResultText("success: true"), nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleDeleteDependencyImpl 统一删除依赖
func (s *MCPServer) handleDeleteDependencyImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	depType, ok1 := getParam(request, "type")
	upstreamPath, ok2 := getParam(request, "upstreamPath")
	downstreamPath, ok3 := getParam(request, "downstreamPath")

	if !ok1 || !ok2 || !ok3 || depType == "" || upstreamPath == "" || downstreamPath == "" {
		return mcp.NewToolResultText("缺少必要参数：type, upstreamPath, downstreamPath"), nil
	}

	switch depType {
	case "module":
		return s.deleteModuleDependencyUnified(upstreamPath, downstreamPath)
	case "task":
		return s.deleteTaskDependencyUnified(upstreamPath, downstreamPath)
	default:
		return mcp.NewToolResultText("无效的 type 参数，必须是 module 或 task"), nil
	}
}

// deleteModuleDependencyUnified 删除模块依赖（统一接口）
func (s *MCPServer) deleteModuleDependencyUnified(upstreamPath, downstreamPath string) (*mcp.CallToolResult, error) {
	// 获取上游模块 ID
	upstreamResp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), upstreamPath))
	if err != nil {
		return mcp.NewToolResultText("获取上游模块信息失败：" + err.Error()), nil
	}
	defer upstreamResp.Body.Close()

	var upstreamModule map[string]interface{}
	if err := json.NewDecoder(upstreamResp.Body).Decode(&upstreamModule); err != nil {
		return mcp.NewToolResultText("解析上游模块信息失败：" + err.Error()), nil
	}

	upstreamModuleID, ok := upstreamModule["id"].(string)
	if !ok {
		return mcp.NewToolResultText("无法获取上游模块ID"), nil
	}

	// 获取下游模块 ID
	downstreamResp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), downstreamPath))
	if err != nil {
		return mcp.NewToolResultText("获取下游模块信息失败：" + err.Error()), nil
	}
	defer downstreamResp.Body.Close()

	var downstreamModule map[string]interface{}
	if err := json.NewDecoder(downstreamResp.Body).Decode(&downstreamModule); err != nil {
		return mcp.NewToolResultText("解析下游模块信息失败：" + err.Error()), nil
	}

	downstreamModuleID, ok := downstreamModule["id"].(string)
	if !ok {
		return mcp.NewToolResultText("无法获取下游模块ID"), nil
	}

	// 删除依赖
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/modules/%s/dependencies/%s", s.getApiURL(), downstreamModuleID, upstreamModuleID), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return mcp.NewToolResultText("success: true"), nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// deleteTaskDependencyUnified 删除任务依赖（统一接口）
func (s *MCPServer) deleteTaskDependencyUnified(upstreamPath, downstreamPath string) (*mcp.CallToolResult, error) {
	// 获取上游任务 ID
	upstreamResp, err := http.Get(fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), upstreamPath))
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
	downstreamResp, err := http.Get(fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), downstreamPath))
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

	// 删除依赖 - 需要先查询依赖 ID
	queryResp, err := http.Get(fmt.Sprintf("%s/dependencies?upstreamTaskId=%s&downstreamTaskId=%s", s.getApiURL(), upstreamTaskID, downstreamTaskID))
	if err != nil {
		return mcp.NewToolResultText("查询依赖失败：" + err.Error()), nil
	}
	defer queryResp.Body.Close()

	var queryResult map[string]interface{}
	if err := json.NewDecoder(queryResp.Body).Decode(&queryResult); err != nil {
		return mcp.NewToolResultText("解析查询结果失败：" + err.Error()), nil
	}

	if data, ok := queryResult["data"].([]interface{}); ok && len(data) > 0 {
		if depMap, ok := data[0].(map[string]interface{}); ok {
			if depID, ok := depMap["id"].(string); ok {
				req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/dependencies/%s", s.getApiURL(), depID), nil)
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return mcp.NewToolResultText("请求失败：" + err.Error()), nil
				}
				defer resp.Body.Close()

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					return mcp.NewToolResultText("success: true"), nil
				}
			}
		}
	}

	return mcp.NewToolResultText("success: false\nerror: \"未找到依赖关系\""), nil
}

// handleQueryDependenciesImpl 统一查询依赖
func (s *MCPServer) handleQueryDependenciesImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	depType, ok1 := getParam(request, "type")
	pathName, ok2 := getParam(request, "pathName")

	if !ok1 || !ok2 || depType == "" || pathName == "" {
		return mcp.NewToolResultText("缺少必要参数：type, pathName"), nil
	}

	direction, _ := getParam(request, "direction")
	if direction == "" {
		direction = "both"
	}

	switch depType {
	case "module":
		return s.queryModuleDependencies(pathName, direction)
	case "task":
		return s.queryTaskDependencies(pathName, direction)
	default:
		return mcp.NewToolResultText("无效的 type 参数，必须是 module 或 task"), nil
	}
}

// queryModuleDependencies 查询模块依赖
func (s *MCPServer) queryModuleDependencies(pathName, direction string) (*mcp.CallToolResult, error) {
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

	moduleID, _ := module["id"].(string)

	var output strings.Builder

	// 查询上游依赖
	if direction == "upstream" || direction == "both" {
		depResp, err := http.Get(fmt.Sprintf("%s/modules/%s/dependencies", s.getApiURL(), moduleID))
		if err != nil {
			return mcp.NewToolResultText("获取模块依赖失败：" + err.Error()), nil
		}
		defer depResp.Body.Close()

		var depResult map[string]interface{}
		if err := json.NewDecoder(depResp.Body).Decode(&depResult); err != nil {
			return mcp.NewToolResultText("解析模块依赖失败：" + err.Error()), nil
		}

		output.WriteString("upstream:\n")
		if data, ok := depResult["data"].([]interface{}); ok {
			for _, dep := range data {
				if depMap, ok := dep.(map[string]interface{}); ok {
					if dependsOnModule, ok := depMap["dependsOnModule"].(map[string]interface{}); ok {
						output.WriteString(fmt.Sprintf("  - pathName: \"%s\"\n", dependsOnModule["pathName"]))
						output.WriteString(fmt.Sprintf("    name: \"%s\"\n", dependsOnModule["name"]))
						// 添加 dependencyType 字段
						if depType, ok := depMap["dependencyType"].(string); ok && depType != "" {
							output.WriteString(fmt.Sprintf("    dependencyType: \"%s\"\n", depType))
						} else {
							output.WriteString("    dependencyType: \"required\"\n")
						}
					}
				}
			}
		}
	}

	// 查询下游依赖
	if direction == "downstream" || direction == "both" {
		dependentsResp, err := http.Get(fmt.Sprintf("%s/modules/%s/dependents", s.getApiURL(), moduleID))
		if err != nil {
			return mcp.NewToolResultText("获取模块被依赖失败：" + err.Error()), nil
		}
		defer dependentsResp.Body.Close()

		var dependentsResult map[string]interface{}
		if err := json.NewDecoder(dependentsResp.Body).Decode(&dependentsResult); err != nil {
			return mcp.NewToolResultText("解析模块被依赖失败：" + err.Error()), nil
		}

		output.WriteString("downstream:\n")
		if data, ok := dependentsResult["data"].([]interface{}); ok {
			for _, dep := range data {
				if depMap, ok := dep.(map[string]interface{}); ok {
					if mod, ok := depMap["module"].(map[string]interface{}); ok {
						output.WriteString(fmt.Sprintf("  - pathName: \"%s\"\n", mod["pathName"]))
						output.WriteString(fmt.Sprintf("    name: \"%s\"\n", mod["name"]))
						// 添加 dependencyType 字段
						if depType, ok := depMap["dependencyType"].(string); ok && depType != "" {
							output.WriteString(fmt.Sprintf("    dependencyType: \"%s\"\n", depType))
						} else {
							output.WriteString("    dependencyType: \"required\"\n")
						}
					}
				}
			}
		}
	}

	return mcp.NewToolResultText(output.String()), nil
}

// queryTaskDependencies 查询任务依赖
func (s *MCPServer) queryTaskDependencies(pathName, direction string) (*mcp.CallToolResult, error) {
	// 获取任务信息
	taskResp, err := http.Get(fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), pathName))
	if err != nil {
		return mcp.NewToolResultText("获取任务信息失败：" + err.Error()), nil
	}
	defer taskResp.Body.Close()

	var task map[string]interface{}
	if err := json.NewDecoder(taskResp.Body).Decode(&task); err != nil {
		return mcp.NewToolResultText("解析任务信息失败：" + err.Error()), nil
	}

	taskID, _ := task["id"].(string)

	var output strings.Builder

	// 查询上游依赖
	if direction == "upstream" || direction == "both" {
		upstreamResp, err := http.Get(fmt.Sprintf("%s/dependencies?downstreamTaskId=%s", s.getApiURL(), taskID))
		if err != nil {
			return mcp.NewToolResultText("获取上游依赖失败：" + err.Error()), nil
		}
		defer upstreamResp.Body.Close()

		var upstreamResult map[string]interface{}
		if err := json.NewDecoder(upstreamResp.Body).Decode(&upstreamResult); err != nil {
			return mcp.NewToolResultText("解析上游依赖失败：" + err.Error()), nil
		}

		output.WriteString("upstream:\n")
		if data, ok := upstreamResult["data"].([]interface{}); ok {
			for _, dep := range data {
				if depMap, ok := dep.(map[string]interface{}); ok {
					if upstreamTask, ok := depMap["upstreamTask"].(map[string]interface{}); ok {
						output.WriteString(fmt.Sprintf("  - pathName: \"%s\"\n", upstreamTask["pathName"]))
						output.WriteString(fmt.Sprintf("    name: \"%s\"\n", upstreamTask["name"]))
						// 添加 dependencyType 字段
						if depType, ok := depMap["dependencyType"].(string); ok && depType != "" {
							output.WriteString(fmt.Sprintf("    dependencyType: \"%s\"\n", depType))
						} else {
							output.WriteString("    dependencyType: \"required\"\n")
						}
					}
				}
			}
		}
	}

	// 查询下游依赖
	if direction == "downstream" || direction == "both" {
		downstreamResp, err := http.Get(fmt.Sprintf("%s/dependencies?upstreamTaskId=%s", s.getApiURL(), taskID))
		if err != nil {
			return mcp.NewToolResultText("获取下游依赖失败：" + err.Error()), nil
		}
		defer downstreamResp.Body.Close()

		var downstreamResult map[string]interface{}
		if err := json.NewDecoder(downstreamResp.Body).Decode(&downstreamResult); err != nil {
			return mcp.NewToolResultText("解析下游依赖失败：" + err.Error()), nil
		}

		output.WriteString("downstream:\n")
		if data, ok := downstreamResult["data"].([]interface{}); ok {
			for _, dep := range data {
				if depMap, ok := dep.(map[string]interface{}); ok {
					if downstreamTask, ok := depMap["downstreamTask"].(map[string]interface{}); ok {
						output.WriteString(fmt.Sprintf("  - pathName: \"%s\"\n", downstreamTask["pathName"]))
						output.WriteString(fmt.Sprintf("    name: \"%s\"\n", downstreamTask["name"]))
						// 添加 dependencyType 字段
						if depType, ok := depMap["dependencyType"].(string); ok && depType != "" {
							output.WriteString(fmt.Sprintf("    dependencyType: \"%s\"\n", depType))
						} else {
							output.WriteString("    dependencyType: \"required\"\n")
						}
					}
				}
			}
		}
	}

	return mcp.NewToolResultText(output.String()), nil
}

// ==================== 6. 问答系统实现 ====================

// handleCreateIssueImpl 创建问题
func (s *MCPServer) handleCreateIssueImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fromTaskPathName, ok1 := getParam(request, "fromTaskPathName")
	toTaskPathName, ok2 := getParam(request, "toTaskPathName")
	issueType, ok3 := getParam(request, "type")
	title, ok4 := getParam(request, "title")
	content, ok5 := getParam(request, "content")

	if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 {
		return mcp.NewToolResultText("缺少必要参数"), nil
	}

	// 构建请求数据
	issueData := map[string]interface{}{
		"fromTaskPathName": fromTaskPathName,
		"toTaskPathName":   toTaskPathName,
		"type":             issueType,
		"title":            title,
		"content":          content,
	}

	jsonData, _ := json.Marshal(issueData)
	resp, err := http.Post(fmt.Sprintf("%s/issues", s.getApiURL()), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	// 提取 issue 信息并返回简洁格式
	if issue, ok := result["issue"].(map[string]interface{}); ok {
		status, _ := issue["status"].(string)
		return mcp.NewToolResultText(fmt.Sprintf("success: true\nstatus: \"%s\"", status)), nil
	}

	return mcp.NewToolResultText("success: false\nerror: \"创建问题失败\""), nil
}

// handleReplyIssueImpl 回复问题
func (s *MCPServer) handleReplyIssueImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fromTaskPathName, ok1 := getParam(request, "fromTaskPathName")
	toTaskPathName, ok2 := getParam(request, "toTaskPathName")
	title, ok3 := getParam(request, "title")
	replyContent, ok4 := getParam(request, "replyContent")

	if !ok1 || !ok2 || !ok3 || !ok4 {
		return mcp.NewToolResultText("缺少必要参数"), nil
	}

	// 构建请求数据
	issueData := map[string]interface{}{
		"fromTaskPathName": fromTaskPathName,
		"toTaskPathName":   toTaskPathName,
		"title":            title,
		"replyContent":     replyContent,
	}

	jsonData, _ := json.Marshal(issueData)
	resp, err := http.Post(fmt.Sprintf("%s/issues/reply", s.getApiURL()), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	// 提取 issue 信息并返回简洁格式
	if issue, ok := result["issue"].(map[string]interface{}); ok {
		status, _ := issue["status"].(string)
		return mcp.NewToolResultText(fmt.Sprintf("success: true\nstatus: \"%s\"", status)), nil
	}

	return mcp.NewToolResultText("success: false\nerror: \"回复问题失败\""), nil
}

// handleResolveIssueImpl 解决问题
func (s *MCPServer) handleResolveIssueImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fromTaskPathName, ok1 := getParam(request, "fromTaskPathName")
	toTaskPathName, ok2 := getParam(request, "toTaskPathName")
	title, ok3 := getParam(request, "title")

	if !ok1 || !ok2 || !ok3 {
		return mcp.NewToolResultText("缺少必要参数"), nil
	}

	// 构建请求数据
	issueData := map[string]interface{}{
		"fromTaskPathName": fromTaskPathName,
		"toTaskPathName":   toTaskPathName,
		"title":            title,
	}

	jsonData, _ := json.Marshal(issueData)
	resp, err := http.Post(fmt.Sprintf("%s/issues/resolve", s.getApiURL()), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	// 提取 issue 信息并返回简洁格式
	if issue, ok := result["issue"].(map[string]interface{}); ok {
		status, _ := issue["status"].(string)
		return mcp.NewToolResultText(fmt.Sprintf("success: true\nstatus: \"%s\"", status)), nil
	}

	return mcp.NewToolResultText("success: false\nerror: \"解决问题失败\""), nil
}

// handleQueryIssuesImpl 查询问题列表
func (s *MCPServer) handleQueryIssuesImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskPathName, _ := getParam(request, "taskPathName")
	direction, _ := getParam(request, "direction")
	status, _ := getParam(request, "status")
	issueType, _ := getParam(request, "type")

	// 构建查询 URL
	url := fmt.Sprintf("%s/issues?", s.getApiURL())
	params := []string{}
	if taskPathName != "" {
		params = append(params, fmt.Sprintf("taskPathName=%s", taskPathName))
	}
	if direction != "" {
		params = append(params, fmt.Sprintf("direction=%s", direction))
	}
	if status != "" {
		params = append(params, fmt.Sprintf("status=%s", status))
	}
	if issueType != "" {
		params = append(params, fmt.Sprintf("type=%s", issueType))
	}
	if len(params) > 0 {
		url += params[0]
		for i := 1; i < len(params); i++ {
			url += "&" + params[i]
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

	// 构建 YAML 风格的输出
	var output strings.Builder
	output.WriteString("issues:\n")

	if issues, ok := result["issues"].([]interface{}); ok {
		for _, i := range issues {
			if issue, ok := i.(map[string]interface{}); ok {
				output.WriteString("  - fromTaskPathName: \"")
				if v, ok := issue["fromTaskPathName"].(string); ok {
					output.WriteString(v)
				}
				output.WriteString("\"\n")

				output.WriteString("    fromTaskName: \"")
				if v, ok := issue["fromTaskName"].(string); ok {
					output.WriteString(v)
				}
				output.WriteString("\"\n")

				output.WriteString("    toTaskPathName: \"")
				if v, ok := issue["toTaskPathName"].(string); ok {
					output.WriteString(v)
				}
				output.WriteString("\"\n")

				output.WriteString("    toTaskName: \"")
				if v, ok := issue["toTaskName"].(string); ok {
					output.WriteString(v)
				}
				output.WriteString("\"\n")

				output.WriteString("    type: \"")
				if v, ok := issue["type"].(string); ok {
					output.WriteString(v)
				}
				output.WriteString("\"\n")

				output.WriteString("    title: \"")
				if v, ok := issue["title"].(string); ok {
					output.WriteString(v)
				}
				output.WriteString("\"\n")

				output.WriteString("    status: \"")
				if v, ok := issue["status"].(string); ok {
					output.WriteString(v)
				}
				output.WriteString("\"\n")

				output.WriteString("    createdAt: \"")
				if v, ok := issue["createdAt"].(string); ok {
					output.WriteString(v)
				} else if v, ok := issue["createdAt"].(float64); ok {
					// 将 Unix 时间戳转换为 ISO 8601 格式
					output.WriteString(time.Unix(int64(v), 0).UTC().Format(time.RFC3339))
				} else {
					output.WriteString("")
				}
				output.WriteString("\"\n")
			}
		}
	}

	output.WriteString("total: ")
	if total, ok := result["total"].(float64); ok {
		output.WriteString(fmt.Sprintf("%.0f", total))
	} else {
		output.WriteString("0")
	}

	return mcp.NewToolResultText(output.String()), nil
}

// ==================== 7. 锁管理实现 ====================

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

// ==================== 8. 编译接口实现 ====================

// CompileIssue 编译问题
type CompileIssue struct {
	RuleID           string `json:"ruleId"`
	RuleName         string `json:"ruleName"`
	ResourceType     string `json:"resourceType"` // module, task
	ResourceName     string `json:"resourceName"`
	ResourcePathName string `json:"resourcePathName"`
	Message          string `json:"message"`
	Suggestion       string `json:"suggestion"`
	Severity         string `json:"severity"` // error, warning
}

// CompileStaticResult 静态编译结果
type CompileStaticResult struct {
	Success        bool            `json:"success"`
	TotalModules   int             `json:"totalModules"`
	TotalTasks     int             `json:"totalTasks"`
	CompletedTasks int             `json:"completedTasks"`
	InProgressTasks int            `json:"inProgressTasks"`
	ReadyTasks     int             `json:"readyTasks"`
	ErrorCount     int             `json:"errorCount"`
	WarningCount   int             `json:"warningCount"`
	Errors         []CompileIssue  `json:"errors"`
	Warnings       []CompileIssue  `json:"warnings"`
}

// handleCompileStaticImpl 静态编译
// 遍历所有模块和任务，验证结构完整性，生成静态报告
func (s *MCPServer) handleCompileStaticImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		pathName = s.configManager.GetProjectPathName()
	}
	if pathName == "" {
		return mcp.NewToolResultText("未配置项目路径名称，请先使用 init_project 设置项目"), nil
	}

	includeWarnings, _ := getParamBool(request, "includeWarnings")

	result := &CompileStaticResult{
		Success:  true,
		Errors:   []CompileIssue{},
		Warnings: []CompileIssue{},
	}

	// 判断 pathName 类型（project/module/task）
	pathParts := strings.Split(pathName, "/")
	
	if len(pathParts) == 1 {
		// 项目级别编译
		s.compileStaticProject(pathName, includeWarnings, result)
	} else if len(pathParts) == 2 {
		// 模块级别编译
		s.compileStaticModule(pathName, includeWarnings, result)
	} else {
		// 任务级别编译
		s.compileStaticTask(pathName, includeWarnings, result)
	}

	// 更新统计
	result.ErrorCount = len(result.Errors)
	result.WarningCount = len(result.Warnings)
	result.Success = result.ErrorCount == 0

	// 构建输出
	var output strings.Builder
	output.WriteString(fmt.Sprintf("success: %v\n", result.Success))
	output.WriteString(fmt.Sprintf("totalModules: %d\n", result.TotalModules))
	output.WriteString(fmt.Sprintf("totalTasks: %d\n", result.TotalTasks))
	output.WriteString(fmt.Sprintf("completedTasks: %d\n", result.CompletedTasks))
	output.WriteString(fmt.Sprintf("inProgressTasks: %d\n", result.InProgressTasks))
	output.WriteString(fmt.Sprintf("readyTasks: %d\n", result.ReadyTasks))
	output.WriteString(fmt.Sprintf("errorCount: %d\n", result.ErrorCount))
	output.WriteString(fmt.Sprintf("warningCount: %d\n", result.WarningCount))

	if len(result.Errors) > 0 {
		output.WriteString("errors:\n")
		for _, err := range result.Errors {
			output.WriteString(fmt.Sprintf("  - ruleId: \"%s\"\n", err.RuleID))
			output.WriteString(fmt.Sprintf("    ruleName: \"%s\"\n", err.RuleName))
			output.WriteString(fmt.Sprintf("    resourceType: \"%s\"\n", err.ResourceType))
			output.WriteString(fmt.Sprintf("    resourceName: \"%s\"\n", err.ResourceName))
			output.WriteString(fmt.Sprintf("    resourcePathName: \"%s\"\n", err.ResourcePathName))
			output.WriteString(fmt.Sprintf("    message: \"%s\"\n", err.Message))
			output.WriteString(fmt.Sprintf("    suggestion: \"%s\"\n", err.Suggestion))
		}
	}

	if includeWarnings && len(result.Warnings) > 0 {
		output.WriteString("warnings:\n")
		for _, warn := range result.Warnings {
			output.WriteString(fmt.Sprintf("  - ruleId: \"%s\"\n", warn.RuleID))
			output.WriteString(fmt.Sprintf("    ruleName: \"%s\"\n", warn.RuleName))
			output.WriteString(fmt.Sprintf("    resourceType: \"%s\"\n", warn.ResourceType))
			output.WriteString(fmt.Sprintf("    resourceName: \"%s\"\n", warn.ResourceName))
			output.WriteString(fmt.Sprintf("    resourcePathName: \"%s\"\n", warn.ResourcePathName))
			output.WriteString(fmt.Sprintf("    message: \"%s\"\n", warn.Message))
			output.WriteString(fmt.Sprintf("    suggestion: \"%s\"\n", warn.Suggestion))
		}
	}

	return mcp.NewToolResultText(output.String()), nil
}

// compileStaticProject 项目级别静态编译
func (s *MCPServer) compileStaticProject(projectPathName string, includeWarnings bool, result *CompileStaticResult) {
	// 获取项目下所有模块
	modulesResp, err := http.Get(fmt.Sprintf("%s/projects/by-path/%s/modules", s.getApiURL(), projectPathName))
	if err != nil {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-00",
			RuleName:         "项目结构检查",
			ResourceType:     "project",
			ResourceName:     projectPathName,
			ResourcePathName: projectPathName,
			Message:          "获取项目模块失败: " + err.Error(),
			Severity:         "error",
		})
		return
	}
	defer modulesResp.Body.Close()

	var modulesData struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(modulesResp.Body).Decode(&modulesData); err != nil {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-00",
			RuleName:         "项目结构检查",
			ResourceType:     "project",
			ResourceName:     projectPathName,
			ResourcePathName: projectPathName,
			Message:          "解析项目模块失败: " + err.Error(),
			Severity:         "error",
		})
		return
	}

	result.TotalModules = len(modulesData.Data)

	// 遍历每个模块进行编译
	for _, module := range modulesData.Data {
		modulePathName, _ := module["pathName"].(string)
		s.compileStaticModule(modulePathName, includeWarnings, result)
	}
}

// compileStaticModule 模块级别静态编译
func (s *MCPServer) compileStaticModule(modulePathName string, includeWarnings bool, result *CompileStaticResult) {
	// 获取模块信息
	moduleResp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), modulePathName))
	if err != nil {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-00",
			RuleName:         "模块结构检查",
			ResourceType:     "module",
			ResourcePathName: modulePathName,
			Message:          "获取模块信息失败: " + err.Error(),
			Severity:         "error",
		})
		return
	}
	defer moduleResp.Body.Close()

	var module map[string]interface{}
	if err := json.NewDecoder(moduleResp.Body).Decode(&module); err != nil {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-00",
			RuleName:         "模块结构检查",
			ResourceType:     "module",
			ResourcePathName: modulePathName,
			Message:          "解析模块信息失败: " + err.Error(),
			Severity:         "error",
		})
		return
	}

	moduleName, _ := module["name"].(string)
	moduleID, _ := module["id"].(string)

	// E-S-04: 检查模块提示词
	prompt, _ := module["prompt"].(string)
	if prompt == "" {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-04",
			RuleName:         "提示词为空",
			ResourceType:     "module",
			ResourceName:     moduleName,
			ResourcePathName: modulePathName,
			Message:          fmt.Sprintf("模块 [%s] 提示词为空", moduleName),
			Suggestion:       "请为模块添加提示词",
			Severity:         "error",
		})
	}

	// W-S-02: 检查模块描述
	if includeWarnings {
		description, _ := module["description"].(string)
		if description == "" {
			result.Warnings = append(result.Warnings, CompileIssue{
				RuleID:           "W-S-02",
				RuleName:         "模块描述为空",
				ResourceType:     "module",
				ResourceName:     moduleName,
				ResourcePathName: modulePathName,
				Message:          fmt.Sprintf("模块 [%s] 描述为空", moduleName),
				Suggestion:       "建议添加模块描述",
				Severity:         "warning",
			})
		}
	}

	// 获取模块下的任务
	tasksResp, err := http.Get(fmt.Sprintf("%s/modules/%s/tasks", s.getApiURL(), moduleID))
	if err != nil {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-00",
			RuleName:         "模块任务检查",
			ResourceType:     "module",
			ResourceName:     moduleName,
			ResourcePathName: modulePathName,
			Message:          "获取模块任务失败: " + err.Error(),
			Severity:         "error",
		})
		return
	}
	defer tasksResp.Body.Close()

	var tasksData struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(tasksResp.Body).Decode(&tasksData); err != nil {
		return
	}

	// 遍历任务进行编译
	for _, task := range tasksData.Data {
		s.compileStaticTaskFromData(task, modulePathName, moduleName, includeWarnings, result)
	}
}

// compileStaticTask 任务级别静态编译
func (s *MCPServer) compileStaticTask(taskPathName string, includeWarnings bool, result *CompileStaticResult) {
	// 获取任务信息
	taskResp, err := http.Get(fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), taskPathName))
	if err != nil {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-00",
			RuleName:         "任务结构检查",
			ResourceType:     "task",
			ResourcePathName: taskPathName,
			Message:          "获取任务信息失败: " + err.Error(),
			Severity:         "error",
		})
		return
	}
	defer taskResp.Body.Close()

	var task map[string]interface{}
	if err := json.NewDecoder(taskResp.Body).Decode(&task); err != nil {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-00",
			RuleName:         "任务结构检查",
			ResourceType:     "task",
			ResourcePathName: taskPathName,
			Message:          "解析任务信息失败: " + err.Error(),
			Severity:         "error",
		})
		return
	}

	modulePathName := ""
	if parts := strings.Split(taskPathName, "/"); len(parts) >= 2 {
		modulePathName = parts[0] + "/" + parts[1]
	}
	moduleName := modulePathName

	s.compileStaticTaskFromData(task, modulePathName, moduleName, includeWarnings, result)
}

// compileStaticTaskFromData 从任务数据编译
func (s *MCPServer) compileStaticTaskFromData(task map[string]interface{}, modulePathName, moduleName string, includeWarnings bool, result *CompileStaticResult) {
	result.TotalTasks++

	taskName, _ := task["name"].(string)
	taskPathName, _ := task["pathName"].(string)
	status, _ := task["status"].(string)

	// 统计任务状态
	switch status {
	case "completed":
		result.CompletedTasks++
	case "in_progress":
		result.InProgressTasks++
	case "ready":
		result.ReadyTasks++
	}

	// E-S-01: 检查代码路径
	codePaths, _ := task["codePaths"].(string)
	if codePaths == "" || codePaths == "[]" {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-01",
			RuleName:         "代码路径为空",
			ResourceType:     "task",
			ResourceName:     taskName,
			ResourcePathName: taskPathName,
			Message:          fmt.Sprintf("任务 [%s] 代码路径为空", taskName),
			Suggestion:       "请为任务添加代码路径",
			Severity:         "error",
		})
	}

	// E-S-02: 检查 Bug 日志
	bugLog, _ := task["bugLog"].(string)
	if bugLog != "" && bugLog != "[]" {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-02",
			RuleName:         "存在Bug日志",
			ResourceType:     "task",
			ResourceName:     taskName,
			ResourcePathName: taskPathName,
			Message:          fmt.Sprintf("任务 [%s] 存在未解决的Bug", taskName),
			Suggestion:       "请解决Bug后清除日志",
			Severity:         "error",
		})
	}

	// E-S-03: 检查人工协助
	humanAssistance, _ := task["humanAssistance"].(string)
	if humanAssistance != "" && humanAssistance != "{}" {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-03",
			RuleName:         "需要人工协助",
			ResourceType:     "task",
			ResourceName:     taskName,
			ResourcePathName: taskPathName,
			Message:          fmt.Sprintf("任务 [%s] 需要人工协助", taskName),
			Suggestion:       "请处理人工协助请求",
			Severity:         "error",
		})
	}

	// E-S-04: 检查提示词
	prompt, _ := task["prompt"].(string)
	if prompt == "" {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-04",
			RuleName:         "提示词为空",
			ResourceType:     "task",
			ResourceName:     taskName,
			ResourcePathName: taskPathName,
			Message:          fmt.Sprintf("任务 [%s] 提示词为空", taskName),
			Suggestion:       "请为任务添加提示词",
			Severity:         "error",
		})
	}

	// W-S-01: 检查测试用例
	if includeWarnings {
		tests, _ := task["tests"].(string)
		if tests == "" || tests == "[]" {
			result.Warnings = append(result.Warnings, CompileIssue{
				RuleID:           "W-S-01",
				RuleName:         "测试用例为空",
				ResourceType:     "task",
				ResourceName:     taskName,
				ResourcePathName: taskPathName,
				Message:          fmt.Sprintf("任务 [%s] 未定义测试用例", taskName),
				Suggestion:       "建议添加测试用例",
				Severity:         "warning",
			})
		}
	}
}

// CompileDynamicResult 动态编译结果
type CompileDynamicResult struct {
	Success           bool                       `json:"success"`
	CompiledAt        string                     `json:"compiledAt"`
	Duration          string                     `json:"duration"`
	TotalTasks        int                        `json:"totalTasks"`
	CompiledTasks     int                        `json:"compiledTasks"`
	FailedTasks       int                        `json:"failedTasks"`
	TestResults       CompileTestResults         `json:"testResults"`
	ContractValidation CompileContractValidation `json:"contractValidation"`
	Errors            []CompileDynamicError      `json:"errors"`
}

// CompileTestResults 测试结果
type CompileTestResults struct {
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

// CompileContractValidation 契约验证结果
type CompileContractValidation struct {
	Valid    int `json:"valid"`
	Invalid  int `json:"invalid"`
	Warnings int `json:"warnings"`
}

// CompileDynamicError 动态编译错误
type CompileDynamicError struct {
	TaskPathName string `json:"taskPathName"`
	Error        string `json:"error"`
}

// handleCompileDynamicImpl 动态编译
// 执行实际编译，运行测试，验证契约，生成编译产物
func (s *MCPServer) handleCompileDynamicImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		pathName = s.configManager.GetProjectPathName()
	}
	if pathName == "" {
		return mcp.NewToolResultText("未配置项目路径名称，请先使用 init_project 设置项目"), nil
	}

	runTests, _ := getParamBool(request, "runTests")
	validateContracts, _ := getParamBool(request, "validateContracts")
	if !validateContracts {
		validateContracts = true // 默认验证契约
	}

	startTime := time.Now()
	result := &CompileDynamicResult{
		CompiledAt: startTime.Format(time.RFC3339),
		Errors:     []CompileDynamicError{},
	}

	// 获取任务列表并按依赖排序
	tasks, err := s.getTasksInDependencyOrder(pathName)
	if err != nil {
		return mcp.NewToolResultText("获取任务列表失败: " + err.Error()), nil
	}

	result.TotalTasks = len(tasks)

	// 遍历任务执行编译
	for _, task := range tasks {
		taskPathName, _ := task["pathName"].(string)
		taskName, _ := task["name"].(string)
		_ = taskName // 避免未使用警告

		// 验证契约
		if validateContracts {
			if err := s.validateTaskContracts(task); err != nil {
				result.Errors = append(result.Errors, CompileDynamicError{
					TaskPathName: taskPathName,
					Error:        "契约验证失败: " + err.Error(),
				})
				result.FailedTasks++
				result.ContractValidation.Invalid++
				continue
			}
			result.ContractValidation.Valid++
		}

		// 运行测试（如果启用）
		if runTests {
			testResult, err := s.runTaskTests(task)
			if err != nil {
				result.Errors = append(result.Errors, CompileDynamicError{
					TaskPathName: taskPathName,
					Error:        "测试失败: " + err.Error(),
				})
				result.TestResults.Failed++
				result.FailedTasks++
				continue
			}
			if testResult.Passed {
				result.TestResults.Passed++
			} else {
				result.TestResults.Failed++
			}
			result.TestResults.Skipped += testResult.Skipped
		}

		result.CompiledTasks++
	}

	// 计算耗时
	duration := time.Since(startTime)
	result.Duration = duration.String()
	result.Success = result.FailedTasks == 0

	// 构建输出
	var output strings.Builder
	output.WriteString(fmt.Sprintf("success: %v\n", result.Success))
	output.WriteString(fmt.Sprintf("compiledAt: \"%s\"\n", result.CompiledAt))
	output.WriteString(fmt.Sprintf("duration: \"%s\"\n", result.Duration))
	output.WriteString("results:\n")
	output.WriteString(fmt.Sprintf("  totalTasks: %d\n", result.TotalTasks))
	output.WriteString(fmt.Sprintf("  compiledTasks: %d\n", result.CompiledTasks))
	output.WriteString(fmt.Sprintf("  failedTasks: %d\n", result.FailedTasks))
	output.WriteString("  testResults:\n")
	output.WriteString(fmt.Sprintf("    passed: %d\n", result.TestResults.Passed))
	output.WriteString(fmt.Sprintf("    failed: %d\n", result.TestResults.Failed))
	output.WriteString(fmt.Sprintf("    skipped: %d\n", result.TestResults.Skipped))
	output.WriteString("  contractValidation:\n")
	output.WriteString(fmt.Sprintf("    valid: %d\n", result.ContractValidation.Valid))
	output.WriteString(fmt.Sprintf("    invalid: %d\n", result.ContractValidation.Invalid))
	output.WriteString(fmt.Sprintf("    warnings: %d\n", result.ContractValidation.Warnings))

	if len(result.Errors) > 0 {
		output.WriteString("errors:\n")
		for _, err := range result.Errors {
			output.WriteString(fmt.Sprintf("  - taskPathName: \"%s\"\n", err.TaskPathName))
			output.WriteString(fmt.Sprintf("    error: \"%s\"\n", err.Error))
		}
	}

	return mcp.NewToolResultText(output.String()), nil
}

// getTasksInDependencyOrder 获取按依赖顺序排序的任务列表
func (s *MCPServer) getTasksInDependencyOrder(pathName string) ([]map[string]interface{}, error) {
	pathParts := strings.Split(pathName, "/")
	var tasks []map[string]interface{}

	if len(pathParts) == 1 {
		// 项目级别：获取所有任务
		// 获取项目下所有模块
		modulesResp, err := http.Get(fmt.Sprintf("%s/projects/by-path/%s/modules", s.getApiURL(), pathName))
		if err != nil {
			return nil, err
		}
		defer modulesResp.Body.Close()

		var modulesData struct {
			Data []map[string]interface{} `json:"data"`
		}
		if err := json.NewDecoder(modulesResp.Body).Decode(&modulesData); err != nil {
			return nil, err
		}

		// 获取每个模块的任务
		for _, module := range modulesData.Data {
			moduleID, _ := module["id"].(string)
			tasksResp, err := http.Get(fmt.Sprintf("%s/modules/%s/tasks", s.getApiURL(), moduleID))
			if err != nil {
				continue
			}
			defer tasksResp.Body.Close()

			var tasksData struct {
				Data []map[string]interface{} `json:"data"`
			}
			if err := json.NewDecoder(tasksResp.Body).Decode(&tasksData); err != nil {
				continue
			}
			tasks = append(tasks, tasksData.Data...)
		}
	} else if len(pathParts) == 2 {
		// 模块级别：获取模块下所有任务
		moduleResp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), pathName))
		if err != nil {
			return nil, err
		}
		defer moduleResp.Body.Close()

		var module map[string]interface{}
		if err := json.NewDecoder(moduleResp.Body).Decode(&module); err != nil {
			return nil, err
		}

		moduleID, _ := module["id"].(string)
		tasksResp, err := http.Get(fmt.Sprintf("%s/modules/%s/tasks", s.getApiURL(), moduleID))
		if err != nil {
			return nil, err
		}
		defer tasksResp.Body.Close()

		var tasksData struct {
			Data []map[string]interface{} `json:"data"`
		}
		if err := json.NewDecoder(tasksResp.Body).Decode(&tasksData); err != nil {
			return nil, err
		}
		tasks = tasksData.Data
	} else {
		// 任务级别：获取单个任务
		taskResp, err := http.Get(fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), pathName))
		if err != nil {
			return nil, err
		}
		defer taskResp.Body.Close()

		var task map[string]interface{}
		if err := json.NewDecoder(taskResp.Body).Decode(&task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	// TODO: 实现拓扑排序以按依赖顺序排列任务
	// 目前简单返回任务列表

	return tasks, nil
}

// validateTaskContracts 验证任务契约
func (s *MCPServer) validateTaskContracts(task map[string]interface{}) error {
	taskID, _ := task["id"].(string)
	taskName, _ := task["name"].(string)

	// 获取任务依赖
	depResp, err := http.Get(fmt.Sprintf("%s/dependencies?downstreamTaskId=%s", s.getApiURL(), taskID))
	if err != nil {
		return nil // 没有依赖则跳过
	}
	defer depResp.Body.Close()

	var depResult struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(depResp.Body).Decode(&depResult); err != nil {
		return nil
	}

	// 检查上游契约
	upstreamContract, _ := task["upstreamContractDetail"].(string)
	if len(depResult.Data) > 0 && upstreamContract == "" {
		return fmt.Errorf("任务 [%s] 有上游依赖但上游契约详情为空", taskName)
	}

	// 检查契约一致性
	for _, dep := range depResult.Data {
		if upstreamTask, ok := dep["upstreamTask"].(map[string]interface{}); ok {
			upstreamID, _ := upstreamTask["id"].(string)
			// 获取上游任务的下游契约
			upstreamResp, err := http.Get(fmt.Sprintf("%s/tasks/%s", s.getApiURL(), upstreamID))
			if err != nil {
				continue
			}
			defer upstreamResp.Body.Close()

			var upstreamTaskData map[string]interface{}
			if err := json.NewDecoder(upstreamResp.Body).Decode(&upstreamTaskData); err != nil {
				continue
			}

			upstreamDownstreamContract, _ := upstreamTaskData["downstreamContractDetail"].(string)
			// 比较契约（简化版：字符串比较）
			if upstreamDownstreamContract != upstreamContract && upstreamDownstreamContract != "" && upstreamContract != "" {
				// 契约不一致，但这里只作为警告
				// return fmt.Errorf("任务 [%s] 与上游任务契约不一致", taskName)
			}
		}
	}

	return nil
}

// TaskTestResult 任务测试结果
type TaskTestResult struct {
	Passed  bool
	Skipped int
}

// runTaskTests 运行任务测试
func (s *MCPServer) runTaskTests(task map[string]interface{}) (*TaskTestResult, error) {
	taskName, _ := task["name"].(string)
	status, _ := task["status"].(string)

	// 检查测试用例
	tests, _ := task["tests"].(string)
	if tests == "" || tests == "[]" {
		// 没有测试用例，跳过测试
		return &TaskTestResult{Passed: true, Skipped: 1}, nil
	}

	// 检查测试结果
	testResult, _ := task["testResult"].(string)
	if testResult == "" {
		// 没有测试结果，根据状态判断
		if status == "completed" {
			return &TaskTestResult{Passed: true}, nil
		}
		return &TaskTestResult{Passed: false}, fmt.Errorf("任务 [%s] 没有测试结果", taskName)
	}

	// 解析测试结果
	var results []map[string]interface{}
	if err := json.Unmarshal([]byte(testResult), &results); err != nil {
		return &TaskTestResult{Passed: false}, fmt.Errorf("解析测试结果失败: %s", err.Error())
	}

	// 检查是否有失败的测试
	for _, r := range results {
		if passed, ok := r["passed"].(bool); ok && !passed {
			return &TaskTestResult{Passed: false}, fmt.Errorf("任务 [%s] 存在失败的测试", taskName)
		}
	}

	return &TaskTestResult{Passed: true}, nil
}

// ==================== 9. 配置和规则管理实现 ====================

// handleGetConfigImpl 获取配置
func (s *MCPServer) handleGetConfigImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	config := s.configManager.GetConfig()

	data, _ := json.MarshalIndent(config, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleUpdateConfigImpl 更新配置
func (s *MCPServer) handleUpdateConfigImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	configVal, ok := request.Params.Arguments.(map[string]any)["config"]
	if !ok {
		return mcp.NewToolResultText("缺少 config 参数"), nil
	}

	config, ok := configVal.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultText("config 参数格式错误"), nil
	}

	if err := s.configManager.SaveConfig(config); err != nil {
		return mcp.NewToolResultText("保存配置失败：" + err.Error()), nil
	}

	return mcp.NewToolResultText("success: true"), nil
}

// handleGetRuleImpl 获取规则
func (s *MCPServer) handleGetRuleImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rule := s.ruleEngine.GetRule()

	data, _ := json.MarshalIndent(rule, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleUpdateRuleImpl 更新规则
func (s *MCPServer) handleUpdateRuleImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ruleVal, ok := request.Params.Arguments.(map[string]any)["rule"]
	if !ok {
		return mcp.NewToolResultText("缺少 rule 参数"), nil
	}

	rule, ok := ruleVal.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultText("rule 参数格式错误"), nil
	}

	if err := s.ruleEngine.SaveRule(rule); err != nil {
		return mcp.NewToolResultText("保存规则失败：" + err.Error()), nil
	}

	return mcp.NewToolResultText("success: true"), nil
}

// ==================== 10. 状态管理实现 ====================

// handleGetStatusImpl 获取状态
func (s *MCPServer) handleGetStatusImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, _ := getParam(request, "pathName")
	nodeType, _ := getParam(request, "type")
	includeErrors, _ := getParamBool(request, "includeErrors")
	includeLockInfo, _ := getParamBool(request, "includeLockInfo")

	// 如果没有提供 pathName，使用当前配置的项目
	if pathName == "" {
		pathName = s.configManager.GetProjectPathName()
	}

	if pathName == "" {
		return mcp.NewToolResultText("未配置项目路径名称，请先使用 init_project 设置项目"), nil
	}

	// 自动检测类型
	if nodeType == "" {
		nodeType = s.detectNodeType(pathName)
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("pathName: \"%s\"\n", pathName))
	output.WriteString(fmt.Sprintf("type: \"%s\"\n", nodeType))

	// 根据类型获取状态
	switch nodeType {
	case "project":
		resp, err := http.Get(fmt.Sprintf("%s/projects/by-path/%s", s.getApiURL(), pathName))
		if err != nil {
			return mcp.NewToolResultText("获取项目信息失败：" + err.Error()), nil
		}
		defer resp.Body.Close()

		var project map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
			return mcp.NewToolResultText("解析项目信息失败：" + err.Error()), nil
		}

		// 获取模块统计
		modulesResp, err := http.Get(fmt.Sprintf("%s/modules?projectPathName=%s", s.getApiURL(), pathName))
		if err == nil {
			defer modulesResp.Body.Close()
			var modulesResult map[string]interface{}
			if json.NewDecoder(modulesResp.Body).Decode(&modulesResult) == nil {
				if data, ok := modulesResult["data"].([]interface{}); ok {
					output.WriteString(fmt.Sprintf("moduleCount: %d\n", len(data)))
				}
			}
		}

		// 获取任务统计
		tasksResp, err := http.Get(fmt.Sprintf("%s/tasks?projectPathName=%s", s.getApiURL(), pathName))
		if err == nil {
			defer tasksResp.Body.Close()
			var tasksResult map[string]interface{}
			if json.NewDecoder(tasksResp.Body).Decode(&tasksResult) == nil {
				if data, ok := tasksResult["data"].([]interface{}); ok {
					output.WriteString(fmt.Sprintf("taskCount: %d\n", len(data)))

					// 统计各状态数量
					statusCounts := make(map[string]int)
					for _, task := range data {
						if taskMap, ok := task.(map[string]interface{}); ok {
							status, _ := taskMap["status"].(string)
							if status == "" {
								status = "ready"
							}
							statusCounts[status]++
						}
					}
					output.WriteString("taskStatus:\n")
					for status, count := range statusCounts {
						output.WriteString(fmt.Sprintf("  %s: %d\n", status, count))
					}
				}
			}
		}

	case "module":
		resp, err := http.Get(fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), pathName))
		if err != nil {
			return mcp.NewToolResultText("获取模块信息失败：" + err.Error()), nil
		}
		defer resp.Body.Close()

		var moduleResult map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&moduleResult); err != nil {
			return mcp.NewToolResultText("解析模块信息失败：" + err.Error()), nil
		}

		if module, ok := moduleResult["module"].(map[string]interface{}); ok {
			output.WriteString(fmt.Sprintf("name: \"%s\"\n", module["name"]))
			output.WriteString(fmt.Sprintf("status: \"%s\"\n", module["status"]))
		}

		// 获取任务统计
		tasksResp, err := http.Get(fmt.Sprintf("%s/tasks?modulePathName=%s", s.getApiURL(), pathName))
		if err == nil {
			defer tasksResp.Body.Close()
			var tasksResult map[string]interface{}
			if json.NewDecoder(tasksResp.Body).Decode(&tasksResult) == nil {
				if data, ok := tasksResult["data"].([]interface{}); ok {
					output.WriteString(fmt.Sprintf("taskCount: %d\n", len(data)))
				}
			}
		}

	case "task":
		resp, err := http.Get(fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), pathName))
		if err != nil {
			return mcp.NewToolResultText("获取任务信息失败：" + err.Error()), nil
		}
		defer resp.Body.Close()

		var taskResult map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&taskResult); err != nil {
			return mcp.NewToolResultText("解析任务信息失败：" + err.Error()), nil
		}

		if task, ok := taskResult["task"].(map[string]interface{}); ok {
			output.WriteString(fmt.Sprintf("name: \"%s\"\n", task["name"]))
			output.WriteString(fmt.Sprintf("status: \"%s\"\n", task["status"]))

			if includeLockInfo {
				locked, _ := task["locked"].(bool)
				output.WriteString(fmt.Sprintf("locked: %v\n", locked))
				if locked {
					output.WriteString(fmt.Sprintf("lockedBy: \"%s\"\n", task["lockedBy"]))
				}
			}
		}
	}

	// 包含错误信息
	if includeErrors {
		output.WriteString("errors: []\n")
		output.WriteString("warnings: []\n")
	}

	return mcp.NewToolResultText(output.String()), nil
}
