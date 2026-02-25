// API类型定义
import { BaseEntity, SyncStatus, ModuleStatus, TaskStatus, NotificationType, LockInfo } from './common';

// ==================== 项目相关 ====================

export interface Project extends BaseEntity {
  name: string;
  constitution: string;
}

export interface GetProjectResponse {
  project: Project | null;
}

export interface GetConstitutionResponse {
  constitution: string;
}

export interface UpdateConstitutionRequest {
  constitution: string;
  version: number;
}

// ==================== 模块相关 ====================

export interface Module extends BaseEntity, LockInfo {
  parentId: string | null;
  projectId: string;
  name: string;
  description: string;
  prompt: string;
  status: ModuleStatus;
  testCoverage: number;
  upstreamContractSummary: string;
  downstreamContractSummary: string;
}

export interface GetModulesParams {
  projectId?: string;
  parentId?: string;
  status?: ModuleStatus;
  flat?: boolean;
  fields?: string;
}

export interface GetModulesResponse {
  modules: Module[];
  total?: number;
}

export interface GetModuleResponse {
  module: Module;
}

export interface CreateModuleRequest {
  projectId: string;
  parentId?: string;
  name: string;
  description?: string;
  prompt?: string;
}

export interface UpdateModuleRequest {
  name?: string;
  description?: string;
  prompt?: string;
  status?: ModuleStatus;
  testCoverage?: number;
  upstreamContractSummary?: string;
  downstreamContractSummary?: string;
  version: number;
}

export interface GetModuleTasksResponse {
  tasks: Task[];
  total: number;
}

// ==================== 任务相关 ====================

export interface Task extends BaseEntity, LockInfo {
  moduleId: string;
  name: string;
  description: string;
  status: TaskStatus;
  assignee: string | null;
  upstreamContractDetail: string;
  downstreamContractDetail: string;
  prompt: string;
  tests: string; // JSON array of TestItem objects
  testResult: string; // JSON array of test evidence strings
  logs: string; // JSON array
  codePaths: string; // JSON array
  humanAssistance: string; // JSON object
}

export interface GetTasksParams {
  moduleId?: string;
  status?: TaskStatus;
  assignee?: string;
  page?: number;
  pageSize?: number;
}

export interface GetTasksResponse {
  tasks: Task[];
  total: number;
  page: number;
  pageSize: number;
}

export interface GetTaskResponse {
  task: Task;
}

export interface CreateTaskRequest {
  moduleId: string;
  name: string;
  description?: string;
  prompt?: string;
  upstreamContractDetail?: string;
  downstreamContractDetail?: string;
}

export interface UpdateTaskRequest {
  name?: string;
  description?: string;
  status?: TaskStatus;
  assignee?: string;
  prompt?: string;
  upstreamContractDetail?: string;
  downstreamContractDetail?: string;
  tests?: string;
  testResult?: string;
  logs?: string;
  codePaths?: string;
  humanAssistance?: string;
  version: number;
}

// ==================== 依赖相关 ====================

export interface Dependency extends BaseEntity {
  upstreamTaskId: string;
  downstreamTaskId: string;
  contractSummary: string;
}

export interface GetDependenciesParams {
  taskId?: string;
  upstream?: boolean;
}

export interface GetDependenciesResponse {
  dependencies: Dependency[];
}

export interface CreateDependencyRequest {
  upstreamTaskId: string;
  downstreamTaskId: string;
  contractSummary?: string;
}

// ==================== 通知相关 ====================

export interface Notification extends Omit<BaseEntity, 'updatedAt' | 'version' | 'syncStatus'> {
  fromTaskId: string;
  toTaskId: string;
  type: NotificationType;
  title: string;
  content: string;
  read: boolean;
  syncStatus: SyncStatus;
}

export interface GetNotificationsParams {
  toTaskId?: string;
  unreadOnly?: boolean;
  page?: number;
  pageSize?: number;
}

export interface GetNotificationsResponse {
  notifications: Notification[];
  total: number;
  unreadCount: number;
}

export interface CreateNotificationRequest {
  fromTaskId: string;
  toTaskId: string;
  type: NotificationType;
  title: string;
  content?: string;
}

// ==================== 锁相关 ====================

export interface LockRequest {
  resourceType: 'module' | 'task';
  resourceId: string;
  holder: string;
  ttl?: number; // 锁定时长（秒）
}

export interface UnlockRequest {
  resourceType: 'module' | 'task';
  resourceId: string;
  holder: string;
}

export interface LockStatusResponse {
  locked: boolean;
  lockedBy?: string;
  lockedAt?: number;
  lockExpiresAt?: number;
}

// ==================== 工具相关 ====================

export interface OpenBrowserRequest {
  url: string;
}

// ==================== WebSocket消息 ====================

export interface WebSocketMessage<T = unknown> {
  type: 'notification' | 'task_update' | 'module_update' | 'lock_change';
  payload: T;
  timestamp: number;
}

export interface NotificationPayload {
  notification: Notification;
}

export interface TaskUpdatePayload {
  task: Task;
  action: 'create' | 'update' | 'delete';
}

export interface ModuleUpdatePayload {
  module: Module;
  action: 'create' | 'update' | 'delete';
}

export interface LockChangePayload {
  resourceType: 'module' | 'task';
  resourceId: string;
  locked: boolean;
  lockedBy?: string;
}
