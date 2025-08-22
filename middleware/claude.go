package middleware

import (
	"claude-code-relay/common"
	"claude-code-relay/constant"
	"claude-code-relay/model"
	"claude-code-relay/service"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// SystemMessage 系统消息结构体
type SystemMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// RequestBody Claude API请求体结构
type RequestBody struct {
	System interface{} `json:"system"`
}

// ClaudeCodeAuth API Key鉴权中间件
func ClaudeCodeAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 判断是否来自真实的 Claude Code 请求
		if !isRealClaudeCodeRequest(c) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "仅支持来自 Claude Code 的请求",
				"code":  40003,
			})
			c.Abort()
			return
		}

		// 从多个可能的请求头中获取API Key
		apiKey := getApiKeyFromHeaders(c)
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少API Key",
				"code":  40001,
			})
			c.Abort()
			return
		}

		// 从数据库查询API Key
		keyInfo, err := model.GetApiKeyByKey(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的API Key",
				"code":  40001,
			})
			c.Abort()
			return
		}

		// 添加调试日志 - API Key认证
		common.SysLog(fmt.Sprintf("[API_KEY_AUTH] API Key: %s (masked), User ID: %d",
			maskApiKey(apiKey), keyInfo.UserID))

		// 判断是否达到每日限额
		if keyInfo.DailyLimit > 0 && keyInfo.TodayTotalCost >= keyInfo.DailyLimit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "API Key已达到每日使用限额",
				"code":  40004,
			})
			c.Abort()
			return
		}

		// API Key已经在model层验证了状态和过期时间
		// 将API Key信息存储到上下文中供后续使用
		c.Set("api_key_id", keyInfo.ID)
		c.Set("api_key", keyInfo)
		c.Set("user_id", keyInfo.UserID)
		c.Set("group_id", keyInfo.GroupID)

		// 计费检查：在请求前检查用户配额
		billingService := service.NewBillingService()

		// 预估请求费用（这里使用一个基础费用，实际费用会在请求后计算）
		estimatedCost := 0.5 // 预估最低费用，避免完全没有余额的用户发起请求

		// 检查用户配额
		common.SysLog(fmt.Sprintf("[QUOTA_CHECK] Checking quota for User ID: %d, Cost: $%.6f",
			keyInfo.UserID, estimatedCost))
		quotaResponse, err := billingService.CheckQuota(keyInfo.UserID, estimatedCost)
		if err != nil {
			common.SysError(fmt.Sprintf("[QUOTA_CHECK] Failed to check user quota for User ID %d: %v",
				keyInfo.UserID, err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "计费系统检查失败",
				"code":  50001,
			})
			c.Abort()
			return
		}

		if !quotaResponse.HasQuota {
			common.SysError(fmt.Sprintf("[QUOTA_CHECK] Insufficient quota for User ID %d: %s",
				keyInfo.UserID, quotaResponse.Message))
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error": "余额不足或无可用套餐，请充值后再试",
				"code":  40006,
			})
			c.Abort()
			return
		}

		common.SysLog(fmt.Sprintf("[QUOTA_CHECK] Quota check passed for User ID: %d, Type: %s",
			keyInfo.UserID, quotaResponse.QuotaType))

		// 将计费信息存储到上下文中
		c.Set("billing_service", billingService)
		c.Set("quota_response", quotaResponse)

		c.Next()
	}
}

// getApiKeyFromHeaders 从多个可能的请求头中提取API Key
func getApiKeyFromHeaders(c *gin.Context) string {
	// 1. 检查 X-API-Key
	if apiKey := c.GetHeader("X-api-key"); apiKey != "" {
		return apiKey
	}

	// 2. 检查 X-Goog-API-Key (Google Cloud API格式)
	if apiKey := c.GetHeader("X-Goog-API-Key"); apiKey != "" {
		return apiKey
	}

	// 3. 检查 Authorization Bearer Token
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			return strings.TrimSpace(authHeader[7:])
		}
	}

	// 4. 检查 API-Key
	if apiKey := c.GetHeader("API-Key"); apiKey != "" {
		return apiKey
	}

	// 5. 检查小写变体 (某些客户端可能发送小写头)
	if apiKey := c.GetHeader("x-api-key"); apiKey != "" {
		return apiKey
	}

	if apiKey := c.GetHeader("api-key"); apiKey != "" {
		return apiKey
	}

	return ""
}

// maskApiKey 遮蔽API Key用于安全日志输出
func maskApiKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
}

// isRealClaudeCodeRequest 判断是否是真实的 Claude Code 请求
func isRealClaudeCodeRequest(c *gin.Context) bool {
	// 检查 User-Agent 是否匹配 Claude Code 格式
	userAgent := c.GetHeader("User-Agent")
	isClaudeCodeUserAgent := isClaudeCodeUserAgent(userAgent)

	// 检查系统提示词
	hasClaudeCodeSystemPrompt := hasClaudeCodeSystemPrompt(c)

	// 只有当 User-Agent 匹配且系统提示词正确时，才认为是真实的 Claude Code 请求
	return isClaudeCodeUserAgent && hasClaudeCodeSystemPrompt
}

// isClaudeCodeUserAgent 检查 User-Agent 是否匹配 Claude Code 格式
func isClaudeCodeUserAgent(userAgent string) bool {
	if userAgent == "" {
		return false
	}
	// 匹配 claude-cli/x.x.x 格式
	matched, _ := regexp.MatchString(`claude-cli/\d+\.\d+\.\d+`, userAgent)
	return matched
}

// hasClaudeCodeSystemPrompt 检查请求中是否包含 Claude Code 系统提示词
func hasClaudeCodeSystemPrompt(c *gin.Context) bool {
	// 读取请求体
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return false
	}

	// 重新设置请求体，以便后续处理可以再次读取
	c.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	// 解析请求体为通用map结构
	var requestBody map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &requestBody); err != nil {
		return false
	}

	// 将解析后的请求体存储到上下文中，供后续使用
	c.Set("request_body", requestBody)

	// 检查system字段
	system, exists := requestBody["system"]
	if !exists || system == nil {
		return false
	}

	// 如果是字符串格式，一定不是真实的 Claude Code 请求
	if systemStr, ok := system.(string); ok {
		_ = systemStr // 避免未使用变量警告
		return false
	}

	// 处理数组格式
	if systemArray, ok := system.([]interface{}); ok && len(systemArray) > 0 {
		if firstItem, ok := systemArray[0].(map[string]interface{}); ok {
			// 检查第一个元素是否包含 Claude Code 提示词
			if itemType, exists := firstItem["type"]; exists && itemType == "text" {
				if text, exists := firstItem["text"]; exists && text == constant.ClaudeCodeSystemPrompt {
					return true
				}
			}
		}
	}

	return false
}

// BillingMiddleware 计费中间件 - 已废弃，计费逻辑已移至日志记录处
// 现在计费处理在 relay/claude.go 的 saveRequestLog 函数中进行
func BillingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 计费逻辑已移至日志记录处，确保数据一致性
		c.Next()
	}
}

// processPostRequestBilling 已废弃 - 计费逻辑已移至 relay/claude.go 的 saveRequestLog 函数
// 在那里使用真实的TokenUsage数据进行计费，确保数据一致性和准确性

// 以下函数已废弃 - 计费逻辑已移至 relay/claude.go 的 saveRequestLog 函数
// 现在直接使用 TokenUsage 数据进行计费，无需从响应中提取token信息
