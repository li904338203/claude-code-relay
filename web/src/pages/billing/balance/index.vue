<template>
  <div class="billing-balance-page">
    <!-- 页面标题 -->
    <t-breadcrumb class="breadcrumb">
      <t-breadcrumb-item>计费管理</t-breadcrumb-item>
      <t-breadcrumb-item>余额管理</t-breadcrumb-item>
    </t-breadcrumb>

    <!-- 余额概览卡片 -->
    <div class="balance-overview">
      <t-row :gutter="[24, 24]">
        <t-col :lg="6" :md="12" :sm="12">
          <t-card class="balance-card" title="当前余额">
            <div class="balance-amount">
              <span class="currency">$</span>
              <span class="amount">{{ balanceData.balance?.balance?.toFixed(4) || '0.0000' }}</span>
            </div>
            <div class="balance-tip">美元</div>
          </t-card>
        </t-col>

        <t-col :lg="6" :md="12" :sm="12">
          <t-card class="stat-card" title="累计充值">
            <div class="stat-amount">
              ${{ balanceData.balance?.total_recharged?.toFixed(4) || '0.0000' }}
            </div>
          </t-card>
        </t-col>

        <t-col :lg="6" :md="12" :sm="12">
          <t-card class="stat-card" title="累计消费">
            <div class="stat-amount">
              ${{ balanceData.balance?.total_consumed?.toFixed(4) || '0.0000' }}
            </div>
          </t-card>
        </t-col>

        <t-col :lg="6" :md="12" :sm="12">
          <t-card class="stat-card" title="今日消费">
            <div class="stat-amount">
              ${{ balanceData.today_consumption?.toFixed(4) || '0.0000' }}
            </div>
          </t-card>
        </t-col>
      </t-row>
    </div>

    <!-- 充值卡兑换 -->
    <t-card title="充值卡兑换" class="redeem-card">
      <t-form :data="redeemForm" @submit="onRedeemSubmit" ref="redeemFormRef" class="redeem-form">
        <t-form-item label="卡密" name="cardCode" :rules="redeemRules.cardCode">
          <t-input
            v-model="redeemForm.cardCode"
            placeholder="请输入充值卡卡密，格式：XXXX-XXXX-XXXX-XXXX"
            clearable
          />
        </t-form-item>
        <t-form-item>
          <t-button theme="primary" :loading="redeemLoading" @click="onRedeemSubmit">
            兑换充值卡
          </t-button>
        </t-form-item>
      </t-form>
    </t-card>

    <!-- 我的套餐 -->
    <t-card title="我的套餐" class="plans-card">
      <t-loading :loading="plansLoading">
        <!-- 时间卡套餐 -->
        <div v-if="timePlans.length > 0" class="plans-section">
          <h4>时间卡套餐</h4>
          <t-row :gutter="[16, 16]">
            <t-col v-for="plan in timePlans" :key="plan.id" :lg="8" :md="12" :sm="24">
              <div class="plan-item time-plan">
                <div class="plan-header">
                  <span class="plan-type">{{ getTimeTypeLabel(plan.time_type) }}</span>
                  <t-tag :theme="getPlanStatusTheme(plan.status)" variant="light">
                    {{ getPlanStatusLabel(plan.status) }}
                  </t-tag>
                </div>
                <div class="plan-content">
                  <div class="plan-limit">每日限制：{{ plan.daily_limit }}次</div>
                  <div class="plan-usage">今日已用：{{ plan.today_used }}次</div>
                  <div class="plan-remaining">今日剩余：{{ plan.daily_limit - plan.today_used }}次</div>
                  <div class="plan-period">
                    有效期：{{ formatDate(plan.start_date) }} - {{ formatDate(plan.end_date) }}
                  </div>
                </div>
              </div>
            </t-col>
          </t-row>
        </div>

        <!-- 次数卡套餐 -->
        <div v-if="usagePlans.length > 0" class="plans-section">
          <h4>次数卡套餐</h4>
          <t-row :gutter="[16, 16]">
            <t-col v-for="plan in usagePlans" :key="plan.id" :lg="8" :md="12" :sm="24">
              <div class="plan-item usage-plan">
                <div class="plan-header">
                  <span class="plan-type">次数卡</span>
                  <t-tag :theme="getPlanStatusTheme(plan.status)" variant="light">
                    {{ getPlanStatusLabel(plan.status) }}
                  </t-tag>
                </div>
                <div class="plan-content">
                  <div class="plan-total">总次数：{{ plan.total_usage }}次</div>
                  <div class="plan-used">已使用：{{ plan.used_usage }}次</div>
                  <div class="plan-remaining">剩余：{{ plan.remaining_usage }}次</div>
                  <div class="plan-progress">
                    <t-progress
                      :percentage="(plan.used_usage / plan.total_usage) * 100"
                      color="#1890ff"
                      size="small"
                    />
                  </div>
                </div>
              </div>
            </t-col>
          </t-row>
        </div>

        <!-- 无套餐提示 -->
        <div v-if="timePlans.length === 0 && usagePlans.length === 0" class="no-plans">
          <t-empty description="暂无可用套餐，请兑换充值卡获取套餐" />
        </div>
      </t-loading>
    </t-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { billingApi, type BillingStats, type UserCardPlan } from '@/api/billing';

// 响应式数据
const balanceData = ref<BillingStats>({} as BillingStats);
const plansLoading = ref(false);
const redeemLoading = ref(false);

// 充值卡兑换表单
const redeemForm = ref({
  cardCode: ''
});

const redeemFormRef = ref();

// 表单验证规则
const redeemRules = {
  cardCode: [
    { required: true, message: '请输入充值卡卡密' },
    {
      pattern: /^[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$/,
      message: '卡密格式不正确，应为XXXX-XXXX-XXXX-XXXX格式'
    }
  ]
};

// 计算属性
const timePlans = computed(() => balanceData.value.time_plans || []);
const usagePlans = computed(() => balanceData.value.usage_plans || []);

// 方法
const loadBalanceData = async () => {
  try {
    const response = await billingApi.getUserBalance();
    console.log('余额数据响应:', response);

    if (response.success) {
      balanceData.value = response.data;
    } else {
      console.error('余额数据获取失败:', response);
    }
  } catch (error) {
    console.error('获取余额数据失败:', error);
    MessagePlugin.error('获取余额数据失败');
  }
};

const onRedeemSubmit = async () => {
  // 表单验证
  const validateResult = await redeemFormRef.value?.validate();
  if (!validateResult) return;

  redeemLoading.value = true;
  try {
    const response = await billingApi.redeemCard(redeemForm.value.cardCode);
    console.log('充值卡兑换响应:', response);

    if (response.success) {
      MessagePlugin.success(response.message || '充值卡兑换成功');
      redeemForm.value.cardCode = '';
      // 重新加载余额数据
      await loadBalanceData();
    } else {
      MessagePlugin.error('兑换失败');
    }
  } catch (error: any) {
    console.error('充值卡兑换失败:', error);
    MessagePlugin.error(error?.response?.data?.error?.message || '充值卡兑换失败');
  } finally {
    redeemLoading.value = false;
  }
};

const getTimeTypeLabel = (timeType?: string) => {
  const labels = {
    daily: '日卡',
    weekly: '周卡',
    monthly: '月卡'
  };
  return labels[timeType as keyof typeof labels] || '未知';
};

const getPlanStatusLabel = (status: string) => {
  const labels = {
    active: '生效中',
    expired: '已过期',
    exhausted: '已用完'
  };
  return labels[status as keyof typeof labels] || status;
};

const getPlanStatusTheme = (status: string): "default" | "primary" | "success" | "warning" | "danger" => {
  const themes: Record<string, "default" | "primary" | "success" | "warning" | "danger"> = {
    active: 'success',
    expired: 'warning',
    exhausted: 'danger'
  };
  return themes[status] || 'default';
};

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleDateString();
};

// 生命周期
onMounted(() => {
  loadBalanceData();
});
</script>

<style scoped lang="less">
.billing-balance-page {
  padding: 24px;

  .breadcrumb {
    margin-bottom: 24px;
  }

  .balance-overview {
    margin-bottom: 24px;

    .balance-card {
      .balance-amount {
        display: flex;
        align-items: baseline;
        margin-bottom: 8px;

        .currency {
          font-size: 18px;
          color: var(--td-text-color-secondary);
          margin-right: 4px;
        }

        .amount {
          font-size: 32px;
          font-weight: 600;
          color: var(--td-brand-color);
        }
      }

      .balance-tip {
        color: var(--td-text-color-placeholder);
        font-size: 12px;
      }
    }

    .stat-card {
      .stat-amount {
        font-size: 24px;
        font-weight: 500;
        color: var(--td-text-color-primary);
      }
    }
  }

  .redeem-card {
    margin-bottom: 24px;

    .redeem-form {
      max-width: 500px;
    }
  }

  .plans-card {
    .plans-section {
      margin-bottom: 24px;

      h4 {
        margin-bottom: 16px;
        color: var(--td-text-color-primary);
      }

      .plan-item {
        border: 1px solid var(--td-border-level-2-color);
        border-radius: 6px;
        padding: 16px;
        background: var(--td-bg-color-container);
        transition: all 0.2s;

        &:hover {
          border-color: var(--td-brand-color);
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
        }

        .plan-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 12px;

          .plan-type {
            font-weight: 500;
            color: var(--td-text-color-primary);
          }
        }

        .plan-content {
          font-size: 14px;

          > div {
            margin-bottom: 8px;
            color: var(--td-text-color-secondary);

            &:last-child {
              margin-bottom: 0;
            }
          }

          .plan-progress {
            margin-top: 12px;
          }
        }

        &.time-plan {
          border-left: 4px solid var(--td-success-color);
        }

        &.usage-plan {
          border-left: 4px solid var(--td-warning-color);
        }
      }
    }

    .no-plans {
      text-align: center;
      padding: 40px 0;
    }
  }
}
</style>
