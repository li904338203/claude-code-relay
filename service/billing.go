package service

import (
	"claude-code-relay/common"
	"claude-code-relay/model"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// BillingService 计费服务结构体
type BillingService struct{}

// NewBillingService 创建计费服务实例
func NewBillingService() *BillingService {
	return &BillingService{}
}

// CheckQuota 检查用户是否有足够的配额进行API调用
func (bs *BillingService) CheckQuota(userID uint, estimatedCost float64) (*model.CheckQuotaResponse, error) {
	// 1. 检查时间卡套餐
	timeCardPlan, err := bs.getActiveTimeCardPlan(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check time card plan: %v", err)
	}

	if timeCardPlan != nil {
		dailyRemaining := timeCardPlan.DailyLimit - timeCardPlan.TodayUsed
		if dailyRemaining > 0 {
			return &model.CheckQuotaResponse{
				HasQuota:       true,
				QuotaType:      "time_limit",
				DailyRemaining: &dailyRemaining,
				Message:        "使用时间卡套餐",
			}, nil
		}
	}

	// 2. 检查次数卡套餐
	usageCardPlan, err := bs.getActiveUsageCardPlan(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check usage card plan: %v", err)
	}

	if usageCardPlan != nil && usageCardPlan.RemainingUsage > 0 {
		return &model.CheckQuotaResponse{
			HasQuota:       true,
			QuotaType:      "usage_count",
			RemainingUsage: &usageCardPlan.RemainingUsage,
			Message:        "使用次数卡套餐",
		}, nil
	}

	// 3. 检查余额
	userBalance, err := bs.getUserBalance(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %v", err)
	}

	if userBalance.Balance >= estimatedCost {
		return &model.CheckQuotaResponse{
			HasQuota:         true,
			QuotaType:        "balance",
			RemainingBalance: &userBalance.Balance,
			Message:          "使用余额扣费",
		}, nil
	}

	// 4. 没有足够的配额
	return &model.CheckQuotaResponse{
		HasQuota: false,
		Message:  "余额不足且无可用套餐",
	}, nil
}

// ProcessDeduction 处理扣费
func (bs *BillingService) ProcessDeduction(req *model.DeductionRequest) (*model.DeductionResponse, error) {
	// 添加调试日志 - 扣费请求
	common.SysLog(fmt.Sprintf("[BILLING_DEDUCTION] Starting deduction for User ID: %d, Cost: $%.6f, API Key ID: %d",
		req.UserID, req.CostUSD, getUintFromPointer(req.ApiKeyID)))

	// 开启事务
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 优先检查时间卡套餐
	timeCardPlan, err := bs.getActiveTimeCardPlanWithTx(tx, req.UserID)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to check time card plan: %v", err)
	}

	if timeCardPlan != nil && timeCardPlan.DailyLimit > timeCardPlan.TodayUsed {
		// 使用时间卡扣费
		response, err := bs.deductFromTimeCard(tx, timeCardPlan, req)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		tx.Commit()
		return response, nil
	}

	// 2. 检查次数卡套餐
	usageCardPlan, err := bs.getActiveUsageCardPlanWithTx(tx, req.UserID)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to check usage card plan: %v", err)
	}

	if usageCardPlan != nil && usageCardPlan.RemainingUsage > 0 {
		// 使用次数卡扣费
		response, err := bs.deductFromUsageCard(tx, usageCardPlan, req)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		tx.Commit()
		return response, nil
	}

	// 3. 使用余额扣费
	userBalance, err := bs.getUserBalanceWithTx(tx, req.UserID)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get user balance: %v", err)
	}

	if userBalance.Balance < req.CostUSD {
		tx.Rollback()
		common.SysError(fmt.Sprintf("[BILLING_DEDUCTION] Insufficient balance for User ID %d: Balance=%.4f, Required=%.6f",
			req.UserID, userBalance.Balance, req.CostUSD))
		return &model.DeductionResponse{
			Success: false,
			Message: "余额不足",
		}, nil
	}

	response, err := bs.deductFromBalance(tx, userBalance, req)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	common.SysLog(fmt.Sprintf("[BILLING_DEDUCTION] Deduction successful for User ID: %d, Type: %s",
		req.UserID, response.DeductionType))
	return response, nil
}

// getUintFromPointer 辅助函数：从指针获取uint值，如果为nil返回0
func getUintFromPointer(ptr *uint) uint {
	if ptr == nil {
		return 0
	}
	return *ptr
}

// deductFromTimeCard 从时间卡扣费
func (bs *BillingService) deductFromTimeCard(tx *gorm.DB, plan *model.UserCardPlan, req *model.DeductionRequest) (*model.DeductionResponse, error) {
	// 检查是否需要重置今日使用次数
	today := time.Now().Format("2006-01-02")
	if plan.LastResetDate == nil || plan.LastResetDate.Format("2006-01-02") != today {
		plan.TodayUsed = 0
		todayTime, _ := time.Parse("2006-01-02", today)
		plan.LastResetDate = &todayTime
	}

	// 增加今日使用次数
	plan.TodayUsed++

	// 更新套餐记录
	if err := tx.Save(plan).Error; err != nil {
		return nil, fmt.Errorf("failed to update time card plan: %v", err)
	}

	// 记录消费日志
	consumptionLog := &model.ConsumptionLog{
		UserID:              req.UserID,
		PlanID:              &plan.ID,
		RequestID:           req.RequestID,
		ApiKeyID:            req.ApiKeyID,
		AccountID:           req.AccountID,
		CostUSD:             req.CostUSD,
		UsageCount:          1,
		DeductionType:       "time_limit",
		InputTokens:         req.InputTokens,
		OutputTokens:        req.OutputTokens,
		CacheReadTokens:     req.CacheReadTokens,
		CacheCreationTokens: req.CacheCreationTokens,
		TotalTokens:         req.InputTokens + req.OutputTokens + req.CacheReadTokens + req.CacheCreationTokens,
		Model:               req.Model,
		PlatformType:        req.PlatformType,
		IsStream:            req.IsStream,
	}

	if err := tx.Create(consumptionLog).Error; err != nil {
		return nil, fmt.Errorf("failed to create consumption log: %v", err)
	}

	dailyRemaining := plan.DailyLimit - plan.TodayUsed
	return &model.DeductionResponse{
		Success:       true,
		DeductionType: "time_limit",
		CostUSD:       0, // 时间卡套餐内免费
		Message:       fmt.Sprintf("时间卡扣费成功，今日剩余：%d次", dailyRemaining),
	}, nil
}

// deductFromUsageCard 从次数卡扣费
func (bs *BillingService) deductFromUsageCard(tx *gorm.DB, plan *model.UserCardPlan, req *model.DeductionRequest) (*model.DeductionResponse, error) {
	// 扣除1次使用次数
	plan.UsedUsage++
	plan.RemainingUsage--

	// 检查是否用完
	if plan.RemainingUsage <= 0 {
		plan.Status = "exhausted"
	}

	// 更新套餐记录
	if err := tx.Save(plan).Error; err != nil {
		return nil, fmt.Errorf("failed to update usage card plan: %v", err)
	}

	// 记录消费日志
	consumptionLog := &model.ConsumptionLog{
		UserID:              req.UserID,
		PlanID:              &plan.ID,
		RequestID:           req.RequestID,
		ApiKeyID:            req.ApiKeyID,
		AccountID:           req.AccountID,
		CostUSD:             req.CostUSD,
		UsageCount:          1,
		DeductionType:       "usage_count",
		InputTokens:         req.InputTokens,
		OutputTokens:        req.OutputTokens,
		CacheReadTokens:     req.CacheReadTokens,
		CacheCreationTokens: req.CacheCreationTokens,
		TotalTokens:         req.InputTokens + req.OutputTokens + req.CacheReadTokens + req.CacheCreationTokens,
		Model:               req.Model,
		PlatformType:        req.PlatformType,
		IsStream:            req.IsStream,
	}

	if err := tx.Create(consumptionLog).Error; err != nil {
		return nil, fmt.Errorf("failed to create consumption log: %v", err)
	}

	return &model.DeductionResponse{
		Success:        true,
		DeductionType:  "usage_count",
		CostUSD:        0, // 次数卡按次计费，不扣除美元
		RemainingUsage: &plan.RemainingUsage,
		Message:        fmt.Sprintf("次数卡扣费成功，剩余次数：%d", plan.RemainingUsage),
	}, nil
}

// deductFromBalance 从余额扣费
func (bs *BillingService) deductFromBalance(tx *gorm.DB, balance *model.UserBalance, req *model.DeductionRequest) (*model.DeductionResponse, error) {
	balanceBefore := balance.Balance

	// 扣除余额
	balance.Balance -= req.CostUSD
	balance.TotalConsumed += req.CostUSD

	// 更新余额记录
	if err := tx.Save(balance).Error; err != nil {
		return nil, fmt.Errorf("failed to update user balance: %v", err)
	}

	balanceAfter := balance.Balance

	// 记录消费日志
	consumptionLog := &model.ConsumptionLog{
		UserID:              req.UserID,
		RequestID:           req.RequestID,
		ApiKeyID:            req.ApiKeyID,
		AccountID:           req.AccountID,
		CostUSD:             req.CostUSD,
		UsageCount:          1,
		DeductionType:       "balance",
		InputTokens:         req.InputTokens,
		OutputTokens:        req.OutputTokens,
		CacheReadTokens:     req.CacheReadTokens,
		CacheCreationTokens: req.CacheCreationTokens,
		TotalTokens:         req.InputTokens + req.OutputTokens + req.CacheReadTokens + req.CacheCreationTokens,
		Model:               req.Model,
		PlatformType:        req.PlatformType,
		IsStream:            req.IsStream,
		BalanceBefore:       &balanceBefore,
		BalanceAfter:        &balanceAfter,
	}

	if err := tx.Create(consumptionLog).Error; err != nil {
		return nil, fmt.Errorf("failed to create consumption log: %v", err)
	}

	return &model.DeductionResponse{
		Success:          true,
		DeductionType:    "balance",
		CostUSD:          req.CostUSD,
		RemainingBalance: balanceAfter,
		Message:          fmt.Sprintf("余额扣费成功，剩余余额：$%.4f", balanceAfter),
	}, nil
}

// RechargeBalance 手动充值余额（管理员功能）
func (bs *BillingService) RechargeBalance(req *model.RechargeBalanceRequest, operatorID uint) error {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取用户余额
	userBalance, err := bs.getUserBalanceWithTx(tx, req.UserID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get user balance: %v", err)
	}

	// 增加余额
	userBalance.Balance += req.Amount
	userBalance.TotalRecharged += req.Amount

	// 更新余额记录
	if err := tx.Save(userBalance).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update user balance: %v", err)
	}

	// 记录充值日志
	rechargeLog := &model.RechargeLog{
		UserID:       req.UserID,
		Amount:       req.Amount,
		RechargeType: "manual",
		Description:  &req.Description,
		OperatorID:   &operatorID,
	}

	if err := tx.Create(rechargeLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create recharge log: %v", err)
	}

	tx.Commit()
	return nil
}

// RedeemCard 兑换充值卡
func (bs *BillingService) RedeemCard(userID uint, cardCode string) error {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查找充值卡，确保不是禁用状态
	var card model.RechargeCard
	if err := tx.Where("card_code = ? AND status = 'unused'", cardCode).First(&card).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("卡密不存在、已使用或已被禁用")
		}
		return fmt.Errorf("failed to find card: %v", err)
	}

	// 检查卡是否过期
	if card.ExpiredAt != nil && card.ExpiredAt.Before(time.Now()) {
		tx.Rollback()
		return errors.New("充值卡已过期")
	}

	// 更新卡状态
	now := time.Now()
	card.Status = "used"
	card.UserID = &userID
	card.UsedAt = &now

	if err := tx.Save(&card).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update card status: %v", err)
	}

	// 根据卡类型处理兑换
	switch card.CardType {
	case "balance":
		return bs.redeemBalanceCard(tx, userID, &card)
	case "usage_count":
		return bs.redeemUsageCard(tx, userID, &card)
	case "time_limit":
		return bs.redeemTimeCard(tx, userID, &card)
	default:
		tx.Rollback()
		return errors.New("未知的卡类型")
	}
}

// redeemBalanceCard 兑换余额卡
func (bs *BillingService) redeemBalanceCard(tx *gorm.DB, userID uint, card *model.RechargeCard) error {
	// 获取用户余额
	userBalance, err := bs.getUserBalanceWithTx(tx, userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get user balance: %v", err)
	}

	// 增加余额
	userBalance.Balance += card.Value
	userBalance.TotalRecharged += card.Value

	// 更新余额记录
	if err := tx.Save(userBalance).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update user balance: %v", err)
	}

	// 记录充值日志
	description := fmt.Sprintf("兑换余额卡: %s", card.CardCode)
	rechargeLog := &model.RechargeLog{
		UserID:       userID,
		CardID:       &card.ID,
		Amount:       card.Value,
		RechargeType: "card",
		Description:  &description,
	}

	if err := tx.Create(rechargeLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create recharge log: %v", err)
	}

	tx.Commit()
	return nil
}

// redeemUsageCard 兑换次数卡
func (bs *BillingService) redeemUsageCard(tx *gorm.DB, userID uint, card *model.RechargeCard) error {
	// 创建次数卡套餐
	plan := &model.UserCardPlan{
		UserID:         userID,
		CardID:         card.ID,
		PlanType:       "usage_count",
		TotalUsage:     card.UsageCount,
		UsedUsage:      0,
		RemainingUsage: card.UsageCount,
		Status:         "active",
	}

	if err := tx.Create(plan).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create usage card plan: %v", err)
	}

	tx.Commit()
	return nil
}

// redeemTimeCard 兑换时间卡
func (bs *BillingService) redeemTimeCard(tx *gorm.DB, userID uint, card *model.RechargeCard) error {
	// 计算开始和结束日期
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, card.DurationDays)

	// 创建时间卡套餐
	plan := &model.UserCardPlan{
		UserID:     userID,
		CardID:     card.ID,
		PlanType:   "time_limit",
		TimeType:   card.TimeType,
		DailyLimit: card.DailyLimit,
		StartDate:  &startDate,
		EndDate:    &endDate,
		TodayUsed:  0,
		Status:     "active",
	}

	if err := tx.Create(plan).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create time card plan: %v", err)
	}

	tx.Commit()
	return nil
}

// 获取用户活跃的时间卡套餐
func (bs *BillingService) getActiveTimeCardPlan(userID uint) (*model.UserCardPlan, error) {
	var plan model.UserCardPlan
	err := model.DB.Where("user_id = ? AND plan_type = 'time_limit' AND status = 'active' AND start_date <= ? AND end_date >= ?",
		userID, time.Now(), time.Now()).First(&plan).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &plan, nil
}

// 获取用户活跃的次数卡套餐
func (bs *BillingService) getActiveUsageCardPlan(userID uint) (*model.UserCardPlan, error) {
	var plan model.UserCardPlan
	err := model.DB.Where("user_id = ? AND plan_type = 'usage_count' AND status = 'active' AND remaining_usage > 0",
		userID).First(&plan).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &plan, nil
}

// 获取用户余额
func (bs *BillingService) getUserBalance(userID uint) (*model.UserBalance, error) {
	var balance model.UserBalance
	err := model.DB.Where("user_id = ?", userID).First(&balance).Error
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

// 带事务获取用户活跃的时间卡套餐
func (bs *BillingService) getActiveTimeCardPlanWithTx(tx *gorm.DB, userID uint) (*model.UserCardPlan, error) {
	var plan model.UserCardPlan
	err := tx.Where("user_id = ? AND plan_type = 'time_limit' AND status = 'active' AND start_date <= ? AND end_date >= ?",
		userID, time.Now(), time.Now()).First(&plan).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &plan, nil
}

// 带事务获取用户活跃的次数卡套餐
func (bs *BillingService) getActiveUsageCardPlanWithTx(tx *gorm.DB, userID uint) (*model.UserCardPlan, error) {
	var plan model.UserCardPlan
	err := tx.Where("user_id = ? AND plan_type = 'usage_count' AND status = 'active' AND remaining_usage > 0",
		userID).First(&plan).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &plan, nil
}

// 带事务获取用户余额
func (bs *BillingService) getUserBalanceWithTx(tx *gorm.DB, userID uint) (*model.UserBalance, error) {
	var balance model.UserBalance
	err := tx.Where("user_id = ?", userID).First(&balance).Error
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

// GetBillingConfig 获取计费配置
func (bs *BillingService) GetBillingConfig(key string) (string, error) {
	var config model.BillingConfig
	err := model.DB.Where("config_key = ?", key).First(&config).Error
	if err != nil {
		return "", err
	}
	return config.ConfigValue, nil
}

// SetBillingConfig 设置计费配置
func (bs *BillingService) SetBillingConfig(key, value string, description *string) error {
	var config model.BillingConfig
	err := model.DB.Where("config_key = ?", key).First(&config).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建新配置
		config = model.BillingConfig{
			ConfigKey:   key,
			ConfigValue: value,
			Description: description,
		}
		return model.DB.Create(&config).Error
	} else if err != nil {
		return err
	} else {
		// 更新现有配置
		config.ConfigValue = value
		if description != nil {
			config.Description = description
		}
		return model.DB.Save(&config).Error
	}
}

// IsUsageCardEnabled 检查是否启用次数卡
func (bs *BillingService) IsUsageCardEnabled() bool {
	value, err := bs.GetBillingConfig("enable_usage_cards")
	if err != nil {
		return true // 默认启用
	}
	enabled, _ := strconv.ParseBool(value)
	return enabled
}

// IsTimeCardEnabled 检查是否启用时间卡
func (bs *BillingService) IsTimeCardEnabled() bool {
	value, err := bs.GetBillingConfig("enable_time_cards")
	if err != nil {
		return true // 默认启用
	}
	enabled, _ := strconv.ParseBool(value)
	return enabled
}

// GetUserBillingStats 获取用户计费统计信息
func (bs *BillingService) GetUserBillingStats(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取用户余额
	balance, err := bs.getUserBalance(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %v", err)
	}
	stats["balance"] = balance

	// 获取用户套餐
	var timePlans []model.UserCardPlan
	if err := model.DB.Where("user_id = ? AND plan_type = 'time_limit' AND status = 'active'", userID).Find(&timePlans).Error; err != nil {
		return nil, fmt.Errorf("failed to get time plans: %v", err)
	}
	stats["time_plans"] = timePlans

	var usagePlans []model.UserCardPlan
	if err := model.DB.Where("user_id = ? AND plan_type = 'usage_count' AND status = 'active'", userID).Find(&usagePlans).Error; err != nil {
		return nil, fmt.Errorf("failed to get usage plans: %v", err)
	}
	stats["usage_plans"] = usagePlans

	// 获取消费统计
	var totalConsumption float64
	if err := model.DB.Model(&model.ConsumptionLog{}).Where("user_id = ?", userID).Select("COALESCE(SUM(cost_usd), 0)").Scan(&totalConsumption).Error; err != nil {
		return nil, fmt.Errorf("failed to get total consumption: %v", err)
	}
	stats["total_consumption"] = totalConsumption

	// 获取今日消费
	today := time.Now().Format("2006-01-02")
	var todayConsumption float64
	if err := model.DB.Model(&model.ConsumptionLog{}).Where("user_id = ? AND DATE(created_at) = ?", userID, today).Select("COALESCE(SUM(cost_usd), 0)").Scan(&todayConsumption).Error; err != nil {
		return nil, fmt.Errorf("failed to get today consumption: %v", err)
	}
	stats["today_consumption"] = todayConsumption

	return stats, nil
}

// ResetDailyUsage 重置时间卡的每日使用次数（定时任务使用）
func (bs *BillingService) ResetDailyUsage() error {
	now := time.Now()
	today := now.Format("2006-01-02")

	// 重置所有活跃时间卡的今日使用次数
	err := model.DB.Model(&model.UserCardPlan{}).
		Where("plan_type = 'time_limit' AND status = 'active' AND (last_reset_date IS NULL OR DATE(last_reset_date) != ?)", today).
		Updates(map[string]interface{}{
			"today_used":      0,
			"last_reset_date": now,
		}).Error

	if err != nil {
		return fmt.Errorf("failed to reset daily usage: %v", err)
	}

	common.SysLog("Daily usage reset completed")
	return nil
}

// CleanupExpiredPlans 清理过期的套餐（定时任务使用）
func (bs *BillingService) CleanupExpiredPlans() error {
	now := time.Now()

	// 将过期的时间卡设置为expired状态
	err := model.DB.Model(&model.UserCardPlan{}).
		Where("plan_type = 'time_limit' AND status = 'active' AND end_date < ?", now).
		Update("status", "expired").Error

	if err != nil {
		return fmt.Errorf("failed to cleanup expired time plans: %v", err)
	}

	// 清理过期的充值卡
	err = model.DB.Model(&model.RechargeCard{}).
		Where("status = 'unused' AND expired_at IS NOT NULL AND expired_at < ?", now).
		Update("status", "expired").Error

	if err != nil {
		return fmt.Errorf("failed to cleanup expired cards: %v", err)
	}

	common.SysLog("Expired plans cleanup completed")
	return nil
}
