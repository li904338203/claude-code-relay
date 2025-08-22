<template>
  <div class="admin-billing-stats-page">
    <!-- 页面标题 -->
    <t-breadcrumb class="breadcrumb">
      <t-breadcrumb-item>管理员</t-breadcrumb-item>
      <t-breadcrumb-item>计费管理</t-breadcrumb-item>
      <t-breadcrumb-item>消费统计</t-breadcrumb-item>
    </t-breadcrumb>

    <!-- 总体统计 -->
    <div class="stats-overview">
      <t-row :gutter="[24, 24]">
        <t-col :lg="3" :md="6" :sm="12">
          <t-card class="stat-card">
            <div class="stat-item">
              <div class="stat-icon total">
                <t-icon name="money-circle" />
              </div>
              <div class="stat-content">
                <div class="stat-value">${{ totalStats.total_consumption?.toFixed(4) || '0.0000' }}</div>
                <div class="stat-label">总消费金额</div>
              </div>
            </div>
          </t-card>
        </t-col>

        <t-col :lg="3" :md="6" :sm="12">
          <t-card class="stat-card">
            <div class="stat-item">
              <div class="stat-icon today">
                <t-icon name="calendar" />
              </div>
              <div class="stat-content">
                <div class="stat-value">${{ todayStats.today_consumption?.toFixed(4) || '0.0000' }}</div>
                <div class="stat-label">今日消费</div>
              </div>
            </div>
          </t-card>
        </t-col>

        <t-col :lg="3" :md="6" :sm="12">
          <t-card class="stat-card">
            <div class="stat-item">
              <div class="stat-icon users">
                <t-icon name="user" />
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ totalStats.total_users || 0 }}</div>
                <div class="stat-label">总用户数</div>
              </div>
            </div>
          </t-card>
        </t-col>

        <t-col :lg="3" :md="6" :sm="12">
          <t-card class="stat-card">
            <div class="stat-item">
              <div class="stat-icon requests">
                <t-icon name="api" />
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ totalStats.total_requests || 0 }}</div>
                <div class="stat-label">总请求数</div>
              </div>
            </div>
          </t-card>
        </t-col>
      </t-row>
    </div>

    <!-- 今日统计 -->
    <t-card title="今日详细统计" class="today-stats-card">
      <t-row :gutter="[24, 24]">
        <t-col :lg="6" :md="12" :sm="24">
          <div class="today-stat-item">
            <div class="label">今日消费金额</div>
            <div class="value primary">${{ todayStats.today_consumption?.toFixed(4) || '0.0000' }}</div>
          </div>
        </t-col>

        <t-col :lg="6" :md="12" :sm="24">
          <div class="today-stat-item">
            <div class="label">今日请求次数</div>
            <div class="value success">{{ todayStats.today_requests || 0 }}</div>
          </div>
        </t-col>

        <t-col :lg="6" :md="12" :sm="24">
          <div class="today-stat-item">
            <div class="label">今日平均费用</div>
            <div class="value info">
              ${{ todayAverageCost.toFixed(6) }}
            </div>
          </div>
        </t-col>

        <t-col :lg="6" :md="12" :sm="24">
          <div class="today-stat-item">
            <div class="label">对比昨日</div>
            <div class="value" :class="todayGrowthClass">
              {{ todayGrowthText }}
            </div>
          </div>
        </t-col>
      </t-row>
    </t-card>

    <!-- 用户消费排行榜 -->
    <t-card title="用户消费排行榜 (前10名)" class="ranking-card">
      <t-table
        :data="userRanking"
        :columns="rankingColumns"
        :loading="statsLoading"
        :pagination="null"
        row-key="user_id"
        stripe
      >
        <!-- 排名列 -->
        <template #rank="{ rowIndex }">
          <div class="rank-badge" :class="getRankClass(rowIndex)">
            {{ rowIndex + 1 }}
          </div>
        </template>

        <!-- 用户名列 -->
        <template #username="{ row }">
          <div class="username-cell">
            <span class="username">{{ row.username }}</span>
            <span class="user-id">ID: {{ row.user_id }}</span>
          </div>
        </template>

        <!-- 消费金额列 -->
        <template #consumption="{ row }">
          <span class="consumption-amount">${{ row.consumption?.toFixed(4) || '0.0000' }}</span>
        </template>

        <!-- 请求次数列 -->
        <template #request_count="{ row }">
          <span class="request-count">{{ row.request_count || 0 }}</span>
        </template>

        <!-- 平均费用列 -->
        <template #avg_cost="{ row }">
          <span class="avg-cost">
            ${{ (row.request_count > 0 ? (row.consumption / row.request_count) : 0).toFixed(6) }}
          </span>
        </template>

        <!-- 占比列 -->
        <template #percentage="{ row }">
          <div class="percentage-cell">
            <span class="percentage-text">
              {{ ((row.consumption / totalStats.total_consumption) * 100).toFixed(2) }}%
            </span>
            <t-progress
              :percentage="(row.consumption / totalStats.total_consumption) * 100"
              color="#1890ff"
              size="small"
              :show-text="false"
            />
          </div>
        </template>
      </t-table>
    </t-card>

    <!-- 用户余额管理 -->
    <t-card title="用户余额管理" class="balance-management-card">
      <t-form :data="rechargeForm" @submit="onRechargeSubmit" ref="rechargeFormRef" layout="inline">
        <t-form-item label="用户ID" name="userId" :rules="rechargeRules.userId">
          <t-input-number v-model="rechargeForm.userId" :min="1" placeholder="用户ID" style="width: 120px" />
        </t-form-item>

        <t-form-item label="充值金额($)" name="amount" :rules="rechargeRules.amount">
          <t-input-number v-model="rechargeForm.amount" :min="0.01" :step="0.01" placeholder="金额" style="width: 120px" />
        </t-form-item>

        <t-form-item label="充值说明" name="description">
          <t-input v-model="rechargeForm.description" placeholder="可选" style="width: 200px" />
        </t-form-item>

        <t-form-item>
          <t-button theme="primary" :loading="rechargeLoading" @click="onRechargeSubmit">
            充值
          </t-button>
        </t-form-item>
      </t-form>
    </t-card>

    <!-- 刷新按钮 -->
    <div class="refresh-section">
      <t-button theme="primary" :loading="statsLoading" @click="loadStatsData">
        <template #icon>
          <t-icon name="refresh" />
        </template>
        刷新数据
      </t-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, reactive } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { adminBillingApi, type ConsumptionStats } from '@/api/billing';

// 响应式数据
const statsData = ref<ConsumptionStats>({} as ConsumptionStats);
const statsLoading = ref(false);
const rechargeLoading = ref(false);

// 充值表单
const rechargeForm = reactive({
  userId: undefined as number | undefined,
  amount: undefined as number | undefined,
  description: ''
});

const rechargeFormRef = ref();

// 计算属性
const totalStats = computed(() => statsData.value.total_stats || {} as any);
const todayStats = computed(() => statsData.value.today_stats || {} as any);
const userRanking = computed(() => statsData.value.user_ranking || []);

const todayAverageCost = computed(() => {
  const requests = todayStats.value.today_requests || 0;
  const consumption = todayStats.value.today_consumption || 0;
  return requests > 0 ? consumption / requests : 0;
});

const todayGrowthText = computed(() => {
  // 这里应该有昨日数据比较，暂时显示占位信息
  return '暂无昨日数据';
});

const todayGrowthClass = computed(() => {
  return 'warning';
});

// 表单验证规则
const rechargeRules = {
  userId: [{ required: true, message: '请输入用户ID' }],
  amount: [{ required: true, message: '请输入充值金额' }]
};

// 排行榜表格列配置
const rankingColumns = [
  { colKey: 'rank', title: '排名', width: 80 },
  { colKey: 'username', title: '用户', width: 150 },
  { colKey: 'consumption', title: '消费金额', width: 120 },
  { colKey: 'request_count', title: '请求次数', width: 100 },
  { colKey: 'avg_cost', title: '平均费用', width: 120 },
  { colKey: 'percentage', title: '占比', width: 150 }
];

// 方法
const loadStatsData = async () => {
  statsLoading.value = true;
  try {
    const response = await adminBillingApi.getConsumptionStats();
    console.log('统计数据响应:', response);

    if (response.success) {
      statsData.value = response.data;
    } else {
      console.error('统计数据获取失败:', response);
    }
  } catch (error) {
    console.error('获取消费统计失败:', error);
    MessagePlugin.error('获取消费统计失败');
  } finally {
    statsLoading.value = false;
  }
};

const onRechargeSubmit = async () => {
  const validateResult = await rechargeFormRef.value?.validate();
  if (!validateResult) return;

  rechargeLoading.value = true;
  try {
    const response = await adminBillingApi.rechargeUserBalance({
      user_id: rechargeForm.userId!,
      amount: rechargeForm.amount!,
      description: rechargeForm.description || undefined
    });

    console.log('用户充值响应:', response);

    if (response.success) {
      MessagePlugin.success(response.message || '充值成功');
      rechargeForm.userId = undefined;
      rechargeForm.amount = undefined;
      rechargeForm.description = '';
      // 刷新统计数据
      await loadStatsData();
    } else {
      MessagePlugin.error('充值失败');
    }
  } catch (error: any) {
    console.error('充值失败:', error);
    MessagePlugin.error(error?.response?.data?.error?.message || '充值失败');
  } finally {
    rechargeLoading.value = false;
  }
};

const getRankClass = (index: number) => {
  if (index === 0) return 'gold';
  if (index === 1) return 'silver';
  if (index === 2) return 'bronze';
  return 'default';
};

const getProgressTheme = (percentage: number) => {
  if (percentage >= 0.3) return 'default';
  if (percentage >= 0.1) return 'default';
  return 'default';
};

// 生命周期
onMounted(() => {
  loadStatsData();
});
</script>

<style scoped lang="less">
.admin-billing-stats-page {
  padding: 24px;

  .breadcrumb {
    margin-bottom: 24px;
  }

  .stats-overview {
    margin-bottom: 24px;

    .stat-card {
      .stat-item {
        display: flex;
        align-items: center;
        gap: 16px;

        .stat-icon {
          width: 48px;
          height: 48px;
          border-radius: 8px;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 24px;
          color: white;

          &.total {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          }

          &.today {
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
          }

          &.users {
            background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
          }

          &.requests {
            background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
          }
        }

        .stat-content {
          .stat-value {
            font-size: 24px;
            font-weight: 600;
            color: var(--td-text-color-primary);
            margin-bottom: 4px;
          }

          .stat-label {
            font-size: 14px;
            color: var(--td-text-color-secondary);
          }
        }
      }
    }
  }

  .today-stats-card {
    margin-bottom: 24px;

    .today-stat-item {
      text-align: center;

      .label {
        font-size: 14px;
        color: var(--td-text-color-secondary);
        margin-bottom: 8px;
      }

      .value {
        font-size: 20px;
        font-weight: 600;

        &.primary {
          color: var(--td-brand-color);
        }

        &.success {
          color: var(--td-success-color);
        }

        &.info {
          color: var(--td-info-color);
        }

        &.warning {
          color: var(--td-warning-color);
        }
      }
    }
  }

  .ranking-card {
    margin-bottom: 24px;

    .rank-badge {
      width: 24px;
      height: 24px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 12px;
      font-weight: 600;
      color: white;

      &.gold {
        background: linear-gradient(135deg, #ffd700, #ffb347);
      }

      &.silver {
        background: linear-gradient(135deg, #c0c0c0, #a8a8a8);
      }

      &.bronze {
        background: linear-gradient(135deg, #cd7f32, #b8860b);
      }

      &.default {
        background: var(--td-text-color-placeholder);
      }
    }

    .username-cell {
      .username {
        display: block;
        font-weight: 500;
        color: var(--td-text-color-primary);
      }

      .user-id {
        display: block;
        font-size: 12px;
        color: var(--td-text-color-placeholder);
      }
    }

    .consumption-amount {
      font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
      font-weight: 500;
      color: var(--td-brand-color);
    }

    .request-count {
      font-weight: 500;
    }

    .avg-cost {
      font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
      font-size: 12px;
    }

    .percentage-cell {
      .percentage-text {
        display: block;
        margin-bottom: 4px;
        font-size: 12px;
        font-weight: 500;
      }
    }
  }

  .balance-management-card {
    margin-bottom: 24px;
  }

  .refresh-section {
    text-align: center;
  }
}
</style>
