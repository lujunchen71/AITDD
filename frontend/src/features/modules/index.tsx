import React, { useState, useEffect } from 'react';
import { Typography, Button, Empty, message, Spin } from 'antd';
import { DeploymentUnitOutlined, AppstoreOutlined } from '@ant-design/icons';
import { ReactFlowProvider } from 'reactflow';
import ModuleTree from './components/ModuleTree';
import ProjectGroupedTree from './components/ProjectGroupedTree';
import ModuleDetail from './components/ModuleDetail';
import ModuleForm from './components/ModuleForm';
import TaskForm from '../tasks/components/TaskForm';
import ViewSwitcher from './components/ViewSwitcher';
import ModuleGraphView from './components/ModuleGraphView';
import { apiClient } from '../../services/api';
import { useProjectId, useProjectStore, useProjects, useSelectedProjectIds } from '../../stores/useProjectStore';
import { Module, ModuleDependency, Task, TaskDependency, ViewMode } from '../../types';
import { useQueryClient } from '@tanstack/react-query';

const { Title } = Typography;

const ModulesPage: React.FC = () => {
  const queryClient = useQueryClient();
  const projectId = useProjectId();
  const { isInitialized, isLoading } = useProjectStore();
  const projects = useProjects();
  const selectedProjectIds = useSelectedProjectIds();
  
  // 当前选中的项目ID（用于模块详情等单项目操作）
  const [currentProjectId, setCurrentProjectId] = useState<string | null>(null);
  
  const [selectedModuleId, setSelectedModuleId] = useState<string | null>(null);
  const [moduleFormVisible, setModuleFormVisible] = useState(false);
  const [editingModuleId, setEditingModuleId] = useState<string | undefined>();
  const [refreshKey, setRefreshKey] = useState(0);
  const [parentIdForNew, setParentIdForNew] = useState<string | undefined>();
  const [formLoading, setFormLoading] = useState(false);
  const [editingModule, setEditingModule] = useState<any>(null);
  
  // 数据状态
  const [modules, setModules] = useState<Module[]>([]);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [taskDependencies, setTaskDependencies] = useState<TaskDependency[]>([]);
  const [moduleDependencies, setModuleDependencies] = useState<ModuleDependency[]>([]);

  const [taskFormVisible, setTaskFormVisible] = useState(false);
  const [editingTaskId, setEditingTaskId] = useState<string | undefined>();
  const [editingTask, setEditingTask] = useState<any>(null);
  const [taskFormLoading, setTaskFormLoading] = useState(false);

  // 视图模式
  const [viewMode, setViewMode] = useState<ViewMode>('list');

  const isProjectReady = isInitialized && projectId;

  // 加载模块
  useEffect(() => {
    if (projectId) {
      loadModules();
    }
  }, [projectId, refreshKey]);

  // 加载选中模块的依赖
  useEffect(() => {
    if (selectedModuleId && viewMode === 'list') {
      loadModuleDependencies(selectedModuleId);
    }
  }, [selectedModuleId, viewMode]);

  // 加载项目图表数据（图形视图模式）
  useEffect(() => {
    if (viewMode === 'graph') {
      loadProjectGraph();
    }
  }, [selectedProjectIds, viewMode]);

  const loadModules = async () => {
    try {
      const response = await apiClient.get(`/modules?projectId=${projectId}`);
      console.log('loadModules response:', response);
      // API 响应格式：{success: true, data: {modules: [], total: 0}} 或 {success: true, data: {data: {modules: []}}}
      const modules = response.data?.data?.modules || response.data?.modules || [];
      
      // 为每个模块加载任务
      const modulesWithTasks = await Promise.all(
        modules.map(async (module: any) => {
          try {
            const taskResponse = await apiClient.get(`/modules/${module.id}/tasks`);
            const tasks = taskResponse.data?.data?.tasks || taskResponse.data?.data || [];
            return { ...module, tasks };
          } catch (error) {
            console.error(`加载模块 ${module.id} 的任务失败:`, error);
            return { ...module, tasks: [] };
          }
        })
      );
      
      setModules(modulesWithTasks);
    } catch (error) {
      console.error('加载模块失败:', error);
    }
  };

  const loadModuleDependencies = async (moduleId: string) => {
    try {
      const response = await apiClient.get(`/modules/${moduleId}/dependencies`);
      setModuleDependencies(response.data?.dependencies || []);
    } catch (error) {
      console.error('加载模块依赖失败:', error);
    }
  };

  const loadProjectGraph = async () => {
    try {
      // 如果没有选中的项目，清空所有图形数据
      // 注意：不再回退使用 projectId，保持与列表模式一致的行为
      if (selectedProjectIds.length === 0) {
        setModules([]);
        setTasks([]);
        setTaskDependencies([]);
        setModuleDependencies([]);
        return;
      }
      
      // 使用 selectedProjectIds 加载数据
      const projectIds = selectedProjectIds;
      
      // 修改API调用，支持多个项目ID
      const response = await apiClient.get(`/graph/project?projectIds=${projectIds.join(',')}`);
      // apiClient 响应拦截器已经解包了 axios 响应，所以 response = {success: true, data: {...}}
      // response.data 就是 {modules: [], tasks: [], ...}
      const graphData = response.data;
      if (graphData) {
        setModules(graphData.modules || []);
        setTasks(graphData.tasks || []);
        setTaskDependencies(graphData.taskDependencies || []);
        setModuleDependencies(graphData.moduleDependencies || []);
      }
    } catch (error) {
      // 静默失败，图形视图会显示空状态
      console.log('图形视图数据加载失败，将显示空状态', error);
    }
  };

  const handleSelectModule = (moduleId: string, projectId?: string) => {
    setSelectedModuleId(moduleId);
    if (projectId) {
      setCurrentProjectId(projectId);
    }
  };

  const handleAddModule = (projectId: string, parentId?: string) => {
    if (!projectId) {
      message.warning('请先选择一个项目');
      return;
    }
    setCurrentProjectId(projectId);
    setEditingModuleId(undefined);
    setEditingModule(null);
    setParentIdForNew(parentId);
    setModuleFormVisible(true);
  };

  const handleEditModule = async (moduleId: string) => {
    setFormLoading(true);
    try {
      const response = await apiClient.get(`/modules/${moduleId}`);
      setEditingModule(response.data);
      setEditingModuleId(moduleId);
      setParentIdForNew(undefined);
      setModuleFormVisible(true);
    } catch (error) {
      message.error('获取模块详情失败');
    } finally {
      setFormLoading(false);
    }
  };

  const handleDeleteModule = async (moduleId: string, projectId?: string) => {
    try {
      await apiClient.delete(`/modules/${moduleId}`);
      message.success('模块删除成功');
      if (selectedModuleId === moduleId) {
        setSelectedModuleId(null);
      }
      setRefreshKey((prev) => prev + 1);
      // 刷新对应项目的模块缓存
      if (projectId) {
        queryClient.invalidateQueries({ queryKey: ['modules', projectId] });
      }
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error?.message || '删除模块失败';
      message.error(errorMsg);
    }
  };

  const handleModuleFormClose = () => {
    setModuleFormVisible(false);
    setEditingModuleId(undefined);
    setEditingModule(null);
    setParentIdForNew(undefined);
  };

  const handleModuleFormSubmit = async (values: any) => {
    // 使用当前选中的项目ID或currentProjectId
    const targetProjectId = currentProjectId || projectId;
    if (!targetProjectId) {
      message.error('项目 ID 不存在，无法创建模块');
      return;
    }
    
    setFormLoading(true);
    try {
      if (editingModuleId) {
        await apiClient.put(`/modules/${editingModuleId}`, values);
        message.success('模块更新成功');
      } else {
        await apiClient.post('/modules', {
          ...values,
          projectId: targetProjectId,
        });
        message.success('模块创建成功');
      }
      setModuleFormVisible(false);
      setEditingModuleId(undefined);
      setEditingModule(null);
      setParentIdForNew(undefined);
      setRefreshKey((prev) => prev + 1);
      // 刷新对应项目的模块缓存
      queryClient.invalidateQueries({ queryKey: ['modules', targetProjectId] });
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error?.message || '操作失败';
      message.error(errorMsg);
    } finally {
      setFormLoading(false);
    }
  };

  const handleAddTask = (moduleId: string, projectId?: string) => {
    if (!projectId && !currentProjectId) {
      message.warning('请先选择一个项目');
      return;
    }
    setEditingTaskId(undefined);
    setEditingTask(null);
    setSelectedModuleId(moduleId);
    if (projectId) {
      setCurrentProjectId(projectId);
    }
    setTaskFormVisible(true);
  };

  const handleEditTask = async (taskId: string, projectId?: string) => {
    setTaskFormLoading(true);
    try {
      const response = await apiClient.get(`/tasks/${taskId}`);
      setEditingTask(response.data?.task);
      setEditingTaskId(taskId);
      setTaskFormVisible(true);
    } catch (error) {
      message.error('获取任务详情失败');
    } finally {
      setTaskFormLoading(false);
    }
  };

  const handleTaskFormClose = () => {
    setTaskFormVisible(false);
    setEditingTaskId(undefined);
    setEditingTask(null);
  };

  const handleTaskFormSubmit = async (values: any) => {
    if (!selectedModuleId) {
      message.error('请先选择一个模块');
      return;
    }
    
    setTaskFormLoading(true);
    try {
      if (editingTaskId) {
        await apiClient.put(`/tasks/${editingTaskId}`, {
          ...values,
          moduleId: selectedModuleId,
        });
        message.success('任务更新成功');
      } else {
        await apiClient.post('/tasks', {
          ...values,
          moduleId: selectedModuleId,
        });
        message.success('任务创建成功');
      }
      setTaskFormVisible(false);
      setEditingTaskId(undefined);
      setEditingTask(null);
      // 刷新模块数据（包含任务）
      await loadModules();
      // 刷新 ModuleDetail 中的任务查询
      queryClient.invalidateQueries({ queryKey: ['moduleTasks', selectedModuleId] });
      queryClient.invalidateQueries({ queryKey: ['modules', projectId] });
      if (viewMode === 'graph') {
        loadProjectGraph();
      }
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error?.message || '操作失败';
      message.error(errorMsg);
    } finally {
      setTaskFormLoading(false);
    }
  };

  // 删除任务（从树节点调用）
  const handleDeleteTaskFromTree = async (taskId: string, projectId?: string) => {
    try {
      await apiClient.delete(`/tasks/${taskId}`);
      message.success('任务删除成功');
      await loadModules();
      // 刷新对应项目的模块缓存
      if (projectId) {
        queryClient.invalidateQueries({ queryKey: ['modules', projectId] });
      }
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error?.message || '删除失败';
      message.error(errorMsg);
    }
  };

  // 编辑任务（从树节点调用）
  const handleEditTaskFromTree = async (taskId: string, projectId?: string) => {
    setTaskFormLoading(true);
    try {
      const response = await apiClient.get(`/tasks/${taskId}`);
      const task = response.data?.task;
      setEditingTask(task);
      setEditingTaskId(taskId);
      // 设置所属模块ID
      if (task?.moduleId) {
        setSelectedModuleId(task.moduleId);
      }
      // 设置当前项目ID
      if (projectId) {
        setCurrentProjectId(projectId);
      }
      setTaskFormVisible(true);
    } catch (error) {
      message.error('获取任务详情失败');
    } finally {
      setTaskFormLoading(false);
    }
  };

  // 检查任务
  const handleCheckTask = async (taskId: string, projectId?: string) => {
    try {
      await apiClient.post(`/tasks/${taskId}/check`);
      message.success('任务检查完成');
      await loadModules();
      // 刷新对应项目的模块缓存
      if (projectId) {
        queryClient.invalidateQueries({ queryKey: ['modules', projectId] });
      }
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error?.message || '检查失败';
      message.error(errorMsg);
    }
  };

  // 重构任务
  const handleRefactorTask = async (taskId: string, projectId?: string) => {
    try {
      await apiClient.post(`/tasks/${taskId}/refactor`);
      message.success('任务重构完成');
      await loadModules();
      // 刷新对应项目的模块缓存
      if (projectId) {
        queryClient.invalidateQueries({ queryKey: ['modules', projectId] });
      }
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error?.message || '重构失败';
      message.error(errorMsg);
    }
  };

  const handleViewModeChange = (mode: ViewMode) => {
    setViewMode(mode);
  };

  if (isLoading && !isInitialized) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%', background: '#0f0f23' }}>
        <Spin size="large" tip="正在加载项目..." />
      </div>
    );
  }

  if (!isProjectReady) {
    return (
      <div style={{ padding: 24, background: '#0f0f23', minHeight: '100vh' }}>
        <Empty
          description="项目尚未初始化，请刷新页面重试"
          style={{ padding: '40px 0' }}
        >
          <Button type="primary" onClick={() => window.location.reload()}>
            刷新页面
          </Button>
        </Empty>
      </div>
    );
  }

  const selectedModule = modules.find(m => m.id === selectedModuleId);

  return (
    <div style={{
      height: '100%',
      display: 'flex',
      flexDirection: 'column',
      background: '#0f0f23',
      overflow: 'hidden',
    }}>
      {/* 顶部工具栏 - 固定在顶部 */}
      <div style={{
        padding: '12px 24px',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        background: 'linear-gradient(135deg, #1a1a2e 0%, #16213e 100%)',
        borderBottom: '1px solid #2d2d44',
        boxShadow: '0 2px 8px rgba(0, 0, 0, 0.3)',
        flexShrink: 0,
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <DeploymentUnitOutlined style={{ fontSize: 20, color: '#e94560' }} />
          <Title level={4} style={{ margin: 0, color: '#ffffff' }}>
            模块与任务管理
          </Title>
        </div>
        <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
          <ViewSwitcher currentMode={viewMode} onModeChange={handleViewModeChange} />
        </div>
      </div>

      {/* 主内容区 */}
      <div style={{ 
        flex: 1, 
        display: 'flex', 
        overflow: 'hidden',
        gap: '1px',
        background: '#1a1a2e',
      }}>
        {viewMode === 'list' ? (
          // 列表视图模式
          <>
            {/* 左栏：模块树 */}
            <div style={{
              width: '280px',
              background: '#16213e',
              borderRight: '1px solid #2d2d44',
              display: 'flex',
              flexDirection: 'column',
            }}>
              <div style={{
                padding: '12px 16px',
                borderBottom: '1px solid #2d2d44',
                background: 'linear-gradient(135deg, #1a1a2e 0%, #16213e 100%)',
                color: '#a0a0a0',
                fontSize: '13px',
                fontWeight: 500,
                display: 'flex',
                alignItems: 'center',
                gap: 8,
              }}>
                <AppstoreOutlined />
                模块结构
              </div>
              <div style={{ flex: 1, overflow: 'auto', padding: '4px' }}>
                <ProjectGroupedTree
                  key={refreshKey}
                  selectedProjectIds={selectedProjectIds}
                  projects={projects}
                  onSelectModule={handleSelectModule}
                  selectedModuleId={selectedModuleId}
                  onAddModule={handleAddModule}
                  onAddTask={handleAddTask}
                  onDeleteModule={handleDeleteModule}
                  onEditTask={handleEditTaskFromTree}
                  onDeleteTask={handleDeleteTaskFromTree}
                  onCheckTask={handleCheckTask}
                  onRefactorTask={handleRefactorTask}
                />
              </div>
            </div>

            {/* 中栏：详情面板 */}
            <div style={{
              flex: 1,
              background: '#16213e',
              display: 'flex',
              flexDirection: 'column',
              overflow: 'hidden',
            }}>
              {selectedModule ? (
                <>
                  <div style={{
                    padding: '12px 16px',
                    borderBottom: '1px solid #2d2d44',
                    background: 'linear-gradient(135deg, #1a1a2e 0%, #16213e 100%)',
                    color: '#a0a0a0',
                    fontSize: '13px',
                    fontWeight: 500,
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                  }}>
                    <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      <AppstoreOutlined />
                      模块详情
                    </span>
                    <span style={{ color: '#e94560', fontSize: '12px', fontWeight: 500 }}>{selectedModule.name}</span>
                  </div>
                  <div style={{ flex: 1, overflow: 'auto' }}>
                    <ModuleDetail
                      moduleId={selectedModuleId!}
                      onEdit={handleEditModule}
                      onDelete={handleDeleteModule}
                      onAddTask={handleAddTask}
                      onEditTask={handleEditTask}
                    />
                  </div>
                </>
              ) : (
                <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
                  <Empty
                    description="请从左侧选择一个模块查看详情和任务"
                    style={{ padding: '40px 0' }}
                    imageStyle={{ filter: 'hue-rotate(200deg)' }}
                  />
                </div>
              )}
            </div>
          </>
        ) : (
          // 图形视图模式
          <ReactFlowProvider>
            <div style={{ flex: 1, overflow: 'hidden' }}>
              <ModuleGraphView
                modules={modules}
                tasks={tasks}
                taskDependencies={taskDependencies}
                moduleDependencies={moduleDependencies}
                onModuleClick={handleSelectModule}
                onTaskClick={(taskId: string) => handleEditTask(taskId, selectedModuleId || '')}
              />
            </div>
          </ReactFlowProvider>
        )}
      </div>

      {/* 模块表单 */}
      <ModuleForm
        open={moduleFormVisible}
        initialValues={editingModule}
        parentId={parentIdForNew}
        onSubmit={handleModuleFormSubmit}
        onCancel={handleModuleFormClose}
        loading={formLoading}
      />

      {/* 任务表单 */}
      {selectedModuleId && (
        <TaskForm
          open={taskFormVisible}
          moduleId={selectedModuleId}
          initialValues={editingTask}
          onSubmit={handleTaskFormSubmit}
          onCancel={handleTaskFormClose}
          loading={taskFormLoading}
        />
      )}
    </div>
  );
};

export default ModulesPage;
