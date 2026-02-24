// 项目类型
export interface Project {
  id: string;
  name: string;
  description: string;
  constitution?: string;
  createdAt: string;
  updatedAt: string;
}

// 模块类型
export interface Module {
  id: string;
  projectId: string;
  parentId?: string;
  name: string;
  description?: string;
  path: string;
  order: number;
  createdAt: string;
  updatedAt: string;
  children?: Module[];
}

// 模块依赖类型
export type ModuleDependencyType = 'required' | 'optional' | 'conditional';

// 模块依赖
export interface ModuleDependency {
  id: string;
  moduleId: string;
  dependsOnModuleId: string;
  dependencyType: ModuleDependencyType;
  contractSummary?: string;
  createdAt: number;
  updatedAt: number;
  version: number;
  syncStatus: string;
}

// 模块依赖详情（包含被依赖模块的信息）
export interface ModuleDependencyWithModule extends ModuleDependency {
  dependsOnModule?: Module;
}

// 模块被依赖详情（包含依赖方模块的信息）
export interface ModuleDependentWithModule extends ModuleDependency {
  module?: Module;
}

// 创建模块依赖请求
export interface CreateModuleDependencyRequest {
  dependsOnModuleId: string;
  dependencyType?: ModuleDependencyType;
  contractSummary?: string;
}

// 更新模块依赖请求
export interface UpdateModuleDependencyRequest {
  dependencyType?: ModuleDependencyType;
  contractSummary?: string;
}

// 任务状态
export type TaskStatus = 'pending' | 'in_progress' | 'completed' | 'blocked';

// 任务优先级
export type TaskPriority = 'low' | 'medium' | 'high' | 'urgent';

// 任务类型
export interface Task {
  id: string;
  projectId: string;
  moduleId?: string;
  title: string;
  description?: string;
  status: TaskStatus;
  priority: TaskPriority;
  order: number;
  version: number;
  lockedBy?: string;
  lockedAt?: string;
  contracts?: TaskContract[];
  dependencies?: TaskDependency[];
  createdAt: string;
  updatedAt: string;
}

// 任务合约
export interface TaskContract {
  id: string;
  taskId: string;
  type: 'input' | 'output';
  name: string;
  description?: string;
  schema?: Record<string, unknown>;
  example?: string;
  createdAt: string;
}

// 任务依赖
export interface TaskDependency {
  id: string;
  taskId: string;
  dependsOnTaskId: string;
  type: 'finish_to_start' | 'start_to_start' | 'finish_to_finish' | 'start_to_finish';
  status: 'pending' | 'satisfied' | 'blocked';
  createdAt: string;
}

// 依赖详情（包含被依赖任务的信息）
export interface DependencyWithTask extends TaskDependency {
  dependsOnTask?: Task;
}

// 通知类型
export type NotificationType = 'info' | 'warning' | 'error' | 'success';

// 通知
export interface Notification {
  id: string;
  userId?: string;
  type: NotificationType;
  title: string;
  message: string;
  read: boolean;
  entityType?: string;
  entityId?: string;
  createdAt: string;
}

// 变更历史
export interface ChangeHistory {
  id: string;
  entityType: string;
  entityId: string;
  action: 'create' | 'update' | 'delete';
  changes?: Record<string, { old: unknown; new: unknown }>;
  userId?: string;
  createdAt: string;
}

// 同步状态
export type SyncStatus = 'pending' | 'syncing' | 'synced' | 'conflict' | 'error';

// 同步记录
export interface SyncRecord {
  id: string;
  entityType: string;
  entityId: string;
  action: 'create' | 'update' | 'delete';
  status: SyncStatus;
  localData?: Record<string, unknown>;
  remoteData?: Record<string, unknown>;
  conflictResolved?: boolean;
  errorMessage?: string;
  createdAt: string;
  syncedAt?: string;
}

// 配置项
export interface Config {
  id: string;
  key: string;
  value: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
}

// API 响应类型
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

// 分页响应
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

// WebSocket 消息
export interface WebSocketMessage {
  type: string;
  payload: unknown;
  timestamp: string;
}
