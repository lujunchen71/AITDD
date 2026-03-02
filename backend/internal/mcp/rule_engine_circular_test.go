package mcp

import (
	"testing"
)

// =============================================
// E-S-09: 模块循环引用错误检测 - 单元测试
// =============================================

// TestCheckModuleCircularReference_NoCircular 测试用例1: 无循环引用（单向传递）
// 场景：模块X: A → B → C，模块Y: D → E → F
// 跨模块引用: A → E, C → D (都是 X → Y 方向)
// 预期结果: 无错误
func TestCheckModuleCircularReference_NoCircular(t *testing.T) {
	// 创建规则引擎
	engine := NewRuleEngine("")

	// 准备测试数据
	// 模块X的任务A引用模块Y的任务E，模块X的任务C引用模块Y的任务D
	// 所有引用都是 X → Y 方向，不存在循环
	modulesData := []map[string]interface{}{
		{
			"pathName": "Project/ModuleX",
			"name":     "模块X",
			"tasks": []interface{}{
				map[string]interface{}{
					"pathName": "Project/ModuleX/TaskA",
					"name":     "任务A",
					"codePaths": []interface{}{
						"Project/ModuleY/TaskE", // A → E (X → Y)
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleX/TaskB",
					"name":     "任务B",
					"codePaths": []interface{}{
						"Project/ModuleX/TaskC",
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleX/TaskC",
					"name":     "任务C",
					"codePaths": []interface{}{
						"Project/ModuleY/TaskD", // C → D (X → Y)
					},
				},
			},
		},
		{
			"pathName": "Project/ModuleY",
			"name":     "模块Y",
			"tasks": []interface{}{
				map[string]interface{}{
					"pathName": "Project/ModuleY/TaskD",
					"name":     "任务D",
					"codePaths": []interface{}{
						"Project/ModuleY/TaskE",
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleY/TaskE",
					"name":     "任务E",
					"codePaths": []interface{}{
						"Project/ModuleY/TaskF",
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleY/TaskF",
					"name":     "任务F",
					"codePaths": []interface{}{},
				},
			},
		},
	}

	// 执行检查
	result := engine.CheckModuleCircularReference(modulesData)

	// 验证结果：应该没有循环引用
	if result.HasCircularReference {
		t.Errorf("预期无循环引用，但检测到循环引用")
	}

	if len(result.CircularPairs) != 0 {
		t.Errorf("预期循环引用对数为0，实际为 %d", len(result.CircularPairs))
	}
}

// TestCheckModuleCircularReference_SimpleCircular 测试用例2: 存在循环引用 - 应报错
// 场景：模块X: A → B → C，模块Y: D → E → F
// 跨模块引用: A → E (X → Y) 和 D → B (Y → X)
// 预期结果: 报告 E-S-09 错误
func TestCheckModuleCircularReference_SimpleCircular(t *testing.T) {
	// 创建规则引擎
	engine := NewRuleEngine("")

	// 准备测试数据
	// 模块X的任务A引用模块Y的任务E (X → Y)
	// 模块Y的任务D引用模块X的任务B (Y → X)
	// 形成循环引用
	modulesData := []map[string]interface{}{
		{
			"pathName": "Project/ModuleX",
			"name":     "模块X",
			"tasks": []interface{}{
				map[string]interface{}{
					"pathName": "Project/ModuleX/TaskA",
					"name":     "任务A",
					"codePaths": []interface{}{
						"Project/ModuleY/TaskE", // A → E (X → Y)
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleX/TaskB",
					"name":     "任务B",
					"codePaths": []interface{}{
						"Project/ModuleX/TaskC",
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleX/TaskC",
					"name":     "任务C",
					"codePaths": []interface{}{},
				},
			},
		},
		{
			"pathName": "Project/ModuleY",
			"name":     "模块Y",
			"tasks": []interface{}{
				map[string]interface{}{
					"pathName": "Project/ModuleY/TaskD",
					"name":     "任务D",
					"codePaths": []interface{}{
						"Project/ModuleX/TaskB", // D → B (Y → X) 形成循环
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleY/TaskE",
					"name":     "任务E",
					"codePaths": []interface{}{
						"Project/ModuleY/TaskF",
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleY/TaskF",
					"name":     "任务F",
					"codePaths": []interface{}{},
				},
			},
		},
	}

	// 执行检查
	result := engine.CheckModuleCircularReference(modulesData)

	// 验证结果：应该检测到循环引用
	if !result.HasCircularReference {
		t.Errorf("预期检测到循环引用，但未检测到")
	}

	if len(result.CircularPairs) != 1 {
		t.Errorf("预期循环引用对数为1，实际为 %d", len(result.CircularPairs))
	}

	// 验证循环引用详情
	if len(result.CircularPairs) > 0 {
		pair := result.CircularPairs[0]
		// 验证模块名称（可能是 ModuleX/ModuleY 或 ModuleY/ModuleX）
		validModules := (pair.ModuleX == "模块X" && pair.ModuleY == "模块Y") ||
			(pair.ModuleX == "模块Y" && pair.ModuleY == "模块X")
		if !validModules {
			t.Errorf("循环引用模块名称不正确: ModuleX=%s, ModuleY=%s", pair.ModuleX, pair.ModuleY)
		}

		// 验证存在双向引用
		if len(pair.XToYReferences) == 0 {
			t.Errorf("预期存在 X→Y 引用，但未找到")
		}
		if len(pair.YToXReferences) == 0 {
			t.Errorf("预期存在 Y→X 引用，但未找到")
		}

		// 验证建议不为空
		if pair.Suggestion == "" {
			t.Errorf("预期有解决建议，但建议为空")
		}
	}
}

// TestCheckModuleCircularReference_MultipleModules 测试用例3: 多模块复杂场景
// 场景：模块X, Y, Z 三者之间
// X ↔ Y (双向引用), Y ↔ Z (双向引用)
// 预期结果: 报告2对循环引用错误
func TestCheckModuleCircularReference_MultipleModules(t *testing.T) {
	// 创建规则引擎
	engine := NewRuleEngine("")

	// 准备测试数据
	// X → Y (任务A引用任务D)
	// Y → X (任务E引用任务B)
	// Y → Z (任务F引用任务H)
	// Z → Y (任务I引用任务G)
	// 形成两对双向循环引用：X↔Y, Y↔Z
	modulesData := []map[string]interface{}{
		{
			"pathName": "Project/ModuleX",
			"name":     "模块X",
			"tasks": []interface{}{
				map[string]interface{}{
					"pathName": "Project/ModuleX/TaskA",
					"name":     "任务A",
					"codePaths": []interface{}{
						"Project/ModuleY/TaskD", // A → D (X → Y)
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleX/TaskB",
					"name":     "任务B",
					"codePaths": []interface{}{},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleX/TaskC",
					"name":     "任务C",
					"codePaths": []interface{}{},
				},
			},
		},
		{
			"pathName": "Project/ModuleY",
			"name":     "模块Y",
			"tasks": []interface{}{
				map[string]interface{}{
					"pathName": "Project/ModuleY/TaskD",
					"name":     "任务D",
					"codePaths": []interface{}{},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleY/TaskE",
					"name":     "任务E",
					"codePaths": []interface{}{
						"Project/ModuleX/TaskB", // E → B (Y → X)
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleY/TaskF",
					"name":     "任务F",
					"codePaths": []interface{}{
						"Project/ModuleZ/TaskH", // F → H (Y → Z)
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleY/TaskG",
					"name":     "任务G",
					"codePaths": []interface{}{},
				},
			},
		},
		{
			"pathName": "Project/ModuleZ",
			"name":     "模块Z",
			"tasks": []interface{}{
				map[string]interface{}{
					"pathName": "Project/ModuleZ/TaskH",
					"name":     "任务H",
					"codePaths": []interface{}{},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleZ/TaskI",
					"name":     "任务I",
					"codePaths": []interface{}{
						"Project/ModuleY/TaskG", // I → G (Z → Y)
					},
				},
			},
		},
	}

	// 执行检查
	result := engine.CheckModuleCircularReference(modulesData)

	// 验证结果：应该检测到循环引用
	if !result.HasCircularReference {
		t.Errorf("预期检测到循环引用，但未检测到")
	}

	// 应该检测到2对循环引用：X↔Y, Y↔Z
	if len(result.CircularPairs) != 2 {
		t.Errorf("预期循环引用对数为2，实际为 %d", len(result.CircularPairs))
	}

	// 验证循环引用详情格式正确
	for i, pair := range result.CircularPairs {
		if pair.ModuleX == "" || pair.ModuleY == "" {
			t.Errorf("循环引用对 #%d 模块名称为空", i+1)
		}
		if pair.Suggestion == "" {
			t.Errorf("循环引用对 #%d 建议为空", i+1)
		}
		// 验证双向引用都存在
		if len(pair.XToYReferences) == 0 {
			t.Errorf("循环引用对 #%d 缺少 X→Y 引用", i+1)
		}
		if len(pair.YToXReferences) == 0 {
			t.Errorf("循环引用对 #%d 缺少 Y→X 引用", i+1)
		}
	}
}

// TestCheckModuleCircularReference_NoCrossModuleRef 测试用例4: 无跨模块引用 - 应通过
// 场景：模块X: A → B → C，模块Y: D → E → F
// 无跨模块引用
// 预期结果: 无错误
func TestCheckModuleCircularReference_NoCrossModuleRef(t *testing.T) {
	// 创建规则引擎
	engine := NewRuleEngine("")

	// 准备测试数据
	// 所有任务只引用同模块内的其他任务，无跨模块引用
	modulesData := []map[string]interface{}{
		{
			"pathName": "Project/ModuleX",
			"name":     "模块X",
			"tasks": []interface{}{
				map[string]interface{}{
					"pathName": "Project/ModuleX/TaskA",
					"name":     "任务A",
					"codePaths": []interface{}{
						"Project/ModuleX/TaskB", // A → B (模块内)
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleX/TaskB",
					"name":     "任务B",
					"codePaths": []interface{}{
						"Project/ModuleX/TaskC", // B → C (模块内)
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleX/TaskC",
					"name":     "任务C",
					"codePaths": []interface{}{},
				},
			},
		},
		{
			"pathName": "Project/ModuleY",
			"name":     "模块Y",
			"tasks": []interface{}{
				map[string]interface{}{
					"pathName": "Project/ModuleY/TaskD",
					"name":     "任务D",
					"codePaths": []interface{}{
						"Project/ModuleY/TaskE", // D → E (模块内)
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleY/TaskE",
					"name":     "任务E",
					"codePaths": []interface{}{
						"Project/ModuleY/TaskF", // E → F (模块内)
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleY/TaskF",
					"name":     "任务F",
					"codePaths": []interface{}{},
				},
			},
		},
	}

	// 执行检查
	result := engine.CheckModuleCircularReference(modulesData)

	// 验证结果：应该没有循环引用
	if result.HasCircularReference {
		t.Errorf("预期无循环引用，但检测到循环引用")
	}

	if len(result.CircularPairs) != 0 {
		t.Errorf("预期循环引用对数为0，实际为 %d", len(result.CircularPairs))
	}
}

// TestCheckModuleCircularReference_SingleModule 测试用例5: 单模块场景
// 场景：只有一个模块
// 预期结果: 无错误（单模块不可能存在跨模块循环引用）
func TestCheckModuleCircularReference_SingleModule(t *testing.T) {
	// 创建规则引擎
	engine := NewRuleEngine("")

	// 准备测试数据：只有一个模块
	modulesData := []map[string]interface{}{
		{
			"pathName": "Project/ModuleX",
			"name":     "模块X",
			"tasks": []interface{}{
				map[string]interface{}{
					"pathName": "Project/ModuleX/TaskA",
					"name":     "任务A",
					"codePaths": []interface{}{
						"Project/ModuleX/TaskB",
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleX/TaskB",
					"name":     "任务B",
					"codePaths": []interface{}{
						"Project/ModuleX/TaskA", // 模块内循环，但不属于跨模块循环引用
					},
				},
			},
		},
	}

	// 执行检查
	result := engine.CheckModuleCircularReference(modulesData)

	// 验证结果：应该没有跨模块循环引用
	if result.HasCircularReference {
		t.Errorf("单模块场景不应检测到跨模块循环引用")
	}

	if len(result.CircularPairs) != 0 {
		t.Errorf("预期循环引用对数为0，实际为 %d", len(result.CircularPairs))
	}
}

// TestCheckModuleCircularReference_EmptyModules 测试用例6: 空模块列表
// 场景：没有模块数据
// 预期结果: 无错误
func TestCheckModuleCircularReference_EmptyModules(t *testing.T) {
	// 创建规则引擎
	engine := NewRuleEngine("")

	// 准备测试数据：空模块列表
	modulesData := []map[string]interface{}{}

	// 执行检查
	result := engine.CheckModuleCircularReference(modulesData)

	// 验证结果：应该没有循环引用
	if result.HasCircularReference {
		t.Errorf("空模块列表不应检测到循环引用")
	}

	if len(result.CircularPairs) != 0 {
		t.Errorf("预期循环引用对数为0，实际为 %d", len(result.CircularPairs))
	}
}

// TestCheckModuleCircularReference_MultipleCrossRefs 测试用例7: 多个跨模块引用对形成循环
// 场景：模块X和Y之间存在多个任务的相互引用
// 预期结果: 检测到循环引用，并列出所有引用对
func TestCheckModuleCircularReference_MultipleCrossRefs(t *testing.T) {
	// 创建规则引擎
	engine := NewRuleEngine("")

	// 准备测试数据
	// X → Y: 任务A→任务D, 任务B→任务E
	// Y → X: 任务F→任务C, 任务G→任务A
	modulesData := []map[string]interface{}{
		{
			"pathName": "Project/ModuleX",
			"name":     "模块X",
			"tasks": []interface{}{
				map[string]interface{}{
					"pathName": "Project/ModuleX/TaskA",
					"name":     "任务A",
					"codePaths": []interface{}{
						"Project/ModuleY/TaskD", // X → Y
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleX/TaskB",
					"name":     "任务B",
					"codePaths": []interface{}{
						"Project/ModuleY/TaskE", // X → Y
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleX/TaskC",
					"name":     "任务C",
					"codePaths": []interface{}{},
				},
			},
		},
		{
			"pathName": "Project/ModuleY",
			"name":     "模块Y",
			"tasks": []interface{}{
				map[string]interface{}{
					"pathName": "Project/ModuleY/TaskD",
					"name":     "任务D",
					"codePaths": []interface{}{},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleY/TaskE",
					"name":     "任务E",
					"codePaths": []interface{}{},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleY/TaskF",
					"name":     "任务F",
					"codePaths": []interface{}{
						"Project/ModuleX/TaskC", // Y → X
					},
				},
				map[string]interface{}{
					"pathName": "Project/ModuleY/TaskG",
					"name":     "任务G",
					"codePaths": []interface{}{
						"Project/ModuleX/TaskA", // Y → X
					},
				},
			},
		},
	}

	// 执行检查
	result := engine.CheckModuleCircularReference(modulesData)

	// 验证结果：应该检测到循环引用
	if !result.HasCircularReference {
		t.Errorf("预期检测到循环引用，但未检测到")
	}

	if len(result.CircularPairs) != 1 {
		t.Errorf("预期循环引用对数为1，实际为 %d", len(result.CircularPairs))
	}

	// 验证引用详情
	if len(result.CircularPairs) > 0 {
		pair := result.CircularPairs[0]

		// 应该有2个 X→Y 引用
		if len(pair.XToYReferences) != 2 {
			t.Errorf("预期 X→Y 引用数为2，实际为 %d", len(pair.XToYReferences))
		}

		// 应该有2个 Y→X 引用
		if len(pair.YToXReferences) != 2 {
			t.Errorf("预期 Y→X 引用数为2，实际为 %d", len(pair.YToXReferences))
		}
	}
}

// TestGenerateCircularReferenceReport 测试报告生成功能
func TestGenerateCircularReferenceReport(t *testing.T) {
	// 测试无循环引用的报告
	noCircularResult := &CheckModuleCircularReferenceResult{
		HasCircularReference: false,
		CircularPairs:        []CircularReferenceDetail{},
	}
	report := noCircularResult.GenerateCircularReferenceReport()
	if report == "" {
		t.Errorf("报告不应为空")
	}
	// 检查报告包含成功标识
	if !containsString(report, "✅") {
		t.Errorf("无循环引用报告应包含成功标识")
	}

	// 测试有循环引用的报告
	circularResult := &CheckModuleCircularReferenceResult{
		HasCircularReference: true,
		CircularPairs: []CircularReferenceDetail{
			{
				ModuleX: "模块X",
				ModuleY: "模块Y",
				XToYReferences: []CrossModuleReference{
					{SourceTask: "任务A", TargetTask: "任务D"},
				},
				YToXReferences: []CrossModuleReference{
					{SourceTask: "任务F", TargetTask: "任务C"},
				},
				Suggestion: "建议引入第三方共享模块来打破循环依赖",
			},
		},
	}
	report = circularResult.GenerateCircularReferenceReport()
	if report == "" {
		t.Errorf("报告不应为空")
	}
	// 检查报告包含错误标识
	if !containsString(report, "❌") {
		t.Errorf("有循环引用报告应包含错误标识")
	}
	// 检查报告包含模块名称
	if !containsString(report, "模块X") || !containsString(report, "模块Y") {
		t.Errorf("报告应包含模块名称")
	}
	// 检查报告包含任务引用
	if !containsString(report, "任务A") || !containsString(report, "任务D") {
		t.Errorf("报告应包含任务引用详情")
	}
}

// TestCheckModuleCircularReference_WithJSONCodePaths 测试使用 JSON 字符串格式的 codePaths
func TestCheckModuleCircularReference_WithJSONCodePaths(t *testing.T) {
	// 创建规则引擎
	engine := NewRuleEngine("")

	// 准备测试数据，codePaths 使用 JSON 字符串格式
	modulesData := []map[string]interface{}{
		{
			"pathName": "Project/ModuleX",
			"name":     "模块X",
			"tasks": []interface{}{
				map[string]interface{}{
					"pathName":  "Project/ModuleX/TaskA",
					"name":      "任务A",
					"codePaths": `["Project/ModuleY/TaskD"]`, // JSON 字符串格式
				},
			},
		},
		{
			"pathName": "Project/ModuleY",
			"name":     "模块Y",
			"tasks": []interface{}{
				map[string]interface{}{
					"pathName":  "Project/ModuleY/TaskD",
					"name":      "任务D",
					"codePaths": `["Project/ModuleX/TaskA"]`, // JSON 字符串格式，形成循环
				},
			},
		},
	}

	// 执行检查
	result := engine.CheckModuleCircularReference(modulesData)

	// 验证结果：应该检测到循环引用
	if !result.HasCircularReference {
		t.Errorf("预期检测到循环引用，但未检测到")
	}
}

// 辅助函数：检查字符串是否包含子串
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
