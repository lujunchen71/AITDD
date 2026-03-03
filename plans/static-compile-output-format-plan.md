# 静态编译输出格式修改策略

## 1. 概述

### 1.1 背景
当前静态编译输出的警告/错误格式是每条单独显示，当同一规则触发多次时会重复显示规则信息，不够简洁。

### 1.2 目标
将同一类错误/警告合并显示，并输出为JSON格式，减少冗余信息，提高可读性。

## 2. 现有代码结构分析

### 2.1 数据结构定义

#### CompileIssue - 编译问题（单条）
**文件**: [`backend/internal/mcp/tools_other.go:1426-1435`](backend/internal/mcp/tools_other.go:1426)

```go
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
```

#### CompileStaticResult - 静态编译结果
**文件**: [`backend/internal/mcp/tools_other.go:1437-1449`](backend/internal/mcp/tools_other.go:1437)

```go
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
```

### 2.2 输出逻辑

**文件**: [`backend/internal/mcp/tools_other.go:1513-1550`](backend/internal/mcp/tools_other.go:1513)

当前输出逻辑在 `handleCompileStaticImpl` 函数中：

```go
// 构建输出
var output strings.Builder
output.WriteString(fmt.Sprintf("success: %v\n", result.Success))
// ... 统计信息 ...

if len(result.Errors) > 0 {
    output.WriteString("errors:\n")
    for _, err := range result.Errors {
        output.WriteString(fmt.Sprintf("  - ruleId: \"%s\"\n", err.RuleID))
        output.WriteString(fmt.Sprintf("    ruleName: \"%s\"\n", err.RuleName))
        // ... 每条单独输出 ...
    }
}
```

### 2.3 问题生成位置

警告/错误通过以下函数生成：
- [`compileStaticTaskFromData`](backend/internal/mcp/tools_other.go:2000) - 任务级别检查
- [`checkUpstreamContract`](backend/internal/mcp/tools_other.go:2112) - 上游契约检查
- [`checkDownstreamContract`](backend/internal/mcp/tools_other.go:2196) - 下游契约检查
- [`checkStatusAnomaly`](backend/internal/mcp/tools_other.go:2278) - 状态异常检查
- [`checkContractConsistency`](backend/internal/mcp/tools_other.go:2410) - 契约一致性检查
- [`checkModuleCrossDependency`](backend/internal/mcp/tools_other.go:3069) - 模块跨依赖检查
- [`checkModuleCircularReference`](backend/internal/mcp/tools_other.go:3211) - 模块循环引用检查
- [`checkOrphanTask`](backend/internal/mcp/tools_other.go:3339) - 孤立任务检查

## 3. 新的数据结构设计

### 3.1 新增结构体

#### CompileIssueResource - 资源引用
```go
// CompileIssueResource 编译问题关联的资源
type CompileIssueResource struct {
    Name     string `json:"name"`     // 资源名称
    PathName string `json:"pathName"` // 资源路径
}
```

#### CompileIssueGroup - 分组后的编译问题
```go
// CompileIssueGroup 按规则分组的编译问题
type CompileIssueGroup struct {
    RuleID       string                 `json:"ruleId"`       // 规则ID
    RuleName     string                 `json:"ruleName"`     // 规则名称
    ResourceType string                 `json:"resourceType"` // module, task
    Resources    []CompileIssueResource `json:"resources"`    // 受影响的资源列表
    Message      string                 `json:"message"`      // 通用消息
    Suggestion   string                 `json:"suggestion"`   // 建议操作
}
```

### 3.2 修改后的 CompileStaticResult

```go
// CompileStaticResult 静态编译结果（修改后）
type CompileStaticResult struct {
    Success         bool                `json:"success"`
    TotalModules    int                 `json:"totalModules"`
    TotalTasks      int                 `json:"totalTasks"`
    CompletedTasks  int                 `json:"completedTasks"`
    InProgressTasks int                 `json:"inProgressTasks"`
    ReadyTasks      int                 `json:"readyTasks"`
    ErrorCount      int                 `json:"errorCount"`
    WarningCount    int                 `json:"warningCount"`
    Errors          []CompileIssueGroup `json:"errors"`    // 改为分组格式
    Warnings        []CompileIssueGroup `json:"warnings"`  // 改为分组格式
}
```

### 3.3 JSON 输出格式示例

```json
{
  "success": false,
  "totalModules": 2,
  "totalTasks": 5,
  "completedTasks": 2,
  "inProgressTasks": 1,
  "readyTasks": 2,
  "errorCount": 1,
  "warningCount": 2,
  "errors": [
    {
      "ruleId": "E-S-09",
      "ruleName": "模块循环引用错误",
      "resourceType": "module",
      "resources": [
        {"name": "计算引擎", "pathName": "计算器程序/计算引擎"},
        {"name": "显示模块", "pathName": "计算器程序/显示模块"}
      ],
      "message": "检测到跨模块循环引用",
      "suggestion": "建议引入第三方共享模块来打破循环依赖"
    }
  ],
  "warnings": [
    {
      "ruleId": "W-S-01",
      "ruleName": "测试用例为空",
      "resourceType": "task",
      "resources": [
        {"name": "数字解析", "pathName": "计算器程序/计算引擎/数字解析"},
        {"name": "运算执行", "pathName": "计算器程序/计算引擎/运算执行"}
      ],
      "message": "未定义测试用例",
      "suggestion": "建议添加测试用例"
    },
    {
      "ruleId": "W-S-02",
      "ruleName": "上游契约为空",
      "resourceType": "task",
      "resources": [
        {"name": "运算执行", "pathName": "计算器程序/计算引擎/运算执行"}
      ],
      "message": "未定义上游契约详情",
      "suggestion": "如果任务有上游依赖，请定义上游契约详情"
    }
  ]
}
```

## 4. 需要修改的文件和函数

### 4.1 文件列表

| 文件 | 修改类型 | 说明 |
|------|----------|------|
| `backend/internal/mcp/tools_other.go` | 修改 | 主要修改文件 |

### 4.2 函数修改列表

| 函数名 | 行号范围 | 修改内容 |
|--------|----------|----------|
| `CompileIssueResource` | 新增 | 新增结构体 |
| `CompileIssueGroup` | 新增 | 新增结构体 |
| `CompileStaticResult` | 1437-1449 | 修改 Errors/Warnings 字段类型 |
| `handleCompileStaticImpl` | 1477-1551 | 修改输出逻辑，添加分组和JSON输出 |
| `groupCompileIssues` | 新增 | 新增分组函数 |

## 5. 具体修改步骤

### 步骤 1: 添加新的数据结构

在 [`backend/internal/mcp/tools_other.go`](backend/internal/mcp/tools_other.go:1425) 的 `CompileIssue` 定义之后添加：

```go
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
```

### 步骤 2: 添加分组函数

添加一个新函数将 `[]CompileIssue` 转换为 `[]CompileIssueGroup`：

```go
// groupCompileIssues 将编译问题按规则分组
func groupCompileIssues(issues []CompileIssue) []CompileIssueGroup {
    groups := make(map[string]*CompileIssueGroup)
    
    for _, issue := range issues {
        // 使用 ruleID + resourceType 作为分组键
        key := issue.RuleID + "_" + issue.ResourceType
        
        if group, exists := groups[key]; exists {
            // 添加到现有分组
            group.Resources = append(group.Resources, CompileIssueResource{
                Name:     issue.ResourceName,
                PathName: issue.ResourcePathName,
            })
        } else {
            // 创建新分组
            groups[key] = &CompileIssueGroup{
                RuleID:       issue.RuleID,
                RuleName:     issue.RuleName,
                ResourceType: issue.ResourceType,
                Resources: []CompileIssueResource{
                    {
                        Name:     issue.ResourceName,
                        PathName: issue.ResourcePathName,
                    },
                },
                Message:    issue.Message,
                Suggestion: issue.Suggestion,
            }
        }
    }
    
    // 转换为切片
    result := make([]CompileIssueGroup, 0, len(groups))
    for _, group := range groups {
        result = append(result, *group)
    }
    
    return result
}
```

### 步骤 3: 修改输出逻辑

修改 [`handleCompileStaticImpl`](backend/internal/mcp/tools_other.go:1477) 函数的输出部分：

```go
// handleCompileStaticImpl 静态编译
func (s *MCPServer) handleCompileStaticImpl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // ... 现有逻辑保持不变直到输出部分 ...
    
    // 更新统计
    result.ErrorCount = len(result.Errors)
    result.WarningCount = len(result.Warnings)
    result.Success = result.ErrorCount == 0

    // 构建JSON输出
    output := CompileStaticOutput{
        Success:         result.Success,
        TotalModules:    result.TotalModules,
        TotalTasks:      result.TotalTasks,
        CompletedTasks:  result.CompletedTasks,
        InProgressTasks: result.InProgressTasks,
        ReadyTasks:      result.ReadyTasks,
        ErrorCount:      result.ErrorCount,
        WarningCount:    result.WarningCount,
        Errors:          groupCompileIssues(result.Errors),
        Warnings:        groupCompileIssues(result.Warnings),
    }

    // 序列化为JSON
    data, err := json.MarshalIndent(output, "", "  ")
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("序列化结果失败: %v", err)), nil
    }

    return mcp.NewToolResultText(string(data)), nil
}
```

### 步骤 4: 添加输出结构体（可选，用于明确输出格式）

```go
// CompileStaticOutput 静态编译输出格式
type CompileStaticOutput struct {
    Success         bool                `json:"success"`
    TotalModules    int                 `json:"totalModules"`
    TotalTasks      int                 `json:"totalTasks"`
    CompletedTasks  int                 `json:"completedTasks"`
    InProgressTasks int                 `json:"inProgressTasks"`
    ReadyTasks      int                 `json:"readyTasks"`
    ErrorCount      int                 `json:"errorCount"`
    WarningCount    int                 `json:"warningCount"`
    Errors          []CompileIssueGroup `json:"errors"`
    Warnings        []CompileIssueGroup `json:"warnings"`
}
```

## 6. 影响分析

### 6.1 向后兼容性

- **不兼容变更**: 输出格式从纯文本变为JSON
- **影响范围**: 所有调用 `compile_static` MCP 工具的客户端
- **缓解措施**: 
  - 可以添加一个参数 `outputFormat` 支持两种格式
  - 或者在文档中明确标注这是Breaking Change

### 6.2 测试影响

需要更新以下测试文件：
- [`backend/internal/mcp/tools_other_test.go`](backend/internal/mcp/tools_other_test.go:618) 中的 `TestHandleCompileStaticImpl_*` 系列测试
- [`backend/internal/mcp/tools_other_test.go`](backend/internal/mcp/tools_other_test.go:779) 中的 `TestCompileStaticResult_Structure` 测试

## 7. 实施计划

### 阶段 1: 数据结构准备
1. 添加 `CompileIssueResource` 结构体
2. 添加 `CompileIssueGroup` 结构体
3. 添加 `CompileStaticOutput` 结构体（可选）

### 阶段 2: 核心逻辑实现
1. 实现 `groupCompileIssues` 分组函数
2. 编写分组函数的单元测试

### 阶段 3: 输出逻辑修改
1. 修改 `handleCompileStaticImpl` 的输出部分
2. 更新相关测试

### 阶段 4: 测试和验证
1. 运行所有单元测试
2. 手动测试各种场景
3. 验证JSON输出格式

## 8. 备选方案

### 方案 A: 添加输出格式参数

保留原有文本格式，添加参数支持JSON格式：

```go
outputFormat, _ := getParam(request, "outputFormat") // "text" or "json"
```

### 方案 B: 同时输出两种格式

在结果中同时包含文本和JSON：

```go
{
  "format": "grouped",
  "text": "原始文本格式...",
  "json": { ... }
}
```

## 9. 总结

本修改策略通过以下方式改进静态编译输出：

1. **减少冗余**: 同一规则的多个问题合并显示
2. **结构清晰**: JSON格式便于程序解析
3. **信息完整**: 保留所有必要的诊断信息
4. **易于扩展**: 分组结构便于添加更多元数据

建议采用**阶段1-4**的实施计划，同时考虑添加 `outputFormat` 参数以保持向后兼容性。
