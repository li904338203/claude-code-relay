import { request } from '@/utils/request';

// 计费相关API路径
const Api = {
  // 用户接口
  GetUserBalance: '/api/v1/billing/balance',
  RedeemCard: '/api/v1/billing/redeem',
  GetUserPlans: '/api/v1/billing/plans',
  GetConsumptionHistory: '/api/v1/billing/consumption',

  // 管理员接口
  GenerateRechargeCards: '/api/v1/admin/billing/cards/generate',
  GetRechargeCards: '/api/v1/admin/billing/cards',
  UpdateCardStatus: '/api/v1/admin/billing/cards',
  GetAllUserPlans: '/api/v1/admin/billing/plans',
  UpdateUserPlanStatus: '/api/v1/admin/billing/plans',
  RechargeUserBalance: '/api/v1/admin/billing/balance/recharge',
  GetConsumptionStats: '/api/v1/admin/billing/stats',
  GetBillingConfig: '/api/v1/admin/billing/config',
  UpdateBillingConfig: '/api/v1/admin/billing/config',
};

// 计费相关API类型定义
export interface UserBalance {
  id: number;
  user_id: number;
  balance: number;
  frozen_balance: number;
  total_recharged: number;
  total_consumed: number;
  created_at: string;
  updated_at: string;
}

export interface UserCardPlan {
  id: number;
  user_id: number;
  card_id: number;
  plan_type: 'usage_count' | 'time_limit';
  total_usage: number;
  used_usage: number;
  remaining_usage: number;
  time_type?: 'daily' | 'weekly' | 'monthly';
  daily_limit: number;
  start_date?: string;
  end_date?: string;
  today_used: number;
  status: 'active' | 'expired' | 'exhausted' | 'disabled';
  created_at: string;
  recharge_card?: RechargeCard;
}

export interface RechargeCard {
  id: number;
  card_code: string;
  card_type: 'usage_count' | 'time_limit' | 'balance';
  usage_count: number;
  time_type?: 'daily' | 'weekly' | 'monthly';
  duration_days: number;
  daily_limit: number;
  value: number;
  status: 'unused' | 'used' | 'expired' | 'disabled';
  user_id?: number;
  used_at?: string;
  expired_at?: string;
  batch_id?: string;
  created_by?: number;
  created_at: string;
}

export interface ConsumptionLog {
  id: number;
  user_id: number;
  plan_id?: number;
  request_id?: string;
  api_key_id?: number;
  account_id?: number;
  cost_usd: number;
  usage_count: number;
  deduction_type: 'balance' | 'usage_count' | 'time_limit';
  input_tokens: number;
  output_tokens: number;
  cache_read_tokens: number;
  cache_creation_tokens: number;
  total_tokens: number;
  model?: string;
  platform_type?: string;
  is_stream: boolean;
  balance_before?: number;
  balance_after?: number;
  created_at: string;
}

export interface BillingStats {
  balance: UserBalance;
  time_plans: UserCardPlan[];
  usage_plans: UserCardPlan[];
  total_consumption: number;
  today_consumption: number;
}

export interface ConsumptionStats {
  total_stats: {
    total_consumption: number;
    total_users: number;
    total_requests: number;
  };
  today_stats: {
    today_consumption: number;
    today_requests: number;
  };
  user_ranking: Array<{
    user_id: number;
    username: string;
    consumption: number;
    request_count: number;
  }>;
}

export interface BillingConfig {
  id: number;
  config_key: string;
  config_value: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

// 用户计费接口
export const billingApi = {
  // 获取用户余额和套餐信息
  getUserBalance: (): Promise<{ success: boolean; data: BillingStats }> => {
    return request.get({
      url: Api.GetUserBalance,
    });
  },

  // 充值卡兑换
  redeemCard: (cardCode: string): Promise<{ success: boolean; message: string }> => {
    return request.post({
      url: Api.RedeemCard,
      data: { card_code: cardCode },
    });
  },

  // 获取用户套餐列表
  getUserPlans: (): Promise<{ success: boolean; data: UserCardPlan[] }> => {
    return request.get({
      url: Api.GetUserPlans,
    });
  },

  // 获取消费历史
  getConsumptionHistory: (params: {
    page?: number;
    page_size?: number;
  }): Promise<{
    success: boolean;
    data: {
      logs: ConsumptionLog[];
      total: number;
      page: number;
      page_size: number;
      total_page: number;
    };
  }> => {
    return request.get({
      url: Api.GetConsumptionHistory,
      params,
    });
  },
};

// 管理员计费接口
export const adminBillingApi = {
  // 生成充值卡
  generateRechargeCards: (data: {
    card_type: 'usage_count' | 'time_limit' | 'balance';
    count: number;
    value: number;
    usage_count?: number;
    time_type?: 'daily' | 'weekly' | 'monthly';
    duration_days?: number;
    daily_limit?: number;
    batch_id?: string;
    expired_at?: string;
  }): Promise<{ success: boolean; message: string; data: { count: number } }> => {
    return request.post({
      url: Api.GenerateRechargeCards,
      data,
    });
  },

  // 查询充值卡列表
  getRechargeCards: (params: {
    page?: number;
    page_size?: number;
    status?: string;
    card_type?: string;
    batch_id?: string;
  }): Promise<{
    success: boolean;
    data: {
      cards: RechargeCard[];
      total: number;
      page: number;
      page_size: number;
      total_page: number;
    };
  }> => {
    return request.get({
      url: Api.GetRechargeCards,
      params,
    });
  },

  // 修改充值卡状态
  updateCardStatus: (
    id: number,
    status: 'unused' | 'used' | 'expired' | 'disabled',
  ): Promise<{ success: boolean; message: string }> => {
    return request.put({
      url: `${Api.UpdateCardStatus}/${id}/status`,
      data: { status },
    });
  },

  // 获取所有用户套餐列表
  getAllUserPlans: (params: {
    page?: number;
    page_size?: number;
    status?: string;
    plan_type?: string;
    user_id?: string;
  }): Promise<{
    success: boolean;
    data: {
      plans: UserCardPlan[];
      total: number;
      page: number;
      page_size: number;
      total_page: number;
    };
  }> => {
    return request.get({
      url: Api.GetAllUserPlans,
      params,
    });
  },

  // 修改用户套餐状态
  updateUserPlanStatus: (
    id: number,
    status: 'active' | 'expired' | 'exhausted' | 'disabled',
  ): Promise<{ success: boolean; message: string }> => {
    return request.put({
      url: `${Api.UpdateUserPlanStatus}/${id}/status`,
      data: { status },
    });
  },

  // 手动充值用户余额
  rechargeUserBalance: (data: {
    user_id: number;
    amount: number;
    description?: string;
  }): Promise<{ success: boolean; message: string }> => {
    return request.post({
      url: Api.RechargeUserBalance,
      data,
    });
  },

  // 获取消费统计
  getConsumptionStats: (): Promise<{ success: boolean; data: ConsumptionStats }> => {
    return request.get({
      url: Api.GetConsumptionStats,
    });
  },

  // 获取计费配置
  getBillingConfig: (): Promise<{ success: boolean; data: BillingConfig[] }> => {
    return request.get({
      url: Api.GetBillingConfig,
    });
  },

  // 更新计费配置
  updateBillingConfig: (data: {
    config_key: string;
    config_value: string;
    description?: string;
  }): Promise<{ success: boolean; message: string }> => {
    return request.put({
      url: Api.UpdateBillingConfig,
      data,
    });
  },
};

export default billingApi;
