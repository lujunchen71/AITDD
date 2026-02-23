package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// 根据错误类型返回不同的响应
			c.JSON(500, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": err.Error(),
				},
				"timestamp": time.Now().UnixMilli(),
			})
		}
	}
}
