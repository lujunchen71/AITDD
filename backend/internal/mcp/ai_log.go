package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AIAnalysisLog AI分析日志结构
// 用于记录每次AI请求和响应的完整信息
// 每个问题单独一个请求，对应一个日志文件
type AIAnalysisLog struct {
	// 请求时间戳
	Timestamp string `json:"timestamp"`
	// 项目路径名
	ProjectPathName string `json:"project_path_name"`
	// 目标路径名（任务/模块/项目的完整路径）
	TargetPathName string `json:"target_path_name"`
	// 目标名称（简短名称）
	TargetName string `json:"target_name"`
	// 分析类型（task/module/project）
	AnalysisType string `json:"analysis_type"`
	// 问题索引（从0开始，每个问题单独一个请求）
	QuestionIndex int `json:"question_index"`
	// 问题内容
	Question string `json:"question,omitempty"`
	// 问题分类（prompt/test/contract/architecture/requirement等）
	QuestionCategory string `json:"question_category,omitempty"`
	// 问题严重级别（error/warning/info）
	QuestionSeverity string `json:"question_severity,omitempty"`
	// 发送给AI的系统提示词
	SystemPrompt string `json:"system_prompt"`
	// 发送给AI的用户提示词
	UserPrompt string `json:"user_prompt"`
	// AI返回的原始响应
	RawResponse string `json:"raw_response"`
	// 解析后的结构化结果（如果有）
	ParsedResult *AnalysisResult `json:"parsed_result,omitempty"`
	// 错误信息（如果有）
	Error string `json:"error,omitempty"`
}

// SaveAIAnalysisLog 保存AI分析日志到JSON文件
// projectName: 项目路径名（如 PYQT6Calculator）
// targetPathName: 目标完整路径名（如 PYQT6Calculator/计算核心模块/表达式解析器）
// questionIndex: 问题索引（从0开始）
// analysisType: 分析类型（task/module/project）
// log: 日志内容
func SaveAIAnalysisLog(projectName string, targetPathName string, log *AIAnalysisLog) error {
	if log == nil {
		return nil
	}

	// 确保目录存在
	// 文件路径格式：.aitdd/compile/{projectPathName}/{prefix}_{targetPathName_flat}_{questionIndex}.json
	dirPath := filepath.Join(".aitdd", "compile", projectName)
	if err := os.MkdirAll(dirPath, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 生成文件名
	fileName := generateLogFileName(log.AnalysisType, projectName, targetPathName, log.QuestionIndex)

	// 生成文件路径
	filePath := filepath.Join(dirPath, fileName)

	// 设置日志的时间戳
	if log.Timestamp == "" {
		log.Timestamp = time.Now().Format(time.RFC3339)
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化日志失败: %w", err)
	}

	// 写入文件（UTF-8编码）
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入日志文件失败: %w", err)
	}

	return nil
}

// generateLogFileName 根据分析类型和目标路径名生成日志文件名
// 日志文件路径规范：
//   - task 问题：{taskPathName_flat}_{questionIndex}.json
//   - module 问题：module_{modulePathName_flat}_{questionIndex}.json
//   - project 问题：project_{questionIndex}.json
func generateLogFileName(analysisType string, projectName string, targetPathName string, questionIndex int) string {
	// 将路径中的 / 替换为 _，作为文件名的一部分
	flatPath := strings.ReplaceAll(targetPathName, "/", "_")
	// 去除项目名前缀（避免重复）
	projectPrefix := projectName + "_"
	if strings.HasPrefix(flatPath, projectPrefix) {
		flatPath = flatPath[len(projectPrefix):]
	}

	switch analysisType {
	case QuestionTypeTask:
		// task: {projectName}_{taskFlatPath}_{questionIndex}.json
		return fmt.Sprintf("%s_%s_%d.json", projectName, flatPath, questionIndex)
	case QuestionTypeModule:
		// module: module_{projectName}_{moduleFlatPath}_{questionIndex}.json
		return fmt.Sprintf("module_%s_%s_%d.json", projectName, flatPath, questionIndex)
	case QuestionTypeProject:
		// project: project_{questionIndex}.json
		return fmt.Sprintf("project_%d.json", questionIndex)
	default:
		return fmt.Sprintf("%s_%s_%d.json", analysisType, flatPath, questionIndex)
	}
}
