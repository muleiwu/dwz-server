# API Bearer Token 认证文档（传统方式）

> ⚠️ **安全提示**：Bearer Token 认证方式安全性较低，Token 在网络传输中可能被截获。建议仅在内网环境或测试场景使用。生产环境推荐使用 [签名认证](./API_SIGNATURE_AUTH.md)。

## 概述

Bearer Token 是一种简单的认证方式，客户端在请求头中携带 Token 即可完成认证。

## 认证流程

```
┌─────────────┐                                    ┌─────────────┐
│   客户端    │                                    │   服务端    │
└──────┬──────┘                                    └──────┬──────┘
       │                                                  │
       │  ─────────────────────────────────────────────>  │
       │  请求 + Authorization: Bearer <token>            │
       │                                                  │
       │                                                  │  验证 Token
       │                                                  │
       │  <─────────────────────────────────────────────  │
       │  响应结果                                         │
       │                                                  │
```

## 请求头

| Header | 必填 | 描述 | 示例 |
|--------|------|------|------|
| `Authorization` | 是 | Bearer Token | `Bearer a1b2c3d4e5f6...` |

## 使用方式

在每个需要认证的 API 请求中，添加 `Authorization` 请求头：

```
Authorization: Bearer <your_token_here>
```

## 代码示例

### cURL

```bash
# 创建短链接
curl -X POST "https://api.example.com/api/v1/short_links" \
  -H "Authorization: Bearer a1b2c3d4e5f6789012345678901234567890123456789012345678901234" \
  -H "Content-Type: application/json" \
  -d '{"original_url": "https://www.example.com", "title": "示例网站"}'

# 获取短链接列表
curl -X GET "https://api.example.com/api/v1/short_links?page=1&page_size=10" \
  -H "Authorization: Bearer a1b2c3d4e5f6789012345678901234567890123456789012345678901234"

# 删除短链接
curl -X DELETE "https://api.example.com/api/v1/short_links/123" \
  -H "Authorization: Bearer a1b2c3d4e5f6789012345678901234567890123456789012345678901234"
```

### Python

```python
import requests

class BearerAuth:
    def __init__(self, token: str, base_url: str):
        self.token = token
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }
    
    def get(self, path: str, params: dict = None) -> dict:
        url = f"{self.base_url}{path}"
        response = requests.get(url, params=params, headers=self.headers)
        return response.json()
    
    def post(self, path: str, data: dict = None) -> dict:
        url = f"{self.base_url}{path}"
        response = requests.post(url, json=data, headers=self.headers)
        return response.json()
    
    def put(self, path: str, data: dict = None) -> dict:
        url = f"{self.base_url}{path}"
        response = requests.put(url, json=data, headers=self.headers)
        return response.json()
    
    def delete(self, path: str) -> dict:
        url = f"{self.base_url}{path}"
        response = requests.delete(url, headers=self.headers)
        return response.json()

# 使用示例
auth = BearerAuth(
    token="a1b2c3d4e5f6789012345678901234567890123456789012345678901234",
    base_url="https://api.example.com"
)

# 创建短链接
result = auth.post('/api/v1/short_links', {
    'original_url': 'https://www.example.com',
    'title': '示例网站'
})
print(result)

# 获取短链接列表
result = auth.get('/api/v1/short_links', {'page': 1, 'page_size': 10})
print(result)

# 删除短链接
result = auth.delete('/api/v1/short_links/123')
print(result)
```

### JavaScript / Node.js

```javascript
const axios = require('axios');

class BearerAuth {
    constructor(token, baseUrl) {
        this.token = token;
        this.baseUrl = baseUrl;
        this.headers = {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        };
    }

    async get(path, params = null) {
        const url = `${this.baseUrl}${path}`;
        const response = await axios.get(url, { headers: this.headers, params });
        return response.data;
    }

    async post(path, data = null) {
        const url = `${this.baseUrl}${path}`;
        const response = await axios.post(url, data, { headers: this.headers });
        return response.data;
    }

    async put(path, data = null) {
        const url = `${this.baseUrl}${path}`;
        const response = await axios.put(url, data, { headers: this.headers });
        return response.data;
    }

    async delete(path) {
        const url = `${this.baseUrl}${path}`;
        const response = await axios.delete(url, { headers: this.headers });
        return response.data;
    }
}

// 使用示例
const auth = new BearerAuth(
    'a1b2c3d4e5f6789012345678901234567890123456789012345678901234',
    'https://api.example.com'
);

// 创建短链接
auth.post('/api/v1/short_links', {
    original_url: 'https://www.example.com',
    title: '示例网站'
}).then(result => console.log(result));

// 获取短链接列表
auth.get('/api/v1/short_links', { page: 1, page_size: 10 })
    .then(result => console.log(result));

// 删除短链接
auth.delete('/api/v1/short_links/123')
    .then(result => console.log(result));
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type BearerAuth struct {
    Token   string
    BaseURL string
}

func (b *BearerAuth) request(method, path string, data map[string]interface{}) (map[string]interface{}, error) {
    var body io.Reader
    if data != nil {
        jsonData, _ := json.Marshal(data)
        body = bytes.NewBuffer(jsonData)
    }

    req, err := http.NewRequest(method, b.BaseURL+path, body)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", "Bearer "+b.Token)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    return result, nil
}

func (b *BearerAuth) Get(path string) (map[string]interface{}, error) {
    return b.request("GET", path, nil)
}

func (b *BearerAuth) Post(path string, data map[string]interface{}) (map[string]interface{}, error) {
    return b.request("POST", path, data)
}

func (b *BearerAuth) Put(path string, data map[string]interface{}) (map[string]interface{}, error) {
    return b.request("PUT", path, data)
}

func (b *BearerAuth) Delete(path string) (map[string]interface{}, error) {
    return b.request("DELETE", path, nil)
}

func main() {
    auth := &BearerAuth{
        Token:   "a1b2c3d4e5f6789012345678901234567890123456789012345678901234",
        BaseURL: "https://api.example.com",
    }

    // 创建短链接
    result, _ := auth.Post("/api/v1/short_links", map[string]interface{}{
        "original_url": "https://www.example.com",
        "title":        "示例网站",
    })
    fmt.Println(result)

    // 获取短链接列表
    result, _ = auth.Get("/api/v1/short_links?page=1&page_size=10")
    fmt.Println(result)

    // 删除短链接
    result, _ = auth.Delete("/api/v1/short_links/123")
    fmt.Println(result)
}
```

## 错误响应

| HTTP 状态码 | 错误消息 | 说明 |
|------------|---------|------|
| 401 | 缺少认证信息 | 缺少 Authorization 头 |
| 401 | Token格式错误 | Authorization 头格式不正确 |
| 401 | Token已失效 | Token 已过期或被禁用 |
| 401 | 用户已被禁用 | 关联的用户已被禁用 |

## 安全建议

1. **使用 HTTPS**：务必使用 HTTPS 传输，防止 Token 被中间人截获
2. **定期轮换**：定期更换 Token，降低泄露风险
3. **最小权限**：为不同用途创建不同的 Token
4. **监控使用**：定期检查 Token 的使用情况
5. **考虑升级**：如果安全要求较高，建议升级到 [签名认证](./API_SIGNATURE_AUTH.md)

## 与签名认证的对比

| 特性 | Bearer Token | 签名认证 |
|------|-------------|---------|
| 实现复杂度 | 简单 | 中等 |
| 安全性 | 较低 | 高 |
| 密钥传输 | Token 在网络传输 | 密钥不传输 |
| 防重放攻击 | 无 | 有（时间戳+nonce） |
| 请求完整性 | 无 | 有（签名验证） |
| 适用场景 | 内网/测试 | 生产环境 |
