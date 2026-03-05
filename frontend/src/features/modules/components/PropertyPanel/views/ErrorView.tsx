import React from 'react';
import { List, Tag, Typography, Empty, Collapse } from 'antd';
import { Task } from '../../../../../types';
import { parseBugLog } from '../../BugLogDisplay';

interface ErrorViewProps {
  taskId: string | null;
  tasks: Task[];
}

const levelColors: Record<string, string> = {
  error: 'error',
  warn: 'warning',
  info: 'processing',
  debug: 'default',
};

const ErrorView: React.FC<ErrorViewProps> = ({ taskId, tasks }) => {
  const task = tasks.find(t => t.id === taskId);

  if (!task) {
    return (
      <div style={{ color: '#666', textAlign: 'center', padding: 20 }}>
        {taskId ? '未找到任务信息' : '请选择一个任务'}
      </div>
    );
  }

  // 解析 bugLog
  const bugLogRaw = task.bugLog;
  const bugLogParsed = bugLogRaw ? parseBugLog(bugLogRaw as string) : [];
  const bugLog: any[] = Array.isArray(bugLogParsed) ? bugLogParsed as any[] : [];

  const formatTimestamp = (ts: string | number | undefined) => {
    if (!ts) return '';
    const d = new Date(typeof ts === 'number' ? ts * 1000 : ts);
    if (isNaN(d.getTime())) return String(ts);
    return d.toLocaleString('zh-CN');
  };

  return (
    <div>
      {bugLog.length === 0 ? (
        <Empty
          description={<span style={{ color: '#555', fontSize: 12 }}>无错误信息</span>}
        />
      ) : (
        <List
          size="small"
          dataSource={bugLog}
          renderItem={(entry: any, index: number) => (
            <List.Item
              key={index}
              style={{ borderColor: '#3d3d5c', padding: '8px 0', flexDirection: 'column', alignItems: 'flex-start' }}
            >
              <div
                style={{
                  width: '100%',
                  borderLeft: `3px solid ${entry.level === 'error' ? '#ff4d4f' : entry.level === 'warn' ? '#faad14' : '#1890ff'}`,
                  paddingLeft: 8,
                  marginBottom: 4,
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                  <Tag color={levelColors[entry.level] || 'default'} style={{ fontSize: 11 }}>
                    {(entry.level || 'info').toUpperCase()}
                  </Tag>
                  {entry.timestamp && (
                    <Typography.Text style={{ color: '#555', fontSize: 11 }}>
                      {formatTimestamp(entry.timestamp)}
                    </Typography.Text>
                  )}
                </div>
                <Typography.Text style={{ color: '#c0c0c0', fontSize: 12, whiteSpace: 'pre-wrap', wordBreak: 'break-word', display: 'block' }}>
                  {entry.message || String(entry)}
                </Typography.Text>
              </div>
            </List.Item>
          )}
        />
      )}

      {/* issueDetails */}
      {task.issueDetails && (
        <Collapse
          ghost
          style={{ marginTop: 8 }}
        >
          <Collapse.Panel
            header={<span style={{ color: '#a0a0a0', fontSize: 12 }}>AI 发现的问题详情</span>}
            key="issueDetails"
          >
            <pre style={{
              whiteSpace: 'pre-wrap',
              wordBreak: 'break-word',
              color: '#c0c0c0',
              fontSize: 12,
              lineHeight: 1.6,
              margin: 0,
            }}>
              {task.issueDetails}
            </pre>
          </Collapse.Panel>
        </Collapse>
      )}
    </div>
  );
};

export default ErrorView;
