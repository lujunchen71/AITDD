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
  
  "upstreamContractDetail": {...},     // 依赖的上游接口定义（结构化对象）
  "downstreamContractDetail": {...},   // 提供给下游的接口定义（结构化对象）
  
  "tests": [...],                      // 测试用例列表（对象数组）
  "test_result": [...],                // 测试结果证据（字符串数组）
  "codePaths": "[\"path/to/file.py\"]",   // 代码文件路径（JSON字符串）
  "bug_log": "",                        // 执行日志/错误信息（专门记录bug和错误）
  "humanAssistance": "{}",              // 需要人工协助的信息（JSON字符串）
  "issue_details": ""                   // AI在实施计划时发现缺少明确信息时提交的问题详情
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

### upstreamContractDetail 字段规范（结构化格式）

`upstreamContractDetail` 是一个结构化对象，用于描述当前任务依赖的上游接口。

#### 字段定义

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| title | string | 是 | 依赖描述标题，格式为"依赖 *任务名 提供" |
| list | array | 是 | 接口列表，每项包含 label、contract_api 和 from |

#### list 数组项定义

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| label | string | 是 | 接口用途描述 |
| contract_api | string | 是 | 函数签名 |
| from | string | 否 | 标明该 API 来自哪个 task_id，方便 AI 在数据库中查找对应代码位置 |

#### 结构定义

```json
{
  "title": "依赖 *任务提供",
  "list": [
    {
      "label": "接口用途描述",
      "contract_api": "函数签名",
      "from": "task-xxx"  // 新增：标明该 API 来自哪个 task_id
    }
  ]
}
```

#### 示例

```json
{
  "title": "依赖 *task-main-window 提供",
  "list": [
    {
      "label": "将按钮面板添加到布局中",
      "contract_api": "get_main_layout() -> QVBoxLayout",
      "from": "task-main-window"
    },
    {
      "label": "点击按钮时更新表达式显示",
      "contract_api": "setExpression(text: str)",
      "from": "task-main-window"
    }
  ]
}
```

### downstreamContractDetail 字段规范（结构化格式）

`downstreamContractDetail` 是一个结构化对象，用于描述当前任务为下游提供的接口。

#### 字段定义

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| title | string | 是 | 提供描述标题，格式为"为下游任务提供以下接口" |
| list | array | 是 | 接口列表，每项包含 label、contract_api 和 from |

#### list 数组项定义

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| label | string | 是 | 接口用途描述 |
| contract_api | string | 是 | 函数签名或信号定义 |
| from | string | 否 | 标明该 API 来自哪个 task_id（对于 downstream 通常为当前任务） |

#### 结构定义

```json
{
  "title": "为下游任务提供以下接口",
  "list": [
    {
      "label": "接口用途描述",
      "contract_api": "函数签名或信号定义",
      "from": "task-xxx"  // 新增：标明该 API 来自哪个 task_id
    }
  ]
}
```

#### 示例

```json
{
  "title": "为下游任务提供以下接口",
  "list": [
    {
      "label": "按钮点击信号，传递按钮文本",
      "contract_api": "button_clicked: Signal(str)",
      "from": "task-buttons"
    },
    {
      "label": "获取按钮面板组件",
      "contract_api": "get_button_panel() -> QWidget",
      "from": "task-buttons"
    },
    {
      "label": "设置按钮启用状态",
      "contract_api": "set_button_enabled(text: str, enabled: bool)",
      "from": "task-buttons"
    }
  ]
}
```

### tests 字段规范（结构化格式）

`tests` 是一个对象数组，每个对象描述一个测试用例。

#### 字段定义

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| target | string | 是 | 测试目标描述，说明该测试验证什么功能 |
| api | string | 是 | 测试函数名称及签名 |

#### 结构定义

```json
"tests": [
  {
    "target": "测试目标描述",
    "api": "test_function_name()"
  }
]
```

#### 示例

```json
"tests": [
  {
    "target": "验证按钮网格布局是否正确创建4x5结构",
    "api": "test_button_grid_layout()"
  },
  {
    "target": "验证按钮点击信号是否正确发送",
    "api": "test_button_click_signal()"
  },
  {
    "target": "验证所有按钮是否正确初始化并显示",
    "api": "test_button_initialization()"
  }
]
```

### test_result 字段规范（新增）

`test_result` 是一个字符串数组，用于记录测试执行的结果证据。

#### 字段定义

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| test_result | array | 否 | 测试结果证据列表，每项为字符串描述 |

#### 结构定义

```json
"test_result": [
  "测试证据描述1",
  "测试证据描述2"
]
```

#### 示例

```json
"test_result": [
  "test_button_grid_layout() - 通过: 网格布局包含20个按钮，排列为4行5列",
  "test_button_click_signal() - 通过: 点击按钮'5'后收到信号，参数为'5'",
  "test_button_initialization() - 通过: 所有按钮文本正确，样式已应用"
]
```

#### 用途说明

- 记录测试执行的实际情况
- 作为任务完成的质量证据
- 便于后续审核和追溯
- 可包含失败原因和错误信息

### issue_details 字段规范（新增）

`issue_details` 是一个字符串字段，用于 AI 在实施计划时发现缺少明确信息时提交的问题详情。

#### 字段定义

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| issue_details | string | 否 | AI发现信息不足时提交的问题详情 |

#### 示例

```json
{
  "issue_details": "上游*任务定义不清，或者已经找不到定义，需要*任务求助"
}
```

#### 用途说明

- 当 AI 在实施过程中发现上游任务定义不清晰时记录问题
- 帮助追踪需要人工干预或补充信息的任务
- 便于后续优化任务定义的完整性

### bug_log 字段规范

`bug_log` 是一个字符串字段，专门用于记录 bug 和错误日志。

#### 字段定义

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| bug_log | string | 否 | 执行日志/错误信息，专门记录 bug 和错误 |

#### 示例

```json
{
  "bug_log": "TypeError: 'NoneType' object is not callable\n  at button_click_handler line 42"
}
```

#### 用途说明

- 专门记录 bug 和错误日志
- 便于调试和问题追踪
- 记录任务执行过程中的异常信息

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

以下是一个完整的AITDD任务数据JSON文件结构示例：

```json
{
  "project": {
    "name": "项目名称",
    "description": "项目描述"
  },
  "modules": [
    {
      "id": "mod-ui-main",
      "name": "主界面模块",
      "description": "计算器主窗口UI组件",
      "status": "developing",
      "file_path": "src/ui/main_window.py"
    },
    {
      "id": "mod-calc-core",
      "name": "计算核心模块",
      "description": "核心计算引擎",
      "status": "developing",
      "file_path": "src/core/calculator.py"
    }
  ],
  "tasks": [
    {
      "id": "task-main-window",
      "name": "主窗口框架",
      "description": "创建QMainWindow主窗口类",
      "moduleId": "mod-ui-main",
      "status": "completed",
      "prompt": "创建CalculatorMainWindow类继承自QMainWindow。\n\n要求：\n1. 设置窗口标题\n2. 创建中央widget\n\n输出文件：src/ui/main_window.py",
      "upstreamContractDetail": {
        "title": "无上游依赖",
        "list": []
      },
      "downstreamContractDetail": {
        "title": "为下游任务提供以下接口",
        "list": [
          {
            "label": "获取主布局",
            "contract_api": "get_main_layout() -> QVBoxLayout",
            "from": "task-main-window"
          }
        ]
      },
      "tests": [
        {
          "target": "验证窗口创建",
          "api": "test_window_creation()"
        }
      ],
      "test_result": ["test_window_creation() - 通过"],
      "codePaths": "[\"src/ui/main_window.py\"]",
      "bug_log": "",
      "humanAssistance": "{}",
      "issue_details": ""
    },
    {
      "id": "task-buttons",
      "name": "按钮面板",
      "description": "创建计算器按钮网格布局",
      "moduleId": "mod-ui-main",
      "status": "ready",
      "prompt": "创建ButtonPanel类继承自QWidget。\n\n要求：\n1. 使用QGridLayout创建按钮网格\n2. 按钮点击发送button_clicked信号\n\n输出文件：src/ui/buttons.py",
      "upstreamContractDetail": {
        "title": "依赖 *task-main-window 提供",
        "list": [
          {
            "label": "将按钮面板添加到布局中",
            "contract_api": "get_main_layout() -> QVBoxLayout",
            "from": "task-main-window"
          }
        ]
      },
      "downstreamContractDetail": {
        "title": "为下游任务提供以下接口",
        "list": [
          {
            "label": "按钮点击信号",
            "contract_api": "button_clicked: Signal(str)",
            "from": "task-buttons"
          }
        ]
      },
      "tests": [
        {
          "target": "验证按钮布局",
          "api": "test_button_layout()"
        }
      ],
      "test_result": [],
      "codePaths": "[\"src/ui/buttons.py\"]",
      "bug_log": "",
      "humanAssistance": "{}",
      "issue_details": ""
    }
  ],
  "taskDependencies": [
    {
      "id": "dep-001",
      "upstreamTaskId": "task-main-window",
      "downstreamTaskId": "task-buttons",
      "interfaceContract": "get_main_layout() -> QVBoxLayout"
    }
  ]
}
```

### 单个任务示例

如果只需要查看单个任务的详细结构：

```json
{
  "id": "task-buttons",
  "name": "按钮面板",
  "description": "创建计算器按钮网格布局，包含数字按钮和运算符按钮",
  "moduleId": "mod-ui-main",
  "status": "ready",
  "prompt": "创建ButtonPanel类继承自QWidget。\n\n要求：\n1. 使用QGridLayout创建4x5按钮网格\n2. 数字按钮0-9\n3. 运算符按钮：+、-、*、/、=、.\n4. 功能按钮：C(清除)、CE(清除所有)、退格\n5. 按钮点击发送button_clicked信号\n\n输出文件：src/ui/buttons.py",
  "upstreamContractDetail": {
    "title": "依赖 *task-main-window 提供",
    "list": [
      {
        "label": "将按钮面板添加到布局中",
        "contract_api": "get_main_layout() -> QVBoxLayout",
        "from": "task-main-window"
      },
      {
        "label": "点击按钮时更新表达式显示",
        "contract_api": "setExpression(text: str)",
        "from": "task-display"
      }
    ]
  },
  "downstreamContractDetail": {
    "title": "为下游任务提供以下接口",
    "list": [
      {
        "label": "按钮点击信号，传递按钮文本",
        "contract_api": "button_clicked: Signal(str)",
        "from": "task-buttons"
      },
      {
        "label": "获取按钮面板组件",
        "contract_api": "get_button_panel() -> QWidget",
        "from": "task-buttons"
      },
      {
        "label": "设置按钮启用状态",
        "contract_api": "set_button_enabled(text: str, enabled: bool)",
        "from": "task-buttons"
      }
    ]
  },
  "tests": [
    {
      "target": "验证按钮网格布局是否正确创建4x5结构",
      "api": "test_button_grid_layout()"
    },
    {
      "target": "验证按钮点击信号是否正确发送",
      "api": "test_button_click_signal()"
    }
  ],
  "test_result": [],
  "codePaths": "[\"src/ui/buttons.py\"]",
  "bug_log": "",
  "humanAssistance": "{}",
  "issue_details": ""
}
```

## 字段对照表

### 任务字段一览

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | 任务唯一标识，格式：task-{缩写} |
| name | string | 是 | 任务名称 |
| description | string | 是 | 任务详细描述 |
| moduleId | string | 是 | 所属模块ID |
| status | string | 是 | 任务状态 |
| prompt | string | 是 | AI提示词 |
| upstreamContractDetail | object | 是 | 上游依赖接口（结构化） |
| downstreamContractDetail | object | 是 | 下游提供接口（结构化） |
| tests | array | 是 | 测试用例列表（对象数组） |
| test_result | array | 否 | 测试结果证据（字符串数组） |
| codePaths | string | 否 | 代码路径（JSON字符串） |
| bug_log | string | 否 | 执行日志/错误信息（专门记录bug和错误） |
| humanAssistance | string | 否 | 人工协助信息（JSON字符串） |
| issue_details | string | 否 | AI发现信息不足时提交的问题详情 |

### ContractInterfaceItem 字段一览（list 数组项）

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| label | string | 是 | 接口用途描述 |
| contract_api | string | 是 | 函数签名或信号定义 |
| from | string | 否 | 标明该 API 来自哪个 task_id，方便 AI 在数据库中查找对应代码位置 |

## AI生成指南

当AI生成任务数据时，必须遵循以下原则：

1. **完整性**：每个任务必须有完整的prompt、upstreamContractDetail和downstreamContractDetail
2. **清晰性**：接口定义必须包含参数类型和返回类型
3. **可追溯性**：依赖关系必须明确指出依赖哪个任务的哪个接口，使用 from 字段标明来源
4. **可测试性**：每个任务必须有对应的测试用例，包含target和api两个字段
5. **层次性**：任务应该按照依赖顺序排列，基础任务在前
6. **结构化**：upstreamContractDetail和downstreamContractDetail必须使用结构化对象格式

### 依赖关系检查清单

- [ ] 每个依赖的upstreamTaskId都对应一个存在的任务
- [ ] 每个依赖的downstreamTaskId都对应一个存在的任务
- [ ] interfaceContract与upstreamTask的downstreamContractDetail匹配
- [ ] interfaceContract与downstreamTask的upstreamContractDetail匹配
- [ ] 没有循环依赖

### 数据格式检查清单

- [ ] upstreamContractDetail 包含 title 和 list 两个字段
- [ ] downstreamContractDetail 包含 title 和 list 两个字段
- [ ] list 中每项包含 label、contract_api 和可选的 from 字段
- [ ] tests 数组中每项包含 target 和 api 两个字段
- [ ] test_result 为字符串数组（可为空）
- [ ] bug_log 为字符串（可为空）
- [ ] issue_details 为字符串（可为空）
