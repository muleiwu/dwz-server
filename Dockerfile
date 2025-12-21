FROM node:22-slim AS builder-web

ENV TZ=Asia/Shanghai

RUN npm i -g corepack

WORKDIR /app

# copy package.json and pnpm-lock.yaml to workspace
COPY ./admin-webui /app

# 安装 pnpm
RUN pnpm config set registry https://mirrors.cloud.tencent.com/npm/ && pnpm install && pnpm build:antd --filter=\!./docs

# 构建阶段
FROM golang:1.25.0-alpine AS builder-go

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk --no-cache add ca-certificates git tzdata

# 设置时区
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 设置Go环境变量
#ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .
# 复制构建产物到 nginx 的默认静态文件目录
COPY --from=builder-web /app/apps/web-antd/dist /app/static/admin

# 声明构建参数
ARG APP_VERSION
ARG BUILD_TIME
ARG GIT_COMMIT

# 构建应用
RUN go build -ldflags="-s -w \
    -X 'main.Version=${APP_VERSION}' \
    -X 'main.BuildTime=${BUILD_TIME}' \
    -X 'main.GitCommit=${GIT_COMMIT}'" \
    -o dwz main.go

# 运行阶段
FROM alpine:latest

# 安装ca证书和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder-go /app/dwz .

# 创建日志目录
RUN mkdir -p logs && \
    mkdir -p config


# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health/simple || exit 1

# 启动应用
CMD ["./dwz"]