# 木雷短网址 - 企业级短链接服务平台

[![Go Version](https://img.shields.io/github/go-mod/go-version/your-org/dwz-server)](https://golang.org/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

> 🚀 一个功能完善、高性能的企业级短链接服务平台，支持多域名、AB测试、用户管理、实时统计等功能。

## ✨ 功能特性

### 🔗 核心功能
- **短链接生成**: 支持自定义短码，自动生成唯一标识
- **多域名支持**: 支持配置多个短链接域名，灵活管理
- **链接管理**: 完整的CRUD操作，支持批量管理
- **过期管理**: 支持设置链接过期时间，自动失效
- **链接状态**: 支持启用/禁用链接状态控制

### 🧪 AB测试系统
- **多版本测试**: 为同一短链接创建多个目标URL版本
- **智能分流**: 支持平均分配、权重分配等流量分配策略
- **会话一致性**: 同一用户在测试期间始终访问相同版本
- **实时统计**: 实时收集各版本的点击数据和转化率
- **测试管理**: 完整的测试生命周期管理

### 👥 用户管理
- **用户认证**: 支持用户注册、登录、密码管理
- **Token管理**: 支持API Token和登录Token双重认证
- **权限控制**: 基于用户的访问权限管理
- **操作日志**: 详细记录用户操作，支持审计追踪

### 📊 统计分析
- **点击统计**: 实时记录点击数据，包括IP、地理位置、设备信息
- **数据分析**: 提供多维度统计分析，包括地理分布、时间分布等
- **AB测试分析**: 专门的AB测试数据分析和转化率统计
- **导出功能**: 支持数据导出，便于进一步分析

### 🛡️ 安全与监控
- **操作日志**: 自动记录所有操作，支持敏感信息脱敏
- **健康检查**: 提供服务健康状态监控
- **性能监控**: 高并发场景下的性能优化
- **安全防护**: 防止恶意访问和数据泄露

## 🏗️ 技术架构

### 技术栈
- **语言**: Go 1.23+
- **Web框架**: Gin
- **数据库**: MySQL/PostgreSQL 
- **缓存**: Redis
- **ORM**: GORM
- **配置管理**: Viper
- **日志**: Zap
- **HTTP客户端**: go-resty

### 架构设计
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │   Mobile App    │    │   API Client    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌─────────────────────────────────────────────────┐
         │                Load Balancer                    │
         └─────────────────────────────────────────────────┘
                                 │
         ┌─────────────────────────────────────────────────┐
         │                 DWZ Server                      │
         │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
         │  │ Controller  │  │ Middleware  │  │   Router    │
         │  └─────────────┘  └─────────────┘  └─────────────┘
         │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
         │  │  Service    │  │     DAO     │  │   Model     │
         │  └─────────────┘  └─────────────┘  └─────────────┘
         └─────────────────────────────────────────────────┘
                                 │
         ┌─────────────────────────────────────────────────┐
         │                 Data Layer                      │
         │  ┌─────────────┐              ┌─────────────┐    │
         │  │   MySQL     │              │    Redis    │    │
         │  │ PostgreSQL  │              │   Cache     │    │
         │  └─────────────┘              └─────────────┘    │
         └─────────────────────────────────────────────────┘
```

### 分层架构
- **Controller层**: 处理HTTP请求，参数验证，调用Service
- **Service层**: 业务逻辑处理，事务管理  
- **DAO层**: 数据访问，数据库操作
- **Model层**: 数据模型定义
- **Middleware层**: 认证、日志、CORS等中间件

## 🚀 快速开始

### 环境要求
- Go 1.23+
- MySQL 5.7+ 或 PostgreSQL 9.6+
- Redis 6.0+

### 安装步骤

1. **克隆项目**
```bash
git clone https://github.com/your-org/dwz-server.git
cd dwz-server
```

2. **安装依赖**
```bash
go mod download
```

3. **配置数据库**
```bash
# 复制配置文件
cp config.yaml.example config.yaml

# 编辑配置文件，设置数据库连接信息
vim config.yaml
```

4. **初始化数据库**
```bash
# 创建数据库表结构
# 执行项目中的数据库迁移脚本
```

5. **启动服务**
```bash
go run main.go
```

6. **验证服务**
```bash
# 健康检查
curl http://localhost:8080/health

# API测试
curl -X POST http://localhost:8080/api/v1/short_links \
  -H "Content-Type: application/json" \
  -d '{"original_url": "https://example.com"}'
```

### Docker 部署

```bash
# 构建镜像
docker build -t dwz-server .

# 运行容器
docker run -d \
  --name dwz-server \
  -p 8080:8080 \
  -v /path/to/config.yaml:/app/config.yaml \
  dwz-server
```

## 📖 API 文档

### 基础信息
- **基础URL**: `http://localhost:8080`
- **内容类型**: `application/json`
- **认证方式**: Bearer Token

### 主要接口

#### 短链接管理
```bash
# 创建短链接
POST /api/v1/short_links
{
  "original_url": "https://example.com",
  "domain": "short.ly",
  "custom_code": "abc123"
}

# 获取短链接列表
GET /api/v1/short_links?page=1&page_size=10

# 获取短链接详情
GET /api/v1/short_links/{id}

# 更新短链接
PUT /api/v1/short_links/{id}

# 删除短链接
DELETE /api/v1/short_links/{id}
```

#### 用户管理
```bash
# 用户登录
POST /api/v1/login
{
  "username": "admin",
  "password": "admin123"
}

# 创建用户
POST /api/v1/users
{
  "username": "newuser",
  "password": "password123",
  "email": "user@example.com"
}
```

#### AB测试
```bash
# 创建AB测试
POST /api/v1/ab_tests
{
  "short_link_id": 1,
  "name": "按钮颜色测试",
  "variants": [
    {
      "name": "红色按钮",
      "target_url": "https://example.com/red"
    },
    {
      "name": "蓝色按钮", 
      "target_url": "https://example.com/blue"
    }
  ]
}

# 获取AB测试统计
GET /api/v1/ab_tests/{id}/statistics
```

详细的API文档请参考 [API.md](docs/API.md)

## 🔧 配置说明

### 配置文件结构
```yaml
app:
  name: "DWZ Server"
  version: "1.0.0"
  port: 8080
  mode: "debug"

database:
  driver: "mysql"
  host: "localhost"
  port: 3306
  database: "dwz_server"
  username: "root"
  password: "password"

redis:
  host: "localhost"
  port: 6379
  password: ""
  database: 0

middleware:
  operation_log:
    enable: true
    max_request_size: 1048576
    sensitive_fields: ["password", "token"]
    async_logging: true
```

### 环境变量
- `APP_PORT`: 服务端口 (默认: 8080)
- `DB_HOST`: 数据库主机
- `DB_PORT`: 数据库端口
- `DB_USER`: 数据库用户名
- `DB_PASSWORD`: 数据库密码
- `REDIS_HOST`: Redis主机
- `REDIS_PORT`: Redis端口

## 🔍 性能特点

### 高性能设计
- **并发优化**: 支持高并发访问，经过性能测试验证
- **缓存策略**: 多级缓存机制，提升响应速度
- **异步处理**: 统计记录异步处理，不影响主流程性能
- **连接池**: 数据库连接池优化，减少连接开销

### 性能基准
- **响应时间**: 平均响应时间 < 10ms
- **并发处理**: 支持万级并发请求
- **吞吐量**: 单实例支持 10,000+ QPS
- **可扩展性**: 支持水平扩展，多实例部署

## 🛡️ 安全特性

### 数据安全
- **敏感信息脱敏**: 自动脱敏密码、Token等敏感信息
- **访问控制**: 基于Token的访问控制机制
- **操作审计**: 完整的操作日志记录
- **数据加密**: 敏感数据加密存储

### 系统安全
- **防刷机制**: 防止恶意刷取短链接
- **访问限制**: 支持IP访问频率限制
- **输入验证**: 严格的输入参数验证
- **错误处理**: 安全的错误信息返回

## 📊 监控与运维

### 健康检查
```bash
# 详细健康检查
GET /health

# 简单健康检查
GET /health/simple
```

### 日志管理
- **结构化日志**: JSON格式日志输出
- **日志级别**: 支持不同级别日志配置
- **日志轮转**: 自动日志文件轮转
- **监控集成**: 支持主流监控系统集成

### 性能监控
- **实时监控**: 实时性能指标监控
- **告警机制**: 异常情况自动告警
- **性能分析**: 详细的性能分析报告
- **容量规划**: 基于历史数据的容量规划

## 🤝 参与贡献

我们欢迎所有形式的贡献，包括但不限于：

- 🐛 Bug 报告
- 🆕 功能建议
- 📝 文档改进
- 🔧 代码优化
- 🧪 测试用例

### 开发指南

1. **Fork 项目**
2. **创建功能分支** (`git checkout -b feature/amazing-feature`)
3. **提交更改** (`git commit -m 'Add amazing feature'`)
4. **推送到分支** (`git push origin feature/amazing-feature`)
5. **创建 Pull Request**

### 代码规范
- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 编写单元测试
- 添加必要的注释

详细的开发指南请参考 [CONTRIBUTING.md](CONTRIBUTING.md)

## 📄 许可证

本项目可以二次开发，但是禁止二次分发。

## 🙏 致谢

感谢所有贡献者的努力和开源社区的支持！

## 📞 联系方式

- **Issue**: [GitHub Issues](https://cnb.cool/mliev/open/dwz-server/-/issues)
\- **文档**: [项目文档](docs/)

---

⭐ 如果这个项目对您有帮助，请给我们一个星标！
