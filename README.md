<p align="center">
  <img src="docs/brand/logo-lockup.svg" alt="木雷短网址" height="72">
</p>

<h1 align="center">木雷短网址</h1>

<p align="center">
  面向团队与私有化部署的短链接管理平台
</p>

<p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white" alt="Go 1.26"></a>
  <a href="https://github.com/gin-gonic/gin"><img src="https://img.shields.io/badge/Gin-1.12.0-008ECF" alt="Gin 1.12.0"></a>
  <a href="https://gorm.io/"><img src="https://img.shields.io/badge/GORM-1.31.1-00A1EA" alt="GORM 1.31.1"></a>
</p>

木雷短网址（dwz-server）集短链生成、多域名、访问策略、A/B 测试、营销归因和统计分析于一体。服务端使用 Go 开发，管理端基于 Vue 3 与 Ant Design Vue，并随二进制或 Docker 镜像一同分发。

项目既可以使用 SQLite、内存缓存和本地发号器单机运行，也可以接入 MySQL/PostgreSQL 与 Redis，用于生产环境和多实例部署。

> **部署与首次使用的关键前置**：创建正式短链前，必须先配置公开短域名的 DNS、证书和反向代理。程序使用应用实际收到的 HTTP `Host + 短码` 精确查找记录；Host 被改成内部地址、容器名或错误端口时，后台和健康检查可能正常，但公开短链仍会返回 404。详见[首次使用说明](https://mdoc.cc/mliev/dwz/v2.22.3/391)和[宝塔面板安装教程](https://mdoc.cc/mliev/dwz/v2.22.3/335)。

> 使用、修改或部署本项目之前，请先阅读 [LICENSE](LICENSE)。本项目允许在协议范围内免费用于商业或非商业用途，但不是标准开源许可证；必须保留版权标识，也不得重新分发派生版本。

## 功能概览

| 模块 | 当前能力 |
| --- | --- |
| 短链管理 | 创建、编辑、删除、批量操作、自定义短码、过期时间、启停控制、301/302/307/308 跳转、重复 URL 策略 |
| 域名管理 | 多域名、HTTP/HTTPS、参数透传、短码生成策略、微信/QQ 防红引导、站点备案信息 |
| 高级路由 | 按国家/省市、设备、浏览器、操作系统、语言、Referer、查询参数分流；支持优先级、条件组、兜底地址和规则测试 |
| 链接安全 | 访问密码、访问时间窗、最大访问次数、IP allowlist/blocklist、Bot 策略、URL 规则扫描、安全事件和滥用举报 |
| A/B 测试 | 多变体、平均/权重分流、会话一致性、点击与转化统计、签名反馈 Token、幂等事件回传 |
| 数据分析 | 点击明细、独立 IP、地域、来源、时间、设备、浏览器、OS、Bot、UTM 与路由维度分析，支持 CSV 导出 |
| 营销管理 | Campaign、Tag、UTM Builder、活动聚合报表 |
| 工作区 | 多工作区切换、成员管理，以及 owner/admin/member/viewer 四级角色 |
| 用户与认证 | 用户管理、JWT 登录、API Bearer Token、请求签名认证、OIDC 登录与账号绑定、操作日志 |
| 品牌与前端 | 系统品牌名称与 Logo、公开页品牌展示、管理后台、前端二维码样式生成与下载 |
| 运维 | 首次安装向导、数据库迁移、健康检查、嵌入式静态资源与模板、前台/守护进程运行 |

二维码目前由管理端直接生成和下载，尚未作为服务端资源持久化；功能边界和后续计划可查看 [功能路线图](docs/feature-roadmap.md)。

## 界面预览

### 工作台与全球访问热力

![工作台概览与全球访问热力图](https://static.1ms.run/mdoc/uploads/2026/07/21/a8ef0c61-d5ac-460d-9709-aee55315be35.png)

### 短链接管理

![短链接管理列表](https://static.1ms.run/mdoc/uploads/2026/07/21/0a0148ac-9d1b-4356-b312-10113fc4b613.png)

### 多维统计分析

![短链接多维统计分析](https://static.1ms.run/mdoc/uploads/2026/07/21/c3b5fb50-1622-411d-be53-246c97ea06a7.png)

### A/B 测试

![A/B 测试变体统计](https://static.1ms.run/mdoc/uploads/2026/07/21/7270a465-c634-4578-b388-b601c45c1e9b.png)

## 部署模式

| 场景 | 数据库 | 缓存 | ID 发号器 | 说明 |
| --- | --- | --- | --- | --- |
| 轻量单机 | SQLite | memory/local | local | 无外部服务，适合个人、演示和小规模使用 |
| 标准生产 | MySQL 或 PostgreSQL | Redis | Redis | 适合持续运行、高并发和多实例部署 |

多实例部署必须让各实例共用数据库，并建议使用 Redis 发号器，避免本地计数器在实例之间产生冲突。

## Docker 快速开始

创建一个空目录，并写入以下 `compose.yaml`：

```yaml
services:
  dwz-server:
    image: docker.cnb.cool/mliev/dwz/dwz-server:latest
    container_name: dwz-server
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      TZ: Asia/Shanghai
    volumes:
      - ./config:/app/config
      - ./data:/app/data
      - ./logs:/app/logs
```

启动服务：

```bash
mkdir -p config data logs
docker compose up -d
docker compose logs -f dwz-server
```

浏览器打开 <http://localhost:8080>。未安装时服务会自动跳转至 `/install/index`，按向导完成以下配置：

1. 选择 SQLite，或填写已有的 MySQL/PostgreSQL 连接。
2. 单机模式选择内存缓存与本地发号器；使用 Redis 时填写 Redis 连接。
3. 创建首个系统管理员。

安装完成后会生成 `config/config.yaml` 和 `config/install.lock`，执行数据库迁移并自动重启服务。上述三个挂载目录应持续保留；其中 `data` 还用于保存上传的品牌 Logo。

健康检查：

```bash
curl http://localhost:8080/health/simple
```

升级前请备份配置目录、数据目录和外部数据库，然后执行：

```bash
docker compose pull
docker compose up -d
```

使用宝塔面板部署时，请参阅 [宝塔面板安装教程](https://mdoc.cc/mliev/dwz/v2.22.3/335)。教程保留应用商店与容器编排两种入口，并说明域名代理、HTTPS 和数据库/Redis 的安全边界。

## 使用发布版二进制

从 [GitHub Releases](https://github.com/muleiwu/dwz-server/releases) 下载与操作系统、CPU 架构匹配的压缩包，解压到一个可写的固定目录：

```bash
chmod +x dwz-server
./dwz-server start
```

首次启动后访问 <http://localhost:8080> 完成安装。轻量模式不需要预先安装数据库或 Redis。

进程管理命令：

```bash
# 守护进程模式启动
./dwz-server start --daemon

# 查看、重启和停止守护进程
./dwz-server status
./dwz-server restart
./dwz-server stop
```

修改配置后建议重启进程。前台运行时使用 `Ctrl+C` 停止，再重新执行 `./dwz-server start`。

## 从源码构建

### 环境要求

- Go 1.26.8+
- Node.js 22.23.2+（版本记录于前端 `.node-version`）
- pnpm 10.34.5，使用前端 `packageManager` 指定的版本
- Git（管理端位于 `admin-webui` 子模块）

依赖安全升级与复查命令见 [依赖安全记录](docs/dependency-security.md)。

### 构建前端与后端

```bash
git clone --recurse-submodules https://github.com/muleiwu/dwz-server.git
cd dwz-server

corepack enable
cd admin-webui
pnpm install
pnpm run build:antd --filter='!./docs'
cd ..

mkdir -p static/admin
cp -R admin-webui/apps/web-antd/dist/. static/admin/

go mod download
CGO_ENABLED=0 go build -o dwz-server .
./dwz-server start
```

`templates`、`static` 和数据库迁移会通过 `go:embed` 编译进可执行文件，因此生产部署不需要额外复制这些目录。

更多平台构建和 GoReleaser 用法见 [手动构建文档](docs/manual-build.md)。也可以直接在仓库根目录构建镜像：

```bash
docker build -t dwz-server:local .
```

### 本地开发与测试

仅运行当前已构建的管理端资源和后端：

```bash
go run . start
```

修改管理端时，可在另一个终端启动 Vite：

```bash
cd admin-webui
corepack enable
pnpm install
pnpm dev:antd
```

运行后端测试：

```bash
go test ./...
```

## 配置说明

首次安装后，主配置文件位于 `config/config.yaml`。仓库根目录的 [config.yaml.example](config.yaml.example) 展示了可配置项；环境变量会覆盖同名 YAML 配置。

常用配置如下：

| YAML 键 | 环境变量 | 说明 |
| --- | --- | --- |
| `http.addr` | `HTTP_ADDR` | HTTP 监听地址，默认 `:8080` |
| `http.mode` | `HTTP_MODE` | `debug`、`release` 或 `test` |
| `database.driver` | `DATABASE_DRIVER` | `sqlite`、`mysql` 或 `postgresql` |
| `database.filepath` | `DATABASE_FILEPATH` | SQLite 文件路径 |
| `database.host` | `DATABASE_HOST` | MySQL/PostgreSQL 主机 |
| `database.port` | `DATABASE_PORT` | 数据库端口 |
| `database.dbname` | `DATABASE_DBNAME` | 数据库名 |
| `database.username` | `DATABASE_USERNAME` | 数据库用户名 |
| `database.password` | `DATABASE_PASSWORD` | 数据库密码 |
| `redis.host` | `REDIS_HOST` | Redis 主机 |
| `redis.port` | `REDIS_PORT` | Redis 端口 |
| `redis.password` | `REDIS_PASSWORD` | Redis 密码 |
| `redis.db` | `REDIS_DB` | Redis DB 编号 |
| `cache.driver` | `CACHE_DRIVER` | `memory`/`local`、`redis` 或 `none` |
| `id_generator.driver` | `ID_GENERATOR_DRIVER` | `local` 或 `redis` |
| `jwt.secret` | `JWT_SECRET` | JWT 签名密钥，生产环境必须妥善保管 |
| `jwt.expire_hours` | `JWT_EXPIRE_HOURS` | 登录 Token 有效时长 |

若使用 Docker，容器内数据库主机应填写 Compose 服务名，而不是 `localhost`。例如数据库服务名为 `mysql`，则填写 `mysql`。

静态资源与模板支持 `embed`、`disk` 两种模式，开发时的热更新配置见 [静态资源与模板说明](docs/static-templates.md)。

## 访问入口

| 地址 | 用途 | 是否需要认证 |
| --- | --- | --- |
| `/admin/` | 管理后台 | 登录后使用 |
| `/<short-code>` | 短链跳转 | 否，可能受链接安全策略限制 |
| `/preview/<short-code>` | 短链预览 | 否 |
| `/health` | 详细健康信息 | 否 |
| `/health/simple` | 轻量健康检查 | 否 |
| `/api/v1/auth/login` | 账号密码登录 | 否 |
| `/api/v1/public/*` | 品牌、密码访问、举报、A/B 反馈等公开接口 | 否 |
| `/api/v1/*` | 业务与管理 API | 是 |

受保护 API 支持三种认证方式：

- 登录 JWT：`Authorization: Bearer <token>`
- API Bearer Token：`Authorization: Bearer <token>`
- 请求签名：`X-App-Id`、`X-Signature`、`X-Timestamp`、`X-Nonce`

切换工作区时通过 `X-Workspace-Id` 请求头传递工作区 ID；不传时使用当前用户可访问的默认工作区。

完整接口和认证规则：

- [API 参考](docs/api/API_REFERENCE.md)
- [Bearer Token 认证](docs/api/API_BEARER_AUTH.md)
- [请求签名认证](docs/api/API_SIGNATURE_AUTH.md)

## 项目结构

```text
dwz-server/
├── app/
│   ├── controller/        # HTTP 控制器
│   ├── service/           # 业务逻辑
│   ├── dao/               # GORM 数据访问
│   ├── model/             # 数据模型
│   ├── dto/               # 请求与响应对象
│   └── middleware/        # 安装、认证、工作区、日志、CORS
├── config/autoload/       # 配置项、路由和中间件注册
├── pkg/service/           # 数据库、缓存、Redis、迁移、发号器等基础服务
├── migrations/            # MySQL/PostgreSQL/SQLite 的 Goose 迁移
├── templates/             # 公开页、错误页和安装页模板
├── static/admin/          # 构建后嵌入的管理端资源
├── admin-webui/           # Vue 3 管理端子模块
├── docs/                  # API、构建、品牌和设计文档
└── main.go                # 应用入口与嵌入资源声明
```

请求主要沿以下分层流转：

```text
HTTP 请求
  → 全局中间件（安装检查、短码分发、CORS）
  → API 中间件（操作日志、认证、工作区）
  → Controller → Service → DAO → Database
                         ↘ Cache / Redis / ID Generator / IP Region
```

数据库结构使用 Goose 管理，并为 MySQL、PostgreSQL、SQLite 维护同名迁移。升级或新增迁移前请阅读 [迁移说明](MIGRATION.md)。

## 相关地址

### 服务端

- [CNB](https://cnb.cool/mliev/dwz/dwz-server)
- [GitHub](https://github.com/muleiwu/dwz-server)
- [Gitee](https://gitee.com/muleiwu/dwz-server)

### 管理端

- [CNB](https://cnb.cool/mliev/dwz/dwz-admin-webui)
- [GitHub](https://github.com/muleiwu/dwz-admin-webui)
- [Gitee](https://gitee.com/muleiwu/dwz-admin-webui)

### 文档与交流

- [木雷短网址 v2.22.3 完整文档](https://mdoc.cc/mliev/dwz/v2.22.3)
- [产品介绍](https://mdoc.cc/mliev/dwz/v2.22.3/222) · [快速入门](https://mdoc.cc/mliev/dwz/v2.22.3/223) · [管理后台操作指南](https://mdoc.cc/mliev/dwz/v2.22.3/225)
- [部署与运维配置](https://mdoc.cc/mliev/dwz/v2.22.3/224) · [宝塔面板安装教程](https://mdoc.cc/mliev/dwz/v2.22.3/335)
- [API 与系统集成](https://mdoc.cc/mliev/dwz/v2.22.3/226) · [开发与架构](https://mdoc.cc/mliev/dwz/v2.22.3/249) · [常见问题](https://mdoc.cc/mliev/dwz/v2.22.3/228)
- QQ 群：`1021660914`（[点击加入](https://n3.ink/lmKc)）

## 参与贡献

欢迎提交问题、改进文档、补充测试或发起 Pull Request。提交代码前请至少执行：

```bash
gofmt -w <修改的 Go 文件>
go test ./...
```

请不要修改已经发布的历史迁移；新的结构变更需要在三个数据库方言目录中添加同名迁移文件。

## 许可证

版权所有 © 2025 合肥木雷坞信息技术有限公司。项目允许在许可协议范围内用于商业或非商业场景，但要求保留版权标识，并限制再授权和派生版本再分发。完整条款以 [LICENSE](LICENSE) 为准。

感谢贡献者 [@bh1xaq](https://cnb.cool/bh1xaq) 对项目的帮助。
