import React, { useState } from 'react';
import { Button, Tooltip, message, Modal, Input, Space } from 'antd';
import { LockOutlined, UnlockOutlined, LoadingOutlined } from '@ant-design/icons';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../../../services/api';

interface LockButtonProps {
  resourceType: 'module' | 'task';
  resourceId: string;
  locked: boolean;
  lockedBy?: string | null;
  currentAgent?: string;
  onLockChange?: (locked: boolean) => void;
  size?: 'small' | 'middle' | 'large';
}

const LockButton: React.FC<LockButtonProps> = ({
  resourceType,
  resourceId,
  locked,
  lockedBy,
  currentAgent = 'user',
  onLockChange,
  size = 'small',
}) => {
  const [showAgentInput, setShowAgentInput] = useState(false);
  const [agentId, setAgentId] = useState(currentAgent);
  const queryClient = useQueryClient();

  const lockMutation = useMutation({
    mutationFn: async (agent: string) => {
      const response = await apiClient.post('/lock', {
        resourceType,
        resourceId,
        lockedBy: agent,
      });
      return response.data;
    },
    onSuccess: () => {
      message.success('资源已锁定');
      queryClient.invalidateQueries({ queryKey: [resourceType + 's'] });
      onLockChange?.(true);
    },
    onError: (error: any) => {
      if (error.response?.status === 423) {
        message.warning('资源已被其他代理锁定');
      } else {
        message.error('锁定失败');
      }
    },
  });

  const unlockMutation = useMutation({
    mutationFn: async (agent: string) => {
      const response = await apiClient.delete('/lock', {
        data: {
          resourceType,
          resourceId,
          lockedBy: agent,
        },
      });
      return response.data;
    },
    onSuccess: () => {
      message.success('资源已解锁');
      queryClient.invalidateQueries({ queryKey: [resourceType + 's'] });
      onLockChange?.(false);
    },
    onError: (error: any) => {
      if (error.response?.status === 404) {
        message.warning('无权解锁此资源');
      } else {
        message.error('解锁失败');
      }
    },
  });

  const handleLock = () => {
    if (!agentId || agentId === 'user') {
      setShowAgentInput(true);
    } else {
      lockMutation.mutate(agentId);
    }
  };

  const handleUnlock = () => {
    unlockMutation.mutate(agentId);
  };

  const handleAgentConfirm = () => {
    if (agentId.trim()) {
      lockMutation.mutate(agentId.trim());
      setShowAgentInput(false);
    } else {
      message.warning('请输入代理标识');
    }
  };

  const isLoading = lockMutation.isPending || unlockMutation.isPending;
  const canUnlock = locked && lockedBy === agentId;

  if (locked) {
    if (canUnlock) {
      return (
        <>
          <Tooltip title="点击解锁">
            <Button
              type="primary"
              danger
              size={size}
              icon={isLoading ? <LoadingOutlined /> : <UnlockOutlined />}
              onClick={handleUnlock}
              loading={isLoading}
            >
              解锁
            </Button>
          </Tooltip>
          <Modal
            title="确认解锁"
            open={showAgentInput}
            onOk={handleAgentConfirm}
            onCancel={() => setShowAgentInput(false)}
          >
            <p>请输入您的代理标识以确认解锁:</p>
            <Input
              value={agentId}
              onChange={(e) => setAgentId(e.target.value)}
              placeholder="代理标识"
            />
          </Modal>
        </>
      );
    }

    return (
      <Tooltip title={`已被 ${lockedBy || '未知'} 锁定`}>
        <Button
          size={size}
          icon={<LockOutlined />}
          disabled
          style={{ opacity: 0.6 }}
        >
          已锁定
        </Button>
      </Tooltip>
    );
  }

  return (
    <>
      <Tooltip title="锁定资源以防止他人修改">
        <Button
          size={size}
          icon={isLoading ? <LoadingOutlined /> : <LockOutlined />}
          onClick={handleLock}
          loading={isLoading}
        >
          锁定
        </Button>
      </Tooltip>
      <Modal
        title="锁定资源"
        open={showAgentInput}
        onOk={handleAgentConfirm}
        onCancel={() => setShowAgentInput(false)}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <p>请输入您的代理标识 (AI Agent ID):</p>
          <Input
            value={agentId}
            onChange={(e) => setAgentId(e.target.value)}
            placeholder="例如: cursor-agent-001"
            onPressEnter={handleAgentConfirm}
          />
          <p style={{ fontSize: 12, color: '#888' }}>
            锁定后，其他代理将无法修改此资源
          </p>
        </Space>
      </Modal>
    </>
  );
};

export default LockButton;
