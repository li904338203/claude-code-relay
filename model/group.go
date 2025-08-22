package model

import (
	"gorm.io/gorm"
)

type Group struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"type:varchar(100);not null;uniqueIndex:idx_groups_user_name"`
	Remark    string         `json:"remark" gorm:"type:text"`
	Status    int            `json:"status" gorm:"default:1"` // 1:启用 0:禁用
	UserID    uint           `json:"user_id" gorm:"not null;uniqueIndex:idx_groups_user_name"`
	CreatedAt Time           `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt Time           `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"uniqueIndex:idx_groups_user_name"`

	// 统计字段，不存储在数据库中
	ApiKeyCount  int `json:"api_key_count" gorm:"-"`
	AccountCount int `json:"account_count" gorm:"-"`

	// 关联查询
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type CreateGroupRequest struct {
	Name   string `json:"name" binding:"required"`
	Remark string `json:"remark"`
	Status int    `json:"status"`
}

type UpdateGroupRequest struct {
	Name   string `json:"name"`
	Remark string `json:"remark"`
	Status *int   `json:"status"`
}

type GroupListResult struct {
	Groups []Group `json:"groups"`
	Total  int64   `json:"total"`
	Page   int     `json:"page"`
	Limit  int     `json:"limit"`
}

func (g *Group) TableName() string {
	return "groups"
}

func CreateGroup(group *Group) error {
	group.ID = 0
	return DB.Create(group).Error
}

func GetGroupById(id int, userID uint) (*Group, error) {
	var group Group
	err := DB.Where("id = ? AND user_id = ?", id, userID).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func GetGroupByName(name string, userID uint) (*Group, error) {
	var group Group
	err := DB.Where("name = ? AND user_id = ?", name, userID).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func UpdateGroup(group *Group) error {
	return DB.Save(group).Error
}

func DeleteGroup(id uint) error {
	return DB.Delete(&Group{}, id).Error
}

// GetAllGroups 获取所有分组（不分页）
func GetAllGroups(userID uint) ([]Group, error) {
	var groups []Group

	err := DB.Where("user_id = ? AND status = 1", userID).Find(&groups).Error
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func GetGroups(page, limit int, userID uint) ([]Group, int64, error) {
	var groups []Group
	var total int64

	err := DB.Model(&Group{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = DB.Where("user_id = ?", userID).Offset(offset).Limit(limit).Find(&groups).Error
	if err != nil {
		return nil, 0, err
	}

	// 为每个分组统计API密钥和账号数量
	for i := range groups {
		// 统计API密钥数量
		var apiKeyCount int64
		err = DB.Model(&ApiKey{}).Where("group_id = ? AND user_id = ?", groups[i].ID, userID).Count(&apiKeyCount).Error
		if err != nil {
			return nil, 0, err
		}
		groups[i].ApiKeyCount = int(apiKeyCount)

		// 统计账号数量
		var accountCount int64
		err = DB.Model(&Account{}).Where("group_id = ? AND user_id = ?", groups[i].ID, userID).Count(&accountCount).Error
		if err != nil {
			return nil, 0, err
		}
		groups[i].AccountCount = int(accountCount)
	}

	return groups, total, nil
}

// AdminGetGroups 管理员分页获取所有用户的分组
func AdminGetGroups(page, limit int, userID *uint) ([]Group, int64, error) {
	var groups []Group
	var total int64

	query := DB.Model(&Group{})

	// 如果指定了用户ID，则只查询该用户的分组
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = query.Preload("User").Offset(offset).Limit(limit).Find(&groups).Error
	if err != nil {
		return nil, 0, err
	}

	// 为每个分组统计API密钥和账号数量
	for i := range groups {
		// 统计API密钥数量
		var apiKeyCount int64
		err = DB.Model(&ApiKey{}).Where("group_id = ? AND user_id = ?", groups[i].ID, groups[i].UserID).Count(&apiKeyCount).Error
		if err != nil {
			return nil, 0, err
		}
		groups[i].ApiKeyCount = int(apiKeyCount)

		// 统计账号数量
		var accountCount int64
		err = DB.Model(&Account{}).Where("group_id = ? AND user_id = ?", groups[i].ID, groups[i].UserID).Count(&accountCount).Error
		if err != nil {
			return nil, 0, err
		}
		groups[i].AccountCount = int(accountCount)
	}

	return groups, total, nil
}

// AdminGetAllGroups 管理员获取所有分组（不分页）
func AdminGetAllGroups(userID *uint) ([]Group, error) {
	var groups []Group

	query := DB.Where("status = 1")

	// 如果指定了用户ID，则只查询该用户的分组
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	err := query.Find(&groups).Error
	if err != nil {
		return nil, err
	}

	return groups, nil
}
