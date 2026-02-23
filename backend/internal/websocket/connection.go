package websocket

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有来源 (生产环境应限制)
		return true
	},
}

// Handler 存储WebSocket Hub
var Handler *WSHandler

// WSHandler WebSocket处理器
type WSHandler struct {
	Hub *Hub
}

// NewWSHandler 创建WebSocket处理器
func NewWSHandler() *WSHandler {
	hub := NewHub()
	go hub.Run()

	return &WSHandler{
		Hub: hub,
	}
}

// HandleWebSocket 处理WebSocket连接
func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// 生成客户端ID
	clientID := c.Query("clientId")
	if clientID == "" {
		clientID = uuid.New().String()
	}

	client := &Client{
		ID:   clientID,
		Conn: conn,
		Send: make(chan []byte, 256),
		Hub:  h.Hub,
	}

	h.Hub.Register <- client

	// 返回客户端ID
	welcomeMsg, _ := json.Marshal(Message{
		Type:    "connected",
		Payload: map[string]string{"clientId": clientID},
	})
	conn.WriteMessage(websocket.TextMessage, welcomeMsg)

	// 启动读写协程
	go client.WritePump()
	go client.ReadPump()
}
