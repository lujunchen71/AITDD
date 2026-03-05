package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSaveAIAnalysisLog(t *testing.T) {
	// 创建临时测试目录
	tempDir := ".aitdd_test_compile"
	defer os.RemoveAll(tempDir)

	// 保存当前目录
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// 切换到临时目录
	os.MkdirAll(tempDir, 0755)
	os.Chdir(tempDir)

	// 创建测试日志（task 类型，问题索引 0）
	log := &AIAnalysisLog{
		Timestamp:        time.Now().Format(time.RFC3339),
		ProjectPathName:  "TestProject",
		TargetPathName:   "TestProject/测试模块/测试任务",
		TargetName:       "测试任务",
		AnalysisType:     "task",
		QuestionIndex:    0,
		Question:         "请分析以下任务的提示词是否清晰...",
		QuestionCategory: "prompt",
		QuestionSeverity: "error",
		SystemPrompt:     "你是一个专业的软件工程分析助手",
		UserPrompt:       "请分析以下任务...",
		RawResponse:      `{"summary": "测试响应", "score": 85}`,
	}

	// 测试保存日志
	err := SaveAIAnalysisLog("TestProject", "TestProject/测试模块/测试任务", log)
	if err != nil {
		t.Fatalf("SaveAIAnalysisLog 失败: %v", err)
	}

	// 新格式：task_{projectName}_{taskFlatPath}_{questionIndex}.json
	// targetPathName = "TestProject/测试模块/测试任务"
	// flatPath 去掉项目名前缀后 = "测试模块_测试任务"
	// fileName = "TestProject_测试模块_测试任务_0.json"
	expectedPath := filepath.Join(".aitdd", "compile", "TestProject", "TestProject_测试模块_测试任务_0.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("日志文件未创建: %s", expectedPath)
	}

	// 读取并验证文件内容
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	// 验证JSON包含预期字段
	content := string(data)
	expectedFields := []string{
		`"timestamp"`,
		`"project_path_name"`,
		`"target_path_name"`,
		`"target_name"`,
		`"analysis_type"`,
		`"question_index"`,
		`"question"`,
		`"question_category"`,
		`"question_severity"`,
		`"system_prompt"`,
		`"user_prompt"`,
		`"raw_response"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(content, field) {
			t.Errorf("日志文件缺少字段: %s", field)
		}
	}

	t.Logf("日志文件内容验证成功，文件路径: %s", expectedPath)
	t.Logf("文件内容:\n%s", content)
}

func TestSaveAIAnalysisLogWithResult(t *testing.T) {
	// 创建临时测试目录
	tempDir := ".aitdd_test_result"
	defer os.RemoveAll(tempDir)

	// 保存当前目录
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// 切换到临时目录
	os.MkdirAll(tempDir, 0755)
	os.Chdir(tempDir)

	// 创建包含解析结果的测试日志（task 类型，问题索引 2）
	log := &AIAnalysisLog{
		Timestamp:        time.Now().Format(time.RFC3339),
		ProjectPathName:  "ResultProject",
		TargetPathName:   "ResultProject/模块/任务",
		TargetName:       "任务",
		AnalysisType:     "task",
		QuestionIndex:    2,
		Question:         "请分析接口完整性",
		QuestionCategory: "contract",
		QuestionSeverity: "warning",
		SystemPrompt:     "系统提示",
		UserPrompt:       "用户提示",
		RawResponse:      `{"summary": "测试", "score": 90}`,
		ParsedResult: &AnalysisResult{
			AnalysisID:   "test_123",
			AnalysisType: "task",
			TargetID:     1,
			TargetName:   "测试任务",
			Score:        90,
			Confidence:   0.85,
			Summary:      "测试摘要",
			Issues: []Issue{
				{
					ID:          "issue_1",
					Category:    "completeness",
					Severity:    "warning",
					Description: "缺少描述",
				},
			},
			Suggestions: []string{"建议1"},
		},
	}

	// 测试保存日志
	err := SaveAIAnalysisLog("ResultProject", "ResultProject/模块/任务", log)
	if err != nil {
		t.Fatalf("SaveAIAnalysisLog 失败: %v", err)
	}

	// 新格式：task 类型
	// targetPathName = "ResultProject/模块/任务"
	// flatPath 去掉项目名前缀 "ResultProject_" 后 = "模块_任务"
	// fileName = "ResultProject_模块_任务_2.json"
	expectedPath := filepath.Join(".aitdd", "compile", "ResultProject", "ResultProject_模块_任务_2.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("日志文件未创建: %s", expectedPath)
	}

	// 读取并验证文件内容
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	content := string(data)

	// 验证包含解析结果字段
	expectedResultFields := []string{
		`"parsed_result"`,
		`"analysis_id"`,
		`"test_123"`,
		`"score"`,
		`90`,
		`"question_index"`,
	}

	for _, field := range expectedResultFields {
		if !strings.Contains(content, field) {
			t.Errorf("日志文件缺少解析结果字段: %s", field)
		}
	}

	t.Logf("带解析结果的日志文件验证成功")
}

func TestSaveAIAnalysisLogWithError(t *testing.T) {
	// 创建临时测试目录
	tempDir := ".aitdd_test_error"
	defer os.RemoveAll(tempDir)

	// 保存当前目录
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// 切换到临时目录
	os.MkdirAll(tempDir, 0755)
	os.Chdir(tempDir)

	// 创建包含错误的测试日志（task 类型，问题索引 3）
	log := &AIAnalysisLog{
		Timestamp:       time.Now().Format(time.RFC3339),
		ProjectPathName: "ErrorProject",
		TargetPathName:  "ErrorProject/模块/失败任务",
		TargetName:      "失败任务",
		AnalysisType:    "task",
		QuestionIndex:   3,
		SystemPrompt:    "系统提示",
		UserPrompt:      "用户提示",
		RawResponse:     "",
		Error:           "AI调用失败: 连接超时",
	}

	// 测试保存日志
	err := SaveAIAnalysisLog("ErrorProject", "ErrorProject/模块/失败任务", log)
	if err != nil {
		t.Fatalf("SaveAIAnalysisLog 失败: %v", err)
	}

	// 新格式：fileName = "ErrorProject_模块_失败任务_3.json"
	expectedPath := filepath.Join(".aitdd", "compile", "ErrorProject", "ErrorProject_模块_失败任务_3.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("日志文件未创建: %s", expectedPath)
	}

	// 读取并验证文件内容
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	content := string(data)

	// 验证包含错误字段
	if !strings.Contains(content, `"error"`) {
		t.Errorf("日志文件缺少错误字段")
	}
	if !strings.Contains(content, "AI调用失败") {
		t.Errorf("日志文件缺少错误内容")
	}

	t.Logf("带错误的日志文件验证成功")
}

func TestSaveAIAnalysisLogModule(t *testing.T) {
	// 创建临时测试目录
	tempDir := ".aitdd_test_module"
	defer os.RemoveAll(tempDir)

	// 保存当前目录
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// 切换到临时目录
	os.MkdirAll(tempDir, 0755)
	os.Chdir(tempDir)

	// 创建 module 类型的测试日志（问题索引 1）
	log := &AIAnalysisLog{
		Timestamp:        time.Now().Format(time.RFC3339),
		ProjectPathName:  "MyProject",
		TargetPathName:   "MyProject/用户管理",
		TargetName:       "用户管理",
		AnalysisType:     "module",
		QuestionIndex:    1,
		Question:         "请分析模块职责是否清晰",
		QuestionCategory: "architecture",
		QuestionSeverity: "warning",
		SystemPrompt:     "系统提示",
		UserPrompt:       "用户提示",
		RawResponse:      `{"summary": "模块职责清晰", "score": 88}`,
	}

	err := SaveAIAnalysisLog("MyProject", "MyProject/用户管理", log)
	if err != nil {
		t.Fatalf("SaveAIAnalysisLog 失败: %v", err)
	}

	// 新格式：module类型 → module_{projectName}_{moduleFlatPath}_{questionIndex}.json
	// targetPathName = "MyProject/用户管理"
	// flatPath 去掉项目名前缀后 = "用户管理"
	// fileName = "module_MyProject_用户管理_1.json"
	expectedPath := filepath.Join(".aitdd", "compile", "MyProject", "module_MyProject_用户管理_1.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("module日志文件未创建: %s", expectedPath)
	}

	t.Logf("module日志文件验证成功，文件路径: %s", expectedPath)
}

func TestSaveAIAnalysisLogProject(t *testing.T) {
	// 创建临时测试目录
	tempDir := ".aitdd_test_project"
	defer os.RemoveAll(tempDir)

	// 保存当前目录
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// 切换到临时目录
	os.MkdirAll(tempDir, 0755)
	os.Chdir(tempDir)

	// 创建 project 类型的测试日志（问题索引 0）
	log := &AIAnalysisLog{
		Timestamp:        time.Now().Format(time.RFC3339),
		ProjectPathName:  "BigProject",
		TargetPathName:   "BigProject",
		TargetName:       "BigProject",
		AnalysisType:     "project",
		QuestionIndex:    0,
		Question:         "请分析项目架构是否合理",
		QuestionCategory: "architecture",
		QuestionSeverity: "error",
		SystemPrompt:     "系统提示",
		UserPrompt:       "用户提示",
		RawResponse:      `{"summary": "架构合理", "score": 92}`,
	}

	err := SaveAIAnalysisLog("BigProject", "BigProject", log)
	if err != nil {
		t.Fatalf("SaveAIAnalysisLog 失败: %v", err)
	}

	// project类型 → project_{questionIndex}.json
	expectedPath := filepath.Join(".aitdd", "compile", "BigProject", "project_0.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("project日志文件未创建: %s", expectedPath)
	}

	t.Logf("project日志文件验证成功，文件路径: %s", expectedPath)
}

func TestGenerateLogFileName(t *testing.T) {
	tests := []struct {
		name           string
		analysisType   string
		projectName    string
		targetPathName string
		questionIndex  int
		expected       string
	}{
		{
			name:           "task问题",
			analysisType:   "task",
			projectName:    "MyProject",
			targetPathName: "MyProject/用户管理/登录功能",
			questionIndex:  0,
			expected:       "MyProject_用户管理_登录功能_0.json",
		},
		{
			name:           "module问题",
			analysisType:   "module",
			projectName:    "MyProject",
			targetPathName: "MyProject/用户管理",
			questionIndex:  2,
			expected:       "module_MyProject_用户管理_2.json",
		},
		{
			name:           "project问题",
			analysisType:   "project",
			projectName:    "BigProject",
			targetPathName: "BigProject",
			questionIndex:  1,
			expected:       "project_1.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateLogFileName(tt.analysisType, tt.projectName, tt.targetPathName, tt.questionIndex)
			if result != tt.expected {
				t.Errorf("generateLogFileName() = %q, want %q", result, tt.expected)
			}
		})
	}
}
