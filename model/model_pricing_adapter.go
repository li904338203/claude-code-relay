package model

import (
	"claude-code-relay/common"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ModelPricingAdapter 模型定价服务适配器，实现了common.ModelPricingService接口
type ModelPricingAdapter struct{}

// NewModelPricingAdapter 创建模型定价服务适配器
func NewModelPricingAdapter() *ModelPricingAdapter {
	return &ModelPricingAdapter{}
}

// GetAllCurrentModelPricing 获取所有当前有效的模型定价
func (adapter *ModelPricingAdapter) GetAllCurrentModelPricing() (map[string]common.ModelPricing, error) {
	if DB == nil {
		return nil, errors.New("数据库连接未初始化")
	}

	// 获取所有启用的模型及其有效定价
	var models []ModelConfig
	now := time.Now()

	err := DB.Preload("Pricing", func(db *gorm.DB) *gorm.DB {
		return db.Where("status = 1").
			Where("effective_time <= ?", now).
			Where("expire_time IS NULL OR expire_time > ?", now).
			Order("effective_time DESC")
	}).Where("status = 1").Find(&models).Error

	if err != nil {
		return nil, err
	}

	result := make(map[string]common.ModelPricing)
	for _, model := range models {
		// 获取最新的有效定价
		if len(model.Pricing) > 0 {
			pricing := model.Pricing[0] // 由于按时间倒序排列，第一个是最新的
			result[model.Name] = common.ModelPricing{
				Input:      pricing.InputPrice,
				Output:     pricing.OutputPrice,
				CacheWrite: pricing.CacheWritePrice,
				CacheRead:  pricing.CacheReadPrice,
			}
		}
	}

	// 如果数据库中没有数据，返回空map让CostCalculator使用硬编码数据
	return result, nil
}

// GetCurrentModelPricing 获取指定模型的当前有效定价
func (adapter *ModelPricingAdapter) GetCurrentModelPricing(modelName string) (common.ModelPricing, bool, error) {
	if DB == nil {
		return common.ModelPricing{}, false, errors.New("数据库连接未初始化")
	}

	if modelName == "" {
		modelName = "unknown"
	}

	// 查询模型的当前有效定价
	var pricing ModelPricing
	now := time.Now()

	err := DB.Joins("LEFT JOIN model_configs ON model_configs.id = model_pricing.model_id").
		Where("model_configs.name = ? AND model_configs.status = 1", modelName).
		Where("model_pricing.status = 1").
		Where("model_pricing.effective_time <= ?", now).
		Where("model_pricing.expire_time IS NULL OR model_pricing.expire_time > ?", now).
		Order("model_pricing.effective_time DESC").
		First(&pricing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ModelPricing{}, false, nil // 没有找到，让调用方使用默认值
		}
		return common.ModelPricing{}, false, err
	}

	return common.ModelPricing{
		Input:      pricing.InputPrice,
		Output:     pricing.OutputPrice,
		CacheWrite: pricing.CacheWritePrice,
		CacheRead:  pricing.CacheReadPrice,
	}, true, nil
}

// InitializeModelPricingService 初始化模型定价服务
func InitializeModelPricingService() {
	// 创建适配器并设置到common包中
	adapter := NewModelPricingAdapter()
	common.SetModelPricingService(adapter)
	common.SysLog("Model pricing service initialized with database adapter")
}
