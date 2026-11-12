# 项目目录与后端架构

本文梳理当前仓库的目录职责，重点说明 Go 后端的启动方式、分层结构、请求链路和商品模块在目录中的分布。前端只保留整体结构及与后端联调相关的位置。

目录树省略 `node_modules`、锁文件、本地工具产物以及重复的 `*_test.go` 测试文件。

## 1. 项目根目录

```text
gin-vue-admin/
├── server/                     Go 后端
│   ├── cmd/                    程序启动入口
│   ├── configs/                配置文件
│   ├── docs/                   Swagger 和后端学习文档
│   ├── internal/               后端核心业务代码
│   ├── go.mod                  Go 模块和依赖声明
│   ├── go.sum                  依赖版本校验
│   └── .golangci.yml           Go Lint 配置
├── web/                        Vue 3 前端
│   ├── src/                    前端源码
│   ├── test/                   前端测试
│   ├── scripts/                前端辅助脚本
│   ├── docs/                   前端规范文档
│   ├── public/                 静态资源
│   ├── vite.config.ts          Vite、代理和构建配置
│   ├── vitest.config.ts        Vitest 配置
│   ├── playwright.config.ts    浏览器测试配置
│   └── package.json            前端依赖和命令
├── deploy/                     部署配置
│   ├── docker-compose.yml
│   ├── docker-compose.tls.yml
│   ├── nginx.conf
│   ├── nginx.tls.conf
│   ├── server.Dockerfile       后端镜像
│   ├── web.Dockerfile          前端镜像
│   └── scripts/                部署脚本
├── docs-site/                  VitePress 文档站点
│   ├── guide/                  使用指南
│   ├── public/                 文档静态资源
│   ├── index.md
│   └── package.json
├── .github/                    GitHub Actions、PR 模板等
├── Makefile                    常用开发命令
├── README.md                   项目总说明
├── DEPLOYMENT.md               部署说明
└── CHANGELOG.md                版本变更记录
```

`.playwright-mcp/` 等目录是本地浏览器验证产生的截图和快照，不属于业务架构。

## 2. 后端启动入口：`cmd/`

```text
server/cmd/
├── api/
│   └── main.go                 API 服务主入口
└── scaffold/
    └── main.go                 代码脚手架命令入口
```

`cmd/api/main.go` 是整个后端的装配入口，启动时大致执行：

```text
读取配置
   ↓
初始化日志
   ↓
连接 MySQL
   ↓
AutoMigrate 自动迁移数据表
   ↓
创建 Repository
   ↓
创建 Service
   ↓
创建 Handler
   ↓
注册 Gin Router 和中间件
   ↓
启动 HTTP 服务
```

商品模块的依赖装配：

```go
productRepo := repository.NewProductRepository(gdb)
productSvc := service.NewProductService(productRepo)
productHandler := handler.NewProductHandler(productSvc)
```

对象关系为：

```text
GORM 数据库连接
    ↓
ProductRepository
    ↓
ProductService
    ↓
ProductHandler
    ↓
注册到 Gin Router
```

当阅读业务代码时，如果不知道 `handler.svc`、`service.repo` 或 `repository.db` 从哪里来，就回到 `cmd/api/main.go` 查看装配过程。

## 3. 配置和数据库连接

```text
server/
├── configs/
│   ├── config.yaml             当前运行配置
│   └── config.example.yaml     配置示例
└── internal/
    ├── config/
    │   └── config.go           配置结构、加载和环境变量覆盖
    └── db/
        └── mysql.go            MySQL/GORM 连接初始化
```

两者职责不同：

- `configs/config.yaml`：保存 MySQL、JWT、服务端口等具体配置。
- `internal/config/config.go`：定义配置结构并负责加载配置。
- `internal/db/mysql.go`：使用配置创建 GORM 的 `*gorm.DB`。

## 4. 后端核心目录：`internal/`

Go 的 `internal` 目录用于存放项目内部实现，不作为对外公共 SDK。

```text
server/internal/
├── config/                     配置加载
│   └── config.go
├── db/                         数据库连接
│   └── mysql.go
├── server/                     HTTP 服务和路由注册
│   └── router.go
├── middleware/                 Gin 中间件
│   ├── auth.go
│   ├── permission.go
│   ├── cors.go
│   ├── logger.go
│   ├── operation_log.go
│   ├── recovery.go
│   └── traceid.go
├── handler/                    HTTP 接口层
│   ├── auth.go
│   ├── user.go
│   ├── role.go
│   ├── permission.go
│   ├── menu.go
│   ├── dept.go
│   ├── crud.go
│   ├── product.go
│   ├── dashboard.go
│   ├── dict.go
│   ├── sys_config.go
│   ├── notice.go
│   ├── operation_log.go
│   ├── login_log.go
│   ├── health.go
│   └── common.go
├── service/                    业务逻辑层
│   ├── auth.go
│   ├── user.go
│   ├── role.go
│   ├── permission.go
│   ├── menu.go
│   ├── dept.go
│   ├── crud.go
│   ├── product.go
│   ├── dashboard.go
│   ├── dict.go
│   ├── sys_config.go
│   ├── notice.go
│   ├── operation_log.go
│   └── login_log.go
├── repository/                 数据库访问层
│   ├── generic.go
│   ├── user.go
│   ├── role.go
│   ├── permission.go
│   ├── menu.go
│   ├── dept.go
│   ├── crud.go
│   ├── product.go
│   ├── dict.go
│   ├── sys_config.go
│   ├── notice.go
│   ├── operation_log.go
│   └── login_log.go
├── model/                      GORM 实体和数据库表结构
│   ├── base.go
│   ├── rbac.go
│   ├── crud.go
│   ├── product.go
│   ├── dept.go
│   ├── dict.go
│   ├── menu.go
│   ├── notice.go
│   ├── sys_config.go
│   ├── operation_log.go
│   └── login_log.go
├── pkg/                        项目内部通用基础能力
└── testutil/                   后端测试工具
    └── db.go
```

对应的单元测试和集成测试放在相同目录，以 `*_test.go` 命名。

## 5. 一次请求经过的核心目录

```text
internal/server/router.go
        ↓
internal/middleware/
        ↓
internal/handler/
        ↓
internal/service/
        ↓
internal/repository/
        ↓
internal/model/ + GORM + MySQL
```

### 5.1 `server/`：注册路由

`internal/server/router.go` 负责：

- 创建 `/api` 路由组。
- 注册全局中间件。
- 注册各业务模块路由。
- 把 HTTP Method 和 URL 映射到 Handler 方法。

商品路由：

```go
product := api.Group("/product")
product.Use(middleware.AuthRequired(jwtMgr))
{
    product.GET("", middleware.RequirePermission(permRepo, "product:list"), productHandler.List)
    product.GET("/:id", middleware.RequirePermission(permRepo, "product:list"), productHandler.Get)
    product.POST("", middleware.RequirePermission(permRepo, "product:create"), productHandler.Create)
    product.PUT("/:id", middleware.RequirePermission(permRepo, "product:edit"), productHandler.Update)
    product.DELETE("/:id", middleware.RequirePermission(permRepo, "product:delete"), productHandler.Delete)
    product.DELETE("", middleware.RequirePermission(permRepo, "product:delete"), productHandler.BatchDelete)
}
```

### 5.2 `middleware/`：公共请求处理

| 文件 | 职责 |
| --- | --- |
| `auth.go` | 验证 JWT，识别当前用户 |
| `permission.go` | 验证接口权限码 |
| `cors.go` | 跨域处理 |
| `logger.go` | HTTP 请求日志 |
| `operation_log.go` | 操作日志 |
| `recovery.go` | 捕获 panic，避免服务直接崩溃 |
| `traceid.go` | 生成并传递请求追踪 ID |

学习 CRUD 时，先理解 `auth.go` 和 `permission.go` 即可，其余中间件按需深入。

### 5.3 `handler/`：HTTP 接口层

Handler 负责：

```text
接收路径参数、查询参数、JSON Body
        ↓
参数绑定与校验
        ↓
调用 Service
        ↓
成功使用 response.Success
失败使用 apperr.Write
```

商品 Handler 的接口分布：

| 方法 | 接口 |
| --- | --- |
| `ProductHandler.List` | `GET /api/product` |
| `ProductHandler.Get` | `GET /api/product/:id` |
| `ProductHandler.Create` | `POST /api/product` |
| `ProductHandler.Update` | `PUT /api/product/:id` |
| `ProductHandler.Delete` | `DELETE /api/product/:id` |
| `ProductHandler.BatchDelete` | `DELETE /api/product` |

### 5.4 `service/`：业务逻辑层

Service 负责：

- 参数规范化。
- 数据是否存在等业务判断。
- 将数据库错误翻译为业务错误。
- 组合多个 Repository。
- 复杂业务流程和事务。

例如商品详情会把 `gorm.ErrRecordNotFound` 翻译为“商品不存在”的 404 错误。

`service/auth.go` 除认证业务外，还包含权限码、菜单等初始化种子。新增模块的权限和菜单需要在这里登记。

### 5.5 `repository/`：数据库访问层

`repository/generic.go` 提供通用能力：

| 方法 | 作用 |
| --- | --- |
| `FindByID` | 按 ID 查询 |
| `List` | 分页列表 |
| `ListAll` | 查询全部 |
| `Create` | 新增 |
| `Update` | 全量更新 |
| `Delete` | 单条删除 |
| `BatchDelete` | 批量删除 |

商品 Repository 嵌入通用实现：

```go
type productRepository struct {
    *GenericRepository[model.Product]
}
```

因此只需在 `repository/product.go` 中实现商品特有的名称和状态过滤，基础 CRUD 由 `GenericRepository` 提供。

### 5.6 `model/`：数据库实体

`model/base.go` 定义公共字段：

```go
type Model struct {
    ID        uint
    CreatedAt time.Time
    UpdatedAt time.Time
    CreatedBy uint
    UpdatedBy uint
    DeletedAt gorm.DeletedAt
    DeletedBy uint
}
```

业务 Model 嵌入 `Model`，自动获得主键、时间戳、审计字段和软删除能力。

`model/product.go` 定义商品业务字段，并通过 `TableName()` 映射到 `products` 表。`model/rbac.go` 中的 `AutoMigrate` 负责将各 Model 注册到数据库自动迁移流程。

## 6. `pkg/`：通用基础能力

```text
internal/pkg/
├── apperr/                     应用错误和 HTTP 状态映射
├── response/
│   ├── response.go             成功响应 ApiResult
│   └── problem.go              RFC 7807 错误响应
├── pagination/                 通用分页参数、结果和 COUNT/LIMIT
├── audit/                      创建人、修改人等 GORM 审计回调
├── jwt/                        JWT 生成与验证
├── hash/                       密码哈希
├── log/                        日志初始化
├── async/                      异步任务辅助
├── csvutil/                    CSV 导出
├── datascope/                  数据权限范围
└── gormx/                      GORM 辅助和错误处理
```

推荐学习顺序：

```text
response → apperr → pagination → audit → 其他包按需阅读
```

`pkg` 不应依赖具体的 `product`、`user` 等业务模块。

## 7. 商品模块的目录分布

一个完整业务模块通常分布在多个目录，而不是集中在一个文件中：

```text
server/
├── cmd/api/main.go
│   └── 创建 ProductRepository → ProductService → ProductHandler
├── internal/
│   ├── model/
│   │   ├── product.go           商品表结构
│   │   └── rbac.go              把 Product 加入 AutoMigrate
│   ├── repository/
│   │   ├── product.go           商品列表过滤
│   │   └── generic.go           复用通用 CRUD
│   ├── service/
│   │   ├── product.go           商品业务逻辑
│   │   └── auth.go              商品权限码和菜单种子
│   ├── handler/
│   │   ├── product.go           商品 HTTP 接口
│   │   └── product_integration_test.go
│   ├── server/
│   │   └── router.go            注册 /api/product 路由
│   └── repository/
│       └── product_test.go      商品 Repository 测试
└── docs/
    └── swagger.*                商品接口文档
```

一句话记忆：

```text
model 定义数据
repository 保存和查询数据
service 处理业务
handler 接收请求
router 注册地址
main.go 把各层组装起来
```

## 8. 前端目录简图

```text
web/src/
├── app/                        应用骨架
│   ├── components/             SearchTable、FormDrawer 等公共组件
│   ├── composables/            useCrud 等通用组合函数
│   ├── constants/              权限码、枚举
│   ├── directives/             v-permission 等指令
│   └── stores/                 Pinia 全局状态
├── layout/                     后台整体布局
├── lib/                        前端基础设施
│   ├── http/                   Axios、Token、响应和错误处理
│   ├── auth/                   登录状态
│   ├── router/                 路由守卫、动态路由
│   └── i18n/                   国际化
├── modules/                    业务模块
│   ├── auth/
│   ├── dashboard/
│   ├── product/                商品管理
│   ├── crud/                   原 CRUD 示例
│   ├── system/
│   ├── profile/
│   └── about/
├── mock/                       MSW Mock
├── router/                     基础路由
├── shared/                     可共享代码
├── assets/                     静态资源
├── App.vue
└── main.ts
```

商品前后端接口的衔接位置：

```text
web/src/modules/product/api.ts
              ↓ HTTP
server/internal/server/router.go
              ↓
server/internal/handler/product.go
```

## 9. 代码依赖方向

业务代码按下面的方向依赖：

```text
Handler → Service → Repository → Model
```

各层都可以复用 `internal/pkg` 的基础能力，但 `pkg` 不反向依赖业务层。

`cmd/api/main.go` 位于最外层，负责创建具体实现并进行依赖注入。因此：

- 阅读请求时，从 Router 向 Handler、Service、Repository 跟踪。
- 阅读启动和对象创建时，从 `main.go` 向各层查看装配。

## 10. 推荐后端阅读顺序

### 第一遍：商品业务主线

```text
model/product.go
    ↓
server/router.go
    ↓
handler/product.go
    ↓
service/product.go
    ↓
repository/product.go
    ↓
repository/generic.go
```

### 第二遍：通用支撑

```text
model/base.go
pkg/response/
pkg/apperr/
pkg/pagination/
pkg/audit/
```

### 第三遍：启动、配置和接口保护

```text
cmd/api/main.go
internal/config/
internal/db/
middleware/auth.go
middleware/permission.go
service/auth.go
```

新增其他业务模块时，配合 [新增业务模块指南](./new-module-guide.md) 阅读；原 CRUD 示例的完整请求流程见 [CRUD 示例：前后端完整实现流程](./crud-flow.md)。
