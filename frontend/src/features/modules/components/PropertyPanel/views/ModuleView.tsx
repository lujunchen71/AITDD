import React from 'react';
import { Descriptions, Divider, Tag, List, Card, Typography, Button } from 'antd';
import { BugOutlined } from '@ant-design/icons';
import { Module, Task } from '../../../../../types';
import { usePropertyPanelStore } from '../../../../../stores/usePropertyPanelStore';
import { parseBugLog } from '../../BugLogPanel';

interface ModuleViewProps {
  moduleId: string | null;
  modules: Module[];
  tasks: Task[];
}

const moduleStatusColors: Record<string, string> = {
  designing: 'blue',
  developing: 'orange',
  completed: 'success',
  deprecated: 'default',
};

const moduleStatusLabels: Record<string, string> = {
  designing: '设计中',
  developing: '开发中',
  completed: '已完成',
  deprecated: '已废弃',
};

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

const ModuleView: React.FC<ModuleViewProps> = ({ moduleId, modules, tasks }) => {
  const { showTaskInfo, showContent } = usePropertyPanelStore();

  const module = modules.find(m => m.id === moduleId);

  if (!module) {
    return (
      <div style={{ color: '#666', textAlign: 'center', padding: 20 }}>
        {moduleId ? '未找到模块信息' : '请选择一个模块'}
      </div>
    );
  }

  const moduleTasks = tasks.filter(t => t.moduleId === module.id);

  // 计算模块级错误统计
  const moduleErrorStats = moduleTasks.reduce((acc, task) => {
    const entries = parseBugLog(task.bugLog as string);
    acc.errors += entries.filter(e => e.level === 'error').length;
    acc.warns += entries.filter(e => e.level === 'warn').length;
    if (task.issueDetails) acc.issues += 1;
    return acc;
  }, { errors: 0, warns: 0, issues: 0 });
  const hasModuleErrors = moduleErrorStats.errors + moduleErrorStats.warns + moduleErrorStats.issues > 0;

  return (
    <div>
      <Descriptions
        column={1}
        size="small"
        labelStyle={{ color: '#888', fontSize: 12, width: 80 }}
        contentStyle={{ color: '#e0e0e0', fontSize: 12 }}
      >
        <Descriptions.Item label="模块名称">{module.name}</Descriptions.Item>
        <Descriptions.Item label="状态">
          <Tag color={moduleStatusColors[module.status] || 'default'}>
            {moduleStatusLabels[module.status] || module.status}
          </Tag>
        </Descriptions.Item>
        <Descriptions.Item label="描述">
          {module.description || <span style={{ color: '#555' }}>—</span>}
        </Descriptions.Item>
        <Descriptions.Item label="测试覆盖率">{module.testCoverage}%</Descriptions.Item>
        <Descriptions.Item label="锁定状态">
          {module.locked
            ? <Tag color="red">已锁定 {module.lockedBy ? `by ${module.lockedBy}` : ''}</Tag>
            : <Tag color="green">未锁定</Tag>
          }
        </Descriptions.Item>
      </Descriptions>

      {/* 模块错误统计 */}
      {hasModuleErrors && (
        <>
          <Divider style={{ margin: '8px 0' }} />
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 8 }}>
            <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
              {moduleErrorStats.errors > 0 && <Tag color="error">{moduleErrorStats.errors} 个错误</Tag>}
              {moduleErrorStats.warns > 0 && <Tag color="warning">{moduleErrorStats.warns} 个警告</Tag>}
              {moduleErrorStats.issues > 0 && <Tag color="orange">{moduleErrorStats.issues} 个问题</Tag>}
            </div>
            <Button
              type="link"
              size="small"
              icon={<BugOutlined />}
              onClick={() => showContent('error', undefined, module.id)}
              style={{ color: '#ff4d4f', padding: '0 4px', fontSize: 12 }}
            >
              查看错误
            </Button>
          </div>
        </>
      )}

      {/* 上游契约摘要 */}
      {module.upstreamContractSummary && (
        <>
          <Divider />
          <Card
            size="small"
            title={<span style={{ color: '#a0a0a0', fontSize: 12 }}>上游契约摘要</span>}
            style={{ background: '#12122a', borderColor: '#3d3d5c' }}
            bodyStyle={{ padding: 10 }}
          >
            <Typography.Text style={{ color: '#c0c0c0', fontSize: 12, whiteSpace: 'pre-wrap' }}>
              {module.upstreamContractSummary}
            </Typography.Text>
          </Card>
        </>
      )}

      {/* 下游契约摘要 */}
      {module.downstreamContractSummary && (
        <Card
          size="small"
          title={<span style={{ color: '#a0a0a0', fontSize: 12 }}>下游契约摘要</span>}
          style={{ background: '#12122a', borderColor: '#3d3d5c', marginTop: 8 }}
          bodyStyle={{ padding: 10 }}
        >
          <Typography.Text style={{ color: '#c0c0c0', fontSize: 12, whiteSpace: 'pre-wrap' }}>
            {module.downstreamContractSummary}
          </Typography.Text>
        </Card>
      )}

      <Divider />

      {/* 任务列表 */}
      <div>
        <Typography.Text style={{ color: '#a0a0a0', fontSize: 12, marginBottom: 8, display: 'block' }}>
          任务列表 ({moduleTasks.length})
        </Typography.Text>
        {moduleTasks.length === 0 ? (
          <div style={{ color: '#555', fontSize: 12 }}>暂无任务</div>
        ) : (
          <List
            size="small"
            dataSource={moduleTasks}
            renderItem={task => (
              <List.Item
                extra={
                  <Tag color={taskStatusColors[task.status] || 'default'} style={{ fontSize: 11 }}>
                    {taskStatusLabels[task.status] || task.status}
                  </Tag>
                }
                onClick={() => showTaskInfo(task.id)}
                style={{ cursor: 'pointer', borderColor: '#3d3d5c' }}
              >
                <Typography.Text
                  style={{ color: '#c0c0c0', fontSize: 12 }}
                  ellipsis={{ tooltip: task.name }}
                >
                  {task.name}
                </Typography.Text>
              </List.Item>
            )}
          />
        )}
      </div>
    </div>
  );
};

export default ModuleView;
