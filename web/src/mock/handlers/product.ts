import { http } from 'msw'
import { ok, fail } from './_utils'

type ProductStatus = 'on' | 'off'

interface ProductRecord {
  id: number
  name: string
  category: string
  price: number
  stock: number
  status: ProductStatus
  createTime: string
  updateTime: string
}

type ProductPayload = Omit<ProductRecord, 'id' | 'createTime' | 'updateTime'>

let nextId = 5
let products: ProductRecord[] = [
  {
    id: 1,
    name: '机械键盘',
    category: '数码',
    price: 299.9,
    stock: 20,
    status: 'on',
    createTime: '2026-11-01T09:00:00+08:00',
    updateTime: '2026-11-01T09:00:00+08:00'
  },
  {
    id: 2,
    name: '人体工学椅',
    category: '家居',
    price: 899,
    stock: 8,
    status: 'on',
    createTime: '2026-11-02T09:00:00+08:00',
    updateTime: '2026-11-02T09:00:00+08:00'
  },
  {
    id: 3,
    name: 'Go 语言实战',
    category: '图书',
    price: 79.8,
    stock: 35,
    status: 'off',
    createTime: '2026-11-03T09:00:00+08:00',
    updateTime: '2026-11-03T09:00:00+08:00'
  },
  {
    id: 4,
    name: '咖啡豆',
    category: '食品',
    price: 69,
    stock: 0,
    status: 'off',
    createTime: '2026-11-04T09:00:00+08:00',
    updateTime: '2026-11-04T09:00:00+08:00'
  }
]

const validPayload = (body: Partial<ProductPayload>) =>
  typeof body.name === 'string' &&
  body.name.trim() !== '' &&
  typeof body.category === 'string' &&
  body.category !== '' &&
  Number.isFinite(Number(body.price)) &&
  Number(body.price) >= 0 &&
  Number.isInteger(Number(body.stock)) &&
  Number(body.stock) >= 0 &&
  (body.status === 'on' || body.status === 'off')

export const productHandlers = [
  http.get('/api/product', ({ request }) => {
    const searchParams = new URL(request.url).searchParams
    const page = Math.max(1, Number(searchParams.get('page')) || 1)
    const size = Math.min(100, Math.max(1, Number(searchParams.get('size')) || 10))
    const keyword = searchParams.get('keyword')?.trim().toLowerCase() ?? ''
    const status = searchParams.get('status') ?? ''

    const filtered = products.filter(
      (product) =>
        (!keyword || product.name.toLowerCase().includes(keyword)) &&
        (!status || product.status === status)
    )
    const start = (page - 1) * size
    return ok({
      records: filtered.slice(start, start + size),
      total: filtered.length,
      current: page,
      size
    })
  }),

  http.get('/api/product/:id', ({ params }) => {
    const product = products.find((item) => String(item.id) === params.id)
    return product ? ok(product) : fail(404, '商品不存在')
  }),

  http.post('/api/product', async ({ request }) => {
    const body = (await request.json()) as Partial<ProductPayload>
    if (!validPayload(body)) return fail(422, '商品参数不正确')

    const now = new Date().toISOString()
    const product: ProductRecord = {
      id: nextId++,
      name: body.name!.trim(),
      category: body.category!,
      price: Number(body.price),
      stock: Number(body.stock),
      status: body.status!,
      createTime: now,
      updateTime: now
    }
    products = [product, ...products]
    return ok(product, '创建成功')
  }),

  http.put('/api/product/:id', async ({ request, params }) => {
    const index = products.findIndex((item) => String(item.id) === params.id)
    if (index === -1) return fail(404, '商品不存在')

    const body = (await request.json()) as Partial<ProductPayload>
    if (!validPayload(body)) return fail(422, '商品参数不正确')

    products[index] = {
      ...products[index],
      name: body.name!.trim(),
      category: body.category!,
      price: Number(body.price),
      stock: Number(body.stock),
      status: body.status!,
      updateTime: new Date().toISOString()
    }
    return ok(products[index], '更新成功')
  }),

  http.delete('/api/product/:id', ({ params }) => {
    const index = products.findIndex((item) => String(item.id) === params.id)
    if (index === -1) return fail(404, '商品不存在')
    products.splice(index, 1)
    return ok(true, '删除成功')
  }),

  http.delete('/api/product', async ({ request }) => {
    const body = (await request.json()) as { ids?: Array<string | number> }
    if (!Array.isArray(body.ids) || body.ids.length === 0) {
      return fail(422, 'ids 不能为空')
    }
    const ids = new Set(body.ids.map(String))
    products = products.filter((product) => !ids.has(String(product.id)))
    return ok(true, '删除成功')
  })
]
