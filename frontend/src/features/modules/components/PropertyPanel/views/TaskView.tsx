import React from 'react';
import { Descriptions, Divider, Tag, Button, Space, List, Typography } from 'antd';
import { Module, Task } from '../../../../../types';
import { usePropertyPanelStore, PanelContentType } from '../../../../../stores/usePropertyPanelStore';

interface TaskViewProps {
  taskId: string | null;
  modules: Module[];
  tasks: Task[];
}

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

const TaskView: React.FC<TaskViewProps> = ({ taskId, modules, tasks }) => {
  const { showContent } = usePropertyPanelStore();

  const task = tasks.find(t => t.id === taskId);
  const module = task ? modules.find(m => m.id === task.moduleId) : null;

  if (!task) {
    return (
      <div style={{ color: '#666', textAlign: 'center', padding: 20 }}>
        {taskId ? '未找到任务信息' : '请选择一个任务'}
      </div>
    );
  }

  // 解析代码文件路径
  const parseCodePaths = (codePaths: string | string[] | null | undefined): string[] => {
    if (!codePaths) return [];
    if (Array.isArray(codePaths)) return codePaths;
    try {
      const parsed = JSON.parse(codePaths);
      return Array.isArray(parsed) ? parsed : [];
    } catch {
      return typeof codePaths === 'string' ? [codePaths] : [];
    }
  };

  const codePaths = parseCodePaths(task.codePaths as any);

  const handleShowContent = (type: PanelContentType) => {
    showContent(type, task.id);
  };

  return (
    <div>
      <Descriptions
        column={1}
        size="small"
        labelStyle={{ color: '#888', fontSize: 12, width: 80 }}
        contentStyle={{ color: '#e0e0e0', fontSize: 12 }}
      >
        <Descriptions.Item label="任务名称">{task.name}</Descriptions.Item>
        <Descriptions.Item label="所属模块">
          {module ? module.name : <span style={{ color: '#555' }}>{task.moduleId}</span>}
        </Descriptions.Item>
        <Descriptions.Item label="状态">
          <Tag color={taskStatusColors[task.status] || 'default'}>
            {taskStatusLabels[task.status] || task.status}
          </Tag>
        </Descriptions.Item>
        <Descriptions.Item label="负责人">
          {task.assignee || <span style={{ color: '#555' }}>未分配</span>}
        </Descriptions.Item>
        <Descriptions.Item label="描述">
          {task.description || <span style={{ color: '#555' }}>—</span>}
        </Descriptions.Item>
        <Descriptions.Item label="锁定状态">
          {task.locked
            ? <Tag color="red">已锁定 {task.lockedBy ? `by ${task.lockedBy}` : ''}</Tag>
            : <Tag color="green">未锁定</Tag>
          }
        </Descriptions.Item>
      </Descriptions>

      <Divider />

      {/* 快捷跳转按钮 */}
      <div style={{ marginBottom: 12 }}>
        <Typography.Text style={{ color: '#888', fontSize: 12, display: 'block', marginBottom: 8 }}>
          详情查看
        </Typography.Text>
        <Space wrap size={6}>
          {task.prompt && (
            <Button
              size="small"
              onClick={() => handleShowContent('prompt')}
              style={{ background: '#1a1a2e', borderColor: '#3d3d5c', color: '#c0c0c0', fontSize: 11 }}
            >
              📝 提示词
            </Button>
          )}
          {task.tests && (
            <Button
              size="small"
              onClick={() => handleShowContent('tests')}
              style={{ background: '#1a1a2e', borderColor: '#3d3d5c', color: '#c0c0c0', fontSize: 11 }}
            >
              🧪 测试用例
            </Button>
          )}
          {task.upstreamContractDetail && (
            <Button
              size="small"
              onClick={() => handleShowContent('upstream-contract')}
              style={{ background: '#1a1a2e', borderColor: '#3d3d5c', color: '#c0c0c0', fontSize: 11 }}
            >
              ↑ 上游契约
            </Button>
          )}
          {task.downstreamContractDetail && (
            <Button
              size="small"
              onClick={() => handleShowContent('downstream-contract')}
              style={{ background: '#1a1a2e', borderColor: '#3d3d5c', color: '#c0c0c0', fontSize: 11 }}
            >
              ↓ 下游契约
            </Button>
          )}
          {task.bugLog && (
            <Button
              size="small"
              danger
              onClick={() => handleShowContent('error')}
              style={{ fontSize: 11 }}
            >
              ❌ 错误信息
            </Button>
          )}
          <Button
            size="small"
            onClick={() => handleShowContent('dependencies')}
            style={{ background: '#1a1a2e', borderColor: '#3d3d5c', color: '#c0c0c0', fontSize: 11 }}
          >
            🔗 依赖关系
          </Button>
        </Space>
      </div>

      {/* 代码文件路径 */}
      {codePaths.length > 0 && (
        <>
          <Divider />
          <div>
            <Typography.Text style={{ color: '#888', fontSize: 12, display: 'block', marginBottom: 8 }}>
              代码文件 ({codePaths.length})
            </Typography.Text>
            <List
              size="small"
              dataSource={codePaths}
              renderItem={p => (
                <List.Item style={{ borderColor: '#3d3d5c', padding: '4px 0' }}>
                  <Typography.Text
                    code
                    style={{ fontSize: 11, color: '#00d9ff', wordBreak: 'break-all' }}
                  >
                    {p}
                  </Typography.Text>
                </List.Item>
              )}
            />
          </div>
        </>
      )}

      {/* testResult */}
      {task.testResult && (
        <>
          <Divider />
          <div>
            <Typography.Text style={{ color: '#888', fontSize: 12, display: 'block', marginBottom: 8 }}>
              测试结果摘要
            </Typography.Text>
            <Typography.Text style={{ color: '#c0c0c0', fontSize: 12, whiteSpace: 'pre-wrap' }}>
              {typeof task.testResult === 'string' ? task.testResult : JSON.stringify(task.testResult, null, 2)}
            </Typography.Text>
          </div>
        </>
      )}
    </div>
  );
};

export default TaskView;
