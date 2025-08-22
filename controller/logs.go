package controller

import (
	"claude-code-relay/common"
	"claude-code-relay/constant"
	"claude-code-relay/model"
	"claude-code-relay/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// LogQueryRequest 日志查询请求参数
type LogQueryRequest struct {
	Page      int      `form:"page"`       // 页码，默认为1
	Limit     int      `form:"limit"`      // 每页数量，默认为10，最大100
	UserID    uint     `form:"user_id"`    // 用户ID筛选
	AccountID uint     `form:"account_id"` // 账号ID筛选
	ApiKeyID  uint     `form:"api_key_id"` // API Key ID筛选
	ModelName string   `form:"model_name"` // 模型名称筛选
	IsStream  *bool    `form:"is_stream"`  // 是否流式请求筛选
	StartTime string   `form:"start_time"` // 开始时间 格式: 2024-01-01 15:04:05
	EndTime   string   `form:"end_time"`   // 结束时间 格式: 2024-01-01 15:04:05
	MinCost   *float64 `form:"min_cost"`   // 最小费用筛选
	MaxCost   *float64 `form:"max_cost"`   // 最大费用筛选
}

// GetLogs 获取日志列表（支持多种筛选条件）
func GetLogs(c *gin.Context) {
	user := c.MustGet("user").(*model.User)

	var req LogQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "查询参数错误: " + err.Error(),
			"code":  constant.InvalidParams,
		})
		return
	}

	// 权限检查：非管理员只能查看自己的日志
	if user.Role != "admin" {
		req.UserID = user.ID
	}

	// 参数验证和默认值设置
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 10
	}

	logService := service.NewLogService()

	// 构建查询条件
	filters := buildLogFilters(&req)

	// 调用service层查询
	result, err := logService.GetLogsWithFilters(filters, req.Page, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取日志列表失败: " + err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取日志列表成功",
		"code":    constant.Success,
		"data":    result,
	})
}

// GetLogById 根据ID获取日志详情
func GetLogById(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "日志ID不能为空",
			"code":  constant.InvalidParams,
		})
		return
	}

	logService := service.NewLogService()
	log, err := logService.GetLogById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "日志不存在",
			"code":  constant.NotFound,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取日志详情成功",
		"code":    constant.Success,
		"data":    log,
	})
}

// GetMyLogs 获取当前用户的日志记录
func GetMyLogs(c *gin.Context) {
	// 从上下文获取当前用户信息
	user := c.MustGet("user").(*model.User)

	var req LogQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "查询参数错误: " + err.Error(),
			"code":  constant.InvalidParams,
		})
		return
	}

	// 强制设置为当前用户ID
	req.UserID = user.ID

	// 参数验证和默认值设置
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 10
	}

	logService := service.NewLogService()
	filters := buildLogFilters(&req)

	result, err := logService.GetLogsWithFilters(filters, req.Page, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取我的日志失败: " + err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取我的日志成功",
		"code":    constant.Success,
		"data":    result,
	})
}

// GetLogStats 获取日志统计信息
func GetLogStats(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	var userID *uint

	// 如果是管理员，可以查看所有用户或指定用户的统计
	if user.Role == "admin" {
		// 检查是否指定了用户ID
		if userIDParam := c.Query("user_id"); userIDParam != "" {
			if id, err := strconv.ParseUint(userIDParam, 10, 32); err == nil {
				uid := uint(id)
				userID = &uid
			}
		}
		// 如果没有指定用户ID，userID保持为nil，表示查看所有用户统计
	} else {
		// 普通用户只能查看自己的统计
		userID = &user.ID
	}

	logService := service.NewLogService()
	stats, err := logService.GetLogStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取统计信息失败: " + err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取统计信息成功",
		"code":    constant.Success,
		"data":    stats,
	})
}

// GetMyLogStats 获取当前用户的日志统计信息
func GetMyLogStats(c *gin.Context) {
	user := c.MustGet("user").(*model.User)

	logService := service.NewLogService()
	stats, err := logService.GetUserLogStats(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取我的统计信息失败: " + err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取我的统计信息成功",
		"code":    constant.Success,
		"data":    stats,
	})
}

// DeleteLogById 删除日志记录（管理员功能）
func DeleteLogById(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "日志ID不能为空",
			"code":  constant.InvalidParams,
		})
		return
	}

	logService := service.NewLogService()
	err := logService.DeleteLogById(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "删除日志失败: " + err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除日志成功",
		"code":    constant.Success,
	})
}

// DeleteExpiredLogs 删除过期日志（管理员功能）
func DeleteExpiredLogs(c *gin.Context) {
	monthsStr := c.Query("months")
	if monthsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请指定保留月数",
			"code":  constant.InvalidParams,
		})
		return
	}

	months, err := strconv.Atoi(monthsStr)
	if err != nil || months <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "保留月数必须是大于0的整数",
			"code":  constant.InvalidParams,
		})
		return
	}

	logService := service.NewLogService()
	deletedCount, err := logService.DeleteExpiredLogs(months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "删除过期日志失败: " + err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除过期日志成功",
		"code":    constant.Success,
		"data": gin.H{
			"deleted_count": deletedCount,
		},
	})
}

// GetUsageStats 获取使用统计数据
func GetUsageStats(c *gin.Context) {
	var req model.StatsQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "查询参数错误: " + err.Error(),
			"code":  constant.InvalidParams,
		})
		return
	}

	// 不再需要period和days参数验证

	// 解析时间参数
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse("2006-01-02 15:04:05", startTimeStr); err == nil {
			req.StartTime = &startTime
		} else {
			common.SysLog("解析开始时间失败: " + err.Error() + ", 时间字符串: " + startTimeStr)
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse("2006-01-02 15:04:05", endTimeStr); err == nil {
			req.EndTime = &endTime
		} else {
			common.SysLog("解析结束时间失败: " + err.Error() + ", 时间字符串: " + endTimeStr)
		}
	}

	// 权限检查：非管理员只能查看自己的统计
	user := c.MustGet("user").(*model.User)
	if user.Role != "admin" {
		req.UserID = &user.ID
	}

	logService := service.NewLogService()
	result, err := logService.GetCompleteStats(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取使用统计失败: " + err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取使用统计成功",
		"code":    constant.Success,
		"data":    result,
	})
}

// GetMyUsageStats 获取我的使用统计数据
func GetMyUsageStats(c *gin.Context) {
	user := c.MustGet("user").(*model.User)

	var req model.StatsQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "查询参数错误: " + err.Error(),
			"code":  constant.InvalidParams,
		})
		return
	}

	// 强制设置为当前用户
	req.UserID = &user.ID

	// 不再需要period和days参数验证

	// 解析时间参数
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse("2006-01-02 15:04:05", startTimeStr); err == nil {
			req.StartTime = &startTime
		} else {
			common.SysLog("解析开始时间失败: " + err.Error() + ", 时间字符串: " + startTimeStr)
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse("2006-01-02 15:04:05", endTimeStr); err == nil {
			req.EndTime = &endTime
		} else {
			common.SysLog("解析结束时间失败: " + err.Error() + ", 时间字符串: " + endTimeStr)
		}
	}

	logService := service.NewLogService()
	result, err := logService.GetCompleteStats(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取我的使用统计失败: " + err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取我的使用统计成功",
		"code":    constant.Success,
		"data":    result,
	})
}

// GetDashboardStats 获取仪表盘统计数据
func GetDashboardStats(c *gin.Context) {
	logService := service.NewLogService()
	result, err := logService.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取仪表盘统计数据失败: " + err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取仪表盘统计数据成功",
		"code":    constant.Success,
		"data":    result,
	})
}

// buildLogFilters 构建日志查询过滤条件
func buildLogFilters(req *LogQueryRequest) *model.LogFilters {
	filters := &model.LogFilters{}

	if req.UserID > 0 {
		filters.UserID = &req.UserID
	}

	if req.AccountID > 0 {
		filters.AccountID = &req.AccountID
	}

	if req.ApiKeyID > 0 {
		filters.ApiKeyID = &req.ApiKeyID
	}

	if req.ModelName != "" {
		filters.ModelName = &req.ModelName
	}

	if req.IsStream != nil {
		filters.IsStream = req.IsStream
	}

	// 解析时间范围
	if req.StartTime != "" {
		if startTime, err := time.Parse("2006-01-02 15:04:05", req.StartTime); err == nil {
			filters.StartTime = &startTime
		}
	}

	if req.EndTime != "" {
		if endTime, err := time.Parse("2006-01-02 15:04:05", req.EndTime); err == nil {
			filters.EndTime = &endTime
		}
	}

	// 费用范围筛选
	if req.MinCost != nil {
		filters.MinCost = req.MinCost
	}

	if req.MaxCost != nil {
		filters.MaxCost = req.MaxCost
	}

	return filters
}
