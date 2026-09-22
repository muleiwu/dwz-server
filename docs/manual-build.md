# 手动打包教程

本文档详细介绍如何手动构建和打包 DWZ 短网址服务。

## 📋 环境要求

### 前端构建环境
- Node.js 22.23.2+
- pnpm 10.34.5（与前端 packageManager 一致）

### 后端构建环境
- Go 1.26.8+
- goreleaser（可选，用于跨平台构建）

## 🎨 前端打包

前端项目位于 `admin-webui` 目录，基于 Vue 3 + Ant Design Vue 开发。

### 1. 安装依赖

```bash
# 进入前端目录
cd admin-webui

# 安装 pnpm（如果未安装）
npm install -g pnpm

# 安装项目依赖
pnpm install
```

### 2. 构建生产版本

```bash
# 构建 Ant Design Vue 版本
pnpm run build:antd

# 或者排除文档构建（推荐，速度更快）
pnpm run build:antd --filter=\!./docs
```

### 3. 构建产物

构建完成后，产物位于：
```
admin-webui/apps/web-antd/dist/
```

## 🔧 后端打包

后端使用 Go 语言开发，支持多种打包方式。

### 方式一：简单构建（当前平台）

```bash
# 在项目根目录执行
go mod download
go build -o dwz-server main.go
```

#### 带版本信息构建

```bash
# 设置构建变量
APP_VERSION="v1.0.0"
BUILD_TIME=$(date +"%Y-%m-%d %H:%M:%S")
GIT_COMMIT=$(git rev-parse --short HEAD)
ENVIRONMENT="production"

# 构建
go build -ldflags="-s -w \
    -X 'main.Version=${APP_VERSION}' \
    -X 'main.BuildTime=${BUILD_TIME}' \
    -X 'main.GitCommit=${GIT_COMMIT}' \
    -X 'main.Environment=${ENVIRONMENT}'" \
    -o dwz-server main.go
```

### 方式二：静态链接构建（推荐部署使用）

静态链接可以减少运行时依赖，便于在各种 Linux 发行版上运行。

```bash
# 禁用 CGO，启用静态链接
CGO_ENABLED=0 go build -a -installsuffix cgo \
    -tags "netgo osusergo" \
    -ldflags="-s -w -extldflags '-static' \
        -X 'main.Version=${APP_VERSION}' \
        -X 'main.BuildTime=${BUILD_TIME}' \
        -X 'main.GitCommit=${GIT_COMMIT}'" \
    -o dwz-server main.go
```

### 方式三：跨平台构建

#### Linux AMD64

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
    -ldflags="-s -w" \
    -o dwz-server-linux-amd64 main.go
```

#### Linux ARM64

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build \
    -ldflags="-s -w" \
    -o dwz-server-linux-arm64 main.go
```

#### macOS AMD64

```bash
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build \
    -ldflags="-s -w" \
    -o dwz-server-darwin-amd64 main.go
```

#### macOS ARM64 (Apple Silicon)

```bash
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build \
    -ldflags="-s -w" \
    -o dwz-server-darwin-arm64 main.go
```

#### Windows AMD64

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build \
    -ldflags="-s -w" \
    -o dwz-server-windows-amd64.exe main.go
```

#### 龙芯 LoongArch64

```bash
GOOS=linux GOARCH=loong64 CGO_ENABLED=0 go build \
    -ldflags="-s -w" \
    -o dwz-server-linux-loong64 main.go
```

### 方式四：使用 GoReleaser（推荐批量构建）

项目已配置 `.goreleaser.yaml`，支持一键构建多平台版本。

```bash
# 安装 goreleaser
go install github.com/goreleaser/goreleaser@latest

# 本地快照构建（不发布）
goreleaser release --snapshot --clean

# 构建产物位于 dist/ 目录
```

## 📦 完整打包（前端 + 后端）

将前端和后端打包到一起，实现单文件部署。

### 手动步骤

```bash
# 1. 构建前端
cd admin-webui
pnpm install
pnpm run build:antd --filter=\!./docs

# 2. 复制前端产物到后端静态目录
cd ..
mkdir -p static/admin
cp -r admin-webui/apps/web-antd/dist/* static/admin/

# 3. 构建后端（前端资源会被嵌入）
CGO_ENABLED=0 go build -ldflags="-s -w" -o dwz-server main.go
```

### 一键脚本

创建 `build.sh` 脚本：

```bash
#!/bin/bash

set -e

# 版本信息
APP_VERSION=${APP_VERSION:-"dev"}
BUILD_TIME=$(date +"%Y-%m-%d %H:%M:%S")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
ENVIRONMENT=${ENVIRONMENT:-"production"}

echo "🎨 正在构建前端..."
cd admin-webui
pnpm install
pnpm run build:antd --filter=\!./docs
cd ..

echo "📁 复制前端产物..."
mkdir -p static/admin
rm -rf static/admin/*
cp -r admin-webui/apps/web-antd/dist/* static/admin/

echo "🔧 正在构建后端..."
CGO_ENABLED=0 go build -a -installsuffix cgo \
    -tags "netgo osusergo" \
    -ldflags="-s -w \
        -X 'main.Version=${APP_VERSION}' \
        -X 'main.BuildTime=${BUILD_TIME}' \
        -X 'main.GitCommit=${GIT_COMMIT}' \
        -X 'main.Environment=${ENVIRONMENT}'" \
    -o dwz-server main.go

echo "✅ 构建完成！"
echo "   可执行文件: ./dwz-server"
echo "   版本: ${APP_VERSION}"
echo "   提交: ${GIT_COMMIT}"
```

使用方法：

```bash
chmod +x build.sh
./build.sh

# 或指定版本号
APP_VERSION=v1.0.0 ./build.sh
```

## 🛠️ 本地开发与测试

### 克隆仓库（含管理端子模块）

```bash
git clone --recurse-submodules https://github.com/muleiwu/dwz-server.git
cd dwz-server
```

### 运行后端（使用已构建的管理端资源）

```bash
go mod download
go run . start
```

### 前端热更新开发

修改管理端时，在另一个终端启动 Vite：

```bash
cd admin-webui
corepack enable
pnpm install
pnpm dev:antd
```

### 运行测试

```bash
go test ./...
```

依赖安全升级与复查命令见 [依赖安全记录](./dependency-security.md)。

## 🐳 Docker 镜像构建

### 标准构建

```bash
# 构建镜像
docker build -t dwz-server:latest .

# 带版本参数构建
docker build \
    --build-arg APP_VERSION=v1.0.0 \
    --build-arg BUILD_TIME="$(date +%Y-%m-%d\ %H:%M:%S)" \
    --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
    -t dwz-server:v1.0.0 .
```

### 多架构构建

使用 Docker Buildx 构建多架构镜像：

```bash
# 创建并使用 buildx 构建器
docker buildx create --name multiarch --use

# 构建并推送多架构镜像
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --build-arg APP_VERSION=v1.0.0 \
    -t your-registry/dwz-server:latest \
    --push .
```

### 龙芯架构构建

```bash
docker buildx build \
    --platform linux/loong64 \
    --build-arg APP_VERSION=v1.0.0 \
    -f Dockerfile.loong64 \
    -t dwz-server:loong64 .
```

## 📁 目录结构说明

```
dwz-server/
├── admin-webui/                 # 前端项目
│   ├── apps/
│   │   └── web-antd/
│   │       └── dist/            # 前端构建产物
│   └── ...
├── static/
│   └── admin/                   # 嵌入的前端静态文件
├── dist/                        # goreleaser 构建产物
│   ├── dwz-server_Darwin_arm64.tar.gz
│   ├── dwz-server_Darwin_x86_64.tar.gz
│   ├── dwz-server_Linux_arm64.tar.gz
│   ├── dwz-server_Linux_x86_64.tar.gz
│   ├── dwz-server_Linux_loong64.tar.gz
│   └── ...
├── Dockerfile                   # 标准 Dockerfile
├── Dockerfile.loong64           # 龙芯架构 Dockerfile
├── .goreleaser.yaml             # GoReleaser 配置
└── main.go                      # 入口文件
```

## 🔍 常见问题

### 1. 前端构建失败

```bash
# 清理缓存重试
cd admin-webui
rm -rf node_modules
pnpm store prune
pnpm install
```

### 2. Go 依赖下载慢

```bash
# 设置 Go 代理
export GOPROXY=https://goproxy.cn,direct
go mod download
```

### 3. 跨平台编译失败

确保已设置 `CGO_ENABLED=0`，否则可能需要安装对应平台的交叉编译工具链。

### 4. 构建产物过大

使用 `-ldflags="-s -w"` 参数移除调试信息：
- `-s`: 去除符号表
- `-w`: 去除 DWARF 调试信息

### 5. 静态文件未嵌入

确保 `static/admin` 目录存在且包含前端构建产物，然后重新构建后端。

## 📚 参考资料

- [Go 交叉编译官方文档](https://go.dev/doc/install/source#environment)
- [GoReleaser 文档](https://goreleaser.com/)
- [Docker Buildx 多架构构建](https://docs.docker.com/buildx/working-with-buildx/)

