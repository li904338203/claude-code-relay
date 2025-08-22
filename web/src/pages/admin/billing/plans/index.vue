<template>
  <div class="admin-billing-plans-page">
    <!-- 页面标题 -->
    <t-breadcrumb class="breadcrumb">
      <t-breadcrumb-item>管理员</t-breadcrumb-item>
      <t-breadcrumb-item>计费管理</t-breadcrumb-item>
      <t-breadcrumb-item>用户套餐管理</t-breadcrumb-item>
    </t-breadcrumb>

    <!-- 用户套餐列表 -->
    <t-card title="用户套餐列表" class="plans-table-card">
      <!-- 筛选条件 -->
      <div class="filter-section">
        <t-form layout="inline" :data="filterForm">
          <t-form-item label="状态">
            <t-select v-model="filterForm.status" placeholder="全部状态" clearable style="width: 120px">
              <t-option value="active" label="活跃" />
              <t-option value="expired" label="已过期" />
              <t-option value="exhausted" label="已耗尽" />
              <t-option value="disabled" label="已禁用" />
            </t-select>
          </t-form-item>

          <t-form-item label="套餐类型">
            <t-select v-model="filterForm.planType" placeholder="全部类型" clearable style="width: 120px">
              <t-option value="usage_count" label="次数卡" />
              <t-option value="time_limit" label="时间卡" />
            </t-select>
          </t-form-item>

          <t-form-item label="用户ID">
            <t-input v-model="filterForm.userID" placeholder="用户ID" clearable style="width: 150px" />
          </t-form-item>

          <t-form-item>
            <t-button theme="primary" @click="loadPlansData">查询</t-button>
            <t-button theme="default" @click="resetFilter">重置</t-button>
          </t-form-item>
        </t-form>
      </div>

      <!-- 表格 -->
      <t-table
        :data="plansData"
        :columns="columns"
        :loading="plansLoading"
        :pagination="{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total: pagination.total,
          showJumper: true,
          showPageSize: true,
          pageSizeOptions: [10, 20, 50, 100],

          onChange: (pageInfo: any) => onPageChange(pageInfo.current),
          onPageSizeChange,
        }"
        row-key="id"
        stripe
        empty="暂无用户套餐数据"
      >
        <!-- 用户信息列 -->
        <template #user_info="{ row }">
          <div class="user-info">
            <div>ID: {{ row.user_id }}</div>
            <div v-if="row.user">用户名: {{ row.user.username }}</div>
          </div>
        </template>

        <!-- 套餐类型列 -->
        <template #plan_type="{ row }">
          <t-tag :theme="getPlanTypeTheme(row.plan_type)" variant="light">
            {{ getPlanTypeLabel(row.plan_type) }}
          </t-tag>
        </template>

        <!-- 状态列 -->
        <template #status="{ row }">
          <t-tag :theme="getStatusTheme(row.status)" variant="light">
            {{ getStatusLabel(row.status) }}
          </t-tag>
        </template>

        <!-- 套餐信息列 -->
        <template #plan_info="{ row }">
          <div class="plan-info">
            <div v-if="row.plan_type === 'usage_count'">
              <div>总次数: {{ row.total_usage }}</div>
              <div>已用: {{ row.used_usage }}</div>
              <div>剩余: {{ row.remaining_usage }}</div>
            </div>
            <div v-else-if="row.plan_type === 'time_limit'">
              <div>类型: {{ getTimeTypeLabel(row.time_type) }}</div>
              <div>每日限制: {{ row.daily_limit }}</div>
              <div>今日已用: {{ row.today_used }}</div>
            </div>
          </div>
        </template>

        <!-- 有效期列 -->
        <template #validity="{ row }">
          <div v-if="row.plan_type === 'time_limit'" class="validity-info">
            <div v-if="row.start_date">开始: {{ formatDate(row.start_date) }}</div>
            <div v-if="row.end_date">结束: {{ formatDate(row.end_date) }}</div>
          </div>
          <span v-else>-</span>
        </template>

        <!-- 充值卡信息列 -->
        <template #card_info="{ row }">
          <div v-if="row.recharge_card" class="card-info">
            <div>卡密: {{ row.recharge_card.card_code }}</div>
            <div>面值: ${{ (row.recharge_card.value || 0).toFixed(2) }}</div>
          </div>
          <span v-else>-</span>
        </template>

        <!-- 创建时间列 -->
        <template #created_at="{ row }">
          <span>{{ formatDateTime(row.created_at) }}</span>
        </template>

        <!-- 操作列 -->
        <template #action="{ row }">
          <t-dropdown :options="getActionOptions(row)" @click="handleAction($event, row)">
            <t-button theme="primary" variant="text">
              操作
              <template #suffix>
                <t-icon name="chevron-down" />
              </template>
            </t-button>
          </t-dropdown>
        </template>
      </t-table>
    </t-card>

    <!-- 状态修改确认对话框 -->
    <t-dialog
      v-model:visible="statusDialogVisible"
      title="修改用户套餐状态"
      :confirm-btn="{ content: '确认', loading: statusLoading }"
      @confirm="confirmStatusChange"
    >
      <p>确定要将用户套餐状态修改为 <strong>{{ getStatusLabel(targetStatus) }}</strong> 吗？</p>
      <div v-if="selectedPlan" class="plan-summary">
        <p>用户ID: {{ selectedPlan.user_id }}</p>
        <p>套餐类型: {{ getPlanTypeLabel(selectedPlan.plan_type) }}</p>
        <p>当前状态: {{ getStatusLabel(selectedPlan.status) }}</p>
      </div>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { adminBillingApi } from '@/api/billing';

// 用户套餐接口类型
interface UserCardPlan {
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
  last_reset_date?: string;
  status: 'active' | 'expired' | 'exhausted' | 'disabled';
  created_at: string;
  updated_at: string;
  user?: {
    id: number;
    username: string;
  };
  recharge_card?: {
    id: number;
    card_code: string;
    value: number;
  };
}

// 响应式数据
const plansData = ref<UserCardPlan[]>([]);
const plansLoading = ref(false);
const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
});

// 筛选表单
const filterForm = reactive({
  status: '',
  planType: '',
  userID: '',
});

// 状态修改相关
const statusDialogVisible = ref(false);
const statusLoading = ref(false);
const selectedPlan = ref<UserCardPlan | null>(null);
const targetStatus = ref('');

// 表格列配置
const columns = [
  {
    colKey: 'id',
    title: 'ID',
    width: 80,
  },
  {
    colKey: 'user_info',
    title: '用户信息',
    width: 120,
  },
  {
    colKey: 'plan_type',
    title: '套餐类型',
    width: 100,
  },
  {
    colKey: 'status',
    title: '状态',
    width: 80,
  },
  {
    colKey: 'plan_info',
    title: '套餐信息',
    width: 150,
  },
  {
    colKey: 'validity',
    title: '有效期',
    width: 140,
  },
  {
    colKey: 'card_info',
    title: '充值卡信息',
    width: 150,
  },
  {
    colKey: 'created_at',
    title: '创建时间',
    width: 160,
  },
  {
    colKey: 'action',
    title: '操作',
    width: 100,
    fixed: 'right' as const,
  },
];

// 方法
const loadPlansData = async () => {
  plansLoading.value = true;
  try {
    const params: any = {
      page: pagination.value.current,
      page_size: pagination.value.pageSize,
    };

    if (filterForm.status) params.status = filterForm.status;
    if (filterForm.planType) params.plan_type = filterForm.planType;
    if (filterForm.userID) params.user_id = filterForm.userID;

    const response = await adminBillingApi.getAllUserPlans(params);

    if (response.success) {
      plansData.value = (response.data as any).plans || [];
      pagination.value.total = response.data.total || 0;
    } else {
      console.error('用户套餐数据获取失败:', response);
    }
  } catch (error) {
    console.error('获取用户套餐失败:', error);
    MessagePlugin.error('获取用户套餐失败');
  } finally {
    plansLoading.value = false;
  }
};

const onPageChange = (current: number) => {
  pagination.value.current = current;
  loadPlansData();
};

const onPageSizeChange = (pageSize: number) => {
  pagination.value.pageSize = pageSize;
  pagination.value.current = 1;
  loadPlansData();
};

const resetFilter = () => {
  filterForm.status = '';
  filterForm.planType = '';
  filterForm.userID = '';
  pagination.value.current = 1;
  loadPlansData();
};

const getPlanTypeLabel = (type: string) => {
  const labels = { usage_count: '次数卡', time_limit: '时间卡' };
  return labels[type as keyof typeof labels] || type;
};

const getPlanTypeTheme = (type: string): "default" | "primary" | "success" | "warning" | "danger" => {
  const themes: Record<string, "default" | "primary" | "success" | "warning" | "danger"> = {
    usage_count: 'warning',
    time_limit: 'success'
  };
  return themes[type] || 'default';
};

const getStatusLabel = (status: string) => {
  const labels = {
    active: '活跃',
    expired: '已过期',
    exhausted: '已耗尽',
    disabled: '已禁用'
  };
  return labels[status as keyof typeof labels] || status;
};

const getStatusTheme = (status: string): "default" | "primary" | "success" | "warning" | "danger" => {
  const themes: Record<string, "default" | "primary" | "success" | "warning" | "danger"> = {
    active: 'success',
    expired: 'danger',
    exhausted: 'warning',
    disabled: 'default'
  };
  return themes[status] || 'default';
};

const getTimeTypeLabel = (timeType?: string) => {
  const labels = { daily: '日卡', weekly: '周卡', monthly: '月卡' };
  return labels[timeType as keyof typeof labels] || timeType || '-';
};

const formatDateTime = (dateStr?: string) => {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleString('zh-CN');
};

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleDateString('zh-CN');
};

const getActionOptions = (row: UserCardPlan) => {
  const options = [];

  if (row.status === 'active') {
    options.push(
      { content: '禁用', value: 'mark_disabled' },
      { content: '标记已过期', value: 'mark_expired' },
      { content: '标记已耗尽', value: 'mark_exhausted' }
    );
  } else if (row.status === 'disabled') {
    options.push({ content: '启用', value: 'mark_active' });
  } else if (row.status === 'expired' || row.status === 'exhausted') {
    options.push(
      { content: '重新激活', value: 'mark_active' },
      { content: '禁用', value: 'mark_disabled' }
    );
  }

  return options;
};

const handleAction = (action: any, row: UserCardPlan) => {
  selectedPlan.value = row;

  switch (action.value) {
    case 'mark_active':
      targetStatus.value = 'active';
      break;
    case 'mark_expired':
      targetStatus.value = 'expired';
      break;
    case 'mark_exhausted':
      targetStatus.value = 'exhausted';
      break;
    case 'mark_disabled':
      targetStatus.value = 'disabled';
      break;
  }

  statusDialogVisible.value = true;
};

const confirmStatusChange = async () => {
  if (!selectedPlan.value) return;

  statusLoading.value = true;
  try {
    const response = await adminBillingApi.updateUserPlanStatus(
      selectedPlan.value.id,
      targetStatus.value as any
    );

    if (response.success) {
      MessagePlugin.success(response.message || '套餐状态修改成功');
      statusDialogVisible.value = false;
      await loadPlansData();
    } else {
      MessagePlugin.error('套餐状态修改失败');
    }
  } catch (error: any) {
    console.error('修改套餐状态失败:', error);
    MessagePlugin.error(error?.response?.data?.error?.message || '修改套餐状态失败');
  } finally {
    statusLoading.value = false;
  }
};

// 生命周期
onMounted(() => {
  loadPlansData();
});
</script>

<style scoped lang="less">
.admin-billing-plans-page {
  padding: 24px;

  .breadcrumb {
    margin-bottom: 24px;
  }

  .plans-table-card {
    .filter-section {
      margin-bottom: 16px;
      padding: 16px;
      background: var(--td-bg-color-container-select);
      border-radius: 6px;
    }

    .user-info {
      font-size: 12px;
      line-height: 1.4;
    }

    .plan-info {
      font-size: 12px;
      line-height: 1.4;
    }

    .validity-info {
      font-size: 12px;
      line-height: 1.4;
    }

    .card-info {
      font-size: 12px;
      line-height: 1.4;

      .code {
        font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
      }
    }
  }

  .plan-summary {
    margin-top: 16px;
    padding: 12px;
    background: var(--td-bg-color-container-select);
    border-radius: 4px;
    font-size: 12px;

    p {
      margin: 4px 0;
    }
  }
}
</style>
