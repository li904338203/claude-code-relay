<template>
  <div>
    <div class="info-banner">
      <div class="info-banner-icon">
        <t-icon name="info-circle-filled" />
      </div>
      <div class="info-banner-content">
        <div class="info-banner-title">管理员分组管理</div>
        <div class="info-banner-text">
          管理所有用户的分组，查看分组统计信息
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
          <t-input v-model="searchValue" placeholder="搜索分组名称" clearable @enter="handleSearch">
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
            {{ row.user?.username || '未知用户' }}
          </t-tag>
        </template>

        <template #status="{ row }">
          <t-tag :theme="row.status === 1 ? 'success' : 'danger'" variant="light">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </t-tag>
        </template>

        <template #api_key_count="{ row }">
          <t-tag theme="primary" variant="outline">
            {{ row.api_key_count }} 个密钥
          </t-tag>
        </template>

        <template #account_count="{ row }">
          <t-tag theme="success" variant="outline">
            {{ row.account_count }} 个账号
          </t-tag>
        </template>

        <template #remark="{ row }">
          <div class="remark-cell">
            {{ row.remark || '-' }}
          </div>
        </template>

        <template #created_at="{ row }">
          {{ formatTime(row.created_at) }}
        </template>

        <template #updated_at="{ row }">
          {{ formatTime(row.updated_at) }}
        </template>

        <template #op="{ row }">
          <t-space>
            <t-button variant="text" size="small" @click="viewDetails(row)">
              <t-icon name="view" />
              查看详情
            </t-button>
          </t-space>
        </template>
      </t-table>
    </t-card>

    <!-- 详情对话框 -->
    <t-dialog
      v-model:visible="detailDialogVisible"
      :header="dialogHeader"
      width="800px"
      @close="handleDetailClose"
    >
      <div v-if="selectedGroup" class="detail-content">
        <t-descriptions :data="groupDetails" />

        <t-divider>分组统计</t-divider>
        <t-row :gutter="16">
          <t-col :span="6">
            <t-card size="small" :bordered="false" class="stat-card">
              <div class="stat-number">{{ selectedGroup.api_key_count }}</div>
              <div class="stat-label">API密钥</div>
            </t-card>
          </t-col>
          <t-col :span="6">
            <t-card size="small" :bordered="false" class="stat-card">
              <div class="stat-number">{{ selectedGroup.account_count }}</div>
              <div class="stat-label">账号数量</div>
            </t-card>
          </t-col>
        </t-row>
      </div>
    </t-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { SearchIcon } from 'tdesign-icons-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { getUsers } from '@/api/user';
import { getAdminGroups } from '@/api/admin';

// 数据
const data = ref([]);
const userList = ref([]);
const dataLoading = ref(false);
const searchValue = ref('');
const filterUserId = ref(null);

// 详情对话框
const detailDialogVisible = ref(false);
const selectedGroup = ref(null);

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
  { colKey: 'name', title: '分组名称', width: 150 },
  { colKey: 'user', title: '所属用户', width: 120 },
  { colKey: 'status', title: '状态', width: 80 },
  { colKey: 'api_key_count', title: 'API密钥', width: 100 },
  { colKey: 'account_count', title: '账号数量', width: 100 },
  { colKey: 'remark', title: '备注', width: 200 },
  { colKey: 'created_at', title: '创建时间', width: 150 },
  { colKey: 'updated_at', title: '更新时间', width: 150 },
  { colKey: 'op', title: '操作', width: 120, fixed: 'right' },
];

// 计算属性
const dialogHeader = computed(() => {
  return selectedGroup.value ? `分组详情 - ${selectedGroup.value.name}` : '分组详情';
});

const groupDetails = computed(() => {
  if (!selectedGroup.value) return [];

  return [
    { label: '分组ID', content: selectedGroup.value.id },
    { label: '分组名称', content: selectedGroup.value.name },
    { label: '所属用户', content: selectedGroup.value.user?.username || '未知用户' },
    { label: '状态', content: selectedGroup.value.status === 1 ? '启用' : '禁用' },
    { label: '备注', content: selectedGroup.value.remark || '无' },
    { label: '创建时间', content: formatTime(selectedGroup.value.created_at) },
    { label: '更新时间', content: formatTime(selectedGroup.value.updated_at) },
  ];
});

// 工具函数
const formatTime = (timeStr) => {
  if (!timeStr) return '-';
  return new Date(timeStr).toLocaleString();
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

    const result = await getAdminGroups(params);

    data.value = result.groups || [];
    pagination.value.total = result.total || 0;
  } catch (error) {
    console.error('获取分组列表失败:', error);
    MessagePlugin.error('获取分组列表失败: ' + error.message);
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

const viewDetails = (row) => {
  selectedGroup.value = row;
  detailDialogVisible.value = true;
};

const handleDetailClose = () => {
  selectedGroup.value = null;
  detailDialogVisible.value = false;
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

.remark-cell {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.detail-content {
  padding: 16px 0;
}

.stat-card {
  text-align: center;
  background: #f8f9fa;
}

.stat-number {
  font-size: 24px;
  font-weight: 600;
  color: #1976d2;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: #666;
}
</style>
