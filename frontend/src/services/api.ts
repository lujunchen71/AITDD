import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import {
  ModuleDependencyWithModule,
  ModuleDependentWithModule,
  CreateModuleDependencyRequest,
  UpdateModuleDependencyRequest,
} from '../types';

// API基础配置
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:34567/api/v1';

// 创建axios实例
const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
apiClient.interceptors.request.use(
  (config) => {
    // 可以在这里添加认证token
    const token = localStorage.getItem('auth_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    return response.data;
  },
  (error) => {
    const { response } = error;
    if (response) {
      // 处理特定错误码
      switch (response.status) {
        case 401:
          console.error('未授权访问');
          break;
        case 403:
          console.error('禁止访问');
          break;
        case 404:
          console.error('资源不存在');
          break;
        case 500:
          console.error('服务器错误');
          break;
        default:
          console.error(`请求错误: ${response.status}`);
      }
    }
    return Promise.reject(error);
  }
);

// API响应类型
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

// 分页响应类型
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

// 通用请求方法
export const api = {
  get<T>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return apiClient.get(url, config);
  },

  post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return apiClient.post(url, data, config);
  },

  put<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return apiClient.put(url, data, config);
  },

  delete<T>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return apiClient.delete(url, config);
  },

  patch<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return apiClient.patch(url, data, config);
  },
};

// Export both named and default exports for flexibility
export { apiClient };
export default api;

// 模块依赖API
export const moduleDependencyApi = {
  // 获取模块的依赖列表
  getDependencies(moduleId: string) {
    return api.get<{ dependencies: ModuleDependencyWithModule[]; total: number }>(`/modules/${moduleId}/dependencies`);
  },

  // 获取依赖此模块的模块列表
  getDependents(moduleId: string) {
    return api.get<{ dependents: ModuleDependentWithModule[]; total: number }>(`/modules/${moduleId}/dependents`);
  },

  // 创建模块依赖
  createDependency(moduleId: string, data: CreateModuleDependencyRequest) {
    return api.post<{ dependency: ModuleDependencyWithModule }>(`/modules/${moduleId}/dependencies`, data);
  },

  // 更新模块依赖
  updateDependency(moduleId: string, dependencyId: string, data: UpdateModuleDependencyRequest) {
    return api.put<{ dependency: ModuleDependencyWithModule }>(`/modules/${moduleId}/dependencies/${dependencyId}`, data);
  },

  // 删除模块依赖
  deleteDependency(moduleId: string, dependencyId: string) {
    return api.delete<{ deleted: boolean }>(`/modules/${moduleId}/dependencies/${dependencyId}`);
  },
};
