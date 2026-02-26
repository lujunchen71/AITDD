/**
 * 本地存储服务
 * 
 * 提供统一的本地存储封装，支持：
 * - 版本控制，便于未来迁移
 * - 错误处理
 * - 类型安全
 * - 默认值回退
 */

// 存储版本号，用于数据迁移
const STORAGE_VERSION = 1;

// 存储键名前缀
const STORAGE_PREFIX = 'aitdd-';

// 存储键名
const STORAGE_KEYS = {
  VERSION: `${STORAGE_PREFIX}storage-version`,
  DISPLAY_SETTINGS: `${STORAGE_PREFIX}display-settings`,
  SELECTED_PROJECTS: `${STORAGE_PREFIX}selected-projects`,
  GRAPH_VIEWPORT: `${STORAGE_PREFIX}graph-viewport`,
  EXPANDED_KEYS: `${STORAGE_PREFIX}expanded-keys`,
  COLLAPSED_MODULES: `${STORAGE_PREFIX}collapsed-modules`,
  SELECTED_MODULES: `${STORAGE_PREFIX}selected-modules`,
} as const;

// 节点显示设置接口
export interface DisplaySettings {
  showInternalEdges: boolean;
  showCrossModuleEdges: boolean;
  showPrompt: boolean;
  showError: boolean;
  showNotification: boolean;
  showStatus: boolean;
  fontSize: number;
  requireAltForTooltip: boolean;
  tooltipScale: number;
}

// 图形视图相机位置接口
export interface GraphViewport {
  x: number;
  y: number;
  zoom: number;
}

// 默认显示设置
export const defaultDisplaySettings: DisplaySettings = {
  showInternalEdges: true,
  showCrossModuleEdges: true,
  showPrompt: true,
  showError: true,
  showNotification: true,
  showStatus: true,
  fontSize: 12,
  requireAltForTooltip: true,
  tooltipScale: 1,
};

// 默认视口设置
export const defaultGraphViewport: GraphViewport = {
  x: 0,
  y: 0,
  zoom: 1,
};

/**
 * 安全地从 localStorage 读取数据
 */
function safeGetItem<T>(key: string, defaultValue: T): T {
  try {
    const item = localStorage.getItem(key);
    if (item === null) {
      return defaultValue;
    }
    const parsed = JSON.parse(item);
    return parsed as T;
  } catch (error) {
    console.warn(`[LocalStorage] 读取 ${key} 失败:`, error);
    return defaultValue;
  }
}

/**
 * 安全地写入 localStorage
 */
function safeSetItem<T>(key: string, value: T): boolean {
  try {
    localStorage.setItem(key, JSON.stringify(value));
    return true;
  } catch (error) {
    console.error(`[LocalStorage] 写入 ${key} 失败:`, error);
    return false;
  }
}

/**
 * 检查存储版本，必要时进行迁移
 */
function checkAndMigrate(): void {
  const currentVersion = safeGetItem<number>(STORAGE_KEYS.VERSION, 0);
  
  if (currentVersion < STORAGE_VERSION) {
    console.log(`[LocalStorage] 检测到旧版本 ${currentVersion}，迁移到版本 ${STORAGE_VERSION}`);
    
    // 未来版本迁移逻辑可以在这里添加
    // if (currentVersion < 2) {
    //   migrateFromV1ToV2();
    // }
    
    // 更新版本号
    safeSetItem(STORAGE_KEYS.VERSION, STORAGE_VERSION);
  }
}

// 初始化时检查版本
try {
  checkAndMigrate();
} catch (error) {
  console.error('[LocalStorage] 版本检查失败:', error);
}

/**
 * 本地存储服务
 */
export const localStorageService = {
  // ==================== 显示设置 ====================

  /**
   * 获取显示设置
   */
  getDisplaySettings: (): DisplaySettings => {
    const settings = safeGetItem<DisplaySettings>(
      STORAGE_KEYS.DISPLAY_SETTINGS,
      defaultDisplaySettings
    );
    
    // 验证并补全缺失的字段
    return {
      ...defaultDisplaySettings,
      ...settings,
    };
  },

  /**
   * 保存显示设置
   */
  setDisplaySettings: (settings: DisplaySettings): boolean => {
    console.log('[LocalStorage] 保存显示设置:', settings);
    return safeSetItem(STORAGE_KEYS.DISPLAY_SETTINGS, settings);
  },

  // ==================== 选中的项目ID ====================

  /**
   * 获取选中的项目ID列表
   */
  getSelectedProjectIds: (): string[] => {
    const ids = safeGetItem<string[]>(STORAGE_KEYS.SELECTED_PROJECTS, []);
    return Array.isArray(ids) ? ids : [];
  },

  /**
   * 保存选中的项目ID列表
   */
  setSelectedProjectIds: (ids: string[]): boolean => {
    console.log('[LocalStorage] 保存选中的项目ID:', ids);
    return safeSetItem(STORAGE_KEYS.SELECTED_PROJECTS, ids);
  },

  /**
   * 清理无效的项目ID
   * @param validProjectIds 有效项目ID列表
   * @returns 清理后的选中项目ID列表
   */
  cleanupInvalidProjectIds: (validProjectIds: string[]): string[] => {
    const selectedIds = localStorageService.getSelectedProjectIds();
    const validSet = new Set(validProjectIds);
    
    // 过滤出仍然有效的项目ID
    const cleanedIds = selectedIds.filter(id => validSet.has(id));
    
    if (cleanedIds.length !== selectedIds.length) {
      console.log(
        `[LocalStorage] 清理了 ${selectedIds.length - cleanedIds.length} 个无效项目ID:`,
        selectedIds.filter(id => !validSet.has(id))
      );
      localStorageService.setSelectedProjectIds(cleanedIds);
    }
    
    return cleanedIds;
  },

  // ==================== 图形视图相机位置 ====================

  /**
   * 获取图形视图相机位置
   */
  getGraphViewport: (): GraphViewport => {
    const viewport = safeGetItem<GraphViewport>(
      STORAGE_KEYS.GRAPH_VIEWPORT,
      defaultGraphViewport
    );
    
    // 验证并补全缺失的字段
    return {
      ...defaultGraphViewport,
      ...viewport,
    };
  },

  /**
   * 保存图形视图相机位置
   */
  setGraphViewport: (viewport: GraphViewport): boolean => {
    console.log('[LocalStorage] 保存图形视图相机位置:', viewport);
    return safeSetItem(STORAGE_KEYS.GRAPH_VIEWPORT, viewport);
  },

  // ==================== 展开的节点Key列表 ====================

  /**
   * 获取展开的节点Key列表
   */
  getExpandedKeys: (): string[] => {
    const keys = safeGetItem<string[]>(STORAGE_KEYS.EXPANDED_KEYS, []);
    return Array.isArray(keys) ? keys : [];
  },

  /**
   * 保存展开的节点Key列表
   */
  setExpandedKeys: (keys: string[]): boolean => {
    console.log('[LocalStorage] 保存展开的节点Key:', keys);
    return safeSetItem(STORAGE_KEYS.EXPANDED_KEYS, keys);
  },

  /**
   * 清理无效的展开节点Key
   * @param validKeys 有效Key列表
   * @returns 清理后的展开Key列表
   */
  cleanupInvalidExpandedKeys: (validKeys: string[]): string[] => {
    const expandedKeys = localStorageService.getExpandedKeys();
    const validSet = new Set(validKeys);
    
    // 过滤出仍然有效的Key
    const cleanedKeys = expandedKeys.filter(key => validSet.has(key));
    
    if (cleanedKeys.length !== expandedKeys.length) {
      console.log(
        `[LocalStorage] 清理了 ${expandedKeys.length - cleanedKeys.length} 个无效展开Key:`,
        expandedKeys.filter(key => !validSet.has(key))
      );
      localStorageService.setExpandedKeys(cleanedKeys);
    }
    
    return cleanedKeys;
  },

  // ==================== 折叠的模块ID列表 ====================

  /**
   * 获取折叠的模块ID列表
   */
  getCollapsedModules: (): string[] => {
    const ids = safeGetItem<string[]>(STORAGE_KEYS.COLLAPSED_MODULES, []);
    return Array.isArray(ids) ? ids : [];
  },

  /**
   * 保存折叠的模块ID列表
   */
  setCollapsedModules: (ids: string[]): boolean => {
    console.log('[LocalStorage] 保存折叠的模块ID:', ids);
    return safeSetItem(STORAGE_KEYS.COLLAPSED_MODULES, ids);
  },

  /**
   * 清理无效的折叠模块ID
   * @param validModuleIds 有效模块ID列表
   * @returns 清理后的折叠模块ID列表
   */
  cleanupInvalidCollapsedModules: (validModuleIds: string[]): string[] => {
    const collapsedIds = localStorageService.getCollapsedModules();
    const validSet = new Set(validModuleIds);
    
    // 过滤出仍然有效的模块ID
    const cleanedIds = collapsedIds.filter(id => validSet.has(id));
    
    if (cleanedIds.length !== collapsedIds.length) {
      console.log(
        `[LocalStorage] 清理了 ${collapsedIds.length - cleanedIds.length} 个无效折叠模块ID:`,
        collapsedIds.filter(id => !validSet.has(id))
      );
      localStorageService.setCollapsedModules(cleanedIds);
    }
    
    return cleanedIds;
  },

// ==================== 选中的模块ID列表 ====================

/**
 * 获取选中的模块ID列表
 */
getSelectedModules: (): string[] => {
  const ids = safeGetItem<string[]>(STORAGE_KEYS.SELECTED_MODULES, []);
  return Array.isArray(ids) ? ids : [];
},

/**
 * 保存选中的模块ID列表
 */
setSelectedModules: (ids: string[]): boolean => {
  console.log('[LocalStorage] 保存选中的模块ID:', ids);
  return safeSetItem(STORAGE_KEYS.SELECTED_MODULES, ids);
},

/**
 * 清理无效的选中模块ID
 * @param validModuleIds 有效模块ID列表
 * @returns 清理后的选中模块ID列表
 */
cleanupInvalidSelectedModules: (validModuleIds: string[]): string[] => {
  const selectedIds = localStorageService.getSelectedModules();
  const validSet = new Set(validModuleIds);
  
  // 过滤出仍然有效的模块ID
  const cleanedIds = selectedIds.filter(id => validSet.has(id));
  
  if (cleanedIds.length !== selectedIds.length) {
    console.log(
      `[LocalStorage] 清理了 ${selectedIds.length - cleanedIds.length} 个无效选中模块ID:`,
      selectedIds.filter(id => !validSet.has(id))
    );
    localStorageService.setSelectedModules(cleanedIds);
  }
  
  return cleanedIds;
},

// ==================== 工具方法 ====================

  /**
   * 清除所有本地存储数据
   */
  clearAll: (): void => {
    try {
      Object.values(STORAGE_KEYS).forEach(key => {
        localStorage.removeItem(key);
      });
      console.log('[LocalStorage] 已清除所有数据');
    } catch (error) {
      console.error('[LocalStorage] 清除数据失败:', error);
    }
  },

  /**
   * 导出所有存储数据（用于调试）
   */
  exportAll: (): Record<string, unknown> => {
    return {
      version: safeGetItem(STORAGE_KEYS.VERSION, 0),
      displaySettings: localStorageService.getDisplaySettings(),
      selectedProjectIds: localStorageService.getSelectedProjectIds(),
      graphViewport: localStorageService.getGraphViewport(),
      expandedKeys: localStorageService.getExpandedKeys(),
      collapsedModules: localStorageService.getCollapsedModules(),
      selectedModules: localStorageService.getSelectedModules(),
    };
  },
};

export default localStorageService;
