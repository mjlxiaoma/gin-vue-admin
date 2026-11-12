import { api } from '@/lib/http/client'

export type ProductStatus = 'on' | 'off'

export interface ProductInfo {
  id: string
  name: string
  category: string
  price: number
  stock: number
  status: ProductStatus
  createTime: string
  updateTime: string
}

export interface ProductPayload {
  name: string
  category: string
  price: number
  stock: number
  status: ProductStatus
}

export interface ProductSearchRequest {
  page: number
  size: number
  keyword: string
  status: ProductStatus | ''
}

export interface ProductListResponse {
  records: ProductInfo[]
  total: number
  current: number
  size: number
}

// Go 的 uint ID 会序列化为数字；只在 HTTP 边界感知该 DTO，业务组件统一使用字符串 ID。
type ProductDTO = Omit<ProductInfo, 'id'> & { id: number }
type ProductListDTO = Omit<ProductListResponse, 'records'> & {
  records: ProductDTO[]
}

const toProductInfo = (record: ProductDTO): ProductInfo => ({
  ...record,
  id: String(record.id)
})

export const fetchProductList = async (
  params: ProductSearchRequest
): Promise<ProductListResponse> => {
  const result = await api.get<ProductListDTO>('/api/product', { params })
  return { ...result, records: result.records.map(toProductInfo) }
}

export const fetchProductDetail = async (id: string): Promise<ProductInfo> => {
  const result = await api.get<ProductDTO>(`/api/product/${id}`)
  return toProductInfo(result)
}

export const createProduct = async (
  data: ProductPayload
): Promise<ProductInfo> => {
  const result = await api.post<ProductDTO>('/api/product', data)
  return toProductInfo(result)
}

export const updateProduct = async (
  id: string,
  data: ProductPayload
): Promise<ProductInfo> => {
  const result = await api.put<ProductDTO>(`/api/product/${id}`, data)
  return toProductInfo(result)
}

export const deleteProduct = (id: string) =>
  api.del<boolean>(`/api/product/${id}`)

export const batchDeleteProducts = (ids: string[]) =>
  api.del<boolean>('/api/product', { data: { ids } })
