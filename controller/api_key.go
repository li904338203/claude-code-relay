package controller

import (
	"claude-code-relay/common"
	"claude-code-relay/constant"
	"claude-code-relay/model"
	"claude-code-relay/service"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateApiKey 创建API Key
func CreateApiKey(c *gin.Context) {
	var req model.CreateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
			"code":  constant.InvalidParams,
		})
		return
	}

	// 从认证中获取用户ID
	user := c.MustGet("user").(*model.User)
	userID := user.ID

	apiKey, err := service.CreateApiKey(userID, &req)
	if err != nil {
		var statusCode int
		var code int
		switch err.Error() {
		case "API Key名称不能为空", "指定的分组不存在", "过期时间不能早于当前时间":
			statusCode = http.StatusBadRequest
			code = constant.InvalidParams
		default:
			statusCode = http.StatusInternalServerError
			code = constant.InternalServerError
		}
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": constant.Success,
		"data": gin.H{
			"key": apiKey.Key,
		},
	})
}

// GetApiKey 获取API Key详情
func GetApiKey(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的API Key ID",
			"code":  constant.InvalidParams,
		})
		return
	}

	// 从认证中获取用户ID
	user := c.MustGet("user").(*model.User)
	userID := user.ID

	apiKey, err := service.GetApiKeyById(uint(idInt), userID)
	if err != nil {
		var statusCode int
		var code int
		if err.Error() == "API Key不存在" {
			statusCode = http.StatusNotFound
			code = constant.InvalidParams
		} else {
			statusCode = http.StatusInternalServerError
			code = constant.InternalServerError
		}
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取API Key成功",
		"code":    constant.Success,
		"data":    apiKey,
	})
}

// UpdateApiKey 更新API Key
func UpdateApiKey(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的API Key ID",
			"code":  constant.InvalidParams,
		})
		return
	}

	var req model.UpdateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
			"code":  constant.InvalidParams,
		})
		return
	}

	// 从认证中获取用户ID
	user := c.MustGet("user").(*model.User)
	userID := user.ID

	apiKey, err := service.UpdateApiKey(uint(idInt), userID, &req)
	if err != nil {
		var statusCode int
		var code int
		switch err.Error() {
		case "API Key不存在", "指定的分组不存在", "过期时间不能早于当前时间":
			statusCode = http.StatusBadRequest
			code = constant.InvalidParams
		default:
			statusCode = http.StatusInternalServerError
			code = constant.InternalServerError
		}
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新API Key成功",
		"code":    constant.Success,
		"data":    apiKey,
	})
}

// DeleteApiKey 删除API Key
func DeleteApiKey(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的API Key ID",
			"code":  constant.InvalidParams,
		})
		return
	}

	// 从认证中获取用户ID
	user := c.MustGet("user").(*model.User)
	userID := user.ID

	err = service.DeleteApiKey(uint(idInt), userID)
	if err != nil {
		var statusCode int
		var code int
		if err.Error() == "API Key不存在" {
			statusCode = http.StatusNotFound
			code = constant.InvalidParams
		} else {
			statusCode = http.StatusInternalServerError
			code = constant.InternalServerError
		}
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除API Key成功",
		"code":    constant.Success,
	})
}

// UpdateApiKeyStatus 更新API Key状态
func UpdateApiKeyStatus(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的API Key ID",
			"code":  constant.InvalidParams,
		})
		return
	}

	var req struct {
		Status *int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
			"code":  constant.InvalidParams,
		})
		return
	}

	// 验证status值范围
	if req.Status == nil || (*req.Status != 0 && *req.Status != 1) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "status参数必须为0或1",
			"code":  constant.InvalidParams,
		})
		return
	}

	// 从认证中获取用户ID
	user := c.MustGet("user").(*model.User)
	userID := user.ID

	err = service.UpdateApiKeyStatusCom(uint(idInt), userID, *req.Status)
	if err != nil {
		var statusCode int
		var code int
		if err.Error() == "API Key不存在" {
			statusCode = http.StatusNotFound
			code = constant.InvalidParams
		} else {
			statusCode = http.StatusInternalServerError
			code = constant.InternalServerError
		}
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	statusText := "禁用"
	if *req.Status == 1 {
		statusText = "启用"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新API Key状态成功",
		"code":    constant.Success,
		"data": gin.H{
			"status": statusText,
		},
	})
}

// GetApiKeys 获取API Key列表
func GetApiKeys(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 可选的分组ID过滤
	var groupID *uint
	if groupIDStr := c.Query("group_id"); groupIDStr != "" {
		if groupIDInt, err := strconv.ParseUint(groupIDStr, 10, 32); err == nil {
			groupIDValue := uint(groupIDInt)
			groupID = &groupIDValue
		}
	}

	// 从认证中获取用户ID
	user := c.MustGet("user").(*model.User)
	userID := user.ID

	result, err := service.GetApiKeys(page, limit, userID, groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取API Key列表成功",
		"code":    constant.Success,
		"data":    result,
	})
}

// AdminGetApiKeys 管理员获取所有用户的API Key列表
func AdminGetApiKeys(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 可选的用户ID过滤
	var userID *uint
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userIDInt, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			userIDValue := uint(userIDInt)
			userID = &userIDValue
		}
	}

	// 可选的分组ID过滤
	var groupID *uint
	if groupIDStr := c.Query("group_id"); groupIDStr != "" {
		if groupIDInt, err := strconv.ParseUint(groupIDStr, 10, 32); err == nil {
			groupIDValue := uint(groupIDInt)
			groupID = &groupIDValue
		}
	}

	result, err := service.AdminGetApiKeys(page, limit, userID, groupID)
	if err != nil {
		common.SysLog(fmt.Sprintf("AdminGetApiKeys error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	common.SysLog(fmt.Sprintf("AdminGetApiKeys result: page=%d, limit=%d, userID=%v, groupID=%v, total=%d, apiKeysCount=%d",
		page, limit, userID, groupID, result.Total, len(result.ApiKeys)))

	c.JSON(http.StatusOK, gin.H{
		"message": "获取API Key列表成功",
		"code":    constant.Success,
		"data":    result,
	})
}

// GetApiKeyInfo 根据API Key获取统计信息和日志（公开接口，不需要登录）
func GetApiKeyInfo(c *gin.Context) {
	// 获取API Key，支持从URL参数或查询参数获取
	apiKeyStr := c.Param("api_key")
	if apiKeyStr == "" {
		apiKeyStr = c.Query("api_key")
	}

	if apiKeyStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "API Key不能为空",
			"code":  constant.InvalidParams,
		})
		return
	}

	// 验证API Key是否存在且有效
	apiKey, err := model.GetApiKeyByKey(apiKeyStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "API Key不存在或已失效",
			"code":  constant.InvalidParams,
		})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 获取30天统计数据
	statsReq := &model.StatsQueryRequest{
		ApiKeyID: &apiKey.ID,
	}

	// 使用最近30天作为默认时间范围
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day()-30, 0, 0, 0, 0, now.Location())
	endTime := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	statsReq.StartTime = &startTime
	statsReq.EndTime = &endTime

	// 获取完整统计数据
	stats, err := model.GetCompleteStats(statsReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取统计数据失败",
			"code":  constant.InternalServerError,
		})
		return
	}

	// 获取日志列表
	filters := &model.LogFilters{
		ApiKeyID:  &apiKey.ID,
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	logs, total, err := model.GetLogsWithFilters(filters, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取日志数据失败",
			"code":  constant.InternalServerError,
		})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, gin.H{
		"code": constant.Success,
		"data": gin.H{
			"api_key_info": gin.H{
				"id":     apiKey.ID,
				"name":   apiKey.Name,
				"status": apiKey.Status,
			},
			"stats": stats,
			"logs": gin.H{
				"list":  logs,
				"total": total,
				"page":  page,
				"limit": limit,
			},
		},
	})
}
