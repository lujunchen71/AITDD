# AITDD 任务数据格式规范

本文档定义了AITDD系统中项目和任务的数据格式规范，用于指导AI生成完整、详细的任务依赖关系。

## 数据结构

### 项目结构

```json
{
  "project": {
    "name": "项目名称",
    "description": "项目描述"
  },
  "modules": [...],
  "tasks": [...],
  "taskDependencies": [...]
}
```

## 模块定义

每个模块必须包含以下字段：

```json
{
  "id": "mod-xxx",           // 模块ID，格式：mod-{模块名缩写}
  "name": "模块名称",          // 人类可读的模块名称
  "description": "模块描述",   // 详细描述模块的职责
  "status": "developing",    // 状态：designing/developing/completed/deprecated
  "file_path": "path/to/module"  // 模块主要文件路径
}
```

## 任务定义（重要）

每个任务必须包含完整的字段：

```json
{
  "id": "task-xxx",                    // 任务ID，格式：task-{任务名缩写}
  "name": "任务名称",                   // 简短的任务名称
  "description": "任务描述",            // 详细描述任务要完成的工作
  "moduleId": "mod-xxx",               // 所属模块ID
  "status": "ready",                   // 状态：ready/claimed/in_progress/pending_review/completed/failed/blocked
  
  // ===== 重要字段（必须详细填写）=====
  
  "prompt": "详细的AI提示词...",        // 给AI的详细指令，包含具体要求
  
  "upstreamContractDetail": "依赖的上游接口定义...",  // 依赖哪些上游任务的哪些接口
  "downstreamContractDetail": "提供给下游的接口定义...", // 为下游任务提供哪些接口
  
  "tests": "[\"test1()\", \"test2()\"]",  // 测试用例列表（JSON字符串）
  "codePaths": "[\"path/to/file.py\"]",   // 代码文件路径（JSON字符串）
  "logs": "",                           // 执行日志/错误信息
  "humanAssistance": "{}"               // 需要人工协助的信息（JSON字符串）
}
```

### prompt 字段规范

提示词必须包含：
1. 创建的类名和继承关系
2. 具体的功能要求（列表形式）
3. 输出文件路径

示例：
```
创建ButtonPanel类继承自QWidget。

要求：
1. 使用QGridLayout创建4x5按钮网格
2. 数字按钮0-9
3. 运算符按钮：+、-、*、/、=、.
4. 功能按钮：C(清除)、CE(清除所有)、退格
5. 按钮点击发送button_clicked信号
6. 按钮样式：圆角、渐变背景、悬停效果

输出文件：src/ui/buttons.py
```

### upstreamContractDetail 字段规范

必须明确说明：
1. 依赖哪个上游任务
2. 使用该任务的哪些接口
3. 接口的参数和返回类型

示例：
```
依赖 task-main-window 提供：
1. get_main_layout() -> QVBoxLayout: 将按钮面板添加到布局中

依赖 task-display 提供：
1. setExpression(text: str): 点击按钮时更新表达式显示
```

### downstreamContractDetail 字段规范

必须明确说明：
1. 为下游任务提供哪些接口
2. 接口的参数和返回类型
3. 信号定义（如果有）

示例：
```
提供以下接口给下游任务使用：
1. button_clicked: Signal(str) - 按钮点击信号，传递按钮文本
2. get_button_panel() -> QWidget: 获取按钮面板组件
3. set_button_enabled(text: str, enabled: bool): 设置按钮启用状态
```

## 任务依赖定义

每个依赖关系必须包含接口契约：

```json
{
  "id": "dep-xxx",                      // 依赖ID
  "upstreamTaskId": "task-upstream",    // 上游任务ID（被依赖的任务）
  "downstreamTaskId": "task-downstream", // 下游任务ID（依赖者）
  "interfaceContract": "具体的接口定义"   // 上下游之间传递的接口
}
```

### interfaceContract 示例

```
get_main_layout() -> QVBoxLayout
calculate(expression: str) -> str
mode_changed: Signal(str)
get_history(limit: int, offset: int) -> List[Record]
```

## 状态说明

### 模块状态
- `designing`: 设计中
- `developing`: 开发中
- `completed`: 已完成
- `deprecated`: 已废弃

### 任务状态
- `ready`: 准备就绪，可以开始
- `claimed`: 已认领
- `in_progress`: 进行中
- `pending_review`: 等待审核
- `completed`: 已完成
- `failed`: 失败
- `blocked`: 被阻塞

## 完整示例

```json
{
  "id": "task-buttons",
  "name": "按钮面板",
  "description": "创建计算器按钮网格布局，包含数字按钮和运算符按钮",
  "moduleId": "mod-ui-main",
  "status": "ready",
  "prompt": "创建ButtonPanel类继承自QWidget。\n\n要求：\n1. 使用QGridLayout创建4x5按钮网格\n2. 数字按钮0-9\n3. 运算符按钮：+、-、*、/、=、.\n4. 功能按钮：C(清除)、CE(清除所有)、退格\n5. 按钮点击发送button_clicked信号\n\n输出文件：src/ui/buttons.py",
  "upstreamContractDetail": "依赖 task-main-window 提供：\n1. get_main_layout() -> QVBoxLayout: 将按钮面板添加到布局中\n\n依赖 task-display 提供：\n1. setExpression(text: str): 点击按钮时更新表达式显示",
  "downstreamContractDetail": "提供以下接口给下游任务使用：\n1. button_clicked: Signal(str) - 按钮点击信号\n2. get_button_panel() -> QWidget: 获取按钮面板组件\n3. set_button_enabled(text: str, enabled: bool): 设置按钮启用状态",
  "tests": "[\"test_button_grid_layout()\", \"test_button_click_signal()\"]",
  "codePaths": "[\"src/ui/buttons.py\"]",
  "logs": "",
  "humanAssistance": "{}"
}
```

## AI生成指南

当AI生成任务数据时，必须遵循以下原则：

1. **完整性**：每个任务必须有完整的prompt、upstreamContractDetail和downstreamContractDetail
2. **清晰性**：接口定义必须包含参数类型和返回类型
3. **可追溯性**：依赖关系必须明确指出依赖哪个任务的哪个接口
4. **可测试性**：每个任务必须有对应的测试用例
5. **层次性**：任务应该按照依赖顺序排列，基础任务在前

### 依赖关系检查清单

- [ ] 每个依赖的upstreamTaskId都对应一个存在的任务
- [ ] 每个依赖的downstreamTaskId都对应一个存在的任务
- [ ] interfaceContract与upstreamTask的downstreamContractDetail匹配
- [ ] interfaceContract与downstreamTask的upstreamContractDetail匹配
- [ ] 没有循环依赖
