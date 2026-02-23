import { useEffect, useRef, useCallback, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { message } from 'antd';

interface WebSocketMessage {
  type: string;
  payload: any;
}

interface UseWebSocketOptions {
  url?: string;
  onMessage?: (data: WebSocketMessage) => void;
  onConnect?: () => void;
  onDisconnect?: () => void;
  reconnect?: boolean;
  reconnectInterval?: number;
}

const useWebSocket = (options: UseWebSocketOptions = {}) => {
  const {
    url = `ws://${window.location.hostname}:34567/ws`,
    onMessage,
    onConnect,
    onDisconnect,
    reconnect = true,
    reconnectInterval = 3000,
  } = options;

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const [connected, setConnected] = useState(false);
  const [clientId, setClientId] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    const ws = new WebSocket(url);

    ws.onopen = () => {
      setConnected(true);
      console.log('WebSocket connected');
      onConnect?.();
    };

    ws.onmessage = (event) => {
      try {
        const data: WebSocketMessage = JSON.parse(event.data);

        // 处理连接消息
        if (data.type === 'connected') {
          setClientId(data.payload.clientId);
          return;
        }

        // 处理心跳
        if (data.type === 'pong') {
          return;
        }

        // 处理通知消息
        if (data.type === 'notification') {
          queryClient.invalidateQueries({ queryKey: ['notifications'] });
          message.info(data.payload.title || '收到新通知');
        }

        // 处理任务更新
        if (data.type === 'task_updated') {
          queryClient.invalidateQueries({ queryKey: ['tasks'] });
        }

        // 处理模块更新
        if (data.type === 'module_updated') {
          queryClient.invalidateQueries({ queryKey: ['modules'] });
        }

        // 调用自定义消息处理
        onMessage?.(data);
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };

    ws.onclose = () => {
      setConnected(false);
      setClientId(null);
      console.log('WebSocket disconnected');
      onDisconnect?.();

      // 自动重连
      if (reconnect) {
        reconnectTimeoutRef.current = setTimeout(() => {
          connect();
        }, reconnectInterval);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    wsRef.current = ws;
  }, [url, onMessage, onConnect, onDisconnect, reconnect, reconnectInterval, queryClient]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    wsRef.current?.close();
    wsRef.current = null;
  }, []);

  const send = useCallback((type: string, payload: any) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type, payload }));
      return true;
    }
    return false;
  }, []);

  const ping = useCallback(() => {
    return send('ping', {});
  }, [send]);

  useEffect(() => {
    connect();

    // 心跳定时器
    const pingInterval = setInterval(ping, 30000);

    return () => {
      clearInterval(pingInterval);
      disconnect();
    };
  }, [connect, disconnect, ping]);

  return {
    connected,
    clientId,
    send,
    disconnect,
    reconnect: connect,
  };
};

export default useWebSocket;
