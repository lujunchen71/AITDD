import React, { useState, useEffect } from 'react';
import { Typography, Button, Empty, message, Tooltip, Spin } from 'antd';
import { PlusOutlined, DeploymentUnitOutlined, AppstoreOutlined } from '@ant-design/icons';
import ModuleTree from './components/ModuleTree';
import ModuleDetail from './components/ModuleDetail';
import ModuleForm from './components/ModuleForm';
import TaskForm from '../tasks/components/TaskForm';
import ViewSwitcher from './components/ViewSwitcher';
import ModuleGraphView from './components/ModuleGraphView';
import { apiClient } from '../../services/api';
import { useProjectId, useProjectStore } from '../../stores/useProjectStore';
import { Module, ModuleDependency, Task, TaskDependency, ViewMode } from '../../types';

const { Title } = Typography;

const ModulesPage: React.FC = () => {
  const projectId = useProjectId();
  const { isInitialized, isLoading } = useProjectStore();
  
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
    if (projectId && viewMode === 'graph') {
      loadProjectGraph();
    }
  }, [projectId, viewMode]);

  const loadModules = async () => {
    try {
      const response = await apiClient.get(`/modules?projectId=${projectId}`);
      // API 响应格式：{success: true, data: {modules: [], total: 0}}
      setModules(response.data?.data?.modules || []);
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
      const response = await apiClient.get(`/graph/project?projectId=${projectId}`);
      // API 响应格式：{success: true, data: {modules: [], tasks: [], ...}}
      const graphData = response.data?.data;
      if (graphData) {
        setModules(graphData.modules || []);
        setTasks(graphData.tasks || []);
        setTaskDependencies(graphData.taskDependencies || []);
        setModuleDependencies(graphData.moduleDependencies || []);
      }
    } catch (error) {
      // 静默失败，图形视图会显示空状态
      console.log('图形视图数据加载失败，将显示空状态');
    }
  };

  const handleSelectModule = (moduleId: string) => {
    setSelectedModuleId(moduleId);
  };

  const handleAddModule = (parentId?: string) => {
    if (!projectId) {
      message.warning('项目尚未初始化，请稍候再试');
      return;
    }
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

  const handleDeleteModule = async (moduleId: string) => {
    try {
      await apiClient.delete(`/modules/${moduleId}`);
      message.success('模块删除成功');
      if (selectedModuleId === moduleId) {
        setSelectedModuleId(null);
      }
      setRefreshKey((prev) => prev + 1);
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
    if (!projectId) {
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
          projectId: projectId,
        });
        message.success('模块创建成功');
      }
      setModuleFormVisible(false);
      setEditingModuleId(undefined);
      setEditingModule(null);
      setParentIdForNew(undefined);
      setRefreshKey((prev) => prev + 1);
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error?.message || '操作失败';
      message.error(errorMsg);
    } finally {
      setFormLoading(false);
    }
  };

  const handleAddTask = (_moduleId: string) => {
    if (!projectId) {
      message.warning('项目尚未初始化，请稍候再试');
      return;
    }
    setEditingTaskId(undefined);
    setEditingTask(null);
    setTaskFormVisible(true);
  };

  const handleEditTask = async (taskId: string, _moduleId: string) => {
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
      // 刷新数据
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
      height: 'calc(100vh - 64px)', 
      display: 'flex', 
      flexDirection: 'column',
      background: '#0f0f23',
    }}>
      {/* 顶部工具栏 */}
      <div style={{ 
        padding: '12px 24px', 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center',
        background: 'linear-gradient(135deg, #1a1a2e 0%, #16213e 100%)',
        borderBottom: '1px solid #2d2d44',
        boxShadow: '0 2px 8px rgba(0, 0, 0, 0.3)',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <DeploymentUnitOutlined style={{ fontSize: 20, color: '#e94560' }} />
          <Title level={4} style={{ margin: 0, color: '#ffffff' }}>
            模块与任务管理
          </Title>
        </div>
        <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
          <ViewSwitcher currentMode={viewMode} onModeChange={handleViewModeChange} />
          <Tooltip title={!projectId ? '项目初始化中...' : ''}>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => handleAddModule()}
              disabled={!projectId}
              style={{ 
                background: 'linear-gradient(135deg, #e94560 0%, #ff6b6b 100%)', 
                border: 'none',
                boxShadow: '0 2px 8px rgba(233, 69, 96, 0.4)',
              }}
            >
              新建模块
            </Button>
          </Tooltip>
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
              <div style={{ flex: 1, overflow: 'auto', padding: '8px' }}>
                <ModuleTree
                  key={refreshKey}
                  projectId={projectId ?? ''}
                  onSelectModule={handleSelectModule}
                  selectedModuleId={selectedModuleId}
                  onAddModule={handleAddModule}
                />
              </div>
            </div>

            {/* 中栏：详情面板 */}
            <div style={{ 
              width: '450px', 
              background: '#16213e',
              borderRight: '1px solid #2d2d44',
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

            {/* 右栏：占位 */}
            <div style={{ 
              flex: 1, 
              background: '#0f0f23',
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
            }}>
              <Empty
                description="切换到节点图表视图查看完整的依赖关系图"
                style={{ padding: '40px 0' }}
                imageStyle={{ filter: 'hue-rotate(200deg)' }}
              />
            </div>
          </>
        ) : (
          // 图形视图模式
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
