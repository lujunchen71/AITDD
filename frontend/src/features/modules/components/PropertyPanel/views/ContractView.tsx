import React from 'react';
import { Divider, List, Typography, Empty, Tooltip } from 'antd';
import { LinkOutlined } from '@ant-design/icons';
import { Task, ContractDetail, ContractInterfaceItem } from '../../../../../types';
import { usePropertyPanelStore } from '../../../../../stores/usePropertyPanelStore';

interface ContractViewProps {
  taskId: string | null;
  tasks: Task[];
  type: 'upstream' | 'downstream';
}

const ContractView: React.FC<ContractViewProps> = ({ taskId, tasks, type }) => {
  const { showTaskInfo } = usePropertyPanelStore();
  const task = tasks.find(t => t.id === taskId);

  if (!task) {
    return (
      <div style={{ color: '#666', textAlign: 'center', padding: 20 }}>
        {taskId ? '未找到任务信息' : '请选择一个任务'}
      </div>
    );
  }

  // 解析契约详情（JSON 字符串）
  const parseContractDetail = (raw: string | null | undefined): ContractDetail | null => {
    if (!raw) return null;
    try {
      return JSON.parse(raw);
    } catch {
      return null;
    }
  };

  const rawContract = type === 'upstream'
    ? task.upstreamContractDetail
    : task.downstreamContractDetail;

  const contractDetail = parseContractDetail(rawContract as string);

  // 通过 from 路径（格式: projectName/moduleName/taskName）找到对应任务
  const findTaskByFromPath = (fromPath: string): Task | null => {
    if (!fromPath) return null;
    // from 字段格式为 "projectName/moduleName/taskName"，取最后一段作为任务名
    const parts = fromPath.split('/');
    const taskName = parts[parts.length - 1];
    if (!taskName) return null;
    // 先精确匹配名称，后者可能有多个同名任务，优先匹配路径更多段的
    return tasks.find(t => t.name === taskName) || null;
  };

  const handleFromClick = (fromPath: string) => {
    const targetTask = findTaskByFromPath(fromPath);
    if (targetTask) {
      showTaskInfo(targetTask.id);
    }
  };

  if (!contractDetail) {
    return (
      <div>
        <Typography.Text style={{ color: '#888', fontSize: 13 }}>
          {type === 'upstream' ? '上游契约' : '下游契约'}
        </Typography.Text>
        <Empty
          description={<span style={{ color: '#555', fontSize: 13 }}>无契约信息</span>}
          style={{ marginTop: 20 }}
        />
      </div>
    );
  }

  return (
    <div>
      <Typography.Text style={{ color: '#e0e0e0', fontSize: 14, fontWeight: 600 }}>
        {contractDetail.title || (type === 'upstream' ? '上游契约' : '下游契约')}
      </Typography.Text>

      <Divider style={{ borderColor: '#3d3d5c', margin: '8px 0' }} />

      {!contractDetail.list || contractDetail.list.length === 0 ? (
        <Empty
          description={<span style={{ color: '#555', fontSize: 13 }}>契约接口列表为空</span>}
        />
      ) : (
        <List
          size="small"
          dataSource={contractDetail.list}
          renderItem={(item: ContractInterfaceItem, index: number) => {
            const targetTask = item.from ? findTaskByFromPath(item.from) : null;
            return (
              <List.Item
                key={index}
                style={{ borderColor: '#3d3d5c', flexDirection: 'column', alignItems: 'flex-start', padding: '10px 0' }}
              >
                <Typography.Text style={{ color: '#c0c0c0', fontSize: 13, fontWeight: 600, marginBottom: 6 }}>
                  {item.label}
                </Typography.Text>
                <Typography.Text
                  code
                  style={{ fontSize: 12, color: '#00d9ff', wordBreak: 'break-all', display: 'block', marginBottom: 4 }}
                >
                  {item.contract_api}
                </Typography.Text>
                {item.from && (
                  <div style={{ display: 'flex', alignItems: 'center', gap: 4, marginTop: 2 }}>
                    <Typography.Text style={{ color: '#888', fontSize: 12 }}>
                      来自：
                    </Typography.Text>
                    {targetTask ? (
                      <Tooltip title={`跳转到任务：${targetTask.name}`}>
                        <Typography.Text
                          style={{
                            color: '#00d9ff',
                            fontSize: 12,
                            cursor: 'pointer',
                            textDecoration: 'underline',
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: 3,
                          }}
                          onClick={() => handleFromClick(item.from!)}
                        >
                          <LinkOutlined style={{ fontSize: 11 }} />
                          {item.from}
                        </Typography.Text>
                      </Tooltip>
                    ) : (
                      <Typography.Text style={{ color: '#666', fontSize: 12 }}>
                        {item.from}
                      </Typography.Text>
                    )}
                  </div>
                )}
              </List.Item>
            );
          }}
        />
      )}
    </div>
  );
};

export default ContractView;
