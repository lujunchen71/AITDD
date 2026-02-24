import React, { useState } from 'react';
import { Row, Col, Card, Typography, Button, Empty, message, Tooltip, Spin } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import ModuleTree from './components/ModuleTree';
import ModuleDetail from './components/ModuleDetail';
import ModuleForm from './components/ModuleForm';
import TaskForm from '../tasks/components/TaskForm';
import TaskList from '../tasks/components/TaskList';
import { apiClient } from '../../services/api';
import { useProjectId, useProjectStore } from '../../stores/useProjectStore';

const { Title } = Typography;

const ModulesPage: React.FC = () => {
  // 获取项目ID和初始化状态
  const projectId = useProjectId();
  const { isInitialized, isLoading } = useProjectStore();
  
  // 模块相关状态
  const [selectedModuleId, setSelectedModuleId] = useState<string | null>(null);
  const [moduleFormVisible, setModuleFormVisible] = useState(false);
  const [editingModuleId, setEditingModuleId] = useState<string | undefined>();
  const [refreshKey, setRefreshKey] = useState(0);
  const [parentIdForNew, setParentIdForNew] = useState<string | undefined>();
  const [formLoading, setFormLoading] = useState(false);
  const [editingModule, setEditingModule] = useState<any>(null);

  // 任务相关状态
  const [taskFormVisible, setTaskFormVisible] = useState(false);
  const [editingTaskId, setEditingTaskId] = useState<string | undefined>();
  const [editingTask, setEditingTask] = useState<any>(null);
  const [taskFormLoading, setTaskFormLoading] = useState(false);
  const [taskRefreshKey, setTaskRefreshKey] = useState(0);

  // 检查项目是否已初始化
  const isProjectReady = isInitialized && projectId;

  // ========== 模块相关处理函数 ==========
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
      message.error('项目ID不存在，无法创建模块');
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

  // ========== 任务相关处理函数 ==========
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
      setTaskRefreshKey((prev) => prev + 1);
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error?.message || '操作失败';
      message.error(errorMsg);
    } finally {
      setTaskFormLoading(false);
    }
  };

  const handleTaskDeleted = () => {
    setTaskRefreshKey((prev) => prev + 1);
  };

  // 显示加载状态
  if (isLoading && !isInitialized) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%', padding: 24 }}>
        <Spin size="large" tip="正在加载项目..." />
      </div>
    );
  }

  // 显示项目未初始化提示
  if (!isProjectReady) {
    return (
      <div style={{ padding: 24 }}>
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

  return (
    <div style={{ padding: 24 }}>
      <Row justify="space-between" align="middle" style={{ marginBottom: 16 }}>
        <Col>
          <Title level={4} style={{ margin: 0 }}>
            模块与任务管理
          </Title>
        </Col>
        <Col>
          <Tooltip title={!projectId ? '项目初始化中...' : ''}>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => handleAddModule()}
              disabled={!projectId}
            >
              新建模块
            </Button>
          </Tooltip>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col xs={24} lg={8}>
          <Card
            title="模块结构"
            bordered={false}
            style={{ height: '100%' }}
          >
            <ModuleTree
              key={refreshKey}
              projectId={projectId ?? ''}
              onSelectModule={handleSelectModule}
              selectedModuleId={selectedModuleId}
              onAddModule={handleAddModule}
            />
          </Card>
        </Col>
        <Col xs={24} lg={16}>
          <Card
            title={selectedModuleId ? '模块详情与任务列表' : '模块详情'}
            bordered={false}
            style={{ height: '100%' }}
          >
            {selectedModuleId ? (
              <>
                <ModuleDetail
                  moduleId={selectedModuleId}
                  onEdit={handleEditModule}
                  onDelete={handleDeleteModule}
                  onAddTask={handleAddTask}
                  onEditTask={handleEditTask}
                />
                <div style={{ marginTop: 24 }}>
                  <TaskList
                    key={taskRefreshKey}
                    moduleId={selectedModuleId}
                    onEdit={(taskId) => handleEditTask(taskId, selectedModuleId)}
                    onDelete={handleTaskDeleted}
                  />
                </div>
              </>
            ) : (
              <Empty
                description="请从左侧选择一个模块查看详情和任务"
                style={{ padding: '40px 0' }}
              />
            )}
          </Card>
        </Col>
      </Row>

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
