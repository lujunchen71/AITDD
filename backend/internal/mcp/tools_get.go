package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// ==================== 辅助函数 ====================

// detectNodeType 通过 path 中的斜杠数量识别类型
func (s *MCPServer) detectNodeType(path string) string {
	slashes := strings.Count(path, "/")
	switch slashes {
	case 0:
		return "project"
	case 1:
		return "module"
	default:
		return "task"
	}
}

// formatAsYAML 将 map 格式化为简洁的 YAML 格式
func formatAsYAML(data map[string]interface{}, fields []string) string {
	var sb strings.Builder

	if len(fields) > 0 {
		// 只返回指定字段
		for _, field := range fields {
			if val, ok := data[field]; ok {
				sb.WriteString(fmt.Sprintf("%s: %v\n", field, formatValue(val)))
			}
		}
	} else {
		// 返回所有字段
		for key, val := range data {
			sb.WriteString(fmt.Sprintf("%s: %v\n", key, formatValue(val)))
		}
	}
	return sb.String()
}

// formatValue 格式化值用于 YAML 输出（单行格式）
func formatValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		// 空字符串
		if v == "" {
			return ""
		}
		// 尝试解析为 JSON 数组或对象
		var jsonArr []interface{}
		if err := json.Unmarshal([]byte(v), &jsonArr); err == nil {
			// 是 JSON 数组，转为 YAML 格式
			if len(jsonArr) == 0 {
				return "[]"
			}
			parts := make([]string, len(jsonArr))
			for i, item := range jsonArr {
				parts[i] = fmt.Sprintf("%v", item)
			}
			return "[" + strings.Join(parts, ", ") + "]"
		}
		var jsonObj map[string]interface{}
		if err := json.Unmarshal([]byte(v), &jsonObj); err == nil {
			// 是 JSON 对象，转为 JSON 字符串
			data, _ := json.Marshal(jsonObj)
			return string(data)
		}
		// 多行字符串或包含特殊字符用引号包裹
		if strings.Contains(v, "\n") || strings.Contains(v, ":") || strings.Contains(v, "\"") {
			return fmt.Sprintf("%q", v)
		}
		return v
	case nil:
		return ""
	case bool:
		return fmt.Sprintf("%v", v)
	case float64, float32, int, int64, int32:
		return fmt.Sprintf("%v", v)
	case []interface{}:
		if len(v) == 0 {
			return "[]"
		}
		parts := make([]string, len(v))
		for i, item := range v {
			parts[i] = fmt.Sprintf("%v", item)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]interface{}:
		// 嵌套对象转为 JSON 字符串（单行）
		data, _ := json.Marshal(v)
		return string(data)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatValueYAML 格式化值用于 YAML 输出，支持多行缩进
func formatValueYAML(val interface{}, indent string) string {
	switch v := val.(type) {
	case string:
		// 空字符串
		if v == "" {
			return ""
		}
		// 尝试解析为 JSON
		var jsonVal interface{}
		if err := json.Unmarshal([]byte(v), &jsonVal); err == nil {
			// 是 JSON 字符串，递归格式化
			return formatValueYAML(jsonVal, indent)
		}
		// 多行字符串或包含特殊字符用引号包裹
		if strings.Contains(v, "\n") || strings.Contains(v, ":") || strings.Contains(v, "\"") {
			return fmt.Sprintf("%q", v)
		}
		return v
	case nil:
		return ""
	case bool:
		return fmt.Sprintf("%v", v)
	case float64, float32, int, int64, int32:
		return fmt.Sprintf("%v", v)
	case []interface{}:
		if len(v) == 0 {
			return "[]"
		}
		var parts []string
		for _, item := range v {
			parts = append(parts, formatValueYAML(item, indent+"  "))
		}
		if len(parts) == 1 && !strings.Contains(parts[0], "\n") {
			return "[" + parts[0] + "]"
		}
		return "\n" + indent + "  - " + strings.Join(parts, "\n"+indent+"  - ")
	case map[string]interface{}:
		if len(v) == 0 {
			return "{}"
		}
		var sb strings.Builder
		sb.WriteString("")
		// 按键排序
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("\n%s  %s: %s", indent, k, formatValueYAML(v[k], indent+"  ")))
		}
		return sb.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ==================== 1. 项目上下文实现 ====================

// handleInitProjectImpl 初始化项目配置
func (s *MCPServer) handleInitProjectImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, _ := getParam(request, "projectId")
	projectName, _ := getParam(request, "projectName")
	pathName, _ := getParam(request, "pathName")

	// 如果没有提供任何参数，返回项目列表
	if projectId == "" && projectName == "" && pathName == "" {
		resp, err := http.Get(fmt.Sprintf("%s/projects", s.getApiURL()))
		if err != nil {
			return mcp.NewToolResultText("请求失败：" + err.Error()), nil
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return mcp.NewToolResultText("解析失败：" + err.Error()), nil
		}

		// 格式化项目列表为 YAML 格式
		var sb strings.Builder
		sb.WriteString("projects:\n")

		// API 返回的结构是 {"success":true,"data":{"projects": [...], "total": N}}
		var projects []interface{}
		if data, ok := result["data"].(map[string]interface{}); ok {
			projects, _ = data["projects"].([]interface{})
		}
		// 兼容直接返回 {"projects": [...]} 的情况
		if projects == nil {
			projects, _ = result["projects"].([]interface{})
		}

		for _, item := range projects {
			if project, ok := item.(map[string]interface{}); ok {
				name, _ := project["name"].(string)
				pn, _ := project["pathName"].(string)
				if pn == "" {
					pn, _ = project["path_name"].(string)
				}
				desc, _ := project["constitution"].(string)
				if desc == "" {
					desc, _ = project["description"].(string)
				}
				sb.WriteString(fmt.Sprintf("  - name: %q\n    pathName: %q\n    description: %q\n", name, pn, desc))
			}
		}
		return mcp.NewToolResultText(sb.String()), nil
	}

	// 如果提供了 pathName，直接通过 pathName 获取项目
	if pathName != "" {
		resp, err := http.Get(fmt.Sprintf("%s/projects/by-path/%s", s.getApiURL(), pathName))
		if err != nil {
			return mcp.NewToolResultText("获取项目失败：" + err.Error()), nil
		}
		defer resp.Body.Close()

		if resp.StatusCode == 404 {
			return mcp.NewToolResultText(fmt.Sprintf("未找到项目: %s", pathName)), nil
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return mcp.NewToolResultText("解析项目失败：" + err.Error()), nil
		}

		// API 返回 {"success":true,"data":{"project":{...}}}
		var project map[string]interface{}
		var found bool
		
		// 首先尝试从 data 字段中获取
		if data, ok := result["data"].(map[string]interface{}); ok {
			project, ok = data["project"].(map[string]interface{})
			if ok {
				found = true
			}
		}
		
		// 如果没有从 data 字段获取到，尝试直接获取 project 字段
		if !found {
			project, found = result["project"].(map[string]interface{})
		}
		
		if !found || project == nil {
			// 输出原始数据用于调试
			resultBytes, _ := json.Marshal(result)
			return mcp.NewToolResultText(fmt.Sprintf("项目数据格式错误，原始数据: %s", string(resultBytes))), nil
		}

		// 保存配置
		config := map[string]interface{}{
			"projectId":   project["id"],
			"projectName": project["name"],
			"pathName":    project["pathName"],
		}

		if err := s.configManager.SaveConfig(config); err != nil {
			return mcp.NewToolResultText("保存配置失败：" + err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("success: true\nprojectName: %s\npathName: %s\n", project["name"], project["pathName"])), nil
	}

	// 如果提供了 projectId 或 projectName（兼容旧逻辑）
	if projectId != "" || projectName != "" {
		var projectURL string
		if projectId != "" {
			projectURL = fmt.Sprintf("%s/projects/%s", s.getApiURL(), projectId)
		} else {
			projectURL = fmt.Sprintf("%s/projects/by-name/%s", s.getApiURL(), projectName)
		}

		resp, err := http.Get(projectURL)
		if err != nil {
			return mcp.NewToolResultText("获取项目失败：" + err.Error()), nil
		}
		defer resp.Body.Close()

		if resp.StatusCode == 404 {
			return mcp.NewToolResultText(fmt.Sprintf("未找到项目: %s", projectId+projectName)), nil
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return mcp.NewToolResultText("解析项目失败：" + err.Error()), nil
		}

		// API 返回 {"success":true,"data":{"project":{...}}}
		project, ok := result["project"].(map[string]interface{})
		if !ok {
			// 尝试从 data 字段中获取
			if data, ok := result["data"].(map[string]interface{}); ok {
				project, ok = data["project"].(map[string]interface{})
			}
		}
		if !ok {
			return mcp.NewToolResultText("项目数据格式错误"), nil
		}

		// 保存配置
		config := map[string]interface{}{
			"projectId":   project["id"],
			"projectName": project["name"],
			"pathName":    project["pathName"],
		}

		if err := s.configManager.SaveConfig(config); err != nil {
			return mcp.NewToolResultText("保存配置失败：" + err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("success: true\nprojectName: %s\npathName: %s\n", project["name"], project["pathName"])), nil
	}

	return mcp.NewToolResultText("请提供 projectId、projectName 或 pathName"), nil
}

// handleGetContextImpl 获取当前上下文
func (s *MCPServer) handleGetContextImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	config := s.configManager.GetConfig()

	// 如果未设置项目，返回提示信息
	if config == nil || config.PathName == "" {
		return mcp.NewToolResultText("未设置项目，请先使用 init_project 设置项目"), nil
	}

	// 从配置获取基本信息
	projectName := config.ProjectName
	pathName := config.PathName
	apiBaseUrl := s.getApiURL()

	// 如果配置中没有 projectName，尝试从 API 获取
	if projectName == "" && pathName != "" {
		resp, err := http.Get(fmt.Sprintf("%s/projects/by-path/%s", s.getApiURL(), pathName))
		if err == nil {
			defer resp.Body.Close()
			var result map[string]interface{}
			if json.NewDecoder(resp.Body).Decode(&result) == nil {
				// 尝试从响应中获取项目名称
				if project, ok := result["project"].(map[string]interface{}); ok {
					projectName, _ = project["name"].(string)
				} else {
					projectName, _ = result["name"].(string)
				}
			}
		}
	}

	// 返回简洁的 YAML 格式输出
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("projectName: %q\n", projectName))
	sb.WriteString(fmt.Sprintf("pathName: %q\n", pathName))
	sb.WriteString(fmt.Sprintf("apiBaseUrl: %q\n", apiBaseUrl))

	return mcp.NewToolResultText(sb.String()), nil
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ==================== 2. 信息查询实现 ====================

// handleQueryNodeImpl 通用查询接口实现
func (s *MCPServer) handleQueryNodeImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, ok := getParam(request, "path")
	if !ok || path == "" {
		return mcp.NewToolResultText("缺少 path 参数"), nil
	}

	// 获取可选的 fields 参数
	var fields []string
	if fieldsVal, ok := getParamAny(request, "fields"); ok {
		if fieldsArr, ok := fieldsVal.([]interface{}); ok {
			for _, f := range fieldsArr {
				if fStr, ok := f.(string); ok {
					fields = append(fields, fStr)
				}
			}
		}
	}

	// 识别节点类型
	nodeType := s.detectNodeType(path)

	var url string
	switch nodeType {
	case "project":
		url = fmt.Sprintf("%s/projects/by-path/%s", s.getApiURL(), path)
	case "module":
		url = fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), path)
	case "task":
		url = fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), path)
	default:
		return mcp.NewToolResultText(fmt.Sprintf("未知的节点类型: %s", nodeType)), nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return mcp.NewToolResultText(fmt.Sprintf("未找到节点: %s (类型: %s)", path, nodeType)), nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	// 从 API 响应中提取实际节点数据
	// API 返回格式: {"success":true,"data":{...}} 或 {"project":{...}} 等
	var nodeData map[string]interface{}
	
	// 首先尝试从 data 字段获取
	if data, ok := result["data"].(map[string]interface{}); ok {
		// 检查 data 中是否有对应类型的字段
		if n, ok := data[nodeType].(map[string]interface{}); ok {
			nodeData = n
		} else {
			// data 本身就是节点数据
			nodeData = data
		}
	}
	
	// 如果没有从 data 获取到，尝试直接从 result 获取
	if nodeData == nil {
		if n, ok := result[nodeType].(map[string]interface{}); ok {
			nodeData = n
		}
	}
	
	// 如果仍然没有获取到，使用整个 result
	if nodeData == nil {
		nodeData = result
	}

	// 根据节点类型格式化输出
	var yamlOutput string
	if len(fields) > 0 {
		// 只返回指定字段
		yamlOutput = formatNodeFields(nodeData, fields)
	} else {
		// 返回所有字段，按节点类型排序
		yamlOutput = formatNodeByType(nodeData, nodeType)
	}
	
	return mcp.NewToolResultText(yamlOutput), nil
}

// formatNodeFields 按指定字段格式化节点数据
func formatNodeFields(data map[string]interface{}, fields []string) string {
	var sb strings.Builder
	for _, field := range fields {
		if val, ok := data[field]; ok {
			sb.WriteString(fmt.Sprintf("%s: %s\n", field, formatValueYAML(val, "")))
		}
	}
	return sb.String()
}

// formatNodeByType 按节点类型格式化输出
func formatNodeByType(data map[string]interface{}, nodeType string) string {
	var sb strings.Builder
	
	switch nodeType {
	case "project":
		// Project 输出格式
		sb.WriteString(fmt.Sprintf("pathName: %s\n", formatValue(data["pathName"])))
		sb.WriteString(fmt.Sprintf("name: %s\n", formatValue(data["name"])))
		sb.WriteString(fmt.Sprintf("constitution: %s\n", formatValue(data["constitution"])))
		// 统计信息
		if totalModules, ok := data["totalModules"]; ok {
			sb.WriteString(fmt.Sprintf("totalModules: %v\n", totalModules))
		}
		if totalTasks, ok := data["totalTasks"]; ok {
			sb.WriteString(fmt.Sprintf("totalTasks: %v\n", totalTasks))
		}
		if completedTasks, ok := data["completedTasks"]; ok {
			sb.WriteString(fmt.Sprintf("completedTasks: %v\n", completedTasks))
		}
		
	case "module":
		// Module 输出格式
		sb.WriteString(fmt.Sprintf("pathName: %s\n", formatValue(data["pathName"])))
		sb.WriteString(fmt.Sprintf("name: %s\n", formatValue(data["name"])))
		sb.WriteString(fmt.Sprintf("description: %s\n", formatValue(data["description"])))
		sb.WriteString(fmt.Sprintf("prompt: %s\n", formatValue(data["prompt"])))
		sb.WriteString(fmt.Sprintf("status: %s\n", formatValue(data["status"])))
		sb.WriteString(fmt.Sprintf("upstreamContractSummary: %s\n", formatValue(data["upstreamContractSummary"])))
		sb.WriteString(fmt.Sprintf("downstreamContractSummary: %s\n", formatValue(data["downstreamContractSummary"])))
		if version, ok := data["version"]; ok {
			sb.WriteString(fmt.Sprintf("version: %v\n", version))
		}
		
	case "task":
		// Task 输出格式
		sb.WriteString(fmt.Sprintf("pathName: %s\n", formatValue(data["pathName"])))
		if modulePathName, ok := data["modulePathName"]; ok {
			sb.WriteString(fmt.Sprintf("modulePathName: %s\n", formatValue(modulePathName)))
		}
		sb.WriteString(fmt.Sprintf("name: %s\n", formatValue(data["name"])))
		sb.WriteString(fmt.Sprintf("description: %s\n", formatValue(data["description"])))
		sb.WriteString(fmt.Sprintf("status: %s\n", formatValue(data["status"])))
		sb.WriteString(fmt.Sprintf("prompt: %s\n", formatValue(data["prompt"])))
		// 契约字段使用多行 YAML 格式
		sb.WriteString(fmt.Sprintf("upstreamContractDetail:%s\n", formatValueYAML(data["upstreamContractDetail"], "")))
		sb.WriteString(fmt.Sprintf("downstreamContractDetail:%s\n", formatValueYAML(data["downstreamContractDetail"], "")))
		sb.WriteString(fmt.Sprintf("tests: %s\n", formatValue(data["tests"])))
		sb.WriteString(fmt.Sprintf("testResult: %s\n", formatValue(data["testResult"])))
		sb.WriteString(fmt.Sprintf("codePaths: %s\n", formatValue(data["codePaths"])))
		sb.WriteString(fmt.Sprintf("bugLog: %s\n", formatValue(data["bugLog"])))
		sb.WriteString(fmt.Sprintf("humanAssistance: %s\n", formatValue(data["humanAssistance"])))
		if locked, ok := data["locked"]; ok {
			sb.WriteString(fmt.Sprintf("locked: %v\n", locked))
		} else {
			sb.WriteString("locked: false\n")
		}
		if lockedBy, ok := data["lockedBy"]; ok && lockedBy != nil {
			sb.WriteString(fmt.Sprintf("lockedBy: %s\n", formatValue(lockedBy)))
		}
		if version, ok := data["version"]; ok {
			sb.WriteString(fmt.Sprintf("version: %v\n", version))
		}
		
	default:
		// 默认输出所有字段
		for key, val := range data {
			sb.WriteString(fmt.Sprintf("%s: %s\n", key, formatValue(val)))
		}
	}
	
	return sb.String()
}

// handleQueryProjectIndexTreeImpl 查询项目计划索引树实现
func (s *MCPServer) handleQueryProjectIndexTreeImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, _ := getParam(request, "pathName")

	// 如果没有提供 pathName，使用当前配置的项目
	if pathName == "" {
		pathName = s.configManager.GetProjectPathName()
	}
	if pathName == "" {
		return mcp.NewToolResultText("未配置项目路径名称，请先使用 init_project 设置项目，或提供 pathName"), nil
	}

	// 获取项目信息
	projectURL := fmt.Sprintf("%s/projects/by-path/%s", s.getApiURL(), pathName)
	projectResp, err := http.Get(projectURL)
	if err != nil {
		return mcp.NewToolResultText("获取项目失败：" + err.Error()), nil
	}
	defer projectResp.Body.Close()

	if projectResp.StatusCode == 404 {
		return mcp.NewToolResultText(fmt.Sprintf("未找到项目: %s", pathName)), nil
	}

	var projectResult map[string]interface{}
	if err := json.NewDecoder(projectResp.Body).Decode(&projectResult); err != nil {
		return mcp.NewToolResultText("解析项目失败：" + err.Error()), nil
	}

	// API 返回 {"success":true,"data":{"project":{...}}}
	var project map[string]interface{}
	if data, ok := projectResult["data"].(map[string]interface{}); ok {
		project, ok = data["project"].(map[string]interface{})
		if !ok {
			return mcp.NewToolResultText("项目数据格式错误"), nil
		}
	} else {
		// 兼容旧格式 {"project":{...}}
		project, ok = projectResult["project"].(map[string]interface{})
		if !ok {
			return mcp.NewToolResultText("项目数据格式错误"), nil
		}
	}

	projectId, _ := project["id"].(string)
	projectName, _ := project["name"].(string)

	// 使用 projectId 获取模块列表
	modulesURL := fmt.Sprintf("%s/modules?projectId=%s", s.getApiURL(), projectId)
	modulesResp, err := http.Get(modulesURL)
	if err != nil {
		return mcp.NewToolResultText("获取模块失败：" + err.Error()), nil
	}
	defer modulesResp.Body.Close()

	var modulesResult map[string]interface{}
	if err := json.NewDecoder(modulesResp.Body).Decode(&modulesResult); err != nil {
		return mcp.NewToolResultText("解析模块失败：" + err.Error()), nil
	}

	// 构建树形结构
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-35s # %s, project, -\n", pathName, projectName))

	// 处理模块 - API 返回 {"success":true,"data":{"modules":[...]}}
	var modules []interface{}
	if data, ok := modulesResult["data"].(map[string]interface{}); ok {
		modules, _ = data["modules"].([]interface{})
	} else {
		// 兼容旧格式 {"modules": [...]}
		modules, _ = modulesResult["modules"].([]interface{})
	}
	if len(modules) > 0 {
		buildModuleTree(&sb, modules, "", s.getApiURL())
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// getLastPathSegment 获取路径的最后一段
func getLastPathSegment(pathName string) string {
	if idx := strings.LastIndex(pathName, "/"); idx >= 0 {
		return pathName[idx+1:]
	}
	return pathName
}

// buildModuleTree 递归构建模块树
func buildModuleTree(sb *strings.Builder, modules []interface{}, prefix string, apiURL string) {
	total := len(modules)
	for i, module := range modules {
		moduleMap, ok := module.(map[string]interface{})
		if !ok {
			continue
		}

		isLast := i == total-1
		var connector, childPrefix string
		if isLast {
			connector = "└── "
			childPrefix = "    "
		} else {
			connector = "├── "
			childPrefix = "│   "
		}

		modulePathName, _ := moduleMap["pathName"].(string)
		moduleName, _ := moduleMap["name"].(string)
		moduleStatus, _ := moduleMap["status"].(string)
		if moduleStatus == "" {
			moduleStatus = "-"
		}

		// 只显示路径的最后一段
		displayName := getLastPathSegment(modulePathName)
		sb.WriteString(fmt.Sprintf("%s%s%-30s # %s, module, %s\n", prefix, connector, displayName, moduleName, moduleStatus))

		// 获取该模块的任务
		tasksURL := fmt.Sprintf("%s/modules/by-path/%s?action=tasks", apiURL, modulePathName)
		tasksResp, err := http.Get(tasksURL)
		if err == nil {
			defer tasksResp.Body.Close()
			var tasksResult map[string]interface{}
			if err := json.NewDecoder(tasksResp.Body).Decode(&tasksResult); err == nil {
				// API 返回 {"success":true,"data":{"tasks":[...]}}
				var tasks []interface{}
				if data, ok := tasksResult["data"].(map[string]interface{}); ok {
					tasks, _ = data["tasks"].([]interface{})
				} else {
					// 兼容旧格式 {"tasks": [...]}
					tasks, _ = tasksResult["tasks"].([]interface{})
				}
				for j, task := range tasks {
					taskMap, ok := task.(map[string]interface{})
					if !ok {
						continue
					}
					taskIsLast := j == len(tasks)-1
					var taskConnector string
					if taskIsLast {
						taskConnector = "└── "
					} else {
						taskConnector = "├── "
					}

					taskPathName, _ := taskMap["pathName"].(string)
					taskName, _ := taskMap["name"].(string)
					taskStatus, _ := taskMap["status"].(string)
					if taskStatus == "" {
						taskStatus = "ready"
					}

					// 只显示路径的最后一段
					taskDisplayName := getLastPathSegment(taskPathName)
					sb.WriteString(fmt.Sprintf("%s%s%s%-30s # %s, task, %s\n", prefix, childPrefix, taskConnector, taskDisplayName, taskName, taskStatus))
				}
			}
		}
	}
}

// handleQueryFileCodePathTreeImpl 查询代码文件路径树实现
func (s *MCPServer) handleQueryFileCodePathTreeImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, _ := getParam(request, "pathName")

	// 如果 pathName 为空，使用当前配置的项目
	if pathName == "" {
		pathName = s.configManager.GetProjectPathName()
	}
	if pathName == "" {
		return mcp.NewToolResultText("未配置项目路径名称，请先使用 init_project 设置项目，或提供 pathName"), nil
	}

	nodeType := s.detectNodeType(pathName)

	var sb strings.Builder
	var totalFiles int

	// 根据节点类型获取代码路径
	if nodeType == "project" {
		// 获取项目下所有模块的代码路径
		projectURL := fmt.Sprintf("%s/projects/by-path/%s", s.getApiURL(), pathName)
		projectResp, err := http.Get(projectURL)
		if err != nil {
			return mcp.NewToolResultText("获取项目失败：" + err.Error()), nil
		}
		defer projectResp.Body.Close()

		if projectResp.StatusCode == 404 {
			return mcp.NewToolResultText(fmt.Sprintf("未找到项目: %s", pathName)), nil
		}

		var projectResult map[string]interface{}
		if err := json.NewDecoder(projectResp.Body).Decode(&projectResult); err != nil {
			return mcp.NewToolResultText("解析项目失败：" + err.Error()), nil
		}

		// 从 API 响应中提取项目数据
		project, ok := projectResult["project"].(map[string]interface{})
		if !ok {
			if data, ok := projectResult["data"].(map[string]interface{}); ok {
				project, _ = data["project"].(map[string]interface{})
			}
		}
		if project == nil {
			return mcp.NewToolResultText("项目数据格式错误"), nil
		}

		projectId, _ := project["id"].(string)
		projectName, _ := project["name"].(string)
		sb.WriteString(fmt.Sprintf("%-35s # %s, project\n", pathName, projectName))

		// 获取项目下所有模块
		modulesURL := fmt.Sprintf("%s/modules?projectId=%s", s.getApiURL(), projectId)
		modulesResp, err := http.Get(modulesURL)
		if err != nil {
			return mcp.NewToolResultText("获取模块失败：" + err.Error()), nil
		}
		defer modulesResp.Body.Close()

		var modulesResult map[string]interface{}
		if err := json.NewDecoder(modulesResp.Body).Decode(&modulesResult); err != nil {
			return mcp.NewToolResultText("解析模块失败：" + err.Error()), nil
		}

		// 处理模块 - API 返回 {"success":true,"data":{"modules":[...]}}
		var modules []interface{}
		if data, ok := modulesResult["data"].(map[string]interface{}); ok {
			modules, _ = data["modules"].([]interface{})
		} else {
			modules, _ = modulesResult["modules"].([]interface{})
		}

		// 收集所有文件的代码路径信息
		type fileEntry struct {
			codePath     string
			taskName     string
			taskPathName string
		}
		var allFiles []fileEntry

		// 遍历所有模块，获取任务的代码路径
		for _, module := range modules {
			moduleMap, ok := module.(map[string]interface{})
			if !ok {
				continue
			}
			modulePathName, _ := moduleMap["pathName"].(string)

			// 获取该模块的任务
			tasksURL := fmt.Sprintf("%s/modules/by-path/%s?action=tasks", s.getApiURL(), modulePathName)
			tasksResp, err := http.Get(tasksURL)
			if err != nil {
				continue
			}
			defer tasksResp.Body.Close()

			var tasksResult map[string]interface{}
			if err := json.NewDecoder(tasksResp.Body).Decode(&tasksResult); err != nil {
				continue
			}

			var tasks []interface{}
			if data, ok := tasksResult["data"].(map[string]interface{}); ok {
				tasks, _ = data["tasks"].([]interface{})
			} else {
				tasks, _ = tasksResult["tasks"].([]interface{})
			}

			for _, task := range tasks {
				taskMap, ok := task.(map[string]interface{})
				if !ok {
					continue
				}
				taskPathName, _ := taskMap["pathName"].(string)
				taskName, _ := taskMap["name"].(string)

				// codePaths 在数据库中是 JSON 字符串，需要解析
				var codePaths []string
				if cpStr, ok := taskMap["codePaths"].(string); ok && cpStr != "" {
					json.Unmarshal([]byte(cpStr), &codePaths)
				} else if cpArray, ok := taskMap["codePaths"].([]interface{}); ok {
					for _, cp := range cpArray {
						if cpStr, ok := cp.(string); ok {
							codePaths = append(codePaths, cpStr)
						}
					}
				}

				for _, codePathStr := range codePaths {
					if codePathStr != "" {
						allFiles = append(allFiles, fileEntry{
							codePath:     codePathStr,
							taskName:     taskName,
							taskPathName: taskPathName,
						})
					}
				}
			}
		}

		// 构建目录树结构
		type treeNode struct {
			name     string
			isFile   bool
			children map[string]*treeNode
			fileInfo *fileEntry // 仅文件节点有
		}

		root := &treeNode{children: make(map[string]*treeNode)}

		for i := range allFiles {
			file := &allFiles[i]
			parts := strings.Split(file.codePath, "/")
			current := root
			for i, part := range parts {
				if part == "" {
					continue
				}
				isLast := i == len(parts)-1
				if _, exists := current.children[part]; !exists {
					if isLast {
						current.children[part] = &treeNode{
							name:     part,
							isFile:   true,
							fileInfo: file,
							children: nil,
						}
					} else {
						current.children[part] = &treeNode{
							name:     part,
							isFile:   false,
							children: make(map[string]*treeNode),
						}
					}
				}
				if !isLast {
					current = current.children[part]
				}
			}
		}

		// 递归渲染树
		var renderTree func(node *treeNode, prefix string, isLast bool, isFirst bool) int
		renderTree = func(node *treeNode, prefix string, isLast bool, isFirst bool) int {
			count := 0
			if node.isFile {
				connector := "├── "
				if isLast {
					connector = "└── "
				}
				sb.WriteString(fmt.Sprintf("%s%s%-30s # %s (%s)\n", prefix, connector, node.name, node.fileInfo.taskName, node.fileInfo.taskPathName))
				return 1
			}

			// 目录节点
			if !isFirst {
				connector := "├── "
				if isLast {
					connector = "└── "
				}
				sb.WriteString(fmt.Sprintf("%s%s%s/\n", prefix, connector, node.name))
			}

			// 获取排序后的子节点名称
			var names []string
			for name := range node.children {
				names = append(names, name)
			}
			// 排序：目录在前，文件在后，同类型按字母排序
			sort.Slice(names, func(i, j int) bool {
				ni, nj := node.children[names[i]], node.children[names[j]]
				if ni.isFile != nj.isFile {
					return !ni.isFile // 目录在前
				}
				return names[i] < names[j]
			})

			newPrefix := prefix
			if !isFirst {
				if isLast {
					newPrefix = prefix + "    "
				} else {
					newPrefix = prefix + "│   "
				}
			}

			for i, name := range names {
				childIsLast := i == len(names)-1
				count += renderTree(node.children[name], newPrefix, childIsLast, false)
			}
			return count
		}

		// 渲染根节点的子节点
		var rootNames []string
		for name := range root.children {
			rootNames = append(rootNames, name)
		}
		sort.Slice(rootNames, func(i, j int) bool {
			ni, nj := root.children[rootNames[i]], root.children[rootNames[j]]
			if ni.isFile != nj.isFile {
				return !ni.isFile
			}
			return rootNames[i] < rootNames[j]
		})

		for i, name := range rootNames {
			childIsLast := i == len(rootNames)-1
			totalFiles += renderTree(root.children[name], "", childIsLast, false)
		}

	} else if nodeType == "module" {
		// 获取模块信息
		moduleURL := fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), pathName)
		moduleResp, err := http.Get(moduleURL)
		if err != nil {
			return mcp.NewToolResultText("获取模块失败：" + err.Error()), nil
		}
		defer moduleResp.Body.Close()

		if moduleResp.StatusCode == 404 {
			return mcp.NewToolResultText(fmt.Sprintf("未找到模块: %s", pathName)), nil
		}

		var moduleResult map[string]interface{}
		if err := json.NewDecoder(moduleResp.Body).Decode(&moduleResult); err != nil {
			return mcp.NewToolResultText("解析模块失败：" + err.Error()), nil
		}

		// 从 API 响应中提取模块数据
		module, ok := moduleResult["module"].(map[string]interface{})
		if !ok {
			// 尝试从 data 字段获取
			if data, ok := moduleResult["data"].(map[string]interface{}); ok {
				module, _ = data["module"].(map[string]interface{})
			}
		}
		if module == nil {
			return mcp.NewToolResultText("模块数据格式错误"), nil
		}

		moduleName, _ := module["name"].(string)
		// 只显示路径的最后一段
		displayName := getLastPathSegment(pathName)
		sb.WriteString(fmt.Sprintf("%-35s # %s, module\n", displayName, moduleName))

		// 获取模块下所有任务的代码路径
		tasksURL := fmt.Sprintf("%s/modules/by-path/%s?action=tasks", s.getApiURL(), pathName)
		tasksResp, err := http.Get(tasksURL)
		if err != nil {
			return mcp.NewToolResultText("获取任务失败：" + err.Error()), nil
		}
		defer tasksResp.Body.Close()

		var tasksResult map[string]interface{}
		if err := json.NewDecoder(tasksResp.Body).Decode(&tasksResult); err != nil {
			return mcp.NewToolResultText("解析任务失败：" + err.Error()), nil
		}

		// 收集所有任务的代码路径
		type fileEntry struct {
			codePath     string
			taskName     string
			taskPathName string
		}
		var allFiles []fileEntry

		// API 返回 {"success":true,"data":{"tasks":[...]}} 或 {"tasks": [...]}
		var tasks []interface{}
		if data, ok := tasksResult["data"].(map[string]interface{}); ok {
			tasks, _ = data["tasks"].([]interface{})
		} else {
			// 兼容旧格式 {"tasks": [...]}
			tasks, _ = tasksResult["tasks"].([]interface{})
		}

		for _, task := range tasks {
			taskMap, ok := task.(map[string]interface{})
			if !ok {
				continue
			}
			taskPathName, _ := taskMap["pathName"].(string)
			taskName, _ := taskMap["name"].(string)

			// codePaths 在数据库中是 JSON 字符串，需要解析
			var codePaths []string
			if cpStr, ok := taskMap["codePaths"].(string); ok && cpStr != "" {
				json.Unmarshal([]byte(cpStr), &codePaths)
			} else if cpArray, ok := taskMap["codePaths"].([]interface{}); ok {
				// 兼容已经是数组的情况
				for _, cp := range cpArray {
					if cpStr, ok := cp.(string); ok {
						codePaths = append(codePaths, cpStr)
					}
				}
			}

			for _, codePathStr := range codePaths {
				if codePathStr != "" {
					allFiles = append(allFiles, fileEntry{
						codePath:     codePathStr,
						taskName:     taskName,
						taskPathName: taskPathName,
					})
				}
			}
		}

		// 输出文件树
		totalFiles = len(allFiles)
		for i, file := range allFiles {
			isLast := i == totalFiles-1
			var connector string
			if isLast {
				connector = "└── "
			} else {
				connector = "├── "
			}
			sb.WriteString(fmt.Sprintf("%s%-30s # %s (%s)\n", connector, file.codePath, file.taskName, file.taskPathName))
		}

	} else if nodeType == "task" {
		// 获取单个任务的代码路径
		taskURL := fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), pathName)
		taskResp, err := http.Get(taskURL)
		if err != nil {
			return mcp.NewToolResultText("获取任务失败：" + err.Error()), nil
		}
		defer taskResp.Body.Close()

		if taskResp.StatusCode == 404 {
			return mcp.NewToolResultText(fmt.Sprintf("未找到任务: %s", pathName)), nil
		}

		var taskResult map[string]interface{}
		if err := json.NewDecoder(taskResp.Body).Decode(&taskResult); err != nil {
			return mcp.NewToolResultText("解析任务失败：" + err.Error()), nil
		}

		// 从 API 响应中提取任务数据
		task, ok := taskResult["task"].(map[string]interface{})
		if !ok {
			// 尝试从 data 字段获取
			if data, ok := taskResult["data"].(map[string]interface{}); ok {
				task, _ = data["task"].(map[string]interface{})
			}
		}
		if task == nil {
			return mcp.NewToolResultText("任务数据格式错误"), nil
		}

		taskName, _ := task["name"].(string)
		// 只显示路径的最后一段
		displayName := getLastPathSegment(pathName)
		sb.WriteString(fmt.Sprintf("%-35s # %s, task\n", displayName, taskName))

		// codePaths 在数据库中是 JSON 字符串，需要解析
		var codePaths []string
		if cpStr, ok := task["codePaths"].(string); ok && cpStr != "" {
			json.Unmarshal([]byte(cpStr), &codePaths)
		} else if cpArray, ok := task["codePaths"].([]interface{}); ok {
			// 兼容已经是数组的情况
			for _, cp := range cpArray {
				if cpStr, ok := cp.(string); ok {
					codePaths = append(codePaths, cpStr)
				}
			}
		}

		fileCount := len(codePaths)
		for i, codePathStr := range codePaths {
			if codePathStr != "" {
				isLast := i == fileCount-1
				var connector string
				if isLast {
					connector = "└── "
				} else {
					connector = "├── "
				}
				sb.WriteString(fmt.Sprintf("%s%-30s\n", connector, codePathStr))
				totalFiles++
			}
		}
	} else {
		return mcp.NewToolResultText(fmt.Sprintf("不支持的节点类型: %s (支持 project, module 和 task)", nodeType)), nil
	}

	sb.WriteString(fmt.Sprintf("totalFiles: %d\n", totalFiles))
	return mcp.NewToolResultText(sb.String()), nil
}

// ==================== query_module 和 query_task 实现 ====================

// Module 可查询字段:
// - pathName: 模块路径名称 (只读)
// - name: 模块名称
// - description: 模块描述
// - status: 状态 (pending/in_progress/completed/blocked)
// - prompt: 提示词内容
// - upstreamContractSummary: 上游契约摘要
// - downstreamContractSummary: 下游契约摘要
// - testCoverage: 测试覆盖率
// - locked: 是否被锁定 (只读)
// - version: 版本号 (只读，用于乐观锁)

// handleQueryModuleImpl 查询模块信息
// 参数:
//   - pathName: 模块路径名称 (必填)，格式: "项目名/模块名"
//   - fields: 要查询的字段列表 (可选)，不填则返回所有字段
// 可查询字段: pathName, name, description, status, prompt, upstreamContractSummary, downstreamContractSummary, testCoverage, locked, version
func (s *MCPServer) handleQueryModuleImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	// 获取可选的 fields 参数
	var fields []string
	if fieldsVal, ok := getParamAny(request, "fields"); ok {
		if fieldsArr, ok := fieldsVal.([]interface{}); ok {
			for _, f := range fieldsArr {
				if fStr, ok := f.(string); ok {
					fields = append(fields, fStr)
				}
			}
		}
	}

	// 调用 API 获取模块数据
	url := fmt.Sprintf("%s/modules/by-path/%s", s.getApiURL(), pathName)
	resp, err := http.Get(url)
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return mcp.NewToolResultText(fmt.Sprintf("未找到模块: %s", pathName)), nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	// 从 API 响应中提取模块数据
	var moduleData map[string]interface{}
	if data, ok := result["data"].(map[string]interface{}); ok {
		if m, ok := data["module"].(map[string]interface{}); ok {
			moduleData = m
		} else {
			moduleData = data
		}
	}
	if moduleData == nil {
		if m, ok := result["module"].(map[string]interface{}); ok {
			moduleData = m
		}
	}
	if moduleData == nil {
		moduleData = result
	}

	// 格式化输出
	var yamlOutput string
	if len(fields) > 0 {
		yamlOutput = formatNodeFields(moduleData, fields)
	} else {
		yamlOutput = formatNodeByType(moduleData, "module")
	}

	return mcp.NewToolResultText(yamlOutput), nil
}

// Task 可查询字段:
// - pathName: 任务路径名称 (只读)
// - name: 任务名称
// - description: 任务描述
// - status: 状态 (pending/in_progress/completed/blocked)
// - prompt: 提示词内容
// - upstreamContractDetail: 上游契约详情，格式: {title:string, list:[{label:string, contract_api:string, from:string}]}
// - downstreamContractDetail: 下游契约详情，格式: {title:string, list:[{label:string, contract_api:string, from:string}]}
// - tests: 测试用例列表，格式: [{target:string, api:string}]
// - testResult: 测试结果 (pass/fail/pending)
// - codePaths: 代码文件路径列表，格式: ["path/to/file1.ts", "path/to/file2.ts"]
// - bugLog: bug 日志
// - humanAssistance: 是否需要人工协助
// - issueDetails: 问题详情
// - locked: 是否被锁定 (只读)
// - version: 版本号 (只读，用于乐观锁)

// handleQueryTaskImpl 查询任务信息
// 参数:
//   - pathName: 任务路径名称 (必填)，格式: "项目名/模块名/任务名"
//   - fields: 要查询的字段列表 (可选)，不填则返回所有字段
// 可查询字段: pathName, name, description, status, prompt, upstreamContractDetail, downstreamContractDetail, tests, testResult, codePaths, bugLog, humanAssistance, issueDetails, locked, version
func (s *MCPServer) handleQueryTaskImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathName, ok := getParam(request, "pathName")
	if !ok || pathName == "" {
		return mcp.NewToolResultText("缺少 pathName 参数"), nil
	}

	// 获取可选的 fields 参数
	var fields []string
	if fieldsVal, ok := getParamAny(request, "fields"); ok {
		if fieldsArr, ok := fieldsVal.([]interface{}); ok {
			for _, f := range fieldsArr {
				if fStr, ok := f.(string); ok {
					fields = append(fields, fStr)
				}
			}
		}
	}

	// 调用 API 获取任务数据
	url := fmt.Sprintf("%s/tasks/by-path/%s", s.getApiURL(), pathName)
	resp, err := http.Get(url)
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return mcp.NewToolResultText(fmt.Sprintf("未找到任务: %s", pathName)), nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	// 从 API 响应中提取任务数据
	var taskData map[string]interface{}
	if data, ok := result["data"].(map[string]interface{}); ok {
		if t, ok := data["task"].(map[string]interface{}); ok {
			taskData = t
		} else {
			taskData = data
		}
	}
	if taskData == nil {
		if t, ok := result["task"].(map[string]interface{}); ok {
			taskData = t
		}
	}
	if taskData == nil {
		taskData = result
	}

	// 格式化输出
	var yamlOutput string
	if len(fields) > 0 {
		yamlOutput = formatNodeFields(taskData, fields)
	} else {
		yamlOutput = formatNodeByType(taskData, "task")
	}

	return mcp.NewToolResultText(yamlOutput), nil
}
