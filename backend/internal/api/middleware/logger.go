package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		// 简单的控制台日志
		if status >= 400 {
			gin.DefaultWriter.Write([]byte(
				"[" + time.Now().Format("2006/01/02 - 15:04:05") + "] " +
					method + " " + path + " " +
					string(rune(status)) + " " +
					latency.String() + "\n",
			))
		}
	}
}
