package routes

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterCommonRoutes 注册通用路由（健康检查、状态等）
func RegisterCommonRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Proxy the public quote service so browsers do not hit its missing CORS headers.
	r.GET("/api/v1/public/quote", func(c *gin.Context) {
		client := &http.Client{Timeout: 5 * time.Second}
		request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, "https://api.imlcd.cn/yy/api.php", nil)
		if err != nil {
			c.String(http.StatusBadGateway, "quote unavailable")
			return
		}
		response, err := client.Do(request)
		if err != nil {
			c.String(http.StatusBadGateway, "quote unavailable")
			return
		}
		defer response.Body.Close()
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			c.String(http.StatusBadGateway, "quote unavailable")
			return
		}
		body, err := io.ReadAll(io.LimitReader(response.Body, 4096))
		if err != nil || strings.TrimSpace(string(body)) == "" {
			c.String(http.StatusBadGateway, "quote unavailable")
			return
		}
		c.Data(http.StatusOK, "text/plain; charset=utf-8", body)
	})

	// Claude Code 遥测日志（忽略，直接返回200）
	r.POST("/api/event_logging/batch", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Setup status endpoint (always returns needs_setup: false in normal mode)
	// This is used by the frontend to detect when the service has restarted after setup
	r.GET("/setup/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"needs_setup": false,
				"step":        "completed",
			},
		})
	})
}
