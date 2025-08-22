package common

import (
	"fmt"
	"sync"
	"time"
)

// ModelPricing Claude模型价格配置 (USD per 1M tokens)
type ModelPricing struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheWrite float64 `json:"cache_write"`
	CacheRead  float64 `json:"cache_read"`
}

// CostDetails 费用详情
type CostDetails struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheWrite float64 `json:"cache_write"`
	CacheRead  float64 `json:"cache_read"`
	Total      float64 `json:"total"`
}

// FormattedCosts 格式化的费用字符串
type FormattedCosts struct {
	Input      string `json:"input"`
	Output     string `json:"output"`
	CacheWrite string `json:"cache_write"`
	CacheRead  string `json:"cache_read"`
	Total      string `json:"total"`
}

// UsageDetails 详细的token使用量数据（包含总计）
type UsageDetails struct {
	InputTokens       int `json:"input_tokens"`
	OutputTokens      int `json:"output_tokens"`
	CacheCreateTokens int `json:"cache_creation_input_tokens"`
	CacheReadTokens   int `json:"cache_read_input_tokens"`
	TotalTokens       int `json:"total_tokens"`
}

// CostCalculationResult 费用计算结果
type CostCalculationResult struct {
	Model     string         `json:"model"`
	Pricing   ModelPricing   `json:"pricing"`
	Usage     UsageDetails   `json:"usage"`
	Costs     CostDetails    `json:"costs"`
	Formatted FormattedCosts `json:"formatted"`
}

// SavingsResult 缓存节省信息
type SavingsResult struct {
	NormalCost        float64 `json:"normal_cost"`
	CacheCost         float64 `json:"cache_cost"`
	Savings           float64 `json:"savings"`
	SavingsPercentage float64 `json:"savings_percentage"`
	Formatted         struct {
		NormalCost        string `json:"normal_cost"`
		CacheCost         string `json:"cache_cost"`
		Savings           string `json:"savings"`
		SavingsPercentage string `json:"savings_percentage"`
	} `json:"formatted"`
}

// 删除硬编码的模型定价配置，所有定价都从数据库获取

// CostCalculator 费用计算器
type CostCalculator struct {
	pricingCache map[string]ModelPricing
	cacheMutex   sync.RWMutex
	cacheTime    time.Time
	cacheExpiry  time.Duration
}

// NewCostCalculator 创建费用计算器实例
func NewCostCalculator() *CostCalculator {
	return &CostCalculator{
		pricingCache: make(map[string]ModelPricing),
		cacheExpiry:  5 * time.Minute, // 缓存5分钟
	}
}

// 已删除getModelPricingFromDB方法，所有定价都通过ModelPricingService接口获取

// refreshPricingCache 刷新定价缓存
func (c *CostCalculator) refreshPricingCache() error {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	// 检查缓存是否过期
	if time.Since(c.cacheTime) < c.cacheExpiry && len(c.pricingCache) > 0 {
		return nil // 缓存未过期
	}

	// 清空现有缓存
	c.pricingCache = make(map[string]ModelPricing)

	// 从数据库服务获取定价
	service := GetModelPricingService()
	if service == nil {
		SysLog("Error: Model pricing service is not initialized")
		return fmt.Errorf("model pricing service is not initialized")
	}

	pricingMap, err := service.GetAllCurrentModelPricing()
	if err != nil {
		SysLog("Error: Failed to get pricing from database service: " + err.Error())
		return fmt.Errorf("failed to get pricing from database: %v", err)
	}

	// 设置从数据库获取的定价到缓存
	c.pricingCache = pricingMap
	c.cacheTime = time.Now()

	SysLog(fmt.Sprintf("Pricing cache refreshed with %d models", len(pricingMap)))
	return nil
}

// getPricing 获取模型定价（从数据库缓存获取）
func (c *CostCalculator) getPricing(modelName string) ModelPricing {
	if modelName == "" {
		modelName = "unknown"
	}

	// 刷新缓存
	err := c.refreshPricingCache()
	if err != nil {
		SysLog("Warning: Failed to refresh pricing cache: " + err.Error())
		// 如果刷新失败，尝试使用现有缓存
	}

	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	// 从缓存获取指定模型定价
	if pricing, exists := c.pricingCache[modelName]; exists {
		return pricing
	}

	// 如果指定模型不存在，尝试获取默认模型定价
	if pricing, exists := c.pricingCache["unknown"]; exists {
		return pricing
	}

	// 如果数据库中没有任何定价配置，返回空定价（会导致0费用）
	SysLog(fmt.Sprintf("Warning: No pricing found for model '%s' and no default pricing available", modelName))
	return ModelPricing{
		Input:      0.0,
		Output:     0.0,
		CacheWrite: 0.0,
		CacheRead:  0.0,
	}
}

// CalculateCost 计算单次请求的费用
func (c *CostCalculator) CalculateCost(usage *TokenUsage) *CostCalculationResult {
	model := usage.Model
	if model == "" {
		model = "unknown"
	}

	// 获取定价信息（优先数据库，回退到硬编码）
	pricing := c.getPricing(model)

	// 计算各类型token的费用 (USD)
	inputCost := (float64(usage.InputTokens) / 1000000) * pricing.Input
	outputCost := (float64(usage.OutputTokens) / 1000000) * pricing.Output
	cacheWriteCost := (float64(usage.CacheCreationInputTokens) / 1000000) * pricing.CacheWrite
	cacheReadCost := (float64(usage.CacheReadInputTokens) / 1000000) * pricing.CacheRead

	totalCost := inputCost + outputCost + cacheWriteCost + cacheReadCost

	return &CostCalculationResult{
		Model:   model,
		Pricing: pricing,
		Usage: UsageDetails{
			InputTokens:       usage.InputTokens,
			OutputTokens:      usage.OutputTokens,
			CacheCreateTokens: usage.CacheCreationInputTokens,
			CacheReadTokens:   usage.CacheReadInputTokens,
			TotalTokens:       usage.InputTokens + usage.OutputTokens + usage.CacheCreationInputTokens + usage.CacheReadInputTokens,
		},
		Costs: CostDetails{
			Input:      inputCost,
			Output:     outputCost,
			CacheWrite: cacheWriteCost,
			CacheRead:  cacheReadCost,
			Total:      totalCost,
		},
		Formatted: FormattedCosts{
			Input:      c.FormatCost(inputCost),
			Output:     c.FormatCost(outputCost),
			CacheWrite: c.FormatCost(cacheWriteCost),
			CacheRead:  c.FormatCost(cacheReadCost),
			Total:      c.FormatCost(totalCost),
		},
	}
}

// CalculateAggregatedCost 计算聚合使用量的费用
func (c *CostCalculator) CalculateAggregatedCost(inputTokens, outputTokens, cacheCreateTokens, cacheReadTokens int, model string) *CostCalculationResult {
	usage := &TokenUsage{
		InputTokens:              inputTokens,
		OutputTokens:             outputTokens,
		CacheCreationInputTokens: cacheCreateTokens,
		CacheReadInputTokens:     cacheReadTokens,
		Model:                    model,
	}

	return c.CalculateCost(usage)
}

// GetModelPricing 获取模型定价信息
func (c *CostCalculator) GetModelPricing(model string) ModelPricing {
	return c.getPricing(model)
}

// GetAllModelPricing 获取所有支持的模型和定价
func (c *CostCalculator) GetAllModelPricing() map[string]ModelPricing {
	// 刷新缓存
	c.refreshPricingCache()

	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	result := make(map[string]ModelPricing)
	for k, v := range c.pricingCache {
		result[k] = v
	}
	return result
}

// IsModelSupported 验证模型是否支持
func (c *CostCalculator) IsModelSupported(model string) bool {
	// 刷新缓存
	c.refreshPricingCache()

	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	_, exists := c.pricingCache[model]
	return exists
}

// FormatCost 格式化费用显示
func (c *CostCalculator) FormatCost(cost float64) string {
	if cost >= 1 {
		return fmt.Sprintf("$%.2f", cost)
	} else if cost >= 0.001 {
		return fmt.Sprintf("$%.4f", cost)
	} else {
		return fmt.Sprintf("$%.6f", cost)
	}
}

// CalculateCacheSavings 计算费用节省（使用缓存的节省）
func (c *CostCalculator) CalculateCacheSavings(usage *TokenUsage) *SavingsResult {
	pricing := c.GetModelPricing(usage.Model)
	cacheReadTokens := usage.CacheReadInputTokens

	// 如果这些token不使用缓存，需要按正常input价格计费
	normalCost := (float64(cacheReadTokens) / 1000000) * pricing.Input
	cacheCost := (float64(cacheReadTokens) / 1000000) * pricing.CacheRead
	savings := normalCost - cacheCost
	savingsPercentage := 0.0
	if normalCost > 0 {
		savingsPercentage = (savings / normalCost) * 100
	}

	result := &SavingsResult{
		NormalCost:        normalCost,
		CacheCost:         cacheCost,
		Savings:           savings,
		SavingsPercentage: savingsPercentage,
	}

	result.Formatted.NormalCost = c.FormatCost(normalCost)
	result.Formatted.CacheCost = c.FormatCost(cacheCost)
	result.Formatted.Savings = c.FormatCost(savings)
	result.Formatted.SavingsPercentage = fmt.Sprintf("%.1f%%", savingsPercentage)

	return result
}

// 全局费用计算器实例
var GlobalCostCalculator = NewCostCalculator()

// 便利函数，使用全局实例
func CalculateCost(usage *TokenUsage) *CostCalculationResult {
	return GlobalCostCalculator.CalculateCost(usage)
}

func CalculateAggregatedCost(inputTokens, outputTokens, cacheCreateTokens, cacheReadTokens int, model string) *CostCalculationResult {
	return GlobalCostCalculator.CalculateAggregatedCost(inputTokens, outputTokens, cacheCreateTokens, cacheReadTokens, model)
}

func GetModelPricing(model string) ModelPricing {
	return GlobalCostCalculator.GetModelPricing(model)
}

func GetAllModelPricing() map[string]ModelPricing {
	return GlobalCostCalculator.GetAllModelPricing()
}

func IsModelSupported(model string) bool {
	return GlobalCostCalculator.IsModelSupported(model)
}

func FormatCost(cost float64) string {
	return GlobalCostCalculator.FormatCost(cost)
}

func CalculateCacheSavings(usage *TokenUsage) *SavingsResult {
	return GlobalCostCalculator.CalculateCacheSavings(usage)
}
