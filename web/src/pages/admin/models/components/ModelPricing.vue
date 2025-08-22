<template>
  <div>
    <!-- 添加新定价 -->
    <t-card class="mb-4">
      <template #header>
        <t-space justify="space-between">
          <span>{{ modelName }} - 定价配置</span>
          <t-button variant="base" @click="handleAddPricing">
            <template #icon><add-icon /></template>
            新增定价
          </t-button>
        </t-space>
      </template>

      <!-- 定价历史列表 -->
      <t-table
        :data="pricingData"
        :columns="pricingColumns"
        :loading="pricingLoading"
        row-key="id"
        hover
        size="medium"
      >
        <!-- 生效时间列 -->
        <template #effective_time="{ row }">
          {{ formatDateTime(row.effective_time) }}
        </template>

        <!-- 失效时间列 -->
        <template #expire_time="{ row }">
          {{ row.expire_time ? formatDateTime(row.expire_time) : '永久有效' }}
        </template>

        <!-- 状态列 -->
        <template #status="{ row }">
          <t-tag v-if="row.status === 1" theme="success" variant="light">启用</t-tag>
          <t-tag v-else theme="danger" variant="light">禁用</t-tag>
        </template>

        <!-- 输入价格列 -->
        <template #input_price="{ row }">
          ${{ row.input_price.toFixed(6) }}
        </template>

        <!-- 输出价格列 -->
        <template #output_price="{ row }">
          ${{ row.output_price.toFixed(6) }}
        </template>

        <!-- 缓存写入价格列 -->
        <template #cache_write_price="{ row }">
          ${{ row.cache_write_price.toFixed(6) }}
        </template>

        <!-- 缓存读取价格列 -->
        <template #cache_read_price="{ row }">
          ${{ row.cache_read_price.toFixed(6) }}
        </template>

        <!-- 操作列 -->
        <template #operation="{ row }">
          <t-space>
            <t-link theme="primary" @click="handleEditPricing(row)">编辑</t-link>
            <t-popconfirm
              content="确认删除这个定价配置吗？"
              @confirm="handleDeletePricing(row)"
            >
              <t-link theme="danger">删除</t-link>
            </t-popconfirm>
          </t-space>
        </template>
      </t-table>
    </t-card>

    <!-- 定价表单对话框 -->
    <t-dialog
      v-model:visible="pricingFormVisible"
      :header="editingPricing ? '编辑定价配置' : '新增定价配置'"
      width="700px"
      :on-confirm="handlePricingFormConfirm"
      :on-cancel="handlePricingFormCancel"
      :confirm-btn="{ loading: pricingSubmitLoading }"
    >
      <t-form ref="pricingFormRef" :model="pricingFormData" label-align="top" :rules="pricingRules">
        <t-row :gutter="16">
          <t-col :span="6">
            <t-form-item label="输入价格 (USD/1M tokens)" name="input_price">
              <t-input-number
                v-model="pricingFormData.input_price"
                :min="0"
                :decimal-places="6"
                placeholder="0.000001"
              />
            </t-form-item>
          </t-col>
          <t-col :span="6">
            <t-form-item label="输出价格 (USD/1M tokens)" name="output_price">
              <t-input-number
                v-model="pricingFormData.output_price"
                :min="0"
                :decimal-places="6"
                placeholder="0.000001"
              />
            </t-form-item>
          </t-col>
        </t-row>

        <t-row :gutter="16">
          <t-col :span="6">
            <t-form-item label="缓存写入价格 (USD/1M tokens)" name="cache_write_price">
              <t-input-number
                v-model="pricingFormData.cache_write_price"
                :min="0"
                :decimal-places="6"
                placeholder="0.000001"
              />
            </t-form-item>
          </t-col>
          <t-col :span="6">
            <t-form-item label="缓存读取价格 (USD/1M tokens)" name="cache_read_price">
              <t-input-number
                v-model="pricingFormData.cache_read_price"
                :min="0"
                :decimal-places="6"
                placeholder="0.000001"
              />
            </t-form-item>
          </t-col>
        </t-row>

        <t-row :gutter="16">
          <t-col :span="6">
            <t-form-item label="生效时间" name="effective_time">
              <t-date-picker
                v-model="pricingFormData.effective_time"
                format="YYYY-MM-DD HH:mm:ss"
                enable-time-picker
                placeholder="选择生效时间"
                style="width: 100%"
              />
            </t-form-item>
          </t-col>
          <t-col :span="6">
            <t-form-item label="失效时间" name="expire_time">
              <t-date-picker
                v-model="pricingFormData.expire_time"
                format="YYYY-MM-DD HH:mm:ss"
                enable-time-picker
                placeholder="选择失效时间（可选）"
                clearable
                style="width: 100%"
              />
              <template #help>留空表示永久有效</template>
            </t-form-item>
          </t-col>
        </t-row>
      </t-form>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, defineProps, watch } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { AddIcon } from 'tdesign-icons-vue-next';
import { getModelPricingHistory, createModelPricing, updateModelPricing, deleteModelPricing } from '@/api/models';

// 引入类型定义
import type { ModelPricing } from '@/api/models';

// Props
const props = defineProps<{
  modelId: number;
  modelName: string;
}>();

// 数据状态
const pricingData = ref<ModelPricing[]>([]);
const pricingLoading = ref(false);

// 定价表单状态
const pricingFormVisible = ref(false);
const pricingSubmitLoading = ref(false);
const editingPricing = ref(null);
const pricingFormRef = ref();

// 定价表单数据
const pricingFormData = reactive({
  input_price: 0,
  output_price: 0,
  cache_write_price: 0,
  cache_read_price: 0,
  effective_time: '',
  expire_time: '',
});

// 定价表单验证规则
const pricingRules = {
  input_price: [{ required: true, message: '输入价格必填' }],
  output_price: [{ required: true, message: '输出价格必填' }],
  cache_write_price: [{ required: true, message: '缓存写入价格必填' }],
  cache_read_price: [{ required: true, message: '缓存读取价格必填' }],
  effective_time: [{ required: true, message: '生效时间必填' }],
};

// 定价表格列定义
const pricingColumns = [
  {
    title: '输入价格',
    colKey: 'input_price',
    width: 120,
  },
  {
    title: '输出价格',
    colKey: 'output_price',
    width: 120,
  },
  {
    title: '缓存写入',
    colKey: 'cache_write_price',
    width: 120,
  },
  {
    title: '缓存读取',
    colKey: 'cache_read_price',
    width: 120,
  },
  {
    title: '生效时间',
    colKey: 'effective_time',
    width: 180,
  },
  {
    title: '失效时间',
    colKey: 'expire_time',
    width: 180,
  },
  {
    title: '状态',
    colKey: 'status',
    width: 80,
  },
  {
    title: '操作',
    colKey: 'operation',
    width: 120,
    fixed: 'right' as const,
  },
];

// 方法
const fetchPricingData = async () => {
  if (!props.modelId) return;

  pricingLoading.value = true;
  try {
    const result = await getModelPricingHistory(props.modelId);
    // API返回格式：{ success: true, data: ModelPricing[] }
    pricingData.value = (result as any).data || [];
  } catch (error) {
    console.error('获取定价历史失败:', error);
    MessagePlugin.error('获取定价历史失败');
  } finally {
    pricingLoading.value = false;
  }
};

const handleAddPricing = () => {
  editingPricing.value = null;
  resetPricingFormData();
  // 设置默认生效时间为当前时间
  pricingFormData.effective_time = new Date().toISOString().slice(0, 19).replace('T', ' ');
  pricingFormVisible.value = true;
};

const handleEditPricing = (row: any) => {
  editingPricing.value = row;
  Object.assign(pricingFormData, {
    input_price: row.input_price,
    output_price: row.output_price,
    cache_write_price: row.cache_write_price,
    cache_read_price: row.cache_read_price,
    effective_time: row.effective_time,
    expire_time: row.expire_time || '',
  });
  pricingFormVisible.value = true;
};

const handleDeletePricing = async (row: ModelPricing) => {
  try {
    await deleteModelPricing(row.id);
    MessagePlugin.success('定价配置删除成功');
    fetchPricingData();
  } catch (error) {
    console.error('删除失败:', error);
    MessagePlugin.error('删除失败');
  }
};

const handlePricingFormConfirm = async () => {
  const valid = await pricingFormRef.value?.validate();
  if (!valid) return false;

  pricingSubmitLoading.value = true;
  try {
    const requestData = {
      input_price: pricingFormData.input_price,
      output_price: pricingFormData.output_price,
      cache_write_price: pricingFormData.cache_write_price,
      cache_read_price: pricingFormData.cache_read_price,
      effective_time: pricingFormData.effective_time,
      expire_time: pricingFormData.expire_time || undefined,
    };

    if (editingPricing.value) {
      await updateModelPricing(editingPricing.value.id, requestData);
      MessagePlugin.success('更新成功');
    } else {
      await createModelPricing(props.modelId, requestData);
      MessagePlugin.success('创建成功');
    }

    pricingFormVisible.value = false;
    fetchPricingData();
    return true;
  } catch (error) {
    console.error('操作失败:', error);
    MessagePlugin.error('操作失败');
    return false;
  } finally {
    pricingSubmitLoading.value = false;
  }
};

const handlePricingFormCancel = () => {
  pricingFormVisible.value = false;
  resetPricingFormData();
};

const resetPricingFormData = () => {
  Object.assign(pricingFormData, {
    input_price: 0,
    output_price: 0,
    cache_write_price: 0,
    cache_read_price: 0,
    effective_time: '',
    expire_time: '',
  });
};

const formatDateTime = (dateStr: any) => {
  if (!dateStr) return '';
  return new Date(dateStr).toLocaleString('zh-CN');
};

// 监听模型ID变化
watch(() => props.modelId, (newId) => {
  if (newId) {
    fetchPricingData();
  }
}, { immediate: true });

// 生命周期
onMounted(() => {
  if (props.modelId) {
    fetchPricingData();
  }
});
</script>

<style scoped>
.mb-4 {
  margin-bottom: 16px;
}
</style>
