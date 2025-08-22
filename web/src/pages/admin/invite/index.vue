<template>
  <div class="admin-invite-page">
    <!-- 页面标题 -->
    <t-breadcrumb class="breadcrumb">
      <t-breadcrumb-item>系统管理</t-breadcrumb-item>
      <t-breadcrumb-item>邀请管理</t-breadcrumb-item>
    </t-breadcrumb>

    <!-- 邀请统计概览 -->
    <div class="stats-overview">
      <t-row :gutter="[24, 24]">
        <t-col :lg="4" :md="6" :sm="12">
          <t-card class="stat-card" title="总邀请数">
            <div class="stat-number">{{ platformStats.total_invites || 0 }}</div>
          </t-card>
        </t-col>

        <t-col :lg="4" :md="6" :sm="12">
          <t-card class="stat-card" title="成功注册">
            <div class="stat-number">{{ platformStats.success_registers || 0 }}</div>
          </t-card>
        </t-col>

        <t-col :lg="4" :md="6" :sm="12">
          <t-card class="stat-card" title="转化率">
            <div class="stat-number">{{ platformStats.conversion_rate?.toFixed(1) || 0 }}%</div>
          </t-card>
        </t-col>

        <t-col :lg="4" :md="6" :sm="12">
          <t-card class="stat-card" title="总奖励支出">
            <div class="stat-number">${{ platformStats.total_rewards?.toFixed(2) || '0.00' }}</div>
          </t-card>
        </t-col>

        <t-col :lg="4" :md="6" :sm="12">
          <t-card class="stat-card" title="邀请用户占比">
            <div class="stat-number">
              {{
                platformStats.total_users > 0
                  ? ((platformStats.total_invites / platformStats.total_users) * 100).toFixed(1)
                  : 0
              }}%
            </div>
          </t-card>
        </t-col>

        <t-col :lg="4" :md="6" :sm="12">
          <t-card class="stat-card" title="总用户数">
            <div class="stat-number">{{ platformStats.total_users || 0 }}</div>
          </t-card>
        </t-col>
      </t-row>
    </div>

    <!-- 邀请配置 -->
    <t-card title="邀请系统配置" class="config-card">
      <t-form ref="configFormRef" :data="configForm" class="config-form" @submit="onConfigSubmit">
        <t-row :gutter="24">
          <t-col :span="8">
            <t-form-item label="邀请奖励金额(美元)" name="invite_reward_amount">
              <t-input
                v-model="configForm.invite_reward_amount"
                placeholder="请输入奖励金额"
                type="number"
                step="0.01"
                min="0"
              />
            </t-form-item>
          </t-col>

          <t-col :span="8">
            <t-form-item label="默认邀请限额" name="max_invite_count_default">
              <t-input
                v-model="configForm.max_invite_count_default"
                placeholder="请输入默认限额"
                type="number"
                min="0"
              />
            </t-form-item>
          </t-col>

          <t-col :span="8">
            <t-form-item label="邀请系统开关" name="invite_system_enabled">
              <t-select v-model="configForm.invite_system_enabled" placeholder="请选择">
                <t-option value="true" label="启用" />
                <t-option value="false" label="禁用" />
              </t-select>
            </t-form-item>
          </t-col>
        </t-row>

        <t-form-item>
          <t-button theme="primary" :loading="configLoading" @click="onConfigSubmit"> 保存配置 </t-button>
          <t-button variant="outline" @click="loadInviteConfig"> 重置 </t-button>
        </t-form-item>
      </t-form>
    </t-card>

    <!-- 用户邀请限额管理 -->
    <t-card title="用户邀请限额管理" class="user-limit-card">
      <div class="user-limit-header">
        <t-form :data="userLimitForm" layout="inline" @submit="onUserLimitSubmit">
          <t-form-item label="用户ID" name="user_id">
            <t-input v-model="userLimitForm.user_id" placeholder="请输入用户ID" type="number" min="1" />
          </t-form-item>
          <t-form-item label="邀请限额" name="max_count">
            <t-input v-model="userLimitForm.max_count" placeholder="请输入限额" type="number" min="0" />
          </t-form-item>
          <t-form-item>
            <t-button theme="primary" :loading="userLimitLoading" @click="onUserLimitSubmit"> 设置限额 </t-button>
          </t-form-item>
        </t-form>
      </div>
    </t-card>

    <!-- 邀请记录列表 -->
    <t-card title="邀请记录管理" class="records-card">
      <div class="records-header">
        <div class="search-area">
          <t-input
            v-model="searchForm.keyword"
            placeholder="搜索邀请人ID、邀请码或被邀请人邮箱"
            clearable
            @enter="handleSearch"
          >
            <template #suffix-icon>
              <t-icon name="search" />
            </template>
          </t-input>
          <t-select v-model="searchForm.status" placeholder="状态筛选" clearable>
            <t-option value="0" label="待注册" />
            <t-option value="1" label="已注册" />
          </t-select>
          <t-button @click="handleSearch">搜索</t-button>
          <t-button variant="outline" @click="resetSearch">重置</t-button>
        </div>
        <t-button variant="outline" @click="refreshData">
          <t-icon name="refresh" />
          刷新
        </t-button>
      </div>

      <t-table
        :data="inviteRecords"
        :columns="recordColumns"
        row-key="id"
        :hover="true"
        :pagination="pagination"
        :loading="recordsLoading"
        @page-change="handlePageChange"
      >
        <template #inviter_info="{ row }">
          <div>
            <div>ID: {{ row.inviter_id }}</div>
            <div class="text-secondary">码: {{ row.inviter_code }}</div>
          </div>
        </template>

        <template #invitee_info="{ row }">
          <div v-if="row.invitee_id">
            <div>ID: {{ row.invitee_id }}</div>
            <div class="text-secondary">{{ row.invitee?.email || row.invitee_email }}</div>
          </div>
          <div v-else class="pending-info">
            <div>待注册</div>
            <div class="text-secondary">{{ row.invitee_email || '-' }}</div>
          </div>
        </template>

        <template #status="{ row }">
                      <t-tag v-if="row.status === 1" theme="success" variant="light"> 已注册 </t-tag>
                      <t-tag v-else theme="warning" variant="light"> 待注册 </t-tag>
        </template>

        <template #reward_amount="{ row }">
                      <span v-if="row.status === 1" class="reward-amount"> +${{ row.reward_amount.toFixed(2) }} </span>
          <span v-else class="pending-reward">待发放</span>
        </template>

        <template #created_at="{ row }">
          {{ formatTime(row.created_at) }}
        </template>

        <template #registered_at="{ row }">
          {{ row.registered_at ? formatTime(row.registered_at) : '-' }}
        </template>
      </t-table>
    </t-card>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import {
  getInviteConfig,
  updateInviteConfig,
  getPlatformInviteStats,
  updateUserInviteLimit,
  getInviteList,
} from '@/api/invite';
import type { InviteConfig, PlatformInviteStats, Invitation, InviteListResponse } from '@/api/invite';

// 响应式数据
const platformStats = ref<PlatformInviteStats>({} as PlatformInviteStats);
const configForm = ref<InviteConfig>({
  invite_reward_amount: '',
  max_invite_count_default: '',
  invite_system_enabled: '',
});
const configFormRef = ref();
const configLoading = ref(false);

// 用户限额管理
const userLimitForm = ref({
  user_id: '',
  max_count: '',
});
const userLimitLoading = ref(false);

// 邀请记录
const inviteRecords = ref<Invitation[]>([]);
const recordsLoading = ref(false);
const searchForm = ref({
  keyword: '',
  status: '',
});

// 分页
const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
  showJumper: true,
  showSizer: true,
  pageSizeOptions: [10, 20, 50, 100],
});

// 表格列定义
const recordColumns = [
  {
    title: 'ID',
    colKey: 'id',
    width: 80,
  },
  {
    title: '邀请人信息',
    colKey: 'inviter_info',
    width: 150,
  },
  {
    title: '被邀请人信息',
    colKey: 'invitee_info',
    width: 200,
  },
  {
    title: '邀请令牌',
    colKey: 'invite_token',
    ellipsis: true,
    width: 200,
  },
  {
    title: '状态',
    colKey: 'status',
    width: 100,
  },
  {
    title: '奖励金额',
    colKey: 'reward_amount',
    width: 120,
  },
  {
    title: '创建时间',
    colKey: 'created_at',
    width: 160,
  },
  {
    title: '注册时间',
    colKey: 'registered_at',
    width: 160,
  },
];

// 方法
const loadPlatformStats = async () => {
  try {
    const response = await getPlatformInviteStats();
    if (response.success || response.code === 20000) {
      platformStats.value = response.data;
    }
  } catch (error) {
    console.error('获取平台统计失败:', error);
    MessagePlugin.error('获取平台统计失败');
  }
};

const loadInviteConfig = async () => {
  try {
    const response = await getInviteConfig();
    if (response.success || response.code === 20000) {
      configForm.value = response.data;
    }
  } catch (error) {
    console.error('获取邀请配置失败:', error);
    MessagePlugin.error('获取邀请配置失败');
  }
};

const onConfigSubmit = async () => {
  configLoading.value = true;
  try {
    const response = await updateInviteConfig(configForm.value);
    if (response.success) {
      MessagePlugin.success('配置保存成功');
      await loadPlatformStats(); // 刷新统计数据
    } else {
      MessagePlugin.error(response.message || '配置保存失败');
    }
  } catch (error: any) {
    console.error('保存配置失败:', error);
    MessagePlugin.error(error?.response?.data?.error || '配置保存失败');
  } finally {
    configLoading.value = false;
  }
};

const onUserLimitSubmit = async () => {
  if (!userLimitForm.value.user_id || !userLimitForm.value.max_count) {
    MessagePlugin.warning('请填写用户ID和邀请限额');
    return;
  }

  userLimitLoading.value = true;
  try {
    const data = {
      user_id: Number.parseInt(userLimitForm.value.user_id),
      max_count: Number.parseInt(userLimitForm.value.max_count),
    };
    const response = await updateUserInviteLimit(data.user_id, data.max_count);
    if (response.success) {
      MessagePlugin.success('用户邀请限额设置成功');
      userLimitForm.value = { user_id: '', max_count: '' };
    } else {
      MessagePlugin.error(response.message || '设置失败');
    }
  } catch (error: any) {
    console.error('设置用户限额失败:', error);
    MessagePlugin.error(error?.response?.data?.error || '设置用户限额失败');
  } finally {
    userLimitLoading.value = false;
  }
};

const loadInviteRecords = async () => {
  recordsLoading.value = true;
  try {
    const params = {
      page: pagination.value.current,
      limit: pagination.value.pageSize,
      ...searchForm.value,
    };
    const response = await getInviteList(params);
    if (response.success) {
      const data = response.data;
      inviteRecords.value = data.invitations;
      pagination.value.total = data.total;
    }
  } catch (error) {
    console.error('获取邀请记录失败:', error);
    MessagePlugin.error('获取邀请记录失败');
  } finally {
    recordsLoading.value = false;
  }
};

const handleSearch = () => {
  pagination.value.current = 1;
  loadInviteRecords();
};

const resetSearch = () => {
  searchForm.value = { keyword: '', status: '' };
  pagination.value.current = 1;
  loadInviteRecords();
};

const handlePageChange = (pageInfo: any) => {
  pagination.value.current = pageInfo.current;
  pagination.value.pageSize = pageInfo.pageSize;
  loadInviteRecords();
};

const refreshData = async () => {
  await Promise.all([loadPlatformStats(), loadInviteRecords()]);
};

const formatTime = (timeStr: string) => {
  if (!timeStr) return '-';
  return new Date(timeStr).toLocaleString();
};

// 生命周期
onMounted(async () => {
  await Promise.all([loadPlatformStats(), loadInviteConfig(), loadInviteRecords()]);
});
</script>

<style scoped lang="less">
.admin-invite-page {
  padding: 24px;

  .breadcrumb {
    margin-bottom: 24px;
  }

  .stats-overview {
    margin-bottom: 24px;

    .stat-card {
      .stat-number {
        font-size: 32px;
        font-weight: 600;
        color: var(--td-brand-color);
        margin-top: 8px;
      }
    }
  }

  .config-card {
    margin-bottom: 24px;

    .config-form {
      max-width: none;
    }
  }

  .user-limit-card {
    margin-bottom: 24px;

    .user-limit-header {
      .t-form {
        margin-bottom: 0;
      }
    }
  }

  .records-card {
    .records-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 16px;
      gap: 16px;

      .search-area {
        display: flex;
        gap: 12px;
        flex-wrap: wrap;
        align-items: center;
        flex: 1;

        .t-input {
          width: 300px;
        }

        .t-select {
          width: 120px;
        }
      }
    }

    .text-secondary {
      color: var(--td-text-color-secondary);
      font-size: 12px;
    }

    .pending-info {
      color: var(--td-text-color-placeholder);
    }

    .reward-amount {
      color: var(--td-success-color);
      font-weight: 500;
    }

    .pending-reward {
      color: var(--td-text-color-placeholder);
    }
  }
}
</style>
