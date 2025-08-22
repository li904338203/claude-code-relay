package controller

import (
	"claude-code-relay/common"
	"claude-code-relay/constant"
	"claude-code-relay/model"
	"claude-code-relay/service"
	"crypto/rand"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// BillingController 计费控制器
type BillingController struct {
	billingService *service.BillingService
}

// NewBillingController 创建计费控制器实例
func NewBillingController() *BillingController {
	return &BillingController{
		billingService: service.NewBillingService(),
	}
}

// ==================== 内部计费接口 ====================

// CheckQuota 检查用户配额
func (bc *BillingController) CheckQuota(c *gin.Context) {
	var req model.CheckQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "参数错误: " + err.Error(),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	response, err := bc.billingService.CheckQuota(req.UserID, req.EstimatedCost)
	if err != nil {
		common.SysError("Failed to check quota: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "检查配额失败",
				"type":    "internal_error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ProcessDeduction 处理扣费
func (bc *BillingController) ProcessDeduction(c *gin.Context) {
	var req model.DeductionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "参数错误: " + err.Error(),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	response, err := bc.billingService.ProcessDeduction(&req)
	if err != nil {
		common.SysError("Failed to process deduction: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "扣费处理失败",
				"type":    "internal_error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ==================== 用户接口 ====================

// GetUserBalance 获取用户余额和套餐信息
func (bc *BillingController) GetUserBalance(c *gin.Context) {
	user := c.MustGet("user").(*model.User)

	stats, err := bc.billingService.GetUserBillingStats(user.ID)
	if err != nil {
		common.SysError("Failed to get user billing stats: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "获取用户计费信息失败",
				"type":    "internal_error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// RedeemCard 兑换充值卡
func (bc *BillingController) RedeemCard(c *gin.Context) {
	user := c.MustGet("user").(*model.User)

	var req struct {
		CardCode string `json:"card_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "参数错误: " + err.Error(),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	err := bc.billingService.RedeemCard(user.ID, req.CardCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "充值卡兑换成功",
	})
}

// GetUserPlans 获取用户套餐列表
func (bc *BillingController) GetUserPlans(c *gin.Context) {
	user := c.MustGet("user").(*model.User)

	var plans []model.UserCardPlan
	err := model.DB.Where("user_id = ?", user.ID).
		Preload("RechargeCard").
		Order("created_at DESC").
		Find(&plans).Error

	if err != nil {
		common.SysError("Failed to get user plans: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "获取用户套餐失败",
				"type":    "internal_error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    plans,
	})
}

// GetConsumptionHistory 获取消费历史
func (bc *BillingController) GetConsumptionHistory(c *gin.Context) {
	user := c.MustGet("user").(*model.User)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	var logs []model.ConsumptionLog

	query := model.DB.Where("user_id = ?", user.ID)

	// 计算总数
	if err := query.Model(&model.ConsumptionLog{}).Count(&total).Error; err != nil {
		common.SysError("Failed to count consumption logs: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "获取消费记录失败",
				"type":    "internal_error",
			},
		})
		return
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := query.Preload("UserCardPlan").Preload("ApiKey").Preload("Account").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&logs).Error; err != nil {
		common.SysError("Failed to get consumption logs: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "获取消费记录失败",
				"type":    "internal_error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"logs":       logs,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// ==================== 管理员接口 ====================

// GenerateRechargeCards 生成充值卡
func (bc *BillingController) GenerateRechargeCards(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	if user.Role != constant.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"message": "权限不足",
				"type":    "permission_denied",
			},
		})
		return
	}

	var req model.CreateRechargeCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "参数错误: " + err.Error(),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	// 验证参数
	if req.CardType == "usage_count" && (req.UsageCount == nil || *req.UsageCount <= 0) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "次数卡必须指定有效的使用次数",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	if req.CardType == "time_limit" {
		if req.TimeType == nil || req.DurationDays == nil || req.DailyLimit == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"message": "时间卡必须指定时间类型、有效天数和每日限制",
					"type":    "invalid_request_error",
				},
			})
			return
		}
		if *req.DurationDays <= 0 || *req.DailyLimit <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"message": "时间卡的有效天数和每日限制必须大于0",
					"type":    "invalid_request_error",
				},
			})
			return
		}
	}

	// 生成充值卡
	cards := make([]model.RechargeCard, req.Count)
	now := time.Now()

	// 解析过期时间
	var expiredAt *time.Time
	if req.ExpiredAt != nil {
		parsed, err := time.Parse("2006-01-02 15:04:05", *req.ExpiredAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"message": "过期时间格式错误，应为：2006-01-02 15:04:05",
					"type":    "invalid_request_error",
				},
			})
			return
		}
		expiredAt = &parsed
	}

	for i := 0; i < req.Count; i++ {
		card := model.RechargeCard{
			CardCode:  generateCardCode(),
			CardType:  req.CardType,
			Value:     req.Value,
			Status:    "unused",
			BatchID:   req.BatchID,
			CreatedBy: &user.ID,
			ExpiredAt: expiredAt,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if req.CardType == "usage_count" {
			card.UsageCount = *req.UsageCount
		} else if req.CardType == "time_limit" {
			card.TimeType = req.TimeType
			card.DurationDays = *req.DurationDays
			card.DailyLimit = *req.DailyLimit
		}

		cards[i] = card
	}

	// 批量插入数据库
	if err := model.DB.CreateInBatches(cards, 100).Error; err != nil {
		common.SysError("Failed to create recharge cards: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "生成充值卡失败",
				"type":    "internal_error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("成功生成 %d 张充值卡", req.Count),
		"data":    gin.H{"count": req.Count},
	})
}

// GetRechargeCards 查询充值卡列表
func (bc *BillingController) GetRechargeCards(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	if user.Role != constant.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"message": "权限不足",
				"type":    "permission_denied",
			},
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	cardType := c.Query("card_type")
	batchID := c.Query("batch_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	var cards []model.RechargeCard

	query := model.DB.Model(&model.RechargeCard{})

	// 添加筛选条件
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if cardType != "" {
		query = query.Where("card_type = ?", cardType)
	}
	if batchID != "" {
		query = query.Where("batch_id = ?", batchID)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		common.SysError("Failed to count recharge cards: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "查询充值卡失败",
				"type":    "internal_error",
			},
		})
		return
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := query.Preload("User").Preload("CreatedByUser").
		Order("id DESC").
		Offset(offset).Limit(pageSize).
		Find(&cards).Error; err != nil {
		common.SysError("Failed to get recharge cards: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "查询充值卡失败",
				"type":    "internal_error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"cards":      cards,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// UpdateCardStatus 修改充值卡状态
func (bc *BillingController) UpdateCardStatus(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	if user.Role != constant.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"message": "权限不足",
				"type":    "permission_denied",
			},
		})
		return
	}

	cardID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "无效的卡ID",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=unused used expired disabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "参数错误: " + err.Error(),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	var card model.RechargeCard
	if err := model.DB.First(&card, cardID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"message": "充值卡不存在",
				"type":    "not_found",
			},
		})
		return
	}

	card.Status = req.Status
	if err := model.DB.Save(&card).Error; err != nil {
		common.SysError("Failed to update card status: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "更新充值卡状态失败",
				"type":    "internal_error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "充值卡状态更新成功",
	})
}

// GetAllUserPlans 获取所有用户套餐列表（管理员）
func (bc *BillingController) GetAllUserPlans(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	if user.Role != constant.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"message": "权限不足",
				"type":    "permission_denied",
			},
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	planType := c.Query("plan_type")
	userID := c.Query("user_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	var plans []model.UserCardPlan

	query := model.DB.Model(&model.UserCardPlan{})

	// 添加筛选条件
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if planType != "" {
		query = query.Where("plan_type = ?", planType)
	}
	if userID != "" {
		if uid, err := strconv.ParseUint(userID, 10, 32); err == nil {
			query = query.Where("user_id = ?", uid)
		}
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		common.SysError("Failed to count user plans: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "查询用户套餐失败",
				"type":    "internal_error",
			},
		})
		return
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := query.Preload("User").Preload("RechargeCard").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&plans).Error; err != nil {
		common.SysError("Failed to get user plans: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "查询用户套餐失败",
				"type":    "internal_error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"plans":      plans,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// UpdateUserPlanStatus 修改用户套餐状态
func (bc *BillingController) UpdateUserPlanStatus(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	if user.Role != constant.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"message": "权限不足",
				"type":    "permission_denied",
			},
		})
		return
	}

	planID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "无效的套餐ID",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active expired exhausted disabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "参数错误: " + err.Error(),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	var plan model.UserCardPlan
	if err := model.DB.Preload("User").First(&plan, planID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"message": "用户套餐不存在",
				"type":    "not_found",
			},
		})
		return
	}

	plan.Status = req.Status
	if err := model.DB.Save(&plan).Error; err != nil {
		common.SysError("Failed to update user plan status: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "更新用户套餐状态失败",
				"type":    "internal_error",
			},
		})
		return
	}

	// 记录操作日志
	common.SysLog(fmt.Sprintf("管理员 %s 将用户 %d 的套餐 %d 状态修改为 %s", user.Username, plan.UserID, plan.ID, req.Status))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户套餐状态更新成功",
	})
}

// RechargeUserBalance 手动充值用户余额
func (bc *BillingController) RechargeUserBalance(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	if user.Role != constant.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"message": "权限不足",
				"type":    "permission_denied",
			},
		})
		return
	}

	var req model.RechargeBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "参数错误: " + err.Error(),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	// 检查目标用户是否存在
	var targetUser model.User
	if err := model.DB.First(&targetUser, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"message": "目标用户不存在",
				"type":    "not_found",
			},
		})
		return
	}

	err := bc.billingService.RechargeBalance(&req, user.ID)
	if err != nil {
		common.SysError("Failed to recharge balance: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "充值失败",
				"type":    "internal_error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("为用户 %s 充值 $%.4f 成功", targetUser.Username, req.Amount),
	})
}

// GetConsumptionStats 获取消费统计
func (bc *BillingController) GetConsumptionStats(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	if user.Role != constant.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"message": "权限不足",
				"type":    "permission_denied",
			},
		})
		return
	}

	// 获取总消费统计
	var totalStats struct {
		TotalConsumption float64 `json:"total_consumption"`
		TotalUsers       int64   `json:"total_users"`
		TotalRequests    int64   `json:"total_requests"`
	}

	if err := model.DB.Model(&model.ConsumptionLog{}).
		Select("COALESCE(SUM(cost_usd), 0) as total_consumption, COUNT(*) as total_requests").
		Scan(&totalStats).Error; err != nil {
		common.SysError("Failed to get total stats: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "获取统计数据失败",
				"type":    "internal_error",
			},
		})
		return
	}

	if err := model.DB.Model(&model.User{}).Count(&totalStats.TotalUsers).Error; err != nil {
		common.SysError("Failed to count users: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "获取用户统计失败",
				"type":    "internal_error",
			},
		})
		return
	}

	// 获取今日统计
	today := time.Now().Format("2006-01-02")
	var todayStats struct {
		TodayConsumption float64 `json:"today_consumption"`
		TodayRequests    int64   `json:"today_requests"`
	}

	if err := model.DB.Model(&model.ConsumptionLog{}).
		Where("DATE(created_at) = ?", today).
		Select("COALESCE(SUM(cost_usd), 0) as today_consumption, COUNT(*) as today_requests").
		Scan(&todayStats).Error; err != nil {
		common.SysError("Failed to get today stats: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "获取今日统计失败",
				"type":    "internal_error",
			},
		})
		return
	}

	// 获取用户消费排行榜（前10名）
	var userRanking []struct {
		UserID       uint    `json:"user_id"`
		Username     string  `json:"username"`
		Consumption  float64 `json:"consumption"`
		RequestCount int64   `json:"request_count"`
	}

	if err := model.DB.Table("consumption_logs").
		Select("consumption_logs.user_id, users.username, COALESCE(SUM(consumption_logs.cost_usd), 0) as consumption, COUNT(*) as request_count").
		Joins("LEFT JOIN users ON consumption_logs.user_id = users.id").
		Group("consumption_logs.user_id, users.username").
		Order("consumption DESC").
		Limit(10).
		Scan(&userRanking).Error; err != nil {
		common.SysError("Failed to get user ranking: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "获取用户排行榜失败",
				"type":    "internal_error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_stats":  totalStats,
			"today_stats":  todayStats,
			"user_ranking": userRanking,
		},
	})
}

// GetBillingConfig 获取计费配置
func (bc *BillingController) GetBillingConfig(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	if user.Role != constant.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"message": "权限不足",
				"type":    "permission_denied",
			},
		})
		return
	}

	var configs []model.BillingConfig
	if err := model.DB.Find(&configs).Error; err != nil {
		common.SysError("Failed to get billing config: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "获取计费配置失败",
				"type":    "internal_error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    configs,
	})
}

// UpdateBillingConfig 更新计费配置
func (bc *BillingController) UpdateBillingConfig(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	if user.Role != constant.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"message": "权限不足",
				"type":    "permission_denied",
			},
		})
		return
	}

	var req struct {
		ConfigKey   string  `json:"config_key" binding:"required"`
		ConfigValue string  `json:"config_value" binding:"required"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "参数错误: " + err.Error(),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	err := bc.billingService.SetBillingConfig(req.ConfigKey, req.ConfigValue, req.Description)
	if err != nil {
		common.SysError("Failed to update billing config: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "更新计费配置失败",
				"type":    "internal_error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "计费配置更新成功",
	})
}

// ==================== 工具函数 ====================

// generateCardCode 生成卡密
func generateCardCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 16

	b := make([]byte, codeLength)
	if _, err := rand.Read(b); err != nil {
		// 如果随机数生成失败，使用时间戳作为后备
		return fmt.Sprintf("CARD%d", time.Now().UnixNano())
	}

	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}

	// 格式化为 XXXX-XXXX-XXXX-XXXX 格式
	code := string(b)
	return fmt.Sprintf("%s-%s-%s-%s",
		code[0:4], code[4:8], code[8:12], code[12:16])
}
