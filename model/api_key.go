package model

import (
	"crypto/rand"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type ApiKey struct {
	ID                            uint           `json:"id" gorm:"primaryKey"`
	Name                          string         `json:"name" gorm:"type:varchar(100);not null"`
	Key                           string         `json:"key" gorm:"type:varchar(100);uniqueIndex;not null"`
	ExpiresAt                     *Time          `json:"expires_at" gorm:"type:datetime"`
	Status                        int            `json:"status" gorm:"default:1"` // 1:启用 0:禁用
	GroupID                       int            `json:"group_id" gorm:"default:0;index"`
	UserID                        uint           `json:"user_id" gorm:"not null;index"`
	TodayUsageCount               int            `json:"today_usage_count" gorm:"default:0;comment:今日使用次数"`
	TodayInputTokens              int            `json:"today_input_tokens" gorm:"default:0;comment:今日输入tokens"`
	TodayOutputTokens             int            `json:"today_output_tokens" gorm:"default:0;comment:今日输出tokens"`
	TodayCacheReadInputTokens     int            `json:"today_cache_read_input_tokens" gorm:"default:0;comment:今日缓存读取输入tokens"`
	TodayCacheCreationInputTokens int            `json:"today_cache_creation_input_tokens" gorm:"default:0;comment:今日缓存创建输入tokens"`
	TodayTotalCost                float64        `json:"today_total_cost" gorm:"default:0;comment:今日使用总费用(USD)"`
	ModelRestriction              string         `json:"model_restriction" gorm:"type:text;comment:模型限制,逗号分隔"`
	DailyLimit                    float64        `json:"daily_limit" gorm:"default:0;comment:日限额(美元),0表示不限制"`
	LastUsedTime                  *Time          `json:"last_used_time" gorm:"comment:最后使用时间;type:datetime"`
	CreatedAt                     Time           `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt                     Time           `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	DeletedAt                     gorm.DeletedAt `json:"-" gorm:"index"`
	// 关联查询
	Group *Group `json:"group" gorm:"-"`
	User  *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type CreateApiKeyRequest struct {
	Name             string  `json:"name" binding:"required"`
	Key              string  `json:"key"`
	ExpiresAt        *Time   `json:"expires_at"`
	Status           int     `json:"status" binding:"oneof=1 2"`
	GroupID          int     `json:"group_id"`
	ModelRestriction string  `json:"model_restriction"`
	DailyLimit       float64 `json:"daily_limit"`
}

type UpdateApiKeyRequest struct {
	Name             string   `json:"name"`
	ExpiresAt        *Time    `json:"expires_at"`
	Status           *int     `json:"status"`
	GroupID          *int     `json:"group_id"`
	ModelRestriction *string  `json:"model_restriction"`
	DailyLimit       *float64 `json:"daily_limit"`
}

type ApiKeyListResult struct {
	ApiKeys []ApiKey `json:"api_keys"`
	Total   int64    `json:"total"`
	Page    int      `json:"page"`
	Limit   int      `json:"limit"`
}

func (a *ApiKey) TableName() string {
	return "api_keys"
}

func (a *ApiKey) BeforeCreate(tx *gorm.DB) error {
	if a.Key == "" {
		key, err := generateApiKey()
		if err != nil {
			return err
		}
		a.Key = key
	}
	return nil
}

func generateApiKey() (string, error) {
	bytes := make([]byte, 30)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("sk-%x", bytes)[:30], nil
}

func CreateApiKey(apiKey *ApiKey) error {
	apiKey.ID = 0
	return DB.Create(apiKey).Error
}

func GetApiKeyById(id uint, userID uint) (*ApiKey, error) {
	var apiKey ApiKey
	err := DB.Where("id = ? AND user_id = ?", id, userID).First(&apiKey).Error
	if err != nil {
		return nil, err
	}

	// 如果有分组ID，查询分组信息
	if apiKey.GroupID > 0 {
		var group Group
		if err := DB.Where("id = ? AND user_id = ?", apiKey.GroupID, userID).First(&group).Error; err == nil {
			apiKey.Group = &group
		}
	}

	return &apiKey, nil
}

// GetApiKeyByKey 根据API Key获取
func GetApiKeyByKey(key string) (*ApiKey, error) {
	var apiKey ApiKey
	err := DB.Where("`key` = ? AND status = 1", key).First(&apiKey).Error
	if err != nil {
		return nil, err
	}

	// 检查是否过期
	if apiKey.ExpiresAt != nil && time.Time(*apiKey.ExpiresAt).Before(time.Now()) {
		return nil, gorm.ErrRecordNotFound
	}

	return &apiKey, nil
}

func UpdateApiKey(apiKey *ApiKey) error {
	return DB.Save(apiKey).Error
}

func DeleteApiKey(id uint) error {
	return DB.Delete(&ApiKey{}, id).Error
}

// GetApiKeys 分页获取API Keys
func GetApiKeys(page, limit int, userID uint, groupID *uint) ([]ApiKey, int64, error) {
	var apiKeys []ApiKey
	var total int64

	query := DB.Model(&ApiKey{}).Where("user_id = ?", userID)
	if groupID != nil {
		query = query.Where("group_id = ?", *groupID)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = query.Offset(offset).Limit(limit).Find(&apiKeys).Error
	if err != nil {
		return nil, 0, err
	}

	// 批量查询分组信息
	groupIDs := make(map[int]bool)
	for _, apiKey := range apiKeys {
		if apiKey.GroupID > 0 {
			groupIDs[apiKey.GroupID] = true
		}
	}

	if len(groupIDs) > 0 {
		var ids []int
		for id := range groupIDs {
			ids = append(ids, id)
		}

		var groups []Group
		DB.Where("id IN ? AND user_id = ?", ids, userID).Find(&groups)

		// 创建分组映射
		groupMap := make(map[int]*Group)
		for i := range groups {
			groupMap[int(groups[i].ID)] = &groups[i]
		}

		// 为每个API Key设置对应的分组信息
		for i := range apiKeys {
			if apiKeys[i].GroupID > 0 {
				if group, exists := groupMap[apiKeys[i].GroupID]; exists {
					apiKeys[i].Group = group
				}
			}
		}
	}

	return apiKeys, total, nil
}

// AdminGetApiKeys 管理员分页获取所有用户的API Keys
func AdminGetApiKeys(page, limit int, userID *uint, groupID *uint) ([]ApiKey, int64, error) {
	var apiKeys []ApiKey
	var total int64

	query := DB.Model(&ApiKey{})

	// 如果指定了用户ID，则只查询该用户的API Keys
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	// 如果指定了分组ID，则只查询该分组的API Keys
	if groupID != nil {
		query = query.Where("group_id = ?", *groupID)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = query.Preload("User").Offset(offset).Limit(limit).Find(&apiKeys).Error
	if err != nil {
		return nil, 0, err
	}

	// 批量查询分组信息
	groupIDs := make(map[int]bool)
	userIDs := make(map[uint]bool)
	for _, apiKey := range apiKeys {
		if apiKey.GroupID > 0 {
			groupIDs[apiKey.GroupID] = true
		}
		userIDs[apiKey.UserID] = true
	}

	if len(groupIDs) > 0 {
		var ids []int
		for id := range groupIDs {
			ids = append(ids, id)
		}

		var groups []Group
		DB.Where("id IN ?", ids).Find(&groups)

		// 创建分组映射
		groupMap := make(map[int]*Group)
		for i := range groups {
			groupMap[int(groups[i].ID)] = &groups[i]
		}

		// 为每个API Key设置对应的分组信息
		for i := range apiKeys {
			if apiKeys[i].GroupID > 0 {
				if group, exists := groupMap[apiKeys[i].GroupID]; exists {
					apiKeys[i].Group = group
				}
			}
		}
	}

	return apiKeys, total, nil
}
