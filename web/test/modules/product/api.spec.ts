import { beforeAll, describe, expect, it } from 'vitest'
import httpAdapter from 'axios/unsafe/adapters/http.js'
import { http } from '@/lib/http/client'
import {
  fetchProductList,
  fetchProductDetail,
  createProduct,
  updateProduct,
  deleteProduct,
  batchDeleteProducts,
  type ProductPayload
} from '@/modules/product/api'

beforeAll(() => {
  http.defaults.adapter = httpAdapter as never
})

const payload = (name: string): ProductPayload => ({
  name,
  category: '数码',
  price: 199.9,
  stock: 10,
  status: 'on'
})

describe('商品 API', () => {
  it('把后端数字 ID 统一转换为字符串', async () => {
    const result = await fetchProductList({ page: 1, size: 10, keyword: '', status: '' })
    expect(result.records.length).toBeGreaterThan(0)
    expect(typeof result.records[0].id).toBe('string')
  })

  it('跑通创建、详情、更新、筛选、批删和单删', async () => {
    const created = await createProduct(payload('接口测试商品'))
    expect(typeof created.id).toBe('string')

    const detail = await fetchProductDetail(created.id)
    expect(detail.name).toBe('接口测试商品')

    const updated = await updateProduct(created.id, {
      ...payload('接口测试商品改名'),
      price: 299.9,
      stock: 6,
      status: 'off'
    })
    expect(updated.name).toBe('接口测试商品改名')
    expect(updated.price).toBe(299.9)
    expect(updated.status).toBe('off')

    const filtered = await fetchProductList({
      page: 1,
      size: 10,
      keyword: '接口测试商品',
      status: 'off'
    })
    expect(filtered.total).toBe(1)
    expect(filtered.records[0].id).toBe(created.id)

    const first = await createProduct(payload('批删商品1'))
    const second = await createProduct(payload('批删商品2'))
    expect(await batchDeleteProducts([first.id, second.id])).toBe(true)
    await expect(fetchProductDetail(first.id)).rejects.toThrow()
    await expect(fetchProductDetail(second.id)).rejects.toThrow()

    expect(await deleteProduct(created.id)).toBe(true)
    await expect(fetchProductDetail(created.id)).rejects.toThrow()
  })
})
