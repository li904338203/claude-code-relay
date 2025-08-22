import { request } from '@/utils/request';

// API 路径定义
const Api = {
  // 管理员账号管理
  AdminAccounts: '/api/v1/admin/accounts',

  // 管理员密钥管理
  AdminKeys: '/api/v1/admin/keys',

  // 管理员分组管理
  AdminGroups: '/api/v1/admin/groups',
  AdminGroupsAll: '/api/v1/admin/groups/all',

  // 管理员日志管理
  AdminLogs: '/api/v1/logs',
  AdminLogsStats: '/api/v1/logs/stats',
};

// 管理员账号管理
export interface AdminAccountListParams {
  page?: number;
  limit?: number;
  user_id?: number;
  name?: string;
}

export interface AdminAccountListResponse {
  accounts: any[];
  total: number;
  page: number;
  limit: number;
}

// 管理员API Key管理
export interface AdminApiKeyListParams {
  page?: number;
  limit?: number;
  user_id?: number;
  group_id?: number;
}

export interface AdminApiKeyListResponse {
  api_keys: any[];
  total: number;
  page: number;
  limit: number;
}

/**
 * 管理员获取所有用户的账号列表
 */
export function getAdminAccounts(params?: AdminAccountListParams) {
  return request.get<AdminAccountListResponse>({
    url: Api.AdminAccounts,
    params,
  });
}

/**
 * 管理员获取所有用户的API Key列表
 */
export function getAdminApiKeys(params?: AdminApiKeyListParams) {
  return request.get<AdminApiKeyListResponse>({
    url: Api.AdminKeys,
    params,
  });
}

// 管理员分组管理
export interface AdminGroupListParams {
  page?: number;
  limit?: number;
  user_id?: number;
  name?: string;
}

export interface AdminGroupListResponse {
  groups: any[];
  total: number;
  page: number;
  limit: number;
}

/**
 * 管理员获取所有用户的分组列表
 */
export function getAdminGroups(params?: AdminGroupListParams) {
  return request.get<AdminGroupListResponse>({
    url: Api.AdminGroups,
    params,
  });
}

/**
 * 管理员获取所有分组选项（用于下拉选择）
 */
export function getAdminGroupsAll(params?: { user_id?: number }) {
  return request.get<any[]>({
    url: Api.AdminGroupsAll,
    params,
  });
}

// 管理员日志管理
export interface AdminLogListParams {
  page?: number;
  limit?: number;
  user_id?: number;
  account_id?: number;
  api_key_id?: number;
  model_name?: string;
  is_stream?: boolean;
  start_time?: string;
  end_time?: string;
  min_cost?: number;
  max_cost?: number;
}

export interface AdminLogListResponse {
  logs: any[];
  total: number;
  page: number;
  limit: number;
}

export interface AdminLogStatsResponse {
  total_requests: number;
  total_tokens: number;
  total_cost: number;
  avg_duration: number;
}

/**
 * 管理员获取所有用户的日志列表
 */
export function getAdminLogs(params?: AdminLogListParams) {
  return request.get<AdminLogListResponse>({
    url: Api.AdminLogs,
    params,
  });
}

/**
 * 管理员获取日志统计信息
 */
export function getAdminLogsStats(params?: AdminLogListParams) {
  return request.get<AdminLogStatsResponse>({
    url: Api.AdminLogsStats,
    params,
  });
}

// 导出所有API
export {
  Api as AdminApi,
};
