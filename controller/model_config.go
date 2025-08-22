package controller

import (
	"claude-code-relay/model"
	"claude-code-relay/service"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var modelConfigService = service.NewModelConfigService()

// CreateModel 创建模型配置
// @Tags Models
// @Summary 创建模型配置
// @Description 创建新的模型配置
// @Accept json
// @Produce json
// @Param input body model.CreateModelRequest true "模型配置信息"
// @Success 200 {object} model.ModelConfig
// @Failure 400 {object} common.APIResponse
// @Failure 500 {object} common.APIResponse
// @Router /api/admin/models [post]
func CreateModel(c *gin.Context) {
	var req model.CreateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	modelConfig, err := modelConfigService.CreateModel(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    modelConfig,
		"message": "模型配置创建成功",
	})
}

// GetModel 获取模型配置详情
// @Tags Models
// @Summary 获取模型配置详情
// @Description 根据ID获取模型配置的详细信息
// @Accept json
// @Produce json
// @Param id path int true "模型ID"
// @Success 200 {object} model.ModelConfig
// @Failure 400 {object} common.APIResponse
// @Failure 404 {object} common.APIResponse
// @Router /api/admin/models/{id} [get]
func GetModel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的模型ID",
		})
		return
	}

	modelConfig, err := modelConfigService.GetModelByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "模型配置不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    modelConfig,
	})
}

// UpdateModel 更新模型配置
// @Tags Models
// @Summary 更新模型配置
// @Description 更新指定的模型配置信息
// @Accept json
// @Produce json
// @Param id path int true "模型ID"
// @Param input body model.UpdateModelRequest true "更新的模型信息"
// @Success 200 {object} model.ModelConfig
// @Failure 400 {object} common.APIResponse
// @Failure 404 {object} common.APIResponse
// @Router /api/admin/models/{id} [put]
func UpdateModel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的模型ID",
		})
		return
	}

	var req model.UpdateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	modelConfig, err := modelConfigService.UpdateModel(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    modelConfig,
		"message": "模型配置更新成功",
	})
}

// DeleteModel 删除模型配置
// @Tags Models
// @Summary 删除模型配置
// @Description 删除指定的模型配置
// @Accept json
// @Produce json
// @Param id path int true "模型ID"
// @Success 200 {object} common.APIResponse
// @Failure 400 {object} common.APIResponse
// @Failure 404 {object} common.APIResponse
// @Router /api/admin/models/{id} [delete]
func DeleteModel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的模型ID",
		})
		return
	}

	err = modelConfigService.DeleteModel(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "模型配置删除成功",
	})
}

// GetModelList 获取模型配置列表
// @Tags Models
// @Summary 获取模型配置列表
// @Description 分页获取模型配置列表
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param name query string false "模型名称（支持模糊搜索）"
// @Param provider query string false "提供商"
// @Param status query int false "状态（0:禁用 1:启用）"
// @Success 200 {object} model.ModelListResponse
// @Failure 400 {object} common.APIResponse
// @Router /api/admin/models [get]
func GetModelList(c *gin.Context) {
	var params model.ModelQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	result, err := modelConfigService.GetModelList(&params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result.Models,
		"total":   result.Total,
	})
}

// GetActiveModels 获取所有启用的模型
// @Tags Models
// @Summary 获取启用的模型列表
// @Description 获取所有状态为启用的模型配置
// @Accept json
// @Produce json
// @Success 200 {object} []model.ModelConfig
// @Failure 500 {object} common.APIResponse
// @Router /api/models/active [get]
func GetActiveModels(c *gin.Context) {
	models, err := modelConfigService.GetActiveModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    models,
	})
}

// UpdateModelStatus 更新模型状态
// @Tags Models
// @Summary 更新模型状态
// @Description 批量更新模型的启用/禁用状态
// @Accept json
// @Produce json
// @Param input body object true "更新状态请求" example({"ids":[1,2,3],"status":1})
// @Success 200 {object} common.APIResponse
// @Failure 400 {object} common.APIResponse
// @Router /api/admin/models/status [put]
func UpdateModelStatus(c *gin.Context) {
	var req struct {
		IDs    []uint `json:"ids" binding:"required,min=1"`
		Status int    `json:"status" binding:"oneof=0 1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	err := modelConfigService.BatchUpdateStatus(req.IDs, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	statusText := "禁用"
	if req.Status == 1 {
		statusText = "启用"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "批量" + statusText + "成功",
	})
}

// CreateModelPricing 创建模型定价
// @Tags Model Pricing
// @Summary 创建模型定价
// @Description 为指定模型创建新的定价配置
// @Accept json
// @Produce json
// @Param id path int true "模型ID"
// @Param input body model.CreatePricingRequest true "定价信息"
// @Success 200 {object} model.ModelPricing
// @Failure 400 {object} common.APIResponse
// @Router /api/admin/models/{id}/pricing [post]
func CreateModelPricing(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的模型ID",
		})
		return
	}

	var req model.CreatePricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	pricing, err := modelConfigService.CreatePricing(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    pricing,
		"message": "定价配置创建成功",
	})
}

// GetModelPricingHistory 获取模型定价历史
// @Tags Model Pricing
// @Summary 获取模型定价历史
// @Description 获取指定模型的所有定价历史记录
// @Accept json
// @Produce json
// @Param id path int true "模型ID"
// @Success 200 {object} []model.ModelPricing
// @Failure 400 {object} common.APIResponse
// @Router /api/admin/models/{id}/pricing [get]
func GetModelPricingHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的模型ID",
		})
		return
	}

	pricingHistory, err := modelConfigService.GetPricingHistory(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    pricingHistory,
	})
}

// UpdateModelPricing 更新模型定价
// @Tags Model Pricing
// @Summary 更新模型定价
// @Description 更新指定的定价配置
// @Accept json
// @Produce json
// @Param id path int true "定价ID"
// @Param input body model.CreatePricingRequest true "定价信息"
// @Success 200 {object} model.ModelPricing
// @Failure 400 {object} common.APIResponse
// @Router /api/admin/models/pricing/{id} [put]
func UpdateModelPricing(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的定价ID",
		})
		return
	}

	var req model.CreatePricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	pricing, err := modelConfigService.UpdatePricing(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    pricing,
		"message": "定价配置更新成功",
	})
}

// DeleteModelPricing 删除模型定价
// @Tags Model Pricing
// @Summary 删除模型定价
// @Description 删除指定的定价配置
// @Accept json
// @Produce json
// @Param id path int true "定价ID"
// @Success 200 {object} common.APIResponse
// @Failure 400 {object} common.APIResponse
// @Router /api/admin/models/pricing/{id} [delete]
func DeleteModelPricing(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的定价ID",
		})
		return
	}

	err = modelConfigService.DeletePricing(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "定价配置删除成功",
	})
}

// GetCurrentModelPricing 获取所有模型的当前有效定价
// @Tags Model Pricing
// @Summary 获取当前有效定价
// @Description 获取所有模型的当前有效定价信息
// @Accept json
// @Produce json
// @Success 200 {object} map[string]model.ModelPricing
// @Failure 500 {object} common.APIResponse
// @Router /api/models/current-pricing [get]
func GetCurrentModelPricing(c *gin.Context) {
	pricingMap, err := modelConfigService.GetAllCurrentModelPricing()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    pricingMap,
	})
}

// ValidateModelName 验证模型名称
// @Tags Models
// @Summary 验证模型名称
// @Description 检查模型名称是否可用
// @Accept json
// @Produce json
// @Param name query string true "模型名称"
// @Success 200 {object} common.APIResponse
// @Failure 400 {object} common.APIResponse
// @Router /api/admin/models/validate-name [get]
func ValidateModelName(c *gin.Context) {
	name := strings.TrimSpace(c.Query("name"))
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "模型名称不能为空",
		})
		return
	}

	err := modelConfigService.ValidateModelName(name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "模型名称可用",
	})
}

// RefreshPricingCache 刷新定价缓存
// @Tags Model Pricing
// @Summary 刷新定价缓存
// @Description 手动刷新CostCalculator的定价缓存
// @Accept json
// @Produce json
// @Success 200 {object} common.APIResponse
// @Failure 500 {object} common.APIResponse
// @Router /api/admin/models/refresh-cache [post]
func RefreshPricingCache(c *gin.Context) {
	err := modelConfigService.SyncPricingToCache()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "缓存刷新成功",
	})
}
