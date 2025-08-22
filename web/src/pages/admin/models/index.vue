<template>
  <div>
    <t-card class="list-card-container" :bordered="false">
      <!-- 搜索和操作区域 -->
      <t-row justify="space-between" class="search-row">
        <div class="left-operation-container">
          <t-space>
            <t-button variant="base" @click="handleAdd">
              <template #icon><add-icon /></template>
              新增模型
            </t-button>
            <t-button variant="outline" @click="handleRefresh">
              <template #icon><refresh-icon /></template>
              刷新
            </t-button>
            <t-button variant="outline" @click="handleRefreshCache">
              <template #icon><swap-icon /></template>
              刷新缓存
            </t-button>
          </t-space>
        </div>
        <div class="right-operation-container">
          <t-space>
            <t-input
              v-model="searchFilters.name"
              placeholder="搜索模型名称"
              clearable
              @clear="handleSearch"
              @enter="handleSearch"
            >
              <template #suffix-icon>
                <search-icon @click="handleSearch" />
              </template>
            </t-input>
            <t-select
              v-model="searchFilters.provider"
              placeholder="选择提供商"
              clearable
              style="width: 150px"
              @change="handleSearch"
            >
              <t-option value="claude" label="Claude" />
              <t-option value="openai" label="OpenAI" />
              <t-option value="gemini" label="Gemini" />
            </t-select>
            <t-select
              v-model="searchFilters.status"
              placeholder="选择状态"
              clearable
              style="width: 120px"
              @change="handleSearch"
            >
              <t-option :value="1" label="启用" />
              <t-option :value="0" label="禁用" />
            </t-select>
          </t-space>
        </div>
      </t-row>

      <!-- 数据表格 -->
      <t-table
        :data="data"
        :columns="columns"
        :loading="dataLoading"
        :pagination="pagination"
        :selected-row-keys="selectedRowKeys"
        row-key="id"
        hover
        @page-change="handlePageChange"
        @select-change="handleSelectChange"
      >
        <!-- 状态列 -->
        <template #status="{ row }">
          <t-tag v-if="row.status === 1" theme="success" variant="light">启用</t-tag>
          <t-tag v-else theme="danger" variant="light">禁用</t-tag>
        </template>

        <!-- 操作列 -->
        <template #operation="{ row }">
          <t-space>
            <t-link theme="primary" @click="handleView(row)">查看</t-link>
            <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
            <t-link theme="primary" @click="handlePricing(row)">定价</t-link>
            <t-popconfirm
              content="确认删除这个模型配置吗？"
              @confirm="handleDelete(row)"
            >
              <t-link theme="danger">删除</t-link>
            </t-popconfirm>
          </t-space>
        </template>
      </t-table>

      <!-- 批量操作 -->
      <t-space v-if="selectedRowKeys.length > 0" class="batch-operations">
        <t-button variant="outline" @click="handleBatchEnable">批量启用</t-button>
        <t-button variant="outline" @click="handleBatchDisable">批量禁用</t-button>
      </t-space>
    </t-card>

    <!-- 模型表单对话框 -->
    <t-dialog
      v-model:visible="formVisible"
      :header="editingItem ? '编辑模型配置' : '新增模型配置'"
      width="800px"
      :on-confirm="handleFormConfirm"
      :on-cancel="handleFormCancel"
      :confirm-btn="{ loading: submitLoading }"
    >
      <t-form ref="formRef" :model="formData" label-align="top" :rules="rules">
        <t-row :gutter="16">
          <t-col :span="6">
            <t-form-item label="模型名称" name="name">
              <t-input v-model="formData.name" :disabled="!!editingItem" placeholder="如：claude-3-5-sonnet-20241022" />
              <template #help>
                <span v-if="editingItem">模型名称创建后不可修改</span>
                <span v-else>用于API调用的唯一标识符</span>
              </template>
            </t-form-item>
          </t-col>
          <t-col :span="6">
            <t-form-item label="显示名称" name="display_name">
              <t-input v-model="formData.display_name" placeholder="如：Claude 3.5 Sonnet" />
            </t-form-item>
          </t-col>
          <t-col :span="4">
            <t-form-item label="提供商" name="provider">
              <t-select v-model="formData.provider" placeholder="选择提供商">
                <t-option value="claude" label="Claude" />
                <t-option value="openai" label="OpenAI" />
                <t-option value="gemini" label="Gemini" />
              </t-select>
            </t-form-item>
          </t-col>
        </t-row>

        <t-row :gutter="16">
          <t-col :span="6">
            <t-form-item label="模型类别" name="category">
              <t-input v-model="formData.category" placeholder="如：sonnet, haiku, opus" />
            </t-form-item>
          </t-col>
          <t-col :span="6">
            <t-form-item label="版本号" name="version">
              <t-input v-model="formData.version" placeholder="如：20241022" />
            </t-form-item>
          </t-col>
          <t-col :span="4">
            <t-form-item label="状态" name="status">
              <t-radio-group v-model="formData.status">
                <t-radio :value="1">启用</t-radio>
                <t-radio :value="0">禁用</t-radio>
              </t-radio-group>
            </t-form-item>
          </t-col>
        </t-row>

        <t-row :gutter="16">
          <t-col :span="4">
            <t-form-item label="排序权重" name="sort_order">
              <t-input-number v-model="formData.sort_order" :min="0" :max="999" placeholder="数字越小越靠前" />
            </t-form-item>
          </t-col>
          <t-col :span="4">
            <t-form-item label="最大Token" name="max_tokens">
              <t-input-number v-model="formData.max_tokens" :min="1" placeholder="可选" />
            </t-form-item>
          </t-col>
          <t-col :span="4">
            <t-form-item label="上下文窗口" name="context_window">
              <t-input-number v-model="formData.context_window" :min="1" placeholder="可选" />
            </t-form-item>
          </t-col>
        </t-row>

        <t-form-item label="模型描述" name="description">
          <t-textarea v-model="formData.description" placeholder="模型的详细描述" :rows="3" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 定价管理对话框 -->
    <t-dialog
      v-model:visible="pricingVisible"
      header="模型定价管理"
      width="1000px"
      :footer="false"
    >
      <ModelPricing v-if="pricingVisible" :model-id="currentModelId" :model-name="currentModelName" />
    </t-dialog>

    <!-- 模型详情对话框 -->
    <t-dialog
      v-model:visible="detailVisible"
      header="模型配置详情"
      width="600px"
      :footer="false"
    >
      <t-descriptions v-if="detailData" :column="2" bordered>
        <t-descriptions-item label="模型名称">{{ detailData.name }}</t-descriptions-item>
        <t-descriptions-item label="显示名称">{{ detailData.display_name }}</t-descriptions-item>
        <t-descriptions-item label="提供商">{{ detailData.provider }}</t-descriptions-item>
        <t-descriptions-item label="类别">{{ detailData.category }}</t-descriptions-item>
        <t-descriptions-item label="版本">{{ detailData.version }}</t-descriptions-item>
        <t-descriptions-item label="状态">
          <t-tag v-if="detailData.status === 1" theme="success" variant="light">启用</t-tag>
          <t-tag v-else theme="danger" variant="light">禁用</t-tag>
        </t-descriptions-item>
        <t-descriptions-item label="排序权重">{{ detailData.sort_order }}</t-descriptions-item>
        <t-descriptions-item label="最大Token">{{ detailData.max_tokens || '无限制' }}</t-descriptions-item>
        <t-descriptions-item label="上下文窗口">{{ detailData.context_window || '未设置' }}</t-descriptions-item>
        <t-descriptions-item label="创建时间">{{ formatDateTime(detailData.created_at) }}</t-descriptions-item>
        <t-descriptions-item label="更新时间">{{ formatDateTime(detailData.updated_at) }}</t-descriptions-item>
        <t-descriptions-item label="描述" :span="2">{{ detailData.description || '无' }}</t-descriptions-item>
      </t-descriptions>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import {
  AddIcon,
  RefreshIcon,
  SearchIcon,
  SwapIcon,
} from 'tdesign-icons-vue-next';

// 引入类型定义和API函数
import type { ModelConfig } from '@/api/models';
import { getModelList, updateModelStatus, deleteModel, createModel, updateModel, refreshPricingCache } from '@/api/models';

// 引入定价管理组件
import ModelPricing from './components/ModelPricing.vue';

// 数据状态
const data = ref<ModelConfig[]>([]);
const dataLoading = ref(false);
const selectedRowKeys = ref<number[]>([]);

// 分页
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
});

// 搜索过滤
const searchFilters = reactive({
  name: '',
  provider: '',
  status: null,
});

// 表单状态
const formVisible = ref(false);
const submitLoading = ref(false);
const editingItem = ref(null);
const formRef = ref();

// 表单数据
const formData = reactive({
  name: '',
  display_name: '',
  provider: '',
  category: '',
  version: '',
  status: 1,
  sort_order: 0,
  description: '',
  max_tokens: null,
  context_window: null,
});

// 定价对话框
const pricingVisible = ref(false);
const currentModelId = ref(null);
const currentModelName = ref('');

// 详情对话框
const detailVisible = ref(false);
const detailData = ref(null);

// 表单验证规则
const rules = {
  name: [
    { required: true, message: '模型名称必填' },
    { min: 1, max: 100, message: '模型名称长度为 1-100 个字符' },
  ],
  display_name: [
    { required: true, message: '显示名称必填' },
    { min: 1, max: 100, message: '显示名称长度为 1-100 个字符' },
  ],
  provider: [{ required: true, message: '请选择提供商' }],
  category: [{ required: true, message: '模型类别必填' }],
  version: [{ required: true, message: '版本号必填' }],
};

// 表格列定义
const columns = [
  {
    colKey: 'row-select',
    type: 'multiple' as const,
    width: 64,
    fixed: 'left' as const,
  },
  {
    title: '模型名称',
    colKey: 'name',
    width: 200,
    fixed: 'left' as const,
    ellipsis: true,
  },
  {
    title: '显示名称',
    colKey: 'display_name',
    width: 150,
    ellipsis: true,
  },
  {
    title: '提供商',
    colKey: 'provider',
    width: 100,
  },
  {
    title: '类别',
    colKey: 'category',
    width: 100,
  },
  {
    title: '版本',
    colKey: 'version',
    width: 120,
  },
  {
    title: '状态',
    colKey: 'status',
    width: 80,
  },
  {
    title: '排序',
    colKey: 'sort_order',
    width: 80,
  },
  {
    title: '最大Token',
    colKey: 'max_tokens',
    width: 100,
  },
  {
    title: '创建时间',
    colKey: 'created_at',
    width: 180,
    cell: (_: any, { row }: any) => formatDateTime(row.created_at),
  },
  {
    title: '操作',
    colKey: 'operation',
    width: 200,
    fixed: 'right' as const,
  },
];

// 方法
const fetchData = async () => {
  dataLoading.value = true;
  try {
    const params = {
      page: pagination.current,
      limit: pagination.pageSize,
      ...(searchFilters.name && { name: searchFilters.name }),
      ...(searchFilters.provider && { provider: searchFilters.provider }),
      ...(searchFilters.status !== null && searchFilters.status !== undefined && { status: searchFilters.status }),
    };

    const result = await getModelList(params);
    // API返回格式：{ success: true, data: ModelConfig[], total: number }
    data.value = (result as any).data || [];
    pagination.total = result.total || 0;
  } catch (error) {
    console.error('获取模型列表失败:', error);
    MessagePlugin.error('获取模型列表失败');
  } finally {
    dataLoading.value = false;
  }
};

const handleSearch = () => {
  pagination.current = 1;
  fetchData();
};

const handleRefresh = () => {
  fetchData();
};

const handleRefreshCache = async () => {
  try {
    await refreshPricingCache();
    MessagePlugin.success('缓存刷新成功');
  } catch (error) {
    console.error('缓存刷新失败:', error);
    MessagePlugin.error('缓存刷新失败');
  }
};

const handlePageChange = (pageInfo: any) => {
  pagination.current = pageInfo.current;
  pagination.pageSize = pageInfo.pageSize;
  fetchData();
};

const handleSelectChange = (keys: any) => {
  selectedRowKeys.value = keys;
};

const handleAdd = () => {
  editingItem.value = null;
  resetFormData();
  formVisible.value = true;
};

const handleEdit = (row: any) => {
  editingItem.value = row;
  Object.assign(formData, row);
  formVisible.value = true;
};

const handleView = (row: any) => {
  detailData.value = row;
  detailVisible.value = true;
};

const handlePricing = (row: any) => {
  currentModelId.value = row.id;
  currentModelName.value = row.name;
  pricingVisible.value = true;
};

const handleDelete = async (row: ModelConfig) => {
  try {
    await deleteModel(row.id);
    MessagePlugin.success('删除成功');
    fetchData();
  } catch (error) {
    console.error('删除失败:', error);
    MessagePlugin.error('删除失败');
  }
};

const handleBatchEnable = () => {
  handleBatchUpdateStatus(1);
};

const handleBatchDisable = () => {
  handleBatchUpdateStatus(0);
};

const handleBatchUpdateStatus = async (status: number) => {
  try {
    await updateModelStatus({
      ids: selectedRowKeys.value,
      status,
    });
    const statusText = status === 1 ? '启用' : '禁用';
    MessagePlugin.success(`批量${statusText}成功`);
    selectedRowKeys.value = [];
    fetchData();
  } catch (error) {
    console.error('批量更新失败:', error);
    MessagePlugin.error('批量更新失败');
  }
};

const handleFormConfirm = async () => {
  const valid = await formRef.value?.validate();
  if (!valid) return false;

  submitLoading.value = true;
  try {
    if (editingItem.value) {
      // 更新模型
      const updateData = {
        display_name: formData.display_name,
        provider: formData.provider,
        category: formData.category,
        version: formData.version,
        status: formData.status,
        sort_order: formData.sort_order,
        description: formData.description,
        max_tokens: formData.max_tokens,
        context_window: formData.context_window,
      };
      await updateModel(editingItem.value.id, updateData);
      MessagePlugin.success('更新成功');
    } else {
      // 创建模型
      await createModel(formData);
      MessagePlugin.success('创建成功');
    }

    formVisible.value = false;
    fetchData();
    return true;
  } catch (error) {
    console.error('操作失败:', error);
    MessagePlugin.error('操作失败');
    return false;
  } finally {
    submitLoading.value = false;
  }
};

const handleFormCancel = () => {
  formVisible.value = false;
  resetFormData();
};

const resetFormData = () => {
  Object.assign(formData, {
    name: '',
    display_name: '',
    provider: '',
    category: '',
    version: '',
    status: 1,
    sort_order: 0,
    description: '',
    max_tokens: null,
    context_window: null,
  });
};

const formatDateTime = (dateStr: any) => {
  if (!dateStr) return '';
  return new Date(dateStr).toLocaleString('zh-CN');
};

// 生命周期
onMounted(() => {
  fetchData();
});
</script>

<style scoped>
.search-row {
  margin-bottom: 20px;
}

.batch-operations {
  margin-top: 16px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 4px;
}
</style>
