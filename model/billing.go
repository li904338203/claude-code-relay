package model

import (
	"time"
)

// UserBalance 用户余额表
type UserBalance struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID         uint      `json:"user_id" gorm:"not null;index"`
	Balance        float64   `json:"balance" gorm:"type:decimal(10,4);default:0.0000;comment:美元余额"`
	FrozenBalance  float64   `json:"frozen_balance" gorm:"type:decimal(10,4);default:0.0000;comment:冻结余额"`
	TotalRecharged float64   `json:"total_recharged" gorm:"type:decimal(10,4);default:0.0000;comment:累计充值金额"`
	TotalConsumed  float64   `json:"total_consumed" gorm:"type:decimal(10,4);default:0.0000;comment:累计消费金额"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// 关联
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// RechargeCard 充值卡表
type RechargeCard struct {
	ID           uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	CardCode     string     `json:"card_code" gorm:"type:varchar(50);unique;not null;comment:卡密"`
	CardType     string     `json:"card_type" gorm:"type:enum('usage_count','time_limit','balance');not null;comment:卡类型"`
	UsageCount   int        `json:"usage_count" gorm:"default:0;comment:可用次数（次数卡）"`
	TimeType     *string    `json:"time_type" gorm:"type:enum('daily','weekly','monthly');comment:时间类型"`
	DurationDays int        `json:"duration_days" gorm:"default:0;comment:有效天数"`
	DailyLimit   int        `json:"daily_limit" gorm:"default:0;comment:每日使用限制"`
	Value        float64    `json:"value" gorm:"type:decimal(10,4);not null;comment:面值（美元）"`
	Status       string     `json:"status" gorm:"type:enum('unused','used','expired','disabled');default:unused"`
	UserID       *uint      `json:"user_id" gorm:"comment:使用用户ID"`
	UsedAt       *time.Time `json:"used_at" gorm:"comment:使用时间"`
	ExpiredAt    *time.Time `json:"expired_at" gorm:"comment:过期时间"`
	BatchID      *string    `json:"batch_id" gorm:"type:varchar(50);comment:批次ID"`
	CreatedBy    *uint      `json:"created_by" gorm:"comment:创建者用户ID"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// 关联
	User          *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CreatedByUser *User `json:"created_by_user,omitempty" gorm:"foreignKey:CreatedBy"`
}

// UserCardPlan 用户卡套餐表
type UserCardPlan struct {
	ID             uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID         uint       `json:"user_id" gorm:"not null;index"`
	CardID         uint       `json:"card_id" gorm:"not null;comment:关联充值卡ID"`
	PlanType       string     `json:"plan_type" gorm:"type:enum('usage_count','time_limit');not null"`
	TotalUsage     int        `json:"total_usage" gorm:"default:0;comment:总次数"`
	UsedUsage      int        `json:"used_usage" gorm:"default:0;comment:已用次数"`
	RemainingUsage int        `json:"remaining_usage" gorm:"default:0;comment:剩余次数"`
	TimeType       *string    `json:"time_type" gorm:"type:enum('daily','weekly','monthly')"`
	DailyLimit     int        `json:"daily_limit" gorm:"default:0"`
	StartDate      *time.Time `json:"start_date" gorm:"type:date;comment:开始日期"`
	EndDate        *time.Time `json:"end_date" gorm:"type:date;comment:结束日期"`
	TodayUsed      int        `json:"today_used" gorm:"default:0;comment:今日已用"`
	LastResetDate  *time.Time `json:"last_reset_date" gorm:"type:date;comment:上次重置日期"`
	Status         string     `json:"status" gorm:"type:enum('active','expired','exhausted','disabled');default:active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// 关联
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	RechargeCard *RechargeCard `json:"recharge_card,omitempty" gorm:"foreignKey:CardID"`
}

// ConsumptionLog 消费记录表
type ConsumptionLog struct {
	ID                  uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID              uint      `json:"user_id" gorm:"not null;index"`
	PlanID              *uint     `json:"plan_id" gorm:"comment:关联套餐ID"`
	RequestID           *string   `json:"request_id" gorm:"type:varchar(50);comment:请求ID"`
	ApiKeyID            *uint     `json:"api_key_id" gorm:"comment:API Key ID"`
	AccountID           *uint     `json:"account_id" gorm:"comment:账号ID"`
	CostUSD             float64   `json:"cost_usd" gorm:"type:decimal(10,6);not null;comment:消费美元"`
	UsageCount          int       `json:"usage_count" gorm:"default:1;comment:消费次数"`
	DeductionType       string    `json:"deduction_type" gorm:"type:enum('balance','usage_count','time_limit');not null;comment:扣费类型"`
	InputTokens         int       `json:"input_tokens" gorm:"default:0"`
	OutputTokens        int       `json:"output_tokens" gorm:"default:0"`
	CacheReadTokens     int       `json:"cache_read_tokens" gorm:"default:0"`
	CacheCreationTokens int       `json:"cache_creation_tokens" gorm:"default:0"`
	TotalTokens         int       `json:"total_tokens" gorm:"default:0"`
	Model               *string   `json:"model" gorm:"type:varchar(100)"`
	PlatformType        *string   `json:"platform_type" gorm:"type:varchar(50)"`
	IsStream            bool      `json:"is_stream" gorm:"default:false"`
	BalanceBefore       *float64  `json:"balance_before" gorm:"type:decimal(10,4);comment:扣费前余额"`
	BalanceAfter        *float64  `json:"balance_after" gorm:"type:decimal(10,4);comment:扣费后余额"`
	CreatedAt           time.Time `json:"created_at"`

	// 关联
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	UserCardPlan *UserCardPlan `json:"user_card_plan,omitempty" gorm:"foreignKey:PlanID"`
	ApiKey       *ApiKey       `json:"api_key,omitempty" gorm:"foreignKey:ApiKeyID"`
	Account      *Account      `json:"account,omitempty" gorm:"foreignKey:AccountID"`
}

// RechargeLog 充值记录表
type RechargeLog struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       uint      `json:"user_id" gorm:"not null;index"`
	CardID       *uint     `json:"card_id" gorm:"comment:充值卡ID"`
	Amount       float64   `json:"amount" gorm:"type:decimal(10,4);not null;comment:充值金额"`
	RechargeType string    `json:"recharge_type" gorm:"type:enum('card','manual','system');not null;comment:充值类型"`
	Description  *string   `json:"description" gorm:"type:text;comment:充值说明"`
	OperatorID   *uint     `json:"operator_id" gorm:"comment:操作员ID（管理员充值时）"`
	CreatedAt    time.Time `json:"created_at"`

	// 关联
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	RechargeCard *RechargeCard `json:"recharge_card,omitempty" gorm:"foreignKey:CardID"`
	Operator     *User         `json:"operator,omitempty" gorm:"foreignKey:OperatorID"`
}

// BillingConfig 计费配置表
type BillingConfig struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ConfigKey   string    `json:"config_key" gorm:"type:varchar(100);unique;not null"`
	ConfigValue string    `json:"config_value" gorm:"type:text;not null"`
	Description *string   `json:"description" gorm:"type:varchar(255);comment:配置说明"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 请求结构体定义

// CreateRechargeCardRequest 创建充值卡请求
type CreateRechargeCardRequest struct {
	CardType     string  `json:"card_type" binding:"required,oneof=usage_count time_limit balance"`
	Count        int     `json:"count" binding:"required,min=1,max=1000"`
	Value        float64 `json:"value" binding:"required,min=0.01"`
	UsageCount   *int    `json:"usage_count"`
	TimeType     *string `json:"time_type" binding:"omitempty,oneof=daily weekly monthly"`
	DurationDays *int    `json:"duration_days"`
	DailyLimit   *int    `json:"daily_limit"`
	BatchID      *string `json:"batch_id"`
	ExpiredAt    *string `json:"expired_at"`
}

// RechargeCardListRequest 充值卡列表请求
type RechargeCardListRequest struct {
	Page      int     `form:"page"`
	Limit     int     `form:"limit"`
	Status    *string `form:"status"`
	CardType  *string `form:"card_type"`
	BatchID   *string `form:"batch_id"`
	CreatedBy *uint   `form:"created_by"`
}

// RedeemCardRequest 兑换充值卡请求
type RedeemCardRequest struct {
	CardCode string `json:"card_code" binding:"required"`
}

// UserBalanceResponse 用户余额响应
type UserBalanceResponse struct {
	Balance        float64            `json:"balance"`
	FrozenBalance  float64            `json:"frozen_balance"`
	TotalRecharged float64            `json:"total_recharged"`
	TotalConsumed  float64            `json:"total_consumed"`
	Plans          []UserCardPlanInfo `json:"plans"`
}

// UserCardPlanInfo 用户套餐信息
type UserCardPlanInfo struct {
	ID             uint    `json:"id"`
	PlanType       string  `json:"plan_type"`
	TotalUsage     int     `json:"total_usage"`
	UsedUsage      int     `json:"used_usage"`
	RemainingUsage int     `json:"remaining_usage"`
	TimeType       *string `json:"time_type"`
	DailyLimit     int     `json:"daily_limit"`
	StartDate      *string `json:"start_date"`
	EndDate        *string `json:"end_date"`
	TodayUsed      int     `json:"today_used"`
	Status         string  `json:"status"`
	CreatedAt      string  `json:"created_at"`
}

// ConsumptionLogListRequest 消费记录列表请求
type ConsumptionLogListRequest struct {
	Page          int     `form:"page"`
	Limit         int     `form:"limit"`
	UserID        *uint   `form:"user_id"`
	DeductionType *string `form:"deduction_type"`
	StartTime     *string `form:"start_time"`
	EndTime       *string `form:"end_time"`
	Model         *string `form:"model"`
}

// BillingStatsResponse 计费统计响应
type BillingStatsResponse struct {
	TotalUsers       int64   `json:"total_users"`
	TotalBalance     float64 `json:"total_balance"`
	TotalConsumed    float64 `json:"total_consumed"`
	TodayConsumption float64 `json:"today_consumption"`
	ActiveUsageCards int64   `json:"active_usage_cards"`
	ActiveTimeCards  int64   `json:"active_time_cards"`
	UnusedCards      int64   `json:"unused_cards"`
	ExpiredCards     int64   `json:"expired_cards"`
}

// DeductionRequest 扣费请求
type DeductionRequest struct {
	UserID              uint    `json:"user_id" binding:"required"`
	CostUSD             float64 `json:"cost_usd" binding:"required,min=0"`
	RequestID           *string `json:"request_id"`
	ApiKeyID            *uint   `json:"api_key_id"`
	AccountID           *uint   `json:"account_id"`
	InputTokens         int     `json:"input_tokens"`
	OutputTokens        int     `json:"output_tokens"`
	CacheReadTokens     int     `json:"cache_read_tokens"`
	CacheCreationTokens int     `json:"cache_creation_tokens"`
	Model               *string `json:"model"`
	PlatformType        *string `json:"platform_type"`
	IsStream            bool    `json:"is_stream"`
}

// DeductionResponse 扣费响应
type DeductionResponse struct {
	Success          bool    `json:"success"`
	DeductionType    string  `json:"deduction_type"`
	CostUSD          float64 `json:"cost_usd"`
	RemainingBalance float64 `json:"remaining_balance,omitempty"`
	RemainingUsage   *int    `json:"remaining_usage,omitempty"`
	Message          string  `json:"message"`
}

// CheckQuotaRequest 检查额度请求
type CheckQuotaRequest struct {
	UserID        uint    `json:"user_id" binding:"required"`
	EstimatedCost float64 `json:"estimated_cost" binding:"required,min=0"`
}

// CheckQuotaResponse 检查额度响应
type CheckQuotaResponse struct {
	HasQuota         bool     `json:"has_quota"`
	QuotaType        string   `json:"quota_type"`
	RemainingBalance *float64 `json:"remaining_balance,omitempty"`
	RemainingUsage   *int     `json:"remaining_usage,omitempty"`
	DailyRemaining   *int     `json:"daily_remaining,omitempty"`
	Message          string   `json:"message"`
}

// RechargeBalanceRequest 手动充值余额请求（管理员）
type RechargeBalanceRequest struct {
	UserID      uint    `json:"user_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,min=0.01"`
	Description string  `json:"description"`
}

// TableName 方法用于指定表名
func (UserBalance) TableName() string    { return "user_balances" }
func (RechargeCard) TableName() string   { return "recharge_cards" }
func (UserCardPlan) TableName() string   { return "user_card_plans" }
func (ConsumptionLog) TableName() string { return "consumption_logs" }
func (RechargeLog) TableName() string    { return "recharge_logs" }
func (BillingConfig) TableName() string  { return "billing_config" }
