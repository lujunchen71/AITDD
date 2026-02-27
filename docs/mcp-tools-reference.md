# AITDD MCP 工具接口参考

本文档列出所有 MCP 工具的输入参数和返回字段，供重构参考。

---

## 一、配置管理工具

### 1.1 init_project

**功能**: 初始化项目配置，获取可用项目列表

**输入参数**: 无

**返回字段**:
| 字段名 | 类型 | 描述 | 来源模型 |
|-------|-----|------|---------|
| id | string | 项目ID | Project.ID |
| name | string | 项目名称 | Project.Name |
| description | string | 项目描述 | Project.Constitution(截取) |

---

### 1.2 get_config

**功能**: 获取当前配置文件内容

**输入参数**: 无

**返回字段**: 返回 `.aitdd/project.json` 的完整 JSON 内容

---

### 1.3 set_project

**功能**: 设置当前项目

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| projectId | string | ✅ | 项目ID |
| projectName | string | ✅ | 项目名称 |

**返回字段**: 成功/失败消息

---

## 二、项目信息工具

### 2.1 get_project_info

**功能**: 获取项目详细信息

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| projectId | string | ❌ | 项目ID，不传则使用配置文件中的ID |
| pathName | string | ❌ | 项目路径名称 |

**返回字段**:
| 字段名 | 类型 | 描述 | 来源模型 |
|-------|-----|------|---------|
| id | string | 项目ID | Project.ID |
| name | string | 项目名称 | Project.Name |
| constitution | string | 项目公约 | Project.Constitution |
| createdAt | number | 创建时间 | Project.CreatedAt |
| updatedAt | number | 更新时间 | Project.UpdatedAt |
| version | number | 版本号 | Project.Version |
| syncStatus | string | 同步状态 | Project.SyncStatus |

---

### 2.2 get_constitution

**功能**: 获取项目公约

**输入参数**: 无（使用配置文件中的项目ID）

**返回字段**:
| 字段名 | 类型 | 描述 | 来源模型 |
|-------|-----|------|---------|
| constitution | string | 项目公约内容 | Project.Constitution |

---

## 三、模块管理工具

### 3.1 get_all_modules

**功能**: 获取所有模块列表

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| projectId | string | ❌ | 项目ID |
| includeStats | boolean | ❌ | 是否包含统计信息 |

**返回字段** (data数组中的每个对象):
| 字段名 | 类型 | 描述 | 来源模型 |
|-------|-----|------|---------|
| id | string | 模块ID | Module.ID |
| name | string | 模块名称 | Module.Name |
| description | string | 模块描述 | Module.Description |
| projectId | string | 项目ID | Module.ProjectID |
| parentId | string | 父模块ID | Module.ParentID |
| status | string | 模块状态 | Module.Status |
| testCoverage | number | 测试覆盖率 | Module.TestCoverage |
| prompt | string | 提示词 | Module.Prompt |
| createdAt | number | 创建时间 | Module.CreatedAt |
| updatedAt | number | 更新时间 | Module.UpdatedAt |
| taskCount | number | 任务数量 | (计算字段) |

---

### 3.2 get_module_overview

**功能**: 获取模块概览

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| moduleId | string | ❌ | 模块ID，与 pathName 二选一 |
| pathName | string | ❌ | 模块路径名称，与 moduleId 二选一 |

**返回字段**: 完整 Module 对象
| 字段名 | 类型 | 描述 | 来源模型 |
|-------|-----|------|---------|
| id | string | 模块ID | Module.ID |
| parentId | string | 父模块ID | Module.ParentID |
| projectId | string | 项目ID | Module.ProjectID |
| name | string | 模块名称 | Module.Name |
| description | string | 描述 | Module.Description |
| prompt | string | 提示词 | Module.Prompt |
| status | string | 状态 | Module.Status |
| testCoverage | number | 测试覆盖率 | Module.TestCoverage |
| upstreamContractSummary | string | 上游契约摘要 | Module.UpstreamContractSummary |
| downstreamContractSummary | string | 下游契约摘要 | Module.DownstreamContractSummary |
| locked | boolean | 是否锁定 | Module.Locked |
| lockedBy | string | 锁定者 | Module.LockedBy |
| lockedAt | number | 锁定时间 | Module.LockedAt |
| lockExpiresAt | number | 锁过期时间 | Module.LockExpiresAt |
| positionX | number | X坐标 | Module.PositionX |
| positionY | number | Y坐标 | Module.PositionY |
| createdAt | number | 创建时间 | Module.CreatedAt |
| updatedAt | number | 更新时间 | Module.UpdatedAt |
| version | number | 版本号 | Module.Version |
| syncStatus | string | 同步状态 | Module.SyncStatus |

---

### 3.3 create_module

**功能**: 创建新模块

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| name | string | ✅ | 模块名称 |
| description | string | ❌ | 模块描述 |
| prompt | string | ❌ | 模块提示词 |
| parentId | string | ❌ | 父模块ID |
| projectId | string | ❌ | 项目ID |

**返回字段**: 完整 Module 对象

---

### 3.4 update_module

**功能**: 更新模块信息

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| moduleId | string | ❌ | 模块ID，与 pathName 二选一 |
| pathName | string | ❌ | 模块路径名称，与 moduleId 二选一 |
| name | string | ❌ | 模块名称 |
| description | string | ❌ | 模块描述 |
| prompt | string | ❌ | 模块提示词 |
| status | string | ❌ | 模块状态 |
| version | number | ✅ | 当前版本号 |

**返回字段**: 更新后的 Module 对象

---

### 3.5 update_module_full

**功能**: 完整更新模块（重新定义所有字段）

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| moduleId | string | ✅ | 模块ID |
| moduleJson | string | ✅ | 完整的模块JSON数据 |

**返回字段**: 更新后的 Module 对象

---

### 3.6 delete_module

**功能**: 删除模块及其所有子任务

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| moduleId | string | ✅ | 模块ID |
| force | boolean | ❌ | 是否强制删除（即使有依赖） |

**返回字段**: 成功/失败消息

---

### 3.7 delete_module_tasks

**功能**: 删除模块的所有任务

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| moduleId | string | ✅ | 模块ID |
| force | boolean | ❌ | 是否强制删除 |

**返回字段**:
| 字段名 | 类型 | 描述 |
|-------|-----|------|
| success | boolean | 是否成功 |
| message | string | 结果消息 |
| deletedCount | number | 删除数量 |
| deletedTasks | array | 已删除的任务列表 |
| errors | array | 错误列表(如有) |

---

### 3.8 get_module_task_ids

**功能**: 获取模块中所有任务ID列表

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| moduleId | string | ✅ | 模块ID |

**返回字段**:
| 字段名 | 类型 | 描述 | 来源模型 |
|-------|-----|------|---------|
| data | array | 任务ID列表 | Task.ID |

---

## 四、任务管理工具

### 4.1 get_module_tasks

**功能**: 获取模块的任务列表

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| moduleId | string | ❌ | 模块ID，与 pathName 二选一 |
| pathName | string | ❌ | 模块路径名称，与 moduleId 二选一 |
| includeContracts | boolean | ❌ | 是否包含契约信息 |

**返回字段** (data数组中的每个对象):
| 字段名 | 类型 | 描述 | 来源模型 |
|-------|-----|------|---------|
| id | string | 任务ID | Task.ID |
| moduleId | string | 模块ID | Task.ModuleID |
| name | string | 任务名称 | Task.Name |
| description | string | 任务描述 | Task.Description |
| status | string | 任务状态 | Task.Status |
| assignee | string | 负责人 | Task.Assignee |
| upstreamContractDetail | string | 上游契约 | Task.UpstreamContractDetail |
| downstreamContractDetail | string | 下游契约 | Task.DownstreamContractDetail |
| prompt | string | 提示词 | Task.Prompt |
| tests | string | 测试用例 | Task.Tests |
| testResult | string | 测试结果 | Task.TestResult |
| bugLog | string | Bug日志 | Task.BugLog |
| codePaths | string | 代码路径 | Task.CodePaths |
| locked | boolean | 是否锁定 | Task.Locked |
| createdAt | number | 创建时间 | Task.CreatedAt |
| updatedAt | number | 更新时间 | Task.UpdatedAt |

---

### 4.2 get_task_detail

**功能**: 获取任务详情

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| taskId | string | ❌ | 任务ID，与 pathName 二选一 |
| pathName | string | ❌ | 任务路径名称，与 taskId 二选一 |

**返回字段**: 完整 Task 对象
| 字段名 | 类型 | 描述 | 来源模型 |
|-------|-----|------|---------|
| id | string | 任务ID | Task.ID |
| moduleId | string | 模块ID | Task.ModuleID |
| name | string | 任务名称 | Task.Name |
| description | string | 描述 | Task.Description |
| status | string | 状态 | Task.Status |
| assignee | string | 负责人 | Task.Assignee |
| upstreamContractDetail | string | 上游契约详情 | Task.UpstreamContractDetail |
| downstreamContractDetail | string | 下游契约详情 | Task.DownstreamContractDetail |
| prompt | string | 提示词 | Task.Prompt |
| tests | string | 测试用例(JSON) | Task.Tests |
| testResult | string | 测试结果(JSON) | Task.TestResult |
| bugLog | string | Bug日志(JSON) | Task.BugLog |
| codePaths | string | 代码路径(JSON) | Task.CodePaths |
| humanAssistance | string | 人工协助(JSON) | Task.HumanAssistance |
| issueDetails | string | 问题详情 | Task.IssueDetails |
| locked | boolean | 是否锁定 | Task.Locked |
| lockedBy | string | 锁定者 | Task.LockedBy |
| lockedAt | number | 锁定时间 | Task.LockedAt |
| lockExpiresAt | number | 锁过期时间 | Task.LockExpiresAt |
| createdAt | number | 创建时间 | Task.CreatedAt |
| updatedAt | number | 更新时间 | Task.UpdatedAt |
| version | number | 版本号 | Task.Version |
| syncStatus | string | 同步状态 | Task.SyncStatus |

---

### 4.3 get_task_contracts

**功能**: 获取任务上下游契约

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| taskId | string | ✅ | 任务ID |
| direction | string | ❌ | 方向: upstream/downstream/both，默认both |

**返回字段**:
| 字段名 | 类型 | 描述 | 来源模型 |
|-------|-----|------|---------|
| task.id | string | 任务ID | Task.ID |
| task.name | string | 任务名称 | Task.Name |
| upstream.title | string | 标题 | (固定值) |
| upstream.contracts | string | 上游契约 | Task.UpstreamContractDetail |
| downstream.title | string | 标题 | (固定值) |
| downstream.contracts | string | 下游契约 | Task.DownstreamContractDetail |
| dependencies | object | 依赖信息 | (来自Dependency表) |

---

### 4.4 create_task

**功能**: 创建新任务

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| moduleId | string | ✅ | 模块ID |
| name | string | ✅ | 任务名称 |
| description | string | ❌ | 任务描述 |
| prompt | string | ❌ | 任务提示词 |
| upstreamContractDetail | string | ❌ | 上游契约详情JSON |
| downstreamContractDetail | string | ❌ | 下游契约详情JSON |
| tests | string | ❌ | 测试用例JSON数组 |
| codePaths | string | ❌ | 代码路径JSON数组 |

**返回字段**: 完整 Task 对象

---

### 4.5 update_task

**功能**: 更新任务信息

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| taskId | string | ❌ | 任务ID，与 pathName 二选一 |
| pathName | string | ❌ | 任务路径名称，与 taskId 二选一 |
| name | string | ❌ | 任务名称 |
| description | string | ❌ | 任务描述 |
| status | string | ❌ | 任务状态 |
| prompt | string | ❌ | 任务提示词 |

**返回字段**: 更新后的 Task 对象

---

### 4.6 update_task_full

**功能**: 完整更新任务（重新定义所有字段）

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| taskId | string | ✅ | 任务ID |
| taskJson | string | ✅ | 完整的任务JSON数据 |

**返回字段**: 更新后的 Task 对象

---

### 4.7 delete_task

**功能**: 删除任务

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| taskId | string | ✅ | 任务ID |

**返回字段**: 成功/失败消息

---

## 五、状态查询工具

### 5.1 get_all_task_status

**功能**: 获取项目中所有任务的状态

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| projectId | string | ❌ | 项目ID |
| status | string | ❌ | 状态过滤 |
| includeLockInfo | boolean | ❌ | 是否包含锁定信息 |

**返回字段** (data数组):
| 字段名 | 类型 | 描述 | 来源模型 |
|-------|-----|------|---------|
| id | string | 任务ID | Task.ID |
| name | string | 任务名称 | Task.Name |
| moduleId | string | 模块ID | Task.ModuleID |
| status | string | 任务状态 | Task.Status |
| locked | boolean | 是否锁定 | Task.Locked (可选) |
| lockedBy | string | 锁定者 | Task.LockedBy (可选) |

---

### 5.2 get_module_task_status

**功能**: 获取模块中所有任务的状态

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| moduleId | string | ❌ | 模块ID，与 pathName 二选一 |
| pathName | string | ❌ | 模块路径名称，与 moduleId 二选一 |
| includeLockInfo | boolean | ❌ | 是否包含锁定信息 |

**返回字段** (data数组):
| 字段名 | 类型 | 描述 | 来源模型 |
|-------|-----|------|---------|
| id | string | 任务ID | Task.ID |
| name | string | 任务名称 | Task.Name |
| status | string | 任务状态 | Task.Status |
| locked | boolean | 是否锁定 | Task.Locked (可选) |

---

### 5.3 get_all_task_code_paths

**功能**: 获取所有任务的代码路径

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| projectId | string | ❌ | 项目ID |

**返回字段** (data数组):
| 字段名 | 类型 | 描述 | 来源模型 |
|-------|-----|------|---------|
| id | string | 任务ID | Task.ID |
| name | string | 任务名称 | Task.Name |
| moduleId | string | 模块ID | Task.ModuleID |
| codePaths | string | 代码路径 | Task.CodePaths |

---

## 六、错误处理工具

### 6.1 get_project_errors

**功能**: 获取项目错误列表

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| projectId | string | ❌ | 项目ID |
| severity | string | ❌ | 严重级别过滤: error/warning/info |
| includeDetails | boolean | ❌ | 是否包含详细信息 |

**返回字段**:
| 字段名 | 类型 | 描述 |
|-------|-----|------|
| projectId | string | 项目ID |
| scanTime | string | 扫描时间 |
| totalTasks | number | 总任务数 |
| errorCount | number | 错误数量 |
| errors | array | 错误列表 |
| errors[].taskId | string | 任务ID |
| errors[].taskName | string | 任务名称 |
| errors[].errors | array | 任务错误列表 |
| errors[].errors[].type | string | 错误类型(issue/bug) |
| errors[].errors[].message | string | 错误消息 |

---

### 6.2 get_module_errors

**功能**: 获取模块错误列表

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| moduleId | string | ✅ | 模块ID |
| severity | string | ❌ | 严重级别过滤 |
| includeDetails | boolean | ❌ | 是否包含详细信息 |

**返回字段**: 同 get_project_errors，但 scope 为模块

---

## 七、检查工具

### 7.1 check_module

**功能**: 根据规则检查模块并生成报告

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| moduleId | string | ✅ | 模块ID |
| rules | array | ❌ | 指定检查规则ID列表 |
| includeDynamic | boolean | ❌ | 是否包含动态检查，默认true |
| format | string | ❌ | 输出格式: json/markdown，默认json |

**返回字段**: 检查报告（结构由规则引擎定义）

---

## 八、锁定工具

### 8.1 lock_resource

**功能**: 锁定资源（模块或任务）

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| resourceType | string | ✅ | 资源类型: module 或 task |
| resourceId | string | ✅ | 资源ID |

**返回字段**:
| 字段名 | 类型 | 描述 |
|-------|-----|------|
| success | boolean | 是否成功 |
| message | string | 结果消息 |
| locked | boolean | 锁定状态 |
| lockedBy | string | 锁定者 |
| lockExpiresAt | number | 锁过期时间 |

---

### 8.2 unlock_resource

**功能**: 解锁资源

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| resourceType | string | ✅ | 资源类型: module 或 task |
| resourceId | string | ✅ | 资源ID |

**返回字段**: 成功/失败消息

---

### 8.3 get_lock_status

**功能**: 查询锁定状态

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| resourceType | string | ✅ | 资源类型: module 或 task |
| resourceId | string | ✅ | 资源ID |

**返回字段**:
| 字段名 | 类型 | 描述 |
|-------|-----|------|
| locked | boolean | 是否锁定 |
| lockedBy | string | 锁定者 |
| lockedAt | number | 锁定时间 |
| lockExpiresAt | number | 锁过期时间 |

---

## 九、依赖管理工具

### 9.1 create_module_dependency

**功能**: 创建模块依赖

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| moduleId | string | ✅ | 模块ID |
| dependsOnModuleId | string | ✅ | 被依赖的模块ID |
| dependencyType | string | ❌ | 依赖类型 |
| contractSummary | string | ❌ | 契约摘要 |

**返回字段**: ModuleDependency 对象

---

### 9.2 get_module_dependencies

**功能**: 获取模块依赖

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| moduleId | string | ✅ | 模块ID |

**返回字段** (data数组):
| 字段名 | 类型 | 描述 | 来源模型 |
|-------|-----|------|---------|
| id | string | 依赖ID | ModuleDependency.ID |
| moduleId | string | 模块ID | ModuleDependency.ModuleID |
| dependsOnModuleId | string | 被依赖模块ID | ModuleDependency.DependsOnModuleID |
| dependencyType | string | 依赖类型 | ModuleDependency.DependencyType |
| contractSummary | string | 契约摘要 | ModuleDependency.ContractSummary |
| dependsOnModule | object | 被依赖模块信息 | Module |

---

### 9.3 create_task_dependency

**功能**: 创建任务依赖

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| upstreamTaskId | string | ✅ | 上游任务ID |
| downstreamTaskId | string | ✅ | 下游任务ID |
| contractSummary | string | ❌ | 契约摘要 |

**返回字段**: Dependency 对象

---

### 9.4 get_task_dependencies

**功能**: 获取任务依赖

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| taskId | string | ✅ | 任务ID |

**返回字段** (data数组):
| 字段名 | 类型 | 描述 | 来源模型 |
|-------|-----|------|---------|
| id | string | 依赖ID | Dependency.ID |
| upstreamTaskId | string | 上游任务ID | Dependency.UpstreamTaskID |
| downstreamTaskId | string | 下游任务ID | Dependency.DownstreamTaskID |
| contractSummary | string | 契约摘要 | Dependency.ContractSummary |

---

## 十、通知工具

### 10.1 send_notification

**功能**: 发送通知

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| toTaskId | string | ✅ | 接收方任务ID |
| type | string | ✅ | 通知类型 |
| title | string | ✅ | 通知标题 |
| message | string | ✅ | 通知内容 |

**返回字段**: Notification 对象

---

### 10.2 read_notifications

**功能**: 读取未读通知

**输入参数**:
| 参数名 | 类型 | 必填 | 描述 |
|-------|-----|------|------|
| toTaskId | string | ❌ | 任务ID(过滤) |

**返回字段** (data数组):
| 字段名 | 类型 | 描述 | 来源模型 |
|-------|-----|------|---------|
| id | string | 通知ID | Notification.ID |
| fromTaskId | string | 发送方任务ID | Notification.FromTaskID |
| toTaskId | string | 接收方任务ID | Notification.ToTaskID |
| type | string | 通知类型 | Notification.Type |
| title | string | 通知标题 | Notification.Title |
| content | string | 通知内容 | Notification.Content |
| read | boolean | 是否已读 | Notification.Read |
| createdAt | number | 创建时间 | Notification.CreatedAt |

---

## 十一、其他工具

### 11.1 open_frontend

**功能**: 打开前端网页

**输入参数**: 无

**返回字段**: 成功消息（包含URL）

---

## 工具分类汇总

| 分类 | 工具数量 | 工具列表 |
|-----|---------|---------|
| 配置管理 | 3 | init_project, get_config, set_project |
| 项目信息 | 2 | get_project_info, get_constitution |
| 模块管理 | 8 | get_all_modules, get_module_overview, create_module, update_module, update_module_full, delete_module, delete_module_tasks, get_module_task_ids |
| 任务管理 | 7 | get_module_tasks, get_task_detail, get_task_contracts, create_task, update_task, update_task_full, delete_task |
| 状态查询 | 3 | get_all_task_status, get_module_task_status, get_all_task_code_paths |
| 错误处理 | 2 | get_project_errors, get_module_errors |
| 检查工具 | 1 | check_module |
| 锁定工具 | 3 | lock_resource, unlock_resource, get_lock_status |
| 依赖管理 | 4 | create_module_dependency, get_module_dependencies, create_task_dependency, get_task_dependencies |
| 通知工具 | 2 | send_notification, read_notifications |
| 其他 | 1 | open_frontend |

**总计**: 36 个 MCP 工具
