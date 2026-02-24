import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { api } from '../services/api';

// 项目接口
export interface Project {
  id: string;
  name: string;
  constitution: string;
  createdAt: number;
  updatedAt: number;
  version: number;
  syncStatus: string;
}

// 项目状态接口
interface ProjectState {
  // 当前项目
  project: Project | null;
  projectId: string | null;
  
  // 加载状态
  isLoading: boolean;
  error: string | null;
  
  // 初始化状态
  isInitialized: boolean;
  
  // 操作方法
  fetchProject: () => Promise<void>;
  ensureProject: () => Promise<void>;
  setProject: (project: Project | null) => void;
  clearError: () => void;
  reset: () => void;
}

// 创建Store - 移除 persist 中间件，每次都从服务器获取项目状态
export const useProjectStore = create<ProjectState>()(
  devtools(
    (set, get) => ({
      // 初始状态
      project: null,
      projectId: null,
      isLoading: false,
      error: null,
      isInitialized: false,

      // 获取项目信息
      fetchProject: async () => {
        set({ isLoading: true, error: null });
        
        try {
          const response = await api.get<{ project: Project | null }>('/project');
          const project = response.data?.project;
          
          if (project) {
            // 兼容后端返回的ID字段（可能是大写ID或小写id）
            const projectId = (project as any).ID || (project as any).id || null;
            set({
              project,
              projectId,
              isLoading: false,
              isInitialized: true,
            });
          } else {
            set({
              project: null,
              projectId: null,
              isLoading: false,
              isInitialized: true,
            });
          }
        } catch (error: any) {
          const errorMessage = error?.response?.data?.error?.message || '获取项目信息失败';
          set({
            error: errorMessage,
            isLoading: false,
            isInitialized: true,
          });
          console.error('Failed to fetch project:', error);
        }
      },

      // 确保项目存在（如果不存在则创建）
      ensureProject: async () => {
        // 防止重复调用
        if (get().isLoading) {
          return;
        }
        
        set({ isLoading: true, error: null });
        
        try {
          // 先尝试获取项目
          const response = await api.get<{ project: Project | null }>('/project');
          const project = response.data?.project;
          
          // 兼容后端返回的ID字段（可能是大写ID或小写id）
          const projectId = project ? ((project as any).ID || (project as any).id) : null;
          
          if (projectId) {
            // 项目已存在
            set({
              projectId,
              project: project,
              isLoading: false,
              isInitialized: true,
            });
            return;
          }
        } catch (error: any) {
          console.log('获取项目失败，尝试创建默认项目...');
        }
        
        // 项目不存在，尝试创建默认项目
        try {
          const response = await api.put<{ project: Project }>('/project/constitution', {
            constitution: '# AITDD 项目宪法\n\n这是自动创建的默认项目宪法。\n\n## 项目规则\n\n- 遵循TDD开发流程\n- 保持代码质量\n- 及时同步任务状态',
            version: 0,
          });
          
          const project = response.data?.project;
          
          if (project) {
            // 兼容后端返回的ID字段（可能是大写ID或小写id）
            const projectId = (project as any).ID || (project as any).id || null;
            set({
              projectId,
              project: project,
              isLoading: false,
              isInitialized: true,
            });
            return;
          }
        } catch (error: any) {
          // 如果是 409 版本冲突，说明项目可能已被其他请求创建，重新获取
          if (error.response?.status === 409) {
            console.log('版本冲突，项目可能已存在，重新获取...');
            try {
              const response = await api.get<{ project: Project | null }>('/project');
              const project = response.data?.project;
              
              // 兼容后端返回的ID字段（可能是大写ID或小写id）
              const projectId = project ? ((project as any).ID || (project as any).id) : null;
              
              if (projectId) {
                set({
                  projectId,
                  project: project,
                  isLoading: false,
                  isInitialized: true,
                });
                return;
              }
            } catch (fetchError) {
              console.error('重新获取项目失败:', fetchError);
            }
          }
          
          console.error('创建默认项目失败:', error);
        }
        
        // 所有尝试都失败
        set({
          isLoading: false,
          isInitialized: true,
          error: '项目初始化失败，请刷新页面重试',
        });
      },

      // 设置项目
      setProject: (project) => {
        set({
          project,
          projectId: project?.id || null,
        });
      },

      // 清除错误
      clearError: () => {
        set({ error: null });
      },

      // 重置状态
      reset: () => {
        set({
          project: null,
          projectId: null,
          isLoading: false,
          error: null,
          isInitialized: false,
        });
      },
    }),
    { name: 'ProjectStore' }
  )
);

// 导出便捷 hook
export const useProjectId = () => useProjectStore((state) => state.projectId);
export const useProject = () => useProjectStore((state) => state.project);

export default useProjectStore;
