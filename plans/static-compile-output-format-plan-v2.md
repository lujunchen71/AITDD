# 静态编译错误输出格式重构计划 v2

## 1. 背景

当前的 `compile_static` MCP 工具的输出格式按规则（ruleId）分组显示错误，导致同一个任务/模块的多个错误分散在不同地方，不便于用户查看和定位问题。

## 2. 核心改进

### 2.1 分组维度变化

从**按规则分组** → **按资源分组**，让用户可以一目了然地看到某个资源的所有问题。

### 2.2 输出优化规则

1. **不显示无错误的资源**：如果资源的 error 和 warning 都为空，则不显示该资源
2. **不显示空的 warning 数组**：如果 warning 为空，则不显示 warning 字段
3. **展示所有错误类型**：完整展示 E-S-01 到 E-S-09 等所有错误类型

## 3. 新数据结构定义

```go
// CompileIssueItem 单个编译问题
type CompileIssueItem struct {
    RuleID     string `json:"ruleId"`     // 规则ID
    RuleName   string `json:"ruleName"`   // 规则名称
    Message    string `json:"message"`    // 具体错误信息
    Suggestion string `json:"suggestion"` // 修复建议
}

// CompileResourceIssues 单个资源的所有问题
type CompileResourceIssues struct {
    Error   []CompileIssueItem `json:"error,omitempty"`   // 错误列表
    Warning []CompileIssueItem `json:"warning,omitempty"` // 警告列表（为空时不显示）
}

// CompileStaticOutputV2 新版静态编译输出
type CompileStaticOutputV2 struct {
    Success         bool                               `json:"success"`
    PathName        string                             `json:"pathName,omitempty"`
    TotalModules    int                                `json:"totalModules"`
    TotalTasks      int                                `json:"totalTasks"`
    ErrorCount      int                                `json:"errorCount"`
    WarningCount    int                                `json:"warningCount"`
    Module          map[string]CompileResourceIssues   `json:"module,omitempty"`
    Task            map[string]CompileResourceIssues   `json:"task,omitempty"`
    Message         string                             `json:"message"`
}
```

## 4. 错误类型清单

| 规则ID | 规则名称 | 检查条件 |
|--------|----------|----------|
| E-S-01 | 代码路径未声明 | 任务的 codePaths 为空 |
| E-S-02 | 存在 Bug 日志 | 任务的 bugLog 包含错误记录 |
| E-S-03 | 需要人工协助 | 任务的 humanAssistance 非空 |
| E-S-04 | 提示词为空 | 模块或任务的 prompt 为空 |
| E-S-05 | 上游契约检查失败 | 上游契约条目未找到提供该接口的上游任务（没有定义这条引用从哪里来） |
| E-S-06 | 下游契约检查失败 | 下游契约条目未找到接收该接口的下游任务（这条给外部的引用，没人接收） |
| E-S-08 | 状态异常 | completed 状态的任务存在 bug |
| E-S-09 | 模块循环引用 | 跨模块任务引用形成循环依赖 |

## 5. 新格式输出示例

以下示例展示了包含多种错误类型的完整输出格式：

```json
{
  "success": false,
  "pathName": "AITDD Painter",
  "totalModules": 5,
  "totalTasks": 14,
  "errorCount": 96,
  "warningCount": 14,
  
  "module": {
    "AITDD Painter/主界面模块": {
      "error": [
        {
          "ruleId": "E-S-09",
          "ruleName": "模块循环引用",
          "message": "检测到跨模块循环引用: 主界面模块/主窗口框架 → 画布模块/画布渲染 → 主界面模块/工具栏组件，由任务 [主窗口框架] 依赖 [画布渲染]，[画布渲染] 依赖 [工具栏组件]，[工具栏组件] 又依赖 [主窗口框架] 形成循环",
          "suggestion": "需要思考整个项目设计来处理"
        },
        {
          "ruleId": "E-S-04",
          "ruleName": "提示词为空",
          "message": "模块 [主界面模块] 提示词为空",
          "suggestion": "请为模块添加提示词"
        }
      ]
    }
  },
  
  "task": {
    "AITDD Painter/主界面模块/主窗口框架": {
      "error": [
        {
          "ruleId": "E-S-05",
          "ruleName": "上游契约检查失败",
          "message": "任务[主窗口框架] 上游契约条目 {\"label\": \"基础运算功能\", \"contract_api\": \"calculate(expression: str) -> str\", \"from\": \"PYQT6Calculator/计算核心模块/基础运算器\"} 未找到提供该接口的上游任务",
          "suggestion": "请检查契约定义是否正确，确保上下游契约匹配"
        },
        {
          "ruleId": "E-S-01",
          "ruleName": "代码路径未声明",
          "message": "任务未声明代码路径",
          "suggestion": "为任务声明相关的实际代码文件相对路径"
        }
      ],
      "warning": [
        {
          "ruleId": "W-S-01",
          "ruleName": "测试用例为空",
          "message": "任务 [主窗口框架] 未定义测试用例",
          "suggestion": "建议添加测试用例"
        }
      ]
    },
    "AITDD Painter/主界面模块/工具栏组件": {
      "error": [
        {
          "ruleId": "E-S-02",
          "ruleName": "存在 Bug 日志",
          "message": "任务 [工具栏组件] 存在未解决的 Bug：点击无响应",
          "suggestion": "请修复 Bug 后清除日志"
        },
        {
          "ruleId": "E-S-03",
          "ruleName": "需要人工协助",
          "message": "任务 [工具栏组件] 需要人工协助：设计决策待确认",
          "suggestion": "请查看 humanAssistance 字段获取详情"
        }
      ]
    },
    "AITDD Painter/画布模块/画布渲染": {
      "error": [
        {
          "ruleId": "E-S-06",
          "ruleName": "下游契约检查失败",
          "message": "任务[画布渲染] 下游契约条目 {\"label\": \"渲染接口\", \"contract_api\": \"render(canvas: Canvas) -> void\", \"from\": \"AITDD Painter/画布模块/画布渲染\"} 未找到接收该接口的下游任务",
          "suggestion": "请检查是否有任务依赖此接口"
        },
        {
          "ruleId": "E-S-08",
          "ruleName": "状态异常",
          "message": "任务 [画布渲染] 状态为 'completed' 但存在未解决的 Bug",
          "suggestion": "请修复 Bug 后更新任务状态"
        }
      ]
    }
  }
}
```

**格式要点**：
- `"AITDD Painter/工具模块"` 没有错误，因此不显示
- `"AITDD Painter/主界面模块/工具栏组件"` 没有 warning，因此不显示 warning 字段
- 所有错误类型都可以在同一资源下集中展示

## 6. 需要修改的代码文件

| 文件路径 | 修改内容 |
|---------|---------|
| [`backend/internal/mcp/tools_other.go`](backend/internal/mcp/tools_other.go) | 1. 添加新的数据结构定义<br>2. 修改 `groupCompileIssuesByResource` 函数<br>3. 更新 `handleCompileStaticImpl` 输出逻辑 |
| [`backend/internal/mcp/tools_other_test.go`](backend/internal/mcp/tools_other_test.go) | 更新相关测试用例以适配新格式 |

## 7. 实现要点

### 7.1 过滤无错误资源

在构建输出时，只添加有错误或警告的资源：

```go
// 只添加有问题的资源
if len(issues.Error) > 0 || len(issues.Warning) > 0 {
    output.Module[pathName] = issues
}
```

### 7.2 省略空 warning

使用 `omitempty` 标签自动省略空数组：

```go
type CompileResourceIssues struct {
    Error   []CompileIssueItem `json:"error,omitempty"`
    Warning []CompileIssueItem `json:"warning,omitempty"`
}
```

## 8. 总结

本计划的核心变化：

1. **分组维度变化**：从按规则分组 → 按资源分组
2. **信息更集中**：同一资源的所有错误和警告集中显示
3. **输出更简洁**：不显示无错误的资源，不显示空的 warning 数组
4. **错误类型完整**：展示所有 E-S-01 到 E-S-09 错误类型

这种新格式更符合用户的思维模式：用户通常关心的是"某个任务有什么问题"，而不是"某个规则影响了哪些任务"。
