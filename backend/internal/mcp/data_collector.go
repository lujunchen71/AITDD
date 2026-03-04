package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DataCollector 数据收集器
// 用于从REST API收集任务、模块和项目的分析数据
type DataCollector struct {
	apiBaseURL string
	httpClient *http.Client
	logger     *MCPLogger
}

// TaskAnalysisData 任务分析数据
// 包含任务的完整信息，用于AI分析
type TaskAnalysisData struct {
	TaskID              uint                   `json:"task_id"`
	TaskName            string                 `json:"task_name"`
	TaskDescription     string                 `json:"task_description"`
	Status              string                 `json:"status"`
	Priority            int                    `json:"priority"`
	ModuleID            *uint                  `json:"module_id,omitempty"`
	ModuleName          string                 `json:"module_name,omitempty"`
	ProjectID           uint                   `json:"project_id"`
	ProjectName         string                 `json:"project_name"`
	Dependencies        []DependencyInfo       `json:"dependencies"`
	CodePaths           []string               `json:"code_paths"`
	AcceptanceCriteria  string                 `json:"acceptance_criteria"`
	TestRequirements    string                 `json:"test_requirements"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
	AdditionalInfo      map[string]interface{} `json:"additional_info,omitempty"`
}

// ModuleAnalysisData 模块分析数据
// 包含模块的完整信息及其关联任务
type ModuleAnalysisData struct {
	ModuleID            uint                   `json:"module_id"`
	ModuleName          string                 `json:"module_name"`
	ModuleDescription   string                 `json:"module_description"`
	ProjectID           uint                   `json:"project_id"`
	ProjectName         string                 `json:"project_name"`
	Tasks               []TaskSummary          `json:"tasks"`
	Dependencies        []DependencyInfo       `json:"dependencies"`
	CodePaths           []string               `json:"code_paths"`
	Status              string                 `json:"status"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
	AdditionalInfo      map[string]interface{} `json:"additional_info,omitempty"`
}

// ProjectAnalysisData 项目分析数据
// 包含项目的完整信息及其关联模块和任务
type ProjectAnalysisData struct {
	ProjectID           uint                   `json:"project_id"`
	ProjectName         string                 `json:"project_name"`
	ProjectDescription  string                 `json:"project_description"`
	Modules             []ModuleSummary        `json:"modules"`
	Tasks               []TaskSummary          `json:"tasks"`
	TotalModules        int                    `json:"total_modules"`
	TotalTasks          int                    `json:"total_tasks"`
	Constitution        string                 `json:"constitution"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
	AdditionalInfo      map[string]interface{} `json:"additional_info,omitempty"`
}

// TaskSummary 任务摘要信息
type TaskSummary struct {
	TaskID          uint   `json:"task_id"`
	TaskName        string `json:"task_name"`
	Status          string `json:"status"`
	Priority        int    `json:"priority"`
	ModuleID        *uint  `json:"module_id,omitempty"`
	ModuleName      string `json:"module_name,omitempty"`
}

// ModuleSummary 模块摘要信息
type ModuleSummary struct {
	ModuleID          uint     `json:"module_id"`
	ModuleName        string   `json:"module_name"`
	Status            string   `json:"status"`
	TaskCount         int      `json:"task_count"`
	Dependencies      []string `json:"dependencies,omitempty"`
}

// DependencyInfo 依赖信息
type DependencyInfo struct {
	DependencyID      uint   `json:"dependency_id"`
	UpstreamTaskID    *uint  `json:"upstream_task_id,omitempty"`
	UpstreamTaskName  string `json:"upstream_task_name,omitempty"`
	DownstreamTaskID  *uint  `json:"downstream_task_id,omitempty"`
	DownstreamTaskName string `json:"downstream_task_name,omitempty"`
	ContractSummary   string `json:"contract_summary,omitempty"`
	Type              string `json:"type"` // task, module
}

// APIResponse 通用API响应结构
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}

// TaskAPIResponse 任务API响应结构
type TaskAPIResponse struct {
	ID                     string  `json:"id"`
	ModuleID               string  `json:"moduleId"`
	Name                   string  `json:"name"`
	PathName               string  `json:"pathName"`
	Description            string  `json:"description"`
	Status                 string  `json:"status"`
	Assignee               *string `json:"assignee"`
	UpstreamContractDetail string  `json:"upstreamContractDetail"`
	DownstreamContractDetail string `json:"downstreamContractDetail"`
	Prompt                 string  `json:"prompt"`
	Tests                  string  `json:"tests"`
	TestResult             string  `json:"testResult"`
	CodePaths              string  `json:"codePaths"`
	CreatedAt              int64   `json:"createdAt"`
	UpdatedAt              int64   `json:"updatedAt"`
}

// ModuleAPIResponse 模块API响应结构
type ModuleAPIResponse struct {
	ID                       string   `json:"id"`
	ParentID                 *string  `json:"parentId"`
	ProjectID                string   `json:"projectId"`
	Name                     string   `json:"name"`
	PathName                 string   `json:"pathName"`
	Description              string   `json:"description"`
	Prompt                   string   `json:"prompt"`
	Status                   string   `json:"status"`
	TestCoverage             float64  `json:"testCoverage"`
	UpstreamContractSummary  string   `json:"upstreamContractSummary"`
	DownstreamContractSummary string   `json:"downstreamContractSummary"`
	CreatedAt                int64    `json:"createdAt"`
	UpdatedAt                int64    `json:"updatedAt"`
}

// ProjectAPIResponse 项目API响应结构
type ProjectAPIResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	PathName     string `json:"pathName"`
	Constitution string `json:"constitution"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
}

// NewDataCollector 创建新的数据收集器
// apiBaseURL 是REST API的基础URL，例如 "http://localhost:8080/api/v1"
func NewDataCollector(apiBaseURL string) *DataCollector {
	return &DataCollector{
		apiBaseURL: apiBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: GetMCPLogger(),
	}
}

// NewDataCollectorWithLogger 创建带自定义日志的数据收集器
func NewDataCollectorWithLogger(apiBaseURL string, logger *MCPLogger) *DataCollector {
	return &DataCollector{
		apiBaseURL: apiBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// CollectTaskData 收集任务分析数据
// 从API获取任务的完整信息，包括依赖关系和关联的模块/项目信息
func (dc *DataCollector) CollectTaskData(ctx context.Context, taskID uint) (*TaskAnalysisData, error) {
	dc.logger.Debug("开始收集任务数据", map[string]interface{}{"task_id": taskID})

	// 获取任务基本信息
	taskResp, err := dc.get(ctx, fmt.Sprintf("/tasks/%d", taskID))
	if err != nil {
		return nil, NewNetworkErrorf(err, "获取任务 %d 失败", taskID)
	}

	var taskData TaskAPIResponse
	if err := dc.parseResponse(taskResp, &taskData); err != nil {
		return nil, NewInternalError("解析任务数据失败", err)
	}

	// 构建分析数据
	analysisData := &TaskAnalysisData{
		TaskID:             taskID,
		TaskName:           taskData.Name,
		TaskDescription:    taskData.Description,
		Status:             taskData.Status,
		CodePaths:          parseJSONArray(taskData.CodePaths),
		AcceptanceCriteria: taskData.DownstreamContractDetail,
		TestRequirements:   taskData.Tests,
		CreatedAt:          time.UnixMilli(taskData.CreatedAt),
		UpdatedAt:          time.UnixMilli(taskData.UpdatedAt),
		AdditionalInfo: map[string]interface{}{
			"prompt":                 taskData.Prompt,
			"upstreamContractDetail": taskData.UpstreamContractDetail,
			"testResult":             taskData.TestResult,
			"assignee":               taskData.Assignee,
		},
	}

	// 获取模块信息
	if taskData.ModuleID != "" {
		moduleID := parseUint(taskData.ModuleID)
		if moduleID > 0 {
			analysisData.ModuleID = &moduleID
			moduleData, err := dc.collectModuleInfo(ctx, moduleID)
			if err == nil {
				analysisData.ModuleName = moduleData.Name
				analysisData.ProjectID = parseUint(moduleData.ProjectID)
				analysisData.ProjectName = "" // 将在获取项目信息后填充
			}
		}
	}

	// 获取项目信息
	if analysisData.ProjectID > 0 {
		projectData, err := dc.collectProjectInfo(ctx, analysisData.ProjectID)
		if err == nil {
			analysisData.ProjectName = projectData.Name
		}
	}

	// 获取任务依赖
	dependencies, err := dc.collectTaskDependencies(ctx, taskID)
	if err == nil {
		analysisData.Dependencies = dependencies
	}

	dc.logger.Info("任务数据收集完成", map[string]interface{}{
		"task_id":       taskID,
		"task_name":     analysisData.TaskName,
		"dependencies":  len(analysisData.Dependencies),
	})

	return analysisData, nil
}

// CollectModuleData 收集模块分析数据
// 从API获取模块的完整信息，包括关联的任务和项目信息
func (dc *DataCollector) CollectModuleData(ctx context.Context, moduleID uint) (*ModuleAnalysisData, error) {
	dc.logger.Debug("开始收集模块数据", map[string]interface{}{"module_id": moduleID})

	// 获取模块基本信息
	moduleResp, err := dc.get(ctx, fmt.Sprintf("/modules/%d", moduleID))
	if err != nil {
		return nil, NewNetworkErrorf(err, "获取模块 %d 失败", moduleID)
	}

	var moduleData ModuleAPIResponse
	if err := dc.parseResponse(moduleResp, &moduleData); err != nil {
		return nil, NewInternalError("解析模块数据失败", err)
	}

	// 构建分析数据
	analysisData := &ModuleAnalysisData{
		ModuleID:          moduleID,
		ModuleName:        moduleData.Name,
		ModuleDescription: moduleData.Description,
		Status:            moduleData.Status,
		CreatedAt:         time.UnixMilli(moduleData.CreatedAt),
		UpdatedAt:         time.UnixMilli(moduleData.UpdatedAt),
		AdditionalInfo: map[string]interface{}{
			"prompt":                    moduleData.Prompt,
			"testCoverage":              moduleData.TestCoverage,
			"upstreamContractSummary":   moduleData.UpstreamContractSummary,
			"downstreamContractSummary": moduleData.DownstreamContractSummary,
		},
	}

	// 获取项目信息
	analysisData.ProjectID = parseUint(moduleData.ProjectID)
	if analysisData.ProjectID > 0 {
		projectData, err := dc.collectProjectInfo(ctx, analysisData.ProjectID)
		if err == nil {
			analysisData.ProjectName = projectData.Name
		}
	}

	// 获取模块的任务
	tasks, err := dc.collectModuleTasks(ctx, moduleID)
	if err == nil {
		analysisData.Tasks = tasks
	}

	// 获取模块依赖
	dependencies, err := dc.collectModuleDependencies(ctx, moduleID)
	if err == nil {
		analysisData.Dependencies = dependencies
	}

	dc.logger.Info("模块数据收集完成", map[string]interface{}{
		"module_id":     moduleID,
		"module_name":   analysisData.ModuleName,
		"tasks":         len(analysisData.Tasks),
		"dependencies":  len(analysisData.Dependencies),
	})

	return analysisData, nil
}

// CollectProjectData 收集项目分析数据
// 从API获取项目的完整信息，包括所有模块和任务
func (dc *DataCollector) CollectProjectData(ctx context.Context, projectID uint) (*ProjectAnalysisData, error) {
	dc.logger.Debug("开始收集项目数据", map[string]interface{}{"project_id": projectID})

	// 获取项目基本信息
	projectResp, err := dc.get(ctx, fmt.Sprintf("/projects/%d", projectID))
	if err != nil {
		return nil, NewNetworkErrorf(err, "获取项目 %d 失败", projectID)
	}

	var projectData ProjectAPIResponse
	if err := dc.parseResponse(projectResp, &projectData); err != nil {
		return nil, NewInternalError("解析项目数据失败", err)
	}

	// 构建分析数据
	analysisData := &ProjectAnalysisData{
		ProjectID:          projectID,
		ProjectName:        projectData.Name,
		ProjectDescription: projectData.Constitution,
		Constitution:       projectData.Constitution,
		CreatedAt:          time.UnixMilli(projectData.CreatedAt),
		UpdatedAt:          time.UnixMilli(projectData.UpdatedAt),
		Modules:            make([]ModuleSummary, 0),
		Tasks:              make([]TaskSummary, 0),
	}

	// 获取项目的所有模块
	modules, err := dc.collectProjectModules(ctx, projectID)
	if err == nil {
		analysisData.Modules = modules
		analysisData.TotalModules = len(modules)
	}

	// 获取项目的所有任务
	tasks, err := dc.collectProjectTasks(ctx, projectID)
	if err == nil {
		analysisData.Tasks = tasks
		analysisData.TotalTasks = len(tasks)
	}

	dc.logger.Info("项目数据收集完成", map[string]interface{}{
		"project_id":    projectID,
		"project_name":  analysisData.ProjectName,
		"total_modules": analysisData.TotalModules,
		"total_tasks":   analysisData.TotalTasks,
	})

	return analysisData, nil
}

// get 执行GET请求
func (dc *DataCollector) get(ctx context.Context, path string) ([]byte, error) {
	url := dc.apiBaseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	startTime := time.Now()
	resp, err := dc.httpClient.Do(req)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		LogAPICall("GET", url, 0, duration)
		return nil, err
	}
	defer resp.Body.Close()

	LogAPICall("GET", url, resp.StatusCode, duration)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API请求失败: %d - %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// parseResponse 解析API响应
func (dc *DataCollector) parseResponse(data []byte, target interface{}) error {
	// 首先尝试解析为通用响应
	var apiResp APIResponse
	if err := json.Unmarshal(data, &apiResp); err == nil && apiResp.Data != nil {
		// 如果是包装响应，解析data字段
		dataBytes, err := json.Marshal(apiResp.Data)
		if err != nil {
			return err
		}
		return json.Unmarshal(dataBytes, target)
	}

	// 直接解析为目标类型
	return json.Unmarshal(data, target)
}

// collectModuleInfo 获取模块信息
func (dc *DataCollector) collectModuleInfo(ctx context.Context, moduleID uint) (*ModuleAPIResponse, error) {
	resp, err := dc.get(ctx, fmt.Sprintf("/modules/%d", moduleID))
	if err != nil {
		return nil, err
	}

	var moduleData ModuleAPIResponse
	if err := dc.parseResponse(resp, &moduleData); err != nil {
		return nil, err
	}

	return &moduleData, nil
}

// collectProjectInfo 获取项目信息
func (dc *DataCollector) collectProjectInfo(ctx context.Context, projectID uint) (*ProjectAPIResponse, error) {
	resp, err := dc.get(ctx, fmt.Sprintf("/projects/%d", projectID))
	if err != nil {
		return nil, err
	}

	var projectData ProjectAPIResponse
	if err := dc.parseResponse(resp, &projectData); err != nil {
		return nil, err
	}

	return &projectData, nil
}

// collectTaskDependencies 获取任务依赖
func (dc *DataCollector) collectTaskDependencies(ctx context.Context, taskID uint) ([]DependencyInfo, error) {
	// 获取下游依赖（当前任务依赖的其他任务）
	resp, err := dc.get(ctx, fmt.Sprintf("/tasks/%d/dependencies", taskID))
	if err != nil {
		dc.logger.Debug("获取任务依赖失败", map[string]interface{}{
			"task_id": taskID,
			"error":   err.Error(),
		})
		return nil, err
	}

	var dependencies []DependencyInfo
	var depsData []struct {
		ID               string `json:"id"`
		UpstreamTaskID   string `json:"upstreamTaskId"`
		DownstreamTaskID string `json:"downstreamTaskId"`
		ContractSummary  string `json:"contractSummary"`
	}

	if err := dc.parseResponse(resp, &depsData); err != nil {
		return nil, err
	}

	for _, dep := range depsData {
		depInfo := DependencyInfo{
			DependencyID:    parseUint(dep.ID),
			ContractSummary: dep.ContractSummary,
			Type:            "task",
		}
		if dep.UpstreamTaskID != "" {
			upstreamID := parseUint(dep.UpstreamTaskID)
			depInfo.UpstreamTaskID = &upstreamID
		}
		if dep.DownstreamTaskID != "" {
			downstreamID := parseUint(dep.DownstreamTaskID)
			depInfo.DownstreamTaskID = &downstreamID
		}
		dependencies = append(dependencies, depInfo)
	}

	return dependencies, nil
}

// collectModuleTasks 获取模块的任务列表
func (dc *DataCollector) collectModuleTasks(ctx context.Context, moduleID uint) ([]TaskSummary, error) {
	resp, err := dc.get(ctx, fmt.Sprintf("/modules/%d/tasks", moduleID))
	if err != nil {
		return nil, err
	}

	var tasks []TaskSummary
	var tasksData []TaskAPIResponse

	if err := dc.parseResponse(resp, &tasksData); err != nil {
		return nil, err
	}

	for _, t := range tasksData {
		task := TaskSummary{
			TaskID:     parseUint(t.ID),
			TaskName:   t.Name,
			Status:     t.Status,
			ModuleName: t.Name,
		}
		if t.ModuleID != "" {
			moduleID := parseUint(t.ModuleID)
			task.ModuleID = &moduleID
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// collectModuleDependencies 获取模块依赖
func (dc *DataCollector) collectModuleDependencies(ctx context.Context, moduleID uint) ([]DependencyInfo, error) {
	// 获取模块的下游依赖
	resp, err := dc.get(ctx, fmt.Sprintf("/modules/%d/dependencies", moduleID))
	if err != nil {
		dc.logger.Debug("获取模块依赖失败", map[string]interface{}{
			"module_id": moduleID,
			"error":     err.Error(),
		})
		return nil, err
	}

	var dependencies []DependencyInfo
	var depsData []struct {
		ID             string `json:"id"`
		SourceModuleID string `json:"sourceModuleId"`
		TargetModuleID string `json:"targetModuleId"`
		DependencyType string `json:"dependencyType"`
	}

	if err := dc.parseResponse(resp, &depsData); err != nil {
		return nil, err
	}

	for _, dep := range depsData {
		depInfo := DependencyInfo{
			DependencyID: parseUint(dep.ID),
			Type:         "module",
		}
		dependencies = append(dependencies, depInfo)
	}

	return dependencies, nil
}

// collectProjectModules 获取项目的模块列表
func (dc *DataCollector) collectProjectModules(ctx context.Context, projectID uint) ([]ModuleSummary, error) {
	resp, err := dc.get(ctx, "/modules")
	if err != nil {
		return nil, err
	}

	var modules []ModuleSummary
	var modulesData []ModuleAPIResponse

	if err := dc.parseResponse(resp, &modulesData); err != nil {
		return nil, err
	}

	for _, m := range modulesData {
		if parseUint(m.ProjectID) == projectID {
			module := ModuleSummary{
				ModuleID:   parseUint(m.ID),
				ModuleName: m.Name,
				Status:     m.Status,
			}
			modules = append(modules, module)
		}
	}

	return modules, nil
}

// collectProjectTasks 获取项目的任务列表
func (dc *DataCollector) collectProjectTasks(ctx context.Context, projectID uint) ([]TaskSummary, error) {
	// 首先获取项目的所有模块
	modules, err := dc.collectProjectModules(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var allTasks []TaskSummary
	for _, module := range modules {
		tasks, err := dc.collectModuleTasks(ctx, module.ModuleID)
		if err != nil {
			continue
		}
		for _, task := range tasks {
			task.ModuleName = module.ModuleName
			moduleID := module.ModuleID
			task.ModuleID = &moduleID
		}
		allTasks = append(allTasks, tasks...)
	}

	return allTasks, nil
}

// parseJSONArray 解析JSON数组字符串
func parseJSONArray(s string) []string {
	if s == "" {
		return nil
	}

	var result []string
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return nil
	}
	return result
}

// parseUint 将字符串转换为uint
func parseUint(s string) uint {
	var result uint
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + uint(c-'0')
		}
	}
	return result
}
