package controller

import (
	"claude-code-relay/constant"
	"claude-code-relay/model"
	"claude-code-relay/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateGroup 创建分组
func CreateGroup(c *gin.Context) {
	var req model.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
			"code":  constant.InvalidParams,
		})
		return
	}

	// 从认证中获取用户ID
	user := c.MustGet("user").(*model.User)
	userID := user.ID

	group, err := service.CreateGroup(&req, userID)
	if err != nil {
		var statusCode int
		var code int
		if err.Error() == "组名已存在" || err.Error() == "组名不能为空" {
			statusCode = http.StatusBadRequest
			code = constant.InvalidParams
		} else {
			statusCode = http.StatusInternalServerError
			code = constant.InternalServerError
		}
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "创建分组成功",
		"code":    constant.Success,
		"data":    group,
	})
}

// GetGroup 获取分组详情
func GetGroup(c *gin.Context) {
	id := c.Param("id")

	// 从认证中获取用户ID
	user := c.MustGet("user").(*model.User)
	userID := user.ID

	group, err := service.GetGroup(id, userID)
	if err != nil {
		var statusCode int
		var code int
		if err.Error() == "组不存在" || err.Error() == "无效的组ID" {
			statusCode = http.StatusNotFound
			code = constant.InvalidParams
		} else {
			statusCode = http.StatusInternalServerError
			code = constant.InternalServerError
		}
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取分组成功",
		"code":    constant.Success,
		"data":    group,
	})
}

// UpdateGroup 更新分组
func UpdateGroup(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
			"code":  constant.InvalidParams,
		})
		return
	}

	// 从认证中获取用户ID
	user := c.MustGet("user").(*model.User)
	userID := user.ID

	group, err := service.UpdateGroup(id, &req, userID)
	if err != nil {
		var statusCode int
		var code int
		switch err.Error() {
		case "组名已存在", "组不存在", "无效的组ID":
			statusCode = http.StatusBadRequest
			code = constant.InvalidParams
		default:
			statusCode = http.StatusInternalServerError
			code = constant.InternalServerError
		}
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新分组成功",
		"code":    constant.Success,
		"data":    group,
	})
}

// DeleteGroup 删除分组
func DeleteGroup(c *gin.Context) {
	id := c.Param("id")

	// 从认证中获取用户ID
	user := c.MustGet("user").(*model.User)
	userID := user.ID

	err := service.DeleteGroup(id, userID)
	if err != nil {
		var statusCode int
		var code int
		switch err.Error() {
		case "组不存在", "无效的组ID":
			statusCode = http.StatusBadRequest
			code = constant.InvalidParams
		default:
			statusCode = http.StatusInternalServerError
			code = constant.InternalServerError
		}
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除分组成功",
		"code":    constant.Success,
	})
}

// GetAllGroups 获取所有分组（用于下拉选择）
func GetAllGroups(c *gin.Context) {
	// 从认证中获取用户ID
	user := c.MustGet("user").(*model.User)
	userID := user.ID

	groups, err := service.GetAllGroups(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取分组选项成功",
		"code":    constant.Success,
		"data":    groups,
	})
}

// GetGroups 获取分组列表
func GetGroups(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 从认证中获取用户ID
	user := c.MustGet("user").(*model.User)
	userID := user.ID

	result, err := service.GetGroupList(page, limit, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取分组列表成功",
		"code":    constant.Success,
		"data":    result,
	})
}

// AdminGetGroups 管理员获取所有用户的分组列表
func AdminGetGroups(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 可选的用户ID过滤
	var userID *uint
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userIDInt, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			userIDValue := uint(userIDInt)
			userID = &userIDValue
		}
	}

	result, err := service.AdminGetGroupList(page, limit, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取分组列表成功",
		"code":    constant.Success,
		"data":    result,
	})
}

// AdminGetAllGroups 管理员获取所有用户的分组（用于下拉选择）
func AdminGetAllGroups(c *gin.Context) {
	// 可选的用户ID过滤
	var userID *uint
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userIDInt, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			userIDValue := uint(userIDInt)
			userID = &userIDValue
		}
	}

	groups, err := service.AdminGetAllGroups(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取分组选项成功",
		"code":    constant.Success,
		"data":    groups,
	})
}
