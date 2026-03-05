import React from 'react';
import { Divider, List, Typography, Empty } from 'antd';
import { Task, ContractDetail, ContractInterfaceItem } from '../../../../../types';

interface ContractViewProps {
  taskId: string | null;
  tasks: Task[];
  type: 'upstream' | 'downstream';
}

const ContractView: React.FC<ContractViewProps> = ({ taskId, tasks, type }) => {
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

  if (!contractDetail) {
    return (
      <div>
        <Typography.Text style={{ color: '#888', fontSize: 12 }}>
          {type === 'upstream' ? '上游契约' : '下游契约'}
        </Typography.Text>
        <Empty
          description={<span style={{ color: '#555', fontSize: 12 }}>无契约信息</span>}
          style={{ marginTop: 20 }}
        />
      </div>
    );
  }

  return (
    <div>
      <Typography.Text style={{ color: '#e0e0e0', fontSize: 13, fontWeight: 600 }}>
        {contractDetail.title || (type === 'upstream' ? '上游契约' : '下游契约')}
      </Typography.Text>

      <Divider style={{ borderColor: '#3d3d5c', margin: '8px 0' }} />

      {!contractDetail.list || contractDetail.list.length === 0 ? (
        <Empty
          description={<span style={{ color: '#555', fontSize: 12 }}>契约接口列表为空</span>}
        />
      ) : (
        <List
          size="small"
          dataSource={contractDetail.list}
          renderItem={(item: ContractInterfaceItem, index: number) => (
            <List.Item
              key={index}
              style={{ borderColor: '#3d3d5c', flexDirection: 'column', alignItems: 'flex-start', padding: '8px 0' }}
            >
              <Typography.Text style={{ color: '#c0c0c0', fontSize: 12, fontWeight: 600, marginBottom: 4 }}>
                {item.label}
              </Typography.Text>
              <Typography.Text
                code
                style={{ fontSize: 11, color: '#00d9ff', wordBreak: 'break-all', display: 'block', marginBottom: 2 }}
              >
                {item.contract_api}
              </Typography.Text>
              {item.from && (
                <Typography.Text style={{ color: '#666', fontSize: 11 }}>
                  来自：{item.from}
                </Typography.Text>
              )}
            </List.Item>
          )}
        />
      )}
    </div>
  );
};

export default ContractView;
