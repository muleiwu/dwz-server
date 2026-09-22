# 项目结构

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

## 请求分层

业务请求主要沿 **Controller → Service → DAO → Database** 分层流转，缓存、Redis、ID 发号器与 IP 归属地解析作为旁路依赖。短码 `GET /:code` 不进入常规路由，而是由全局中间件链（CORS → 短码分发 → 安装检查）拦截后直接转给跳转控制器。

启动过程按「装配链（Env → Config → Logger → Installed → … → Cache）→ 服务器链（Migration → IDGenerator → HttpServer）」执行，未安装时自动降级进入安装向导。

## 架构图

以上链路的可视化见 [架构图目录](./archify/README.md)（中英双语，交互式 HTML）：

- [短码跳转时序](./archify/zh-CN/redirect-flow.html) — 请求分发、缓存回源、安全门控与 A/B 分流
- [启动装配流程](./archify/zh-CN/startup-assembly.html) — 装配链、服务器链与安装门控
