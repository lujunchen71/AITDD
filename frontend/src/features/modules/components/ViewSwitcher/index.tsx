import React from 'react';
import { Button, Tooltip } from 'antd';
import { UnorderedListOutlined, DeploymentUnitOutlined } from '@ant-design/icons';
import type { ViewMode } from '../../../types';

interface ViewSwitcherProps {
  currentMode: ViewMode;
  onModeChange: (mode: ViewMode) => void;
}

const ViewSwitcher: React.FC<ViewSwitcherProps> = ({
  currentMode,
  onModeChange,
}) => {
  return (
    <div style={{ 
      display: 'flex', 
      gap: '8px', 
      background: '#1a1a2e',
      padding: '4px 8px',
      borderRadius: '6px',
      border: '1px solid #2d2d44',
    }}>
      <Tooltip title="文件列表视图">
        <Button
          type={currentMode === 'list' ? 'primary' : 'text'}
          icon={<UnorderedListOutlined />}
          onClick={() => onModeChange('list')}
          size="small"
          style={{
            background: currentMode === 'list' 
              ? 'linear-gradient(135deg, #e94560 0%, #ff6b6b 100%)' 
              : 'transparent',
            border: 'none',
            color: currentMode === 'list' ? '#ffffff' : '#a0a0a0',
          }}
        />
      </Tooltip>
      <Tooltip title="节点图表视图">
        <Button
          type={currentMode === 'graph' ? 'primary' : 'text'}
          icon={<DeploymentUnitOutlined />}
          onClick={() => onModeChange('graph')}
          size="small"
          style={{
            background: currentMode === 'graph' 
              ? 'linear-gradient(135deg, #e94560 0%, #ff6b6b 100%)' 
              : 'transparent',
            border: 'none',
            color: currentMode === 'graph' ? '#ffffff' : '#a0a0a0',
          }}
        />
      </Tooltip>
    </div>
  );
};

export default ViewSwitcher;
