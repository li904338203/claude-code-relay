<template>
  <div>
    <div class="info-banner">
      <div class="info-banner-icon">
        <t-icon name="info-circle-filled" />
      </div>
      <div class="info-banner-content">
        <div class="info-banner-title">管理员密钥管理</div>
        <div class="info-banner-text">
          管理所有用户的API密钥，可以按用户和分组筛选查看
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
          <t-select
            v-model="filterGroupId"
            placeholder="选择分组"
            clearable
            @change="handleGroupFilter"
            style="width: 200px; margin-right: 16px"
          >
            <t-option
              v-for="group in groupList"
              :key="group.id"
              :value="group.id"
              :label="group.name"
            />
          </t-select>
        </div>
        <div class="search-input">
          <t-input v-model="searchValue" placeholder="搜索密钥名称" clearable @enter="handleSearch">
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
        <template #user="{ row }">
          <t-tag theme="primary" variant="light">
            {{ getUserName(row.user_id) }}
          </t-tag>
        </template>

        <template #group="{ row }">
          <t-tag v-if="row.group" theme="success" variant="light">
            {{ row.group.name }}
          </t-tag>
          <span v-else>-</span>
        </template>

        <template #key="{ row }">
          <div class="key-display">
            <span class="key-prefix">{{ row.key.substring(0, 8) }}</span>
            <span class="key-mask">...</span>
            <span class="key-suffix">{{ row.key.substring(row.key.length - 8) }}</span>
            <t-button
              variant="text"
              size="small"
              @click="copyKey(row.key)"
              style="margin-left: 8px"
            >
              <t-icon name="copy" />
            </t-button>
          </div>
        </template>

        <template #status="{ row }">
          <t-tag :theme="row.status === 1 ? 'success' : 'danger'" variant="light">
            {{ row.status === 1 ? '启用' : '禁用' }}
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
          <span v-else>永久有效</span>
        </template>

        <template #today_usage="{ row }">
          <div class="usage-info">
            <div>请求: {{ row.today_usage_count }}</div>
            <div>Token: {{ formatTokens(row.today_input_tokens + row.today_output_tokens) }}</div>
            <div>费用: ${{ row.today_total_cost?.toFixed(4) || '0.0000' }}</div>
          </div>
        </template>

        <template #daily_limit="{ row }">
          <span v-if="row.daily_limit > 0">
            {{ row.daily_limit }}
            <div class="limit-progress">
              <t-progress
                :percentage="Math.min((row.today_usage_count / row.daily_limit) * 100, 100)"
                size="small"
                :theme="row.today_usage_count >= row.daily_limit ? 'danger' : 'primary'"
              />
            </div>
          </span>
          <span v-else>无限制</span>
        </template>
      </t-table>
    </t-card>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { SearchIcon } from 'tdesign-icons-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { getUsers } from '@/api/user';
import { getAdminApiKeys, getAdminGroupsAll } from '@/api/admin';

// 数据
const data = ref([]);
const userList = ref([]);
const groupList = ref([]);
const dataLoading = ref(false);
const searchValue = ref('');
const filterUserId = ref(null);
const filterGroupId = ref(null);

// 分页
const pagination = ref({
  current: 1,
  pageSize: 10,
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
  { colKey: 'name', title: '密钥名称', width: 150 },
  { colKey: 'user', title: '所属用户', width: 120 },
  { colKey: 'key', title: 'API Key', width: 200 },
  { colKey: 'group', title: '分组', width: 120 },
  { colKey: 'status', title: '状态', width: 80 },
  { colKey: 'expires_at', title: '过期时间', width: 180 },
  { colKey: 'daily_limit', title: '日限制', width: 120 },
  { colKey: 'today_usage', title: '今日使用', width: 150 },
  { colKey: 'last_used_time', title: '最后使用', width: 150 },
  { colKey: 'created_at', title: '创建时间', width: 150 },
];

// 工具函数
const getUserName = (userId) => {
  const user = userList.value.find(u => u.id === userId);
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

const isExpiringSoon = (expiresAt) => {
  if (!expiresAt) return false;
  const now = new Date();
  const expireDate = new Date(expiresAt);
  const timeLeft = expireDate - now;
  return timeLeft > 0 && timeLeft < 24 * 60 * 60 * 1000; // 24小时内过期
};

const copyKey = async (key) => {
  try {
    await navigator.clipboard.writeText(key);
    MessagePlugin.success('API Key已复制到剪贴板');
  } catch (error) {
    MessagePlugin.error('复制失败');
  }
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

    if (filterGroupId.value) {
      params.group_id = filterGroupId.value;
    }

    const result = await getAdminApiKeys(params);

    data.value = result.api_keys || [];
    pagination.value.total = result.total || 0;
  } catch (error) {
    console.error('获取API Key列表失败:', error);
    MessagePlugin.error('获取API Key列表失败: ' + error.message);
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

const fetchGroupList = async () => {
  try {
    const result = await getAdminGroupsAll();
    groupList.value = result || [];
  } catch (error) {
    console.error('获取分组列表失败:', error);
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

const handleGroupFilter = () => {
  pagination.value.current = 1;
  fetchData();
};

// 初始化
onMounted(() => {
  fetchData();
  fetchUserList();
  fetchGroupList();
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

.key-display {
  display: flex;
  align-items: center;
  font-family: monospace;
}

.key-prefix, .key-suffix {
  color: #333;
}

.key-mask {
  color: #999;
}

.usage-info {
  font-size: 12px;
  line-height: 1.4;
}

.usage-info div {
  margin: 2px 0;
}

.limit-progress {
  margin-top: 4px;
  width: 80px;
}
</style>
