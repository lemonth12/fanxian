# 京东返现平台 - 设计文档

## 概述

打造一个面向多用户的京东购买返现平台（Web版）。用户粘贴京东商品链接，系统转换为京东联盟推广链接，用户通过推广链接下单后，系统获取京东联盟佣金并按固定比例返还给用户。

架构：**前后端分离**。后端 Gin 提供 REST API（JSON），前端 Vue 3 SPA 调用 API，独立开发和部署。

适用规模：<100 用户的小团队/个人使用。替代市面上高抽成的第三方返现平台。

## 技术选型

| 层 | 技术 |
|---|------|
| 后端框架 | Gin（纯 REST API，JSON 响应） |
| ORM | GORM |
| 数据库 | SQLite（开启WAL模式，提升并发读） |
| 认证 | JWT（Authorization: Bearer 头） |
| CORS | Gin CORS 中间件 |
| 定时任务 | in-process ticker（Gin启动时启动goroutine） |
| 前端框架 | Vue 3（Composition API） |
| 前端构建 | Vite |
| 前端路由 | Vue Router 4 |
| HTTP 客户端 | Axios |
| UI | 纯 CSS（不使用组件库，保持轻量） |

## 项目结构

```
backend/
  cmd/server/main.go          # 入口
  internal/
    config/config.go           # 配置加载（YAML文件 + 环境变量覆盖）
    handler/                   # Gin API handler（JSON 响应）
      auth_handler.go          # 注册/登录/刷新Token
      product_handler.go       # 链接转换（含限流）
      order_handler.go         # 订单列表/余额
    middleware/
      auth.go                  # JWT认证中间件（读Authorization头）
      cors.go                  # CORS中间件
      ratelimit.go             # 转链接口限流
    service/
      auth_service.go          # 注册/登录业务
      product_service.go       # 链接解析 + 转链业务
      order_service.go         # 订单同步/返现计算（事务）
      cron_service.go          # 定时任务调度
    repository/
      user_repo.go             # 用户表操作
      order_repo.go            # 订单表操作
    model/
      user.go                  # User模型
      order.go                 # Order模型
    jd/
      client.go                # 京东联盟API客户端（签名、调用、错误码处理）
      parser.go                # 商品链接解析（正则提取商品ID/SKU）
  config.yaml

frontend/
  index.html                   # Vite 入口 HTML
  vite.config.js
  package.json
  src/
    main.js                    # Vue 应用入口
    App.vue                    # 根组件（layout）
    api/
      client.js                # Axios 实例（baseURL、拦截器、Token刷新）
      auth.js                  # 注册/登录/登出 API
      product.js               # 链接转换 API
      order.js                 # 订单列表/余额 API
    router/
      index.js                 # Vue Router 路由配置 + 导航守卫
    views/
      LoginView.vue            # 登录页
      RegisterView.vue         # 注册页
      ConvertView.vue          # 转链页（首页）
      OrderListView.vue        # 我的订单
    components/
      AppHeader.vue            # 顶部导航（用户信息、登出）
    stores/
      auth.js                  # Pinia store（Token、用户状态）
    assets/
      style.css
```

**分层约定（后端）：**

- `handler` — 参数校验、调用service、返回JSON，不直接操作数据库
- `service` — 业务逻辑，不碰HTTP；涉及资金的操作必须用事务
- `repository` — 数据库操作，不碰业务
- `jd/client` — 封装京东联盟API调用、签名、错误码分类
- `jd/parser` — 商品链接解析逻辑

**分层约定（前端）：**

- `views/` — 路由对应的页面组件
- `components/` — 可复用的UI组件
- `api/` — API请求函数，封装Axios调用
- `stores/` — Pinia状态管理（登录态、用户信息）
- `router/` — 路由配置 + 导航守卫（未登录重定向登录页）

## 配置文件

使用 `config.yaml` 文件，支持环境变量覆盖：

```yaml
server:
  port: 8080
  mode: debug          # debug / release

database:
  path: ./data.db
  wal: true            # 开启WAL模式

jwt:
  access_expire: 2h
  refresh_expire: 168h # 7天
  secret: ${JWT_SECRET}

cors:
  allowed_origins:
    - http://localhost:5173    # Vite 开发服务器

jd_union:
  app_key: ${JD_APP_KEY}
  app_secret: ${JD_APP_SECRET}
  site_id: ${JD_SITE_ID}           # 推广位ID
  pid: ${JD_PID}                   # 联盟PID，子PID在此基础上生成

cashback:
  default_rate: 0.7                # 全局返现比例（0.7 = 返70%佣金）

rate_limit:
  convert_per_minute: 10           # 每用户每分钟最多转链次数
```

配置加载：`internal/config/config.go` 读取YAML，敏感值通过 `${ENV_VAR}` 格式从环境变量注入，启动时校验必填项（jd_union相关配置），缺失则终止启动并提示。

## 数据库设计

### users 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 自增主键 |
| username | TEXT UNIQUE NOT NULL | 用户名，登录用 |
| password_hash | TEXT NOT NULL | bcrypt哈希 |
| sub_pid | TEXT NOT NULL | 京东联盟子PID，格式 `{pid}_{user_id}`，注册时生成 |
| total_earned | REAL DEFAULT 0 | 累计已返现金额（冗余，避免每次SUM计算） |
| created_at | DATETIME | |
| updated_at | DATETIME | |

子PID生成规则：`{京东联盟PID}_{user_id}`。例如联盟PID为 `jd_123456`，用户ID为 `1`，则子PID为 `jd_123456_1`。子PID在用户注册时自动创建，后续所有转链和订单追踪通过此子PID关联。

### orders 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 自增主键 |
| user_id | INTEGER FK NOT NULL | 关联用户 |
| sub_pid | TEXT NOT NULL | 匹配订单归属的子PID |
| jd_order_id | TEXT NOT NULL UNIQUE | 京东订单号（唯一约束，防重） |
| sku_id | TEXT | 商品SKU ID |
| product_name | TEXT | 商品名称 |
| product_url | TEXT | 原始商品链接 |
| estimated_price | REAL | 商品面价 |
| actual_price | REAL | 实际付款金额（用户可能用券/改规格） |
| jd_commission_rate | REAL | 京东原始佣金比例（品类决定，方便对账） |
| commission_amount | REAL DEFAULT 0 | 京东实际结算佣金 |
| cashback_amount | REAL DEFAULT 0 | 返还用户的金额 = commission_amount × cashback_rate |
| cashback_rate | REAL NOT NULL | 返现时的全局比例快照（下单时从config读入） |
| status | TEXT DEFAULT 'pending' | pending/confirmed/settled/invalid |
| order_time | DATETIME | 京东下单时间 |
| settled_at | DATETIME | 实际结算时间（京东确认佣金后） |
| created_at | DATETIME | |
| updated_at | DATETIME | |

## 订单状态机

```
  pending ──→ confirmed ──→ settled
     │            │
     └──→ invalid └──→ invalid
```

| 状态 | 含义 | 触发条件 |
|------|------|----------|
| pending | 待确认（用户刚下单，京东未确认） | 用户通过推广链接下单 |
| confirmed | 已确认（京东确认收货，等待结算佣金） | 定时任务拉取到已完成订单 |
| settled | 已结算（京东已结算佣金，可返现） | 定时任务拉取到已结算佣金 |
| invalid | 已失效（拒收/退货/退款） | 订单被取消或退款 |

**京东结算周期说明：** 用户确认收货后，京东联盟通常需要30-60天结算佣金。因此 `confirmed` → `settled` 之间有较长延迟。前端需向用户说明这一账期。

## 核心流程

### 流程一：用户注册 → 创建子PID

```
用户填写用户名密码 → 创建User记录 → 生成sub_pid（{pid}_{user_id}）
→ 调用京东联盟API（jd.union.open.user.pid.get）创建子推广位 → 完成注册
```

### 流程二：用户转链下单

```
用户粘贴商品链接 → parser解析提取商品ID/SKU
→ 调用京东联盟转链API（jd.union.open.goods.link.query，带sub_pid）
→ 返回推广链接 → 用户复制 → 去京东下单
→ （可选）写入pending订单记录，标记预估金额和佣金比例
```

**链接解析策略（`internal/jd/parser.go`）：**

京东链接格式多样，按优先级匹配：

| 格式 | 示例 | 提取方式 |
|------|------|----------|
| PC商品页 | `item.jd.com/123456.html` | 正则 `/item\.jd\.com/(\d+)\.html` |
| 移动端商品页 | `item.m.jd.com/product/123456.html` | 正则 `/product/(\d+)\.html` |
| 短链接 | `3.cn/xxxxx` | HTTP HEAD请求获取重定向URL后解析 |
| 带参数长链接 | `item.jd.com/123456.html?xxx` | 同上正则（先strip参数） |

不支持或无法解析的链接返回明确错误提示。

### 流程三：返现结算（定时任务）

```
定时任务（每小时） → 调用京东联盟订单查询API
→ 查询时间窗口：上次成功同步时间 至 当前时间（增量拉取）
→ 遍历返回订单，按sub_pid匹配本地用户
→ UPSERT写入订单表（jd_order_id唯一约束防重）
→ 如果订单状态变为settled且本地为confirmed：
    → 在事务中：更新订单状态 + 更新佣金金额 + 累加用户total_earned
→ 记录本次同步窗口，作为下次查询起点
```

**定时任务实现：**
- Go `time.Ticker`，在main.go启动时作为goroutine运行
- 默认每小时执行一次
- 每次同步记录窗口 `[last_sync_time, now]`
- 使用 `INSERT ... ON CONFLICT(jd_order_id) DO UPDATE` 实现幂等同步
- 单条订单处理失败不影响其他订单（continue to next）

## 页面 & 路由

Vue Router 4 路由设计：

| 路径 | 页面 | 说明 |
|------|------|------|
| `/login` | LoginView | 登录页（未登录入口） |
| `/register` | RegisterView | 注册页 |
| `/` | ConvertView | 转链页（首页，需登录） |
| `/orders` | OrderListView | 我的订单（需登录） |

**导航守卫：** 未登录用户访问需登录页面 → 重定向 `/login`。已登录用户访问 `/login` → 重定向 `/`。

**页面详情：**

1. **LoginView** — 用户名、密码、登录按钮、去注册链接。成功后 store 存储 token，跳转首页
2. **RegisterView** — 用户名、密码（最少6位）、确认密码、注册按钮
3. **ConvertView（首页）** — 输入框粘贴京东商品链接、转换按钮、展示推广链接（一键复制按钮）、展示预估可返金额（按品类最低佣金率估算，提示"以实际结算为准"）
4. **OrderListView** — 表格：商品名、下单时间、订单状态（含状态说明）、佣金、返现金额；顶部显示累计可返现金额

**App.vue 布局：** 顶部 `AppHeader`（已登录时显示用户名 + 登出按钮），下方 `<router-view>` 渲染页面。

## API 设计

后端返回统一 JSON 格式：

```json
// 成功
{ "code": 0, "msg": "ok", "data": { ... } }

// 失败
{ "code": 40001, "msg": "用户名已存在", "data": null }
```

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/auth/register` | 注册 | 否 |
| POST | `/api/auth/login` | 登录，返回 access_token + refresh_token | 否 |
| POST | `/api/auth/refresh` | 刷新 Token | 否（用 refresh_token） |
| POST | `/api/product/convert` | 转链，Body: `{url: "..."}` | 是 |
| GET | `/api/orders` | 订单列表 + 累计余额 | 是 |
| GET | `/api/orders/summary` | 累计可返现金额 | 是 |

## 京东联盟API

| 接口 | 用途 | 调用时机 |
|------|------|----------|
| `jd.union.open.user.pid.get` | 创建子推广位 | 用户注册时 |
| `jd.union.open.goods.link.query` | 商品链接转推广链接 | 用户粘贴链接点击转换 |
| `jd.union.open.order.query` | 查询推广订单明细 | 定时任务同步 |
| `jd.union.open.order.bonus.query` | 查询奖励订单 | 定时任务补充同步 |

**API鉴权：** 京东联盟使用签名鉴权，请求参数按字典序排序后拼接 AppSecret 生成 MD5 sign。统一封装在 `internal/jd/client.go`。

**API错误处理与重试：**

| 错误类型 | 示例 | 策略 |
|----------|------|------|
| 可重试 | 网络超时、限流 | 最多重试3次，指数退避（1s/2s/4s） |
| 不可重试 | 签名错误、参数错误 | 立即失败，记录日志 |
| 批量同步中单条失败 | — | 记录日志，continue 下一条，不中断整批 |

**频率限制：** 京东联盟API有QPS限制。转链接口对同一用户限流（每用户每分钟最多10次），定时任务控制调用频率。

## 认证与安全

### JWT方案

- Access Token 有效期 2 小时，Refresh Token 有效期 7 天
- Token 通过 HTTP `Authorization: Bearer <token>` 头传递
- 登录成功返回双 Token，前端存入 Pinia store + localStorage
- 后端 `/api/auth/refresh` 端点接受 refresh_token，返回新 access_token
- 前端 Axios 响应拦截器：收到 401 时自动调用 refresh 接口续期，续期失败则清除登录态跳转登录页
- 登出时清除前端 Token（短期不做服务端黑名单，2h 过期即失效）

### CORS 配置

- Gin CORS 中间件，只允许 `config.yaml` 中配置的 `allowed_origins`
- 允许 `Authorization`、`Content-Type` 请求头
- 开发环境允许 `localhost:5173`，生产环境配置实际前端域名

### 密码策略

- 最小长度 6 位
- 使用 bcrypt 哈希存储

### 转链接口防滥用

- 每用户每分钟最多 10 次转链请求
- 超过限流返回 HTTP 429，Body: `{"code": 42900, "msg": "操作太频繁，请稍后再试"}`
- 后端限流中间件使用用户ID（已登录用户）或 IP（未登录请求）作为 key

## 事务管理

涉及资金的操作必须在事务中执行：

```
// 订单结算时的原子操作
tx := db.Begin()
tx.UpdateOrderStatus(orderID, "settled", commissionAmount, cashbackAmount)
tx.IncrementUserTotalEarned(userID, cashbackAmount)
tx.Commit()
```

失败时回滚，确保订单状态和用户余额一致。

## 错误处理

后端统一返回 JSON 错误：

- 京东API调用失败 → `{"code": 50001, "msg": "转链服务暂时不可用，请稍后重试"}` + 记录错误日志
- API签名错误 → 记录详细日志（错误码、请求参数摘要），不暴露密钥到日志
- 用户重复注册 → `{"code": 40001, "msg": "用户名已存在"}`
- 无效商品链接 → `{"code": 40002, "msg": "无法识别该链接，请粘贴京东商品链接"}`
- 未登录访问 → JWT中间件拦截，返回 401 `{"code": 40100, "msg": "未登录"}`
- Token过期 → 返回 401 `{"code": 40101, "msg": "Token已过期"}`
- 数据库查询错误 → 记录日志，返回 500 `{"code": 50000, "msg": "服务器内部错误"}`
- 参数校验失败 → 返回 400 `{"code": 40000, "msg": "参数错误"}`
- 定时任务失败 → 记录日志，下次执行时重试（增量窗口不会丢数据）
- 启动时配置缺失 → 终止启动，明确提示缺失项

## 测试策略

- **单元测试：** service层逻辑（返现计算、链接解析）；边界情况：佣金为0、退款订单（commission_amount为负）、部分退货
- **集成测试：** repository层用SQLite内存库测试事务、UPSERT幂等
- **Handler测试：** `httptest` 测试Gin路由、JWT中间件、CORS、限流
- **京东API：** mock测试正常返回、限流、签名错误、超时四条路径
- **前端测试：** Vitest 测试组件渲染、Axios mock测试API调用
- **手动验证：** 京东联盟API用实际账号验证一次完整流程

## 后续扩展点（不做，预留空间）

- 后端新增handler + 前端新增路由/页面即可扩展功能
- 用户表加 `openid` 字段对接微信登录
- 前端新增路由页面 + 后端新增handler 即可扩展平台（`pdd/`、`tb/`）
- 提现：后端加 `withdrawals` 表 + API，前端加提现页面
- SQLite换MySQL/PostgreSQL：只需改GORM配置，SQLite的写锁瓶颈自然解决
