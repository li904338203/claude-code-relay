package router

import (
	"claude-code-relay/controller"
	"claude-code-relay/middleware"

	"github.com/gin-gonic/gin"
)

func SetClaudeCodeRouter(server *gin.Engine) {
	claude := server.Group("/claude-code")
	// api key 鉴权
	claude.Use(middleware.ClaudeCodeAuth())
	// 计费中间件
	claude.Use(middleware.BillingMiddleware())
	{
		// 对话接口
		claude.POST("/v1/messages", controller.GetMessages)
	}
}
