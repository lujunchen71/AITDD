import React from 'react';
import { Tooltip } from 'antd';

// ==================== BugLog 类型定义 ====================

export interface BugLogItem {
  error: string[];
  warning: string[];
}

export interface BugLog {
  static: BugLogItem;
  dynamic: BugLogItem;
  exe: BugLogItem;
}

// ==================== 解析工具函数 ====================

/**
 * 解析 bugLog JSON 字符串为 BugLog 对象
 */
export const parseBugLog = (bugLogStr: string | undefined | null): BugLog | null => {
  if (!bugLogStr) return null;
  try {
    const parsed = JSON.parse(bugLogStr);
    // 验证基本结构
    if (parsed && typeof parsed === 'object') {
      // 兼容新的结构化格式
      if ('static' in parsed || 'dynamic' in parsed || 'exe' in parsed) {
        const result: BugLog = {
          static: {
            error: Array.isArray(parsed.static?.error) ? parsed.static.error : [],
            warning: Array.isArray(parsed.static?.warning) ? parsed.static.warning : [],
          },
          dynamic: {
            error: Array.isArray(parsed.dynamic?.error) ? parsed.dynamic.error : [],
            warning: Array.isArray(parsed.dynamic?.warning) ? parsed.dynamic.warning : [],
          },
          exe: {
            error: Array.isArray(parsed.exe?.error) ? parsed.exe.error : [],
            warning: Array.isArray(parsed.exe?.warning) ? parsed.exe.warning : [],
          },
        };
        return result;
      }
    }
    return null;
  } catch {
    return null;
  }
};

/**
 * 获取 bugLog 汇总统计
 */
export const getBugLogSummary = (bugLog: BugLog) => {
  const totalErrors =
    bugLog.static.error.length +
    bugLog.dynamic.error.length +
    bugLog.exe.error.length;
  const totalWarnings =
    bugLog.static.warning.length +
    bugLog.dynamic.warning.length +
    bugLog.exe.warning.length;
  return { totalErrors, totalWarnings };
};

// ==================== BugLogDisplay 组件 ====================

interface BugLogDisplayProps {
  /** bugLog JSON 字符串 */
  bugLogStr: string;
  /** 字体大小 */
  fontSize?: number;
  /** 是否需要 Alt 键才显示 Tooltip */
  requireAltForTooltip?: boolean;
  /** Tooltip 缩放比例 */
  tooltipScale?: number;
}

// 每个分类的名称
const categoryLabels: Record<string, string> = {
  static: '静态检查',
  dynamic: '动态分析',
  exe: '执行检查',
};

/**
 * 渲染警告类型的 Tooltip 内容（展开所有来源的 warning）
 */
const renderWarningTooltip = (bugLog: BugLog, scaledFontSize: number, tooltipWidth: number, totalWarnings: number) => {
  const categories: (keyof BugLog)[] = ['exe', 'static', 'dynamic'];

  return (
    <div style={{ width: tooltipWidth }}>
      {/* 标题行 */}
      <div
        style={{
          fontWeight: 600,
          fontSize: scaledFontSize + 2,
          marginBottom: 10,
          color: '#f59e0b',
          borderBottom: '1px solid #f59e0b33',
          paddingBottom: 6,
          display: 'flex',
          alignItems: 'center',
          gap: 8,
        }}
      >
        ⚠ 警告详情
        <span style={{ color: '#f59e0b', fontWeight: 400, fontSize: scaledFontSize }}>
          共 {totalWarnings} 条
        </span>
      </div>

      {/* 各来源警告 */}
      {categories.map((categoryKey) => {
        const item = bugLog[categoryKey];
        if (!item || item.warning.length === 0) return null;

        return (
          <div
            key={categoryKey}
            style={{
              marginBottom: 10,
              borderRadius: 6,
              overflow: 'hidden',
              border: '1px solid rgba(245,158,11,0.2)',
            }}
          >
            {/* 分类标题 */}
            <div
              style={{
                background: 'rgba(245,158,11,0.08)',
                padding: '4px 10px',
                fontSize: scaledFontSize - 1,
                fontWeight: 600,
                color: '#ccc',
                borderBottom: '1px solid rgba(245,158,11,0.15)',
              }}
            >
              📦 {categoryLabels[categoryKey] || categoryKey}
              <span style={{ color: '#f59e0b', marginLeft: 8 }}>
                {item.warning.length} 警告
              </span>
            </div>

            {/* 警告列表 - 全部展开 */}
            <div style={{ padding: '6px 10px' }}>
              {item.warning.map((warn, idx) => (
                <div
                  key={idx}
                  style={{
                    display: 'flex',
                    alignItems: 'flex-start',
                    gap: 6,
                    padding: '3px 0',
                    borderBottom:
                      idx < item.warning.length - 1
                        ? '1px solid rgba(255,255,255,0.05)'
                        : 'none',
                  }}
                >
                  <span style={{ color: '#f59e0b', flexShrink: 0, fontSize: scaledFontSize - 1 }}>⚠</span>
                  <span
                    style={{
                      color: '#fcd34d',
                      fontSize: scaledFontSize - 1,
                      lineHeight: 1.5,
                      wordBreak: 'break-word',
                    }}
                  >
                    {warn}
                  </span>
                </div>
              ))}
            </div>
          </div>
        );
      })}
    </div>
  );
};

/**
 * 渲染错误类型的 Tooltip 内容（展开所有来源的 error）
 */
const renderErrorTooltip = (bugLog: BugLog, scaledFontSize: number, tooltipWidth: number, totalErrors: number) => {
  const categories: (keyof BugLog)[] = ['exe', 'static', 'dynamic'];

  return (
    <div style={{ width: tooltipWidth }}>
      {/* 标题行 */}
      <div
        style={{
          fontWeight: 600,
          fontSize: scaledFontSize + 2,
          marginBottom: 10,
          color: '#ff6b6b',
          borderBottom: '1px solid #ff6b6b33',
          paddingBottom: 6,
          display: 'flex',
          alignItems: 'center',
          gap: 8,
        }}
      >
        ❌ 错误详情
        <span style={{ color: '#ff6b6b', fontWeight: 400, fontSize: scaledFontSize }}>
          共 {totalErrors} 个
        </span>
      </div>

      {/* 各来源错误 */}
      {categories.map((categoryKey) => {
        const item = bugLog[categoryKey];
        if (!item || item.error.length === 0) return null;

        return (
          <div
            key={categoryKey}
            style={{
              marginBottom: 10,
              borderRadius: 6,
              overflow: 'hidden',
              border: '1px solid rgba(255,107,107,0.2)',
            }}
          >
            {/* 分类标题 */}
            <div
              style={{
                background: 'rgba(255,107,107,0.08)',
                padding: '4px 10px',
                fontSize: scaledFontSize - 1,
                fontWeight: 600,
                color: '#ccc',
                borderBottom: '1px solid rgba(255,107,107,0.15)',
              }}
            >
              📦 {categoryLabels[categoryKey] || categoryKey}
              <span style={{ color: '#ff6b6b', marginLeft: 8 }}>
                {item.error.length} 错误
              </span>
            </div>

            {/* 错误列表 - 全部展开 */}
            <div style={{ padding: '6px 10px' }}>
              {item.error.map((err, idx) => (
                <div
                  key={idx}
                  style={{
                    display: 'flex',
                    alignItems: 'flex-start',
                    gap: 6,
                    padding: '3px 0',
                    borderBottom:
                      idx < item.error.length - 1
                        ? '1px solid rgba(255,255,255,0.05)'
                        : 'none',
                  }}
                >
                  <span style={{ color: '#ff6b6b', flexShrink: 0, fontSize: scaledFontSize - 1 }}>✕</span>
                  <span
                    style={{
                      color: '#ffaaaa',
                      fontSize: scaledFontSize - 1,
                      lineHeight: 1.5,
                      wordBreak: 'break-word',
                    }}
                  >
                    {err}
                  </span>
                </div>
              ))}
            </div>
          </div>
        );
      })}
    </div>
  );
};

/**
 * BugLogDisplay 组件
 * 将结构化 bugLog（static/dynamic/exe）以可读的方式显示在节点界面上：
 * - 错误图标（❌）悬停显示所有来源的错误详情（exe/static/dynamic 全展开）
 * - 警告图标（⚠）悬停显示所有来源的警告详情（exe/static/dynamic 全展开）
 */
const BugLogDisplay: React.FC<BugLogDisplayProps> = ({
  bugLogStr,
  fontSize = 12,
  requireAltForTooltip = true,
  tooltipScale = 1,
}) => {
  const [isAltPressed, setIsAltPressed] = React.useState(false);
  const [showErrorTooltip, setShowErrorTooltip] = React.useState(false);
  const [showWarningTooltip, setShowWarningTooltip] = React.useState(false);

  React.useEffect(() => {
    if (!requireAltForTooltip) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.altKey) setIsAltPressed(true);
    };
    const handleKeyUp = (e: KeyboardEvent) => {
      if (!e.altKey) setIsAltPressed(false);
    };

    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('keyup', handleKeyUp);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      window.removeEventListener('keyup', handleKeyUp);
    };
  }, [requireAltForTooltip]);

  // 尝试解析结构化 bugLog
  const bugLog = parseBugLog(bugLogStr);

  // 如果不是结构化格式，返回 null（不显示）
  if (!bugLog) return null;

  const { totalErrors, totalWarnings } = getBugLogSummary(bugLog);

  // 没有任何错误和警告，不显示
  if (totalErrors === 0 && totalWarnings === 0) return null;

  const scaledFontSize = Math.round(fontSize * tooltipScale);
  const tooltipWidth = Math.round(400 * tooltipScale);

  const canShowTooltip = !requireAltForTooltip || isAltPressed;

  // 渲染错误徽章（带独立 Tooltip）
  const errorBadge = totalErrors > 0 ? (
    <Tooltip
      title={canShowTooltip ? renderErrorTooltip(bugLog, scaledFontSize, tooltipWidth, totalErrors) : undefined}
      color="#1a1a2e"
      overlayInnerStyle={{
        padding: 16 * tooltipScale,
        width: tooltipWidth + 32 * tooltipScale,
      }}
      mouseEnterDelay={0}
      mouseLeaveDelay={0.1}
      open={canShowTooltip ? showErrorTooltip : false}
      onOpenChange={(visible) => {
        if (canShowTooltip) {
          setShowErrorTooltip(visible);
        }
      }}
    >
      <span
        style={{
          fontSize: `${fontSize - 2}px`,
          color: '#ff6b6b',
          fontWeight: 600,
          cursor: 'pointer',
        }}
      >
        ❌{totalErrors}
      </span>
    </Tooltip>
  ) : null;

  // 渲染警告徽章（带独立 Tooltip）
  const warningBadge = totalWarnings > 0 ? (
    <Tooltip
      title={canShowTooltip ? renderWarningTooltip(bugLog, scaledFontSize, tooltipWidth, totalWarnings) : undefined}
      color="#1a1a2e"
      overlayInnerStyle={{
        padding: 16 * tooltipScale,
        width: tooltipWidth + 32 * tooltipScale,
      }}
      mouseEnterDelay={0}
      mouseLeaveDelay={0.1}
      open={canShowTooltip ? showWarningTooltip : false}
      onOpenChange={(visible) => {
        if (canShowTooltip) {
          setShowWarningTooltip(visible);
        }
      }}
    >
      <span
        style={{
          fontSize: `${fontSize - 2}px`,
          color: '#f59e0b',
          fontWeight: 600,
          cursor: 'pointer',
        }}
      >
        ⚠{totalWarnings}
      </span>
    </Tooltip>
  ) : null;

  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 2,
      }}
    >
      {errorBadge}
      {warningBadge}
    </span>
  );
};

export default BugLogDisplay;
