import React, { useState } from 'react';
import { Row, Col, Card, Typography, Button, Space, Segmented, message } from 'antd';
import { PlusOutlined, UnorderedListOutlined, ApartmentOutlined } from '@ant-design/icons';
import TaskList from './components/TaskList';
import TaskForm from './components/TaskForm';
import TaskGraph from './components/TaskGraph';
import TaskDetailPanel from './components/TaskDetailPanel';
import { apiClient } from '../../services/api';
import { useProjectId } from '../../stores/useProjectStore';

const { Title } = Typography;

const TasksPage: React.FC = () => {
  const [viewMode, setViewMode] = useState<'list' | 'graph'>('list');
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);
  const [detailVisible, setDetailVisible] = useState(false);
  const [formVisible, setFormVisible] = useState(false);
  const [editingTaskId, setEditingTaskId] = useState<string | undefined>();
  const [refreshKey, setRefreshKey] = useState(0);
  const [selectedModuleId] = useState<string>('');
  const [formLoading, setFormLoading] = useState(false);
  const [editingTask, setEditingTask] = useState<any>(null);
  const projectId = useProjectId();

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
              onChange={(value) => setViewMode(value as 'list' | 'graph')}
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
          />
        ) : (
          <div style={{ height: 600 }}>
            <TaskGraph
              key={refreshKey}
              moduleId={selectedModuleId}
              onTaskSelect={handleTaskSelect}
            />
          </div>
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
