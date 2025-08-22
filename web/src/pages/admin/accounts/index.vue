<template>
  <div>
    <div class="info-banner">
      <div class="info-banner-icon">
        <t-icon name="info-circle-filled" />
      </div>
      <div class="info-banner-content">
        <div class="info-banner-title">管理员账号管理</div>
        <div class="info-banner-text">
          管理所有用户的Claude账号，可以按用户筛选查看
        </div>
      </div>
    </div>

    <t-card class="list-card-container" :bordered="false">
      <t-row justify="space-between">
        <div class="left-operation-container">
          <t-select
            v-model="filterUserId"
            placeholder="选择用户"
            clearable
            @change="handleUserFilter"
            style="width: 200px; margin-right: 16px"
          >
            <t-option
              v-for="user in userList"
              :key="user.id"
              :value="user.id"
              :label="user.username"
            />
          </t-select>
        </div>
        <div class="search-input">
          <t-input v-model="searchValue" placeholder="搜索账号名称" clearable @enter="handleSearch">
            <template #suffix-icon>
              <search-icon size="16px" />
            </template>
          </t-input>
        </div>
      </t-row>

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
        <template #platform_type="{ row }">
          <t-tag :theme="getPlatformTypeTheme(row.platform_type)" variant="light">
            {{ getPlatformTypeName(row.platform_type) }}
          </t-tag>
        </template>

        <template #user="{ row }">
          <t-tag theme="primary" variant="light">
            {{ row.user?.username || '未知用户' }}
          </t-tag>
        </template>

        <template #group="{ row }">
          <t-tag v-if="row.group" theme="success" variant="light">
            {{ row.group.name }}
          </t-tag>
          <span v-else>-</span>
        </template>

        <template #current_status="{ row }">
          <t-tag :theme="getStatusTheme(row.current_status)" variant="light">
            {{ getStatusText(row.current_status) }}
          </t-tag>
        </template>

        <template #active_status="{ row }">
          <t-tag :theme="row.active_status === 1 ? 'success' : 'danger'" variant="light">
            {{ row.active_status === 1 ? '激活' : '禁用' }}
          </t-tag>
        </template>

        <template #expires_at="{ row }">
          <span v-if="row.expires_at">
            {{ formatTime(row.expires_at) }}
            <t-tag
              v-if="isExpiringSoon(row.expires_at)"
              theme="warning"
              variant="light"
              size="small"
              style="margin-left: 8px"
            >
              即将过期
            </t-tag>
          </span>
          <span v-else>-</span>
        </template>

        <template #today_usage="{ row }">
          <div class="usage-info">
            <div>请求: {{ row.today_usage_count }}</div>
            <div>Token: {{ formatTokens(row.today_input_tokens + row.today_output_tokens) }}</div>
            <div>费用: ${{ row.today_total_cost?.toFixed(4) || '0.0000' }}</div>
          </div>
        </template>
      </t-table>
    </t-card>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { SearchIcon } from 'tdesign-icons-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { getAdminAccounts } from '@/api/admin';
import { getUsers } from '@/api/user';

// 数据
const data = ref([]);
const userList = ref([]);
const dataLoading = ref(false);
const searchValue = ref('');
const filterUserId = ref(null);

// 分页
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showJumper: true,
});

// 表格配置
const rowKey = 'id';
const headerAffixedTop = ref(false);

// 表格列配置
const COLUMNS = [
  { colKey: 'name', title: '账号名称', width: 150 },
  { colKey: 'platform_type', title: '平台类型', width: 120 },
  { colKey: 'user', title: '所属用户', width: 120 },
  { colKey: 'group', title: '分组', width: 120 },
  { colKey: 'priority', title: '优先级', width: 80 },
  { colKey: 'weight', title: '权重', width: 80 },
  { colKey: 'current_status', title: '当前状态', width: 100 },
  { colKey: 'active_status', title: '激活状态', width: 100 },
  { colKey: 'expires_at', title: '过期时间', width: 180 },
  { colKey: 'today_usage', title: '今日使用', width: 150 },
  { colKey: 'last_used_time', title: '最后使用', width: 150 },
];

// 平台类型相关
const getPlatformTypeName = (type) => {
  const typeMap = {
    'claude': 'Claude API',
    'claude_console': 'Claude Console',
    'openai': 'OpenAI',
    'gemini': 'Gemini'
  };
  return typeMap[type] || type;
};

const getPlatformTypeTheme = (type) => {
  const themeMap = {
    'claude': 'primary',
    'claude_console': 'success',
    'openai': 'warning',
    'gemini': 'danger'
  };
  return themeMap[type] || 'default';
};

// 状态相关
const getStatusTheme = (status) => {
  const themeMap = {
    1: 'success',
    2: 'warning',
    3: 'danger'
  };
  return themeMap[status] || 'default';
};

const getStatusText = (status) => {
  const textMap = {
    1: '正常',
    2: '接口异常',
    3: '账号异常'
  };
  return textMap[status] || '未知';
};

// 工具函数
const formatTime = (timestamp) => {
  if (!timestamp) return '-';
  return new Date(timestamp * 1000).toLocaleString();
};

const formatTokens = (tokens) => {
  if (tokens >= 1000000) {
    return (tokens / 1000000).toFixed(1) + 'M';
  } else if (tokens >= 1000) {
    return (tokens / 1000).toFixed(1) + 'K';
  }
  return tokens?.toString() || '0';
};

const isExpiringSoon = (expiresAt) => {
  if (!expiresAt) return false;
  const now = Math.floor(Date.now() / 1000);
  const timeLeft = expiresAt - now;
  return timeLeft > 0 && timeLeft < 24 * 60 * 60; // 24小时内过期
};

// 数据获取
const fetchData = async () => {
  dataLoading.value = true;
  try {
    const params = {
      page: pagination.value.current,
      limit: pagination.value.pageSize,
    };

    if (filterUserId.value) {
      params.user_id = filterUserId.value;
    }

    if (searchValue.value) {
      params.name = searchValue.value;
    }

    const result = await getAdminAccounts(params);
    data.value = result.accounts || [];
    pagination.value.total = result.total || 0;
  } catch (error) {
    console.error('获取账号列表失败:', error);
    MessagePlugin.error('获取账号列表失败: ' + error.message);
  } finally {
    dataLoading.value = false;
  }
};

const fetchUserList = async () => {
  try {
    const result = await getUsers({ page: 1, limit: 1000 });
    userList.value = result.users || [];
  } catch (error) {
    console.error('获取用户列表失败:', error);
  }
};

// 事件处理
const handlePageChange = () => {
  fetchData();
};

const handleSearch = () => {
  pagination.value.current = 1;
  fetchData();
};

const handleUserFilter = () => {
  pagination.value.current = 1;
  fetchData();
};

// 初始化
onMounted(() => {
  fetchData();
  fetchUserList();
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

.left-operation-container {
  display: flex;
  align-items: center;
}

.search-input {
  width: 300px;
}

.usage-info {
  font-size: 12px;
  line-height: 1.4;
}

.usage-info div {
  margin: 2px 0;
}
</style>
