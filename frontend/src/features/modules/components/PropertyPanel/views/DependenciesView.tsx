import React from 'react';
import { List, Tag, Typography, Divider, Tooltip } from 'antd';
import { ArrowDownOutlined, LinkOutlined } from '@ant-design/icons';
import { Task, TaskDependency, Module } from '../../../../../types';
import { usePropertyPanelStore } from '../../../../../stores/usePropertyPanelStore';

interface DependenciesViewProps {
  taskId: string | null;
  tasks: Task[];
  modules: Module[];
  taskDependencies: TaskDependency[];
}

const DependenciesView: React.FC<DependenciesViewProps> = ({
  taskId,
  tasks,
  modules,
  taskDependencies,
}) => {
  const { showTaskInfo } = usePropertyPanelStore();

  if (!taskId) {
    return (
      <div style={{ color: '#666', textAlign: 'center', padding: 20 }}>
        请选择一个任务
      </div>
    );
  }

  const task = tasks.find(t => t.id === taskId);
  if (!task) {
    return (
      <div style={{ color: '#666', textAlign: 'center', padding: 20 }}>
        未找到任务信息
      </div>
    );
  }

  // 从 taskDependencies 中过滤上游和下游依赖
  const upstreamDeps = taskDependencies.filter(d => d.downstreamTaskId === taskId);
  const downstreamDeps = taskDependencies.filter(d => d.upstreamTaskId === taskId);

  const getTaskName = (tid: string) => {
    const t = tasks.find(t => t.id === tid);
    return t ? t.name : tid;
  };

  const getTaskModule = (tid: string) => {
    const t = tasks.find(t => t.id === tid);
    if (!t) return null;
    return modules.find(m => m.id === t.moduleId);
  };

  const depStatusColors: Record<string, string> = {
    active: 'success',
    deprecated: 'default',
    pending: 'warning',
  };

  const handleNavigateToTask = (tid: string) => {
    const targetTask = tasks.find(t => t.id === tid);
    if (targetTask) {
      showTaskInfo(targetTask.id);
    }
  };

  return (
    <div>
      {/* 上游依赖 */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 8 }}>
        <ArrowDownOutlined style={{ color: '#52c41a', fontSize: 14 }} />
        <Typography.Text style={{ color: '#a0a0a0', fontSize: 13, fontWeight: 600 }}>
          上游依赖（Depends On）（{upstreamDeps.length}）
        </Typography.Text>
      </div>
      {upstreamDeps.length === 0 ? (
        <div style={{ color: '#555', fontSize: 13, marginBottom: 12 }}>无上游依赖</div>
      ) : (
        <List
          size="small"
          dataSource={upstreamDeps}
          renderItem={(dep) => {
            const upTask = tasks.find(t => t.id === dep.upstreamTaskId);
            const upModule = upTask ? getTaskModule(dep.upstreamTaskId) : null;
            return (
              <List.Item
                style={{ borderColor: '#3d3d5c', padding: '8px 0', flexDirection: 'column', alignItems: 'flex-start' }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', width: '100%', alignItems: 'center' }}>
                  <Tooltip title="点击跳转到该任务">
                    <Typography.Text
                      style={{
                        color: '#00d9ff',
                        fontSize: 13,
                        cursor: 'pointer',
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: 4,
                      }}
                      onClick={() => handleNavigateToTask(dep.upstreamTaskId)}
                    >
                      <LinkOutlined style={{ fontSize: 12 }} />
                      {getTaskName(dep.upstreamTaskId)}
                    </Typography.Text>
                  </Tooltip>
                  <Tag
                    color={depStatusColors[dep.status] || 'default'}
                    style={{ fontSize: 12 }}
                  >
                    {dep.status || 'active'}
                  </Tag>
                </div>
                {upModule && (
                  <Typography.Text style={{ color: '#666', fontSize: 12, marginTop: 2 }}>
                    模块：{upModule.name}
                  </Typography.Text>
                )}
                {dep.contractSummary && (
                  <Typography.Text style={{ color: '#555', fontSize: 12, display: 'block', marginTop: 2 }}>
                    契约：{dep.contractSummary}
                  </Typography.Text>
                )}
              </List.Item>
            );
          }}
        />
      )}

      <Divider style={{ borderColor: '#3d3d5c', margin: '10px 0' }} />

      {/* 下游依赖 */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 8 }}>
        <ArrowDownOutlined style={{ color: '#ff7a45', fontSize: 14 }} />
        <Typography.Text style={{ color: '#a0a0a0', fontSize: 13, fontWeight: 600 }}>
          下游依赖（Used By）（{downstreamDeps.length}）
        </Typography.Text>
      </div>
      {downstreamDeps.length === 0 ? (
        <div style={{ color: '#555', fontSize: 13 }}>无下游依赖</div>
      ) : (
        <List
          size="small"
          dataSource={downstreamDeps}
          renderItem={(dep) => {
            const downTask = tasks.find(t => t.id === dep.downstreamTaskId);
            const downModule = downTask ? getTaskModule(dep.downstreamTaskId) : null;
            return (
              <List.Item
                style={{ borderColor: '#3d3d5c', padding: '8px 0', flexDirection: 'column', alignItems: 'flex-start' }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', width: '100%', alignItems: 'center' }}>
                  <Tooltip title="点击跳转到该任务">
                    <Typography.Text
                      style={{
                        color: '#00d9ff',
                        fontSize: 13,
                        cursor: 'pointer',
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: 4,
                      }}
                      onClick={() => handleNavigateToTask(dep.downstreamTaskId)}
                    >
                      <LinkOutlined style={{ fontSize: 12 }} />
                      {getTaskName(dep.downstreamTaskId)}
                    </Typography.Text>
                  </Tooltip>
                  <Tag
                    color={depStatusColors[dep.status] || 'default'}
                    style={{ fontSize: 12 }}
                  >
                    {dep.status || 'active'}
                  </Tag>
                </div>
                {downModule && (
                  <Typography.Text style={{ color: '#666', fontSize: 12, marginTop: 2 }}>
                    模块：{downModule.name}
                  </Typography.Text>
                )}
                {dep.contractSummary && (
                  <Typography.Text style={{ color: '#555', fontSize: 12, display: 'block', marginTop: 2 }}>
                    契约：{dep.contractSummary}
                  </Typography.Text>
                )}
              </List.Item>
            );
          }}
        />
      )}
    </div>
  );
};

export default DependenciesView;
