package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// ModelConfig 模型配置
type ModelConfig struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	Name          string `json:"name" gorm:"type:varchar(100);uniqueIndex;not null;comment:模型名称"`
	DisplayName   string `json:"display_name" gorm:"type:varchar(100);not null;comment:显示名称"`
	Provider      string `json:"provider" gorm:"type:varchar(50);not null;comment:提供商"`
	Category      string `json:"category" gorm:"type:varchar(50);not null;comment:模型类别"`
	Version       string `json:"version" gorm:"type:varchar(50);not null;comment:版本号"`
	Status        int    `json:"status" gorm:"default:1;comment:状态(1:启用 0:禁用)"`
	SortOrder     int    `json:"sort_order" gorm:"default:0;comment:排序权重"`
	Description   string `json:"description" gorm:"type:text;comment:模型描述"`
	MaxTokens     *int   `json:"max_tokens" gorm:"comment:最大token数"`
	ContextWindow *int   `json:"context_window" gorm:"comment:上下文窗口大小"`
	CreatedAt     Time   `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt     Time   `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`

	// 关联关系
	Pricing []*ModelPricing `json:"pricing,omitempty" gorm:"foreignKey:ModelID"`
}

// ModelPricing 模型定价
type ModelPricing struct {
	ID              uint    `json:"id" gorm:"primaryKey"`
	ModelID         uint    `json:"model_id" gorm:"not null;index;comment:模型ID"`
	InputPrice      float64 `json:"input_price" gorm:"type:decimal(10,6);not null;comment:输入价格(USD/1M tokens)"`
	OutputPrice     float64 `json:"output_price" gorm:"type:decimal(10,6);not null;comment:输出价格(USD/1M tokens)"`
	CacheWritePrice float64 `json:"cache_write_price" gorm:"type:decimal(10,6);default:0;comment:缓存写入价格(USD/1M tokens)"`
	CacheReadPrice  float64 `json:"cache_read_price" gorm:"type:decimal(10,6);default:0;comment:缓存读取价格(USD/1M tokens)"`
	EffectiveTime   Time    `json:"effective_time" gorm:"type:datetime;not null;comment:生效时间"`
	ExpireTime      *Time   `json:"expire_time" gorm:"type:datetime;comment:失效时间"`
	Status          int     `json:"status" gorm:"default:1;comment:状态(1:启用 0:禁用)"`
	CreatedAt       Time    `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt       Time    `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`

	// 关联关系
	Model *ModelConfig `json:"model,omitempty" gorm:"foreignKey:ModelID"`
}

// 表名
func (m *ModelConfig) TableName() string {
	return "model_configs"
}

func (mp *ModelPricing) TableName() string {
	return "model_pricing"
}

// CreateModelRequest 创建模型请求
type CreateModelRequest struct {
	Name          string `json:"name" binding:"required,min=1,max=100"`
	DisplayName   string `json:"display_name" binding:"required,min=1,max=100"`
	Provider      string `json:"provider" binding:"required,min=1,max=50"`
	Category      string `json:"category" binding:"required,min=1,max=50"`
	Version       string `json:"version" binding:"required,min=1,max=50"`
	Status        int    `json:"status" binding:"oneof=0 1"`
	SortOrder     int    `json:"sort_order"`
	Description   string `json:"description"`
	MaxTokens     *int   `json:"max_tokens"`
	ContextWindow *int   `json:"context_window"`
}

// UpdateModelRequest 更新模型请求
type UpdateModelRequest struct {
	DisplayName   string `json:"display_name" binding:"required,min=1,max=100"`
	Provider      string `json:"provider" binding:"required,min=1,max=50"`
	Category      string `json:"category" binding:"required,min=1,max=50"`
	Version       string `json:"version" binding:"required,min=1,max=50"`
	Status        int    `json:"status" binding:"oneof=0 1"`
	SortOrder     int    `json:"sort_order"`
	Description   string `json:"description"`
	MaxTokens     *int   `json:"max_tokens"`
	ContextWindow *int   `json:"context_window"`
}

// CreatePricingRequest 创建定价请求
type CreatePricingRequest struct {
	InputPrice      float64 `json:"input_price" binding:"required,min=0"`
	OutputPrice     float64 `json:"output_price" binding:"required,min=0"`
	CacheWritePrice float64 `json:"cache_write_price" binding:"min=0"`
	CacheReadPrice  float64 `json:"cache_read_price" binding:"min=0"`
	EffectiveTime   string  `json:"effective_time" binding:"required"`
	ExpireTime      *string `json:"expire_time"`
}

// ModelQueryParams 模型查询参数
type ModelQueryParams struct {
	Page     int    `form:"page" binding:"min=1"`
	Limit    int    `form:"limit" binding:"min=1,max=100"`
	Name     string `form:"name"`
	Provider string `form:"provider"`
	Status   *int   `form:"status" binding:"omitempty,oneof=0 1"`
}

// ModelListResponse 模型列表响应
type ModelListResponse struct {
	Models []ModelConfig `json:"models"`
	Total  int64         `json:"total"`
}

// =================== 数据库操作方法 ===================

// CreateModelConfig 创建模型配置
func CreateModelConfig(model *ModelConfig) error {
	model.ID = 0
	return DB.Create(model).Error
}

// GetModelConfigByID 根据ID获取模型配置
func GetModelConfigByID(id uint) (*ModelConfig, error) {
	var model ModelConfig
	err := DB.Preload("Pricing", "status = 1").First(&model, id).Error
	return &model, err
}

// GetModelConfigByName 根据名称获取模型配置
func GetModelConfigByName(name string) (*ModelConfig, error) {
	var model ModelConfig
	err := DB.Preload("Pricing", "status = 1").Where("name = ? AND status = 1", name).First(&model).Error
	return &model, err
}

// UpdateModelConfig 更新模型配置
func UpdateModelConfig(model *ModelConfig) error {
	return DB.Save(model).Error
}

// DeleteModelConfig 删除模型配置（软删除）
func DeleteModelConfig(id uint) error {
	return DB.Delete(&ModelConfig{}, id).Error
}

// GetModelConfigList 获取模型配置列表
func GetModelConfigList(params *ModelQueryParams) (*ModelListResponse, error) {
	var models []ModelConfig
	var total int64

	query := DB.Model(&ModelConfig{})

	// 条件筛选
	if params.Name != "" {
		query = query.Where("name LIKE ? OR display_name LIKE ?", "%"+params.Name+"%", "%"+params.Name+"%")
	}
	if params.Provider != "" {
		query = query.Where("provider = ?", params.Provider)
	}
	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.Limit
	err := query.Preload("Pricing", "status = 1").
		Order("sort_order ASC, created_at DESC").
		Limit(params.Limit).
		Offset(offset).
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	return &ModelListResponse{
		Models: models,
		Total:  total,
	}, nil
}

// GetActiveModels 获取所有启用的模型
func GetActiveModels() ([]ModelConfig, error) {
	var models []ModelConfig
	err := DB.Preload("Pricing", "status = 1").
		Where("status = 1").
		Order("sort_order ASC, created_at DESC").
		Find(&models).Error
	return models, err
}

// CreateModelPricing 创建模型定价
func CreateModelPricing(pricing *ModelPricing) error {
	// 检查模型是否存在
	var model ModelConfig
	if err := DB.First(&model, pricing.ModelID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("模型不存在")
		}
		return err
	}

	// 检查时间有效性
	if pricing.ExpireTime != nil && time.Time(pricing.EffectiveTime).After(time.Time(*pricing.ExpireTime)) {
		return errors.New("生效时间不能晚于失效时间")
	}

	pricing.ID = 0
	return DB.Create(pricing).Error
}

// GetCurrentModelPricing 获取模型当前有效定价
func GetCurrentModelPricing(modelName string) (*ModelPricing, error) {
	var pricing ModelPricing
	now := time.Now()

	err := DB.Joins("LEFT JOIN model_configs ON model_configs.id = model_pricing.model_id").
		Where("model_configs.name = ? AND model_configs.status = 1", modelName).
		Where("model_pricing.status = 1").
		Where("model_pricing.effective_time <= ?", now).
		Where("model_pricing.expire_time IS NULL OR model_pricing.expire_time > ?", now).
		Order("model_pricing.effective_time DESC").
		First(&pricing).Error

	return &pricing, err
}

// GetAllCurrentModelPricing 获取所有模型的当前有效定价
func GetAllCurrentModelPricing() (map[string]*ModelPricing, error) {
	var models []ModelConfig
	err := DB.Preload("Pricing", func(db *gorm.DB) *gorm.DB {
		now := time.Now()
		return db.Where("status = 1").
			Where("effective_time <= ?", now).
			Where("expire_time IS NULL OR expire_time > ?", now).
			Order("effective_time DESC")
	}).Where("status = 1").Find(&models).Error

	if err != nil {
		return nil, err
	}

	result := make(map[string]*ModelPricing)
	for _, model := range models {
		// 获取最新的有效定价
		if len(model.Pricing) > 0 {
			result[model.Name] = model.Pricing[0]
		}
	}

	return result, nil
}

// BatchUpdateModelStatus 批量更新模型状态
func BatchUpdateModelStatus(ids []uint, status int) error {
	return DB.Model(&ModelConfig{}).Where("id IN ?", ids).Update("status", status).Error
}

// GetModelPricingHistory 获取模型定价历史
func GetModelPricingHistory(modelID uint) ([]*ModelPricing, error) {
	var pricing []*ModelPricing
	err := DB.Where("model_id = ?", modelID).
		Order("effective_time DESC").
		Find(&pricing).Error
	return pricing, err
}

// UpdateModelPricing 更新定价
func UpdateModelPricing(pricing *ModelPricing) error {
	return DB.Save(pricing).Error
}

// DeleteModelPricing 删除定价
func DeleteModelPricing(id uint) error {
	return DB.Delete(&ModelPricing{}, id).Error
}
