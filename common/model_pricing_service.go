package common

import (
	"errors"
	"sync"
)

// ModelPricingService 模型定价服务接口
type ModelPricingService interface {
	GetAllCurrentModelPricing() (map[string]ModelPricing, error)
	GetCurrentModelPricing(modelName string) (ModelPricing, bool, error)
}

// 全局的模型定价服务实例
var (
	modelPricingService ModelPricingService
	serviceMutex        sync.RWMutex
)

// SetModelPricingService 设置模型定价服务实例
func SetModelPricingService(service ModelPricingService) {
	serviceMutex.Lock()
	defer serviceMutex.Unlock()
	modelPricingService = service
}

// GetModelPricingService 获取模型定价服务实例
func GetModelPricingService() ModelPricingService {
	serviceMutex.RLock()
	defer serviceMutex.RUnlock()
	return modelPricingService
}

// DefaultModelPricingService 默认的模型定价服务实现（空实现，提示需要初始化数据库服务）
type DefaultModelPricingService struct{}

// NewDefaultModelPricingService 创建默认模型定价服务
func NewDefaultModelPricingService() *DefaultModelPricingService {
	return &DefaultModelPricingService{}
}

// GetAllCurrentModelPricing 获取所有当前有效的模型定价（返回错误提示需要初始化数据库）
func (s *DefaultModelPricingService) GetAllCurrentModelPricing() (map[string]ModelPricing, error) {
	return nil, errors.New("database model pricing service not initialized, please check database connection")
}

// GetCurrentModelPricing 获取指定模型的当前有效定价（返回错误提示需要初始化数据库）
func (s *DefaultModelPricingService) GetCurrentModelPricing(modelName string) (ModelPricing, bool, error) {
	return ModelPricing{}, false, errors.New("database model pricing service not initialized, please check database connection")
}

// DatabaseModelPricingService 基于数据库的模型定价服务实现
// 这个结构体将在初始化时由main包或其他地方设置
type DatabaseModelPricingService struct {
	// 这里可以包含数据库访问的方法或接口
	GetAllPricingFunc func() (map[string]ModelPricing, error)
	GetPricingFunc    func(modelName string) (ModelPricing, bool, error)
}

// NewDatabaseModelPricingService 创建基于数据库的模型定价服务
func NewDatabaseModelPricingService(
	getAllFunc func() (map[string]ModelPricing, error),
	getFunc func(modelName string) (ModelPricing, bool, error),
) *DatabaseModelPricingService {
	return &DatabaseModelPricingService{
		GetAllPricingFunc: getAllFunc,
		GetPricingFunc:    getFunc,
	}
}

// GetAllCurrentModelPricing 获取所有当前有效的模型定价
func (s *DatabaseModelPricingService) GetAllCurrentModelPricing() (map[string]ModelPricing, error) {
	if s.GetAllPricingFunc == nil {
		return nil, errors.New("数据库定价服务未正确初始化")
	}
	return s.GetAllPricingFunc()
}

// GetCurrentModelPricing 获取指定模型的当前有效定价
func (s *DatabaseModelPricingService) GetCurrentModelPricing(modelName string) (ModelPricing, bool, error) {
	if s.GetPricingFunc == nil {
		return ModelPricing{}, false, errors.New("数据库定价服务未正确初始化")
	}
	return s.GetPricingFunc(modelName)
}
