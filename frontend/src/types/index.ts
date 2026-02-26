// ==================== 项目类型 ====================

export interface Project {
  id: string;
  name: string;
  constitution?: string;
  createdAt: number;
  updatedAt: number;
  version: number;
  syncStatus: SyncStatus;
}

// ==================== 模块类型 ====================

/** 模块状态枚举 */
export type ModuleStatus = 'designing' | 'developing' | 'completed' | 'deprecated';

/** 模块类型 - 与后端 models.Module 对齐 */
export interface Module {
  id: string;
  parentId?: string | null;
  projectId: string;
  name: string;
  description?: string;
  prompt?: string;
  status: ModuleStatus;
  testCoverage: number;
  upstreamContractSummary?: string;
  downstreamContractSummary?: string;
  locked: boolean;
  lockedBy?: string | null;
  lockedAt?: number | null;
  lockExpiresAt?: number | null;
  position_x?: number;
  position_y?: number;
  position_updated_at?: number;
  createdAt: number;
  updatedAt: number;
  version: number;
  syncStatus: SyncStatus;
  children?: Module[];
}

// ==================== 模块依赖类型 ====================

/** 模块依赖类型 */
export type ModuleDependencyType = 'required' | 'optional' | 'conditional';

/** 模块依赖 */
export interface ModuleDependency {
  id: string;
  moduleId: string;
  dependsOnModuleId: string;
  dependencyType: ModuleDependencyType;
  contractSummary?: string;
  createdAt: number;
  updatedAt: number;
  version: number;
  syncStatus: SyncStatus;
}

/** 模块依赖详情 (包含被依赖模块的信息) */
export interface ModuleDependencyWithModule extends ModuleDependency {
  dependsOnModule?: Module;
}

/** 模块被依赖详情 (包含依赖方模块的信息) */
export interface ModuleDependentWithModule extends ModuleDependency {
  module?: Module;
}

/** 创建模块依赖请求 */
export interface CreateModuleDependencyRequest {
  dependsOnModuleId: string;
  dependencyType?: ModuleDependencyType;
  contractSummary?: string;
}

/** 更新模块依赖请求 */
export interface UpdateModuleDependencyRequest {
  dependencyType?: ModuleDependencyType;
  contractSummary?: string;
}

// ==================== 任务类型 ====================

/** 任务状态枚举 - 与后端 models.Task 对齐 */
export type TaskStatus = 'ready' | 'claimed' | 'in_progress' | 'pending_review' | 'completed' | 'failed' | 'blocked';

/** 契约接口项 - 新格式 */
export interface ContractInterfaceItem {
  label: string;
  contract_api: string;
  from?: string;  // 标明该 API 来自哪个 task_id
}

/** 契约详情 - 新格式 (JSON字典) */
export interface ContractDetail {
  title: string;
  list: ContractInterfaceItem[];
}

/** 测试用例 - 实际数据格式 */
export interface TestItem {
  name: string;
  description: string;
  precondition: string;
  steps: string[];
  expected: string;
}

/** 日志条目 */
export interface LogEntry {
  id: string;
  level: 'info' | 'warn' | 'error' | 'debug';
  message: string;
  timestamp: number;
  metadata?: Record<string, unknown>;
}

/** 人类协助事项 */
export interface HumanAssistanceItem {
  id: string;
  description: string;
  status: 'pending' | 'approved' | 'rejected';
  reviewedBy?: string;
  reviewedAt?: number;
}

export interface HumanAssistance {
  items: HumanAssistanceItem[];
  allApproved: boolean;
}

/** 任务类型 - 与后端 models.Task 对齐 */
export interface Task {
  id: string;
  moduleId: string;
  name: string;
  description?: string;
  status: TaskStatus;
  assignee?: string | null;
  upstreamContractDetail?: string; // JSON string of ContractDetail
  downstreamContractDetail?: string; // JSON string of ContractDetail
  prompt?: string;
  tests?: string; // JSON string of TestItem[]
  testResult?: string; // JSON string of string[] (test evidence strings)
  bugLog?: string; // JSON string of LogEntry[] (原 logs)
  issueDetails?: string; // AI 发现问题时记录
  codePaths?: string; // JSON string of string[]
  humanAssistance?: string; // JSON string of HumanAssistance
  locked: boolean;
  lockedBy?: string | null;
  lockedAt?: number | null;
  lockExpiresAt?: number | null;
  createdAt: number;
  updatedAt: number;
  version: number;
  syncStatus: SyncStatus;
}

// ==================== 任务依赖类型 ====================

/** 任务依赖类型 */
export type TaskDependencyType = 'finish_to_start' | 'start_to_start' | 'finish_to_finish' | 'start_to_finish';

/** 任务依赖状态 */
export type TaskDependencyStatus = 'pending' | 'satisfied' | 'blocked';

/** 任务依赖 */
export interface TaskDependency {
  id: string;
  upstreamTaskId: string;
  downstreamTaskId: string;
  contractSummary?: string;
  status: 'active' | 'suspended' | 'removed';
  createdAt: number;
  updatedAt: number;
  syncStatus: SyncStatus;
}

/** 依赖详情 (包含被依赖任务的信息) */
export interface DependencyWithTask extends TaskDependency {
  upstreamTask?: Task;
}

// ==================== 通知类型 ====================

/** 通知类型 */
export type NotificationType = 'task_completed' | 'contract_changed' | 'block_resolved' | 'assistance_required' | 'custom';

/** 通知 */
export interface Notification {
  id: string;
  fromTaskId?: string | null;
  toTaskId?: string | null;
  type: NotificationType;
  title: string;
  message: string;
  read: boolean;
  readAt?: number | null;
  createdAt: number;
  syncStatus: SyncStatus;
}

// ==================== 变更历史类型 ====================

/** 变更类型 */
export type ChangeType = 'INSERT' | 'UPDATE' | 'DELETE';

/** 变更历史 */
export interface ChangeHistory {
  id: string;
  tableName: string;
  recordId: string;
  changeType: ChangeType;
  oldData?: Record<string, unknown>;
  newData?: Record<string, unknown>;
  changedBy: string;
  changedAt: number;
  syncVersion: number;
}

// ==================== 配置类型 ====================

/** 配置项 */
export interface Config {
  key: string;
  value: string; // JSON string
  updatedAt: number;
}

// ==================== 同步状态类型 ====================

/** 同步状态枚举 */
export type SyncStatus = 'SYNCED' | 'PENDING_UPLOAD' | 'PENDING_DOWNLOAD' | 'CONFLICT';

// ==================== 提示词版本类型 ====================

/** 提示词版本 */
export interface PromptVersion {
  id: string;
  entityType: 'module' | 'task';
  entityId: string;
  version: number;
  prompt: string;
  changeSummary?: string;
  createdBy?: string;
  createdAt: number;
}

// ==================== 锁类型 ====================

/** 锁状态 */
export interface LockStatus {
  resourceType: 'module' | 'task';
  resourceId: string;
  locked: boolean;
  lockedBy?: string | null;
  lockedAt?: number | null;
  lockExpiresAt?: number | null;
}

// ==================== API 响应类型 ====================

/** API 响应 */
export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: Record<string, unknown>;
  };
  timestamp: number;
}

/** 分页响应 */
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

// ==================== WebSocket 类型 ====================

/** WebSocket 消息类型 */
export type WebSocketMessageType = 'connected' | 'notification' | 'task_updated' | 'module_updated' | 'ping' | 'pong';

/** WebSocket 消息 */
export interface WebSocketMessage {
  type: WebSocketMessageType;
  payload: unknown;
  timestamp: string;
}

// ==================== 可视化依赖图类型 ====================

/** 视图模式 */
export type ViewMode = 'list' | 'graph';

/** 端口数据（用于模块间依赖） */
export interface PortData {
  id: string;
  taskId: string;
  taskName: string;
  direction: 'input' | 'output';
  connectedModules: string[];
}

/** 模块节点数据（React Flow） */
export interface ModuleNodeData {
  id: string;
  label: string;
  status: ModuleStatus;
  collapsed: boolean;
  tasks: Task[];
  inputPorts: PortData[];
  outputPorts: PortData[];
}

/** 任务节点数据（React Flow） */
export interface TaskNodeData {
  id: string;
  moduleId: string;
  label: string;
  status: TaskStatus;
  upstreamContractDetail?: ContractDetail;
  downstreamContractDetail?: ContractDetail;
  dependencies: TaskDependency[];
}

/** 图表数据 */
export interface ProjectGraphData {
  modules: Module[];
  tasks: Task[];
  taskDependencies: TaskDependency[];
  moduleDependencies: ModuleDependency[];
}

/** 模块端口响应 */
export interface ModulePortsResponse {
  inputPorts: PortData[];
  outputPorts: PortData[];
}
