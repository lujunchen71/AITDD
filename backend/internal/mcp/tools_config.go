package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleInitProjectImpl 初始化项目配置
func (s *MCPServer) handleInitProjectImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取项目列表
	resp, err := http.Get(s.getApiURL() + "/projects")
	if err != nil {
		return mcp.NewToolResultText("获取项目列表失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析项目列表失败：" + err.Error()), nil
	}

	// 构建清晰的项目列表格式，方便 AI 匹配
	var output string
	output = "【可用项目列表】\n\n"
	output += "AI 请根据用户提供的项目名称，在下方列表中找到匹配的项目，然后自动调用 set_project 工具设置项目。\n"
	output += "set_project 需要参数：projectId（项目ID）、projectName（项目名称）和 pathName（项目路径名称，可选）。\n\n"
	output += "==========================================\n"

	// 解析项目列表
	if data, ok := result["data"].([]interface{}); ok {
		for i, item := range data {
			if project, ok := item.(map[string]interface{}); ok {
				id, _ := project["id"].(string)
				name, _ := project["name"].(string)
				pathName, _ := project["pathName"].(string)
				desc, _ := project["description"].(string)
				output += fmt.Sprintf("%d. 项目名称: %s\n   项目ID: %s\n   路径名称: %s\n   简介: %s\n\n", i+1, name, id, pathName, desc)
			}
		}
	} else {
		// 如果无法解析，返回原始 JSON
		data, _ := json.MarshalIndent(result, "", "  ")
		output += string(data)
	}

	output += "==========================================\n"
	output += "\n提示：告诉 AI 您要使用哪个项目（可以说项目名称），AI 会自动调用 set_project 完成设置。"

	return mcp.NewToolResultText(output), nil
}

// handleGetConfigImpl 获取当前配置
func (s *MCPServer) handleGetConfigImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	config, err := s.configManager.ToJSON()
	if err != nil {
		return mcp.NewToolResultText("获取配置失败：" + err.Error()), nil
	}
	return mcp.NewToolResultText(config), nil
}

// handleSetProjectImpl 设置当前项目
func (s *MCPServer) handleSetProjectImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID, ok1 := getParam(request, "projectId")
	projectName, ok2 := getParam(request, "projectName")
	pathName, _ := getParam(request, "pathName")

	if !ok1 || !ok2 {
		return mcp.NewToolResultText("缺少必要参数：projectId 和 projectName"), nil
	}

	if err := s.configManager.SetProject(projectID, projectName, pathName); err != nil {
		return mcp.NewToolResultText("设置项目失败：" + err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("项目设置成功！\n项目ID: %s\n项目名称: %s\n路径名称: %s", projectID, projectName, pathName)), nil
}
