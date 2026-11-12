<template>
  <FormDrawer
    v-model="visible"
    :title="drawerTitle"
    :mode="mode"
    :form-data="formData"
    :fields="fields"
    :rules="rules"
    :loading="submitting"
    width="600px"
    @submit="handleSubmit"
  />
</template>

<script lang="ts" setup>
import { ref, watch, reactive, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { FormDrawer } from '@/app/components'
import type { FormField, FormDrawerMode } from '@/app/components/FormDrawer/types'
import { t } from '@/lib/i18n'
import {
  createProduct,
  updateProduct,
  type ProductInfo,
  type ProductPayload
} from '../api'

const props = defineProps<{
  modelValue: boolean
  mode: FormDrawerMode
  data: ProductInfo | null
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  success: []
}>()

const visible = ref(props.modelValue)
const submitting = ref(false)

watch(
  () => props.modelValue,
  (value) => {
    visible.value = value
    if (value) initForm()
  }
)
watch(visible, (value) => emit('update:modelValue', value))

const formData = reactive<ProductPayload>({
  name: '',
  category: '',
  price: 0,
  stock: 0,
  status: 'on'
})

const drawerTitle = computed(() => {
  if (props.mode === 'add') return t('product.createTitle')
  if (props.mode === 'edit') return t('product.editTitle')
  return t('product.viewTitle')
})

const fields = computed<FormField[]>(() => [
  { prop: 'name', label: t('product.name'), type: 'input', span: 12 },
  {
    prop: 'category',
    label: t('product.category'),
    type: 'select',
    span: 12,
    options: [
      { label: t('product.categoryDigital'), value: '数码' },
      { label: t('product.categoryHome'), value: '家居' },
      { label: t('product.categoryBook'), value: '图书' },
      { label: t('product.categoryFood'), value: '食品' },
      { label: t('product.categoryOther'), value: '其他' }
    ]
  },
  { prop: 'price', label: t('product.price'), type: 'number', span: 12 },
  { prop: 'stock', label: t('product.stock'), type: 'number', span: 12 },
  {
    prop: 'status',
    label: t('product.status'),
    type: 'select',
    span: 24,
    options: [
      { label: t('product.statusOn'), value: 'on' },
      { label: t('product.statusOff'), value: 'off' }
    ]
  }
])

const rules = computed(() => ({
  name: [
    { required: true, message: t('product.nameRequired'), trigger: 'blur' },
    { max: 100, message: t('product.nameMax'), trigger: 'blur' }
  ],
  category: [
    { required: true, message: t('product.categoryRequired'), trigger: 'change' }
  ],
  price: [
    { required: true, message: t('product.priceRequired'), trigger: 'blur' },
    { type: 'number', min: 0, message: t('product.priceMin'), trigger: 'blur' }
  ],
  stock: [
    { required: true, message: t('product.stockRequired'), trigger: 'blur' },
    {
      validator: (_rule: unknown, value: unknown, callback: (error?: Error) => void) => {
        const stock = Number(value)
        if (!Number.isInteger(stock) || stock < 0) {
          callback(new Error(t('product.stockInteger')))
          return
        }
        callback()
      },
      trigger: 'blur'
    }
  ],
  status: [
    { required: true, message: t('product.statusRequired'), trigger: 'change' }
  ]
}))

const initForm = () => {
  if ((props.mode === 'edit' || props.mode === 'view') && props.data) {
    Object.assign(formData, {
      name: props.data.name,
      category: props.data.category,
      price: props.data.price,
      stock: props.data.stock,
      status: props.data.status
    })
    return
  }
  Object.assign(formData, {
    name: '',
    category: '',
    price: 0,
    stock: 0,
    status: 'on'
  })
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    const payload: ProductPayload = {
      name: formData.name.trim(),
      category: formData.category,
      price: Number(formData.price),
      stock: Number(formData.stock),
      status: formData.status
    }
    if (props.mode === 'add') {
      await createProduct(payload)
      ElMessage.success(t('common.message.createSuccess'))
    } else {
      await updateProduct(props.data!.id, payload)
      ElMessage.success(t('common.message.updateSuccess'))
    }
    emit('success')
    emit('update:modelValue', false)
  } finally {
    submitting.value = false
  }
}
</script>
