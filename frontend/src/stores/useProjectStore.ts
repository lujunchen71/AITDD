import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { api } from '../services/api';
import { localStorageService } from '../services/localStorageService';

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
  
  // 所有项目列表
  projects: Project[];
  
  // 选中的项目ID列表（用于多选）
  selectedProjectIds: string[];
  
  // 加载状态
  isLoading: boolean;
  error: string | null;
  
  // 初始化状态
  isInitialized: boolean;
  
  // 操作方法
  fetchProject: () => Promise<void>;
  fetchProjects: () => Promise<void>;
  ensureProject: () => Promise<void>;
  setProject: (project: Project | null) => void;
  clearError: () => void;
  reset: () => void;
  
  // 多选相关方法
  toggleProjectSelection: (projectId: string) => void;
  selectAllProjects: () => void;
  deselectAllProjects: () => void;
  setSelectedProjectIds: (ids: string[]) => void;
}

// 从本地存储加载初始选中的项目ID
const loadInitialSelectedProjectIds = (): string[] => {
  return localStorageService.getSelectedProjectIds();
};

// 创建Store - 移除 persist 中间件，每次都从服务器获取项目状态
export const useProjectStore = create<ProjectState>()(
  devtools(
    (set, get) => ({
      // 初始状态
      project: null,
      projectId: null,
      projects: [],
      selectedProjectIds: loadInitialSelectedProjectIds(),
      isLoading: false,
      error: null,
      isInitialized: false,

      // 获取所有项目列表
      fetchProjects: async () => {
        try {
          console.log('[ProjectStore] fetchProjects 开始...');
          const response = await api.get<{ projects: Project[], total: number }>('/projects');
          console.log('[ProjectStore] API 响应:', response);
          const projects = response.data?.projects || [];
          console.log('[ProjectStore] 解析出的 projects:', projects);
          
          // 兼容后端返回的ID字段（可能是大写ID或小写id）
          const normalizedProjects = projects.map((p: any) => ({
            ...p,
            id: p.ID || p.id,
          }));
          
          console.log('[ProjectStore] 标准化后的 projects:', normalizedProjects);
          
          // 获取有效项目ID列表
          const validProjectIds = normalizedProjects.map((p: Project) => p.id);
          
          // 清理无效的项目ID并获取有效的选中项目
          const cleanedSelectedIds = localStorageService.cleanupInvalidProjectIds(validProjectIds);
          
          console.log('[ProjectStore] 当前 selectedProjectIds:', get().selectedProjectIds);
          console.log('[ProjectStore] 清理后的 selectedProjectIds:', cleanedSelectedIds);
          
          set({ projects: normalizedProjects });
          
          // 如果没有选中的项目，默认选中所有项目
          if (cleanedSelectedIds.length === 0 && normalizedProjects.length > 0) {
            const allProjectIds = normalizedProjects.map((p: Project) => p.id);
            console.log('[ProjectStore] 自动选中所有项目:', allProjectIds);
            localStorageService.setSelectedProjectIds(allProjectIds);
            set({ selectedProjectIds: allProjectIds });
          } else {
            // 使用清理后的选中项目ID
            set({ selectedProjectIds: cleanedSelectedIds });
            console.log('[ProjectStore] 使用清理后的选中项目:', cleanedSelectedIds);
          }
        } catch (error: any) {
          console.error('[ProjectStore] fetchProjects 失败:', error);
          set({ projects: [] });
        }
      },

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
          // 先尝试获取所有项目
          const response = await api.get<{ projects: Project[], total: number }>('/projects');
          const projects = response.data?.projects || [];
          
          if (projects.length > 0) {
            // 项目已存在，使用第一个项目作为当前项目
            const project = projects[0];
            const projectId = (project as any).ID || (project as any).id || project.id;
            
            set({
              projectId,
              project: project,
              projects: projects,
              isLoading: false,
              isInitialized: true,
            });
            return;
          }
          
          // 没有项目，提示用户创建
          console.log('没有找到任何项目，请使用数据导入脚本创建项目');
          set({
            projectId: null,
            project: null,
            projects: [],
            isLoading: false,
            isInitialized: true,
            error: '没有找到任何项目，请使用数据导入脚本创建项目',
          });
        } catch (error: any) {
          console.error('获取项目失败:', error);
          set({
            isLoading: false,
            isInitialized: true,
            error: '获取项目失败，请检查后端服务是否正常运行',
          });
        }
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
          projects: [],
          selectedProjectIds: [],
          isLoading: false,
          error: null,
          isInitialized: false,
        });
      },

      // 切换项目选择状态
      toggleProjectSelection: (projectId: string) => {
        const { selectedProjectIds } = get();
        const newSelectedIds = selectedProjectIds.includes(projectId)
          ? selectedProjectIds.filter(id => id !== projectId)
          : [...selectedProjectIds, projectId];
        // 保存到本地存储
        localStorageService.setSelectedProjectIds(newSelectedIds);
        set({ selectedProjectIds: newSelectedIds });
      },

      // 选择所有项目
      selectAllProjects: () => {
        const { projects } = get();
        const allIds = projects.map(p => p.id);
        // 保存到本地存储
        localStorageService.setSelectedProjectIds(allIds);
        set({ selectedProjectIds: allIds });
      },

      // 取消选择所有项目
      deselectAllProjects: () => {
        // 保存到本地存储
        localStorageService.setSelectedProjectIds([]);
        set({ selectedProjectIds: [] });
      },

      // 设置选中的项目ID列表
      setSelectedProjectIds: (ids: string[]) => {
        // 保存到本地存储
        localStorageService.setSelectedProjectIds(ids);
        set({ selectedProjectIds: ids });
      },
    }),
    { name: 'ProjectStore' }
  )
);

// 导出便捷 hook
export const useProjectId = () => useProjectStore((state) => state.projectId);
export const useProject = () => useProjectStore((state) => state.project);
export const useProjects = () => useProjectStore((state) => state.projects);
export const useSelectedProjectIds = () => useProjectStore((state) => state.selectedProjectIds);

export default useProjectStore;
