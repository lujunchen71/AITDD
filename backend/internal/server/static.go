package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed static/*
var staticFS embed.FS

// StaticHandler 静态文件处理器
type StaticHandler struct {
	fileSystem http.FileSystem
}

// NewStaticHandler 创建静态文件处理器
func NewStaticHandler() *StaticHandler {
	// 在开发模式下，使用本地文件系统
	// 在生产模式下，使用嵌入的文件系统
	return &StaticHandler{
		fileSystem: http.Dir("./frontend/dist"),
	}
}

// ServeHTTP 处理静态文件请求
func (h *StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// 如果是API请求，跳过
	if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/health") {
		http.NotFound(w, r)
		return
	}

	// 尝试打开文件
	f, err := h.fileSystem.Open(path)
	if err != nil {
		// 如果文件不存在，返回index.html（SPA路由）
		f, err = h.fileSystem.Open("/index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}
	defer f.Close()

	// 获取文件信息
	stat, err := f.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// 如果是目录，尝试返回index.html
	if stat.IsDir() {
		f, err = h.fileSystem.Open(path + "/index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()
		stat, err = f.Stat()
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}

	// 设置内容类型
	contentType := getContentType(path)
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	// 返回文件内容
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), f)
}

// getContentType 获取内容类型
func getContentType(path string) string {
	ext := strings.ToLower(path[strings.LastIndex(path, "."):])
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	default:
		return ""
	}
}

// getEmbeddedFS 获取嵌入的文件系统
func getEmbeddedFS() (http.FileSystem, error) {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return nil, err
	}
	return http.FS(sub), nil
}
