import React, { useState } from 'react';
import { Row, Col, Card, Typography, Button, Space, Segmented, message, Modal } from 'antd';
import { PlusOutlined, UnorderedListOutlined, ApartmentOutlined, BranchesOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import TaskList from './components/TaskList';
import TaskForm from './components/TaskForm';
import TaskGraph from './components/TaskGraph';
import TaskMermaidView from './components/TaskMermaidView';
import TaskDetailPanel from './components/TaskDetailPanel';
import { apiClient } from '../../services/api';
import { useProjectId, useSelectedProjectIds } from '../../stores/useProjectStore';

const { Title } = Typography;
const { confirm } = Modal;

const TasksPage: React.FC = () => {
  const [viewMode, setViewMode] = useState<'list' | 'graph' | 'mermaid'>('list');
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);
  const [detailVisible, setDetailVisible] = useState(false);
  const [formVisible, setFormVisible] = useState(false);
  const [editingTaskId, setEditingTaskId] = useState<string | undefined>();
  const [refreshKey, setRefreshKey] = useState(0);
  const [selectedModuleId] = useState<string>('');
  const [formLoading, setFormLoading] = useState(false);
  const [editingTask, setEditingTask] = useState<any>(null);
  const projectId = useProjectId();
  const selectedProjectIds = useSelectedProjectIds();

  // 计算用于 Mermaid 视图的项目ID列表：优先使用选中的项目，否则使用当前项目
  const mermaidProjectIds = selectedProjectIds.length > 0
    ? selectedProjectIds
    : projectId
    ? [projectId]
    : [];

  const handleAddTask = () => {
    setEditingTaskId(undefined);
    setEditingTask(null);
    setFormVisible(true);
  };

  const handleEditTask = async (taskId: string) => {
    setFormLoading(true);
    try {
      const response = await apiClient.get(`/tasks/${taskId}`);
      setEditingTask(response.data);
      setEditingTaskId(taskId);
      setFormVisible(true);
    } catch (error) {
      message.error('获取任务详情失败');
    } finally {
      setFormLoading(false);
    }
  };

  const handleFormClose = () => {
    setFormVisible(false);
    setEditingTaskId(undefined);
    setEditingTask(null);
  };

  const handleFormSubmit = async (values: any) => {
    setFormLoading(true);
    try {
      if (editingTaskId) {
        await apiClient.put(`/tasks/${editingTaskId}`, values);
        message.success('任务更新成功');
      } else {
        await apiClient.post('/tasks', values);
        message.success('任务创建成功');
      }
      setFormVisible(false);
      setEditingTaskId(undefined);
      setEditingTask(null);
      setRefreshKey((prev) => prev + 1);
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error?.message || '操作失败';
      message.error(errorMsg);
    } finally {
      setFormLoading(false);
    }
  };

  const handleTaskSelect = (taskId: string) => {
    setSelectedTaskId(taskId);
    setDetailVisible(true);
  };

  const handleDetailClose = () => {
    setDetailVisible(false);
    setSelectedTaskId(null);
  };

  // 删除任务
  const handleDeleteTask = (taskId: string) => {
    confirm({
      title: '确认删除',
      icon: <ExclamationCircleOutlined />,
      content: '确定要删除这个任务吗？此操作不可恢复。',
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await apiClient.delete(`/tasks/${taskId}`);
          message.success('任务删除成功');
          setRefreshKey((prev) => prev + 1);
        } catch (error: any) {
          const errorMsg = error?.response?.data?.error?.message || '删除失败';
          message.error(errorMsg);
        }
      },
    });
  };

  // 检查任务
  const handleCheckTask = async (taskId: string) => {
    try {
      const response = await apiClient.post(`/tasks/${taskId}/check`);
      if (response.data.success) {
        message.success('任务检查通过');
      } else {
        message.warning(response.data.message || '任务检查发现问题');
      }
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error?.message || '检查失败';
      message.error(errorMsg);
    }
  };

  // 重构任务
  const handleRefactorTask = async (taskId: string) => {
    confirm({
      title: '确认重构',
      icon: <ExclamationCircleOutlined />,
      content: '确定要重构这个任务吗？这将重新分析任务的结构和依赖。',
      okText: '重构',
      cancelText: '取消',
      onOk: async () => {
        try {
          await apiClient.post(`/tasks/${taskId}/refactor`);
          message.success('任务重构成功');
          setRefreshKey((prev) => prev + 1);
        } catch (error: any) {
          const errorMsg = error?.response?.data?.error?.message || '重构失败';
          message.error(errorMsg);
        }
      },
    });
  };

  // 复制任务
  const handleDuplicateTask = async (taskId: string) => {
    try {
      await apiClient.post(`/tasks/${taskId}/duplicate`);
      message.success('任务复制成功');
      setRefreshKey((prev) => prev + 1);
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error?.message || '复制失败';
      message.error(errorMsg);
    }
  };

  // 锁定/解锁任务
  const handleLockTask = async (taskId: string) => {
    try {
      const response = await apiClient.post(`/tasks/${taskId}/toggle-lock`);
      message.success(response.data.locked ? '任务已锁定' : '任务已解锁');
      setRefreshKey((prev) => prev + 1);
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error?.message || '操作失败';
      message.error(errorMsg);
    }
  };

  return (
    <div style={{ padding: 24 }}>
      <Row justify="space-between" align="middle" style={{ marginBottom: 16 }}>
        <Col>
          <Title level={4} style={{ margin: 0 }}>
            任务管理
          </Title>
        </Col>
        <Col>
          <Space>
            <Segmented
              value={viewMode}
              onChange={(value) => setViewMode(value as 'list' | 'graph' | 'mermaid')}
              options={[
                {
                  value: 'list',
                  icon: <UnorderedListOutlined />,
                  label: '列表视图',
                },
                {
                  value: 'graph',
                  icon: <ApartmentOutlined />,
                  label: '依赖图',
                },
                {
                  value: 'mermaid',
                  icon: <BranchesOutlined />,
                  label: 'Mermaid 图',
                },
              ]}
            />
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={handleAddTask}
            >
              新建任务
            </Button>
          </Space>
        </Col>
      </Row>

      <Card bordered={false}>
        {viewMode === 'list' ? (
          <TaskList
            key={refreshKey}
            moduleId={selectedModuleId || undefined}
            onEdit={handleEditTask}
            onAdd={handleAddTask}
            onDelete={handleDeleteTask}
            onCheck={handleCheckTask}
            onRefactor={handleRefactorTask}
            onDuplicate={handleDuplicateTask}
            onLock={handleLockTask}
          />
        ) : viewMode === 'graph' ? (
          <div style={{ height: 600 }}>
            <TaskGraph
              key={refreshKey}
              moduleId={selectedModuleId}
              onTaskSelect={handleTaskSelect}
            />
          </div>
        ) : (
          <TaskMermaidView
            key={refreshKey}
            moduleId={selectedModuleId || undefined}
            projectIds={mermaidProjectIds}
            onTaskSelect={handleTaskSelect}
            refreshKey={refreshKey}
          />
        )}
      </Card>

      <TaskForm
        open={formVisible}
        moduleId={selectedModuleId || projectId || ''}
        initialValues={editingTask}
        onSubmit={handleFormSubmit}
        onCancel={handleFormClose}
        loading={formLoading}
      />

      <TaskDetailPanel
        taskId={detailVisible ? selectedTaskId : null}
        onClose={handleDetailClose}
      />
    </div>
  );
};

export default TasksPage;
