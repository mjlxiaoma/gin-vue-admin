# CRUD 示例：前后端完整实现流程

本文对应前端「系统管理 → 增删改查」页面，以 `crud` 示例模块为例，说明一次操作如何从页面按钮依次经过前端 API、Gin 路由、Handler、Service、Repository、GORM，最终操作 MySQL 并刷新页面。

本文重点学习真实 Go 后端。本地联调时，确保 `VITE_ENABLE_MOCK` 不是 `true`，否则请求可能被 MSW 拦截，不会进入本文所述 Go 代码。

## 1. 接口总览

| 页面操作 | HTTP 请求 | 后端方法 | 权限码 |
| --- | --- | --- | --- |
| 列表、搜索、分页 | `GET /api/crud` | `CrudHandler.List` | `crud:list` |
| 查看详情 | `GET /api/crud/:id` | `CrudHandler.Get` | `crud:list` |
| 新增 | `POST /api/crud` | `CrudHandler.Create` | `crud:create` |
| 编辑 | `PUT /api/crud/:id` | `CrudHandler.Update` | `crud:edit` |
| 删除 | `DELETE /api/crud/:id` | `CrudHandler.Delete` | `crud:delete` |
| 批量删除 | `DELETE /api/crud` | `CrudHandler.BatchDelete` | `crud:delete` |

> 当前页面点击「查看」或「编辑」时直接使用表格行数据，没有调用详情接口；`GET /api/crud/:id` 是后端已经提供、前端已经封装但当前页面未使用的接口。

## 2. 推荐阅读顺序

建议按下面顺序对照源码阅读：

1. `server/internal/model/crud.go`：数据表和实体字段。
2. `server/cmd/api/main.go`：依赖如何组装。
3. `server/internal/server/router.go`：URL 如何映射到 Handler。
4. `server/internal/handler/crud.go`：参数绑定和 HTTP 响应。
5. `server/internal/service/crud.go`：业务规则。
6. `server/internal/repository/crud.go`：当前模块特有的列表过滤。
7. `server/internal/repository/generic.go`：通用增删改查实现。
8. `web/src/modules/crud/api.ts`：前端请求封装。
9. `web/src/modules/crud/views/List.vue`：页面操作入口。
10. `web/src/modules/crud/views/CrudFormDrawer.vue`：新增和编辑表单。

## 3. 后端分层

这个模块比 Dashboard 图表接口多一层 Repository：

```text
HTTP 请求
   ↓
Handler：接收参数、校验格式、返回 HTTP 响应
   ↓
Service：业务逻辑，例如“记录不存在就返回 404”
   ↓
Repository：使用 GORM 操作数据库
   ↓
MySQL：crud_items 表
```

对应文件：

```text
server/internal/handler/crud.go
server/internal/service/crud.go
server/internal/repository/crud.go
server/internal/repository/generic.go
server/internal/model/crud.go
```

各层可以用前端开发中熟悉的概念辅助理解：

| 后端层 | 职责 | 类比 |
| --- | --- | --- |
| Router | 根据请求方法和 URL 找处理函数 | 后端版路由表 |
| Handler | 读取路径参数、查询参数、JSON 请求体 | API 接收层 |
| Service | 处理业务规则和错误翻译 | 独立业务逻辑 |
| Repository | 封装数据访问 | 数据访问模块 |
| GORM | 生成并执行 SQL | Prisma 等 ORM |
| Model | 描述数据库实体 | 数据库实体的类型定义 |

## 4. 启动时的依赖注入

`server/cmd/api/main.go:106`：

```go
crudRepo := repository.NewCrudRepository(gdb)
crudSvc := service.NewCrudService(crudRepo)
crudHandler := handler.NewCrudHandler(crudSvc)
```

组装关系：

```text
数据库连接 gdb
   ↓ 注入
CrudRepository
   ↓ 注入
CrudService
   ↓ 注入
CrudHandler
```

因此 Handler 不直接操作数据库，而是调用：

```go
h.svc.List(...)
h.svc.Get(...)
h.svc.Create(...)
h.svc.Update(...)
h.svc.Delete(...)
h.svc.BatchDelete(...)
```

Service 再调用：

```go
s.repo.List(...)
s.repo.FindByID(...)
s.repo.Create(...)
s.repo.Update(...)
s.repo.Delete(...)
s.repo.BatchDelete(...)
```

这段组装只在服务启动时执行，不是每次请求都重新创建数据库连接和对象。

## 5. Model：Go 对象、JSON 和数据库表

模型定义在 `server/internal/model/crud.go:8`：

```go
type CrudItem struct {
    Model
    Date     string `gorm:"size:10" json:"date"`
    Name     string `gorm:"size:64;not null" json:"name"`
    Province string `gorm:"size:64" json:"province"`
    City     string `gorm:"size:64" json:"city"`
    Address  string `gorm:"size:255" json:"address"`
    Zip      int    `gorm:"default:0" json:"zip"`
}
```

表名在 `server/internal/model/crud.go:19`：

```go
func (CrudItem) TableName() string {
    return "crud_items"
}
```

所以：

```text
Go 模型：model.CrudItem
MySQL 表：crud_items
```

字段对应关系：

| Go 字段 | 数据库字段 | JSON 字段 |
| --- | --- | --- |
| `Name` | `name` | `name` |
| `Province` | `province` | `province` |
| `City` | `city` | `city` |
| `Address` | `address` | `address` |
| `Zip` | `zip` | `zip` |
| `Date` | `date` | `date` |

标签含义：

- `gorm:"size:64;not null"`：数据库列约束。
- `json:"name"`：序列化为 JSON 时使用的字段名。

### 5.1 公共 Model

`CrudItem` 嵌入了 `Model`。公共字段定义在 `server/internal/model/base.go:19`：

```go
type Model struct {
    ID        uint           `json:"id"`
    CreatedAt time.Time      `json:"createTime"`
    UpdatedAt time.Time      `json:"updateTime"`
    CreatedBy uint           `json:"createdBy"`
    UpdatedBy uint           `json:"updatedBy"`
    DeletedAt gorm.DeletedAt `json:"-"`
    DeletedBy uint           `json:"-"`
}
```

一条 `CrudItem` 实际还包含：

```text
ID
创建时间
修改时间
创建人
修改人
删除时间
删除人
```

其中 `DeletedAt gorm.DeletedAt` 表示该模型支持软删除。

## 6. 通用 Repository：为什么找不到单独的 Create 实现？

`server/internal/repository/crud.go:25`：

```go
type crudRepository struct {
    *GenericRepository[model.CrudItem]
}
```

这里嵌入了通用仓储，可以类比 TypeScript：

```ts
GenericRepository<CrudItem>
```

在这个模块中，泛型参数 `T` 就是 `CrudItem`。因此以下方法都由 `GenericRepository` 提供：

```text
FindByID
Create
Update
Delete
BatchDelete
```

`crudRepository` 只额外实现自己的 `List`，因为它需要增加姓名模糊搜索。

复用思路：

```text
新增、修改、删除：各模块基本相同 → 通用基类实现
列表过滤条件：各模块不同         → 各模块自己实现
```

## 7. 每次请求先经过登录和权限检查

路由位于 `server/internal/server/router.go:122`：

```go
crud := api.Group("/crud")
crud.Use(middleware.AuthRequired(jwtMgr))
{
    crud.GET("", middleware.RequirePermission(permRepo, "crud:list"), crudHandler.List)
    crud.GET("/:id", middleware.RequirePermission(permRepo, "crud:list"), crudHandler.Get)
    crud.POST("", middleware.RequirePermission(permRepo, "crud:create"), crudHandler.Create)
    crud.PUT("/:id", middleware.RequirePermission(permRepo, "crud:edit"), crudHandler.Update)
    crud.DELETE("/:id", middleware.RequirePermission(permRepo, "crud:delete"), crudHandler.Delete)
    crud.DELETE("", middleware.RequirePermission(permRepo, "crud:delete"), crudHandler.BatchDelete)
}
```

`api` 已带 `/api` 前缀，因此：

```go
crud := api.Group("/crud")
crud.POST("", crudHandler.Create)
```

对应：

```http
POST /api/crud
```

每次请求的顺序：

```text
AuthRequired
  → 验证 Authorization 中的 JWT
  → 得到当前用户 ID
       ↓
RequirePermission
  → 查询当前用户的权限码
  → 检查是否有 crud:list/create/edit/delete
       ↓
CrudHandler 对应方法
```

超级管理员拥有 `"*"` 权限，会直接放行。

常见错误：

```text
没有登录或 Token 无效 → 401
没有对应权限         → 403
```

---

## 8. 查询列表：打开页面、搜索、分页和刷新

### 8.1 页面挂载时调用 fetchList

`web/src/modules/crud/views/List.vue:172`：

```ts
onMounted(fetchList)
```

「搜索」「重置」「分页」「刷新」最终也都会调用 `fetchList`。

### 8.2 组装分页和搜索参数

`List.vue:117`：

```ts
useCrud<CrudItem>({
  fetch: (params) =>
    fetchCrud({
      name: (params.keyword as string) || '',
      current: Number(params.page ?? 1),
      size: Number(params.size ?? 10)
    }),
  remove: deleteCrudItem,
  batchRemove: batchDeleteCrudItems,
  defaultSearchForm: { keyword: '' },
  pageSize: 10
})
```

假设用户输入「张」，当前第一页、每页 10 条，最终参数：

```ts
{
  name: '张',
  current: 1,
  size: 10
}
```

### 8.3 发送 GET 请求

`web/src/modules/crud/api.ts:33`：

```ts
export const fetchCrud = (req: CrudListRequest) =>
  api.get<CrudListResponse>('/api/crud', { params: req })
```

最终请求：

```http
GET /api/crud?name=张&current=1&size=10
```

这些是 URL 查询参数，不是 JSON 请求体。

### 8.4 Handler 绑定查询参数

`server/internal/handler/crud.go:19`：

```go
type crudListQuery struct {
    Current int    `form:"current"`
    Size    int    `form:"size"`
    Name    string `form:"name"`
}
```

`form:"current"` 表示从 URL 的 `current` 参数取值。

Handler 位于 `crud.go:59`：

```go
func (h *CrudHandler) List(c *gin.Context) {
    var q crudListQuery
    _ = c.ShouldBindQuery(&q)

    pq := pagination.Query{Page: q.Current, Size: q.Size, Keyword: q.Name}
    res, err := h.svc.List(c.Request.Context(), pq)
    if err != nil {
        apperr.Write(c, err)
        return
    }
    response.Success(c, res)
}
```

绑定后相当于：

```go
q = crudListQuery{Current: 1, Size: 10, Name: "张"}
```

随后转换成通用分页参数：

```go
pq = pagination.Query{Page: 1, Size: 10, Keyword: "张"}
```

### 8.5 Service 规范化分页参数

`server/internal/service/crud.go:29`：

```go
func (s *CrudService) List(ctx context.Context, q pagination.Query) (pagination.Result[model.CrudItem], error) {
    q.Normalize()
    res, err := s.repo.List(ctx, q)
    if err != nil {
        return pagination.Result[model.CrudItem]{}, err
    }
    res.Current = q.Page
    res.Size = q.Size
    return res, nil
}
```

`Normalize` 位于 `server/internal/pkg/pagination/pagination.go:21`：

```go
func (q *Query) Normalize() {
    if q.Page < 1 {
        q.Page = 1
    }
    if q.Size < 1 {
        q.Size = DefaultSize
    }
    if q.Size > maxPageSize {
        q.Size = maxPageSize
    }
}
```

规则：

```text
页码小于 1   → 改成 1
每页条数没传 → 使用 DefaultSize，默认 10
每页超过 100 → 限制为 100
```

### 8.6 Repository 添加姓名模糊搜索

`server/internal/repository/crud.go:35`：

```go
func (r *crudRepository) List(ctx context.Context, q pagination.Query) (pagination.Result[model.CrudItem], error) {
    return r.GenericRepository.List(ctx, q, func(db *gorm.DB) *gorm.DB {
        if q.Keyword != "" {
            return db.Where("name LIKE ?", "%"+q.Keyword+"%")
        }
        return db
    })
}
```

搜索「张」时添加：

```sql
name LIKE '%张%'
```

含义是姓名中包含「张」即可。没有输入姓名时不添加该条件。

### 8.7 通用分页执行两条 SQL

分页实现位于 `server/internal/pkg/pagination/pagination.go:43`。

先查询总条数：

```go
countDB := build(db.Session(&gorm.Session{}))
countDB.Model(new(T)).Count(&total)
```

大致相当于：

```sql
SELECT COUNT(*)
FROM crud_items
WHERE name LIKE '%张%'
  AND deleted_at IS NULL;
```

再查询当前页：

```go
offset := (q.Page - 1) * q.Size
listDB.Offset(offset).Limit(q.Size).Find(&records)
```

第一页、每页 10 条：

```text
offset = (1 - 1) × 10 = 0
limit  = 10
```

大致 SQL：

```sql
SELECT *
FROM crud_items
WHERE name LIKE '%张%'
  AND deleted_at IS NULL
LIMIT 10 OFFSET 0;
```

第二页时：

```text
offset = (2 - 1) × 10 = 10
```

即跳过前 10 条。

### 8.8 返回分页结果并更新表格

```go
return pagination.Result[T]{
    Records: records,
    Total:   total,
    Current: q.Page,
    Size:    q.Size,
}, nil
```

HTTP 响应示意：

```json
{
  "code": 0,
  "data": {
    "records": [
      {
        "id": 1,
        "name": "张三",
        "province": "上海",
        "city": "上海",
        "address": "浦东新区某路",
        "zip": 200000,
        "date": "2026-11-12"
      }
    ],
    "total": 1,
    "current": 1,
    "size": 10
  },
  "msg": "ok"
}
```

前端响应拦截器取出 `data` 后，`useCrud` 保存结果：

```ts
listData.value = res.records
pagination.total = res.total
```

因此：

```text
records → 表格数据
total   → 分页总条数
```

没有符合条件的记录时，页面显示「暂无数据」和「共 0 条」。

---

## 9. 新增：从打开抽屉到 INSERT

### 9.1 点击新增只打开抽屉，不请求后端

`List.vue:34`：

```vue
<el-button @click="openDrawer('add')">新增</el-button>
```

```ts
const openDrawer = (mode: 'add' | 'edit' | 'view', row?: CrudItem) => {
  drawerMode.value = mode
  editingRow.value = row ?? null
  drawerVisible.value = true
}
```

新增时：

```ts
drawerMode.value = 'add'
editingRow.value = null
drawerVisible.value = true
```

### 9.2 表单提交并组装数据

`CrudFormDrawer.vue:127`：

```ts
const payload = {
  date: formData.date,
  name: formData.name,
  province: formData.province,
  city: formData.city,
  address: formData.address,
  zip: Number(formData.zip) || 0
}
```

新增模式调用：

```ts
await createCrudItem(payload)
```

请求数据示例：

```json
{
  "date": "2026-11-12",
  "name": "张三",
  "province": "上海",
  "city": "上海",
  "address": "浦东新区某路",
  "zip": 200000
}
```

### 9.3 前端发送 POST

`web/src/modules/crud/api.ts:41`：

```ts
export const createCrudItem = (payload: CrudCreatePayload) =>
  api.post<item>('/api/crud', payload)
```

最终请求：

```http
POST /api/crud
Content-Type: application/json
Authorization: Bearer <token>
```

请求体为上面的 JSON。

### 9.4 Handler 绑定和校验 JSON

请求结构位于 `server/internal/handler/crud.go:26`：

```go
type crudUpsertReq struct {
    Date     string `json:"date"`
    Name     string `json:"name" binding:"required"`
    Province string `json:"province"`
    City     string `json:"city"`
    Address  string `json:"address"`
    Zip      int    `json:"zip"`
}
```

Create 位于 `crud.go:104`：

```go
var req crudUpsertReq
if err := c.ShouldBindJSON(&req); err != nil {
    apperr.Write(c, apperr.Validation(err.Error(), nil))
    return
}
```

`ShouldBindJSON` 做两件事：

1. 把请求体 JSON 转成 Go 结构体。
2. 根据 `binding:"required"` 等标签校验。

后端这里只强制要求 `name`。前端表单还校验城市、地址、邮编等，那些属于前端规则。

JSON 格式错误或缺少姓名时，当前 `apperr.Validation` 实际返回 HTTP 422。

### 9.5 请求对象转换成数据库模型

```go
e := &model.CrudItem{
    Date: req.Date, Name: req.Name, Province: req.Province,
    City: req.City, Address: req.Address, Zip: req.Zip,
}
```

两种结构职责不同：

```text
crudUpsertReq：描述“前端允许传什么”
model.CrudItem：描述“数据库要保存什么”
```

前端不需要传 ID、CreatedAt、CreatedBy、UpdatedAt，这些由后端和数据库处理。

### 9.6 Service 调用 Repository

```go
if err := h.svc.Create(c.Request.Context(), e); err != nil {
    apperr.Write(c, err)
    return
}
```

`server/internal/service/crud.go:53`：

```go
func (s *CrudService) Create(ctx context.Context, e *model.CrudItem) error {
    return s.repo.Create(ctx, e)
}
```

该示例的新增没有额外业务规则，所以 Service 直接调用 Repository。真实业务中的唯一性检查、编号生成等规则应放在 Service。

### 9.7 通用 Repository 执行 INSERT

`server/internal/repository/generic.go:119`：

```go
func (r *GenericRepository[T]) Create(ctx context.Context, e *T) error {
    return r.db.WithContext(ctx).Create(e).Error
}
```

GORM 大致执行：

```sql
INSERT INTO crud_items (
    date, name, province, city, address, zip,
    created_at, updated_at, created_by, updated_by
) VALUES (
    '2026-11-12', '张三', '上海', '上海', '浦东新区某路', 200000,
    NOW(), NOW(), 当前用户ID, 当前用户ID
);
```

创建成功后，GORM 把数据库生成的主键回填到：

```go
e.ID
```

审计回调根据 Context 中的用户 ID 自动写入 `CreatedBy` 和 `UpdatedBy`。

### 9.8 返回对象并刷新列表

```go
response.Success(c, e, "创建成功")
```

响应示意：

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "name": "张三",
    "province": "上海",
    "city": "上海",
    "address": "浦东新区某路",
    "zip": 200000,
    "date": "2026-11-12",
    "createTime": "2026-11-12T15:00:00+08:00"
  },
  "msg": "创建成功"
}
```

前端成功后：

```ts
ElMessage.success('新增成功')
emit('success')
```

父页面执行：

```ts
const onFormSuccess = () => {
  drawerVisible.value = false
  fetchList()
}
```

完整结果：

```text
INSERT 成功
→ 返回成功响应
→ 关闭抽屉
→ 重新 GET /api/crud
→ 新记录出现在表格中
```

---

## 10. 编辑：先查询，再修改，再保存

### 10.1 点击编辑，使用当前行数据回填

`List.vue:63`：

```vue
<el-button @click="openDrawer('edit', row)">编辑</el-button>
```

`CrudFormDrawer.vue:102`：

```ts
if ((props.mode === 'edit' || props.mode === 'view') && props.data) {
  Object.assign(formData, {
    id: props.data.id,
    name: props.data.name,
    province: props.data.province,
    city: props.data.city,
    address: props.data.address,
    zip: props.data.zip,
    date: props.data.date
  })
}
```

当前页面没有调用详情接口，而是直接使用列表中的 `row`。

### 10.2 提交 PUT 请求

`CrudFormDrawer.vue:141`：

```ts
await updateCrudItem(props.data!.id, payload)
```

前端 API：

```ts
export const updateCrudItem = (id: string, payload: CrudUpdatePayload) =>
  api.put<item>(`/api/crud/${id}`, payload)
```

假设修改 ID 为 1 的记录：

```http
PUT /api/crud/1
Content-Type: application/json
```

```json
{
  "date": "2026-11-12",
  "name": "张三",
  "province": "浙江",
  "city": "杭州",
  "address": "西湖区某路",
  "zip": 310000
}
```

这里路径参数 `id = 1`，请求体是修改后的字段。

### 10.3 解析路径 ID

`server/internal/handler/crud.go:132`：

```go
id, err := strconv.ParseUint(c.Param("id"), 10, 64)
if err != nil {
    apperr.Write(c, apperr.Validation("无效的 id", nil))
    return
}
```

`c.Param("id")` 从 `/api/crud/1` 中读取 `"1"`，`ParseUint` 将字符串转换为无符号整数。

请求 `PUT /api/crud/abc` 时无法转换，直接返回参数校验错误。

### 10.4 绑定修改后的 JSON

```go
var req crudUpsertReq
if err := c.ShouldBindJSON(&req); err != nil {
    apperr.Write(c, apperr.Validation(err.Error(), nil))
    return
}
```

新增和编辑共用 `crudUpsertReq`，因为二者字段相同。

### 10.5 查询旧记录

```go
e, err := h.svc.Get(c.Request.Context(), uint(id))
```

`server/internal/service/crud.go:41`：

```go
e, err := s.repo.FindByID(ctx, id)
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, apperr.NotFound("记录不存在")
    }
    return nil, err
}
return e, nil
```

通用 Repository：

```go
e := new(T)
r.db.WithContext(ctx).First(e, id)
```

大致 SQL：

```sql
SELECT *
FROM crud_items
WHERE id = 1
  AND deleted_at IS NULL
LIMIT 1;
```

先查询的目的：

1. 确认记录存在，不存在返回 404。
2. 得到原有对象，保留 ID、创建时间、创建人等字段。
3. 在该对象上修改业务字段。

### 10.6 修改内存对象

```go
e.Date = req.Date
e.Name = req.Name
e.Province = req.Province
e.City = req.City
e.Address = req.Address
e.Zip = req.Zip
```

此时只修改了 Go 内存中的对象，数据库尚未更新。

### 10.7 Save 保存到数据库

```go
if err := h.svc.Update(c.Request.Context(), e); err != nil {
    apperr.Write(c, err)
    return
}
```

Service：

```go
func (s *CrudService) Update(ctx context.Context, e *model.CrudItem) error {
    return s.repo.Update(ctx, e)
}
```

通用 Repository：

```go
func (r *GenericRepository[T]) Update(ctx context.Context, e *T) error {
    return r.db.WithContext(ctx).Save(e).Error
}
```

因为 `e.ID` 已有值，GORM 大致执行：

```sql
UPDATE crud_items
SET date = '2026-11-12',
    name = '张三',
    province = '浙江',
    city = '杭州',
    address = '西湖区某路',
    zip = 310000,
    updated_at = NOW(),
    updated_by = 当前用户ID
WHERE id = 1;
```

`Save` 是全量保存，空字符串、0 等零值也会写入。因此当前前端编辑时提交完整表单，而不是只提交一个修改字段。

### 10.8 返回并刷新列表

```go
response.Success(c, e, "更新成功")
```

前端成功后：

```text
提示“更新成功”
→ 关闭抽屉
→ fetchList()
→ 表格显示修改后的数据
```

---

## 11. 单条删除：实际执行软删除

### 11.1 点击删除，先弹确认框

页面按钮：

```vue
<el-button @click="handleDelete(row.id)">删除</el-button>
```

`handleDelete` 来自 `useCrud`，位于 `web/src/app/composables/useCrud.ts:77`：

```ts
const confirmed = await confirmService.showConfirm('确认删除该记录？', '提示')
if (!confirmed) return

await remove(id)
ElMessage.success('删除成功')
await fetchList()
```

点击取消不会请求后端；点击确认才执行 `deleteCrudItem(id)`。

### 11.2 发送 DELETE 请求

`web/src/modules/crud/api.ts:49`：

```ts
export const deleteCrudItem = (id: string) =>
  api.del<boolean>(`/api/crud/${id}`)
```

例如：

```http
DELETE /api/crud/1
```

没有请求体，ID 位于 URL 中。

### 11.3 Handler 解析 ID 并调用 Service

`server/internal/handler/crud.go:170`：

```go
id, err := strconv.ParseUint(c.Param("id"), 10, 64)
if err != nil {
    apperr.Write(c, apperr.Validation("无效的 id", nil))
    return
}
if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
    apperr.Write(c, err)
    return
}
response.Success(c, true, "删除成功")
```

### 11.4 Service 先确认记录存在

`server/internal/service/crud.go:63`：

```go
func (s *CrudService) Delete(ctx context.Context, id uint) error {
    if _, err := s.repo.FindByID(ctx, id); err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return apperr.NotFound("记录不存在")
        }
        return err
    }
    return s.repo.Delete(ctx, id)
}
```

流程：

```text
根据 ID 查询
  ├── 不存在     → 返回 404
  ├── 数据库出错 → 返回错误
  └── 存在       → 执行删除
```

### 11.5 Repository 执行软删除

`server/internal/repository/generic.go:131`：

```go
func (r *GenericRepository[T]) Delete(ctx context.Context, id uint) error {
    if isSoftDelete[T]() {
        return r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(softDeleteUpdates(ctx)).Error
    }
    return r.db.WithContext(ctx).Delete(new(T), id).Error
}
```

因为 `CrudItem` 包含 `gorm.DeletedAt`，所以走软删除分支。

`softDeleteUpdates` 生成：

```go
map[string]any{
    "deleted_at": time.Now(),
    "deleted_by": 当前用户ID,
}
```

大致 SQL：

```sql
UPDATE crud_items
SET deleted_at = NOW(),
    deleted_by = 当前用户ID
WHERE id = 1
  AND deleted_at IS NULL;
```

不是：

```sql
DELETE FROM crud_items WHERE id = 1;
```

软删除后：

```text
数据库记录还在
deleted_at 有值
普通 GORM 查询自动排除它
页面刷新后不再显示它
```

这就是列表 SQL 中一直出现 `deleted_at IS NULL` 的原因。

### 11.6 返回并刷新

```json
{
  "code": 0,
  "data": true,
  "msg": "删除成功"
}
```

前端随后执行 `fetchList()`，被软删除的记录不再出现。

---

## 12. 批量删除：一次 UPDATE 处理多个 ID

### 12.1 保存勾选行并提取 ID

```ts
const handleSelectionChange = (rows: T[]) => {
  selectedRows.value = rows
}
```

点击批量删除：

```ts
const ids = selectedRows.value.map((r) => r.id)
await batchRemove(ids)
```

后端期望的请求体：

```json
{
  "ids": ["1", "2", "3"]
}
```

### 12.2 DELETE 请求携带 JSON 请求体

`web/src/modules/crud/api.ts:53`：

```ts
export const batchDeleteCrudItems = (ids: string[]) =>
  api.del<boolean>('/api/crud', { data: { ids } })
```

最终请求：

```http
DELETE /api/crud
Content-Type: application/json
```

```json
{
  "ids": ["1", "2", "3"]
}
```

单条与批量删除的区别：

```text
单条删除：DELETE /api/crud/1
批量删除：DELETE /api/crud + { ids: [...] }
```

### 12.3 绑定并校验 IDs

公共请求结构位于 `server/internal/handler/common.go:6`：

```go
type batchDeleteReq struct {
    IDs []string `json:"ids" binding:"required,min=1"`
}
```

要求必须有 `ids`，且至少包含一个元素。

Handler 位于 `server/internal/handler/crud.go:192`：

```go
var req crudBatchDeleteReq
if err := c.ShouldBindJSON(&req); err != nil {
    apperr.Write(c, apperr.Validation(err.Error(), nil))
    return
}
```

再把字符串 ID 全部转换为 `uint`：

```go
ids := make([]uint, 0, len(req.IDs))
for _, s := range req.IDs {
    id, err := strconv.ParseUint(s, 10, 64)
    if err != nil {
        apperr.Write(c, apperr.Validation("无效的 id: "+s, nil))
        return
    }
    ids = append(ids, uint(id))
}
```

转换结果：

```text
["1", "2", "3"]
       ↓
[]uint{1, 2, 3}
```

只要有一个 ID 无效，就立即返回错误，不执行数据库删除。

### 12.4 一条 SQL 批量软删除

Service：

```go
func (s *CrudService) BatchDelete(ctx context.Context, ids []uint) error {
    return s.repo.BatchDelete(ctx, ids)
}
```

Repository 位于 `server/internal/repository/generic.go:139`：

```go
if len(ids) == 0 {
    return nil
}
return r.db.WithContext(ctx).Model(new(T)).Where("id IN ?", ids).Updates(softDeleteUpdates(ctx)).Error
```

大致 SQL：

```sql
UPDATE crud_items
SET deleted_at = NOW(),
    deleted_by = 当前用户ID
WHERE id IN (1, 2, 3)
  AND deleted_at IS NULL;
```

它不是循环执行三次 SQL，而是用一次 `IN` 更新多条记录。

批量删除不会逐个检查记录是否存在；合法但不存在或已删除的 ID 不会匹配到数据，也不会单独返回 404。

成功后：

```go
response.Success(c, true, "删除成功")
```

前端：

```ts
selectedRows.value = []
await fetchList()
```

即清空勾选状态并刷新列表。

---

## 13. 详情接口

后端提供：

```http
GET /api/crud/1
```

Handler 位于 `server/internal/handler/crud.go:80`：

```go
id, err := strconv.ParseUint(c.Param("id"), 10, 64)
if err != nil {
    apperr.Write(c, apperr.Validation("无效的 id", nil))
    return
}
e, err := h.svc.Get(c.Request.Context(), uint(id))
if err != nil {
    apperr.Write(c, err)
    return
}
response.Success(c, e)
```

链路：

```text
GET /api/crud/1
→ 解析 ID
→ Service.Get
→ Repository.FindByID
→ SELECT ... WHERE id = 1 AND deleted_at IS NULL
→ 找到：200 + 记录
→ 没找到：404 记录不存在
```

前端也封装了：

```ts
export const fetchCrudDetail = (id: string) =>
  api.get<item>(`/api/crud/${id}`)
```

如果以后希望编辑时获取最新数据，可以改为「点击编辑 → 请求详情 → 打开抽屉并回填」。当前实现直接使用列表行数据。

---

## 14. 统一响应与异常流程

### 14.1 成功响应

Handler 使用：

```go
response.Success(c, data, "操作成功")
```

统一格式：

```json
{
  "code": 0,
  "data": {},
  "msg": "操作成功"
}
```

响应还可能带 `traceId`，用于请求追踪。

前端响应拦截器会自动取出 `data`，所以业务代码直接得到列表、记录或布尔值。

### 14.2 失败响应

Handler 使用：

```go
apperr.Write(c, err)
return
```

常见情况：

| 情况 | HTTP 状态码 |
| --- | ---: |
| 未登录、Token 无效 | 401 |
| 没有操作权限 | 403 |
| 修改或删除的记录不存在 | 404 |
| JSON、ID 或请求参数校验失败 | 422 |
| 未预期的数据库或服务错误 | 500 |

失败响应使用 ProblemDetail 格式。前端响应拦截器将其转换成 `HttpError` 并抛出，不会进入成功后的刷新逻辑。

---

## 15. 五种操作的流程模板

### 15.1 查询列表

```text
绑定 URL 查询参数
→ 规范化分页参数
→ 添加业务过滤条件
→ COUNT 查询总数
→ LIMIT/OFFSET 查询当前页
→ 返回 records/total/current/size
```

### 15.2 新增

```text
绑定 JSON 请求体
→ 校验参数
→ 创建 model.CrudItem
→ Repository.Create
→ INSERT INTO crud_items
→ GORM 回填 ID
→ 返回创建后的对象
```

### 15.3 编辑

```text
解析 URL 中的 ID
→ 绑定 JSON 请求体
→ 根据 ID 查询旧记录
→ 不存在返回 404
→ 修改旧对象字段
→ Repository.Save
→ UPDATE crud_items
→ 返回修改后的对象
```

### 15.4 单条删除

```text
解析 ID
→ 查询记录是否存在
→ 不存在返回 404
→ UPDATE deleted_at、deleted_by
→ 返回 true
→ 前端刷新列表
```

### 15.5 批量删除

```text
绑定 { ids: [...] }
→ 字符串 ID 转成 uint
→ UPDATE ... WHERE id IN (...)
→ 返回 true
→ 前端清空选择并刷新列表
```

## 16. 联调注意事项：ID 类型存在不一致

后端 `Model.ID` 是 `uint`，JSON 中实际返回数字 ID；但前端 `web/src/modules/crud/api.ts:20` 把 `item.id` 声明为 `string`，批量删除后端又要求 `ids` 为字符串数组。

因此真实后端联调时，如果批量删除出现参数绑定错误，应重点检查：列表返回的数字 ID 是否在发送前被转换成字符串。

单条删除使用 URL 模板拼接，数字也能直接拼入 URL，通常不受这个问题影响。

## 17. 最后需要记住的主线

```text
Vue 页面：决定什么时候请求
Axios：发送请求、携带 Token、解包响应
Gin Router：把 URL 分配给 Handler
AuthRequired：验证登录身份
RequirePermission：验证操作权限
Handler：接收并校验参数
Service：执行业务规则
Repository：调用 GORM
GORM：生成并执行 SQL
MySQL：保存和查询数据
response.Success：统一成功响应
apperr.Write：统一错误响应
```

理解这套流程后，用户、角色、权限、字典等模块也可以按照同样的方式阅读；主要差异通常在过滤条件、关联关系和 Service 中的业务规则。
