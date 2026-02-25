import React from 'react';
import { Drawer, Tabs, Descriptions, Tag, Input, Spin, List, Card, Typography, Collapse, Empty } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../../../../services/api';
import { ContractDetail, TestItem } from '../../../../types';

const { Text, Title } = Typography;
const { Panel } = Collapse;

interface TaskDetailPanelProps {
  taskId: string | null;
  onClose: () => void;
}

const statusColors: Record<string, string> = {
  ready: 'blue',
  in_progress: 'orange',
  done: 'green',
  blocked: 'red',
  claimed: 'cyan',
  pending_review: 'purple',
  completed: 'green',
  failed: 'red',
};

const statusLabels: Record<string, string> = {
  ready: '就绪',
  in_progress: '进行中',
  done: '已完成',
  blocked: '阻塞',
  claimed: '已认领',
  pending_review: '待审核',
  completed: '已完成',
  failed: '失败',
};

// 解析契约详情 JSON
const parseContractDetail = (jsonStr: string | undefined): ContractDetail | null => {
  if (!jsonStr) return null;
  try {
    return JSON.parse(jsonStr);
  } catch {
    return null;
  }
};

// 解析测试数组 JSON
const parseTests = (jsonStr: string | undefined): TestItem[] => {
  if (!jsonStr) return [];
  try {
    return JSON.parse(jsonStr);
  } catch {
    return [];
  }
};

// 解析测试结果数组 JSON
const parseTestResult = (jsonStr: string | undefined): string[] => {
  if (!jsonStr) return [];
  try {
    return JSON.parse(jsonStr);
  } catch {
    return [];
  }
};

// 契约详情渲染组件
const ContractDetailCard: React.FC<{ detail: ContractDetail | null; title: string }> = ({ detail, title }) => {
  if (!detail) {
    return <Empty description={`无${title}`} image={Empty.PRESENTED_IMAGE_SIMPLE} />;
  }

  return (
    <Card size="small" title={title} className="mb-4">
      <Descriptions column={1} size="small">
        <Descriptions.Item label="标题">{detail.title}</Descriptions.Item>
      </Descriptions>
      
      {detail.list && detail.list.length > 0 && (
        <div className="mt-4">
          <Title level={5}>接口列表</Title>
          <List
            size="small"
            dataSource={detail.list}
            renderItem={(item, index) => (
              <List.Item>
                <div className="w-full">
                  <div className="flex items-start">
                    <Tag color="blue">{index + 1}</Tag>
                    <div className="flex-1">
                      <div className="text-gray-600 mb-1">{item.label}</div>
                      <Text code className="text-xs">{item.contract_api}</Text>
                      {item.from && (
                        <div className="mt-1">
                          <Tag color="purple" className="text-xs">来自: {item.from}</Tag>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </List.Item>
            )}
          />
        </div>
      )}
    </Card>
  );
};

// 测试列表渲染组件
const TestsList: React.FC<{ tests: TestItem[] }> = ({ tests }) => {
  if (!tests || tests.length === 0) {
    return <Empty description="无测试" image={Empty.PRESENTED_IMAGE_SIMPLE} />;
  }

  return (
    <List
      size="small"
      dataSource={tests}
      renderItem={(item, index) => (
        <List.Item>
          <Card size="small" className="w-full" title={<><Tag color="green">测试 {index + 1}</Tag></>}>
            <Descriptions column={1} size="small">
              <Descriptions.Item label="测试目标">{item.target}</Descriptions.Item>
              <Descriptions.Item label="API">
                <Text code>{item.api}</Text>
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </List.Item>
      )}
    />
  );
};

// 测试结果渲染组件
const TestResultList: React.FC<{ results: string[] }> = ({ results }) => {
  if (!results || results.length === 0) {
    return <Empty description="无测试结果" image={Empty.PRESENTED_IMAGE_SIMPLE} />;
  }

  return (
    <List
      size="small"
      dataSource={results}
      renderItem={(item, index) => (
        <List.Item>
          <div className="w-full bg-gray-50 p-2 rounded">
            <Tag color="purple">证据 {index + 1}</Tag>
            <Text className="ml-2">{item}</Text>
          </div>
        </List.Item>
      )}
    />
  );
};

const TaskDetailPanel: React.FC<TaskDetailPanelProps> = ({ taskId, onClose }) => {
  const { data, isLoading } = useQuery({
    queryKey: ['task', taskId],
    queryFn: async () => {
      if (!taskId) return null;
      const response = await apiClient.get(`/tasks/${taskId}`);
      return response.data;
    },
    enabled: !!taskId,
  });

  const task = data?.task;

  // 解析结构化数据
  const upstreamContract = task ? parseContractDetail(task.upstreamContractDetail) : null;
  const downstreamContract = task ? parseContractDetail(task.downstreamContractDetail) : null;
  const tests = task ? parseTests(task.tests) : [];
  const testResults = task ? parseTestResult(task.testResult) : [];

  return (
    <Drawer
      open={!!taskId}
      title={task?.name || '任务详情'}
      onClose={onClose}
      width={700}
      placement="right"
    >
      {isLoading ? (
        <div className="flex justify-center items-center h-64">
          <Spin />
        </div>
      ) : task ? (
        <Tabs
          items={[
            {
              key: 'basic',
              label: '基本信息',
              children: (
                <Descriptions column={1} bordered size="small">
                  <Descriptions.Item label="状态">
                    <Tag color={statusColors[task.status] || 'default'}>
                      {statusLabels[task.status] || task.status}
                    </Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label="分配给">
                    {task.assignee || <span className="text-gray-400">未分配</span>}
                  </Descriptions.Item>
                  <Descriptions.Item label="版本">v{task.version}</Descriptions.Item>
                  <Descriptions.Item label="描述">
                    {task.description || <span className="text-gray-400">无描述</span>}
                  </Descriptions.Item>
                  <Descriptions.Item label="问题详情">
                    {task.issueDetails ? (
                      <div className="whitespace-pre-wrap text-red-600 bg-red-50 p-2 rounded">
                        {task.issueDetails}
                      </div>
                    ) : (
                      <span className="text-gray-400">无问题记录</span>
                    )}
                  </Descriptions.Item>
                  <Descriptions.Item label="创建时间">
                    {new Date(task.createdAt).toLocaleString()}
                  </Descriptions.Item>
                  <Descriptions.Item label="更新时间">
                    {new Date(task.updatedAt).toLocaleString()}
                  </Descriptions.Item>
                </Descriptions>
              ),
            },
            {
              key: 'contracts',
              label: '契约信息',
              children: (
                <Collapse defaultActiveKey={['upstream', 'downstream']}>
                  <Panel header="上游契约详情" key="upstream">
                    <ContractDetailCard detail={upstreamContract} title="上游契约" />
                  </Panel>
                  <Panel header="下游契约详情" key="downstream">
                    <ContractDetailCard detail={downstreamContract} title="下游契约" />
                  </Panel>
                </Collapse>
              ),
            },
            {
              key: 'prompt',
              label: 'Prompt',
              children: (
                <Input.TextArea
                  readOnly
                  value={task.prompt || ''}
                  rows={10}
                  placeholder="无Prompt"
                />
              ),
            },
            {
              key: 'tests',
              label: `测试 (${tests.length})`,
              children: (
                <div>
                  <Title level={5}>测试用例</Title>
                  <TestsList tests={tests} />
                  
                  <Title level={5} className="mt-4">测试结果</Title>
                  <TestResultList results={testResults} />
                </div>
              ),
            },
            {
              key: 'bugLog',
              label: 'Bug 日志',
              children: (
                <Input.TextArea
                  readOnly
                  value={task.bugLog || ''}
                  rows={10}
                  placeholder="无Bug日志"
                  className="font-mono"
                />
              ),
            },
            {
              key: 'assistance',
              label: '人类协助',
              children: (
                <Input.TextArea
                  readOnly
                  value={task.humanAssistance || ''}
                  rows={10}
                  placeholder="无协助记录"
                />
              ),
            },
          ]}
        />
      ) : (
        <div className="text-gray-400">未找到任务</div>
      )}
    </Drawer>
  );
};

export default TaskDetailPanel;
