package service

import (
	"claude-code-relay/common"
	"claude-code-relay/model"
	"errors"
	"log"
	"time"

	"gorm.io/gorm"
)

func CreateApiKey(userID uint, req *model.CreateApiKeyRequest) (*model.ApiKey, error) {
	if req.Name == "" {
		return nil, errors.New("API Key名称不能为空")
	}

	// 如果指定了分组ID，验证分组是否存在且属于用户
	if req.GroupID > 0 {
		_, err := model.GetGroupById(req.GroupID, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("指定的分组不存在")
			}
			return nil, err
		}
	}

	// 如果指定了过期时间，验证时间有效性
	if req.ExpiresAt != nil && time.Time(*req.ExpiresAt).Before(time.Now()) {
		return nil, errors.New("过期时间不能早于当前时间")
	}

	apiKey := &model.ApiKey{
		Name:      req.Name,
		Key:       req.Key,
		ExpiresAt: req.ExpiresAt,
		Status:    req.Status,
		GroupID:   req.GroupID,
		UserID:    userID,
	}

	if apiKey.Status == 0 {
		apiKey.Status = 1 // 默认启用
	}

	err := model.CreateApiKey(apiKey)
	if err != nil {
		return nil, err
	}

	return apiKey, nil
}

func GetApiKeyById(id, userID uint) (*model.ApiKey, error) {
	return model.GetApiKeyById(id, userID)
}

func UpdateApiKey(id, userID uint, req *model.UpdateApiKeyRequest) (*model.ApiKey, error) {
	apiKey, err := model.GetApiKeyById(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("API Key不存在")
		}
		return nil, err
	}

	// 如果指定了分组ID，验证分组是否存在且属于用户
	if req.GroupID != nil && *req.GroupID != 0 {
		_, err := model.GetGroupById(*req.GroupID, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("指定的分组不存在")
			}
			return nil, err
		}
	}

	// 如果指定了过期时间，验证时间有效性
	if req.ExpiresAt != nil && time.Time(*req.ExpiresAt).Before(time.Now()) {
		return nil, errors.New("过期时间不能早于当前时间")
	}

	// 更新字段
	if req.Name != "" {
		apiKey.Name = req.Name
	}
	if req.ExpiresAt != nil {
		apiKey.ExpiresAt = req.ExpiresAt
	}
	if req.Status != nil {
		apiKey.Status = *req.Status
	}
	if req.GroupID != nil {
		apiKey.GroupID = *req.GroupID
	}
	if req.ModelRestriction != nil {
		apiKey.ModelRestriction = *req.ModelRestriction
	}
	if req.DailyLimit != nil {
		apiKey.DailyLimit = *req.DailyLimit
	}

	err = model.UpdateApiKey(apiKey)
	if err != nil {
		return nil, err
	}

	return apiKey, nil
}

func DeleteApiKey(id, userID uint) error {
	apiKey, err := model.GetApiKeyById(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("API Key不存在")
		}
		return err
	}

	return model.DeleteApiKey(apiKey.ID)
}

func GetApiKeys(page, limit int, userID uint, groupID *uint) (*model.ApiKeyListResult, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	apiKeys, total, err := model.GetApiKeys(page, limit, userID, groupID)
	if err != nil {
		return nil, err
	}

	return &model.ApiKeyListResult{
		ApiKeys: apiKeys,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}, nil
}

// UpdateApiKeyStatusCom 更新API Key状态
func UpdateApiKeyStatusCom(id, userID uint, status int) error {
	apiKey, err := model.GetApiKeyById(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("API Key不存在")
		}
		return err
	}

	apiKey.Status = status

	err = model.UpdateApiKey(apiKey)
	if err != nil {
		return err
	}

	return nil
}

// AdminGetApiKeys 管理员获取所有用户的API Key列表
func AdminGetApiKeys(page, limit int, userID *uint, groupID *uint) (*model.ApiKeyListResult, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	apiKeys, total, err := model.AdminGetApiKeys(page, limit, userID, groupID)
	if err != nil {
		return nil, err
	}

	return &model.ApiKeyListResult{
		ApiKeys: apiKeys,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}, nil
}

// ValidateApiKey 验证API Key是否有效
func ValidateApiKey(key string) (*model.ApiKey, error) {
	if key == "" {
		return nil, errors.New("API Key不能为空")
	}

	apiKey, err := model.GetApiKeyByKey(key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("API Key无效或已过期")
		}
		return nil, err
	}

	return apiKey, nil
}

// UpdateApiKeyStatus 根据响应状态码更新API Key统计信息
func UpdateApiKeyStatus(apiKey *model.ApiKey, statusCode int, usage *common.TokenUsage) {
	// 只在请求成功时更新API Key统计信息
	if statusCode != 200 && statusCode != 201 {
		return
	}

	now := time.Now()

	// 判断最后使用时间是否为当天
	if apiKey.LastUsedTime != nil {
		lastUsedDate := time.Time(*apiKey.LastUsedTime).Format("2006-01-02")
		todayDate := now.Format("2006-01-02")

		if lastUsedDate == todayDate {
			// 同一天，使用次数+1
			apiKey.TodayUsageCount++
		} else {
			// 不同天，重置为1
			apiKey.TodayUsageCount = 1
		}
	} else {
		// 首次使用，设置为1
		apiKey.TodayUsageCount = 1
	}

	// 更新token使用量和费用（如果有的话）
	if usage != nil {
		// 计算本次请求的费用
		costResult := common.CalculateCost(usage)
		currentCost := costResult.Costs.Total

		if apiKey.LastUsedTime != nil {
			lastUsedDate := time.Time(*apiKey.LastUsedTime).Format("2006-01-02")
			todayDate := now.Format("2006-01-02")

			if lastUsedDate == todayDate {
				// 同一天，累加各类tokens和费用
				apiKey.TodayInputTokens += usage.InputTokens
				apiKey.TodayOutputTokens += usage.OutputTokens
				apiKey.TodayCacheReadInputTokens += usage.CacheReadInputTokens
				apiKey.TodayCacheCreationInputTokens += usage.CacheCreationInputTokens
				apiKey.TodayTotalCost += currentCost
			} else {
				// 不同天，重置各类tokens和费用
				apiKey.TodayInputTokens = usage.InputTokens
				apiKey.TodayOutputTokens = usage.OutputTokens
				apiKey.TodayCacheReadInputTokens = usage.CacheReadInputTokens
				apiKey.TodayCacheCreationInputTokens = usage.CacheCreationInputTokens
				apiKey.TodayTotalCost = currentCost
			}
		} else {
			// 首次使用，设置各类tokens和费用
			apiKey.TodayInputTokens = usage.InputTokens
			apiKey.TodayOutputTokens = usage.OutputTokens
			apiKey.TodayCacheReadInputTokens = usage.CacheReadInputTokens
			apiKey.TodayCacheCreationInputTokens = usage.CacheCreationInputTokens
			apiKey.TodayTotalCost = currentCost
		}
	}

	// 更新最后使用时间
	nowTime := model.Time(now)
	apiKey.LastUsedTime = &nowTime

	// 更新数据库
	if err := model.UpdateApiKey(apiKey); err != nil {
		log.Printf("failed to update api key status: %v", err)
	}
}
