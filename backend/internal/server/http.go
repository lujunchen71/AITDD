package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aitdd/backend/internal/api"
)

// HTTPServer HTTP服务器
type HTTPServer struct {
	port   int
	server *http.Server
}

// NewHTTPServer 创建HTTP服务器
func NewHTTPServer(port int) *HTTPServer {
	return &HTTPServer{
		port: port,
	}
}

// Start 启动服务器
func (s *HTTPServer) Start() error {
	// 设置路由
	router := api.SetupRouter()

	// 配置静态文件服务
	s.setupStaticFiles(router)

	// 创建HTTP服务器
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	return s.server.ListenAndServe()
}

// Stop 停止服务器
func (s *HTTPServer) Stop() error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// setupStaticFiles 设置静态文件服务
func (s *HTTPServer) setupStaticFiles(router http.Handler) {
	// 静态文件服务在 routes.go 中处理
}
