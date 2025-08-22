package service

import (
	"claude-code-relay/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ModelConfigService 模型配置服务
type ModelConfigService struct{}

// NewModelConfigService 创建模型配置服务实例
func NewModelConfigService() *ModelConfigService {
	return &ModelConfigService{}
}

// CreateModel 创建模型配置
func (s *ModelConfigService) CreateModel(req *model.CreateModelRequest) (*model.ModelConfig, error) {
	// 检查模型名称是否已存在
	var existingModel model.ModelConfig
	err := model.DB.Where("name = ?", req.Name).First(&existingModel).Error
	if err == nil {
		return nil, errors.New("模型名称已存在")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 创建模型配置
	modelConfig := &model.ModelConfig{
		Name:          req.Name,
		DisplayName:   req.DisplayName,
		Provider:      req.Provider,
		Category:      req.Category,
		Version:       req.Version,
		Status:        req.Status,
		SortOrder:     req.SortOrder,
		Description:   req.Description,
		MaxTokens:     req.MaxTokens,
		ContextWindow: req.ContextWindow,
	}

	err = model.CreateModelConfig(modelConfig)
	if err != nil {
		return nil, err
	}

	return modelConfig, nil
}

// GetModelByID 根据ID获取模型配置
func (s *ModelConfigService) GetModelByID(id uint) (*model.ModelConfig, error) {
	return model.GetModelConfigByID(id)
}

// GetModelByName 根据名称获取模型配置
func (s *ModelConfigService) GetModelByName(name string) (*model.ModelConfig, error) {
	return model.GetModelConfigByName(name)
}

// UpdateModel 更新模型配置
func (s *ModelConfigService) UpdateModel(id uint, req *model.UpdateModelRequest) (*model.ModelConfig, error) {
	// 获取现有模型
	modelConfig, err := model.GetModelConfigByID(id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	modelConfig.DisplayName = req.DisplayName
	modelConfig.Provider = req.Provider
	modelConfig.Category = req.Category
	modelConfig.Version = req.Version
	modelConfig.Status = req.Status
	modelConfig.SortOrder = req.SortOrder
	modelConfig.Description = req.Description
	modelConfig.MaxTokens = req.MaxTokens
	modelConfig.ContextWindow = req.ContextWindow

	err = model.UpdateModelConfig(modelConfig)
	if err != nil {
		return nil, err
	}

	return modelConfig, nil
}

// DeleteModel 删除模型配置
func (s *ModelConfigService) DeleteModel(id uint) error {
	return model.DeleteModelConfig(id)
}

// GetModelList 获取模型配置列表
func (s *ModelConfigService) GetModelList(params *model.ModelQueryParams) (*model.ModelListResponse, error) {
	// 设置默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	return model.GetModelConfigList(params)
}

// GetActiveModels 获取所有启用的模型
func (s *ModelConfigService) GetActiveModels() ([]model.ModelConfig, error) {
	return model.GetActiveModels()
}

// BatchUpdateStatus 批量更新模型状态
func (s *ModelConfigService) BatchUpdateStatus(ids []uint, status int) error {
	if len(ids) == 0 {
		return errors.New("ID列表不能为空")
	}

	if status != 0 && status != 1 {
		return errors.New("状态值必须是0或1")
	}

	return model.BatchUpdateModelStatus(ids, status)
}

// CreatePricing 创建模型定价
func (s *ModelConfigService) CreatePricing(modelID uint, req *model.CreatePricingRequest) (*model.ModelPricing, error) {
	// 解析时间
	effectiveTime, err := time.Parse("2006-01-02 15:04:05", req.EffectiveTime)
	if err != nil {
		return nil, errors.New("生效时间格式错误")
	}

	var expireTime *time.Time
	if req.ExpireTime != nil && *req.ExpireTime != "" {
		t, err := time.Parse("2006-01-02 15:04:05", *req.ExpireTime)
		if err != nil {
			return nil, errors.New("失效时间格式错误")
		}
		expireTime = &t
	}

	// 创建定价配置
	pricing := &model.ModelPricing{
		ModelID:         modelID,
		InputPrice:      req.InputPrice,
		OutputPrice:     req.OutputPrice,
		CacheWritePrice: req.CacheWritePrice,
		CacheReadPrice:  req.CacheReadPrice,
		EffectiveTime:   model.Time(effectiveTime),
		Status:          1,
	}

	if expireTime != nil {
		timeValue := model.Time(*expireTime)
		pricing.ExpireTime = &timeValue
	}

	err = model.CreateModelPricing(pricing)
	if err != nil {
		return nil, err
	}

	return pricing, nil
}

// GetPricingHistory 获取模型定价历史
func (s *ModelConfigService) GetPricingHistory(modelID uint) ([]*model.ModelPricing, error) {
	return model.GetModelPricingHistory(modelID)
}

// UpdatePricing 更新定价
func (s *ModelConfigService) UpdatePricing(pricingID uint, req *model.CreatePricingRequest) (*model.ModelPricing, error) {
	// 获取现有定价
	var pricing model.ModelPricing
	err := model.DB.First(&pricing, pricingID).Error
	if err != nil {
		return nil, err
	}

	// 解析时间
	effectiveTime, err := time.Parse("2006-01-02 15:04:05", req.EffectiveTime)
	if err != nil {
		return nil, errors.New("生效时间格式错误")
	}

	var expireTime *time.Time
	if req.ExpireTime != nil && *req.ExpireTime != "" {
		t, err := time.Parse("2006-01-02 15:04:05", *req.ExpireTime)
		if err != nil {
			return nil, errors.New("失效时间格式错误")
		}
		expireTime = &t
	}

	// 更新字段
	pricing.InputPrice = req.InputPrice
	pricing.OutputPrice = req.OutputPrice
	pricing.CacheWritePrice = req.CacheWritePrice
	pricing.CacheReadPrice = req.CacheReadPrice
	pricing.EffectiveTime = model.Time(effectiveTime)

	if expireTime != nil {
		timeValue := model.Time(*expireTime)
		pricing.ExpireTime = &timeValue
	} else {
		pricing.ExpireTime = nil
	}

	err = model.UpdateModelPricing(&pricing)
	if err != nil {
		return nil, err
	}

	return &pricing, nil
}

// DeletePricing 删除定价
func (s *ModelConfigService) DeletePricing(pricingID uint) error {
	return model.DeleteModelPricing(pricingID)
}

// GetCurrentModelPricing 获取模型当前有效定价
func (s *ModelConfigService) GetCurrentModelPricing(modelName string) (*model.ModelPricing, error) {
	return model.GetCurrentModelPricing(modelName)
}

// GetAllCurrentModelPricing 获取所有模型的当前有效定价
func (s *ModelConfigService) GetAllCurrentModelPricing() (map[string]*model.ModelPricing, error) {
	return model.GetAllCurrentModelPricing()
}

// ValidateModelName 验证模型名称
func (s *ModelConfigService) ValidateModelName(name string) error {
	if name == "" {
		return errors.New("模型名称不能为空")
	}

	// 检查是否已存在
	var count int64
	err := model.DB.Model(&model.ModelConfig{}).Where("name = ?", name).Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("模型名称已存在")
	}

	return nil
}

// SyncPricingToCache 同步定价到缓存（用于通知CostCalculator刷新缓存）
func (s *ModelConfigService) SyncPricingToCache() error {
	// 这里可以实现缓存刷新的逻辑
	// 例如：发送消息到消息队列，或者调用特定的缓存刷新接口
	return nil
}
