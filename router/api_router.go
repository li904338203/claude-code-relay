package router

import (
	"claude-code-relay/controller"
	"claude-code-relay/middleware"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func SetAPIRouter(server *gin.Engine) {
	// 创建计费控制器实例
	billingController := controller.NewBillingController()
	// 健康检查
	server.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	// API路由组
	api := server.Group("/api/v1")
	api.Use(middleware.RateLimit(300, time.Minute))
	{
		// 公开接口
		auth := api.Group("/auth")
		{
			auth.POST("/login", controller.Login)
			auth.POST("/register", controller.Register)
			auth.POST("/send-verification-code", controller.SendVerificationCode)
			auth.GET("/api-key", controller.GetApiKeyInfo)          // 根据API Key查询统计信息（公开接口）
			auth.GET("/api-key/:api_key", controller.GetApiKeyInfo) // 支持URL路径参数方式
		}

		// 系统状态
		api.GET("/status", controller.GetStatus)

		// 公开的模型信息接口
		models := api.Group("/models")
		{
			models.GET("/active", controller.GetActiveModels)                 // 获取启用的模型列表
			models.GET("/current-pricing", controller.GetCurrentModelPricing) // 获取当前有效定价
		}

		// 需要认证的接口
		authenticated := api.Group("")
		authenticated.Use(middleware.Auth())
		{
			// 用户相关
			user := authenticated.Group("/user")
			{
				user.GET("/profile", controller.GetProfile)
				user.PUT("/profile", controller.UpdateProfile)
				user.PUT("/change-email", controller.ChangeEmail)
			}

			// 菜单相关
			authenticated.GET("/menu-list", controller.GetMenuList)

			// 分组相关
			group := authenticated.Group("/groups")
			{
				group.GET("/list", controller.GetGroups)            // 获取分组列表
				group.GET("/all", controller.GetAllGroups)          // 获取所有分组（用于下拉选择）
				group.POST("/create", controller.CreateGroup)       // 创建分组
				group.GET("/detail/:id", controller.GetGroup)       // 获取分组详情
				group.PUT("/update/:id", controller.UpdateGroup)    // 更新分组
				group.DELETE("/delete/:id", controller.DeleteGroup) // 删除分组
			}

			// 账号管理相关
			account := authenticated.Group("/accounts")
			{
				account.GET("/list", controller.GetAccountList)                                  // 获取账号列表
				account.POST("/create", controller.CreateAccount)                                // 创建账号
				account.GET("/detail/:id", controller.GetAccount)                                // 获取账号详情
				account.PUT("/update/:id", controller.UpdateAccount)                             // 更新账号
				account.DELETE("/delete/:id", controller.DeleteAccount)                          // 删除账号
				account.PUT("/update-active-status/:id", controller.UpdateAccountActiveStatus)   // 更新账号激活状态
				account.PUT("/update-current-status/:id", controller.UpdateAccountCurrentStatus) // 更新账号当前状态
				account.POST("/test/:id", controller.TestGetMessages)                            // 测试账号连通性
			}

			// Claude OAuth 相关
			oauth := authenticated.Group("/oauth")
			{
				oauth.GET("/generate-auth-url", controller.GetOAuthURL) // 获取OAuth授权URL
				oauth.POST("/exchange-code", controller.ExchangeCode)   // 验证授权码并获取token
			}

			// API Key 相关
			apikey := authenticated.Group("/api-keys")
			{
				apikey.GET("/list", controller.GetApiKeys)                      // 获取API Key列表
				apikey.POST("/create", controller.CreateApiKey)                 // 创建API Key
				apikey.GET("/detail/:id", controller.GetApiKey)                 // 获取API Key详情
				apikey.PUT("/update/:id", controller.UpdateApiKey)              // 更新API Key
				apikey.PUT("/update-status/:id", controller.UpdateApiKeyStatus) // 更新API Key状态
				apikey.DELETE("/delete/:id", controller.DeleteApiKey)           // 删除API Key
			}

			// 日志相关（用户接口）
			logs := authenticated.Group("/logs")
			{
				logs.GET("/my", controller.GetMyLogs)                   // 获取当前用户的日志记录
				logs.GET("/stats/my", controller.GetMyLogStats)         // 获取当前用户的日志统计
				logs.GET("/usage-stats/my", controller.GetMyUsageStats) // 获取当前用户的使用统计
				logs.GET("/detail/:id", controller.GetLogById)          // 获取日志详情
			}

			// 仪表盘数据接口
			authenticated.GET("/dashboard/stats", controller.GetDashboardStats) // 获取仪表盘统计数据

			// 计费系统相关接口
			billing := authenticated.Group("/billing")
			{
				// 用户计费接口
				billing.GET("/balance", billingController.GetUserBalance)            // 获取用户余额和套餐信息
				billing.POST("/redeem", billingController.RedeemCard)                // 充值卡兑换
				billing.GET("/plans", billingController.GetUserPlans)                // 获取用户套餐列表
				billing.GET("/consumption", billingController.GetConsumptionHistory) // 获取消费历史
			}

			// 内部计费接口（用于中间件调用）
			internal := authenticated.Group("/internal")
			{
				internal.POST("/billing/check-quota", billingController.CheckQuota)  // 检查用户配额
				internal.POST("/billing/deduct", billingController.ProcessDeduction) // 处理扣费
			}

			// 管理员接口
			admin := authenticated.Group("/admin")
			admin.Use(middleware.AdminAuth())
			{
				admin.GET("/users", controller.GetUsers)
				admin.POST("/users", controller.AdminCreateUser)
				admin.PUT("/users/:id/status", controller.AdminUpdateUserStatus)
				admin.GET("/logs", controller.GetApiLogs)
				admin.GET("/dashboard", controller.GetDashboard)

				// 日志管理（管理员专用）
				adminLogs := admin.Group("/logs")
				{
					adminLogs.GET("/list", controller.GetLogs)                 // 获取所有日志列表（支持筛选）
					adminLogs.GET("/stats", controller.GetLogStats)            // 获取日志统计（支持指定用户）
					adminLogs.GET("/usage-stats", controller.GetUsageStats)    // 获取使用统计（管理员可查看所有用户）
					adminLogs.GET("/detail/:id", controller.GetLogById)        // 获取日志详情
					adminLogs.DELETE("/delete/:id", controller.DeleteLogById)  // 删除指定日志
					adminLogs.DELETE("/cleanup", controller.DeleteExpiredLogs) // 删除过期日志
				}

				// 定时任务测试接口（管理员专用）
				admin.POST("/test/reset-stats", controller.ManualResetStats)       // 手动重置统计数据
				admin.POST("/test/clean-logs", controller.ManualCleanLogs)         // 手动清理过期日志
				admin.POST("/test/refresh-tokens", controller.ManualRefreshTokens) // 手动刷新Claude账号token

				// 计费管理接口（管理员专用）
				adminBilling := admin.Group("/billing")
				{
					// 充值卡管理
					adminBilling.POST("/cards/generate", billingController.GenerateRechargeCards) // 生成充值卡
					adminBilling.GET("/cards", billingController.GetRechargeCards)                // 查询充值卡列表
					adminBilling.PUT("/cards/:id/status", billingController.UpdateCardStatus)     // 修改充值卡状态

					// 用户套餐管理
					adminBilling.GET("/plans", billingController.GetAllUserPlans)                 // 获取所有用户套餐列表
					adminBilling.PUT("/plans/:id/status", billingController.UpdateUserPlanStatus) // 修改用户套餐状态

					// 用户余额管理
					adminBilling.POST("/balance/recharge", billingController.RechargeUserBalance) // 手动充值用户余额

					// 消费统计
					adminBilling.GET("/stats", billingController.GetConsumptionStats) // 获取消费统计

					// 系统配置
					adminBilling.GET("/config", billingController.GetBillingConfig)    // 获取计费配置
					adminBilling.PUT("/config", billingController.UpdateBillingConfig) // 更新计费配置
				}

				// 模型配置管理接口（管理员专用）
				adminModels := admin.Group("/models")
				{
					// 模型管理
					adminModels.POST("", controller.CreateModel)                       // 创建模型配置
					adminModels.GET("", controller.GetModelList)                       // 获取模型配置列表
					adminModels.GET("/:id", controller.GetModel)                       // 获取模型配置详情
					adminModels.PUT("/:id", controller.UpdateModel)                    // 更新模型配置
					adminModels.DELETE("/:id", controller.DeleteModel)                 // 删除模型配置
					adminModels.PUT("/status", controller.UpdateModelStatus)           // 批量更新模型状态
					adminModels.GET("/validate-name", controller.ValidateModelName)    // 验证模型名称
					adminModels.POST("/refresh-cache", controller.RefreshPricingCache) // 刷新定价缓存

					// 模型定价管理
					adminModels.POST("/:id/pricing", controller.CreateModelPricing)    // 创建模型定价
					adminModels.GET("/:id/pricing", controller.GetModelPricingHistory) // 获取模型定价历史
					adminModels.PUT("/pricing/:id", controller.UpdateModelPricing)     // 更新模型定价
					adminModels.DELETE("/pricing/:id", controller.DeleteModelPricing)  // 删除模型定价
				}

				// 账号管理接口（管理员专用）
				adminAccounts := admin.Group("/accounts")
				{
					adminAccounts.GET("", controller.GetAccountList) // 获取所有用户账号列表（复用现有接口，管理员可查看所有）
				}

				// 密钥管理接口（管理员专用）
				adminKeys := admin.Group("/keys")
				{
					adminKeys.GET("", controller.AdminGetApiKeys) // 获取所有用户API Key列表
				}

				// 分组管理接口（管理员专用）
				adminGroups := admin.Group("/groups")
				{
					adminGroups.GET("", controller.AdminGetGroups)        // 获取所有用户分组列表
					adminGroups.GET("/all", controller.AdminGetAllGroups) // 获取所有分组（用于下拉选择）
				}
			}

			// 通用日志接口（管理员权限）
			adminLogsAll := authenticated.Group("/logs")
			adminLogsAll.Use(middleware.AdminAuth())
			{
				adminLogsAll.GET("", controller.GetLogs)           // 获取所有日志列表
				adminLogsAll.GET("/stats", controller.GetLogStats) // 获取日志统计
			}
		}
	}

	// 前端静态文件服务
	server.Static("/assets", "./web/dist/assets")
	server.Static("/static", "./web/dist/static")

	// 前端路由处理 - 对于前端路由，返回 index.html
	server.NoRoute(func(c *gin.Context) {
		// 如果是 API 请求，返回404
		if strings.HasPrefix(c.Request.URL.Path, "/api/") || strings.HasPrefix(c.Request.URL.Path, "/health") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
			return
		}

		// 对于前端路由，返回 index.html
		c.File("./web/dist/index.html")
	})
}
