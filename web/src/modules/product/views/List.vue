<template>
  <PageContainer>
    <SearchTable
      :title="t('product.title')"
      :loading="loading"
      :data="tableData"
      :columns="columns"
      :pagination="pagination"
      :selected-rows="tableSelectedRows"
      selectable
      row-key="id"
      column-settings-key="product-list"
      @search="handleSearch"
      @reset="handleReset"
      @page-change="handlePageChange"
      @selection-change="onSelectionChange"
    >
      <template #search>
        <el-input
          v-model="searchForm.keyword"
          :placeholder="t('product.name')"
          clearable
          style="width: 220px"
          @keyup.enter="handleSearch"
        />
        <el-select
          v-model="searchForm.status"
          clearable
          :placeholder="t('product.status')"
          style="width: 140px"
        >
          <el-option
            :label="t('product.statusOn')"
            value="on"
          />
          <el-option
            :label="t('product.statusOff')"
            value="off"
          />
        </el-select>
      </template>

      <template #actions>
        <el-button
          v-permission="Permissions.PRODUCT_CREATE"
          type="primary"
          :icon="Plus"
          @click="openDrawer('add')"
        >
          {{ t('product.createAction') }}
        </el-button>
        <el-button
          v-permission="Permissions.PRODUCT_DELETE"
          type="danger"
          :icon="Delete"
          :disabled="selectedRows.length === 0"
          @click="handleBatchDelete"
        >
          {{ t('common.action.batchDelete') }}
        </el-button>
        <el-button
          :icon="Refresh"
          @click="fetchList"
        >
          {{ t('common.action.refresh') }}
        </el-button>
      </template>

      <template #col-price="{ row }">
        ¥{{ Number(row.price).toFixed(2) }}
      </template>

      <template #col-status="{ row }">
        <el-tag
          :type="row.status === 'on' ? 'success' : 'info'"
          size="small"
        >
          {{ row.status === 'on' ? t('product.statusOn') : t('product.statusOff') }}
        </el-tag>
      </template>

      <template #col-createTime="{ row }">
        {{ formatDate(row.createTime) }}
      </template>

      <template #col-actions="{ row }">
        <el-button
          link
          type="primary"
          size="small"
          @click="openDrawer('view', row)"
        >
          {{ t('common.action.view') }}
        </el-button>
        <el-button
          v-permission="Permissions.PRODUCT_EDIT"
          link
          type="primary"
          size="small"
          @click="openDrawer('edit', row)"
        >
          {{ t('common.action.edit') }}
        </el-button>
        <el-button
          v-permission="Permissions.PRODUCT_DELETE"
          link
          type="danger"
          size="small"
          @click="handleDelete(row.id)"
        >
          {{ t('common.action.delete') }}
        </el-button>
      </template>
    </SearchTable>

    <ProductFormDrawer
      v-model="drawerVisible"
      :mode="drawerMode"
      :data="editingRow"
      @success="onFormSuccess"
    />
  </PageContainer>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue'
import { Plus, Delete, Refresh } from '@element-plus/icons-vue'
import { PageContainer, SearchTable } from '@/app/components'
import { useCrud } from '@/app/composables/useCrud'
import { Permissions } from '@/app/constants/permissions'
import { t } from '@/lib/i18n'
import { formatDate } from '@/lib/format'
import type { ColumnDef } from '@/app/components/SearchTable/types'
import {
  fetchProductList,
  deleteProduct,
  batchDeleteProducts,
  type ProductInfo,
  type ProductSearchRequest
} from '../api'
import ProductFormDrawer from './ProductFormDrawer.vue'

const {
  listData,
  loading,
  pagination,
  searchForm,
  selectedRows,
  fetchList,
  handleSearch,
  handleReset,
  handlePageChange,
  handleSelectionChange,
  handleDelete,
  handleBatchDelete
} = useCrud<ProductInfo>({
  fetch: (params) => fetchProductList(params as unknown as ProductSearchRequest),
  remove: deleteProduct,
  batchRemove: batchDeleteProducts,
  defaultSearchForm: { keyword: '', status: '' },
  pageSize: 10
})

const onSelectionChange = (rows: Record<string, unknown>[]) => {
  handleSelectionChange(rows as unknown as ProductInfo[])
}

const tableData = computed(
  () => listData.value as unknown as Record<string, unknown>[]
)
const tableSelectedRows = computed(
  () => selectedRows.value as unknown as Record<string, unknown>[]
)

const columns = computed<ColumnDef[]>(() => [
  { prop: 'name', label: t('product.name'), minWidth: 160 },
  { prop: 'category', label: t('product.category'), minWidth: 110 },
  { prop: 'price', label: t('product.price'), width: 120, align: 'right', slot: 'price' },
  { prop: 'stock', label: t('product.stock'), width: 100, align: 'right' },
  { prop: 'status', label: t('product.status'), width: 100, slot: 'status' },
  { prop: 'createTime', label: t('product.createTime'), width: 170, slot: 'createTime' },
  {
    prop: 'actions',
    label: t('product.actions'),
    width: 200,
    fixed: 'right',
    slot: 'actions',
    hideable: false
  }
])

const drawerVisible = ref(false)
const drawerMode = ref<'add' | 'edit' | 'view'>('add')
const editingRow = ref<ProductInfo | null>(null)

const openDrawer = (mode: 'add' | 'edit' | 'view', row?: ProductInfo) => {
  drawerMode.value = mode
  editingRow.value = row ?? null
  drawerVisible.value = true
}

const onFormSuccess = () => {
  drawerVisible.value = false
  fetchList()
}

onMounted(fetchList)
</script>
