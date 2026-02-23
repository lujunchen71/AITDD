import React from 'react';
import { Card, Input, Tabs } from 'antd';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../../../services/api';

interface ContractEditorProps {
  taskId: string;
}

const ContractEditor: React.FC<ContractEditorProps> = ({ taskId }) => {
  const queryClient = useQueryClient();

  // 获取任务详情
  const { data } = useQuery({
    queryKey: ['task', taskId],
    queryFn: async () => {
      const response = await apiClient.get(`/api/v1/tasks/${taskId}`);
      return response.data;
    },
  });

  // 更新契约
  const updateMutation = useMutation({
    mutationFn: async (values: { upstream?: string; downstream?: string }) => {
      const response = await apiClient.put(`/api/v1/tasks/${taskId}`, {
        ...values,
        version: data?.task?.version,
      });
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['task', taskId] });
    },
  });

  const task = data?.task;

  return (
    <Card title="契约编辑器" size="small">
      <Tabs
        items={[
          {
            key: 'upstream',
            label: '上游契约',
            children: (
              <Input.TextArea
                rows={10}
                value={task?.upstreamContractDetail || ''}
                onChange={(e) => {
                  // 防抖保存
                  updateMutation.mutate({ upstream: e.target.value });
                }}
                placeholder="定义上游模块/任务需要提供的接口和数据格式"
              />
            ),
          },
          {
            key: 'downstream',
            label: '下游契约',
            children: (
              <Input.TextArea
                rows={10}
                value={task?.downstreamContractDetail || ''}
                onChange={(e) => {
                  updateMutation.mutate({ downstream: e.target.value });
                }}
                placeholder="定义本任务提供给下游的接口和数据格式"
              />
            ),
          },
        ]}
      />
    </Card>
  );
};

export default ContractEditor;
