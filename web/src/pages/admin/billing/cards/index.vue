<template>
  <div class="admin-billing-cards-page">
    <!-- 页面标题 -->
    <t-breadcrumb class="breadcrumb">
      <t-breadcrumb-item>管理员</t-breadcrumb-item>
      <t-breadcrumb-item>计费管理</t-breadcrumb-item>
      <t-breadcrumb-item>充值卡管理</t-breadcrumb-item>
    </t-breadcrumb>

    <!-- 生成充值卡 -->
    <t-card title="生成充值卡" class="generate-card">
      <t-form :data="generateForm" @submit="onGenerateSubmit" ref="generateFormRef" layout="inline">
        <t-form-item label="卡类型" name="cardType" :rules="generateRules.cardType">
          <t-select v-model="generateForm.cardType" placeholder="选择卡类型" style="width: 150px">
            <t-option value="balance" label="余额卡" />
            <t-option value="usage_count" label="次数卡" />
            <t-option value="time_limit" label="时间卡" />
          </t-select>
        </t-form-item>

        <t-form-item label="生成数量" name="count" :rules="generateRules.count">
          <t-input-number v-model="generateForm.count" :min="1" :max="1000" style="width: 120px" />
        </t-form-item>

        <t-form-item label="面值($)" name="value" :rules="generateRules.value">
          <t-input-number v-model="generateForm.value" :min="0.01" :step="0.01" style="width: 120px" />
        </t-form-item>

        <!-- 次数卡专用字段 -->
        <t-form-item
          v-if="generateForm.cardType === 'usage_count'"
          label="使用次数"
          name="usageCount"
          :rules="generateRules.usageCount"
        >
          <t-input-number v-model="generateForm.usageCount" :min="1" style="width: 120px" />
        </t-form-item>

        <!-- 时间卡专用字段 -->
        <template v-if="generateForm.cardType === 'time_limit'">
          <t-form-item label="时间类型" name="timeType" :rules="generateRules.timeType">
            <t-select v-model="generateForm.timeType" placeholder="选择时间类型" style="width: 120px">
              <t-option value="daily" label="日卡" />
              <t-option value="weekly" label="周卡" />
              <t-option value="monthly" label="月卡" />
            </t-select>
          </t-form-item>

          <t-form-item label="有效天数" name="durationDays" :rules="generateRules.durationDays">
            <t-input-number v-model="generateForm.durationDays" :min="1" style="width: 120px" />
          </t-form-item>

          <t-form-item label="每日限制" name="dailyLimit" :rules="generateRules.dailyLimit">
            <t-input-number v-model="generateForm.dailyLimit" :min="1" style="width: 120px" />
          </t-form-item>
        </template>

        <t-form-item label="批次ID" name="batchId">
          <t-input v-model="generateForm.batchId" placeholder="可选" style="width: 150px" />
        </t-form-item>

        <t-form-item>
          <t-button theme="primary" :loading="generateLoading" @click="onGenerateSubmit">生成充值卡</t-button>
        </t-form-item>
      </t-form>
    </t-card>

    <!-- 充值卡列表 -->
    <t-card title="充值卡列表" class="cards-table-card">
      <!-- 筛选条件 -->
      <div class="filter-section">
        <t-form layout="inline" :data="filterForm">
          <t-form-item label="状态">
            <t-select v-model="filterForm.status" placeholder="全部状态" clearable style="width: 120px">
              <t-option value="unused" label="未使用" />
              <t-option value="used" label="已使用" />
              <t-option value="expired" label="已过期" />
              <t-option value="disabled" label="已禁用" />
            </t-select>
          </t-form-item>

          <t-form-item label="卡类型">
            <t-select v-model="filterForm.cardType" placeholder="全部类型" clearable style="width: 120px">
              <t-option value="balance" label="余额卡" />
              <t-option value="usage_count" label="次数卡" />
              <t-option value="time_limit" label="时间卡" />
            </t-select>
          </t-form-item>

          <t-form-item label="批次ID">
            <t-input v-model="filterForm.batchId" placeholder="批次ID" clearable style="width: 150px" />
          </t-form-item>

          <t-form-item>
            <t-button theme="primary" @click="loadCardsData">查询</t-button>
            <t-button theme="default" @click="resetFilter">重置</t-button>
          </t-form-item>
        </t-form>
      </div>

      <!-- 表格 -->
      <t-table
        :data="cardsData"
        :columns="columns"
        :loading="cardsLoading"
        :pagination="{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total: pagination.total,
          showJumper: true,
          showPageSize: true,
          pageSizeOptions: [10, 20, 50, 100],

          onChange: onPageChange,
          onPageSizeChange,
        }"
        row-key="id"
        stripe
        empty="暂无充值卡数据，请先生成充值卡"
      >
        <!-- 卡密列 -->
        <template #card_code="{ row }">
          <div class="card-code">
            <span class="code">{{ row.card_code }}</span>
            <t-button theme="primary" variant="text" size="small" @click="copyCardCode(row.card_code)">复制</t-button>
          </div>
        </template>

        <!-- 卡类型列 -->
        <template #card_type="{ row }">
          <t-tag :theme="getCardTypeTheme(row.card_type)" variant="light">
            {{ getCardTypeLabel(row.card_type) }}
          </t-tag>
        </template>

        <!-- 状态列 -->
        <template #status="{ row }">
          <t-tag :theme="getStatusTheme(row.status)" variant="light">
            {{ getStatusLabel(row.status) }}
          </t-tag>
        </template>

        <!-- 卡信息列 -->
        <template #card_info="{ row }">
          <div class="card-info">
            <div>面值: ${{ (row.value || 0).toFixed(2) }}</div>
            <div v-if="row.card_type === 'usage_count'">次数: {{ row.usage_count || 0 }}</div>
            <div v-if="row.card_type === 'time_limit'">
              {{ getTimeTypeLabel(row.time_type) }} / {{ row.duration_days || 0 }}天 / 每日{{ row.daily_limit || 0 }}次
            </div>
          </div>
        </template>

        <!-- 使用信息列 -->
        <template #usage_info="{ row }">
          <div v-if="row.status === 'used'" class="usage-info">
            <div>用户ID: {{ row.user_id }}</div>
            <div>使用时间: {{ formatDateTime(row.used_at) }}</div>
          </div>
          <span v-else>-</span>
        </template>

        <!-- 过期时间列 -->
        <template #expired_at="{ row }">
          <span v-if="row.expired_at">{{ formatDateTime(row.expired_at) }}</span>
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
      title="修改状态"
      :confirm-btn="{ content: '确认', loading: statusLoading }"
      @confirm="confirmStatusChange"
    >
      <p>确定要将充值卡状态修改为 <strong>{{ getStatusLabel(targetStatus) }}</strong> 吗？</p>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { adminBillingApi } from '@/api/billing';
import type { RechargeCard } from '@/api/billing';

// 响应式数据
const cardsData = ref<RechargeCard[]>([]);
const cardsLoading = ref(false);
const generateLoading = ref(false);
const statusLoading = ref(false);

const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
  showJumper: true,
  showSizeChanger: true,
  pageSizeOptions: [10, 20, 50, 100],
  showTotal: true,
  theme: 'default'
});

// 生成表单
const generateForm = reactive({
  cardType: 'balance',
  count: 1,
  value: 1.00,
  usageCount: 100,
  timeType: 'monthly',
  durationDays: 30,
  dailyLimit: 10,
  batchId: ''
});

const generateFormRef = ref();

// 筛选表单
const filterForm = reactive({
  status: '',
  cardType: '',
  batchId: ''
});

// 状态修改相关
const statusDialogVisible = ref(false);
const selectedCard = ref<RechargeCard | null>(null);
const targetStatus = ref<string>('');

// 表单验证规则
const generateRules = {
  cardType: [{ required: true, message: '请选择卡类型' }],
  count: [{ required: true, message: '请输入生成数量' }],
  value: [{ required: true, message: '请输入面值' }],
  usageCount: [{ required: true, message: '请输入使用次数' }],
  timeType: [{ required: true, message: '请选择时间类型' }],
  durationDays: [{ required: true, message: '请输入有效天数' }],
  dailyLimit: [{ required: true, message: '请输入每日限制' }]
};

// 表格列配置
const columns = [
  { colKey: 'id', title: 'ID', width: 80 },
  { colKey: 'card_code', title: '卡密', width: 200 },
  { colKey: 'card_type', title: '类型', width: 100 },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'card_info', title: '卡信息', width: 200 },
  { colKey: 'usage_info', title: '使用信息', width: 180 },
  { colKey: 'batch_id', title: '批次ID', width: 120 },
  { colKey: 'expired_at', title: '过期时间', width: 160 },
  { colKey: 'created_at', title: '创建时间', width: 160 },
  { colKey: 'action', title: '操作', width: 100, fixed: 'right' as const }
];

// 方法
const loadCardsData = async () => {
  cardsLoading.value = true;
  try {
    const params = {
      page: pagination.value.current,
      page_size: pagination.value.pageSize,
      ...(filterForm.status && { status: filterForm.status }),
      ...(filterForm.cardType && { card_type: filterForm.cardType }),
      ...(filterForm.batchId && { batch_id: filterForm.batchId })
    };

    const response = await adminBillingApi.getRechargeCards(params);

    if (response.success) {
      const responseData = response.data;
      cardsData.value = responseData.cards || [];
      pagination.value.total = responseData.total || 0;

      // 确保当前页不超过总页数
      const totalPages = Math.ceil((responseData.total || 0) / pagination.value.pageSize);
      if (pagination.value.current > totalPages && totalPages > 0) {
        pagination.value.current = totalPages;
        setTimeout(() => loadCardsData(), 100);
      }
    } else {
      MessagePlugin.error('获取充值卡列表失败');
    }
  } catch {
    MessagePlugin.error('获取充值卡列表失败');
  } finally {
    cardsLoading.value = false;
  }
};

const onGenerateSubmit = async () => {
  const validateResult = await generateFormRef.value?.validate();
  if (!validateResult) return;

  generateLoading.value = true;
  try {
    const data: any = {
      card_type: generateForm.cardType,
      count: generateForm.count,
      value: generateForm.value,
    };

    // 只有当批次ID有值时才添加
    if (generateForm.batchId && generateForm.batchId.trim()) {
      data.batch_id = generateForm.batchId.trim();
    }

    if (generateForm.cardType === 'usage_count') {
      data.usage_count = generateForm.usageCount;
    } else if (generateForm.cardType === 'time_limit') {
      data.time_type = generateForm.timeType;
      data.duration_days = generateForm.durationDays;
      data.daily_limit = generateForm.dailyLimit;
    }

    const response = await adminBillingApi.generateRechargeCards(data);

    if (response.success) {
      MessagePlugin.success(response.message || `成功生成 ${data.count} 张充值卡`);

      // 重置表单
      generateForm.cardType = 'balance';
      generateForm.count = 1;
      generateForm.value = 1.00;
      generateForm.batchId = '';

      // 重置到第一页并重新加载数据
      pagination.value.current = 1;
      await loadCardsData();
  } else {
    MessagePlugin.error('生成充值卡失败');
  }
  } catch (error: any) {
    console.error('生成充值卡失败:', error);
    MessagePlugin.error(error?.response?.data?.error?.message || '生成充值卡失败');
  } finally {
    generateLoading.value = false;
  }
};

const onPageChange = (pageInfo: any) => {
  let currentPage = 1;
  if (typeof pageInfo === 'number') {
    currentPage = pageInfo;
  } else if (pageInfo && typeof pageInfo === 'object') {
    currentPage = pageInfo.current || pageInfo.page || 1;
  }

  pagination.value.current = currentPage;
  loadCardsData();
};

const onPageSizeChange = (sizeInfo: any, pageInfo?: any) => {
  let newPageSize = 20;
  let currentPage = 1;

  if (typeof sizeInfo === 'number') {
    newPageSize = sizeInfo;
  } else if (sizeInfo && typeof sizeInfo === 'object') {
    newPageSize = sizeInfo.pageSize || sizeInfo.size || 20;
    currentPage = sizeInfo.current || sizeInfo.page || 1;
  }

  if (pageInfo && typeof pageInfo === 'object') {
    currentPage = pageInfo.current || pageInfo.page || 1;
  }

  pagination.value.pageSize = newPageSize;
  pagination.value.current = currentPage;
  loadCardsData();
};

const resetFilter = () => {
  filterForm.status = '';
  filterForm.cardType = '';
  filterForm.batchId = '';
  pagination.value.current = 1;
  loadCardsData();
};

const copyCardCode = async (cardCode: string) => {
  try {
    await navigator.clipboard.writeText(cardCode);
    MessagePlugin.success('卡密已复制到剪贴板');
  } catch (error) {
    MessagePlugin.error('复制失败');
  }
};

const getCardTypeLabel = (type: string) => {
  const labels = { balance: '余额卡', usage_count: '次数卡', time_limit: '时间卡' };
  return labels[type as keyof typeof labels] || type;
};

const getCardTypeTheme = (type: string): "default" | "primary" | "success" | "warning" | "danger" => {
  const themes: Record<string, "default" | "primary" | "success" | "warning" | "danger"> = {
    balance: 'primary',
    usage_count: 'warning',
    time_limit: 'success'
  };
  return themes[type] || 'default';
};

const getStatusLabel = (status: string) => {
  const labels = { unused: '未使用', used: '已使用', expired: '已过期', disabled: '已禁用' };
  return labels[status as keyof typeof labels] || status;
};

const getStatusTheme = (status: string): "default" | "primary" | "success" | "warning" | "danger" => {
  const themes: Record<string, "default" | "primary" | "success" | "warning" | "danger"> = {
    unused: 'success',
    used: 'primary',
    expired: 'danger',
    disabled: 'warning'
  };
  return themes[status] || 'default';
};

const getTimeTypeLabel = (timeType?: string) => {
  const labels = { daily: '日卡', weekly: '周卡', monthly: '月卡' };
  return labels[timeType as keyof typeof labels] || timeType;
};

const formatDateTime = (dateStr?: string) => {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleString('zh-CN');
};

const getActionOptions = (row: RechargeCard) => {
  const options = [];

  if (row.status === 'unused') {
    options.push(
      { content: '禁用', value: 'mark_disabled' },
      { content: '标记为已使用', value: 'mark_used' },
      { content: '标记为已过期', value: 'mark_expired' }
    );
  } else if (row.status === 'used') {
    options.push({ content: '标记为未使用', value: 'mark_unused' });
  } else if (row.status === 'expired') {
    options.push({ content: '标记为未使用', value: 'mark_unused' });
  } else if (row.status === 'disabled') {
    options.push({ content: '启用', value: 'mark_unused' });
  }

  return options;
};

const handleAction = (action: any, row: RechargeCard) => {
  selectedCard.value = row;

  switch (action.value) {
    case 'mark_used':
      targetStatus.value = 'used';
      break;
    case 'mark_unused':
      targetStatus.value = 'unused';
      break;
    case 'mark_expired':
      targetStatus.value = 'expired';
      break;
    case 'mark_disabled':
      targetStatus.value = 'disabled';
      break;
  }

  statusDialogVisible.value = true;
};

const confirmStatusChange = async () => {
  if (!selectedCard.value) return;

  statusLoading.value = true;
  try {
    const response = await adminBillingApi.updateCardStatus(
      selectedCard.value.id,
      targetStatus.value as any
    );

    if (response.success) {
      MessagePlugin.success(response.message || '状态修改成功');
      statusDialogVisible.value = false;
      await loadCardsData();
    } else {
      MessagePlugin.error('状态修改失败');
    }
  } catch (error: any) {
    console.error('修改状态失败:', error);
    MessagePlugin.error(error?.response?.data?.error?.message || '修改状态失败');
  } finally {
    statusLoading.value = false;
  }
};

// 生命周期
onMounted(() => {
  loadCardsData();
});
</script>

<style scoped lang="less">
.admin-billing-cards-page {
  padding: 24px;

  .breadcrumb {
    margin-bottom: 24px;
  }

  .generate-card {
    margin-bottom: 24px;
  }

  .cards-table-card {
    .filter-section {
      margin-bottom: 16px;
      padding: 16px;
      background: var(--td-bg-color-container-select);
      border-radius: 6px;
    }

    .card-code {
      display: flex;
      align-items: center;
      gap: 8px;

      .code {
        font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
        font-size: 12px;
      }
    }

    .card-info {
      font-size: 12px;
      line-height: 1.4;
    }

    .usage-info {
      font-size: 12px;
      line-height: 1.4;
    }
  }
}
</style>
