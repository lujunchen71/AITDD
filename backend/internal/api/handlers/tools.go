package handlers

import (
	"os/exec"
	"runtime"

	"github.com/aitdd/backend/internal/api"
	"github.com/gin-gonic/gin"
)

// OpenBrowser 打开浏览器
func OpenBrowser(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		api.ValidationError(c, "无效的请求数据", nil)
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", req.URL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", req.URL)
	case "darwin":
		cmd = exec.Command("open", req.URL)
	default:
		api.ValidationError(c, "不支持的操作系统", nil)
		return
	}

	if err := cmd.Start(); err != nil {
		api.InternalError(c, "打开浏览器失败")
		return
	}

	api.Success(c, gin.H{
		"opened": true,
		"url":    req.URL,
	})
}
