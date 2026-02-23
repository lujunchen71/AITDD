// 通用类型定义

// 基础实体接口
export interface BaseEntity {
  id: string;
  createdAt: number;
  updatedAt?: number;
  version: number;
  syncStatus: SyncStatus;
}

// 同步状态枚举
export type SyncStatus =
  | 'SYNCED'
  | 'PENDING_UPLOAD'
  | 'PENDING_DOWNLOAD'
  | 'CONFLICT';

// 模块状态枚举
export type ModuleStatus =
  | 'designing'
  | 'developing'
  | 'completed'
  | 'deprecated';

// 任务状态枚举
export type TaskStatus =
  | 'ready'
  | 'claimed'
  | 'in_progress'
  | 'pending_review'
  | 'completed'
  | 'failed'
  | 'blocked';

// 通知类型枚举
export type NotificationType =
  | 'task_completed'
  | 'task_failed'
  | 'task_blocked'
  | 'review_required'
  | 'contract_change';

// 锁定信息接口
export interface LockInfo {
  locked: boolean;
  lockedBy?: string;
  lockedAt?: number;
  lockExpiresAt?: number;
}

// 分页参数
export interface PaginationParams {
  page?: number;
  pageSize?: number;
}

// 排序参数
export interface SortParams {
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

// 查询参数
export interface QueryParams extends PaginationParams, SortParams {
  [key: string]: unknown;
}

// 列表响应
export interface ListResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

// 树节点接口
export interface TreeNode<T> {
  id: string;
  name: string;
  children?: TreeNode<T>[];
  data?: T;
}

// 操作结果
export interface OperationResult {
  success: boolean;
  message?: string;
  data?: unknown;
}

// 错误信息
export interface ErrorInfo {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

// API响应
export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: ErrorInfo;
  timestamp: number;
}

// 键值对
export interface KeyValuePair<K = string, V = unknown> {
  key: K;
  value: V;
}

// 时间范围
export interface TimeRange {
  start: number;
  end: number;
}

// 选择项
export interface SelectOption {
  label: string;
  value: string | number;
  disabled?: boolean;
}

// 表单字段错误
export interface FieldError {
  field: string;
  message: string;
}

// 表单验证结果
export interface ValidationResult {
  valid: boolean;
  errors: FieldError[];
}
