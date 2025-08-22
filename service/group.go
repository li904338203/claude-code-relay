package service

import (
	"claude-code-relay/model"
	"errors"
	"strconv"

	"gorm.io/gorm"
)

func CreateGroup(req *model.CreateGroupRequest, userID uint) (*model.Group, error) {
	if req.Name == "" {
		return nil, errors.New("组名不能为空")
	}

	// 检查组名是否已存在（在当前用户下）
	_, err := model.GetGroupByName(req.Name, userID)
	if err == nil {
		return nil, errors.New("组名已存在")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	group := &model.Group{
		Name:   req.Name,
		Remark: req.Remark,
		Status: req.Status,
		UserID: userID,
	}

	// 如果没有指定状态，默认为启用
	if group.Status == 0 && req.Status == 0 {
		group.Status = 1
	}

	err = model.CreateGroup(group)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func GetGroup(id string, userID uint) (*model.Group, error) {
	groupId, err := strconv.Atoi(id)
	if err != nil {
		return nil, errors.New("无效的组ID")
	}

	group, err := model.GetGroupById(groupId, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("组不存在")
		}
		return nil, err
	}

	return group, nil
}

func UpdateGroup(id string, req *model.UpdateGroupRequest, userID uint) (*model.Group, error) {
	group, err := GetGroup(id, userID)
	if err != nil {
		return nil, err
	}

	// 如果要更新组名，检查新组名是否已存在（在当前用户下）
	if req.Name != "" && req.Name != group.Name {
		_, err := model.GetGroupByName(req.Name, userID)
		if err == nil {
			return nil, errors.New("组名已存在")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		group.Name = req.Name
	}

	if req.Remark != "" || req.Remark == "" { // 允许清空备注
		group.Remark = req.Remark
	}

	if req.Status != nil {
		group.Status = *req.Status
	}

	err = model.UpdateGroup(group)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func DeleteGroup(id string, userID uint) error {
	group, err := GetGroup(id, userID)
	if err != nil {
		return err
	}

	return model.DeleteGroup(group.ID)
}

func GetAllGroups(userID uint) ([]model.Group, error) {
	return model.GetAllGroups(userID)
}

func GetGroupList(page, limit int, userID uint) (*model.GroupListResult, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	groups, total, err := model.GetGroups(page, limit, userID)
	if err != nil {
		return nil, err
	}

	return &model.GroupListResult{
		Groups: groups,
		Total:  total,
		Page:   page,
		Limit:  limit,
	}, nil
}

// AdminGetGroupList 管理员获取所有用户的分组列表
func AdminGetGroupList(page, limit int, userID *uint) (*model.GroupListResult, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	groups, total, err := model.AdminGetGroups(page, limit, userID)
	if err != nil {
		return nil, err
	}

	return &model.GroupListResult{
		Groups: groups,
		Total:  total,
		Page:   page,
		Limit:  limit,
	}, nil
}

// AdminGetAllGroups 管理员获取所有用户的分组（用于下拉选择）
func AdminGetAllGroups(userID *uint) ([]model.Group, error) {
	return model.AdminGetAllGroups(userID)
}
