import React from 'react';
import { Table, Typography, Empty, List } from 'antd';
import { Task } from '../../../../../types';

interface TestsViewProps {
  taskId: string | null;
  tasks: Task[];
}

interface TestItem {
  target: string;
  api: string;
}

const TestsView: React.FC<TestsViewProps> = ({ taskId, tasks }) => {
  const task = tasks.find(t => t.id === taskId);

  if (!task) {
    return (
      <div style={{ color: '#666', textAlign: 'center', padding: 20 }}>
        {taskId ? '未找到任务信息' : '请选择一个任务'}
      </div>
    );
  }

  // 解析测试用例
  const parseTests = (raw: string | null | undefined): TestItem[] => {
    if (!raw) return [];
    try {
      const parsed = JSON.parse(raw);
      return Array.isArray(parsed) ? parsed : [];
    } catch {
      return [];
    }
  };

  // 解析测试结果
  const parseTestResult = (raw: string | any[] | null | undefined): string[] => {
    if (!raw) return [];
    if (Array.isArray(raw)) return raw.map(String);
    try {
      const parsed = JSON.parse(raw as string);
      return Array.isArray(parsed) ? parsed.map(String) : [String(raw)];
    } catch {
      return [String(raw)];
    }
  };

  const tests = parseTests(task.tests as string);
  const testResult = parseTestResult(task.testResult as any);

  const columns = [
    {
      title: '测试目标',
      dataIndex: 'target',
      key: 'target',
      render: (text: string) => (
        <Typography.Text style={{ color: '#c0c0c0', fontSize: 12 }}>{text}</Typography.Text>
      ),
    },
    {
      title: '测试 API',
      dataIndex: 'api',
      key: 'api',
      render: (text: string) => (
        <Typography.Text code style={{ fontSize: 11, color: '#00d9ff', wordBreak: 'break-all' }}>
          {text}
        </Typography.Text>
      ),
    },
  ];

  return (
    <div>
      {tests.length === 0 ? (
        <Empty
          description={<span style={{ color: '#555', fontSize: 12 }}>无测试用例</span>}
        />
      ) : (
        <Table
          size="small"
          dataSource={tests.map((t, i) => ({ ...t, key: i }))}
          columns={columns}
          pagination={false}
          style={{ background: 'transparent' }}
        />
      )}

      {/* 测试结果记录 */}
      {testResult.length > 0 && (
        <div style={{ marginTop: 16 }}>
          <Typography.Text style={{ color: '#a0a0a0', fontSize: 12, display: 'block', marginBottom: 8 }}>
            测试结果记录
          </Typography.Text>
          <List
            size="small"
            dataSource={testResult}
            renderItem={(r, i) => (
              <List.Item style={{ borderColor: '#3d3d5c', padding: '4px 0' }}>
                <Typography.Text style={{ color: '#c0c0c0', fontSize: 12 }}>
                  {i + 1}. {r}
                </Typography.Text>
              </List.Item>
            )}
          />
        </div>
      )}
    </div>
  );
};

export default TestsView;
