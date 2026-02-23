import React from 'react';
import { Tag, Tooltip, Space } from 'antd';
import { LockOutlined, UnlockOutlined, UserOutlined, ClockCircleOutlined } from '@ant-design/icons';

interface LockStatusProps {
  locked: boolean;
  lockedBy?: string | null;
  lockedAt?: number | null;
  showDetails?: boolean;
}

const LockStatus: React.FC<LockStatusProps> = ({
  locked,
  lockedBy,
  lockedAt,
  showDetails = true,
}) => {
  if (!locked) {
    return (
      <Tag icon={<UnlockOutlined />} color="success">
        未锁定
      </Tag>
    );
  }

  const formatTime = (timestamp: number) => {
    const date = new Date(timestamp);
    return date.toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (!showDetails) {
    return (
      <Tag icon={<LockOutlined />} color="warning">
        已锁定
      </Tag>
    );
  }

  return (
    <Space direction="vertical" size={4}>
      <Tag icon={<LockOutlined />} color="warning">
        已锁定
      </Tag>
      {lockedBy && (
        <div style={{ fontSize: 12, color: '#888' }}>
          <UserOutlined style={{ marginRight: 4 }} />
          {lockedBy}
        </div>
      )}
      {lockedAt && (
        <div style={{ fontSize: 12, color: '#888' }}>
          <ClockCircleOutlined style={{ marginRight: 4 }} />
          {formatTime(lockedAt)}
        </div>
      )}
    </Space>
  );
};

export default LockStatus;
