import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';

// UI状态接口
interface UIState {
  // 侧边栏状态
  sidebarCollapsed: boolean;
  toggleSidebar: () => void;
  setSidebarCollapsed: (collapsed: boolean) => void;

  // 主题
  theme: 'light' | 'dark';
  setTheme: (theme: 'light' | 'dark') => void;

  // 加载状态
  globalLoading: boolean;
  setGlobalLoading: (loading: boolean) => void;

  // 当前选中的模块ID
  selectedModuleId: string | null;
  setSelectedModuleId: (id: string | null) => void;

  // 当前选中的任务ID
  selectedTaskId: string | null;
  setSelectedTaskId: (id: string | null) => void;

  // 任务详情面板状态
  taskDetailPanelVisible: boolean;
  setTaskDetailPanelVisible: (visible: boolean) => void;

  // 任务详情面板当前标签
  taskDetailActiveTab: string;
  setTaskDetailActiveTab: (tab: string) => void;

  // 模态框状态
  modals: {
    createModule: boolean;
    editModule: boolean;
    createTask: boolean;
    editTask: boolean;
    deleteConfirm: boolean;
  };
  openModal: (modal: keyof UIState['modals']) => void;
  closeModal: (modal: keyof UIState['modals']) => void;

  // 通知抽屉
  notificationDrawerVisible: boolean;
  setNotificationDrawerVisible: (visible: boolean) => void;

  // 重置状态
  reset: () => void;
}

// 初始状态
const initialState = {
  sidebarCollapsed: false,
  theme: 'light' as const,
  globalLoading: false,
  selectedModuleId: null,
  selectedTaskId: null,
  taskDetailPanelVisible: false,
  taskDetailActiveTab: 'basic',
  modals: {
    createModule: false,
    editModule: false,
    createTask: false,
    editTask: false,
    deleteConfirm: false,
  },
  notificationDrawerVisible: false,
};

// 创建Store
export const useUIStore = create<UIState>()(
  devtools(
    persist(
      (set) => ({
        ...initialState,

        // 侧边栏操作
        toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
        setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),

        // 主题操作
        setTheme: (theme) => set({ theme }),

        // 加载状态
        setGlobalLoading: (loading) => set({ globalLoading: loading }),

        // 选中状态
        setSelectedModuleId: (id) => set({ selectedModuleId: id }),
        setSelectedTaskId: (id) => set({ selectedTaskId: id }),

        // 任务详情面板
        setTaskDetailPanelVisible: (visible) => set({ taskDetailPanelVisible: visible }),
        setTaskDetailActiveTab: (tab) => set({ taskDetailActiveTab: tab }),

        // 模态框操作
        openModal: (modal) => set((state) => ({ modals: { ...state.modals, [modal]: true } })),
        closeModal: (modal) => set((state) => ({ modals: { ...state.modals, [modal]: false } })),

        // 通知抽屉
        setNotificationDrawerVisible: (visible) => set({ notificationDrawerVisible: visible }),

        // 重置
        reset: () => set(initialState),
      }),
      {
        name: 'aitdd-ui-storage',
        partialize: (state) => ({
          sidebarCollapsed: state.sidebarCollapsed,
          theme: state.theme,
        }),
      }
    ),
    { name: 'UIStore' }
  )
);

export default useUIStore;
