# API 签名认证文档（推荐）

## 概述

签名认证是一种基于 HMAC-SHA256 的安全认证方式。与传统的 Bearer Token 认证相比，签名认证具有以下优势：

- **密钥不传输**：App Secret 永远不会在网络上传输，即使请求被截获，攻击者也无法伪造有效签名
- **防重放攻击**：通过时间戳和随机数（nonce）机制，防止请求被重放
- **请求完整性**：签名包含请求的所有关键信息，任何篡改都会导致签名验证失败

## 认证流程

```
┌─────────────┐                                    ┌─────────────┐
│   客户端    │                                    │   服务端    │
└──────┬──────┘                                    └──────┬──────┘
       │                                                  │
       │  1. 准备请求参数                                  │
       │  2. 生成时间戳和随机数                            │
       │  3. 使用 App Secret 计算签名                      │
       │                                                  │
       │  ─────────────────────────────────────────────>  │
       │  请求 + X-App-Id + X-Signature +                 │
       │  X-Timestamp + X-Nonce                           │
       │                                                  │
       │                                                  │  4. 验证时间戳
       │                                                  │  5. 查询 App Secret
       │                                                  │  6. 重新计算签名
       │                                                  │  7. 比对签名
       │                                                  │
       │  <─────────────────────────────────────────────  │
       │  响应结果                                         │
       │                                                  │
```

## 请求头

| Header | 必填 | 描述 | 示例 |
|--------|------|------|------|
| `X-App-Id` | 是 | 应用标识符，创建 Token 时获取 | `app_1a2b3c4d5e6f7890` |
| `X-Signature` | 是 | HMAC-SHA256 签名（十六进制） | `a1b2c3d4e5f6...` |
| `X-Timestamp` | 是 | Unix 时间戳（秒） | `1703232000` |
| `X-Nonce` | 是 | 随机字符串（建议 16-32 字符） | `abc123xyz789def456` |

## 签名算法

### 签名公式

```
signature = HMAC-SHA256(app_secret, string_to_sign)
```

### 待签名字符串构建

```
string_to_sign = method + path + sorted_params_json + timestamp + nonce
```

其中：
- `method`: HTTP 方法（大写），如 `GET`、`POST`、`PUT`、`DELETE`
- `path`: 请求路径，如 `/api/v1/short_links`
- `sorted_params_json`: 请求参数按 key 字母顺序排序后的 JSON 字符串
- `timestamp`: Unix 时间戳（秒）
- `nonce`: 随机字符串

### 参数排序规则

1. 对于 GET 请求，使用 URL 查询参数
2. 对于 POST/PUT/PATCH 请求，使用 JSON Body 参数
3. 参数按 key 的字母顺序排序
4. 空参数使用 `{}`

### 签名示例

**请求信息：**
- Method: `POST`
- Path: `/api/v1/short_links`
- Body: `{"original_url": "https://example.com", "title": "示例"}`
- Timestamp: `1703232000`
- Nonce: `abc123xyz789`
- App Secret: `your_app_secret_here`

**步骤 1：构建待签名字符串**

```
string_to_sign = "POST" + "/api/v1/short_links" + '{"original_url":"https://example.com","title":"示例"}' + "1703232000" + "abc123xyz789"
```

注意：JSON 中的 key 需要按字母顺序排序（`original_url` 在 `title` 之前）。

**步骤 2：计算 HMAC-SHA256 签名**

```python
import hmac
import hashlib

app_secret = "your_app_secret_here"
string_to_sign = 'POST/api/v1/short_links{"original_url":"https://example.com","title":"示例"}1703232000abc123xyz789'

signature = hmac.new(
    app_secret.encode('utf-8'),
    string_to_sign.encode('utf-8'),
    hashlib.sha256
).hexdigest()
```

## 代码示例

### Python

```python
import hmac
import hashlib
import json
import time
import uuid
import requests

class SignatureAuth:
    def __init__(self, app_id: str, app_secret: str, base_url: str):
        self.app_id = app_id
        self.app_secret = app_secret
        self.base_url = base_url
    
    def _sort_params(self, params: dict) -> str:
        """将参数按 key 排序后转为 JSON 字符串"""
        if not params:
            return "{}"
        sorted_params = dict(sorted(params.items()))
        return json.dumps(sorted_params, separators=(',', ':'), ensure_ascii=False)
    
    def _generate_signature(self, method: str, path: str, params: dict, timestamp: int, nonce: str) -> str:
        """生成 HMAC-SHA256 签名"""
        sorted_params_json = self._sort_params(params)
        string_to_sign = f"{method.upper()}{path}{sorted_params_json}{timestamp}{nonce}"
        
        signature = hmac.new(
            self.app_secret.encode('utf-8'),
            string_to_sign.encode('utf-8'),
            hashlib.sha256
        ).hexdigest()
        
        return signature
    
    def request(self, method: str, path: str, params: dict = None, data: dict = None) -> dict:
        """发送带签名的请求"""
        timestamp = int(time.time())
        nonce = str(uuid.uuid4()).replace('-', '')[:16]
        
        # 确定用于签名的参数
        sign_params = data if method.upper() in ['POST', 'PUT', 'PATCH'] else (params or {})
        
        signature = self._generate_signature(method, path, sign_params or {}, timestamp, nonce)
        
        headers = {
            'X-App-Id': self.app_id,
            'X-Signature': signature,
            'X-Timestamp': str(timestamp),
            'X-Nonce': nonce,
            'Content-Type': 'application/json'
        }
        
        url = f"{self.base_url}{path}"
        
        if method.upper() == 'GET':
            response = requests.get(url, params=params, headers=headers)
        elif method.upper() == 'POST':
            response = requests.post(url, json=data, headers=headers)
        elif method.upper() == 'PUT':
            response = requests.put(url, json=data, headers=headers)
        elif method.upper() == 'DELETE':
            response = requests.delete(url, headers=headers)
        else:
            raise ValueError(f"Unsupported method: {method}")
        
        return response.json()

# 使用示例
auth = SignatureAuth(
    app_id="app_1a2b3c4d5e6f7890",
    app_secret="your_app_secret_here",
    base_url="https://api.example.com"
)

# 创建短链接
result = auth.request('POST', '/api/v1/short_links', data={
    'original_url': 'https://www.example.com',
    'title': '示例网站'
})
print(result)

# 获取短链接列表
result = auth.request('GET', '/api/v1/short_links', params={'page': 1, 'page_size': 10})
print(result)
```

### JavaScript / Node.js

```javascript
const crypto = require('crypto');
const axios = require('axios');
const { v4: uuidv4 } = require('uuid');

class SignatureAuth {
    constructor(appId, appSecret, baseUrl) {
        this.appId = appId;
        this.appSecret = appSecret;
        this.baseUrl = baseUrl;
    }

    _sortParams(params) {
        if (!params || Object.keys(params).length === 0) {
            return '{}';
        }
        const sortedKeys = Object.keys(params).sort();
        const sortedObj = {};
        sortedKeys.forEach(key => {
            sortedObj[key] = params[key];
        });
        return JSON.stringify(sortedObj);
    }

    _generateSignature(method, path, params, timestamp, nonce) {
        const sortedParamsJson = this._sortParams(params);
        const stringToSign = `${method.toUpperCase()}${path}${sortedParamsJson}${timestamp}${nonce}`;
        
        const signature = crypto
            .createHmac('sha256', this.appSecret)
            .update(stringToSign, 'utf8')
            .digest('hex');
        
        return signature;
    }

    async request(method, path, { params = null, data = null } = {}) {
        const timestamp = Math.floor(Date.now() / 1000);
        const nonce = uuidv4().replace(/-/g, '').substring(0, 16);
        
        // 确定用于签名的参数
        const signParams = ['POST', 'PUT', 'PATCH'].includes(method.toUpperCase()) 
            ? (data || {}) 
            : (params || {});
        
        const signature = this._generateSignature(method, path, signParams, timestamp, nonce);
        
        const headers = {
            'X-App-Id': this.appId,
            'X-Signature': signature,
            'X-Timestamp': timestamp.toString(),
            'X-Nonce': nonce,
            'Content-Type': 'application/json'
        };
        
        const url = `${this.baseUrl}${path}`;
        
        const config = { headers };
        
        let response;
        switch (method.toUpperCase()) {
            case 'GET':
                response = await axios.get(url, { ...config, params });
                break;
            case 'POST':
                response = await axios.post(url, data, config);
                break;
            case 'PUT':
                response = await axios.put(url, data, config);
                break;
            case 'DELETE':
                response = await axios.delete(url, config);
                break;
            default:
                throw new Error(`Unsupported method: ${method}`);
        }
        
        return response.data;
    }
}

// 使用示例
const auth = new SignatureAuth(
    'app_1a2b3c4d5e6f7890',
    'your_app_secret_here',
    'https://api.example.com'
);

// 创建短链接
auth.request('POST', '/api/v1/short_links', {
    data: {
        original_url: 'https://www.example.com',
        title: '示例网站'
    }
}).then(result => console.log(result));

// 获取短链接列表
auth.request('GET', '/api/v1/short_links', {
    params: { page: 1, page_size: 10 }
}).then(result => console.log(result));
```

### Go

```go
package main

import (
    "bytes"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "sort"
    "strconv"
    "strings"
    "time"

    "github.com/google/uuid"
)

type SignatureAuth struct {
    AppID     string
    AppSecret string
    BaseURL   string
}

func (s *SignatureAuth) sortParams(params map[string]interface{}) string {
    if len(params) == 0 {
        return "{}"
    }

    keys := make([]string, 0, len(params))
    for k := range params {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    sortedMap := make(map[string]interface{})
    for _, k := range keys {
        sortedMap[k] = params[k]
    }

    // 使用 Buffer 和 Encoder 禁用 HTML 转义，避免 &<> 等字符被转义
    var buf bytes.Buffer
    encoder := json.NewEncoder(&buf)
    encoder.SetEscapeHTML(false)
    encoder.Encode(sortedMap)
    
    // Encoder.Encode 会添加换行符，需要 TrimSpace
    return strings.TrimSpace(buf.String())
}

func (s *SignatureAuth) generateSignature(method, path string, params map[string]interface{}, timestamp int64, nonce string) string {
    sortedParamsJSON := s.sortParams(params)
    stringToSign := fmt.Sprintf("%s%s%s%d%s", method, path, sortedParamsJSON, timestamp, nonce)

    h := hmac.New(sha256.New, []byte(s.AppSecret))
    h.Write([]byte(stringToSign))
    return hex.EncodeToString(h.Sum(nil))
}

func (s *SignatureAuth) Request(method, path string, data map[string]interface{}) (map[string]interface{}, error) {
    timestamp := time.Now().Unix()
    nonce := uuid.New().String()[:16]

    signature := s.generateSignature(method, path, data, timestamp, nonce)

    var body io.Reader
    if data != nil {
        jsonData, _ := json.Marshal(data)
        body = bytes.NewBuffer(jsonData)
    }

    req, err := http.NewRequest(method, s.BaseURL+path, body)
    if err != nil {
        return nil, err
    }

    req.Header.Set("X-App-Id", s.AppID)
    req.Header.Set("X-Signature", signature)
    req.Header.Set("X-Timestamp", strconv.FormatInt(timestamp, 10))
    req.Header.Set("X-Nonce", nonce)
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

func main() {
    auth := &SignatureAuth{
        AppID:     "app_1a2b3c4d5e6f7890",
        AppSecret: "your_app_secret_here",
        BaseURL:   "https://api.example.com",
    }

    // 创建短链接
    result, _ := auth.Request("POST", "/api/v1/short_links", map[string]interface{}{
        "original_url": "https://www.example.com",
        "title":        "示例网站",
    })
    fmt.Println(result)
}
```

## 错误响应

| HTTP 状态码 | 错误消息 | 说明 |
|------------|---------|------|
| 401 | 缺少认证信息 | 缺少必要的认证头 |
| 401 | 时间戳无效 | 时间戳超出 ±5 分钟窗口 |
| 401 | 签名验证失败 | 签名不匹配 |
| 401 | 无效的AppID | App ID 不存在 |
| 401 | Token已禁用 | Token 已被禁用 |
| 401 | 用户已被禁用 | 关联的用户已被禁用 |

## 注意事项

1. **时间同步**：确保客户端时间与服务器时间同步，时间戳误差不能超过 ±5 分钟
2. **Nonce 唯一性**：建议使用 UUID 或其他随机字符串生成器
3. **密钥安全**：App Secret 应妥善保管，不要在客户端代码中硬编码
4. **HTTPS**：生产环境务必使用 HTTPS
5. **参数编码**：JSON 参数中的中文等特殊字符需要正确编码
