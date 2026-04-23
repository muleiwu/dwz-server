# 框架迁移说明（dwz-server v2 → go-web）

dwz-server 现已迁移到 [`cnb.cool/mliev/open/go-web`](https://cnb.cool/mliev/open/go-web) 框架。本文记录破坏性变更与 EE 升级步骤。

## 主要变更

| 维度 | 老版本 | 新版本 |
|---|---|---|
| **入口** | `cmd.Start(StartOptions{...})` | `gomander.Run(...) → cmd.Start(WithApp(...))` |
| **DI** | `helper.GetHelper().GetXxx()` 单例 + 手动 SetXxx | go-web 容器：`container.MustGet[T]()` / `helper.GetXxx()` 包级函数 |
| **HTTP 上下文** | `*gin.Context` + `interfaces.HelperInterface` | `httpInterfaces.RouterContextInterface` |
| **路由签名** | `func(*gin.Engine, *impl.HttpDeps)` | `func(httpInterfaces.RouterInterface)` |
| **中间件** | `gin.HandlerFunc` | `httpInterfaces.HandlerFunc` |
| **数据库迁移** | GORM `AutoMigrate` + `database.migration` 配置键 | goose SQL 文件，按方言分目录 (`migrations/{mysql,postgresql,sqlite}/`) |
| **生命周期** | 仅 SIGINT/SIGTERM | gomander 包装：SIGHUP 热重载 + `reload.TriggerReload()` |

## 移除的 API

- `cmd.StartOptions` 整体删除（包括 `Version`、`GitCommit`、`BuildTime`、`StaticFs`、`ExtraRoutes`、`ExtraModels`、`ExtraAssembly`、`ExtraConfigs`、`ExtraMigrationsFS`）
- `interfaces.HelperInterface.SetXxx`（容器生命周期由 go-web 管理，运行时切换通过 SIGHUP reload）
- `pkg/service/{env,config,logger,cache,database,redis,http_server}/`（全部交给 go-web 同名 assembly/server）

## EE 升级步骤

EE 不再通过 `cmd.StartOptions.Extra*` 注入扩展点，而是实现自己的 `interfaces.AppProvider` 包装 CE 的默认链。

### 1. 替换入口

```go
// 旧版
cmd.Start(cmd.StartOptions{
    ExtraAssembly: func(h interfaces.HelperInterface) []interfaces.AssemblyInterface {
        return []interfaces.AssemblyInterface{ &eeAssembly.MyService{Helper: h} }
    },
    ExtraRoutes: func(e *gin.Engine, deps *impl.HttpDeps) { ... },
    ExtraModels: []any{ &eeModel.MyTable{} },
    ExtraMigrationsFS: eeMigrationsFS,
})

// 新版
type EEApp struct{
    MigrationsFS embed.FS // CE migrations
    EEMigrationsFS embed.FS // EE migrations
}

func (a EEApp) Assemblies() []interfaces.AssemblyInterface {
    return append(config.DefaultAssemblies(), &eeAssembly.MyService{})
}

func (a EEApp) Servers() []interfaces.ServerInterface {
    return append(config.DefaultServers(a.MigrationsFS), &eeServer.MyServer{})
}

func main() {
    gomander.Run(func() {
        cmd.Start(
            cmd.WithTemplateFs(eeTemplateFS),
            cmd.WithWebStaticFs(eeStaticFS),
            cmd.WithApp(EEApp{MigrationsFS: ceMigrationsFS, EEMigrationsFS: eeMigrationsFS}),
        )
    })
}
```

### 2. 路由扩展

EE 路由通过 `config/autoload/router.go` 风格的 `InitConfig` 提供者添加。在 EE 自己的 InitConfig 里设置 `http.router_ee`，由 CE router 调用，或 EE 直接覆盖 `http.router` 把 CE+EE 路由合并。

### 3. 模型扩展

不再把 GORM 模型塞进 `database.migration` 配置键。EE 自己写 SQL 迁移文件嵌入 `embed.FS`，目录结构与 CE 一致：

```
ee_migrations/
├── mysql/
│   └── 0010_create_ee_table.sql
├── postgresql/
└── sqlite/
```

然后通过 `helper.GetHelper().GetConfig().Set("ee.extra_migrations_fs", eeFS)` 在 EE assembly 里注入；CE migration server 会在 CE 迁移完成后，针对当前方言子目录追加运行 EE 迁移。

### 4. 控制器迁移

控制器签名从 `func(*gin.Context, interfaces.HelperInterface)` 改为 `func(httpInterfaces.RouterContextInterface)`。helper 通过包级函数获取：

```go
import (
    helperPkg "cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
    httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
)

func (ctrl MyController) MyHandler(c httpInterfaces.RouterContextInterface) {
    helper := helperPkg.GetHelper()
    // 业务逻辑
}
```

`*gin.Context` 上的常用方法在 `RouterContextInterface` 上都有对应：

| Gin | RouterContextInterface |
|---|---|
| `c.Request.URL.Path` | `c.Path()` |
| `c.Request.Host` | `c.Host()` |
| `c.Request.Method` | `c.Method()` |
| `c.Request.Context()` | `c.Request().Context()` |
| `c.Writer.Status()` | `c.GetStatus()` |
| `gin.H{}` | `map[string]any{}` |

需要原始 `*http.Request` / `http.ResponseWriter` 时通过 `c.Request()` / `c.ResponseWriter()` 获取。

### 5. 中间件迁移

中间件签名变为 `func(httpInterfaces.RouterContextInterface)`。同样可以通过 `c.Request()` 拿到底层 `*http.Request`。

### 6. 安装流程变更

安装完成后不再手工 `SetCache/SetDatabase/SetRedis`，而是写入 `config/config.yaml` + 创建 `install.lock` 后调用 `reload.TriggerReload()` 让 gomander 重启服务，所有 assembly 用新配置重新装配。

## 暴露给 EE 的辅助函数

- `config.DefaultAssemblies() []interfaces.AssemblyInterface` — CE 默认装配链
- `config.DefaultServers(migrationsFS embed.FS) []interfaces.ServerInterface` — CE 默认服务链
- `config.DefaultConfigs() []interfaces.InitConfig` — CE 默认 InitConfig 列表
