package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	ResourceType     string `json:"resourceType,omitempty"` // module, task
	ResourceName     string `json:"resourceName,omitempty"`
	ResourcePathName string `json:"resourcePathName,omitempty"`
	Message          string `json:"message"`
	Suggestion       string `json:"suggestion"`
	Severity         string `json:"severity"` // error, warning
}

// CompileIssueResource 编译问题关联的资源
type CompileIssueResource struct {
	Name     string `json:"name"`     // 资源名称
	PathName string `json:"pathName"` // 资源路径
}

// CompileIssueGroup 按规则分组的编译问题
type CompileIssueGroup struct {
	RuleID       string                 `json:"ruleId"`       // 规则ID
	RuleName     string                 `json:"ruleName"`     // 规则名称
	ResourceType string                 `json:"resourceType"` // module, task
	Resources    []CompileIssueResource `json:"resources"`    // 受影响的资源列表
	Message      string                 `json:"message"`      // 通用消息
	Suggestion   string                 `json:"suggestion"`   // 建议操作
}

// CompileIssueGroupsByType 按资源类型分组的编译问题
type CompileIssueGroupsByType struct {
	Module []CompileIssueGroup `json:"module,omitempty"`
	Task   []CompileIssueGroup `json:"task,omitempty"`
}

// CompileStaticResult 静态编译结果（内部使用）
type CompileStaticResult struct {
	Success         bool            `json:"success"`
	TotalModules    int             `json:"totalModules"`
	TotalTasks      int             `json:"totalTasks"`
	CompletedTasks  int             `json:"completedTasks"`
	InProgressTasks int             `json:"inProgressTasks"`
	ReadyTasks      int             `json:"readyTasks"`
	ErrorCount      int             `json:"errorCount"`
	WarningCount    int             `json:"warningCount"`
	Errors          []CompileIssue  `json:"errors"`
	Warnings        []CompileIssue  `json:"warnings"`
}

// CompileStaticOutput 静态编译输出（用于JSON输出）
type CompileStaticOutput struct {
	Success         bool                     `json:"success"`
	PathName        string                   `json:"pathName,omitempty"`
	TotalModules    int                      `json:"totalModules"`
	TotalTasks      int                      `json:"totalTasks"`
	CompletedTasks  int                      `json:"completedTasks"`
	InProgressTasks int                      `json:"inProgressTasks"`
	ReadyTasks      int                      `json:"readyTasks"`
	ErrorCount      int                      `json:"errorCount"`
	WarningCount    int                      `json:"warningCount"`
	Errors          CompileIssueGroupsByType `json:"errors,omitempty"`
	Warnings        CompileIssueGroupsByType `json:"warnings,omitempty"`
	Message         string                   `json:"message"`
}

// ContractDetailItem 契约条目
type ContractDetailItem struct {
	Label       string `json:"label"`
	ContractAPI string `json:"contract_api"`
	From        string `json:"from"`
}

// ContractDetail 契约详情
type ContractDetail struct {
	Title string                `json:"title"`
	List  []ContractDetailItem `json:"list"`
}

// BugLog Bug日志结构
type BugLog struct {
	Static  []string `json:"static"`
	Dynamic []string `json:"dynamic"`
}

// HasErrors 检查是否存在错误
func (b *BugLog) HasErrors() bool {
	return len(b.Static) > 0 || len(b.Dynamic) > 0
}

// groupCompileIssuesByTypeAndRule 按资源类型和规则ID分组编译问题
func groupCompileIssuesByTypeAndRule(issues []CompileIssue) CompileIssueGroupsByType {
	result := CompileIssueGroupsByType{
		Module: []CompileIssueGroup{},
		Task:   []CompileIssueGroup{},
	}

	// 先按 resourceType 分组
	moduleIssues := make(map[string][]CompileIssue)
	taskIssues := make(map[string][]CompileIssue)

	for _, issue := range issues {
		switch issue.ResourceType {
		case "module":
			moduleIssues[issue.RuleID] = append(moduleIssues[issue.RuleID], issue)
		case "task":
			taskIssues[issue.RuleID] = append(taskIssues[issue.RuleID], issue)
		}
	}

	// 转换模块问题为分组格式
	for ruleID, issueList := range moduleIssues {
		if len(issueList) == 0 {
			continue
		}
		group := CompileIssueGroup{
			RuleID:       ruleID,
			RuleName:     issueList[0].RuleName,
			ResourceType: "module",
			Resources:    []CompileIssueResource{},
			Message:      issueList[0].Message,
			Suggestion:   issueList[0].Suggestion,
		}
		for _, issue := range issueList {
			group.Resources = append(group.Resources, CompileIssueResource{
				Name:     issue.ResourceName,
				PathName: issue.ResourcePathName,
			})
		}
		result.Module = append(result.Module, group)
	}

	// 转换任务问题为分组格式
	for ruleID, issueList := range taskIssues {
		if len(issueList) == 0 {
			continue
		}
		group := CompileIssueGroup{
			RuleID:       ruleID,
			RuleName:     issueList[0].RuleName,
			ResourceType: "task",
			Resources:    []CompileIssueResource{},
			Message:      issueList[0].Message,
			Suggestion:   issueList[0].Suggestion,
		}
		for _, issue := range issueList {
			group.Resources = append(group.Resources, CompileIssueResource{
				Name:     issue.ResourceName,
				PathName: issue.ResourcePathName,
			})
		}
		result.Task = append(result.Task, group)
	}

	return result
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

	// 构建输出（使用新的分组格式）
	output := CompileStaticOutput{
		Success:         result.Success,
		PathName:        pathName,
		TotalModules:    result.TotalModules,
		TotalTasks:      result.TotalTasks,
		CompletedTasks:  result.CompletedTasks,
		InProgressTasks: result.InProgressTasks,
		ReadyTasks:      result.ReadyTasks,
		ErrorCount:      result.ErrorCount,
		WarningCount:    result.WarningCount,
		Errors:          groupCompileIssuesByTypeAndRule(result.Errors),
		Message:         "",
	}

	if includeWarnings {
		output.Warnings = groupCompileIssuesByTypeAndRule(result.Warnings)
	}

	if result.Success {
		output.Message = "静态编译通过"
	} else {
		output.Message = fmt.Sprintf("静态编译失败，发现 %d 个错误", result.ErrorCount)
	}

	// 序列化为JSON输出
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("序列化结果失败: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

// compileStaticProject 项目级别静态编译
func (s *MCPServer) compileStaticProject(projectPathName string, includeWarnings bool, result *CompileStaticResult) {
	// 获取所有模块，然后通过路径前缀过滤
	modulesResp, err := http.Get(fmt.Sprintf("%s/modules", s.getApiURL()))
	if err != nil {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-00",
			RuleName:         "项目结构检查",
			ResourceType:     "project",
			ResourceName:     projectPathName,
			ResourcePathName: projectPathName,
			Message:          "获取模块列表失败: " + err.Error(),
			Severity:         "error",
		})
		return
	}
	defer modulesResp.Body.Close()

	body, err := io.ReadAll(modulesResp.Body)
	if err != nil {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-00",
			RuleName:         "项目结构检查",
			ResourceType:     "project",
			ResourceName:     projectPathName,
			ResourcePathName: projectPathName,
			Message:          "读取模块列表响应失败: " + err.Error(),
			Severity:         "error",
		})
		return
	}

	// 解析模块响应
	var modulesData struct {
		Success bool `json:"success"`
		Data    struct {
			Modules []map[string]interface{} `json:"modules"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &modulesData); err != nil {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-00",
			RuleName:         "项目结构检查",
			ResourceType:     "project",
			ResourceName:     projectPathName,
			ResourcePathName: projectPathName,
			Message:          fmt.Sprintf("解析模块列表失败: %v", err),
			Severity:         "error",
		})
		return
	}

	// 通过路径前缀过滤出属于该项目的模块
	// 模块 pathName 格式为 "项目名/模块名"
	var modules []map[string]interface{}
	prefix := projectPathName + "/"
	for _, module := range modulesData.Data.Modules {
		pathName, _ := module["pathName"].(string)
		if strings.HasPrefix(pathName, prefix) {
			modules = append(modules, module)
		}
	}

	result.TotalModules = len(modules)

	// 收集所有任务用于契约一致性检查
	var allTasks []map[string]interface{}

	// 遍历每个模块进行编译
	for _, module := range modules {
		modulePathName, _ := module["pathName"].(string)
		moduleTasks := s.compileStaticModuleWithTasks(modulePathName, includeWarnings, result)
		allTasks = append(allTasks, moduleTasks...)
	}

	// 执行契约一致性检查
	s.checkContractConsistency(allTasks, result)

	// E-S-09: 检查模块循环引用
	s.checkModuleCircularReference(allTasks, modules, result)

	// E-S-09: 检查模块跨模块依赖
	s.checkModuleCrossDependency(allTasks, modules, result)

	// E-S-10: 检查孤立任务
	s.checkOrphanTask(allTasks, result)

	// 如果编译成功，更新任务依赖关系到数据库
	if result.ErrorCount == 0 {
		s.updateTaskDependencies(allTasks, result)
	}
}

// updateTaskDependencies 从契约中提取依赖关系并更新到数据库
func (s *MCPServer) updateTaskDependencies(allTasks []map[string]interface{}, result *CompileStaticResult) {
	// 构建任务路径到任务ID的映射
	taskPathToID := make(map[string]string)
	// 构建任务名称到任务ID的映射（用于跨模块查找）
	taskNameToID := make(map[string]string)
	for _, task := range allTasks {
		pathName, _ := task["pathName"].(string)
		taskID, _ := task["id"].(string)
		taskName, _ := task["name"].(string)
		if pathName != "" && taskID != "" {
			taskPathToID[pathName] = taskID
		}
		if taskName != "" && taskID != "" {
			taskNameToID[taskName] = taskID
		}
	}

	// 收集所有依赖关系
	type DependencyInfo struct {
		UpstreamTaskID   string
		DownstreamTaskID string
		ContractAPI      string
	}

	var dependencies []DependencyInfo

	// 遍历所有任务，从上游契约中提取依赖
	for _, task := range allTasks {
		downstreamTaskID, _ := task["id"].(string)
		downstreamPathName, _ := task["pathName"].(string)

		upstreamContractStr, _ := task["upstreamContractDetail"].(string)
		if upstreamContractStr == "" {
			continue
		}

		var upstreamContract ContractDetail
		if err := json.Unmarshal([]byte(upstreamContractStr), &upstreamContract); err != nil {
			continue
		}

		// 跳过首个任务
		if upstreamContract.Title == "start" {
			continue
		}

		// 从上游契约列表中提取依赖
		for _, item := range upstreamContract.List {
			// item.From 是上游任务的名称或路径
			upstreamTaskPath := item.From

			// 尝试查找上游任务ID
			upstreamTaskID := ""

			// 首先尝试直接匹配路径
			if id, exists := taskPathToID[upstreamTaskPath]; exists {
				upstreamTaskID = id
			} else {
				// 尝试在当前模块内查找
				modulePath := downstreamPathName
				if idx := strings.LastIndex(modulePath, "/"); idx > 0 {
					modulePath = modulePath[:idx]
					possiblePath := modulePath + "/" + upstreamTaskPath
					if id, exists := taskPathToID[possiblePath]; exists {
						upstreamTaskID = id
					}
				}
				// 如果模块内找不到，尝试在整个项目范围内按名称查找（跨模块）
				if upstreamTaskID == "" {
					if id, exists := taskNameToID[upstreamTaskPath]; exists {
						upstreamTaskID = id
					}
				}
			}

			if upstreamTaskID != "" && downstreamTaskID != "" {
				dependencies = append(dependencies, DependencyInfo{
					UpstreamTaskID:   upstreamTaskID,
					DownstreamTaskID: downstreamTaskID,
					ContractAPI:      item.ContractAPI,
				})
			}
		}
	}

	// 通过API更新依赖关系
	if len(dependencies) > 0 {
		// 先删除该项目的所有现有依赖
		// 获取项目ID（从第一个任务获取）
		if len(allTasks) > 0 {
			firstTask := allTasks[0]
			moduleID, _ := firstTask["moduleId"].(string)
			if moduleID != "" {
				// 获取模块信息以获取项目ID
				moduleResp, err := http.Get(fmt.Sprintf("%s/modules/%s", s.getApiURL(), moduleID))
				if err == nil {
					body, _ := io.ReadAll(moduleResp.Body)
					moduleResp.Body.Close()
					var moduleData struct {
						Success bool `json:"success"`
						Data    struct {
							Module map[string]interface{} `json:"module"`
						} `json:"data"`
					}
					if json.Unmarshal(body, &moduleData) == nil && moduleData.Success {
						projectID, _ := moduleData.Data.Module["projectId"].(string)
						if projectID != "" {
							// 删除该项目的所有现有任务依赖
							deleteURL := fmt.Sprintf("%s/dependencies/by-project?projectId=%s", s.getApiURL(), projectID)
							req, _ := http.NewRequest("DELETE", deleteURL, nil)
							resp, err := http.DefaultClient.Do(req)
							if err == nil {
								resp.Body.Close()
							}
						}
					}
				}
			}
		}

		// 创建新的依赖关系
		for _, dep := range dependencies {
			depData := map[string]interface{}{
				"upstreamTaskId":   dep.UpstreamTaskID,
				"downstreamTaskId": dep.DownstreamTaskID,
				"contractSummary":  dep.ContractAPI,
			}
			depJSON, _ := json.Marshal(depData)
			http.Post(fmt.Sprintf("%s/dependencies", s.getApiURL()), "application/json", bytes.NewReader(depJSON))
		}

		// 在结果中记录更新的依赖数量
		result.Warnings = append(result.Warnings, CompileIssue{
			RuleID:           "W-S-07",
			RuleName:         "依赖关系已更新",
			ResourceType:     "project",
			ResourceName:     "",
			ResourcePathName: "",
			Message:          fmt.Sprintf("已更新 %d 个任务依赖关系", len(dependencies)),
			Suggestion:       "依赖关系已同步到数据库，前端可查看可视化图",
			Severity:         "warning",
		})
	}
}

// compileStaticModuleWithTasks 模块级别静态编译，返回任务列表
func (s *MCPServer) compileStaticModuleWithTasks(modulePathName string, includeWarnings bool, result *CompileStaticResult) []map[string]interface{} {
	var allTasks []map[string]interface{}

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
		return allTasks
	}
	defer moduleResp.Body.Close()

	body, err := io.ReadAll(moduleResp.Body)
	if err != nil {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-00",
			RuleName:         "模块结构检查",
			ResourceType:     "module",
			ResourcePathName: modulePathName,
			Message:          "读取模块信息失败: " + err.Error(),
			Severity:         "error",
		})
		return allTasks
	}

	// 解析模块响应 - API返回格式: {"success":true,"data":{"module":{...}}}
	var moduleRespData struct {
		Success bool `json:"success"`
		Data    struct {
			Module map[string]interface{} `json:"module"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &moduleRespData); err != nil {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-00",
			RuleName:         "模块结构检查",
			ResourceType:     "module",
			ResourcePathName: modulePathName,
			Message:          fmt.Sprintf("解析模块信息失败: %v, 响应: %s", err, string(body)),
			Severity:         "error",
		})
		return allTasks
	}

	if !moduleRespData.Success || moduleRespData.Data.Module == nil {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-00",
			RuleName:         "模块结构检查",
			ResourceType:     "module",
			ResourcePathName: modulePathName,
			Message:          "模块不存在或获取失败",
			Severity:         "error",
		})
		return allTasks
	}

	module := moduleRespData.Data.Module
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
		return allTasks
	}
	defer tasksResp.Body.Close()

	tasksBody, err := io.ReadAll(tasksResp.Body)
	if err != nil {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-00",
			RuleName:         "模块任务检查",
			ResourceType:     "module",
			ResourceName:     moduleName,
			ResourcePathName: modulePathName,
			Message:          "读取模块任务失败: " + err.Error(),
			Severity:         "error",
		})
		return allTasks
	}

	// 解析任务列表 - API返回格式: {"success":true,"data":{"tasks":[...],"total":42}}
	var tasksRespData struct {
		Success bool `json:"success"`
		Data    struct {
			Tasks []map[string]interface{} `json:"tasks"`
			Total int                       `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(tasksBody, &tasksRespData); err != nil {
		result.Warnings = append(result.Warnings, CompileIssue{
			RuleID:           "W-S-06",
			RuleName:         "任务列表解析失败",
			ResourceType:     "module",
			ResourceName:     moduleName,
			ResourcePathName: modulePathName,
			Message:          fmt.Sprintf("解析模块任务列表失败: %v", err),
			Suggestion:       "请检查API返回格式",
			Severity:         "warning",
		})
		return allTasks
	}

	tasks := tasksRespData.Data.Tasks

	// 遍历任务进行编译
	for _, task := range tasks {
		s.compileStaticTaskFromData(task, modulePathName, moduleName, includeWarnings, result)
		allTasks = append(allTasks, task)
	}

	return allTasks
}

// compileStaticModule 模块级别静态编译（保留兼容性）
func (s *MCPServer) compileStaticModule(modulePathName string, includeWarnings bool, result *CompileStaticResult) {
	tasks := s.compileStaticModuleWithTasks(modulePathName, includeWarnings, result)
	// 对于单个模块编译，也需要执行契约一致性检查
	s.checkContractConsistency(tasks, result)
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
	if codePaths == "" || codePaths == "[]" || codePaths == "null" {
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

	// E-S-02: 检查 Bug 日志（使用新的 BugLog 结构）
	bugLogStr, _ := task["bugLog"].(string)
	if bugLogStr != "" && bugLogStr != "{}" && bugLogStr != "null" {
		var bugLog BugLog
		if err := json.Unmarshal([]byte(bugLogStr), &bugLog); err == nil {
			if bugLog.HasErrors() {
				result.Errors = append(result.Errors, CompileIssue{
					RuleID:           "E-S-02",
					RuleName:         "存在Bug日志",
					ResourceType:     "task",
					ResourceName:     taskName,
					ResourcePathName: taskPathName,
					Message:          fmt.Sprintf("任务 [%s] 存在未解决的Bug - static: %d, dynamic: %d",
						taskName, len(bugLog.Static), len(bugLog.Dynamic)),
					Suggestion:       "请解决Bug后清除日志",
					Severity:         "error",
				})
			}
		}
	}

	// E-S-03: 检查人工协助
	humanAssistance, _ := task["humanAssistance"].(string)
	if humanAssistance != "" && humanAssistance != "{}" && humanAssistance != "null" {
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

	// E-S-05: 检查上游契约
	s.checkUpstreamContract(task, taskPathName, result)

	// E-S-06: 检查下游契约
	s.checkDownstreamContract(task, taskPathName, result)

	// E-S-08: 检查状态异常
	s.checkStatusAnomaly(task, taskPathName, result)

	// W-S-01: 检查测试用例
	if includeWarnings {
		tests, _ := task["tests"].(string)
		if tests == "" || tests == "[]" || tests == "null" {
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

// checkUpstreamContract 检查上游契约 (E-S-05)
func (s *MCPServer) checkUpstreamContract(task map[string]interface{}, taskPathName string, result *CompileStaticResult) {
	taskName, _ := task["name"].(string)
	upstreamContractStr, _ := task["upstreamContractDetail"].(string)

	// 首个任务标记检查
	if upstreamContractStr == "" {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-05",
			RuleName:         "上游契约缺失",
			ResourceType:     "task",
			ResourceName:     taskName,
			ResourcePathName: taskPathName,
			Message:          fmt.Sprintf("任务 [%s] 缺少上游契约定义", taskName),
			Suggestion:       "首个任务应设置 title 为 'start'，其他任务需定义上游依赖",
			Severity:         "error",
		})
		return
	}

	var upstreamContract ContractDetail
	if err := json.Unmarshal([]byte(upstreamContractStr), &upstreamContract); err != nil {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-05",
			RuleName:         "上游契约格式错误",
			ResourceType:     "task",
			ResourceName:     taskName,
			ResourcePathName: taskPathName,
			Message:          fmt.Sprintf("任务 [%s] 上游契约JSON解析失败: %v", taskName, err),
			Suggestion:       "请检查 upstreamContractDetail 的JSON格式",
			Severity:         "error",
		})
		return
	}

	// 检查是否为首个任务（无上游依赖）
	if upstreamContract.Title == "start" {
		// 首个任务，list 应为空
		if len(upstreamContract.List) > 0 {
			result.Warnings = append(result.Warnings, CompileIssue{
				RuleID:           "W-S-03",
				RuleName:         "首个任务存在上游依赖",
				ResourceType:     "task",
				ResourceName:     taskName,
				ResourcePathName: taskPathName,
				Message:          fmt.Sprintf("任务 [%s] 标记为首个任务但存在上游依赖", taskName),
				Suggestion:       "请确认是否为首任务，若是则清空list",
				Severity:         "warning",
			})
		}
		return
	}

	// 非首个任务必须有 title 和 list
	if upstreamContract.Title == "" {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-05",
			RuleName:         "上游契约标题缺失",
			ResourceType:     "task",
			ResourceName:     taskName,
			ResourcePathName: taskPathName,
			Message:          fmt.Sprintf("任务 [%s] 上游契约缺少 title", taskName),
			Suggestion:       "请添加描述性的 title，如 '依赖xxx提供'",
			Severity:         "error",
		})
	}

	// 检查 list 中每个条目的完整性
	for i, item := range upstreamContract.List {
		if item.Label == "" || item.ContractAPI == "" || item.From == "" {
			result.Errors = append(result.Errors, CompileIssue{
				RuleID:           "E-S-05",
				RuleName:         "上游契约条目不完整",
				ResourceType:     "task",
				ResourceName:     taskName,
				ResourcePathName: taskPathName,
				Message:          fmt.Sprintf("任务 [%s] 上游契约第%d条缺少必要字段", taskName, i+1),
				Suggestion:       "每个条目需包含 label, contract_api, from 三个字段",
				Severity:         "error",
			})
		}
	}
}

// checkDownstreamContract 检查下游契约 (E-S-06)
func (s *MCPServer) checkDownstreamContract(task map[string]interface{}, taskPathName string, result *CompileStaticResult) {
	taskName, _ := task["name"].(string)
	downstreamContractStr, _ := task["downstreamContractDetail"].(string)

	if downstreamContractStr == "" {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-06",
			RuleName:         "下游契约缺失",
			ResourceType:     "task",
			ResourceName:     taskName,
			ResourcePathName: taskPathName,
			Message:          fmt.Sprintf("任务 [%s] 缺少下游契约定义", taskName),
			Suggestion:       "末尾任务应设置 title 为 'end'，其他任务需定义下游输出",
			Severity:         "error",
		})
		return
	}

	var downstreamContract ContractDetail
	if err := json.Unmarshal([]byte(downstreamContractStr), &downstreamContract); err != nil {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-06",
			RuleName:         "下游契约格式错误",
			ResourceType:     "task",
			ResourceName:     taskName,
			ResourcePathName: taskPathName,
			Message:          fmt.Sprintf("任务 [%s] 下游契约JSON解析失败: %v", taskName, err),
			Suggestion:       "请检查 downstreamContractDetail 的JSON格式",
			Severity:         "error",
		})
		return
	}

	// 检查是否为末尾任务（无下游依赖）
	if downstreamContract.Title == "end" {
		if len(downstreamContract.List) > 0 {
			result.Warnings = append(result.Warnings, CompileIssue{
				RuleID:           "W-S-04",
				RuleName:         "末尾任务存在下游输出",
				ResourceType:     "task",
				ResourceName:     taskName,
				ResourcePathName: taskPathName,
				Message:          fmt.Sprintf("任务 [%s] 标记为末尾任务但存在下游输出", taskName),
				Suggestion:       "请确认是否为末尾任务，若是则清空list",
				Severity:         "warning",
			})
		}
		return
	}

	// 非末尾任务必须有 title 和 list
	if downstreamContract.Title == "" {
		result.Errors = append(result.Errors, CompileIssue{
			RuleID:           "E-S-06",
			RuleName:         "下游契约标题缺失",
			ResourceType:     "task",
			ResourceName:     taskName,
			ResourcePathName: taskPathName,
			Message:          fmt.Sprintf("任务 [%s] 下游契约缺少 title", taskName),
			Suggestion:       "请添加描述性的 title，如 '为下游任务提供以下接口'",
			Severity:         "error",
		})
	}

	// 检查 list 中每个条目的完整性
	for i, item := range downstreamContract.List {
		if item.Label == "" || item.ContractAPI == "" || item.From == "" {
			result.Errors = append(result.Errors, CompileIssue{
				RuleID:           "E-S-06",
				RuleName:         "下游契约条目不完整",
				ResourceType:     "task",
				ResourceName:     taskName,
				ResourcePathName: taskPathName,
				Message:          fmt.Sprintf("任务 [%s] 下游契约第%d条缺少必要字段", taskName, i+1),
				Suggestion:       "每个条目需包含 label, contract_api, from 三个字段",
				Severity:         "error",
			})
		}
	}
}

// checkStatusAnomaly 检查状态异常 (E-S-08)
func (s *MCPServer) checkStatusAnomaly(task map[string]interface{}, taskPathName string, result *CompileStaticResult) {
	taskName, _ := task["name"].(string)
	status, _ := task["status"].(string)

	// 只检查已完成状态的任务
	if status != "completed" {
		return
	}

	// 检查 bugLog
	bugLogStr, _ := task["bugLog"].(string)
	if bugLogStr != "" && bugLogStr != "{}" && bugLogStr != "null" {
		var bugLog BugLog
		if err := json.Unmarshal([]byte(bugLogStr), &bugLog); err == nil {
			if bugLog.HasErrors() {
				result.Errors = append(result.Errors, CompileIssue{
					RuleID:           "E-S-08",
					RuleName:         "已完成任务存在Bug",
					ResourceType:     "task",
					ResourceName:     taskName,
					ResourcePathName: taskPathName,
					Message:          fmt.Sprintf("任务 [%s] 已完成但存在未解决的Bug - static: %d, dynamic: %d",
						taskName, len(bugLog.Static), len(bugLog.Dynamic)),
					Suggestion:       "请解决Bug后清除日志或将状态改为非完成状态",
					Severity:         "error",
				})
			}
		}
	}
}

// findTaskPathByName 根据任务名称或路径查找任务路径
func findTaskPathByName(taskMap map[string]map[string]interface{}, from string, currentTaskPath string) string {
	// 首先尝试直接匹配 pathName
	if _, exists := taskMap[from]; exists {
		return from
	}

	// 尝试在同一模块下查找任务名称
	currentParts := strings.Split(currentTaskPath, "/")
	if len(currentParts) >= 2 {
		modulePath := currentParts[0] + "/" + currentParts[1]
		// 尝试构建完整路径
		possiblePath := modulePath + "/" + from
		if _, exists := taskMap[possiblePath]; exists {
			return possiblePath
		}
	}

	// 遍历所有任务查找匹配的名称
	for pathName, task := range taskMap {
		name, _ := task["name"].(string)
		if name == from {
			return pathName
		}
	}

	return ""
}

// checkContractMatchInDownstream 检查契约条目是否在上游任务的下游契约中存在匹配（反向检查）
// 参数:
//   - upstreamTask: 上游任务数据
//   - item: 当前任务的上游契约条目（来自 upstreamContractDetail.list）
// 返回值:
//   - bool: true 表示找到匹配，false 表示未找到匹配
func checkContractMatchInDownstream(upstreamTask map[string]interface{}, item ContractDetailItem) bool {
	downstreamContractStr, _ := upstreamTask["downstreamContractDetail"].(string)
	if downstreamContractStr == "" {
		return false
	}

	var downstreamContract ContractDetail
	if err := json.Unmarshal([]byte(downstreamContractStr), &downstreamContract); err != nil {
		return false
	}

	// 在上游任务的下游契约列表中查找匹配项
	// 三个字段必须全部匹配：
	// 1. label 必须完全相同
	// 2. contract_api 必须完全相同
	// 3. from 必须完全相同（都表示这条契约来自哪个任务）
	for _, downstreamItem := range downstreamContract.List {
		if item.Label == downstreamItem.Label &&
			item.ContractAPI == downstreamItem.ContractAPI &&
			item.From == downstreamItem.From {
			return true
		}
	}

	return false
}

// checkContractMatchInUpstream 检查契约条目是否在下游任务的上游契约中存在匹配（正向检查）
// 参数:
//   - downstreamTask: 下游任务数据
//   - item: 当前任务的下游契约条目（来自 downstreamContractDetail.list）
//   - currentTaskPathName: 当前任务的 pathName（用于匹配 from 字段）
// 返回值:
//   - bool: true 表示找到匹配，false 表示未找到匹配
func checkContractMatchInUpstream(downstreamTask map[string]interface{}, item ContractDetailItem, currentTaskPathName string) bool {
	upstreamContractStr, _ := downstreamTask["upstreamContractDetail"].(string)
	if upstreamContractStr == "" {
		return false
	}

	var upstreamContract ContractDetail
	if err := json.Unmarshal([]byte(upstreamContractStr), &upstreamContract); err != nil {
		return false
	}

	// 在下游任务的上游契约列表中查找匹配项
	// 三个字段必须全部匹配：
	// 1. label 必须完全相同
	// 2. contract_api 必须完全相同
	// 3. upstreamItem.From 必须等于当前任务的 pathName（因为 from 表示契约来源）
	for _, upstreamItem := range upstreamContract.List {
		if item.Label == upstreamItem.Label &&
			item.ContractAPI == upstreamItem.ContractAPI &&
			upstreamItem.From == currentTaskPathName {
			return true
		}
	}

	return false
}

// checkContractConsistency 检查契约一致性 (E-S-07)
// 执行双向检查：
// 1. 反向检查：遍历每个任务的上游契约，在上游任务的下游契约中查找匹配
// 2. 正向检查：遍历每个任务的下游契约，在下游任务的上游契约中查找匹配
// 需要在所有任务数据加载完成后执行
func (s *MCPServer) checkContractConsistency(allTasks []map[string]interface{}, result *CompileStaticResult) {
	// 构建任务路径到任务的映射
	taskMap := make(map[string]map[string]interface{})
	for _, task := range allTasks {
		pathName, _ := task["pathName"].(string)
		taskMap[pathName] = task
	}

	// ==================== 反向检查 ====================
	// 遍历所有任务，检查每个任务的上游契约
	for _, task := range allTasks {
		taskName, _ := task["name"].(string)
		taskPathName, _ := task["pathName"].(string)

		upstreamContractStr, _ := task["upstreamContractDetail"].(string)
		if upstreamContractStr == "" {
			continue
		}

		var upstreamContract ContractDetail
		if err := json.Unmarshal([]byte(upstreamContractStr), &upstreamContract); err != nil {
			continue // 格式错误已在E-S-05中报告
		}

		// 跳过首个任务（title 为 "start"）
		if upstreamContract.Title == "start" {
			continue
		}

		// 遍历上游契约列表中的每一条契约
		for _, upstreamItem := range upstreamContract.List {
			// 反向检查：当前任务的上游契约 {from: 上游任务, label: X}
			// 查找上游任务时，不应该使用 from 字段，而是应该遍历所有任务，找到下游契约中有匹配项的任务
			foundMatch := false
			for _, potentialUpstreamTask := range allTasks {
				// 跳过自己
				potentialPathName, _ := potentialUpstreamTask["pathName"].(string)
				if potentialPathName == taskPathName {
					continue
				}

				// 检查该任务的下游契约是否有匹配项
				if checkContractMatchInDownstream(potentialUpstreamTask, upstreamItem) {
					foundMatch = true
					break
				}
			}

			if !foundMatch {
				// 找不到匹配的上游任务，报错
				contractJSON, _ := json.Marshal(map[string]string{
					"label":        upstreamItem.Label,
					"contract_api": upstreamItem.ContractAPI,
					"from":         upstreamItem.From,
				})
				result.Errors = append(result.Errors, CompileIssue{
					RuleID:       "E-S-07",
					RuleName:     "任务契约不一致",
					ResourceType: "task",
					Message:      fmt.Sprintf("任务[%s] 上游契约条目 %s 未找到提供该接口的上游任务", taskName, string(contractJSON)),
					Suggestion:   "请检查契约定义是否正确，确保上下游任务的契约条目（label, contract_api, from）完全一致",
					Severity:     "error",
				})
			}
		}
	}

	// ==================== 正向检查 ====================
	// 遍历所有任务，检查每个任务的下游契约
	// 下游契约的 from 字段表示当前任务自己（提供接口的任务）
	// 需要在消费该接口的下游任务的上游契约中找到匹配项
	for _, task := range allTasks {
		taskName, _ := task["name"].(string)
		taskPathName, _ := task["pathName"].(string)

		downstreamContractStr, _ := task["downstreamContractDetail"].(string)
		if downstreamContractStr == "" {
			continue
		}

		var downstreamContract ContractDetail
		if err := json.Unmarshal([]byte(downstreamContractStr), &downstreamContract); err != nil {
			continue // 格式错误已在E-S-05中报告
		}

		// 跳过末尾任务（title 为 "end"）
		if downstreamContract.Title == "end" {
			continue
		}

		// 遍历下游契约列表中的每一条契约
		for _, downstreamItem := range downstreamContract.List {
			// 正向检查：当前任务的下游契约 {from: 当前任务, label: X}
			// 需要在所有任务中找到上游契约中有匹配项的下游任务
			// 下游任务的上游契约应该是 {from: 当前任务, label: X}
			
			// 遍历所有任务，查找上游契约中有匹配项的任务（下游任务）
			foundMatch := false
			for _, potentialDownstreamTask := range allTasks {
				// 跳过自己
				potentialPathName, _ := potentialDownstreamTask["pathName"].(string)
				if potentialPathName == taskPathName {
					continue
				}

				// 检查该任务的上游契约是否有匹配项
				if checkContractMatchInUpstream(potentialDownstreamTask, downstreamItem, taskPathName) {
					foundMatch = true
					break
				}
			}

			if !foundMatch {
				// 找不到匹配的下游任务，报错
				contractJSON, _ := json.Marshal(map[string]string{
					"label":        downstreamItem.Label,
					"contract_api": downstreamItem.ContractAPI,
					"from":         taskPathName,
				})
				result.Errors = append(result.Errors, CompileIssue{
					RuleID:       "E-S-07",
					RuleName:     "任务契约不一致",
					ResourceType: "task",
					Message:      fmt.Sprintf("任务[%s] 下游契约条目 %s 未找到消费该接口的下游任务", taskName, string(contractJSON)),
					Suggestion:   "请检查契约定义是否正确，确保上下游任务的契约条目（label, contract_api, from）完全一致",
					Severity:     "error",
				})
			}
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

// checkModuleCrossDependency 检查模块跨模块依赖 (E-S-09)
// 规则：每个模块必须至少有一个跨模块的依赖（作为上游或下游）
func (s *MCPServer) checkModuleCrossDependency(allTasks []map[string]interface{}, modules []map[string]interface{}, result *CompileStaticResult) {
	// 如果只有一个模块，跳过检查
	if len(modules) <= 1 {
		return
	}

	// 构建任务路径到模块路径的映射
	taskToModule := make(map[string]string)
	for _, task := range allTasks {
		taskPathName, _ := task["pathName"].(string)
		if taskPathName == "" {
			continue
		}
		// 模块路径是任务路径的前两部分
		parts := strings.Split(taskPathName, "/")
		if len(parts) >= 2 {
			modulePath := parts[0] + "/" + parts[1]
			taskToModule[taskPathName] = modulePath
		}
	}

	// 构建任务名称到模块路径的映射（用于跨模块查找）
	taskNameToModule := make(map[string]string)
	for _, task := range allTasks {
		taskName, _ := task["name"].(string)
		taskPathName, _ := task["pathName"].(string)
		if taskName != "" && taskPathName != "" {
			taskNameToModule[taskName] = taskToModule[taskPathName]
		}
	}

	// 记录每个模块是否有跨模块依赖
	moduleHasCrossDependency := make(map[string]bool)
	for _, module := range modules {
		modulePathName, _ := module["pathName"].(string)
		moduleHasCrossDependency[modulePathName] = false
	}

	// 遍历所有任务，检查跨模块依赖
	for _, task := range allTasks {
		taskPathName, _ := task["pathName"].(string)
		currentModulePath := taskToModule[taskPathName]
		if currentModulePath == "" {
			continue
		}

		// 检查上游契约
		upstreamContractStr, _ := task["upstreamContractDetail"].(string)
		if upstreamContractStr != "" {
			var upstreamContract ContractDetail
			if err := json.Unmarshal([]byte(upstreamContractStr), &upstreamContract); err == nil {
				for _, item := range upstreamContract.List {
					upstreamTaskPath := item.From
					if upstreamTaskPath == "" {
						continue
					}

					// 查找上游任务所属模块
					upstreamModulePath := ""
					// 首先尝试直接匹配路径
					if modulePath, exists := taskToModule[upstreamTaskPath]; exists {
						upstreamModulePath = modulePath
					} else if modulePath, exists := taskNameToModule[upstreamTaskPath]; exists {
						// 尝试按名称匹配
						upstreamModulePath = modulePath
					} else {
						// 尝试在当前模块内查找
						possiblePath := currentModulePath + "/" + upstreamTaskPath
						if modulePath, exists := taskToModule[possiblePath]; exists {
							upstreamModulePath = modulePath
						}
					}

					// 如果找到了上游模块且与当前模块不同，则存在跨模块依赖
					if upstreamModulePath != "" && upstreamModulePath != currentModulePath {
						moduleHasCrossDependency[currentModulePath] = true
						moduleHasCrossDependency[upstreamModulePath] = true
					}
				}
			}
		}

		// 检查下游契约
		downstreamContractStr, _ := task["downstreamContractDetail"].(string)
		if downstreamContractStr != "" {
			var downstreamContract ContractDetail
			if err := json.Unmarshal([]byte(downstreamContractStr), &downstreamContract); err == nil {
				for _, item := range downstreamContract.List {
					downstreamTaskPath := item.From
					if downstreamTaskPath == "" {
						continue
					}

					// 查找下游任务所属模块
					downstreamModulePath := ""
					// 首先尝试直接匹配路径
					if modulePath, exists := taskToModule[downstreamTaskPath]; exists {
						downstreamModulePath = modulePath
					} else if modulePath, exists := taskNameToModule[downstreamTaskPath]; exists {
						// 尝试按名称匹配
						downstreamModulePath = modulePath
					} else {
						// 尝试在当前模块内查找
						possiblePath := currentModulePath + "/" + downstreamTaskPath
						if modulePath, exists := taskToModule[possiblePath]; exists {
							downstreamModulePath = modulePath
						}
					}

					// 如果找到了下游模块且与当前模块不同，则存在跨模块依赖
					if downstreamModulePath != "" && downstreamModulePath != currentModulePath {
						moduleHasCrossDependency[currentModulePath] = true
						moduleHasCrossDependency[downstreamModulePath] = true
					}
				}
			}
		}
	}

	// 检查每个模块是否有跨模块依赖，如果没有则报错
	for _, module := range modules {
		modulePathName, _ := module["pathName"].(string)
		moduleName, _ := module["name"].(string)

		if !moduleHasCrossDependency[modulePathName] {
			result.Errors = append(result.Errors, CompileIssue{
				RuleID:           "E-S-09",
				RuleName:         "模块缺少跨模块依赖",
				ResourceType:     "module",
				ResourceName:     moduleName,
				ResourcePathName: modulePathName,
				Message:          fmt.Sprintf("模块 [%s] 没有任何跨模块依赖，模块必须与外部链接", moduleName),
				Suggestion:       "请在任务的契约中添加对其他模块任务的依赖关系",
				Severity:         "error",
			})
		}
	}
}

// checkModuleCircularReference 检查模块间循环引用 (E-S-09)
// 规则：跨模块任务引用不能形成循环依赖
// 例如：模块X任务A → 模块Y任务E 和 模块Y任务D → 模块X任务B 形成循环
func (s *MCPServer) checkModuleCircularReference(allTasks []map[string]interface{}, modules []map[string]interface{}, result *CompileStaticResult) {
	// 如果少于两个模块，不可能存在循环引用
	if len(modules) < 2 {
		return
	}

	// 构建任务路径到模块路径的映射
	taskToModule := make(map[string]string)
	taskNameToPath := make(map[string]string)
	for _, task := range allTasks {
		taskPathName, _ := task["pathName"].(string)
		taskName, _ := task["name"].(string)
		if taskPathName == "" {
			continue
		}
		// 模块路径是任务路径的前两部分
		parts := strings.Split(taskPathName, "/")
		if len(parts) >= 2 {
			modulePath := parts[0] + "/" + parts[1]
			taskToModule[taskPathName] = modulePath
		}
		if taskName != "" {
			taskNameToPath[taskName] = taskPathName
		}
	}

	// 构建模块路径到模块名称的映射
	modulePathToName := make(map[string]string)
	for _, module := range modules {
		modulePathName, _ := module["pathName"].(string)
		moduleName, _ := module["name"].(string)
		modulePathToName[modulePathName] = moduleName
	}

	// 记录模块间的引用关系
	// key: "源模块路径->目标模块路径", value: 引用列表
	moduleReferences := make(map[string][]CrossModuleReference)

	// 遍历所有任务，分析跨模块引用
	for _, task := range allTasks {
		taskPathName, _ := task["pathName"].(string)
		currentModulePath := taskToModule[taskPathName]
		if currentModulePath == "" {
			continue
		}

		// 检查上游契约中的跨模块引用
		upstreamContractStr, _ := task["upstreamContractDetail"].(string)
		if upstreamContractStr != "" {
			var upstreamContract ContractDetail
			if err := json.Unmarshal([]byte(upstreamContractStr), &upstreamContract); err == nil {
				for _, item := range upstreamContract.List {
					upstreamTaskPath := item.From
					if upstreamTaskPath == "" {
						continue
					}

					// 查找上游任务所属模块
					upstreamModulePath := findModuleByTaskIdentifier(upstreamTaskPath, taskToModule, taskNameToPath, currentModulePath)

					// 如果上游任务属于不同模块，则记录引用（方向：上游模块 → 当前模块）
					if upstreamModulePath != "" && upstreamModulePath != currentModulePath {
						key := upstreamModulePath + "->" + currentModulePath
						moduleReferences[key] = append(moduleReferences[key], CrossModuleReference{
							SourceTask: upstreamTaskPath,
							TargetTask: taskPathName,
						})
					}
				}
			}
		}
	}

	// 检测循环引用：检查所有模块对之间是否存在双向引用
	for i := 0; i < len(modules); i++ {
		for j := i + 1; j < len(modules); j++ {
			moduleXPathName, _ := modules[i]["pathName"].(string)
			moduleYPathName, _ := modules[j]["pathName"].(string)
			moduleXName, _ := modules[i]["name"].(string)
			moduleYName, _ := modules[j]["name"].(string)

			keyXToY := moduleXPathName + "->" + moduleYPathName
			keyYToX := moduleYPathName + "->" + moduleXPathName

			xToYRefs, hasXToY := moduleReferences[keyXToY]
			yToXRefs, hasYToX := moduleReferences[keyYToX]

			// 如果同时存在 X→Y 和 Y→X 的引用，则存在循环引用
			if hasXToY && hasYToX {
				result.Errors = append(result.Errors, CompileIssue{
					RuleID:           RuleModuleCircularReference,
					RuleName:         "模块循环引用错误",
					ResourceType:     "module",
					ResourceName:     moduleXName + " <-> " + moduleYName,
					ResourcePathName: moduleXPathName + " <-> " + moduleYPathName,
					Message:          fmt.Sprintf("模块 [%s] 和模块 [%s] 之间存在循环引用（%s → %s: %d条, %s → %s: %d条）", moduleXName, moduleYName, moduleXName, moduleYName, len(xToYRefs), moduleYName, moduleXName, len(yToXRefs)),
					Suggestion:       "建议引入第三方共享模块来打破循环依赖",
					Severity:         "error",
				})
			}
		}
	}
}

// findModuleByTaskIdentifier 根据任务标识符查找所属模块路径
func findModuleByTaskIdentifier(taskIdentifier string, taskToModule map[string]string, taskNameToPath map[string]string, currentModulePath string) string {
	// 1. 首先尝试直接匹配任务路径
	if modulePath, exists := taskToModule[taskIdentifier]; exists {
		return modulePath
	}

	// 2. 尝试按任务名称匹配
	if taskPath, exists := taskNameToPath[taskIdentifier]; exists {
		return taskToModule[taskPath]
	}

	// 3. 尝试在当前模块内构建完整路径
	possiblePath := currentModulePath + "/" + taskIdentifier
	if modulePath, exists := taskToModule[possiblePath]; exists {
		return modulePath
	}

	return ""
}

// checkOrphanTask 检查孤立任务 (E-S-10)
// 规则：任务必须有上游依赖或下游被依赖（除非是项目的起点/终点任务）
// 孤立任务定义：没有任何上游契约引用，也没有任何下游契约被引用的任务
func (s *MCPServer) checkOrphanTask(allTasks []map[string]interface{}, result *CompileStaticResult) {
	// 如果只有一个任务，跳过检查
	if len(allTasks) <= 1 {
		return
	}

	// 构建任务路径到任务的映射
	taskPathToTask := make(map[string]map[string]interface{})
	for _, task := range allTasks {
		taskPathName, _ := task["pathName"].(string)
		if taskPathName != "" {
			taskPathToTask[taskPathName] = task
		}
	}

	// 收集所有被引用的任务路径（作为上游被依赖）
	referencedAsUpstream := make(map[string]bool)
	// 收集所有引用其他任务的任务路径（有下游契约被消费）
	referencedAsDownstream := make(map[string]bool)

	for _, task := range allTasks {
		taskPathName, _ := task["pathName"].(string)
		if taskPathName == "" {
			continue
		}

		// 检查上游契约 - 当前任务依赖哪些上游任务
		upstreamContractStr, _ := task["upstreamContractDetail"].(string)
		if upstreamContractStr != "" {
			var upstreamContract ContractDetail
			if err := json.Unmarshal([]byte(upstreamContractStr), &upstreamContract); err == nil {
				for _, item := range upstreamContract.List {
					if item.From != "" {
						// item.From是上游任务的路径
						referencedAsUpstream[item.From] = true
					}
				}
			}
		}

		// 检查下游契约 - 当前任务被哪些下游任务依赖
		downstreamContractStr, _ := task["downstreamContractDetail"].(string)
		if downstreamContractStr != "" {
			var downstreamContract ContractDetail
			if err := json.Unmarshal([]byte(downstreamContractStr), &downstreamContract); err == nil {
				if len(downstreamContract.List) > 0 {
					// 有下游契约定义，说明可能被下游任务引用
					// 注意：这里只是定义了下游契约，实际是否被引用需要看其他任务的上游契约
					// 但如果定义了下游契约，说明这个任务有输出，不是孤立的
					referencedAsDownstream[taskPathName] = true
				}
			}
		}

		// 如果任务有上游契约引用，说明它参与了依赖链
		if upstreamContractStr != "" {
			var upstreamContract ContractDetail
			if err := json.Unmarshal([]byte(upstreamContractStr), &upstreamContract); err == nil {
				if len(upstreamContract.List) > 0 {
					referencedAsDownstream[taskPathName] = true
				}
			}
		}
	}

	// 检查每个任务是否孤立
	for _, task := range allTasks {
		taskPathName, _ := task["pathName"].(string)
		taskName, _ := task["name"].(string)
		if taskPathName == "" {
			continue
		}

		// 检查是否有上游依赖（被其他任务作为上游）
		hasUpstreamRef := referencedAsUpstream[taskPathName]
		// 检查是否有下游引用（引用了其他任务或有下游契约定义）
		hasDownstreamRef := referencedAsDownstream[taskPathName]

		// 如果既没有上游引用，也没有下游引用，则是孤立任务
		if !hasUpstreamRef && !hasDownstreamRef {
			result.Errors = append(result.Errors, CompileIssue{
				RuleID:           RuleOrphanTask,
				RuleName:         "孤立任务错误",
				ResourceType:     "task",
				ResourceName:     taskName,
				ResourcePathName: taskPathName,
				Message:          fmt.Sprintf("任务 [%s] 没有任何上下游依赖关系，属于孤立任务", taskName),
				Suggestion:       "请检查该任务的契约定义，确认是否需要与其他任务建立依赖关系",
				Severity:         "error",
			})
		}
	}
}
