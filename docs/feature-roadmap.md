# 短网址功能规划文档

文档状态：Draft
更新时间：2026-05-22
适用范围：`dwz-server` 后端、`admin-webui` 管理端、公开跳转页面与 API。

## 1. 背景

当前项目已经具备短网址服务的核心闭环：短链生成、跳转、统计、域名管理、用户认证、API Token、A/B 测试、二维码生成器和独立部署模式。和 Bitly、Rebrandly、Short.io、Dub 等主流短网址/链接管理平台相比，差距主要不在“能否缩短 URL”，而在营销归因、团队治理、高级路由、风控合规、数据产品化和集成生态。

本规划目标是把项目从“自部署短链管理系统”推进到“可用于营销、增长和企业内控的链接管理平台”。

## 2. 产品定位

### 2.1 核心定位

面向个人、小团队和企业私有化部署的短链接与链接归因系统，强调：

- 数据自持有：点击、地域、来源等数据保存在自有数据库。
- 部署简单：继续保留 SQLite + 内存缓存的独立模式。
- 企业可控：支持多域名、多用户、审计、权限、API 自动化。
- 营销可用：补齐 UTM、活动、二维码、路由、导出和集成能力。

### 2.2 不建议的定位

短期内不建议直接对标 Bitly/Rebrandly 的完整 SaaS 商业平台能力，例如域名购买、托管 DNS、自动 SSL 商业化、企业 SLA、全球边缘网络等。这些能力投入较重，且和当前自部署优势不完全一致。

## 3. 当前能力基线

以下为代码和文档中已经具备或基本具备的能力：

| 模块 | 当前能力 | 依据 |
| --- | --- | --- |
| 短链管理 | 创建、列表、详情、更新、禁用、删除、批量创建、自定义短码、过期时间 | `app/service/shortlink_service.go`、`docs/api/API_REFERENCE.md` |
| 多域名 | 域名 CRUD、启用/禁用、协议、备案信息、参数透传、短码生成策略 | `app/model/domain.go`、`app/service/domain_service.go` |
| 跳转 | 缓存查找、数据库回源、过期/禁用判断、302 跳转、微信/QQ 防红页面 | `app/controller/shortlink_controller.go`、`app/service/shortlink_service.go` |
| 统计 | 点击记录、IP、User-Agent、Referer、Query、国家、省份、城市、ISP、小时/日统计、热门链接 | `app/model/click_statistic.go`、`app/dao/click_statistic_dao.go` |
| A/B 测试 | 多变体、等分/权重分流、会话一致性、启动/停止、变体点击统计 | `app/service/ab_test_service.go` |
| 用户与认证 | 用户、密码、JWT 登录、Bearer Token、HMAC 签名认证、OIDC | `app/model/user.go`、`app/middleware/auth_middleware.go` |
| 操作审计 | 操作日志、中间件记录、敏感字段脱敏 | `app/middleware/operation_log_middleware.go` |
| 二维码 | 管理端前端二维码生成、样式配置、Logo、图层、PNG/JPG 下载 | `admin-webui/apps/web-antd/src/components/QRCodeGenerator` |
| 部署 | MySQL/PostgreSQL/SQLite，Redis/内存缓存，Docker，独立部署 | `README.md`、`config.yaml.example` |

## 4. 对标差异

| 方向 | 主流服务常见能力 | 当前差距 | 优先级 |
| --- | --- | --- | --- |
| 营销归因 | UTM Builder、活动 Campaign、渠道、标签 Tag、文件夹、UTM 维度报表 | 当前只有查询参数透传，没有归因实体和报表维度 | P0 |
| 团队治理 | 工作区 Workspace、团队成员、角色权限、资源归属、邀请、成员审计 | 用户模型无角色，短链无创建者归属，登录后基本是全局管理 | P0 |
| 统计分析 | 设备、浏览器、OS、机器人 Bot 过滤、UTM 分析、实时事件、数据导出 | 当前有 IP/地域/Referer/时间，但没有设备解析和导出 API | P0 |
| 二维码产品化 | QR 实体、动态 QR、扫描统计、SVG/PDF、批量 QR、API 生成 | 当前是前端即时生成图片，没有后端实体和扫码独立统计 | P1 |
| 高级路由 | Geo targeting、device targeting、browser/language/referrer routing、fallback URL | 当前只有 A/B 分流，缺少条件路由和 fallback | P1 |
| 链接安全 | 访问密码、黑白名单、恶意 URL 扫描、滥用举报 Abuse report、机器人 Bot 拦截 | 当前只有登录限流和操作审计，没有链接访问级安全 | P1 |
| 集成生态 | Webhook、SDK、Zapier/Slack/GA/Segment、浏览器插件 | 当前有 API，但没有 Webhook、SDK、第三方集成 | P2 |
| Link-in-bio | Bio Page、链接集合页、主题和访问统计 | 当前没有落地页/集合页构建器 | P2 |
| 域名运维 | DNS 验证、SSL 状态、域名健康检查、域名分组 | 当前只管理域名记录，不校验 DNS/SSL 状态 | P2 |
| 合规治理 | 数据保留策略、IP 匿名化、隐私开关、导出/删除审计数据 | 当前没有数据生命周期和隐私配置 | P2 |

## 5. 规划原则

1. 先补平台地基，再补营销花活：优先做权限、归属、归因和统计维度。
2. 所有新功能都要兼容独立部署模式：SQLite + 内存缓存不可被放弃。
3. 不引入过早的重型平台能力：DNS 托管、域名购买、全球边缘节点暂不做。
4. 数据模型要为后续 SaaS 化留口：即使当前自部署，也应引入工作区 workspace 和资源所有者 resource owner。
5. API 优先：管理端功能应尽量由稳定 API 驱动，方便后续集成和自动化。

## 6. 阶段路线图

### Phase 0：基线修正与文档对齐

目标：修正明显不一致，降低后续迭代风险。

- 修正 API 文档中的 `domain` 示例：当前后端要求域名不带协议，协议来自域名配置。
- 明确 README 中“导出功能”的真实状态：若暂无导出接口，应改为规划项。
- 增加功能矩阵文档：区分已实现、部分实现、规划中。
- 为短链、域名、统计、A/B 测试补充最小回归测试。

交付物：

- `docs/api/API_REFERENCE.md` 更新。
- `docs/feature-roadmap.md` 保持维护。
- 新增基础测试用例。

### Phase 1：多用户归属与权限体系

目标：让系统具备团队协作和企业内控基础。

后端规划：

- 新增 `workspaces` 表：工作区名称、状态、创建人。
- 新增 `workspace_members` 表：用户、工作区、角色。
- 新增 `roles` 或枚举角色：所有者 owner、管理员 admin、成员 member、只读 viewer。
- 为 `short_links`、`domains`、`ab_tests`、`user_tokens` 增加 `workspace_id`。
- 为 `short_links` 增加 `created_by`、`updated_by`。
- AuthMiddleware 后增加权限检查中间件或服务层 Policy。

管理端规划：

- 工作区切换器。
- 成员管理页面。
- 角色权限提示。
- 短链列表默认只展示当前工作区数据。

API 规划：

- `GET /api/v1/workspaces`
- `POST /api/v1/workspaces`
- `GET /api/v1/workspaces/:id/members`
- `POST /api/v1/workspaces/:id/invitations`
- `PUT /api/v1/workspaces/:id/members/:user_id/role`

验收标准：

- 普通成员不能操作其他工作区资源。
- Viewer 只能查看，不能创建、更新、删除。
- 操作日志记录工作区和操作者。

### Phase 2：营销归因与链接组织

目标：补齐营销使用中最常见的 UTM、活动、标签和组织能力。

后端规划：

- 新增 `campaigns` 表：名称、描述、时间范围、工作区。
- 新增 `tags` 表和 `short_link_tags` 关联表。
- 为短链增加营销字段：
  - `campaign_id`
  - `utm_source`
  - `utm_medium`
  - `utm_campaign`
  - `utm_term`
  - `utm_content`
  - `notes`
- 创建短链时支持 UTM Builder，将 UTM 合并到目标 URL。
- 点击统计增加 UTM 维度解析，保存到统计表或事件扩展表。

管理端规划：

- 创建/编辑短链增加 UTM Builder。
- 增加活动 Campaign 管理页面。
- 短链列表支持按活动 Campaign、标签 Tag、创建人筛选。
- 统计页增加 UTM 维度报表。

API 规划：

- `POST /api/v1/campaigns`
- `GET /api/v1/campaigns`
- `POST /api/v1/tags`
- `GET /api/v1/tags`
- `GET /api/v1/click_statistics/utm-analysis`

验收标准：

- 创建短链时可生成带 UTM 的目标 URL。
- 同一活动 Campaign 下可以查看链接汇总点击、唯一 IP、来源、地域。
- API 和管理端都能按标签筛选短链。

### Phase 3：统计增强与数据导出

目标：把统计从“点击计数”升级到可分析、可导出的数据产品。

后端规划：

- 引入 User-Agent 解析库，提取：
  - device_type：desktop、mobile、tablet、bot、unknown
  - browser
  - os
  - bot_name
- 点击统计增加设备、浏览器、OS、机器人 Bot 字段。
- 支持机器人 Bot 过滤开关：全量、排除 Bot、只看 Bot。
- 增加 CSV 导出接口。
- 增加聚合报表接口，避免前端自行拼装。
- 增加数据保留策略配置：保留天数、IP 匿名化。

管理端规划：

- 统计页增加设备、浏览器、OS、机器人 Bot 分布。
- 点击明细支持 CSV 导出。
- 报表筛选支持日期范围、活动 Campaign、标签 Tag、设备、来源。

API 规划：

- `GET /api/v1/click_statistics/export`
- `GET /api/v1/click_statistics/device-analysis`
- `GET /api/v1/click_statistics/browser-analysis`
- `GET /api/v1/reports/links`
- `GET /api/v1/reports/campaigns`

验收标准：

- 统计 API 可按设备、浏览器、OS 分组。
- 导出文件可被 Excel/Numbers 正常打开。
- 机器人 Bot 流量可单独查看或排除。

### Phase 4：二维码产品化

目标：把现有前端二维码工具升级为可管理、可统计、可 API 化的二维码产品。

后端规划：

- 新增 `qr_codes` 表：
  - `short_link_id`
  - `workspace_id`
  - `name`
  - `config_json`
  - `format`
  - `status`
  - `created_by`
- 新增 QR 生成 API，支持 PNG、JPG、SVG。
- 点击统计中标记访问来源：link 或 qr。
- 可选：独立 QR 扫描事件表。

管理端规划：

- 二维码保存为模板。
- 二维码列表、复制、下载、更新。
- QR 扫描统计与短链点击统计区分。
- 批量生成二维码。

API 规划：

- `POST /api/v1/qr_codes`
- `GET /api/v1/qr_codes`
- `GET /api/v1/qr_codes/:id/download?format=png`
- `PUT /api/v1/qr_codes/:id`
- `DELETE /api/v1/qr_codes/:id`

验收标准：

- 管理端生成的二维码可保存并再次编辑。
- 同一短链多个 QR 可以区分扫描数据。
- API 可直接生成并下载 QR 图片。

### Phase 5：高级路由与链接保护

目标：让短链具备场景化跳转能力。

后端规划：

- 新增 `link_routes` 表：
  - `short_link_id`
  - `priority`
  - `condition_type`
  - `condition_value`
  - `target_url`
  - `is_active`
- 支持条件：
  - country、province、city
  - device_type
  - browser
  - os
  - language
  - referer
  - query_param
- 为短链增加：
  - `fallback_url`
  - `redirect_code`，支持 301/302/307/308
  - `password_hash`
  - `max_clicks`
  - `access_window_start`
  - `access_window_end`
- 访问密码页模板。

管理端规划：

- 路由规则配置器。
- 链接密码配置。
- 最大访问次数和访问时间窗口配置。
- 路由命中统计。

API 规划：

- `POST /api/v1/short_links/:id/routes`
- `GET /api/v1/short_links/:id/routes`
- `PUT /api/v1/short_links/:id/routes/:route_id`
- `DELETE /api/v1/short_links/:id/routes/:route_id`
- `POST /api/v1/short_links/:id/password`

验收标准：

- 同一短链可以按国家或设备跳转到不同 URL。
- 带访问密码的短链未验证时不泄露目标 URL。
- 规则冲突时按优先级稳定命中。

### Phase 6：安全风控与合规

目标：降低短链被滥用的风险，并提升企业部署可信度。

后端规划：

- 增加目标 URL 安全检查抽象接口。
- 支持本地黑名单/白名单：
  - domain blocklist
  - domain allowlist
  - keyword blocklist
- 增加 abuse report：
  - 公开举报页面
  - 后台处理状态
  - 自动禁用策略
- 增加 IP 匿名化配置。
- 增加数据保留策略任务。
- 增加审计日志导出。

管理端规划：

- 风控设置页。
- 黑白名单管理。
- 举报处理工作台。
- 数据保留和隐私设置。

API 规划：

- `POST /api/v1/abuse_reports`
- `GET /api/v1/admin/abuse_reports`
- `PUT /api/v1/admin/abuse_reports/:id`
- `GET /api/v1/security/blocklist`
- `POST /api/v1/security/blocklist`

验收标准：

- 命中黑名单域名时不能创建短链。
- 举报后可在后台处理，并可禁用对应短链。
- 可配置统计数据保留天数。

### Phase 7：Webhook 与集成生态

目标：提升自动化和外部系统集成能力。

后端规划：

- 新增 `webhooks` 表：
  - URL
  - secret
  - event types
  - status
  - retry policy
- 支持事件：
  - link.created
  - link.updated
  - link.clicked
  - qr.scanned
  - campaign.completed
  - abuse_report.created
- HMAC 签名 webhook payload。
- 失败重试和投递日志。
- 发布官方 OpenAPI Spec。

管理端规划：

- Webhook 管理页。
- 投递日志和重放按钮。
- API 文档入口。

API 规划：

- `POST /api/v1/webhooks`
- `GET /api/v1/webhooks`
- `GET /api/v1/webhooks/:id/deliveries`
- `POST /api/v1/webhooks/:id/test`

验收标准：

- 外部系统可收到短链创建和点击事件。
- Webhook payload 可验签。
- 失败投递可重试和查看错误。

### Phase 8：Link-in-bio 与轻量落地页

目标：面向创作者、活动页、私域运营补齐链接集合场景。

后端规划：

- 新增 `pages` 表：slug、标题、主题、状态、工作区。
- 新增 `page_links` 表：页面下的链接项。
- 页面访问统计。
- 页面主题配置 JSON。

管理端规划：

- Bio Page 编辑器。
- 链接排序、启用/禁用。
- 页面主题选择。
- 页面统计。

API 规划：

- `POST /api/v1/pages`
- `GET /api/v1/pages`
- `PUT /api/v1/pages/:id`
- `POST /api/v1/pages/:id/links`

验收标准：

- 可创建公开链接集合页。
- 页面访问和单个链接点击可统计。
- 页面 slug 和短码空间不冲突。

## 7. 推荐优先级

### P0：建议优先启动

1. Phase 0：文档和现状修正。
2. Phase 1：工作区 Workspace、资源归属、权限。
3. Phase 2：活动 Campaign、标签 Tag、UTM Builder。
4. Phase 3：设备解析、机器人 Bot 过滤、导出。

理由：这些能力是企业可用和营销可用的基础，也会影响后续所有数据模型。越晚补，迁移成本越高。

### P1：第二阶段启动

1. Phase 4：二维码产品化。
2. Phase 5：高级路由和链接访问保护。
3. Phase 6：安全风控和合规。

理由：当前项目已有二维码前端和 A/B 路由基础，可以复用现有能力继续深化。

### P2：视资源推进

1. Phase 7：Webhook 和集成生态。
2. Phase 8：Link-in-bio 与轻量落地页。

理由：这些功能有助于扩展使用场景，但依赖前面工作区、权限、统计和事件模型。

## 8. 数据模型草案

### 8.1 工作区

```sql
workspaces (
  id BIGINT PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  slug VARCHAR(100) NOT NULL UNIQUE,
  owner_user_id BIGINT NOT NULL,
  status TINYINT DEFAULT 1,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
)
```

```sql
workspace_members (
  id BIGINT PRIMARY KEY,
  workspace_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  role VARCHAR(20) NOT NULL,
  status TINYINT DEFAULT 1,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  UNIQUE (workspace_id, user_id)
)
```

### 8.2 活动与标签

```sql
campaigns (
  id BIGINT PRIMARY KEY,
  workspace_id BIGINT NOT NULL,
  name VARCHAR(120) NOT NULL,
  description TEXT,
  start_at TIMESTAMP,
  end_at TIMESTAMP,
  status VARCHAR(20) DEFAULT 'active',
  created_by BIGINT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
)
```

```sql
tags (
  id BIGINT PRIMARY KEY,
  workspace_id BIGINT NOT NULL,
  name VARCHAR(50) NOT NULL,
  color VARCHAR(20),
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  UNIQUE (workspace_id, name)
)
```

```sql
short_link_tags (
  short_link_id BIGINT NOT NULL,
  tag_id BIGINT NOT NULL,
  PRIMARY KEY (short_link_id, tag_id)
)
```

### 8.3 点击统计增强

建议为 `click_statistics` 增加字段：

- `workspace_id`
- `campaign_id`
- `utm_source`
- `utm_medium`
- `utm_campaign`
- `utm_term`
- `utm_content`
- `device_type`
- `browser`
- `os`
- `is_bot`
- `bot_name`
- `entry_type`：link、qr、page

### 8.4 二维码

```sql
qr_codes (
  id BIGINT PRIMARY KEY,
  workspace_id BIGINT NOT NULL,
  short_link_id BIGINT NOT NULL,
  name VARCHAR(120) NOT NULL,
  config_json TEXT,
  status TINYINT DEFAULT 1,
  created_by BIGINT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
)
```

### 8.5 高级路由

```sql
link_routes (
  id BIGINT PRIMARY KEY,
  short_link_id BIGINT NOT NULL,
  priority INT DEFAULT 100,
  condition_type VARCHAR(50) NOT NULL,
  condition_operator VARCHAR(20) DEFAULT 'eq',
  condition_value VARCHAR(500) NOT NULL,
  target_url VARCHAR(2000) NOT NULL,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
)
```

## 9. API 设计原则

- 所有列表接口统一支持 `page`、`page_size`、`keyword`。
- 新增资源必须带 `workspace_id` 上下文，优先从当前工作区解析。
- 报表接口避免返回过多原始数据，默认返回聚合数据。
- 导出接口走异步任务更稳妥，大数据量场景返回任务 ID。
- Webhook、API Token、签名认证共用 HMAC 工具链。

## 10. 管理端导航建议

建议后续管理端主导航调整为：

- 概览
- 短网址
  - 链接管理
  - 批量创建
  - 高级路由
- 活动与归因
  - 活动 Campaign
  - 标签 Tag
  - UTM 模板
- 二维码
  - QR 列表
  - 模板管理
- 统计分析
  - 链接分析
  - 活动分析
  - 点击明细
  - 数据导出
- 域名
- 团队与权限
- 集成
  - API Token
  - Webhook
- 安全与合规
  - 黑白名单
  - 举报处理
  - 数据保留
- 系统设置

## 11. 风险与注意事项

### 11.1 数据迁移风险

引入 `workspace_id` 会影响大部分主表和查询条件。应先做默认工作区迁移，再逐步改业务查询。

### 11.2 统计表增长风险

点击统计会快速增长。建议尽早设计：

- 按日期索引。
- 聚合表。
- 数据保留策略。
- 大表导出异步化。

### 11.3 SQLite 兼容风险

当前项目支持 SQLite，新增 SQL 需要同时提供 MySQL、PostgreSQL、SQLite 迁移脚本，避免使用单一数据库特性。

### 11.4 隐私风险

IP、User-Agent、地理位置、UTM 可能涉及隐私合规。建议新增配置：

- 是否记录完整 IP。
- 是否匿名化 IP。
- 点击明细保留天数。
- 是否记录 Query Params。

### 11.5 README 与实际实现不一致

当前 README 提到“导出功能”，但现有后端路由和管理端未看到明确导出接口。后续应补实现或调整文案，避免用户预期落差。

## 12. 对标参考

- Bitly：短链创建支持自定义短 URL、UTM、QR Code，平台包含短链接、QR Codes、landing pages、分析和集成。参考：
  - https://support.bitly.com/hc/en-us/articles/230897128-How-do-I-create-links-with-Bitly
  - https://support.bitly.com/hc/en-us/articles/230895688-What-is-Bitly
- Rebrandly：强调 branded links、custom domains、custom QR codes、short link analytics、link in bio、webhooks、integrations、UTM Builder。参考：
  - https://www.rebrandly.com/
  - https://support.rebrandly.com/hc/en-us/articles/13583667895581-What-are-Rebrandly-Analytics
  - https://support.rebrandly.com/hc/en-us/articles/360017468514-Do-You-Offer-Features-Your-Competitors-Don-t
- Short.io：特性包括 geo-targeting、mobile targeting、password protection、QR code、slug editing。参考：
  - https://www.short.io/features/
  - https://help.short.io/en/articles/4065802-what-is-short-io
  - https://short.io/features/mobile-targeting
- Dub：定位为 link attribution platform，强调 UTM Builder、custom previews、QR Code Design、real-time analytics、webhooks、deferred deep linking、affiliate/referral programs。参考：
  - https://dub.co/
