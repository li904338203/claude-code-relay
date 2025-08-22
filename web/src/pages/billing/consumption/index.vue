<template>
  <div class="consumption-history-page">
    <!-- 页面标题 -->
    <t-breadcrumb class="breadcrumb">
      <t-breadcrumb-item>计费管理</t-breadcrumb-item>
      <t-breadcrumb-item>消费历史</t-breadcrumb-item>
    </t-breadcrumb>

    <!-- 消费统计卡片 -->
    <div class="consumption-stats">
      <t-row :gutter="[24, 24]">
        <t-col :lg="6" :md="12" :sm="24">
          <t-card class="stat-card">
            <div class="stat-item">
              <div class="stat-label">总消费金额</div>
              <div class="stat-value primary">
                ${{ totalConsumption.toFixed(4) }}
              </div>
            </div>
          </t-card>
        </t-col>

        <t-col :lg="6" :md="12" :sm="24">
          <t-card class="stat-card">
            <div class="stat-item">
              <div class="stat-label">今日消费</div>
              <div class="stat-value success">
                ${{ todayConsumption.toFixed(4) }}
              </div>
            </div>
          </t-card>
        </t-col>

        <t-col :lg="6" :md="12" :sm="24">
          <t-card class="stat-card">
            <div class="stat-item">
              <div class="stat-label">总请求次数</div>
              <div class="stat-value info">
                {{ totalRequests }}
              </div>
            </div>
          </t-card>
        </t-col>

        <t-col :lg="6" :md="12" :sm="24">
          <t-card class="stat-card">
            <div class="stat-item">
              <div class="stat-label">平均单次费用</div>
              <div class="stat-value warning">
                ${{ averageCost.toFixed(6) }}
              </div>
            </div>
          </t-card>
        </t-col>
      </t-row>
    </div>

    <!-- 消费记录列表 -->
    <t-card title="消费记录" class="consumption-table-card">
      <t-table
        :data="consumptionLogs"
        :columns="columns"
        :loading="loading"
        :pagination="pagination"
        row-key="id"
        stripe
        @page-change="(pageInfo: any, newDataSource: any) => onPageChange(pageInfo.current)"
        @page-size-change="onPageSizeChange"
      >
        <!-- 扣费类型列 -->
        <template #deduction_type="{ row }">
          <t-tag
            :theme="getDeductionTypeTheme(row.deduction_type)"
            variant="light"
          >
            {{ getDeductionTypeLabel(row.deduction_type) }}
          </t-tag>
        </template>

        <!-- 费用列 -->
        <template #cost_usd="{ row }">
          <span class="cost-amount">${{ row.cost_usd.toFixed(6) }}</span>
        </template>

        <!-- Token统计列 -->
        <template #tokens="{ row }">
          <div class="token-info">
            <div>输入: {{ row.input_tokens }}</div>
            <div>输出: {{ row.output_tokens }}</div>
            <div v-if="row.cache_read_tokens > 0">缓存读: {{ row.cache_read_tokens }}</div>
            <div v-if="row.cache_creation_tokens > 0">缓存写: {{ row.cache_creation_tokens }}</div>
            <div class="token-total">总计: {{ row.total_tokens }}</div>
          </div>
        </template>

        <!-- 模型列 -->
        <template #model="{ row }">
          <span class="model-name">{{ getModelDisplayName(row.model) }}</span>
        </template>

        <!-- 时间列 -->
        <template #created_at="{ row }">
          <span class="datetime">{{ formatDateTime(row.created_at) }}</span>
        </template>

        <!-- 操作列 -->
        <template #action="{ row }">
          <t-button theme="primary" variant="text" @click="showDetailDialog(row)">
            详情
          </t-button>
        </template>
      </t-table>
    </t-card>

    <!-- 详情对话框 -->
    <t-dialog
      v-model:visible="detailDialogVisible"
      title="消费记录详情"
      width="600px"
      :footer="false"
    >
      <div v-if="selectedLog" class="detail-content">
        <!-- 使用表格形式显示详情，更可靠 -->
        <table class="detail-table">
          <tbody>
            <tr v-for="item in detailData" :key="item.label">
              <td class="detail-label">{{ item.label }}</td>
              <td class="detail-value">{{ item.value }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else style="text-align: center; color: #999; padding: 20px;">
        selectedLog 为空
      </div>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { billingApi, type ConsumptionLog } from '@/api/billing';

// 响应式数据
const consumptionLogs = ref<ConsumptionLog[]>([]);
const loading = ref(false);
const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
  showJumper: true,
  showSizeChanger: true,
  pageSizeOptions: [10, 20, 50, 100]
});

const detailDialogVisible = ref(false);
const selectedLog = ref<ConsumptionLog | null>(null);

// 计算属性
const totalConsumption = computed(() => {
  return consumptionLogs.value.reduce((sum, log) => sum + log.cost_usd, 0);
});

const todayConsumption = computed(() => {
  const today = new Date().toDateString();
  return consumptionLogs.value
    .filter(log => new Date(log.created_at).toDateString() === today)
    .reduce((sum, log) => sum + log.cost_usd, 0);
});

const totalRequests = computed(() => consumptionLogs.value.length);

const averageCost = computed(() => {
  return totalRequests.value > 0 ? totalConsumption.value / totalRequests.value : 0;
});

const detailData = computed(() => {
  if (!selectedLog.value) return [];

  const log = selectedLog.value;
  return [
    { label: '记录ID', value: log.id },
    { label: '请求ID', value: log.request_id || '-' },
    { label: '扣费类型', value: getDeductionTypeLabel(log.deduction_type) },
    { label: '消费金额', value: `$${log.cost_usd?.toFixed(6) || '0.000000'}` },
    { label: '使用次数', value: log.usage_count || 0 },
    { label: '模型', value: getModelDisplayName(log.model) },
    { label: '输入Token', value: log.input_tokens || 0 },
    { label: '输出Token', value: log.output_tokens || 0 },
    { label: '缓存读Token', value: log.cache_read_tokens || 0 },
    { label: '缓存写Token', value: log.cache_creation_tokens || 0 },
    { label: '总Token', value: log.total_tokens || 0 },
    { label: '是否流式', value: log.is_stream ? '是' : '否' },
    { label: '平台类型', value: log.platform_type || '-' },
    { label: '扣费前余额', value: log.balance_before ? `$${log.balance_before.toFixed(4)}` : '-' },
    { label: '扣费后余额', value: log.balance_after ? `$${log.balance_after.toFixed(4)}` : '-' },
    { label: '创建时间', value: formatDateTime(log.created_at) }
  ];
});

// 表格列配置
const columns = [
  {
    colKey: 'id',
    title: 'ID',
    width: 80,
  },
  {
    colKey: 'deduction_type',
    title: '扣费类型',
    width: 120,
  },
  {
    colKey: 'cost_usd',
    title: '费用',
    width: 120,
  },
  {
    colKey: 'tokens',
    title: 'Token统计',
    width: 150,
  },
  {
    colKey: 'model',
    title: '模型',
    width: 200,
  },
  {
    colKey: 'created_at',
    title: '时间',
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
const loadConsumptionHistory = async () => {
  loading.value = true;
  try {
    const response = await billingApi.getConsumptionHistory({
      page: pagination.value.current,
      page_size: pagination.value.pageSize,
    });

    if (response.success) {
      consumptionLogs.value = response.data.logs || [];
      pagination.value.total = response.data.total || 0;
    } else {
      console.error('消费历史数据获取失败:', response);
    }
  } catch (error) {
    console.error('获取消费历史失败:', error);
    MessagePlugin.error('获取消费历史失败');
  } finally {
    loading.value = false;
  }
};

const onPageChange = (current: number) => {
  pagination.value.current = current;
  loadConsumptionHistory();
};

const onPageSizeChange = (pageSize: number) => {
  pagination.value.pageSize = pageSize;
  pagination.value.current = 1;
  loadConsumptionHistory();
};

const getDeductionTypeLabel = (type: string) => {
  const labels = {
    balance: '余额扣费',
    usage_count: '次数卡',
    time_limit: '时间卡'
  };
  return labels[type as keyof typeof labels] || type;
};

const getDeductionTypeTheme = (type: string): "default" | "primary" | "success" | "warning" | "danger" => {
  const themes: Record<string, "default" | "primary" | "success" | "warning" | "danger"> = {
    balance: 'primary',
    usage_count: 'warning',
    time_limit: 'success'
  };
  return themes[type] || 'default';
};

const getModelDisplayName = (model?: string) => {
  if (!model) return '-';

  // 动态格式化模型名称，不再使用硬编码映射
  // 将 "claude-3-5-sonnet-20241022" 转换为 "Claude 3.5 Sonnet 20241022"
  return model
    .split('-')
    .map((part, index) => {
      if (index === 0) return part.charAt(0).toUpperCase() + part.slice(1); // Claude
      if (/^\d/.test(part)) return part; // 数字部分保持不变
      return part.charAt(0).toUpperCase() + part.slice(1); // 首字母大写
    })
    .join(' ');
};

const formatDateTime = (dateStr: string) => {
  const date = new Date(dateStr);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });
};

const showDetailDialog = (log: ConsumptionLog) => {
  selectedLog.value = log;
  detailDialogVisible.value = true;
};

// 生命周期
onMounted(() => {
  loadConsumptionHistory();
});
</script>

<style scoped lang="less">
.consumption-history-page {
  padding: 24px;

  .breadcrumb {
    margin-bottom: 24px;
  }

  .consumption-stats {
    margin-bottom: 24px;

    .stat-card {
      .stat-item {
        .stat-label {
          font-size: 14px;
          color: var(--td-text-color-secondary);
          margin-bottom: 8px;
        }

        .stat-value {
          font-size: 24px;
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
  }

  .consumption-table-card {
    .cost-amount {
      font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
      font-weight: 500;
    }

    .token-info {
      font-size: 12px;
      line-height: 1.4;

      .token-total {
        font-weight: 500;
        color: var(--td-brand-color);
      }
    }

    .model-name {
      font-size: 12px;
      color: var(--td-text-color-secondary);
    }

    .datetime {
      font-size: 12px;
      color: var(--td-text-color-secondary);
    }
  }

  .detail-content {
    max-height: 500px;
    overflow-y: auto;

    .detail-table {
      width: 100%;
      border-collapse: collapse;
      border: 1px solid var(--td-border-level-1-color);

      tbody tr {
        &:nth-child(even) {
          background-color: var(--td-bg-color-container-active);
        }

        td {
          padding: 12px 16px;
          border-bottom: 1px solid var(--td-border-level-1-color);
          vertical-align: top;
        }

        .detail-label {
          background-color: var(--td-bg-color-component);
          font-weight: 500;
          width: 120px;
          color: var(--td-text-color-primary);
          border-right: 1px solid var(--td-border-level-1-color);
        }

        .detail-value {
          color: var(--td-text-color-secondary);
          font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
        }
      }

      tbody tr:last-child td {
        border-bottom: none;
      }
    }
  }
}
</style>
