import { request } from '@/utils/request';

// 模型配置接口类型定义
export interface ModelConfig {
  id: number;
  name: string;
  display_name: string;
  provider: string;
  category: string;
  version: string;
  status: number;
  sort_order: number;
  description: string;
  max_tokens?: number;
  context_window?: number;
  created_at: string;
  updated_at: string;
  pricing?: ModelPricing[];
}

export interface ModelPricing {
  id: number;
  model_id: number;
  input_price: number;
  output_price: number;
  cache_write_price: number;
  cache_read_price: number;
  effective_time: string;
  expire_time?: string;
  status: number;
  created_at: string;
  updated_at: string;
}

export interface ModelQueryParams {
  page?: number;
  limit?: number;
  name?: string;
  provider?: string;
  status?: number;
}

export interface CreateModelRequest {
  name: string;
  display_name: string;
  provider: string;
  category: string;
  version: string;
  status: number;
  sort_order: number;
  description?: string;
  max_tokens?: number;
  context_window?: number;
}

export interface UpdateModelRequest {
  display_name: string;
  provider: string;
  category: string;
  version: string;
  status: number;
  sort_order: number;
  description?: string;
  max_tokens?: number;
  context_window?: number;
}

export interface CreatePricingRequest {
  input_price: number;
  output_price: number;
  cache_write_price: number;
  cache_read_price: number;
  effective_time: string;
  expire_time?: string;
}

export interface ModelListResponse {
  models: ModelConfig[];
  total: number;
}

// 模型管理API接口

/**
 * 创建模型配置
 */
export const createModel = (data: CreateModelRequest) => {
  return request.post({
    url: '/api/v1/admin/models',
    data,
  });
};

/**
 * 获取模型配置列表
 */
export const getModelList = (params?: ModelQueryParams) => {
  return request.get<ModelListResponse>({
    url: '/api/v1/admin/models',
    params,
  });
};

/**
 * 获取模型配置详情
 */
export const getModel = (id: number) => {
  return request.get<ModelConfig>({
    url: `/api/v1/admin/models/${id}`,
  });
};

/**
 * 更新模型配置
 */
export const updateModel = (id: number, data: UpdateModelRequest) => {
  return request.put({
    url: `/api/v1/admin/models/${id}`,
    data,
  });
};

/**
 * 删除模型配置
 */
export const deleteModel = (id: number) => {
  return request.delete({
    url: `/api/v1/admin/models/${id}`,
  });
};

/**
 * 获取启用的模型列表
 */
export const getActiveModels = () => {
  return request.get<ModelConfig[]>({
    url: '/api/v1/models/active',
  });
};

/**
 * 批量更新模型状态
 */
export const updateModelStatus = (data: { ids: number[]; status: number }) => {
  return request.put({
    url: '/api/v1/admin/models/status',
    data,
  });
};

/**
 * 验证模型名称
 */
export const validateModelName = (name: string) => {
  return request.get({
    url: '/api/v1/admin/models/validate-name',
    params: { name },
  });
};

/**
 * 刷新定价缓存
 */
export const refreshPricingCache = () => {
  return request.post({
    url: '/api/v1/admin/models/refresh-cache',
  });
};

// 模型定价管理API接口

/**
 * 创建模型定价
 */
export const createModelPricing = (modelId: number, data: CreatePricingRequest) => {
  return request.post({
    url: `/api/v1/admin/models/${modelId}/pricing`,
    data,
  });
};

/**
 * 获取模型定价历史
 */
export const getModelPricingHistory = (modelId: number) => {
  return request.get<ModelPricing[]>({
    url: `/api/v1/admin/models/${modelId}/pricing`,
  });
};

/**
 * 更新模型定价
 */
export const updateModelPricing = (pricingId: number, data: CreatePricingRequest) => {
  return request.put({
    url: `/api/v1/admin/models/pricing/${pricingId}`,
    data,
  });
};

/**
 * 删除模型定价
 */
export const deleteModelPricing = (pricingId: number) => {
  return request.delete({
    url: `/api/v1/admin/models/pricing/${pricingId}`,
  });
};

/**
 * 获取当前有效定价
 */
export const getCurrentModelPricing = () => {
  return request.get<Record<string, ModelPricing>>({
    url: '/api/v1/models/current-pricing',
  });
};
