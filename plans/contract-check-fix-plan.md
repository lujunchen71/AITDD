# AITDD 契约检查逻辑修复方案

## 1. 问题概述

用户指出当前的契约上下游检查和模块孤立检查逻辑存在错误。核心问题是：

1. **错误地把 `from` 当成路径使用** - `from` 字段是契约标识的一部分，不是任务路径
2. **契约匹配逻辑不完整** - 应该检查 `contract_api`、`from`、`label` 三个字段都相等才算契约匹配
3. **应该使用双重 for 循环算法** - 循环所有模块中所有任务中所有契约来检查

## 2. 契约数据结构说明

### 2.1 契约条目结构 (ContractDetailItem)

```go
type ContractDetailItem struct {
    Label       string `json:"label"`        // 契约标签/名称
    ContractAPI string `json:"contract_api"` // 契约API/接口
    From        string `json:"from"`         // 契约来源标识（不是路径！）
}
```

### 2.2 契约详情结构 (ContractDetail)

```go
type ContractDetail struct {
    Title string                `json:"title"` // 契约标题
    List  []ContractDetailItem  `json:"list"`  // 契约条目列表
}
```

### 2.3 关键理解

- **`from` 字段是契约标识的一部分**，用于标识这条契约来自哪个任务
- **`from` 不是任务路径**，不应该用于 `taskToModule` 或 `taskNameToPath` 映射查找
- **契约匹配**：两个契约条目相等 = `label` 相同 + `contract_api` 相同 + `from` 相同

## 3. 现有逻辑问题分析

### 3.1 `checkContractMatchInUpstream` 函数 (行 2471-2496)

**文件**: [`backend/internal/mcp/tools_other.go:2471`](backend/internal/mcp/tools_other.go:2471)

**现有逻辑**:
```go
func checkContractMatchInUpstream(downstreamTask map[string]interface{}, item ContractDetailItem, currentTaskPathName string) bool {
    // ... 解析上游契约 ...
    for _, upstreamItem := range upstreamContract.List {
        if item.Label == upstreamItem.Label &&
            item.ContractAPI == upstreamItem.ContractAPI &&
            upstreamItem.From == currentTaskPathName {  // ❌ 错误：把 from 当作路径比较
            return true
        }
    }
    return false
}
```

**问题**:
- 第三个条件 `upstreamItem.From == currentTaskPathName` 错误地把 `from` 当作任务路径
- `from` 是契约标识的一部分，不是路径

**正确逻辑**:
```go
func checkContractMatchInUpstream(downstreamTask map[string]interface{}, item ContractDetailItem) bool {
    upstreamContractStr, _ := downstreamTask["upstreamContractDetail"].(string)
    if upstreamContractStr == "" {
        return false
    }

    var upstreamContract ContractDetail
    if err := json.Unmarshal([]byte(upstreamContractStr), &upstreamContract); err != nil {
        return false
    }

    // 三字段完全匹配
    for _, upstreamItem := range upstreamContract.List {
        if item.Label == upstreamItem.Label &&
            item.ContractAPI == upstreamItem.ContractAPI &&
            item.From == upstreamItem.From {
            return true
        }
    }
    return false
}
```

### 3.2 `checkModuleCircularReference` 函数 (行 3345-3442)

**文件**: [`backend/internal/mcp/tools_other.go:3345`](backend/internal/mcp/tools_other.go:3345)

**现有逻辑**:
```go
func (s *MCPServer) checkModuleCircularReference(allTasks []map[string]interface{}, modules []map[string]interface{}, result *CompileStaticResult) {
    // ...
    for _, task := range allTasks {
        // ...
        upstreamContract := extractContractDetailFromTask(task, "upstreamContractDetail")
        for _, item := range upstreamContract.List {
            upstreamTaskPath := item.From  // ❌ 错误：把 from 当作路径
            if upstreamTaskPath == "" {
                continue
            }
            // 查找上游任务所属模块
            upstreamModulePath := findModuleByTaskIdentifier(upstreamTaskPath, taskToModule, taskNameToPath, currentModulePath)
            // ...
        }
    }
}
```

**问题**:
- `upstreamTaskPath := item.From` 错误地把 `from` 当作任务路径
- 调用 `findModuleByTaskIdentifier` 试图根据 `from` 值查找模块路径

**正确逻辑**:
应该通过契约三字段匹配来找到提供该契约的任务，进而确定模块间的依赖关系。

```go
func (s *MCPServer) checkModuleCircularReference(allTasks []map[string]interface{}, modules []map[string]interface{}, result *CompileStaticResult) {
    if len(modules) < 2 {
        return
    }

    // 构建任务路径到模块路径的映射
    taskToModule := make(map[string]string)
    for _, task := range allTasks {
        taskPathName, _ := task["pathName"].(string)
        if taskPathName == "" {
            continue
        }
        parts := strings.Split(taskPathName, "/")
        if len(parts) >= 2 {
            modulePath := parts[0] + "/" + parts[1]
            taskToModule[taskPathName] = modulePath
        }
    }

    // 契约匹配函数：三字段完全匹配
    contractsMatch := func(c1, c2 ContractDetailItem) bool {
        return c1.ContractAPI == c2.ContractAPI &&
            c1.From == c2.From &&
            c1.Label == c2.Label
    }

    // 记录模块间的引用关系
    // key: "源模块路径->目标模块路径"
    moduleReferences := make(map[string][]CrossModuleReference)

    // 双重 for 循环：遍历所有任务的上游契约，在其他任务的下游契约中找匹配
    for _, task := range allTasks {
        taskPathName, _ := task["pathName"].(string)
        currentModulePath := taskToModule[taskPathName]
        if currentModulePath == "" {
            continue
        }

        upstreamContract := extractContractDetailFromTask(task, "upstreamContractDetail")
        for _, upstreamItem := range upstreamContract.List {
            // 在所有任务中查找下游契约有匹配的任务
            for _, otherTask := range allTasks {
                otherTaskPathName, _ := otherTask["pathName"].(string)
                otherModulePath := taskToModule[otherTaskPathName]
                if otherModulePath == "" || otherModulePath == currentModulePath {
                    continue
                }

                downstreamContract := extractContractDetailFromTask(otherTask, "downstreamContractDetail")
                for _, downstreamItem := range downstreamContract.List {
                    if contractsMatch(upstreamItem, downstreamItem) {
                        // 找到跨模块契约匹配，记录引用关系
                        // 方向：otherModule（提供方）→ currentModule（消费方）
                        key := otherModulePath + "->" + currentModulePath
                        moduleReferences[key] = append(moduleReferences[key], CrossModuleReference{
                            SourceTask: otherTaskPathName,
                            TargetTask: taskPathName,
                        })
                    }
                }
            }
        }
    }

    // 检测循环引用（后续逻辑不变）
    // ...
}
```

### 3.3 `checkOrphanTask` 函数 (行 3468-3544)

**文件**: [`backend/internal/mcp/tools_other.go:3468`](backend/internal/mcp/tools_other.go:3468)

**现有逻辑**:
```go
func (s *MCPServer) checkOrphanTask(allTasks []map[string]interface{}, result *CompileStaticResult) {
    // ...
    for _, task := range allTasks {
        // ...
        upstreamContract := extractContractDetailFromTask(task, "upstreamContractDetail")
        for _, item := range upstreamContract.List {
            if item.From != "" {
                // ❌ 错误：把 item.From 当作任务路径
                referencedAsUpstream[item.From] = true
            }
        }
        // ...
    }
}
```

**问题**:
- `referencedAsUpstream[item.From] = true` 错误地把 `from` 当作任务路径记录
- `from` 是契约标识的一部分，不是任务路径

**正确逻辑**:
应该通过契约匹配来判断任务之间是否有依赖关系。

```go
func (s *MCPServer) checkOrphanTask(allTasks []map[string]interface{}, result *CompileStaticResult) {
    if len(allTasks) <= 1 {
        return
    }

    // 契约匹配函数：三字段完全匹配
    contractsMatch := func(c1, c2 ContractDetailItem) bool {
        return c1.ContractAPI == c2.ContractAPI &&
            c1.From == c2.From &&
            c1.Label == c2.Label
    }

    // 记录每个任务是否有依赖关系
    taskHasDependency := make(map[string]bool)
    for _, task := range allTasks {
        taskPathName, _ := task["pathName"].(string)
        if taskPathName != "" {
            taskHasDependency[taskPathName] = false
        }
    }

    // 双重 for 循环：检查任务间的契约匹配
    for _, taskA := range allTasks {
        taskAPathName, _ := taskA["pathName"].(string)
        if taskAPathName == "" {
            continue
        }

        // 检查 taskA 的上游契约是否与其他任务的下游契约匹配
        upstreamContractA := extractContractDetailFromTask(taskA, "upstreamContractDetail")
        if len(upstreamContractA.List) > 0 {
            // 跳过首个任务
            if upstreamContractA.Title != "start" {
                taskHasDependency[taskAPathName] = true
            }
        }

        // 检查 taskA 的下游契约是否与其他任务的上游契约匹配
        downstreamContractA := extractContractDetailFromTask(taskA, "downstreamContractDetail")
        if len(downstreamContractA.List) > 0 {
            // 跳过末尾任务
            if downstreamContractA.Title != "end" {
                taskHasDependency[taskAPathName] = true
            }
        }

        // 通过契约匹配确认依赖关系
        for _, taskB := range allTasks {
            taskBPathName, _ := taskB["pathName"].(string)
            if taskBPathName == "" || taskBPathName == taskAPathName {
                continue
            }

            // taskA 的上游契约与 taskB 的下游契约匹配
            for _, upstreamItem := range upstreamContractA.List {
                downstreamContractB := extractContractDetailFromTask(taskB, "downstreamContractDetail")
                for _, downstreamItem := range downstreamContractB.List {
                    if contractsMatch(upstreamItem, downstreamItem) {
                        taskHasDependency[taskAPathName] = true
                        taskHasDependency[taskBPathName] = true
                    }
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

        if !taskHasDependency[taskPathName] {
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
```

### 3.4 `findModuleByTaskIdentifier` 函数 (行 3445-3463)

**文件**: [`backend/internal/mcp/tools_other.go:3445`](backend/internal/mcp/tools_other.go:3445)

**现有逻辑**:
```go
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
```

**问题**:
- 这个函数的设计假设 `from` 是任务路径或任务名称
- 但实际上 `from` 是契约标识的一部分，不应该用于路径查找

**修复方案**:
- **废弃此函数**，不再使用
- 改用契约三字段匹配来建立任务间的依赖关系

### 3.5 `findTaskPathByName` 函数 (行 2403-2429)

**文件**: [`backend/internal/mcp/tools_other.go:2403`](backend/internal/mcp/tools_other.go:2403)

**现有逻辑**:
```go
func findTaskPathByName(taskMap map[string]map[string]interface{}, from string, currentTaskPath string) string {
    // 首先尝试直接匹配 pathName
    if _, exists := taskMap[from]; exists {
        return from
    }

    // 尝试在同一模块下查找任务名称
    currentParts := strings.Split(currentTaskPath, "/")
    if len(currentParts) >= 2 {
        modulePath := currentParts[0] + "/" + currentParts[1]
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
```

**问题**:
- 这个函数的设计假设 `from` 是任务路径或任务名称
- 但实际上 `from` 是契约标识的一部分

**修复方案**:
- **废弃此函数**，不再使用
- 改用契约三字段匹配来建立任务间的依赖关系

## 4. 需要修改的函数列表

| 函数名 | 文件 | 行号 | 修改类型 | 优先级 |
|--------|------|------|----------|--------|
| `checkContractMatchInUpstream` | tools_other.go | 2471-2496 | 修改 | 高 |
| `checkModuleCircularReference` | tools_other.go | 3345-3442 | 重写 | 高 |
| `checkOrphanTask` | tools_other.go | 3468-3544 | 重写 | 高 |
| `findModuleByTaskIdentifier` | tools_other.go | 3445-3463 | 废弃 | 中 |
| `findTaskPathByName` | tools_other.go | 2403-2429 | 废弃 | 中 |

## 5. 修改方案详解

### 5.1 统一的契约匹配函数

首先定义一个统一的契约匹配函数，确保所有地方使用相同的匹配逻辑：

```go
// contractsMatch 检查两个契约条目是否匹配
// 契约匹配条件：label、contract_api、from 三个字段完全相同
func contractsMatch(c1, c2 ContractDetailItem) bool {
    return c1.Label == c2.Label &&
        c1.ContractAPI == c2.ContractAPI &&
        c1.From == c2.From
}
```

### 5.2 修改 `checkContractMatchInUpstream`

**修改前**:
```go
func checkContractMatchInUpstream(downstreamTask map[string]interface{}, item ContractDetailItem, currentTaskPathName string) bool {
    // ...
    for _, upstreamItem := range upstreamContract.List {
        if item.Label == upstreamItem.Label &&
            item.ContractAPI == upstreamItem.ContractAPI &&
            upstreamItem.From == currentTaskPathName {  // ❌ 错误
            return true
        }
    }
    return false
}
```

**修改后**:
```go
func checkContractMatchInUpstream(downstreamTask map[string]interface{}, item ContractDetailItem) bool {
    upstreamContractStr, _ := downstreamTask["upstreamContractDetail"].(string)
    if upstreamContractStr == "" {
        return false
    }

    var upstreamContract ContractDetail
    if err := json.Unmarshal([]byte(upstreamContractStr), &upstreamContract); err != nil {
        return false
    }

    for _, upstreamItem := range upstreamContract.List {
        if contractsMatch(item, upstreamItem) {  // ✅ 使用三字段匹配
            return true
        }
    }
    return false
}
```

### 5.3 修改 `checkContractConsistency` 中的调用

由于 `checkContractMatchInUpstream` 的签名改变了，需要修改调用处：

**修改前** (行 2609):
```go
if checkContractMatchInUpstream(potentialDownstreamTask, downstreamItem, taskPathName) {
```

**修改后**:
```go
if checkContractMatchInUpstream(potentialDownstreamTask, downstreamItem) {
```

### 5.4 重写 `checkModuleCircularReference`

使用双重 for 循环算法，基于契约三字段匹配来建立模块间的依赖关系：

```go
func (s *MCPServer) checkModuleCircularReference(allTasks []map[string]interface{}, modules []map[string]interface{}, result *CompileStaticResult) {
    if len(modules) < 2 {
        return
    }

    // 构建任务路径到模块路径的映射
    taskToModule := make(map[string]string)
    for _, task := range allTasks {
        taskPathName, _ := task["pathName"].(string)
        if taskPathName == "" {
            continue
        }
        parts := strings.Split(taskPathName, "/")
        if len(parts) >= 2 {
            modulePath := parts[0] + "/" + parts[1]
            taskToModule[taskPathName] = modulePath
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
    moduleReferences := make(map[string][]CrossModuleReference)

    // 双重 for 循环：遍历所有任务对，检查契约匹配
    for _, taskA := range allTasks {
        taskAPathName, _ := taskA["pathName"].(string)
        moduleA := taskToModule[taskAPathName]
        if moduleA == "" {
            continue
        }

        // 获取 taskA 的上游契约
        upstreamContractA := extractContractDetailFromTask(taskA, "upstreamContractDetail")

        // 遍历其他任务
        for _, taskB := range allTasks {
            taskBPathName, _ := taskB["pathName"].(string)
            moduleB := taskToModule[taskBPathName]
            if moduleB == "" || moduleB == moduleA {
                continue
            }

            // 获取 taskB 的下游契约
            downstreamContractB := extractContractDetailFromTask(taskB, "downstreamContractDetail")

            // 检查 taskA 的上游契约是否与 taskB 的下游契约匹配
            for _, upstreamItem := range upstreamContractA.List {
                for _, downstreamItem := range downstreamContractB.List {
                    if contractsMatch(upstreamItem, downstreamItem) {
                        // taskB 提供的下游契约被 taskA 的上游契约消费
                        // 方向：moduleB -> moduleA
                        key := moduleB + "->" + moduleA
                        moduleReferences[key] = append(moduleReferences[key], CrossModuleReference{
                            SourceTask: taskBPathName,
                            TargetTask: taskAPathName,
                        })
                    }
                }
            }
        }
    }

    // 检测循环引用
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
```

### 5.5 重写 `checkOrphanTask`

使用双重 for 循环算法，基于契约匹配来判断任务是否孤立：

```go
func (s *MCPServer) checkOrphanTask(allTasks []map[string]interface{}, result *CompileStaticResult) {
    if len(allTasks) <= 1 {
        return
    }

    // 记录每个任务是否有依赖关系
    taskHasDependency := make(map[string]bool)
    taskPathToName := make(map[string]string)
    
    for _, task := range allTasks {
        taskPathName, _ := task["pathName"].(string)
        taskName, _ := task["name"].(string)
        if taskPathName != "" {
            taskHasDependency[taskPathName] = false
            taskPathToName[taskPathName] = taskName
        }
    }

    // 双重 for 循环：检查任务间的契约匹配
    for _, taskA := range allTasks {
        taskAPathName, _ := taskA["pathName"].(string)
        if taskAPathName == "" {
            continue
        }

        upstreamContractA := extractContractDetailFromTask(taskA, "upstreamContractDetail")
        downstreamContractA := extractContractDetailFromTask(taskA, "downstreamContractDetail")

        // 跳过首个任务（title 为 "start"）
        if upstreamContractA.Title == "start" {
            taskHasDependency[taskAPathName] = true
            continue
        }

        // 跳过末尾任务（title 为 "end"）
        if downstreamContractA.Title == "end" {
            taskHasDependency[taskAPathName] = true
            continue
        }

        // 检查 taskA 的契约是否与其他任务的契约匹配
        for _, taskB := range allTasks {
            taskBPathName, _ := taskB["pathName"].(string)
            if taskBPathName == "" || taskBPathName == taskAPathName {
                continue
            }

            upstreamContractB := extractContractDetailFromTask(taskB, "upstreamContractDetail")
            downstreamContractB := extractContractDetailFromTask(taskB, "downstreamContractDetail")

            // taskA 的上游契约与 taskB 的下游契约匹配
            for _, itemA := range upstreamContractA.List {
                for _, itemB := range downstreamContractB.List {
                    if contractsMatch(itemA, itemB) {
                        taskHasDependency[taskAPathName] = true
                        taskHasDependency[taskBPathName] = true
                    }
                }
            }

            // taskA 的下游契约与 taskB 的上游契约匹配
            for _, itemA := range downstreamContractA.List {
                for _, itemB := range upstreamContractB.List {
                    if contractsMatch(itemA, itemB) {
                        taskHasDependency[taskAPathName] = true
                        taskHasDependency[taskBPathName] = true
                    }
                }
            }
        }
    }

    // 检查每个任务是否孤立
    for taskPathName, hasDependency := range taskHasDependency {
        if !hasDependency {
            taskName := taskPathToName[taskPathName]
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
```

## 6. 测试用例

### 6.1 契约匹配测试

```go
func TestContractsMatch(t *testing.T) {
    tests := []struct {
        name     string
        c1       ContractDetailItem
        c2       ContractDetailItem
        expected bool
    }{
        {
            name: "三字段完全相同",
            c1:   ContractDetailItem{Label: "用户数据", ContractAPI: "UserService.getData", From: "用户管理模块"},
            c2:   ContractDetailItem{Label: "用户数据", ContractAPI: "UserService.getData", From: "用户管理模块"},
            expected: true,
        },
        {
            name: "label不同",
            c1:   ContractDetailItem{Label: "用户数据", ContractAPI: "UserService.getData", From: "用户管理模块"},
            c2:   ContractDetailItem{Label: "订单数据", ContractAPI: "UserService.getData", From: "用户管理模块"},
            expected: false,
        },
        {
            name: "contract_api不同",
            c1:   ContractDetailItem{Label: "用户数据", ContractAPI: "UserService.getData", From: "用户管理模块"},
            c2:   ContractDetailItem{Label: "用户数据", ContractAPI: "UserService.getData2", From: "用户管理模块"},
            expected: false,
        },
        {
            name: "from不同",
            c1:   ContractDetailItem{Label: "用户数据", ContractAPI: "UserService.getData", From: "用户管理模块"},
            c2:   ContractDetailItem{Label: "用户数据", ContractAPI: "UserService.getData", From: "订单管理模块"},
            expected: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := contractsMatch(tt.c1, tt.c2)
            if result != tt.expected {
                t.Errorf("contractsMatch() = %v, expected %v", result, tt.expected)
            }
        })
    }
}
```

## 7. 实施步骤

1. **添加统一契约匹配函数** `contractsMatch`
2. **修改** `checkContractMatchInUpstream` 函数，移除 `currentTaskPathName` 参数
3. **更新** `checkContractConsistency` 中对 `checkContractMatchInUpstream` 的调用
4. **重写** `checkModuleCircularReference` 函数
5. **重写** `checkOrphanTask` 函数
6. **废弃** `findModuleByTaskIdentifier` 和 `findTaskPathByName` 函数（标记为 deprecated 或删除）
7. **运行测试** 确保修改正确

## 8. 风险评估

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 修改后契约匹配逻辑变化 | 可能导致现有项目编译失败 | 提供迁移指南，说明契约定义的正确方式 |
| 性能影响 | 双重 for 循环可能影响性能 | 对于大型项目，考虑优化数据结构 |
| 向后兼容性 | 旧数据可能不符合新逻辑 | 提供数据迁移脚本 |

## 9. 总结

本次修复的核心是纠正对 `from` 字段的错误理解：

- **错误理解**：`from` 是任务路径，可以用于查找任务
- **正确理解**：`from` 是契约标识的一部分，只能用于契约匹配

修复后的逻辑将使用统一的三字段匹配算法（`label` + `contract_api` + `from`）来判断契约是否匹配，从而正确建立任务间的依赖关系。
