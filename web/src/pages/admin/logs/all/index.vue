<template>
  <div>
    <div class="info-banner">
      <div class="info-banner-icon">
        <t-icon name="info-circle-filled" />
      </div>
      <div class="info-banner-content">
        <div class="info-banner-title">管理员日志管理</div>
        <div class="info-banner-text">
          查看所有用户的API调用日志，支持按用户、模型、时间等条件筛选
        </div>
      </div>
    </div>

    <t-card class="list-card-container" :bordered="false">
      <!-- 搜索和筛选区域 -->
      <t-row justify="space-between" class="search-row">
        <div class="left-operation-container">
          <t-space>
            <t-button variant="outline" @click="handleRefresh">
              <template #icon><refresh-icon /></template>
              刷新
            </t-button>
            <t-button variant="outline" @click="resetFilters"> 重置筛选 </t-button>
            <t-button variant="outline" @click="toggleAdvancedFilter">
              {{ showAdvancedFilter ? '收起筛选' : '展开筛选' }}
            </t-button>
          </t-space>
        </div>
        <div class="right-operation-container">
          <t-space>
            <t-select
              v-model="searchFilters.user_id"
              placeholder="选择用户"
              clearable
              style="width: 180px"
              @change="handleSearch"
            >
              <t-option v-for="user in userOptions" :key="user.id" :value="user.id" :label="user.username" />
            </t-select>

            <t-select
              v-model="searchFilters.model_name"
              placeholder="选择模型"
              clearable
              style="width: 200px"
              @change="handleSearch"
            >
              <t-option v-for="model in modelOptions" :key="model" :value="model" :label="model" />
            </t-select>

            <t-date-range-picker
              v-model="dateRange"
              format="YYYY-MM-DD HH:mm:ss"
              placeholder="选择时间范围"
              clearable
              @change="handleDateRangeChange"
            />
          </t-space>
        </div>
      </t-row>

      <!-- 高级筛选 -->
      <t-row v-if="showAdvancedFilter" class="advanced-filter">
        <t-space>
          <t-input-number
            v-model="searchFilters.min_cost"
            placeholder="最小费用"
            :min="0"
            :step="0.001"
            :decimal-places="4"
            style="width: 150px"
            @blur="handleSearch"
          />
          <span>-</span>
          <t-input-number
            v-model="searchFilters.max_cost"
            placeholder="最大费用"
            :min="0"
            :step="0.001"
            :decimal-places="4"
            style="width: 150px"
            @blur="handleSearch"
          />

          <t-select
            v-model="searchFilters.is_stream"
            placeholder="请求类型"
            clearable
            style="width: 120px"
            @change="handleSearch"
          >
            <t-option :value="true" label="流式" />
            <t-option :value="false" label="非流式" />
          </t-select>
        </t-space>
      </t-row>

      <!-- 统计信息 -->
      <t-row class="stats-row" v-if="statsData">
        <t-col :span="6">
          <t-card size="small" :bordered="false" class="stat-card">
            <div class="stat-number">{{ statsData.total_requests || 0 }}</div>
            <div class="stat-label">总请求数</div>
          </t-card>
        </t-col>
        <t-col :span="6">
          <t-card size="small" :bordered="false" class="stat-card">
            <div class="stat-number">{{ formatTokens(statsData.total_tokens || 0) }}</div>
            <div class="stat-label">总Token数</div>
          </t-card>
        </t-col>
        <t-col :span="6">
          <t-card size="small" :bordered="false" class="stat-card">
            <div class="stat-number">${{ (statsData.total_cost || 0).toFixed(4) }}</div>
            <div class="stat-label">总费用</div>
          </t-card>
        </t-col>
        <t-col :span="6">
          <t-card size="small" :bordered="false" class="stat-card">
            <div class="stat-number">{{ (statsData.avg_duration || 0).toFixed(2) }}s</div>
            <div class="stat-label">平均耗时</div>
          </t-card>
        </t-col>
      </t-row>

      <!-- 日志表格 -->
      <t-table
        :data="data"
        :columns="COLUMNS"
        :row-key="rowKey"
        :hover="true"
        :pagination="pagination"
        :loading="dataLoading"
        :header-affixed-top="headerAffixedTop"
        @page-change="handlePageChange"
      >
        <template #user="{ row }">
          <t-tag theme="primary" variant="light">
            {{ getUserName(row.user_id) }}
          </t-tag>
        </template>

        <template #model_name="{ row }">
          <t-tag theme="success" variant="light">
            {{ row.model_name }}
          </t-tag>
        </template>

        <template #api_key="{ row }">
          <t-tag v-if="row.api_key" theme="warning" variant="light">
            {{ row.api_key.name }}
          </t-tag>
          <span v-else>-</span>
        </template>

        <template #tokens="{ row }">
          <div class="token-info">
            <div class="token-row">
              <span class="token-label">输入:</span>
              <span class="token-value">{{ formatTokens(row.input_tokens) }}</span>
            </div>
            <div class="token-row">
              <span class="token-label">输出:</span>
              <span class="token-value">{{ formatTokens(row.output_tokens) }}</span>
            </div>
            <div v-if="row.cache_read_input_tokens > 0" class="token-row cache">
              <span class="token-label">缓存读:</span>
              <span class="token-value">{{ formatTokens(row.cache_read_input_tokens) }}</span>
            </div>
            <div v-if="row.cache_creation_input_tokens > 0" class="token-row cache">
              <span class="token-label">缓存写:</span>
              <span class="token-value">{{ formatTokens(row.cache_creation_input_tokens) }}</span>
            </div>
          </div>
        </template>

        <template #cost="{ row }">
          <div class="cost-info">
            <div class="cost-total">${{ row.total_cost.toFixed(4) }}</div>
            <div class="cost-breakdown">
              <span>输入: ${{ row.input_cost.toFixed(4) }}</span><br>
              <span>输出: ${{ row.output_cost.toFixed(4) }}</span>
              <span v-if="row.cache_read_cost > 0"><br>缓存读: ${{ row.cache_read_cost.toFixed(4) }}</span>
              <span v-if="row.cache_write_cost > 0"><br>缓存写: ${{ row.cache_write_cost.toFixed(4) }}</span>
            </div>
          </div>
        </template>

        <template #is_stream="{ row }">
          <t-tag :theme="row.is_stream ? 'primary' : 'default'" variant="light">
            {{ row.is_stream ? '流式' : '非流式' }}
          </t-tag>
        </template>

        <template #duration="{ row }">
          <span class="duration">{{ row.duration.toFixed(2) }}s</span>
        </template>

        <template #created_at="{ row }">
          {{ formatTime(row.created_at) }}
        </template>

        <template #op="{ row }">
          <t-button variant="text" size="small" @click="viewDetail(row)">
            详情
          </t-button>
        </template>
      </t-table>
    </t-card>

    <!-- 详情对话框 -->
    <t-dialog
      v-model:visible="detailDialogVisible"
      header="日志详情"
      width="800px"
      @close="handleDetailClose"
    >
      <div v-if="selectedLog" class="detail-content">
        <t-descriptions :data="logDetails" />
      </div>
    </t-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { RefreshIcon } from 'tdesign-icons-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { getUsers } from '@/api/user';
import { getAdminLogs, getAdminLogsStats } from '@/api/admin';

// 数据
const data = ref([]);
const userOptions = ref([]);
const modelOptions = ref([]);
const statsData = ref(null);
const dataLoading = ref(false);
const showAdvancedFilter = ref(false);
const dateRange = ref([]);

// 搜索筛选条件
const searchFilters = ref({
  user_id: null,
  model_name: '',
  is_stream: null,
  min_cost: null,
  max_cost: null,
  start_time: '',
  end_time: '',
});

// 详情对话框
const detailDialogVisible = ref(false);
const selectedLog = ref(null);

// 分页
const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
  showJumper: true,
  onChange: (pageInfo) => {
    pagination.value.current = pageInfo.current;
    pagination.value.pageSize = pageInfo.pageSize;
    fetchData();
  },
});

// 表格配置
const rowKey = 'id';
const headerAffixedTop = ref(false);

// 表格列配置
const COLUMNS = [
  { colKey: 'user', title: '用户', width: 100 },
  { colKey: 'model_name', title: '模型', width: 120 },
  { colKey: 'api_key', title: 'API Key', width: 120 },
  { colKey: 'tokens', title: 'Token统计', width: 160 },
  { colKey: 'cost', title: '费用详情', width: 140 },
  { colKey: 'is_stream', title: '类型', width: 80 },
  { colKey: 'duration', title: '耗时', width: 80 },
  { colKey: 'created_at', title: '时间', width: 150 },
  { colKey: 'op', title: '操作', width: 80, fixed: 'right' },
];

// 计算属性
const logDetails = computed(() => {
  if (!selectedLog.value) return [];

  const log = selectedLog.value;
  return [
    { label: '日志ID', content: log.id },
    { label: '用户', content: getUserName(log.user_id) },
    { label: '模型名称', content: log.model_name },
    { label: 'API Key', content: log.api_key?.name || '-' },
    { label: '输入Token', content: log.input_tokens },
    { label: '输出Token', content: log.output_tokens },
    { label: '缓存读Token', content: log.cache_read_input_tokens },
    { label: '缓存写Token', content: log.cache_creation_input_tokens },
    { label: '输入费用', content: `$${log.input_cost.toFixed(4)}` },
    { label: '输出费用', content: `$${log.output_cost.toFixed(4)}` },
    { label: '缓存读费用', content: `$${log.cache_read_cost.toFixed(4)}` },
    { label: '缓存写费用', content: `$${log.cache_write_cost.toFixed(4)}` },
    { label: '总费用', content: `$${log.total_cost.toFixed(4)}` },
    { label: '请求类型', content: log.is_stream ? '流式' : '非流式' },
    { label: '耗时', content: `${log.duration.toFixed(2)}秒` },
    { label: '创建时间', content: formatTime(log.created_at) },
  ];
});

// 工具函数
const getUserName = (userId) => {
  const user = userOptions.value.find(u => u.id === userId);
  return user?.username || '未知用户';
};

const formatTime = (timeStr) => {
  if (!timeStr) return '-';
  return new Date(timeStr).toLocaleString();
};

const formatTokens = (tokens) => {
  if (tokens >= 1000000) {
    return (tokens / 1000000).toFixed(1) + 'M';
  } else if (tokens >= 1000) {
    return (tokens / 1000).toFixed(1) + 'K';
  }
  return tokens?.toString() || '0';
};

// 数据获取
const fetchData = async () => {
  dataLoading.value = true;
  try {
    const params = {
      page: pagination.value.current,
      limit: pagination.value.pageSize,
      ...searchFilters.value,
    };

    // 移除空值
    Object.keys(params).forEach(key => {
      if (params[key] === null || params[key] === '') {
        delete params[key];
      }
    });

    const result = await getAdminLogs(params);

    data.value = result.logs || [];
    pagination.value.total = result.total || 0;

    // 更新模型选项
    const models = [...new Set(data.value.map(log => log.model_name))];
    modelOptions.value = models;
  } catch (error) {
    console.error('获取日志列表失败:', error);
    MessagePlugin.error('获取日志列表失败: ' + error.message);
  } finally {
    dataLoading.value = false;
  }
};

const fetchUserList = async () => {
  try {
    const result = await getUsers({ page: 1, limit: 1000 });
    userOptions.value = result.users || [];
  } catch (error) {
    console.error('获取用户列表失败:', error);
  }
};

const fetchStats = async () => {
  try {
    const params = { ...searchFilters.value };
    Object.keys(params).forEach(key => {
      if (params[key] === null || params[key] === '') {
        delete params[key];
      }
    });

    const result = await getAdminLogsStats(params);

    statsData.value = result;
  } catch (error) {
    console.error('获取统计数据失败:', error);
  }
};

// 事件处理
const handlePageChange = () => {
  fetchData();
};

const handleRefresh = () => {
  fetchData();
  fetchStats();
};

const handleSearch = () => {
  pagination.value.current = 1;
  fetchData();
  fetchStats();
};

const handleDateRangeChange = () => {
  if (dateRange.value && dateRange.value.length === 2) {
    searchFilters.value.start_time = dateRange.value[0];
    searchFilters.value.end_time = dateRange.value[1];
  } else {
    searchFilters.value.start_time = '';
    searchFilters.value.end_time = '';
  }
  handleSearch();
};

const resetFilters = () => {
  searchFilters.value = {
    user_id: null,
    model_name: '',
    is_stream: null,
    min_cost: null,
    max_cost: null,
    start_time: '',
    end_time: '',
  };
  dateRange.value = [];
  handleSearch();
};

const toggleAdvancedFilter = () => {
  showAdvancedFilter.value = !showAdvancedFilter.value;
};

const viewDetail = (row) => {
  selectedLog.value = row;
  detailDialogVisible.value = true;
};

const handleDetailClose = () => {
  selectedLog.value = null;
  detailDialogVisible.value = false;
};

// 初始化
onMounted(() => {
  fetchUserList();
  fetchData();
  fetchStats();
});
</script>

<style scoped>
.info-banner {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  color: white;
}

.info-banner-icon {
  font-size: 24px;
  margin-right: 12px;
}

.info-banner-content {
  flex: 1;
}

.info-banner-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 4px;
}

.info-banner-text {
  font-size: 14px;
  opacity: 0.9;
}

.list-card-container {
  padding: 24px;
}

.search-row {
  margin-bottom: 16px;
}

.advanced-filter {
  margin-bottom: 16px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 6px;
}

.stats-row {
  margin-bottom: 16px;
}

.stat-card {
  text-align: center;
  background: #f8f9fa;
}

.stat-number {
  font-size: 20px;
  font-weight: 600;
  color: #1976d2;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: #666;
}

.token-info {
  font-size: 12px;
  line-height: 1.4;
}

.token-row {
  display: flex;
  justify-content: space-between;
  margin: 2px 0;
}

.token-row.cache {
  color: #666;
}

.token-label {
  color: #999;
}

.token-value {
  font-weight: 500;
}

.cost-info {
  font-size: 12px;
  line-height: 1.4;
}

.cost-total {
  font-weight: 600;
  color: #e53e3e;
  margin-bottom: 4px;
}

.cost-breakdown {
  color: #666;
}

.duration {
  color: #4a90e2;
  font-weight: 500;
}

.detail-content {
  padding: 16px 0;
}
</style>
