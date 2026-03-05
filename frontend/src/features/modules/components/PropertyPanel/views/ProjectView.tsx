import React from 'react';
import { Descriptions, Divider, Row, Col, Statistic, Tag, Collapse, Typography } from 'antd';
import { Project, Module, Task, TaskDependency } from '../../../../../types';

const { Panel } = Collapse;

interface ProjectViewProps {
  project: Project | null;
  modules: Module[];
  tasks: Task[];
  taskDependencies: TaskDependency[];
}

// 任务状态标签颜色
const taskStatusColors: Record<string, string> = {
  ready: 'blue',
  claimed: 'cyan',
  in_progress: 'processing',
  pending_review: 'orange',
  completed: 'success',
  failed: 'error',
  blocked: 'warning',
};

const taskStatusLabels: Record<string, string> = {
  ready: '待领取',
  claimed: '已领取',
  in_progress: '进行中',
  pending_review: '待审核',
  completed: '已完成',
  failed: '失败',
  blocked: '阻塞',
};

const ProjectView: React.FC<ProjectViewProps> = ({ project, modules, tasks, taskDependencies }) => {
  if (!project) {
    return (
      <div style={{ color: '#666', textAlign: 'center', padding: 20 }}>
        暂无项目信息
      </div>
    );
  }

  // 格式化时间
  const formatDate = (timestamp: number) => {
    if (!timestamp) return '—';
    return new Date(timestamp * 1000).toLocaleString('zh-CN');
  };

  // 统计各状态的任务数
  const statusGroups = Object.entries(
    tasks.reduce((acc, t) => {
      acc[t.status] = (acc[t.status] || 0) + 1;
      return acc;
    }, {} as Record<string, number>)
  );

  return (
    <div>
      <Descriptions
        column={1}
        size="small"
        labelStyle={{ color: '#888', fontSize: 12, width: 80 }}
        contentStyle={{ color: '#e0e0e0', fontSize: 12 }}
      >
        <Descriptions.Item label="项目名称">{project.name}</Descriptions.Item>
        <Descriptions.Item label="创建时间">{formatDate(project.createdAt)}</Descriptions.Item>
        <Descriptions.Item label="更新时间">{formatDate(project.updatedAt)}</Descriptions.Item>
        <Descriptions.Item label="版本">{project.version}</Descriptions.Item>
      </Descriptions>

      <Divider />

      {/* 统计信息 */}
      <Row gutter={16} style={{ marginBottom: 8 }}>
        <Col span={8}>
          <Statistic
            title="模块数"
            value={modules.length}
            valueStyle={{ color: '#00d9ff', fontSize: 20 }}
          />
        </Col>
        <Col span={8}>
          <Statistic
            title="任务数"
            value={tasks.length}
            valueStyle={{ color: '#a9d9ff', fontSize: 20 }}
          />
        </Col>
        <Col span={8}>
          <Statistic
            title="依赖数"
            value={taskDependencies.length}
            valueStyle={{ color: '#ff9f43', fontSize: 20 }}
          />
        </Col>
      </Row>

      <Divider />

      {/* 任务状态分布 */}
      {statusGroups.length > 0 && (
        <div style={{ marginBottom: 12 }}>
          <Typography.Text style={{ color: '#888', fontSize: 12, display: 'block', marginBottom: 8 }}>
            任务状态分布
          </Typography.Text>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
            {statusGroups.map(([status, count]) => (
              <Tag key={status} color={taskStatusColors[status] || 'default'}>
                {taskStatusLabels[status] || status}: {count}
              </Tag>
            ))}
          </div>
        </div>
      )}

      {/* 模块列表 */}
      {modules.length > 0 && (
        <>
          <Divider />
          <div>
            <Typography.Text style={{ color: '#888', fontSize: 12, display: 'block', marginBottom: 8 }}>
              模块列表 ({modules.length})
            </Typography.Text>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
              {modules.map(m => (
                <Tag key={m.id} color="blue" style={{ fontSize: 11 }}>
                  {m.name}
                </Tag>
              ))}
            </div>
          </div>
        </>
      )}

      {/* 项目章程（可折叠） */}
      {project.constitution && (
        <>
          <Divider />
          <Collapse
            ghost
            style={{ background: 'transparent' }}
          >
            <Panel
              header={<span style={{ color: '#a0a0a0', fontSize: 12 }}>项目章程</span>}
              key="constitution"
            >
              <pre style={{
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word',
                color: '#c0c0c0',
                fontSize: 12,
                lineHeight: 1.6,
                margin: 0,
                background: '#0f0f23',
                padding: 10,
                borderRadius: 4,
              }}>
                {project.constitution}
              </pre>
            </Panel>
          </Collapse>
        </>
      )}
    </div>
  );
};

export default ProjectView;
