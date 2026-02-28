# AITDD MCP API 参数说明书

## 数据结构参考

基于 `debug/aitdd_example.json` 示例文件分析，以下是各类型节点的完整字段结构。

---

## 一、Project（项目）字段结构

```json
{
  "name": "PYQT6Calculator",
  "description": "基于PyQt6的科学计算器应用程序"
}
```

### 字段说明

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `name` | string | 是 | 项目名称 |
| `description` | string | 否 | 项目描述 |

---

## 二、Module（模块）字段结构

```json
{
  "name": "主界面模块",
  "description": "计算器主窗口UI组件，包含显示屏和按钮面板",
  "status": "developing",
  "file_path": "src/ui/main_window.py",
  "prompt": "模块开发提示词...",
  "upstreamContractSummary": "依赖核心引擎模块",
  "downstreamContractSummary": "为下游提供UI组件"
}
```

### 字段说明

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `name` | string | 是 | 模块名称 |
| `description` | string | 否 | 模块描述 |
| `status` | string | 否 | 状态：designing/developing/completed |
| `file_path` | string | 否 | 主要文件路径 |
| `prompt` | string | 否 | 模块开发提示词 |
| `upstreamContractSummary` | string | 否 | 上游契约摘要 |
| `downstreamContractSummary` | string | 否 | 下游契约摘要 |

---

## 三、Task（任务）字段结构

```json
{
  "name": "主窗口框架",
  "description": "创建QMainWindow主窗口类，设置窗口属性、布局管理器和基础样式",
  "status": "completed",
  "prompt": "创建CalculatorMainWindow类继承自QMainWindow...\n\n要求：\n1. 设置窗口标题为'科学计算器'\n2. 窗口大小300x500，固定大小\n\n输出文件：src/ui/main_window.py",
  "upstreamContractDetail": {
    "title": "无上游依赖，这是基础组件",
    "list": []
  },
  "downstreamContractDetail": {
    "title": "为下游任务提供以下接口",
    "list": [
      {
        "label": "获取中央widget用于添加子组件",
        "contract_api": "get_central_widget() -> QWidget",
        "from": "task-main-window"
      },
      {
        "label": "获取主布局用于添加其他布局",
        "contract_api": "get_main_layout() -> QVBoxLayout",
        "from": "task-main-window"
      }
    ]
  },
  "tests": [
    {
      "target": "验证窗口标题是否正确设置",
      "api": "test_window_title()"
    },
    {
      "target": "验证窗口大小是否为300x500",
      "api": "test_window_size()"
    }
  ],
  "test_result": [
    "test_window_title() - 通过: 窗口标题为'科学计算器'",
    "test_window_size() - 通过: 窗口大小为300x500"
  ],
  "codePaths": "[\"src/ui/main_window.py\"]",
  "bug_log": "",
  "humanAssistance": "{}",
  "issue_details": ""
}
```

### 字段说明

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `name` | string | 是 | 任务名称 |
| `description` | string | 否 | 任务描述 |
| `status` | string | 否 | 状态：ready/in_progress/completed/blocked |
| `prompt` | string | 否 | 任务开发提示词（可多行） |
| `upstreamContractDetail` | object | 否 | 上游契约详情（JSON对象） |
| `downstreamContractDetail` | object | 否 | 下游契约详情（JSON对象） |
| `tests` | array | 否 | 测试用例数组 |
| `test_result` | array | 否 | 测试结果数组 |
| `codePaths` | string | 否 | 代码路径（JSON字符串格式） |
| `bug_log` | string | 否 | Bug日志 |
| `humanAssistance` | string | 否 | 人工协助信息（JSON字符串格式） |
| `issue_details` | string | 否 | 问题详情 |

---

## 四、特殊字段格式详解

### 4.1 `tests` 字段格式

**正确格式**（JSON数组）：
```json
[
  {
    "target": "验证窗口标题是否正确设置",
    "api": "test_window_title()"
  },
  {
    "target": "验证窗口大小是否为300x500",
    "api": "test_window_size()"
  }
]
```

**字段说明**：
- `target`: 测试目标/描述（string）
- `api`: 测试API/方法名（string）

### 4.2 `upstreamContractDetail` / `downstreamContractDetail` 字段格式

**正确格式**（JSON对象）：
```json
{
  "title": "为下游任务提供以下接口",
  "list": [
    {
      "label": "获取中央widget用于添加子组件",
      "contract_api": "get_central_widget() -> QWidget",
      "from": "task-main-window"
    },
    {
      "label": "获取主布局用于添加其他布局",
      "contract_api": "get_main_layout() -> QVBoxLayout",
      "from": "task-main-window"
    }
  ]
}
```

**字段说明**：
- `title`: 契约标题（string）
- `list`: 契约项数组
  - `label`: 接口描述（string）
  - `contract_api`: API签名（string）
  - `from`: 来源任务ID（string）

### 4.3 `codePaths` 字段格式

**正确格式**（JSON字符串）：
```json
"[\"src/ui/main_window.py\", \"src/ui/display.py\"]"
```

---

## 五、MCP API 使用示例

### 5.1 `query_node` 查询节点

**功能**：通用查询接口，根据 path 自动识别类型（project/module/task）

**参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `path` | string | 是 | 节点 pathName |
| `fields` | array | 否 | 指定返回字段，不填返回所有 |

**示例调用**：
```json
{
  "path": "PYQT6Calculator/主界面模块/主窗口框架",
  "fields": ["tests", "prompt", "downstreamContractDetail"]
}
```

**返回示例**（简洁YAML格式）：
```yaml
tests: 
  - 
    target: 验证窗口标题是否正确设置
    api: test_window_title()
  - 
    target: 验证窗口大小是否为300x500
    api: test_window_size()
prompt: "创建CalculatorMainWindow类继承自QMainWindow..."
downstreamContractDetail: 
  title: 为下游任务提供以下接口
  list: 
    - 
      label: 获取中央widget用于添加子组件
      contract_api: get_central_widget() -> QWidget
      from: task-main-window
```

### 5.2 `modify_node` 修改节点

**功能**：统一修改节点接口，根据 path 自动识别类型，只更新传入的字段

**参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `path` | string | 是 | 节点 pathName |
| `version` | number | 是 | 当前版本号（乐观锁） |
| `data` | object | 是 | 要修改的字段键值对 |

#### 5.2.1 修改 `tests` 字段

**正确调用示例**：
```json
{
  "path": "PYQT6Calculator/主界面模块/主窗口框架",
  "version": 1,
  "data": {
    "tests": [
      {
        "target": "验证窗口标题是否正确设置",
        "api": "test_window_title()"
      },
      {
        "target": "验证窗口大小是否为300x500",
        "api": "test_window_size()"
      },
      {
        "target": "验证中央widget是否存在",
        "api": "test_central_widget_exists()"
      }
    ]
  }
}
```

**错误示例**（不要这样用）：
```json
{
  "data": {
    "tests": "[{\"name\":\"test_constructor\",\"description\":\"测试构造函数\"}]"
  }
}
```
❌ `tests` 的值应该是 JSON 数组对象，而不是 JSON 字符串！

#### 5.2.2 修改 `downstreamContractDetail` 字段

**正确调用示例**：
```json
{
  "path": "PYQT6Calculator/主界面模块/主窗口框架",
  "version": 2,
  "data": {
    "downstreamContractDetail": {
      "title": "为下游任务提供以下接口",
      "list": [
        {
          "label": "获取中央widget",
          "contract_api": "get_central_widget() -> QWidget",
          "from": "task-main-window"
        },
        {
          "label": "获取主布局",
          "contract_api": "get_main_layout() -> QVBoxLayout",
          "from": "task-main-window"
        }
      ]
    }
  }
}
```

#### 5.2.3 修改 `prompt` 字段

**正确调用示例**：
```json
{
  "path": "PYQT6Calculator/主界面模块/主窗口框架",
  "version": 3,
  "data": {
    "prompt": "创建CalculatorMainWindow类继承自QMainWindow。\n\n要求：\n1. 设置窗口标题为'科学计算器'\n2. 窗口大小300x500，固定大小\n3. 设置窗口图标\n\n输出文件：src/ui/main_window.py"
  }
}
```

#### 5.2.4 同时修改多个字段

**正确调用示例**：
```json
{
  "path": "PYQT6Calculator/主界面模块/主窗口框架",
  "version": 4,
  "data": {
    "status": "completed",
    "tests": [
      {
        "target": "验证窗口标题",
        "api": "test_window_title()"
      }
    ],
    "test_result": [
      "test_window_title() - 通过"
    ],
    "codePaths": "[\"src/ui/main_window.py\"]"
  }
}
```

---

## 六、常见错误

### 6.1 JSON 对象传成了字符串

❌ **错误**：
```json
{
  "data": {
    "tests": "[{\"target\":\"测试1\",\"api\":\"test1()\"}]"
  }
}
```

✅ **正确**：
```json
{
  "data": {
    "tests": [
      {
        "target": "测试1",
        "api": "test1()"
      }
    ]
  }
}
```

### 6.2 版本号不匹配

修改节点时必须提供当前正确的版本号，否则会因乐观锁失败。建议：
1. 先用 `query_node` 查询当前 `version`
2. 使用查询到的版本号进行修改

### 6.3 path 格式错误

- Project: `"项目名"` （无斜杠）
- Module: `"项目名/模块名"` （一个斜杠）
- Task: `"项目名/模块名/任务名"` （两个斜杠）

---

## 七、字段与类型对应关系速查表

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `name` | string | 名称 |
| `description` | string | 描述 |
| `status` | string | 状态枚举 |
| `prompt` | string | 提示词（可多行） |
| `file_path` | string | 文件路径 |
| `upstreamContractDetail` | **object** | 上游契约（JSON对象） |
| `downstreamContractDetail` | **object** | 下游契约（JSON对象） |
| `upstreamContractSummary` | string | 上游契约摘要 |
| `downstreamContractSummary` | string | 下游契约摘要 |
| `tests` | **array** | 测试用例数组 |
| `test_result` | **array** | 测试结果数组 |
| `codePaths` | string | 代码路径（JSON字符串） |
| `bug_log` | string | Bug日志 |
| `humanAssistance` | string | 人工协助信息 |
| `issue_details` | string | 问题详情 |
| `version` | number | 版本号（只读，用于乐观锁） |
