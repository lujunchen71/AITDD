import React from 'react';
import { Button, Typography, message } from 'antd';
import { CopyOutlined } from '@ant-design/icons';
import { Task } from '../../../../../types';

interface PromptViewProps {
  taskId: string | null;
  tasks: Task[];
}

const PromptView: React.FC<PromptViewProps> = ({ taskId, tasks }) => {
  const task = tasks.find(t => t.id === taskId);

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      message.success('已复制到剪贴板');
    } catch {
      message.error('复制失败');
    }
  };

  if (!task) {
    return (
      <div style={{ color: '#666', textAlign: 'center', padding: 20 }}>
        {taskId ? '未找到任务信息' : '请选择一个任务'}
      </div>
    );
  }

  const prompt = task.prompt || '';

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
        <Typography.Text style={{ color: '#a0a0a0', fontSize: 12 }}>
          提示词内容
        </Typography.Text>
        {prompt && (
          <Button
            size="small"
            icon={<CopyOutlined />}
            onClick={() => copyToClipboard(prompt)}
            style={{ background: '#1a1a2e', borderColor: '#3d3d5c', color: '#c0c0c0', fontSize: 11 }}
          >
            复制
          </Button>
        )}
      </div>
      <div
        style={{
          background: '#0f0f23',
          border: '1px solid #3d3d5c',
          borderRadius: 4,
          padding: 12,
          maxHeight: 420,
          overflowY: 'auto',
          fontFamily: 'monospace',
          fontSize: 12,
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-word',
          color: '#e0e0e0',
          lineHeight: 1.6,
        }}
      >
        {prompt || <span style={{ color: '#555' }}>（无提示词）</span>}
      </div>
      <Typography.Text style={{ color: '#555', fontSize: 11, marginTop: 4, display: 'block' }}>
        字符数：{prompt.length}
      </Typography.Text>
    </div>
  );
};

export default PromptView;
