import { create } from 'zustand';

// 面板内容类型
export type PanelContentType =
  | 'project'             // 未选中任何对象时显示项目信息
  | 'module'              // 选中模块节点
  | 'task'                // 选中任务节点（基本信息）
  | 'prompt'              // 点击提示词按钮
  | 'upstream-contract'   // 点击上游契约按钮
  | 'downstream-contract' // 点击下游契约按钮
  | 'error'               // 点击错误按钮
  | 'tests'               // 点击测试用例按钮
  | 'dependencies';       // 点击任务名称

// 历史记录条目
interface HistoryEntry {
  contentType: PanelContentType;
  selectedModuleId: string | null;
  selectedTaskId: string | null;
}

// Store 状态接口
interface PropertyPanelState {
  // 面板可见性
  visible: boolean;
  toggleVisible: () => void;
  setVisible: (visible: boolean) => void;
  close: () => void;

  // 当前内容类型
  contentType: PanelContentType;
  setContentType: (type: PanelContentType) => void;

  // 选中的对象 ID
  selectedModuleId: string | null;
  selectedTaskId: string | null;
  setSelectedModule: (moduleId: string | null) => void;
  setSelectedTask: (taskId: string | null) => void;

  // 历史记录（用于返回）
  history: HistoryEntry[];
  canGoBack: boolean;
  goBack: () => void;

  // 复合操作：选中模块
  showModuleInfo: (moduleId: string) => void;
  // 复合操作：选中任务
  showTaskInfo: (taskId: string) => void;
  // 复合操作：显示特定内容（会记录历史）
  showContent: (type: PanelContentType, taskId?: string, moduleId?: string) => void;
  // 复合操作：无选中时显示项目信息
  showProjectInfo: () => void;
}

export const usePropertyPanelStore = create<PropertyPanelState>((set, get) => ({
  // 初始状态
  visible: false,
  contentType: 'project',
  selectedModuleId: null,
  selectedTaskId: null,
  history: [],
  canGoBack: false,

  // 切换可见性
  toggleVisible: () => set((s) => ({ visible: !s.visible })),

  // 设置可见性
  setVisible: (visible: boolean) => set({ visible }),

  // 关闭面板
  close: () => set({ visible: false }),

  // 设置内容类型
  setContentType: (type: PanelContentType) => set({ contentType: type }),

  // 设置选中模块
  setSelectedModule: (moduleId: string | null) => set({ selectedModuleId: moduleId }),

  // 设置选中任务
  setSelectedTask: (taskId: string | null) => set({ selectedTaskId: taskId }),

  // 返回上一个历史记录
  goBack: () => {
    const { history } = get();
    if (history.length === 0) return;
    const newHistory = [...history];
    const prev = newHistory.pop()!;
    set({
      contentType: prev.contentType,
      selectedModuleId: prev.selectedModuleId,
      selectedTaskId: prev.selectedTaskId,
      history: newHistory,
      canGoBack: newHistory.length > 0,
    });
  },

  // 显示模块信息（普通单击模块节点）
  showModuleInfo: (moduleId: string) => {
    const { contentType, selectedModuleId, selectedTaskId, history } = get();
    const newHistory = [...history, { contentType, selectedModuleId, selectedTaskId }].slice(-10);
    set({
      visible: true,
      contentType: 'module',
      selectedModuleId: moduleId,
      selectedTaskId: null,
      history: newHistory,
      canGoBack: newHistory.length > 0,
    });
  },

  // 显示任务信息（单击任务节点）
  showTaskInfo: (taskId: string) => {
    const { contentType, selectedModuleId, selectedTaskId, history } = get();
    const newHistory = [...history, { contentType, selectedModuleId, selectedTaskId }].slice(-10);
    set({
      visible: true,
      contentType: 'task',
      selectedTaskId: taskId,
      history: newHistory,
      canGoBack: newHistory.length > 0,
    });
  },

  // 显示特定内容（点击 InfoBadge 等，会记录历史）
  showContent: (type: PanelContentType, taskId?: string, moduleId?: string) => {
    const { contentType, selectedModuleId, selectedTaskId, history } = get();
    const newHistory = [...history, { contentType, selectedModuleId, selectedTaskId }].slice(-10);
    set({
      visible: true,
      contentType: type,
      ...(taskId !== undefined ? { selectedTaskId: taskId } : {}),
      ...(moduleId !== undefined ? { selectedModuleId: moduleId } : {}),
      history: newHistory,
      canGoBack: newHistory.length > 0,
    });
  },

  // 显示项目信息（点击空白处）
  showProjectInfo: () => set({
    visible: true,
    contentType: 'project',
    selectedModuleId: null,
    selectedTaskId: null,
    history: [],
    canGoBack: false,
  }),
}));
