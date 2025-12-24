# API 接口文档

## 概述

本文档描述了短链接服务的所有 API 接口。

**Base URL**: `https://your-domain.com`

**API 版本**: v1

## 认证方式

系统支持两种认证方式：

1. **签名认证（推荐）**：基于 HMAC-SHA256 的安全认证，详见 [签名认证文档](./API_SIGNATURE_AUTH.md)
2. **Bearer Token 认证**：传统 Token 认证，详见 [Bearer Token 文档](./API_BEARER_AUTH.md)

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
    "domain": "https://short.ly",
    "custom_code": "mylink",
    "title": "示例网站",
    "description": "这是一个示例网站",
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
| expire_at | string | 否 | 过期时间 |

**响应**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 1,
        "short_code": "abc123",
        "domain": "https://short.ly",
        "short_url": "https://short.ly/abc123",
        "original_url": "https://www.example.com/very/long/url",
        "title": "示例网站",
        "description": "这是一个示例网站",
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
    "domain": "https://short.ly"
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
                "short_url": "https://short.ly/abc123",
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
        "domain": "https://short.ly",
        "short_url": "https://short.ly/abc123",
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

## 域名管理接口

### 创建域名

**请求**

```
POST /api/v1/domains
```

**请求体**

```json
{
    "domain": "short.ly",
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
