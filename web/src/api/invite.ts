import { request } from '@/utils/request';

// 邀请配置接口
export interface InviteConfig {
  invite_reward_amount: string;
  max_invite_count_default: string;
  invite_system_enabled: string;
}

// 平台邀请统计接口
export interface PlatformInviteStats {
  total_invites: number;
  total_rewards: number;
  total_users: number;
  success_registers: number;
  conversion_rate: number;
  invited_users: number;
}

// 邀请记录接口
export interface Invitation {
  id: number;
  inviter_id: number;
  invitee_id: number;
  reward_amount: number;
  status: string;
  created_at: string;
  inviter?: {
    username: string;
  };
  invitee?: {
    username: string;
  };
}

// 邀请列表响应接口
export interface InviteListResponse {
  invitations: Invitation[];
  total: number;
  page: number;
  limit: number;
}

/**
 * 获取邀请配置
 */
export function getInviteConfig(): Promise<{ success: boolean; data: InviteConfig; code?: number }> {
  return request.get({
    url: '/api/v1/admin/invite/config',
  });
}

/**
 * 更新邀请配置
 */
export function updateInviteConfig(config: InviteConfig): Promise<{ success: boolean; message?: string; code?: number }> {
  return request.put({
    url: '/api/v1/admin/invite/config',
    data: config,
  });
}

/**
 * 获取平台邀请统计
 */
export function getPlatformInviteStats(): Promise<{ success: boolean; data: PlatformInviteStats; code?: number }> {
  return request.get({
    url: '/api/v1/admin/invite/stats',
  });
}

/**
 * 设置用户邀请限额
 */
export function updateUserInviteLimit(userId: number, maxCount: number): Promise<{ success: boolean; message?: string }> {
  return request.put({
    url: `/api/v1/admin/invite/user/${userId}/limit`,
    data: {
      max_count: maxCount,
    },
  });
}

/**
 * 获取邀请列表
 */
export function getInviteList(params: {
  page?: number;
  limit?: number;
  inviter_id?: number;
  invitee_id?: number;
}): Promise<{ success: boolean; data: InviteListResponse }> {
  return request.get({
    url: '/api/v1/admin/invite/list',
    params,
  });
}
