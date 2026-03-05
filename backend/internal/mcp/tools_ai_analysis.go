package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aitdd/backend/internal/database"
	"github.com/aitdd/backend/internal/models"
)

// CompileDynamicAIParams compile_dynamic工具的AI分析参数
// 这些配置内容由AI助手从项目根目录的 .aitdd/ 文件夹中读取后传入
type CompileDynamicAIParams struct {
	// ProjectName 项目名称
	ProjectName string `json:"project_name"`
	// SubAgentConfig AI子代理配置内容（从 .aitdd/sub_agent.json 读取）
	SubAgentConfig interface{} `json:"sub_agent_config"`
	// QAConfig 问题配置内容（从 .aitdd/qa.json 读取）
	QAConfig interface{} `json:"qa_config"`
}

// ParseAIParams 从compile_dynamic工具参数中解析AI分析参数
func ParseAIParams(args map[string]interface{}) (*CompileDynamicAIParams, error) {
	params := &CompileDynamicAIParams{}

	if v, ok := args["project_name"]; ok {
		if s, ok := v.(string); ok {
			params.ProjectName = s
		}
	}

	if v, ok := args["sub_agent_config"]; ok {
		params.SubAgentConfig = v
	}

	if v, ok := args["qa_config"]; ok {
		params.QAConfig = v
	}

	return params, nil
}

// ========== 单问题分析请求结构 ==========

// singleQuestionJob 单个问题的分析任务
type singleQuestionJob struct {
	// 任务类型：task / module / project
	analysisType string
	// 目标路径名（用于标识和日志）
	targetPathName string
	// 目标名称（显示用）
	targetName string
	// 问题索引（该问题在该目标的问题列表中的位置）
	questionIndex int
	// 问题定义
	question Question
	// 系统提示词
	systemPrompt string
	// 用户提示词（包含完整上下文和该单一问题）
	userPrompt string
}

// singleQuestionResult 单个问题的分析结果
type singleQuestionResult struct {
	job    *singleQuestionJob
	result *AnalysisResult
	err    string
}

// targetAnalysisSummary 一个目标（task/module/project）的汇总分析结果
type targetAnalysisSummary struct {
	analysisType   string
	targetPathName string
	targetName     string
	// 按问题索引顺序存储每个问题的分析结果
	questionResults []*singleQuestionResult
	// 所有问题的平均分
	avgScore float64
}

// ========== RunAIAnalysisForCompile 主入口 ==========

// RunAIAnalysisForCompile 在compile_dynamic工具中执行AI分析
// 新逻辑：
//   - 对每个 task，对每个 qa_config.questions.task 问题，单独发一个 AI 请求
//   - 对每个 module（从 tasks 中提取），对每个 qa_config.questions.module 问题，单独发一个 AI 请求
//   - 对项目本身，对每个 qa_config.questions.project 问题，单独发一个 AI 请求
//   - 所有请求并发执行（受 max_concurrent 控制）
//   - 按目标聚合结果，生成报表
//
// projectName: 项目名称（pathName格式）
// subAgentConfigRaw: 来自参数的sub_agent配置（JSON字符串或map）
// qaConfigRaw: 来自参数的qa配置（JSON字符串或map）
// tasks: 已收集的任务列表（map[string]interface{}）
func RunAIAnalysisForCompile(ctx context.Context, projectName string, subAgentConfigRaw interface{}, qaConfigRaw interface{}, tasks []map[string]interface{}) (string, error) {
	// 1. 清理 compile 目录中的旧 JSON 文件（每次编译前清理）
	compileDir := filepath.Join(".aitdd", "compile", projectName)
	if err := cleanCompileDir(compileDir); err != nil {
		// 清理失败不阻断编译，只记录日志
		logger := GetMCPLogger()
		logger.Warn("清理compile目录失败", map[string]interface{}{"dir": compileDir, "error": err.Error()})
	}

	// 解析 SubAgentConfig
	var subAgentConfig *SubAgentConfig
	if subAgentConfigRaw != nil {
		var err error
		subAgentConfig, err = parseSubAgentConfig(subAgentConfigRaw)
		if err != nil {
			return "", fmt.Errorf("解析SubAgentConfig失败: %w", err)
		}
	}

	// 解析 QAConfig
	var qaConfig *QAConfig
	if qaConfigRaw != nil {
		var err error
		qaConfig, err = parseQAConfig(qaConfigRaw)
		if err != nil {
			return "", fmt.Errorf("解析QAConfig失败: %w", err)
		}
	}
	if qaConfig == nil {
		qaConfig = GetDefaultQAConfig()
	}

	// 如果没有任务，返回提示
	if len(tasks) == 0 {
		return fmt.Sprintf("项目 %s 没有找到任务节点，跳过AI分析", projectName), nil
	}

	// 创建 AI 客户端
	var aiClient *AIClient
	if subAgentConfig != nil && len(subAgentConfig.Models) > 0 {
		aiClient = NewAIClient(subAgentConfig)
	} else {
		aiClient = NewAIClientWithLogger(GetSubAgentConfigManager().GetConfig(), GetMCPLogger())
	}

	logger := GetMCPLogger()
	generator := NewAnalysisGeneratorWithLogger(qaConfig, logger)

	// 确定并发数
	maxConcurrent := 3
	if subAgentConfig != nil && subAgentConfig.RateLimit.MaxConcurrent > 0 {
		maxConcurrent = subAgentConfig.RateLimit.MaxConcurrent
	}

	// ========== 1. 提取模块信息（从 tasks 中推断）==========
	moduleMap := extractModulesFromTasks(tasks, projectName)

	// ========== 2. 提取项目信息 ==========
	projectInfo := extractProjectInfo(tasks, projectName)

	// ========== 3. 生成所有待执行的单问题任务 ==========
	taskQuestions := qaConfig.GetTaskQuestions()
	moduleQuestions := qaConfig.GetModuleQuestions()
	projectQuestions := qaConfig.GetProjectQuestions()

	var jobs []*singleQuestionJob

	// 3.1 task 问题
	for _, taskData := range tasks {
		taskPathName, _ := taskData["pathName"].(string)
		taskName, _ := taskData["name"].(string)
		if taskPathName == "" {
			taskPathName = taskName
		}

		// 构建 task 上下文（含上游/下游任务信息、模块信息、项目信息）
		taskContextStr := buildTaskContext(taskData, tasks, moduleMap, projectInfo, projectName)

		for qi, q := range taskQuestions {
			userPrompt := generator.BuildTaskSingleQuestionPrompt(taskData, taskContextStr, q)
			systemPrompt := generator.BuildSystemPrompt(QuestionTypeTask)

			jobs = append(jobs, &singleQuestionJob{
				analysisType:   QuestionTypeTask,
				targetPathName: taskPathName,
				targetName:     taskName,
				questionIndex:  qi,
				question:       q,
				systemPrompt:   systemPrompt,
				userPrompt:     userPrompt,
			})
		}
	}

	// 3.2 module 问题
	for modPathName, modData := range moduleMap {
		modName, _ := modData["name"].(string)
		if modName == "" {
			modName = modPathName
		}

		// 构建 module 上下文
		moduleContextStr := buildModuleContext(modPathName, modData, tasks, moduleMap, projectInfo, projectName)

		for qi, q := range moduleQuestions {
			userPrompt := generator.BuildModuleSingleQuestionPrompt(modData, moduleContextStr, q)
			systemPrompt := generator.BuildSystemPrompt(QuestionTypeModule)

			jobs = append(jobs, &singleQuestionJob{
				analysisType:   QuestionTypeModule,
				targetPathName: modPathName,
				targetName:     modName,
				questionIndex:  qi,
				question:       q,
				systemPrompt:   systemPrompt,
				userPrompt:     userPrompt,
			})
		}
	}

	// 3.3 project 问题
	for qi, q := range projectQuestions {
		userPrompt := generator.BuildProjectSingleQuestionPrompt(projectInfo, moduleMap, tasks, q)
		systemPrompt := generator.BuildSystemPrompt(QuestionTypeProject)

		jobs = append(jobs, &singleQuestionJob{
			analysisType:   QuestionTypeProject,
			targetPathName: projectName,
			targetName:     projectName,
			questionIndex:  qi,
			question:       q,
			systemPrompt:   systemPrompt,
			userPrompt:     userPrompt,
		})
	}

	logger.Info("开始并发AI分析（一问一请求）", map[string]interface{}{
		"task_count":      len(tasks),
		"module_count":    len(moduleMap),
		"total_jobs":      len(jobs),
		"max_concurrent":  maxConcurrent,
		"task_questions":  len(taskQuestions),
		"module_questions": len(moduleQuestions),
		"project_questions": len(projectQuestions),
	})

	// ========== 4. 并发执行所有问题请求 ==========
	resultCh := make(chan *singleQuestionResult, len(jobs))
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for _, job := range jobs {
		select {
		case <-ctx.Done():
			goto done
		default:
		}

		wg.Add(1)
		go func(j *singleQuestionJob) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			select {
			case <-ctx.Done():
				resultCh <- &singleQuestionResult{job: j, err: "上下文已取消"}
				return
			default:
			}

			aiReq := &AIRequest{
				Messages: []Message{
					{Role: "system", Content: j.systemPrompt},
					{Role: "user", Content: j.userPrompt},
				},
			}

			aiResp, err := aiClient.Call(ctx, aiReq)

			// 准备日志
			logEntry := &AIAnalysisLog{
				ProjectPathName:  projectName,
				TargetPathName:   j.targetPathName,
				TargetName:       j.targetName,
				AnalysisType:     j.analysisType,
				QuestionIndex:    j.questionIndex,
				Question:         j.question.Question,
				QuestionCategory: j.question.Category,
				QuestionSeverity: j.question.Severity,
				SystemPrompt:     j.systemPrompt,
				UserPrompt:       j.userPrompt,
			}

			if err != nil {
				logEntry.Error = err.Error()
				if saveErr := SaveAIAnalysisLog(projectName, j.targetPathName, logEntry); saveErr != nil {
					logger.Warn("保存AI分析日志失败", map[string]interface{}{"error": saveErr.Error()})
				}
				resultCh <- &singleQuestionResult{job: j, err: fmt.Sprintf("[%s] 问题%d AI调用失败: %v", j.targetPathName, j.questionIndex, err)}
				return
			}

			logEntry.RawResponse = aiResp.Content

			result, err := generator.ParseAIResponseByPath(aiResp.Content, j.analysisType, j.targetPathName, j.targetName)
			if err != nil {
				logEntry.Error = err.Error()
				if saveErr := SaveAIAnalysisLog(projectName, j.targetPathName, logEntry); saveErr != nil {
					logger.Warn("保存AI分析日志失败", map[string]interface{}{"error": saveErr.Error()})
				}
				resultCh <- &singleQuestionResult{job: j, err: fmt.Sprintf("[%s] 问题%d 解析AI响应失败: %v", j.targetPathName, j.questionIndex, err)}
				return
			}

			logEntry.ParsedResult = result
			if saveErr := SaveAIAnalysisLog(projectName, j.targetPathName, logEntry); saveErr != nil {
				logger.Warn("保存AI分析日志失败", map[string]interface{}{"error": saveErr.Error()})
			}

			resultCh <- &singleQuestionResult{job: j, result: result}
		}(job)
	}

done:
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// ========== 5. 收集结果 ==========
	var allResults []*singleQuestionResult
	var analyzeErrors []string

	for r := range resultCh {
		if r.err != "" {
			analyzeErrors = append(analyzeErrors, r.err)
		}
		allResults = append(allResults, r)
	}

	logger.Info("并发AI分析完成", map[string]interface{}{
		"total_jobs":    len(jobs),
		"result_count":  len(allResults),
		"error_count":   len(analyzeErrors),
	})

	// ========== 6. 按目标聚合结果，生成报表 ==========
	output := formatNewCompileAIReport(
		projectName,
		tasks,
		moduleMap,
		projectInfo,
		taskQuestions,
		moduleQuestions,
		projectQuestions,
		allResults,
		analyzeErrors,
		maxConcurrent,
	)

	// ========== 7. 将动态编译结果写入数据库 bug_log.dynamic ==========
	// 按目标汇总 issues，然后写入数据库
	go writeDynamicBugLogToDB(projectName, tasks, moduleMap, projectInfo, allResults)

	return output, nil
}

// cleanCompileDir 清理 compile 目录中的所有 .json 文件
func cleanCompileDir(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// 目录不存在，无需清理
		return nil
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("读取目录失败: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			filePath := filepath.Join(dirPath, entry.Name())
			if err := os.Remove(filePath); err != nil {
				// 单个文件删除失败不阻断，继续处理其他文件
				continue
			}
		}
	}
	return nil
}

// writeDynamicBugLogToDB 将动态编译（AI分析）结果写入数据库的 bug_log.dynamic 字段
// 每个 task/module/project 的所有问题汇总，按 severity 分类后写入
// 完全覆盖 dynamic 字段（仅针对被编译范围内的资源）
func writeDynamicBugLogToDB(projectName string, tasks []map[string]interface{}, moduleMap map[string]map[string]interface{}, projectInfo map[string]interface{}, allResults []*singleQuestionResult) {
	if database.DB == nil {
		return
	}

	logger := GetMCPLogger()

	// 按目标路径汇总 issues（错误和警告）
	// key: targetPathName
	taskErrors := make(map[string][]string)   // targetPathName -> error messages
	taskWarnings := make(map[string][]string)  // targetPathName -> warning messages
	moduleErrors := make(map[string][]string)
	moduleWarnings := make(map[string][]string)
	var projectErrors []string
	var projectWarnings []string

	for _, r := range allResults {
		if r.job == nil || r.result == nil {
			continue
		}
		targetPathName := r.job.targetPathName
		for _, issue := range r.result.Issues {
			isSevere := issue.Severity == "error" || issue.Severity == "critical"
			msg := issue.Description
			if issue.Suggestion != "" {
				msg = msg + " → " + issue.Suggestion
			}
			switch r.job.analysisType {
			case QuestionTypeTask:
				if isSevere {
					taskErrors[targetPathName] = append(taskErrors[targetPathName], msg)
				} else {
					taskWarnings[targetPathName] = append(taskWarnings[targetPathName], msg)
				}
			case QuestionTypeModule:
				if isSevere {
					moduleErrors[targetPathName] = append(moduleErrors[targetPathName], msg)
				} else {
					moduleWarnings[targetPathName] = append(moduleWarnings[targetPathName], msg)
				}
			case QuestionTypeProject:
				if isSevere {
					projectErrors = append(projectErrors, msg)
				} else {
					projectWarnings = append(projectWarnings, msg)
				}
			}
		}
	}

	// 更新 task 的 bug_log.dynamic
	for _, taskData := range tasks {
		taskPathName, _ := taskData["pathName"].(string)
		if taskPathName == "" {
			continue
		}
		var task models.Task
		if err := database.DB.Where("path_name = ?", taskPathName).First(&task).Error; err != nil {
			continue
		}
		existingBugLog := ParseBugLog(task.BugLog)
		existingBugLog.Dynamic.Error = taskErrors[taskPathName]
		if existingBugLog.Dynamic.Error == nil {
			existingBugLog.Dynamic.Error = []string{}
		}
		existingBugLog.Dynamic.Warning = taskWarnings[taskPathName]
		if existingBugLog.Dynamic.Warning == nil {
			existingBugLog.Dynamic.Warning = []string{}
		}
		bugLogBytes, err := json.Marshal(existingBugLog)
		if err != nil {
			continue
		}
		if err := database.DB.Model(&task).Updates(map[string]interface{}{
			"bug_log":    string(bugLogBytes),
			"updated_at": time.Now().UnixMilli(),
		}).Error; err != nil {
			logger.Warn("更新task bug_log.dynamic 失败", map[string]interface{}{
				"taskPathName": taskPathName,
				"error":        err.Error(),
			})
		}
	}

	// 更新 module 的 bug_log.dynamic
	for modPathName := range moduleMap {
		var module models.Module
		if err := database.DB.Where("path_name = ?", modPathName).First(&module).Error; err != nil {
			continue
		}
		existingBugLog := ParseBugLog(module.BugLog)
		existingBugLog.Dynamic.Error = moduleErrors[modPathName]
		if existingBugLog.Dynamic.Error == nil {
			existingBugLog.Dynamic.Error = []string{}
		}
		existingBugLog.Dynamic.Warning = moduleWarnings[modPathName]
		if existingBugLog.Dynamic.Warning == nil {
			existingBugLog.Dynamic.Warning = []string{}
		}
		bugLogBytes, err := json.Marshal(existingBugLog)
		if err != nil {
			continue
		}
		if err := database.DB.Model(&module).Updates(map[string]interface{}{
			"bug_log":    string(bugLogBytes),
			"updated_at": time.Now().UnixMilli(),
		}).Error; err != nil {
			logger.Warn("更新module bug_log.dynamic 失败", map[string]interface{}{
				"modPathName": modPathName,
				"error":       err.Error(),
			})
		}
	}

	// 更新 project 的 bug_log.dynamic
	var project models.Project
	if err := database.DB.Where("path_name = ?", projectName).First(&project).Error; err == nil {
		existingBugLog := ParseBugLog(project.BugLog)
		existingBugLog.Dynamic.Error = projectErrors
		if existingBugLog.Dynamic.Error == nil {
			existingBugLog.Dynamic.Error = []string{}
		}
		existingBugLog.Dynamic.Warning = projectWarnings
		if existingBugLog.Dynamic.Warning == nil {
			existingBugLog.Dynamic.Warning = []string{}
		}
		bugLogBytes, err := json.Marshal(existingBugLog)
		if err == nil {
			if err := database.DB.Model(&project).Updates(map[string]interface{}{
				"bug_log":    string(bugLogBytes),
				"updated_at": time.Now().UnixMilli(),
			}).Error; err != nil {
				logger.Warn("更新project bug_log.dynamic 失败", map[string]interface{}{
					"projectName": projectName,
					"error":       err.Error(),
				})
			}
		}
	}

	_ = projectInfo // 已通过 projectName 查询 project 记录
}

// ========== 上下文构建函数 ==========

// extractModulesFromTasks 从任务列表中提取模块信息（通过 pathName 推断）
// key: modulePathName, value: 模块基本信息
func extractModulesFromTasks(tasks []map[string]interface{}, projectName string) map[string]map[string]interface{} {
	moduleMap := make(map[string]map[string]interface{})

	for _, task := range tasks {
		pathName, _ := task["pathName"].(string)
		if pathName == "" {
			continue
		}
		parts := strings.Split(pathName, "/")
		if len(parts) < 2 {
			continue
		}
		// modulePathName = "ProjectName/ModuleName"
		modulePathName := strings.Join(parts[:len(parts)-1], "/")
		moduleName := parts[len(parts)-2]

		if _, exists := moduleMap[modulePathName]; !exists {
			moduleMap[modulePathName] = map[string]interface{}{
				"pathName": modulePathName,
				"name":     moduleName,
			}
		}
	}

	return moduleMap
}

// extractProjectInfo 从任务列表推断项目信息
func extractProjectInfo(tasks []map[string]interface{}, projectName string) map[string]interface{} {
	projectInfo := map[string]interface{}{
		"pathName": projectName,
		"name":     projectName,
	}
	// 尝试从任务字段提取项目信息
	if len(tasks) > 0 {
		if projectNameVal, ok := tasks[0]["projectName"]; ok {
			projectInfo["name"] = projectNameVal
		}
	}
	return projectInfo
}

// buildTaskContext 为 task 构建参考上下文（上游任务、下游任务、模块信息、项目信息）
func buildTaskContext(
	taskData map[string]interface{},
	allTasks []map[string]interface{},
	moduleMap map[string]map[string]interface{},
	projectInfo map[string]interface{},
	projectName string,
) string {
	var sb strings.Builder

	taskPathName, _ := taskData["pathName"].(string)
	taskName, _ := taskData["name"].(string)

	// 推断所属模块路径
	modulePath := ""
	if taskPathName != "" {
		parts := strings.Split(taskPathName, "/")
		if len(parts) >= 2 {
			modulePath = strings.Join(parts[:len(parts)-1], "/")
		}
	}

	// 所属模块信息
	if modulePath != "" {
		if modInfo, ok := moduleMap[modulePath]; ok {
			sb.WriteString("\n## 所属模块信息\n")
			sb.WriteString(fmt.Sprintf("**模块路径**: %s\n", modulePath))
			if modName, ok := modInfo["name"].(string); ok {
				sb.WriteString(fmt.Sprintf("**模块名称**: %s\n", modName))
			}
		}
	}

	// 查找上游任务（基于 upstreamContractDetail 中的 from 字段）
	upstreamTasks := findUpstreamTasks(taskData, allTasks, taskName)
	if len(upstreamTasks) > 0 {
		sb.WriteString("\n## 上游契约任务信息（参考）\n")
		for _, ut := range upstreamTasks {
			sb.WriteString(fmt.Sprintf("\n### 上游任务: %s\n", getStrField(ut, "name")))
			sb.WriteString(fmt.Sprintf("**路径**: %s\n", getStrField(ut, "pathName")))
			if desc := getStrField(ut, "description"); desc != "" {
				sb.WriteString(fmt.Sprintf("**描述**: %s\n", desc))
			}
			if downstream := ut["downstreamContractDetail"]; downstream != nil {
				sb.WriteString(fmt.Sprintf("**下游契约定义**: %v\n", downstream))
			}
		}
	}

	// 查找下游任务（基于 downstreamContractDetail 中的 from 字段）
	downstreamTasks := findDownstreamTasks(taskData, allTasks, taskName)
	if len(downstreamTasks) > 0 {
		sb.WriteString("\n## 下游契约任务信息（参考）\n")
		for _, dt := range downstreamTasks {
			sb.WriteString(fmt.Sprintf("\n### 下游任务: %s\n", getStrField(dt, "name")))
			sb.WriteString(fmt.Sprintf("**路径**: %s\n", getStrField(dt, "pathName")))
			if desc := getStrField(dt, "description"); desc != "" {
				sb.WriteString(fmt.Sprintf("**描述**: %s\n", desc))
			}
			if upstream := dt["upstreamContractDetail"]; upstream != nil {
				sb.WriteString(fmt.Sprintf("**上游契约引用**: %v\n", upstream))
			}
		}
	}

	// 项目信息
	sb.WriteString("\n## 项目信息（参考）\n")
	sb.WriteString(fmt.Sprintf("**项目名称**: %s\n", getStrField(projectInfo, "name")))
	sb.WriteString(fmt.Sprintf("**项目路径**: %s\n", projectName))

	_ = taskPathName
	return sb.String()
}

// buildModuleContext 为 module 构建参考上下文（上游/下游模块信息、项目信息）
func buildModuleContext(
	modulePathName string,
	modData map[string]interface{},
	allTasks []map[string]interface{},
	moduleMap map[string]map[string]interface{},
	projectInfo map[string]interface{},
	projectName string,
) string {
	var sb strings.Builder

	// 当前模块的任务列表
	moduleTasks := getTasksInModule(allTasks, modulePathName)
	if len(moduleTasks) > 0 {
		sb.WriteString("\n## 当前模块任务详情\n")
		for _, t := range moduleTasks {
			sb.WriteString(fmt.Sprintf("\n### 任务: %s\n", getStrField(t, "name")))
			sb.WriteString(fmt.Sprintf("**路径**: %s\n", getStrField(t, "pathName")))
			sb.WriteString(fmt.Sprintf("**状态**: %s\n", getStrField(t, "status")))
			if desc := getStrField(t, "description"); desc != "" {
				sb.WriteString(fmt.Sprintf("**描述**: %s\n", desc))
			}
			if prompt := getStrField(t, "prompt"); prompt != "" {
				sb.WriteString(fmt.Sprintf("**实现提示词**: %s\n", prompt))
			}
			if upstream := t["upstreamContractDetail"]; upstream != nil {
				sb.WriteString(fmt.Sprintf("**上游契约**: %v\n", upstream))
			}
			if downstream := t["downstreamContractDetail"]; downstream != nil {
				sb.WriteString(fmt.Sprintf("**下游契约**: %v\n", downstream))
			}
		}
	}

	// 找到和当前模块有关的上游/下游模块（通过分析任务的上下游契约中的 from 字段）
	upstreamModules, downstreamModules := findRelatedModules(moduleTasks, allTasks, modulePathName, moduleMap)

	if len(upstreamModules) > 0 {
		sb.WriteString("\n## 上游依赖模块信息（参考，排除当前模块）\n")
		for upModPath, upModData := range upstreamModules {
			sb.WriteString(fmt.Sprintf("\n### 上游模块: %s\n", getStrField(upModData, "name")))
			sb.WriteString(fmt.Sprintf("**路径**: %s\n", upModPath))
			// 列出该上游模块的任务摘要
			upTasks := getTasksInModule(allTasks, upModPath)
			for _, ut := range upTasks {
				sb.WriteString(fmt.Sprintf("- 任务: %s [%s]\n", getStrField(ut, "name"), getStrField(ut, "status")))
				if downstream := ut["downstreamContractDetail"]; downstream != nil {
					sb.WriteString(fmt.Sprintf("  下游契约: %v\n", downstream))
				}
			}
		}
	}

	if len(downstreamModules) > 0 {
		sb.WriteString("\n## 下游依赖模块信息（参考，排除当前模块）\n")
		for downModPath, downModData := range downstreamModules {
			sb.WriteString(fmt.Sprintf("\n### 下游模块: %s\n", getStrField(downModData, "name")))
			sb.WriteString(fmt.Sprintf("**路径**: %s\n", downModPath))
			downTasks := getTasksInModule(allTasks, downModPath)
			for _, dt := range downTasks {
				sb.WriteString(fmt.Sprintf("- 任务: %s [%s]\n", getStrField(dt, "name"), getStrField(dt, "status")))
				if upstream := dt["upstreamContractDetail"]; upstream != nil {
					sb.WriteString(fmt.Sprintf("  上游契约: %v\n", upstream))
				}
			}
		}
	}

	// 项目信息
	sb.WriteString("\n## 项目信息（参考）\n")
	sb.WriteString(fmt.Sprintf("**项目名称**: %s\n", getStrField(projectInfo, "name")))
	sb.WriteString(fmt.Sprintf("**项目路径**: %s\n", projectName))

	return sb.String()
}

// ========== AnalysisGenerator 新方法（在此文件中添加以保持功能内聚）==========
// 注意：这些方法实际上添加到 AnalysisGenerator 类型上

// BuildTaskSingleQuestionPrompt 为单个问题构建 task 分析的用户提示词
// taskData: 任务 map 数据
// contextStr: 额外的上下文信息（上游/下游任务、模块信息等）
// question: 当前分析问题
func (g *AnalysisGenerator) BuildTaskSingleQuestionPrompt(
	taskData map[string]interface{},
	contextStr string,
	question Question,
) string {
	var sb strings.Builder

	// 主信息：当前任务所有信息
	taskName, _ := taskData["name"].(string)
	taskPathName, _ := taskData["pathName"].(string)
	taskDescription, _ := taskData["description"].(string)
	taskStatus, _ := taskData["status"].(string)
	taskPrompt, _ := taskData["prompt"].(string)

	sb.WriteString("## 当前任务信息（主信息）\n\n")
	if taskPathName != "" {
		sb.WriteString(fmt.Sprintf("**任务路径**: %s\n", taskPathName))
	}
	if taskName != "" {
		sb.WriteString(fmt.Sprintf("**任务名称**: %s\n", taskName))
	}
	if taskStatus != "" {
		sb.WriteString(fmt.Sprintf("**状态**: %s\n", taskStatus))
	}
	if taskDescription != "" {
		sb.WriteString("\n### 任务描述\n")
		sb.WriteString(taskDescription)
		sb.WriteString("\n")
	}
	if taskPrompt != "" {
		sb.WriteString("\n### 实现提示词（Prompt）\n")
		sb.WriteString(taskPrompt)
		sb.WriteString("\n")
	}

	// tests 字段可能是 string 或 []interface{}
	if tests := taskData["tests"]; tests != nil {
		sb.WriteString("\n### 测试要求\n")
		switch t := tests.(type) {
		case string:
			if t != "" {
				sb.WriteString(t)
				sb.WriteString("\n")
			}
		case []interface{}:
			for _, item := range t {
				if itemMap, ok := item.(map[string]interface{}); ok {
					target, _ := itemMap["target"].(string)
					api, _ := itemMap["api"].(string)
					sb.WriteString(fmt.Sprintf("- %s: %s\n", target, api))
				} else {
					sb.WriteString(fmt.Sprintf("- %v\n", item))
				}
			}
		}
	}

	// 上游契约
	if upstreamContract := taskData["upstreamContractDetail"]; upstreamContract != nil {
		sb.WriteString("\n### 上游接口契约（当前任务自身定义）\n")
		sb.WriteString(fmt.Sprintf("%v\n", upstreamContract))
	}

	// 下游契约
	if downstreamContract := taskData["downstreamContractDetail"]; downstreamContract != nil {
		sb.WriteString("\n### 下游接口契约（当前任务自身定义）\n")
		sb.WriteString(fmt.Sprintf("%v\n", downstreamContract))
	}

	// 代码路径
	if codePaths := taskData["codePaths"]; codePaths != nil {
		sb.WriteString("\n### 代码路径\n")
		sb.WriteString(fmt.Sprintf("%v\n", codePaths))
	}

	// 参考信息：上下文
	if contextStr != "" {
		sb.WriteString("\n---\n")
		sb.WriteString("## 参考信息（上下文）\n")
		sb.WriteString(contextStr)
	}

	// 当前分析问题（单个）
	sb.WriteString("\n---\n")
	sb.WriteString("## 分析问题\n\n")
	sb.WriteString(fmt.Sprintf("**问题 [%s - %s]**: %s\n\n", question.Category, question.Severity, question.Question))
	sb.WriteString("请针对上述单一问题进行分析，按照系统提示词中的JSON格式输出分析结果。\n")

	return sb.String()
}

// BuildModuleSingleQuestionPrompt 为单个问题构建 module 分析的用户提示词
func (g *AnalysisGenerator) BuildModuleSingleQuestionPrompt(
	modData map[string]interface{},
	contextStr string,
	question Question,
) string {
	var sb strings.Builder

	modName, _ := modData["name"].(string)
	modPathName, _ := modData["pathName"].(string)
	modDescription, _ := modData["description"].(string)
	modStatus, _ := modData["status"].(string)
	modPrompt, _ := modData["prompt"].(string)

	sb.WriteString("## 当前模块信息（主信息）\n\n")
	if modPathName != "" {
		sb.WriteString(fmt.Sprintf("**模块路径**: %s\n", modPathName))
	}
	if modName != "" {
		sb.WriteString(fmt.Sprintf("**模块名称**: %s\n", modName))
	}
	if modStatus != "" {
		sb.WriteString(fmt.Sprintf("**状态**: %s\n", modStatus))
	}
	if modDescription != "" {
		sb.WriteString("\n### 模块描述\n")
		sb.WriteString(modDescription)
		sb.WriteString("\n")
	}
	if modPrompt != "" {
		sb.WriteString("\n### 模块提示词\n")
		sb.WriteString(modPrompt)
		sb.WriteString("\n")
	}

	if upstreamSummary, _ := modData["upstreamContractSummary"].(string); upstreamSummary != "" {
		sb.WriteString("\n### 上游契约摘要\n")
		sb.WriteString(upstreamSummary)
		sb.WriteString("\n")
	}
	if downstreamSummary, _ := modData["downstreamContractSummary"].(string); downstreamSummary != "" {
		sb.WriteString("\n### 下游契约摘要\n")
		sb.WriteString(downstreamSummary)
		sb.WriteString("\n")
	}

	// 参考信息
	if contextStr != "" {
		sb.WriteString("\n---\n")
		sb.WriteString("## 参考信息（上下文）\n")
		sb.WriteString(contextStr)
	}

	// 单个问题
	sb.WriteString("\n---\n")
	sb.WriteString("## 分析问题\n\n")
	sb.WriteString(fmt.Sprintf("**问题 [%s - %s]**: %s\n\n", question.Category, question.Severity, question.Question))
	sb.WriteString("请针对上述单一问题进行分析，按照系统提示词中的JSON格式输出分析结果。\n")

	return sb.String()
}

// BuildProjectSingleQuestionPrompt 为单个问题构建 project 分析的用户提示词
func (g *AnalysisGenerator) BuildProjectSingleQuestionPrompt(
	projectInfo map[string]interface{},
	moduleMap map[string]map[string]interface{},
	allTasks []map[string]interface{},
	question Question,
) string {
	var sb strings.Builder

	projectName, _ := projectInfo["name"].(string)
	projectPathName, _ := projectInfo["pathName"].(string)

	sb.WriteString("## 项目信息（主信息）\n\n")
	if projectPathName != "" {
		sb.WriteString(fmt.Sprintf("**项目路径**: %s\n", projectPathName))
	}
	if projectName != "" {
		sb.WriteString(fmt.Sprintf("**项目名称**: %s\n", projectName))
	}

	if constitution, ok := projectInfo["constitution"].(string); ok && constitution != "" {
		sb.WriteString("\n### 项目描述/宪法\n")
		sb.WriteString(constitution)
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("\n**模块总数**: %d\n", len(moduleMap)))
	sb.WriteString(fmt.Sprintf("**任务总数**: %d\n", len(allTasks)))

	// 所有模块信息
	if len(moduleMap) > 0 {
		sb.WriteString("\n## 所有模块信息（参考）\n")
		for modPath, modData := range moduleMap {
			modName, _ := modData["name"].(string)
			sb.WriteString(fmt.Sprintf("\n### 模块: %s\n", modName))
			sb.WriteString(fmt.Sprintf("**路径**: %s\n", modPath))
			if modStatus, ok := modData["status"].(string); ok && modStatus != "" {
				sb.WriteString(fmt.Sprintf("**状态**: %s\n", modStatus))
			}
			if modDesc, ok := modData["description"].(string); ok && modDesc != "" {
				sb.WriteString(fmt.Sprintf("**描述**: %s\n", modDesc))
			}
		}
	}

	// 所有任务摘要（精简版：名称、描述、状态、buglog、关键契约）
	if len(allTasks) > 0 {
		sb.WriteString("\n## 项目任务摘要（参考）\n")
		for _, t := range allTasks {
			sb.WriteString(fmt.Sprintf("\n- **%s** [%s]", getStrField(t, "name"), getStrField(t, "status")))
			if desc := getStrField(t, "description"); desc != "" {
				sb.WriteString(fmt.Sprintf(": %s", desc))
			}
			sb.WriteString("\n")
			if bugLog := getStrField(t, "bugLog"); bugLog != "" {
				sb.WriteString(fmt.Sprintf("  BugLog: %s\n", bugLog))
			}
			if downstream := t["downstreamContractDetail"]; downstream != nil {
				// 只输出函数名摘要，不输出完整契约，避免 prompt 过长
				if downstreamMap, ok := downstream.(map[string]interface{}); ok {
					if list, ok := downstreamMap["list"].([]interface{}); ok && len(list) > 0 {
						sb.WriteString("  下游契约（摘要）: ")
						labels := []string{}
						for _, item := range list {
							if itemMap, ok := item.(map[string]interface{}); ok {
								if label, ok := itemMap["label"].(string); ok {
									labels = append(labels, label)
								}
							}
						}
						sb.WriteString(strings.Join(labels, ", "))
						sb.WriteString("\n")
					}
				}
			}
		}
	}

	// 单个问题
	sb.WriteString("\n---\n")
	sb.WriteString("## 分析问题\n\n")
	sb.WriteString(fmt.Sprintf("**问题 [%s - %s]**: %s\n\n", question.Category, question.Severity, question.Question))
	sb.WriteString("请针对上述单一问题进行分析，按照系统提示词中的JSON格式输出分析结果。\n")

	return sb.String()
}

// ========== 辅助查找函数 ==========

// findUpstreamTasks 根据 upstreamContractDetail.list[].from 字段找到上游任务
func findUpstreamTasks(taskData map[string]interface{}, allTasks []map[string]interface{}, currentTaskName string) []map[string]interface{} {
	var result []map[string]interface{}
	seen := make(map[string]bool)

	upstreamDetail := taskData["upstreamContractDetail"]
	if upstreamDetail == nil {
		return result
	}

	// 提取 from 列表
	fromPaths := extractFromPaths(upstreamDetail)
	if len(fromPaths) == 0 {
		return result
	}

	for _, fromPath := range fromPaths {
		if seen[fromPath] {
			continue
		}
		seen[fromPath] = true

		for _, t := range allTasks {
			tPathName, _ := t["pathName"].(string)
			if tPathName == fromPath {
				result = append(result, t)
				break
			}
		}
	}

	return result
}

// findDownstreamTasks 根据 downstreamContractDetail.list[].from 字段找到下游任务
func findDownstreamTasks(taskData map[string]interface{}, allTasks []map[string]interface{}, currentTaskName string) []map[string]interface{} {
	var result []map[string]interface{}
	seen := make(map[string]bool)

	downstreamDetail := taskData["downstreamContractDetail"]
	if downstreamDetail == nil {
		return result
	}

	fromPaths := extractFromPaths(downstreamDetail)
	if len(fromPaths) == 0 {
		return result
	}

	for _, fromPath := range fromPaths {
		if seen[fromPath] {
			continue
		}
		seen[fromPath] = true

		for _, t := range allTasks {
			tPathName, _ := t["pathName"].(string)
			if tPathName == fromPath {
				result = append(result, t)
				break
			}
		}
	}

	return result
}

// extractFromPaths 从 contractDetail 中提取 list[].from 字段值
// contractDetail 可能是 map[string]interface{} 或 JSON string
func extractFromPaths(contractDetail interface{}) []string {
	var paths []string

	var detailMap map[string]interface{}

	switch v := contractDetail.(type) {
	case map[string]interface{}:
		detailMap = v
	case string:
		if v == "" {
			return paths
		}
		if err := json.Unmarshal([]byte(v), &detailMap); err != nil {
			return paths
		}
	default:
		return paths
	}

	list, ok := detailMap["list"].([]interface{})
	if !ok {
		return paths
	}

	for _, item := range list {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if from, ok := itemMap["from"].(string); ok && from != "" {
				paths = append(paths, from)
			}
		}
	}

	return paths
}

// getTasksInModule 获取属于某模块的所有任务
func getTasksInModule(allTasks []map[string]interface{}, modulePathName string) []map[string]interface{} {
	var result []map[string]interface{}
	for _, t := range allTasks {
		tPathName, _ := t["pathName"].(string)
		if tPathName == "" {
			continue
		}
		parts := strings.Split(tPathName, "/")
		if len(parts) < 2 {
			continue
		}
		// 检查任务的模块路径是否匹配
		taskModulePath := strings.Join(parts[:len(parts)-1], "/")
		if taskModulePath == modulePathName {
			result = append(result, t)
		}
	}
	return result
}

// findRelatedModules 根据当前模块任务的上下游契约，找到相关的上游模块和下游模块
// 排除当前模块自身
func findRelatedModules(
	moduleTasks []map[string]interface{},
	allTasks []map[string]interface{},
	currentModulePath string,
	moduleMap map[string]map[string]interface{},
) (upstreamModules map[string]map[string]interface{}, downstreamModules map[string]map[string]interface{}) {
	upstreamModules = make(map[string]map[string]interface{})
	downstreamModules = make(map[string]map[string]interface{})

	for _, t := range moduleTasks {
		// 上游契约 -> 找上游任务 -> 找上游模块
		upstreamTasks := findUpstreamTasks(t, allTasks, "")
		for _, ut := range upstreamTasks {
			utPath, _ := ut["pathName"].(string)
			if utPath == "" {
				continue
			}
			parts := strings.Split(utPath, "/")
			if len(parts) < 2 {
				continue
			}
			upModPath := strings.Join(parts[:len(parts)-1], "/")
			if upModPath != currentModulePath {
				if modData, ok := moduleMap[upModPath]; ok {
					upstreamModules[upModPath] = modData
				}
			}
		}

		// 下游契约 -> 找下游任务 -> 找下游模块
		downstreamTasks := findDownstreamTasks(t, allTasks, "")
		for _, dt := range downstreamTasks {
			dtPath, _ := dt["pathName"].(string)
			if dtPath == "" {
				continue
			}
			parts := strings.Split(dtPath, "/")
			if len(parts) < 2 {
				continue
			}
			downModPath := strings.Join(parts[:len(parts)-1], "/")
			if downModPath != currentModulePath {
				if modData, ok := moduleMap[downModPath]; ok {
					downstreamModules[downModPath] = modData
				}
			}
		}
	}

	return
}

// getStrField 安全地从 map 中提取字符串字段
func getStrField(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, _ := m[key].(string)
	return v
}

// ========== 报表格式化 ==========

// formatNewCompileAIReport 格式化新版AI分析报告
// 按目标（task/module/project）聚合每个问题的分析结果，计算综合评分
func formatNewCompileAIReport(
	projectName string,
	tasks []map[string]interface{},
	moduleMap map[string]map[string]interface{},
	projectInfo map[string]interface{},
	taskQuestions []Question,
	moduleQuestions []Question,
	projectQuestions []Question,
	allResults []*singleQuestionResult,
	analyzeErrors []string,
	maxConcurrent int,
) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("## AI设计合理性分析报告 - %s\n\n", projectName))
	output.WriteString(fmt.Sprintf("**并发配置**: 最大并发数 = %d  |  任务总数 = %d  |  模块总数 = %d\n\n",
		maxConcurrent, len(tasks), len(moduleMap)))

	// 按目标聚合结果
	// key: targetPathName, value: 按 questionIndex 存储的结果
	type targetKey struct {
		analysisType   string
		targetPathName string
	}
	resultsByTarget := make(map[targetKey]map[int]*singleQuestionResult)

	for _, r := range allResults {
		if r.job == nil {
			continue
		}
		tk := targetKey{r.job.analysisType, r.job.targetPathName}
		if _, ok := resultsByTarget[tk]; !ok {
			resultsByTarget[tk] = make(map[int]*singleQuestionResult)
		}
		resultsByTarget[tk][r.job.questionIndex] = r
	}

	// ========== 任务分析结果 ==========
	if len(taskQuestions) > 0 {
		output.WriteString("---\n\n")
		output.WriteString("# 任务分析结果\n\n")

		for _, taskData := range tasks {
			taskPathName, _ := taskData["pathName"].(string)
			taskName, _ := taskData["name"].(string)
			if taskPathName == "" {
				taskPathName = taskName
			}

			tk := targetKey{QuestionTypeTask, taskPathName}
			qResults := resultsByTarget[tk]

			// 计算综合评分
			avgScore := calcAvgScore(qResults)
			output.WriteString(fmt.Sprintf("### 任务: %s   (综合评分: %.0f/100)\n\n", taskPathName, avgScore))

			// 逐问题展示
			for qi, q := range taskQuestions {
				qr, ok := qResults[qi]
				score := 0
				if ok && qr.result != nil {
					score = qr.result.Score
				}
				output.WriteString(fmt.Sprintf("**[问题%d - %s]** 质量: %d/100\n", qi+1, q.Category, score))
				output.WriteString(fmt.Sprintf("%s\n\n", q.Question))

				if ok && qr.err != "" {
					output.WriteString(fmt.Sprintf("*分析失败: %s*\n\n", qr.err))
					continue
				}

				if ok && qr.result != nil {
					r := qr.result
					if r.Summary != "" {
						output.WriteString(fmt.Sprintf("**分析摘要**: %s\n\n", r.Summary))
					}
					if len(r.Issues) > 0 {
						output.WriteString("**发现的问题：**\n")
						for _, issue := range r.Issues {
							output.WriteString(fmt.Sprintf("- [%s] %s\n", issue.Severity, issue.Description))
							if issue.Suggestion != "" {
								output.WriteString(fmt.Sprintf("  → 建议: %s\n", issue.Suggestion))
							}
						}
						output.WriteString("\n")
					}
					if len(r.Suggestions) > 0 {
						output.WriteString("**优化建议：**\n")
						for _, s := range r.Suggestions {
							output.WriteString(fmt.Sprintf("- %s\n", s))
						}
						output.WriteString("\n")
					}
				}
			}
			output.WriteString("---\n\n")
		}
	}

	// ========== 模块分析结果 ==========
	if len(moduleQuestions) > 0 {
		output.WriteString("# 模块分析结果\n\n")

		for modPath, modData := range moduleMap {
			modName, _ := modData["name"].(string)

			tk := targetKey{QuestionTypeModule, modPath}
			qResults := resultsByTarget[tk]
			avgScore := calcAvgScore(qResults)

			output.WriteString(fmt.Sprintf("### 模块: %s  (综合评分: %.0f/100)\n\n", modPath, avgScore))
			_ = modName

			for qi, q := range moduleQuestions {
				qr, ok := qResults[qi]
				score := 0
				if ok && qr.result != nil {
					score = qr.result.Score
				}
				output.WriteString(fmt.Sprintf("**[问题%d - %s]** 质量: %d/100\n", qi+1, q.Category, score))
				output.WriteString(fmt.Sprintf("%s\n\n", q.Question))

				if ok && qr.err != "" {
					output.WriteString(fmt.Sprintf("*分析失败: %s*\n\n", qr.err))
					continue
				}

				if ok && qr.result != nil {
					r := qr.result
					if r.Summary != "" {
						output.WriteString(fmt.Sprintf("**分析摘要**: %s\n\n", r.Summary))
					}
					if len(r.Issues) > 0 {
						output.WriteString("**发现的问题：**\n")
						for _, issue := range r.Issues {
							output.WriteString(fmt.Sprintf("- [%s] %s\n", issue.Severity, issue.Description))
							if issue.Suggestion != "" {
								output.WriteString(fmt.Sprintf("  → 建议: %s\n", issue.Suggestion))
							}
						}
						output.WriteString("\n")
					}
					if len(r.Suggestions) > 0 {
						output.WriteString("**优化建议：**\n")
						for _, s := range r.Suggestions {
							output.WriteString(fmt.Sprintf("- %s\n", s))
						}
						output.WriteString("\n")
					}
				}
			}
			output.WriteString("---\n\n")
		}
	}

	// ========== 项目分析结果 ==========
	if len(projectQuestions) > 0 {
		output.WriteString("# 项目分析结果\n\n")

		tk := targetKey{QuestionTypeProject, projectName}
		qResults := resultsByTarget[tk]
		avgScore := calcAvgScore(qResults)

		output.WriteString(fmt.Sprintf("### 项目: %s  (综合评分: %.0f/100)\n\n", projectName, avgScore))

		for qi, q := range projectQuestions {
			qr, ok := qResults[qi]
			score := 0
			if ok && qr.result != nil {
				score = qr.result.Score
			}
			output.WriteString(fmt.Sprintf("**[问题%d - %s]** 质量: %d/100\n", qi+1, q.Category, score))
			output.WriteString(fmt.Sprintf("%s\n\n", q.Question))

			if ok && qr.err != "" {
				output.WriteString(fmt.Sprintf("*分析失败: %s*\n\n", qr.err))
				continue
			}

			if ok && qr.result != nil {
				r := qr.result
				if r.Summary != "" {
					output.WriteString(fmt.Sprintf("**分析摘要**: %s\n\n", r.Summary))
				}
				if len(r.Issues) > 0 {
					output.WriteString("**发现的问题：**\n")
					for _, issue := range r.Issues {
						output.WriteString(fmt.Sprintf("- [%s] %s\n", issue.Severity, issue.Description))
						if issue.Suggestion != "" {
							output.WriteString(fmt.Sprintf("  → 建议: %s\n", issue.Suggestion))
						}
					}
					output.WriteString("\n")
				}
				if len(r.Suggestions) > 0 {
					output.WriteString("**优化建议：**\n")
					for _, s := range r.Suggestions {
						output.WriteString(fmt.Sprintf("- %s\n", s))
					}
					output.WriteString("\n")
				}
			}
		}
		output.WriteString("---\n\n")
	}

	// 错误汇总
	if len(analyzeErrors) > 0 {
		output.WriteString("### 分析错误\n")
		for _, errMsg := range analyzeErrors {
			output.WriteString(fmt.Sprintf("- %s\n", errMsg))
		}
		output.WriteString("\n")
	}

	return output.String()
}

// calcAvgScore 计算一组问题结果的平均分
func calcAvgScore(qResults map[int]*singleQuestionResult) float64 {
	if len(qResults) == 0 {
		return 0
	}
	total := 0
	count := 0
	for _, qr := range qResults {
		if qr != nil && qr.result != nil {
			total += qr.result.Score
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return float64(total) / float64(count)
}

// ==================== 保留的辅助函数 ====================

// parseSubAgentConfig 从原始输入解析SubAgentConfig
func parseSubAgentConfig(raw interface{}) (*SubAgentConfig, error) {
	var jsonBytes []byte
	var err error

	switch v := raw.(type) {
	case string:
		jsonBytes = []byte(v)
	case map[string]interface{}:
		jsonBytes, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("不支持的sub_agent_config格式: %T", raw)
	}

	config := &SubAgentConfig{}
	if err := json.Unmarshal(jsonBytes, config); err != nil {
		return nil, fmt.Errorf("解析sub_agent_config JSON失败: %w", err)
	}

	return config, nil
}

// parseQAConfig 从原始输入解析QAConfig
// 验证规则：qa_config 必须包含 "questions" 字段（新格式），否则返回明确错误指引 AI 重新读取文件
func parseQAConfig(raw interface{}) (*QAConfig, error) {
	var jsonBytes []byte
	var err error

	switch v := raw.(type) {
	case string:
		jsonBytes = []byte(v)
	case map[string]interface{}:
		// 在序列化前检查 "questions" 字段是否存在（支持新格式校验）
		if _, hasQuestions := v["questions"]; !hasQuestions {
			// 同时检查旧格式兼容字段
			_, hasTaskQ := v["task_questions"]
			_, hasModuleQ := v["module_questions"]
			_, hasProjectQ := v["project_questions"]
			if !hasTaskQ && !hasModuleQ && !hasProjectQ {
				return nil, fmt.Errorf(
					"qa_config 参数格式无效：缺少必要的 \"questions\" 字段。\n" +
						"请按以下步骤重试：\n" +
						"1. 使用 read_file 工具读取项目根目录下的 .aitdd/qa.json 文件内容\n" +
						"2. 将读取到的完整 JSON 内容作为 qa_config 参数重新调用 compile_dynamic 工具\n" +
						"正确的 qa_config 格式示例：{ \"questions\": { \"task\": [...], \"module\": [...], \"project\": [...] } }")
			}
		}
		jsonBytes, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("不支持的qa_config格式: %T", raw)
	}

	config := &QAConfig{}
	if err := json.Unmarshal(jsonBytes, config); err != nil {
		return nil, fmt.Errorf("解析qa_config JSON失败: %w", err)
	}

	// 验证解析后的结构：新格式必须有 Questions 字段，旧格式允许使用扁平字段
	if config.Questions == nil && len(config.TaskQuestions) == 0 && len(config.ModuleQuestions) == 0 && len(config.ProjectQuestions) == 0 {
		return nil, fmt.Errorf(
			"qa_config 内容无效：\"questions\" 字段存在但为空，或缺少问题定义。\n" +
				"请按以下步骤重试：\n" +
				"1. 使用 read_file 工具读取项目根目录下的 .aitdd/qa.json 文件内容\n" +
				"2. 将读取到的完整 JSON 内容作为 qa_config 参数重新调用 compile_dynamic 工具\n" +
				"正确的 qa_config 格式示例：{ \"questions\": { \"task\": [...], \"module\": [...], \"project\": [...] } }")
	}

	return config, nil
}

// analysisTypeLabel 获取分析类型的中文标签
func analysisTypeLabel(analysisType string) string {
	switch analysisType {
	case QuestionTypeTask:
		return "任务"
	case QuestionTypeModule:
		return "模块"
	case QuestionTypeProject:
		return "项目"
	default:
		return analysisType
	}
}

// filterIssuesBySeverity 按严重程度过滤问题
func filterIssuesBySeverity(issues []Issue, severity string) []Issue {
	var filtered []Issue
	for _, issue := range issues {
		if issue.Severity == severity {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

// ==================== 辅助函数（保留供其他文件使用）====================

// parseIDList 解析ID列表参数
// 支持 []interface{}、[]float64、[]int 等格式
func parseIDList(raw interface{}) ([]uint, error) {
	if raw == nil {
		return nil, fmt.Errorf("ID列表为nil")
	}

	var ids []uint

	switch v := raw.(type) {
	case []interface{}:
		for _, item := range v {
			switch id := item.(type) {
			case float64:
				ids = append(ids, uint(id))
			case int:
				ids = append(ids, uint(id))
			case int64:
				ids = append(ids, uint(id))
			default:
				return nil, fmt.Errorf("无效的ID类型: %T", item)
			}
		}
	case []float64:
		for _, id := range v {
			ids = append(ids, uint(id))
		}
	case []int:
		for _, id := range v {
			ids = append(ids, uint(id))
		}
	default:
		return nil, fmt.Errorf("无效的ID列表类型: %T", raw)
	}

	return ids, nil
}
