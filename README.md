# 木雷短网址 - 企业级短链接服务平台

[![Go Version](https://img.shields.io/badge/Go-1.25.0-blue.svg)](https://golang.org)
[![Gin Framework](https://img.shields.io/badge/Gin-v1.10.1-green.svg)](https://github.com/gin-gonic/gin)
[![GORM](https://img.shields.io/badge/Gorm-v1.30.3-orange.svg)](https://gorm.io)

> 🚀 木雷坞开源的一个功能完善、高性能的企业级短链接服务平台，支持多域名、AB测试、用户管理、实时统计等功能。

### ✨ 新特性：独立部署模式
- 🎯 **零依赖部署**: 支持SQLite + 内存缓存，无需安装数据库和Redis
- ⚡ **快速启动**: 单文件部署，下载即用
- 💡 **新手友好**: 适合个人用户和小型项目快速上手
- 🔧 **灵活选择**: 支持独立模式和完整模式，满足不同场景需求

### 开源地址

1. 后端
   - CNB [https://cnb.cool/mliev/dwz/dwz-server](https://cnb.cool/mliev/dwz/dwz-server)
   - Gitee [https://gitee.com/muleiwu/dwz-server](https://gitee.com/muleiwu/dwz-server)
   - GitHub [https://github.com/muleiwu/dwz-server](https://github.com/muleiwu/dwz-server)
2. 界面
   - CNB [https://cnb.cool/mliev/open/dwz-admin-webui](https://cnb.cool/mliev/open/dwz-admin-webui)
   - Gitee [https://gitee.com/muleiwu/dwz-admin-webui](https://gitee.com/muleiwu/dwz-admin-webui)
   - GitHub[https://github.com/muleiwu/dwz-admin-webui](https://github.com/muleiwu/dwz-admin-webui)
3. 文档地址
   - https://www.mliev.com/docs/dwz

###  📞 加群获取帮助

|                                     QQ                                      |                                 企业微信                                       |
|:---------------------------------------------------------------------------:|:--------------------------------------------------------------------------:|
| ![wechat_qr_code.png](https://static.1ms.run/dwz/image/httpsn3.inklmKc.png) | ![wechat_qr_code.png](https://static.1ms.run/dwz/image/wechat_qr_code.png) |
|       QQ群号：1021660914 <br /> [点击链接加入群聊【木雷坞开源家】](https://n3.ink/lmKc)        |                                扫描上方二维码加入微信群                                |



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

### 🚀 部署模式
- **独立模式**: SQLite + 内存缓存，零依赖部署，适合个人和小型项目
- **完整模式**: MySQL/PostgreSQL + Redis，适合生产环境和高并发场景
- **灵活切换**: 支持运行时配置切换，满足不同阶段需求

## 🏗️ 技术架构

### 技术栈
- **语言**: Go 1.23+
- **Web框架**: Gin
- **数据库**: MySQL/PostgreSQL/SQLite (支持独立部署)
- **缓存**: Redis/内存缓存 (支持独立部署)
- **ORM**: GORM
- **配置管理**: Viper
- **日志**: Zap
- **HTTP客户端**: go-resty

### 架构设计
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │   Mobile App    │    │   API Client    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
       │                       │                         │
       └───────────────────────┼─────────────────────────┘
                               │
     ┌─────────────────────────────────────────────────────┐
     │                   Load Balancer                     │
     └─────────────────────────────────────────────────────┘
                               │
     ┌─────────────────────────────────────────────────────┐
     │                  DWZ Server                         │
     │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │
     │  │ Controller  │  │ Middleware  │  │   Router    │  │
     │  └─────────────┘  └─────────────┘  └─────────────┘  │
     │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │
     │  │  Service    │  │     DAO     │  │   Model     │  │
     │  └─────────────┘  └─────────────┘  └─────────────┘  │
     └─────────────────────────────────────────────────────┘
                               │
     ┌─────────────────────────────────────────────────────┐
     │                     Data Layer                      │
     │       ┌─────────────┐         ┌─────────────┐       │
     │       │   MySQL     │         │    Cache    │       │
     │       │ PostgreSQL  │         │    Redis    │       │
     │       │   SQLite    │         │    Memory   │       │
     │       └─────────────┘         └─────────────┘       │
     └─────────────────────────────────────────────────────┘
```

### 部署模式

#### 完整模式（生产环境）
- 使用 MySQL/PostgreSQL 作为主数据库
- 使用 Redis 作为缓存和ID生成器
- 支持高并发和集群部署

#### 独立模式（轻量部署）
- 使用 SQLite 作为数据库，无需独立数据库服务
- 使用内存缓存，无需 Redis 服务
- 单文件部署，适合小型项目和个人使用

### 分层架构
- **Controller层**: 处理HTTP请求，参数验证，调用Service
- **Service层**: 业务逻辑处理，事务管理  
- **DAO层**: 数据访问，数据库操作
- **Model层**: 数据模型定义
- **Middleware层**: 认证、日志、CORS等中间件


## 🔧 快速安装

### 部署方式选择

#### 方式一：独立部署（推荐新手）
无需安装数据库和Redis，使用SQLite和内存缓存，适合个人使用和小型项目。

#### 方式二：完整部署
使用MySQL/PostgreSQL和Redis，适合生产环境和高并发场景。

---

## 🚀 独立部署（无需外部依赖）

### 1. 下载可执行文件
```bash
# 创建项目目录
mkdir mliev-dwz
cd mliev-dwz

# 下载最新版本（以Linux x86_64为例）
wget https://github.com/muleiwu/dwz-server/releases/latest/download/dwz-server_Linux_x86_64.tar.gz
tar -xzf dwz-server_Linux_x86_64.tar.gz
chmod +x dwz-server
```

### 2. 启动服务
```bash
# 启动服务
./dwz-server

# 后台运行
nohup ./dwz-server > dwz.log 2>&1 &
```

### 3. 访问系统
打开浏览器访问 `http://localhost:8080` 进行初始化配置。

---

## 🐳 Docker 部署

### 1. 创建项目目录
```bash
mkdir mliev-dwz
cd mliev-dwz
```

### 2. 创建 Docker Compose 文件


启动后，后台地址是 `http://{ip}:{端口}/admin/`

#### 创建 `docker-compose.yml` 文件：


```yaml
version: '3.8'

services:
  dwz-server:
    container_name: dwz-server
    image: docker.cnb.cool/mliev/open/dwz-server:latest
    restart: always
    ports:
      - "8080:8080"
    volumes:
      - "./config/:/app/config/"
    environment:
      - TZ=Asia/Shanghai
      - GIN_MODE=release
```

### 3. 创建配置目录
```bash
mkdir -p config
chmod 666 ./config
```

### 4. 启动服务
```bash
# 后台启动所有服务
docker-compose up -d

# 或者前台启动（可以看到日志）
docker-compose up
```

### 5. 验证安装
```bash
# 检查服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f
```

### 6. 页面配置

打开 `http://{您的IP}:8080` 进行继续配置（请注意8080端口放开）

> **提示**：独立模式无需配置数据库和Redis，系统会自动使用SQLite和内存缓存。

## 🚀 系统预览

![Snipaste_2025-07-16_01-30-57.png](https://static.1ms.run/dwz/image/Snipaste_2025-07-16_01-30-57.png)

![Snipaste_2025-07-16_01-32-13.png](https://static.1ms.run/dwz/image/Snipaste_2025-07-16_01-32-13.png)

![Snipaste_2025-07-16_01-32-59.png](https://static.1ms.run/dwz/image/Snipaste_2025-07-16_01-32-59.png)

![Snipaste_2025-07-16_01-33-14.png](https://static.1ms.run/dwz/image/Snipaste_2025-07-16_01-33-14.png)

![Snipaste_2025-07-16_01-33-45.png](https://static.1ms.run/dwz/image/Snipaste_2025-07-16_01-33-45.png)

![Snipaste_2025-07-16_01-33-56.png](https://static.1ms.run/dwz/image/Snipaste_2025-07-16_01-33-56.png)

![Snipaste_2025-07-16_01-34-36.png](https://static.1ms.run/dwz/image/Snipaste_2025-07-16_01-34-36.png)

![Snipaste_2025-07-16_01-34-59.png](https://static.1ms.run/dwz/image/Snipaste_2025-07-16_01-34-59.png)

![Snipaste_2025-07-16_01-35-19.png](https://static.1ms.run/dwz/image/Snipaste_2025-07-16_01-35-19.png)

![Snipaste_2025-07-16_01-35-56.png](https://static.1ms.run/dwz/image/Snipaste_2025-07-16_01-35-56.png)

![Snipaste_2025-07-16_01-36-07.png](https://static.1ms.run/dwz/image/Snipaste_2025-07-16_01-36-07.png)

![Snipaste_2025-07-16_01-36-18.png](https://static.1ms.run/dwz/image/Snipaste_2025-07-16_01-36-18.png)

![Snipaste_2025-07-16_01-36-35.png](https://static.1ms.run/dwz/image/Snipaste_2025-07-16_01-36-35.png)

![Snipaste_2025-07-16_01-36-59.png](https://static.1ms.run/dwz/image/Snipaste_2025-07-16_01-36-59.png)


## 🚀 二次开发

### 环境要求
- Go 1.23+
- Node.js 22+ / pnpm 9.0+（前端构建）
- MySQL 5.7+ 或 PostgreSQL 9.6+（可选，支持 SQLite）
- Redis 6.0+（可选，支持内存缓存）

### 手动打包

详细的手动打包教程请参考：[手动打包教程](docs/manual-build.md)

快速构建命令：
```bash
# 1. 构建前端
cd admin-webui && pnpm install && pnpm run build:antd --filter=\!./docs && cd ..

# 2. 复制前端产物
mkdir -p static/admin && cp -r admin-webui/apps/web-antd/dist/* static/admin/

# 3. 构建后端
CGO_ENABLED=0 go build -ldflags="-s -w" -o dwz-server main.go
```

### 开发步骤

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

详细的API文档请参考 [API.md](temp/docs/API.md)

## 🔧 配置说明

### 配置文件结构

#### 独立模式配置（推荐新手）
```yaml
server:
  mode: release
  addr: ":8080"

# 数据库配置 - SQLite（无需外部数据库）
database:
  driver: sqlite
  filepath: "./config/sqlite.db"

# 缓存配置 - 内存缓存（无需Redis）
cache:
  driver: local

# ID生成器配置 - 本地模式（无需Redis）
id_generator:
  driver: local

# 短链接配置
shortlink:
  domain: "http://localhost:8080"
  length: 6
  custom_length: true

# JWT配置
jwt:
  secret: "your-secret-key-change-this"
  expire_hours: 24
```

#### 完整模式配置（生产环境）
```yaml
server:
  mode: release
  addr: ":8080"

# 数据库配置 - MySQL/PostgreSQL
database:
  driver: mysql  # 或 postgresql
  host: "localhost"
  port: 3306
  username: "root"
  password: "password"
  dbname: "dwz_db"

# Redis配置
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 5

# 缓存配置 - Redis
cache:
  driver: redis

# ID生成器配置 - Redis
id_generator:
  driver: redis

# 短链接配置
shortlink:
  domain: "http://localhost:8080"
  length: 6
  custom_length: true

# JWT配置
jwt:
  secret: "your-secret-key-change-this"
  expire_hours: 24

# 中间件配置
middleware:
  operation_log:
    enable: true
    max_request_size: 1048576
    sensitive_fields: ["password", "token"]
    async_logging: true
```

### 环境变量

#### 通用配置
- `SERVER_ADDR`: 服务端口 (默认: :8080)
- `SERVER_MODE`: 运行模式 (debug/release/test)

#### 数据库配置
- `DATABASE_DRIVER`: 数据库类型 (mysql/postgresql/sqlite)
- `DATABASE_HOST`: 数据库主机 (SQLite不需要)
- `DATABASE_PORT`: 数据库端口 (SQLite不需要)
- `DATABASE_USERNAME`: 数据库用户名 (SQLite不需要)
- `DATABASE_PASSWORD`: 数据库密码 (SQLite不需要)
- `DATABASE_NAME`: 数据库名称 (SQLite不需要)
- `DATABASE_FILEPATH`: SQLite文件路径 (仅SQLite需要)

#### 缓存配置
- `CACHE_DRIVER`: 缓存驱动 (redis/local)

#### ID生成器配置
- `ID_GENERATOR_DRIVER`: ID生成器驱动 (redis/local)

#### Redis配置（仅当使用Redis时）
- `REDIS_HOST`: Redis主机
- `REDIS_PORT`: Redis端口
- `REDIS_PASSWORD`: Redis密码
- `REDIS_DB`: Redis数据库编号

#### 短链接配置
- `SHORTLINK_DOMAIN`: 短链接域名
- `SHORTLINK_LENGTH`: 短码长度
- `SHORTLINK_CUSTOM_LENGTH`: 是否允许自定义长度

#### JWT配置
- `JWT_SECRET`: JWT密钥
- `JWT_EXPIRE_HOURS`: JWT过期时间（小时）

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

本项目可以二次开发用于商业用途，但是禁止发布衍生版本。具体见 [授权协议](LICENSE)

## 🙏 致谢

感谢所有贡献者的努力和开源社区的支持！


### 贡献者

- 小谈谈 [@bh1xaq](https://cnb.cool/bh1xaq)

⭐ 如果这个项目对您有帮助，请给我们一个星标！
