# API 接口文档

## 概述

本文档描述了短链接服务的所有 API 接口。

**Base URL**: `https://your-domain.com`

**API 版本**: v1

## 认证方式

系统支持两种认证方式：

1. **签名认证（推荐）**：基于 HMAC-SHA256 的安全认证，详见 [签名认证文档](./API_SIGNATURE_AUTH.md)
2. **Bearer Token 认证**：传统 Token 认证，详见 [Bearer Token 文档](./API_BEARER_AUTH.md)

所有受保护接口支持工作区上下文请求头：

```http
X-Workspace-Id: 1
```

未传 `X-Workspace-Id` 时，服务端会自动选择当前用户第一个可用工作区，以兼容旧客户端。

## 通用响应格式

### 成功响应

```json
{
    "code": 0,
    "message": "success",
    "data": { ... }
}
```

### 错误响应

```json
{
    "code": 40001,
    "message": "错误描述"
}
```

### 错误码说明

| 错误码 | 说明 |
|-------|------|
| 0 | 成功 |
| 40001 | 请求参数错误 |
| 40101 | 未授权 |
| 40301 | 禁止访问 |
| 40401 | 资源不存在 |
| 40901 | 资源冲突 |
| 50001 | 服务器内部错误 |

---

## 认证接口

### 用户登录

登录获取访问令牌（用于 Web 管理后台）。

**请求**

```
POST /api/v1/auth/login
```

**请求体**

```json
{
    "username": "admin",
    "password": "password123"
}
```

**响应**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "token": "user_1_abc123...",
        "expires_at": "2024-12-31T23:59:59Z",
        "user": {
            "id": 1,
            "username": "admin",
            "real_name": "管理员",
            "email": "admin@example.com",
            "status": 1
        }
    }
}
```

### 用户登出

**请求**

```
POST /api/v1/auth/logout
```

**响应**

```json
{
    "code": 0,
    "message": "登出成功"
}
```

---

## Token 管理接口

### 创建 Token

创建 API 访问 Token，支持签名认证和 Bearer Token 两种类型。

**请求**

```
POST /api/v1/tokens
```

**请求体**

```json
{
    "token_name": "API Token",
    "token_type": "signature",
    "expire_at": "2025-12-31T23:59:59Z"
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| token_name | string | 是 | Token 名称，最大 100 字符 |
| token_type | string | 否 | Token 类型：`signature`（默认）或 `bearer` |
| expire_at | string | 否 | 过期时间，ISO 8601 格式，不填则永不过期 |

**响应 - 签名认证类型**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 1,
        "token_name": "API Token",
        "token_type": "signature",
        "app_id": "app_1a2b3c4d5e6f7890",
        "app_secret": "a1b2c3d4e5f6789012345678901234567890123456789012345678901234",
        "expire_at": "2025-12-31T23:59:59Z",
        "created_at": "2024-01-01T00:00:00Z"
    }
}
```

**响应 - Bearer Token 类型**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 2,
        "token_name": "Bearer Token",
        "token_type": "bearer",
        "token": "a1b2c3d4e5f6789012345678901234567890123456789012345678901234",
        "expire_at": "2025-12-31T23:59:59Z",
        "created_at": "2024-01-01T00:00:00Z"
    }
}
```

> ⚠️ **重要**：`app_secret` 和 `token` 仅在创建时返回一次，请妥善保存！

### 获取 Token 列表

**请求**

```
GET /api/v1/tokens
```

**查询参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页数量，默认 10，最大 100 |
| token_name | string | 否 | Token 名称（模糊搜索） |
| status | int | 否 | 状态：1-启用，0-禁用 |

**响应**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "list": [
            {
                "id": 1,
                "token_name": "API Token",
                "token_type": "signature",
                "app_id": "app_1a2b3c4d5e6f7890",
                "last_used_at": "2024-01-15T10:30:00Z",
                "expire_at": "2025-12-31T23:59:59Z",
                "status": 1,
                "created_at": "2024-01-01T00:00:00Z"
            },
            {
                "id": 2,
                "token_name": "Bearer Token",
                "token_type": "bearer",
                "token": "a1b2c3d4...",
                "last_used_at": null,
                "expire_at": null,
                "status": 1,
                "created_at": "2024-01-02T00:00:00Z"
            }
        ],
        "pagination": {
            "total": 2,
            "page": 1,
            "size": 10
        }
    }
}
```

### 删除 Token

**请求**

```
DELETE /api/v1/tokens/:token_id
```

**响应**

```json
{
    "code": 0,
    "message": "Token删除成功"
}
```

---

## 短链接接口

### 创建短链接

**请求**

```
POST /api/v1/short_links
```

**请求体**

```json
{
    "original_url": "https://www.example.com/very/long/url",
    "domain": "dwz.do",
    "custom_code": "mylink",
    "title": "示例网站",
    "description": "这是一个示例网站",
    "campaign_id": 1,
    "tag_ids": [1, 2],
    "utm_source": "newsletter",
    "utm_medium": "email",
    "utm_campaign": "spring",
    "notes": "投放备注",
    "expire_at": "2025-12-31T23:59:59Z"
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| original_url | string | 是 | 原始 URL |
| domain | string | 否 | 短链接域名，不填使用默认域名 |
| custom_code | string | 否 | 自定义短码 |
| title | string | 否 | 标题 |
| description | string | 否 | 描述 |
| campaign_id | number | 否 | 活动 Campaign ID |
| tag_ids | array | 否 | 标签 Tag ID 列表 |
| utm_source/utm_medium/utm_campaign/utm_term/utm_content | string | 否 | UTM 参数；服务端会合并到原始 URL query，同名参数以请求字段为准 |
| notes | string | 否 | 内部备注 |
| expire_at | string | 否 | 过期时间 |

**响应**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 1,
        "short_code": "abc123",
        "workspace_id": 1,
        "campaign_id": 1,
        "campaign_name": "spring",
        "tags": [{"id": 1, "name": "推广", "color": "#1677ff"}],
        "domain": "dwz.do",
        "short_url": "https://dwz.do/abc123",
        "original_url": "https://www.example.com/very/long/url?utm_campaign=spring&utm_medium=email&utm_source=newsletter",
        "title": "示例网站",
        "description": "这是一个示例网站",
        "utm_source": "newsletter",
        "utm_medium": "email",
        "utm_campaign": "spring",
        "notes": "投放备注",
        "expire_at": "2025-12-31T23:59:59Z",
        "is_active": true,
        "click_count": 0,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
    }
}
```

### 批量创建短链接

**请求**

```
POST /api/v1/short_links/batch
```

**请求体**

```json
{
    "urls": [
        "https://www.example1.com",
        "https://www.example2.com",
        "https://www.example3.com"
    ],
    "domain": "dwz.do"
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| urls | array | 是 | URL 列表，最多 100 个 |
| domain | string | 否 | 短链接域名 |

**响应**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "success": [
            {
                "id": 1,
                "short_code": "abc123",
                "short_url": "https://dwz.do/abc123",
                "original_url": "https://www.example1.com"
            }
        ],
        "failed": [
            {
                "url": "invalid-url",
                "error": "无效的 URL 格式"
            }
        ]
    }
}
```

### 获取短链接列表

**请求**

```
GET /api/v1/short_links
```

**查询参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页数量，默认 10，最大 100 |
| domain | string | 否 | 域名筛选 |
| keyword | string | 否 | 关键词搜索（搜索短码、标题、原始 URL） |

**响应**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "list": [...],
        "total": 100,
        "page": 1,
        "size": 10
    }
}
```

### 获取短链接详情

**请求**

```
GET /api/v1/short_links/:id
```

**响应**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 1,
        "short_code": "abc123",
        "domain": "dwz.do",
        "short_url": "https://dwz.do/abc123",
        "original_url": "https://www.example.com",
        "title": "示例网站",
        "description": "这是一个示例网站",
        "expire_at": "2025-12-31T23:59:59Z",
        "is_active": true,
        "click_count": 1234,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
    }
}
```

### 更新短链接

**请求**

```
PUT /api/v1/short_links/:id
```

**请求体**

```json
{
    "original_url": "https://www.new-example.com",
    "title": "新标题",
    "description": "新描述",
    "expire_at": "2026-12-31T23:59:59Z",
    "is_active": true
}
```

### 删除短链接

**请求**

```
DELETE /api/v1/short_links/:id
```

**响应**

```json
{
    "code": 0,
    "message": "删除成功"
}
```

### 获取短链接统计

**请求**

```
GET /api/v1/short_links/:id/statistics
```

**查询参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| days | int | 否 | 统计天数，默认 7，最大 365 |

**响应**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total_clicks": 12345,
        "today_clicks": 100,
        "week_clicks": 500,
        "month_clicks": 2000,
        "daily_statistics": [
            {"date": "2024-01-15", "click_count": 100},
            {"date": "2024-01-14", "click_count": 95},
            {"date": "2024-01-13", "click_count": 110}
        ]
    }
}
```

---

## 工作区、活动 Campaign 与标签 Tag

### 工作区

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/v1/workspaces` | 当前用户可用工作区 |
| POST | `/api/v1/workspaces` | 创建工作区并成为所有者 owner |
| PUT | `/api/v1/workspaces/current` | 更新当前工作区 |
| GET | `/api/v1/workspaces/current/members` | 当前工作区成员 |
| POST | `/api/v1/workspaces/current/members` | 添加已有用户到当前工作区 |
| PUT | `/api/v1/workspaces/current/members/:user_id` | 更新成员角色 |
| DELETE | `/api/v1/workspaces/current/members/:user_id` | 移除成员 |

### 活动 Campaign

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/v1/campaigns` | 活动列表 |
| POST | `/api/v1/campaigns` | 创建活动 |
| GET | `/api/v1/campaigns/:id` | 活动详情 |
| PUT | `/api/v1/campaigns/:id` | 更新活动 |
| DELETE | `/api/v1/campaigns/:id` | 删除活动 |
| GET | `/api/v1/reports/campaigns` | 活动报表 |

### 标签 Tag

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/v1/tags` | 标签列表 |
| POST | `/api/v1/tags` | 创建标签 |
| GET | `/api/v1/tags/:id` | 标签详情 |
| PUT | `/api/v1/tags/:id` | 更新标签 |
| DELETE | `/api/v1/tags/:id` | 删除标签，短链不会被删除 |

## 链接安全 Link Security

受保护接口继续使用 `X-Workspace-Id` 工作区上下文；公开接口不需要登录。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/v1/short_links/:id/security` | 获取短链安全配置 |
| PUT | `/api/v1/short_links/:id/security` | 更新访问密码、时间窗、最大访问次数、IP/Bot 策略和举报入口 |
| POST | `/api/v1/short_links/:id/security/rescan` | 使用当前 URL 安全规则重扫短链 |
| GET | `/api/v1/security/url_rules` | URL 安全规则列表 |
| POST | `/api/v1/security/url_rules` | 创建域名/关键词 allow/block 规则 |
| PUT | `/api/v1/security/url_rules/:id` | 更新 URL 安全规则 |
| DELETE | `/api/v1/security/url_rules/:id` | 删除 URL 安全规则 |
| GET | `/api/v1/security/events` | 安全事件列表 |
| GET | `/api/v1/abuse_reports` | 滥用举报列表 |
| PUT | `/api/v1/abuse_reports/:id` | 处理举报，可选择禁用短链 |
| POST | `/api/v1/public/link_access/password` | 公开访问密码验证 |
| POST | `/api/v1/public/abuse_reports` | 公开滥用举报提交 |

短链创建/更新可附带 `security` 对象：

```json
{
  "security": {
    "password": "secret",
    "password_enabled": true,
    "access_window_start": "2026-06-01T00:00:00+08:00",
    "access_window_end": "2026-06-30T23:59:59+08:00",
    "max_clicks": 1000,
    "ip_policy": "off",
    "ip_rules": [{"cidr": "203.0.113.0/24"}],
    "bot_policy": "record_only",
    "report_enabled": true
  }
}
```

短链响应增加 `security_enabled`、`security_summary`、`report_enabled`。短链列表支持 `security_status=none|enabled|password|restricted|url_blocked|reported`。

## 域名管理接口

### 创建域名

**请求**

```
POST /api/v1/domains
```

**请求体**

```json
{
    "domain": "dwz.do",
    "protocol": "https",
    "site_name": "短链接服务",
    "icp_number": "京ICP备12345678号",
    "police_number": "京公网安备12345678号",
    "is_active": true,
    "pass_query_params": false,
    "random_suffix_length": 2,
    "enable_checksum": true,
    "description": "主要短链域名"
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| domain | string | 是 | 域名 |
| protocol | string | 是 | 协议：`http` 或 `https` |
| site_name | string | 否 | 网站名称 |
| icp_number | string | 否 | ICP 备案号 |
| police_number | string | 否 | 公安备案号 |
| is_active | bool | 否 | 是否启用 |
| pass_query_params | bool | 否 | 是否透传查询参数 |
| random_suffix_length | int | 否 | 随机后缀长度（0-10） |
| enable_checksum | bool | 否 | 是否启用校验位 |
| description | string | 否 | 描述 |

### 获取域名列表

**请求**

```
GET /api/v1/domains
```

### 获取活跃域名列表

**请求**

```
GET /api/v1/domains/active
```

### 更新域名

**请求**

```
PUT /api/v1/domains/:id
```

### 更新域名状态

**请求**

```
PUT /api/v1/domains/:id/status
```

**请求体**

```json
{
    "is_active": true
}
```

### 删除域名

**请求**

```
DELETE /api/v1/domains/:id
```

---

## 用户管理接口

### 创建用户

**请求**

```
POST /api/v1/users
```

**请求体**

```json
{
    "username": "newuser",
    "password": "password123",
    "real_name": "张三",
    "email": "zhangsan@example.com",
    "phone": "13800138000"
}
```

### 获取用户列表

**请求**

```
GET /api/v1/users
```

**查询参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |
| username | string | 否 | 用户名（模糊搜索） |
| real_name | string | 否 | 真实姓名（模糊搜索） |
| status | int | 否 | 状态 |

### 获取用户详情

**请求**

```
GET /api/v1/users/:id
```

### 更新用户

**请求**

```
PUT /api/v1/users/:id
```

### 删除用户

**请求**

```
DELETE /api/v1/users/:id
```

### 重置用户密码

**请求**

```
POST /api/v1/users/:id/reset-password
```

**请求体**

```json
{
    "new_password": "newpassword123"
}
```

---

## 当前用户接口

### 获取当前用户信息

**请求**

```
GET /api/v1/profile
```

### 修改密码

**请求**

```
POST /api/v1/profile/change-password
```

**请求体**

```json
{
    "old_password": "oldpassword",
    "new_password": "newpassword123"
}
```

---

## 统计接口

### 获取系统统计

**请求**

```
GET /api/v1/statistics/system
```

### 获取仪表盘统计

**请求**

```
GET /api/v1/statistics/dashboard
```

### 获取短链接统计

**请求**

```
GET /api/v1/statistics/short-links
```

---

## 点击统计接口

### 获取点击统计列表

**请求**

```
GET /api/v1/click_statistics
```

### 获取点击统计分析

**请求**

```
GET /api/v1/click_statistics/analysis
```

响应增加设备、浏览器、操作系统、机器人 Bot 和 UTM 维度字段：`top_devices`、`top_browsers`、`top_os`、`bot_stats`、`top_utm_sources`、`top_utm_campaigns`。支持 `short_link_id`、`campaign_id`、`tag_id`、`device_type`、`is_bot`、`start_date`、`end_date` 过滤。

### 获取地图地理聚合

**请求**

```
GET /api/v1/click_statistics/geo-analysis
```

地图专用地理聚合，不做 Top N 截断。支持 `level=country|province|city`，默认 `country`；支持 `short_link_id`、`campaign_id`、`route_id`、`tag_id`、`country`、`province`、`device_type`、`is_bot`、`start_date`、`end_date` 过滤。响应包含 `total_clicks`、`unique_ips`、`level`、`country`、`province` 与 `regions`。

### 导出点击明细

**请求**

```
GET /api/v1/click_statistics/export
```

返回同步 CSV 文件，支持 `short_link_id`、`campaign_id`、`tag_id`、`start_date`、`end_date`、`device_type`、`is_bot`。单次最多 50,000 行，超过时返回 400，需缩小筛选范围。

---

## A/B 测试接口

### A/B 测试反馈流程

1. 管理端创建并启动 A/B 测试。
2. 用户访问短链后，系统按实验配置选择变体并跳转到变体目标 URL。
3. 跳转目标 URL 会追加 `_dwz_abt` 查询参数，例如：`https://example.com/page-a?_dwz_abt=<token>`。
4. 落地页或业务系统在注册、下单、购买等结果发生后，调用公开反馈接口回传转化。
5. 统计接口根据点击与反馈事件计算真实转化数、转化率和转化价值。

`_dwz_abt` 是服务端签名 token，绑定 `workspace_id`、`ab_test_id`、`variant_id`、`short_link_id` 和 `session_id`，默认有效期 30 天。直接访问变体目标 URL 不会生成 token，必须通过短链跳转进入实验。

管理端 A/B 测试统计弹窗中的“分流反馈”表单仅用于手动验证。生产环境应在落地页或业务系统中自动调用反馈接口。

### 获取 A/B 测试统计

**请求**

```
GET /api/v1/ab_tests/{id}/statistics
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | A/B 测试 ID |

**查询参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| days | int | 否 | 统计最近天数，默认 7，最大 365 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "ab_test_id": 1,
    "total_clicks": 1000,
    "total_conversions": 138,
    "conversion_value": 2999.5,
    "variant_stats": [
      {
        "variant": {
          "id": 1,
          "ab_test_id": 1,
          "name": "版本A",
          "target_url": "https://example.com/page-a",
          "weight": 50,
          "is_control": true
        },
        "click_count": 480,
        "unique_clicks": 460,
        "conversion_count": 58,
        "conversion_rate": 12.61,
        "conversion_value": 1200,
        "percentage": 48
      }
    ],
    "daily_stats": [
      {
        "date": "2024-01-15",
        "variants": {
          "1": 45,
          "2": 55
        }
      }
    ],
    "conversion_rate": 13.8,
    "winning_variant": {
      "id": 1,
      "name": "版本A"
    }
  }
}
```

**统计口径**

- `click_count`：变体点击次数。
- `unique_clicks`：变体唯一会话点击数。
- `conversion_count`：通过反馈接口写入的业务结果数。
- 变体 `conversion_rate`：`conversion_count / unique_clicks * 100`，无唯一点击时为 `0`。
- 顶层 `conversion_rate`：`total_conversions / 所有变体 unique_clicks 之和 * 100`。
- `conversion_value`：反馈事件 `value` 汇总。
- `winning_variant`：当前按转化数最高的变体计算。

### 上报转化反馈

**请求**

```
POST /api/v1/public/ab_test_feedback
```

公开接口，不需要登录态或 Bearer Token。写入范围由 `_dwz_abt` 签名 token 限制。

**请求体**

```json
{
  "feedback_token": "<_dwz_abt 参数值>",
  "event_id": "order-202401150001",
  "value": 99.9,
  "currency": "CNY",
  "metadata": {
    "plan": "pro"
  },
  "occurred_at": "2024-01-15T10:30:00Z"
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| feedback_token | string | 是 | 短链 A/B 跳转目标 URL 中 `_dwz_abt` 的值 |
| event_id | string | 是 | 业务事件唯一 ID，同一 A/B 测试内幂等，最长 128 字符 |
| value | number | 否 | 转化价值，必须大于等于 0 |
| currency | string | 否 | 币种，服务端会转为大写，最长 16 字符 |
| metadata | object | 否 | 业务附加信息，序列化后最大 4096 字节 |
| occurred_at | string | 否 | 业务事件发生时间，ISO 8601 格式；不传则使用服务端当前时间 |

**成功响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 10,
    "duplicate": false,
    "workspace_id": 1,
    "ab_test_id": 1,
    "variant_id": 2,
    "short_link_id": 5,
    "session_id": "ab_1_5_xxx",
    "event_id": "order-202401150001"
  }
}
```

重复提交相同 `event_id` 会返回成功，但 `duplicate` 为 `true`，不会重复计入转化：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 10,
    "duplicate": true,
    "workspace_id": 1,
    "ab_test_id": 1,
    "variant_id": 2,
    "short_link_id": 5,
    "session_id": "ab_1_5_xxx",
    "event_id": "order-202401150001"
  }
}
```

**错误说明**

| HTTP 状态 | code | 场景 |
|-----------|------|------|
| 400 | 40001 | 缺少 `feedback_token`、缺少 `event_id`、字段长度非法、`value` 为负数 |
| 401 | 40101 | token 无效、被篡改、过期，或 token 绑定的实验/变体不存在 |
| 500 | 50001 | 服务端写入失败 |

**落地页自动回传示例**

```js
const token = new URLSearchParams(location.search).get('_dwz_abt');

if (token) {
  await fetch('https://your-domain.com/api/v1/public/ab_test_feedback', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      feedback_token: token,
      event_id: 'order-202401150001',
      value: 99.9,
      currency: 'CNY',
      metadata: {
        order_id: '202401150001',
        plan: 'pro'
      }
    })
  });
}
```

---

## 操作日志接口

### 获取操作日志

**请求**

```
GET /api/v1/logs
```

---

## 健康检查接口

### 完整健康检查

**请求**

```
GET /health
```

**响应**

```json
{
    "status": "healthy",
    "database": "connected",
    "cache": "connected",
    "timestamp": "2024-01-01T00:00:00Z"
}
```

### 简单健康检查

**请求**

```
GET /health/simple
```

**响应**

```json
{
    "status": "ok"
}
```
