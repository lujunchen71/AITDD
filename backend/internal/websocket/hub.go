package websocket

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

// Client 表示一个WebSocket客户端连接
type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
	Hub  *Hub
}

// Message 表示WebSocket消息
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Hub 管理所有WebSocket连接
type Hub struct {
	Clients    map[string]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
	mu         sync.RWMutex
}

// NewHub 创建新的Hub
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte, 256),
	}
}

// Run 启动Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.ID] = client
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.ID]; ok {
				delete(h.Clients, client.ID)
				close(client.Send)
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			h.mu.RLock()
			for _, client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client.ID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// SendToClient 向特定客户端发送消息
func (h *Hub) SendToClient(clientID string, message []byte) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, ok := h.Clients[clientID]; ok {
		select {
		case client.Send <- message:
			return true
		default:
			return false
		}
	}
	return false
}

// BroadcastMessage 广播消息到所有客户端
func (h *Hub) BroadcastMessage(msgType string, payload interface{}) error {
	message := Message{
		Type:    msgType,
		Payload: payload,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.Broadcast <- data
	return nil
}

// GetClientCount 获取连接的客户端数量
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Clients)
}

// ReadPump 从客户端读取消息
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		// 处理收到的消息 (心跳等)
		var msg Message
		if err := json.Unmarshal(message, &msg); err == nil {
			if msg.Type == "ping" {
				// 回复pong
				pong, _ := json.Marshal(Message{Type: "pong"})
				c.Send <- pong
			}
		}
	}
}

// WritePump 向客户端写入消息
func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
	}
}
