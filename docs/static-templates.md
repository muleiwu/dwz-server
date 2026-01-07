# 静态资源与模板加载说明

## 概述

dwz-server 支持灵活的静态资源和模板加载方式，提供 **embed（嵌入模式）** 和 **disk（磁盘模式）** 两种加载策略，并根据运行模式自动优化性能。

## 功能特性

### 1. 双模式支持

#### Embed 模式（嵌入模式）
- 静态资源和模板嵌入到二进制文件中
- 部署简单，只需一个可执行文件
- 适合生产环境

#### Disk 模式（磁盘模式）
- 从磁盘加载静态资源和模板
- 支持实时更新（开发模式）
- 支持内存缓存（生产模式）

### 2. 智能缓存策略

系统根据 `http.mode` 配置自动选择缓存策略：

| http.mode | 缓存行为 | 适用场景 |
|-----------|---------|---------|
| `debug` / `test` | 实时读取磁盘，不缓存 | 开发环境，支持热更新 |
| `release` | 首次读取后缓存到内存 | 生产环境，性能优化 |

### 3. 目录自动发现

- 自动扫描 `static_dir` 配置的目录及所有子目录
- 为每个文件注册独立路由
- 自动处理目录访问（如 `/admin/` → `/admin/index.html`）

## 配置说明

### 配置项

```yaml
http:
  mode: release              # 运行模式：debug, test, release
  addr: ":8080"              # 监听地址
  static_dir: static         # 静态资源根目录
  static_mode: embed         # 静态资源模式：embed 或 disk
  templates_mode: embed      # 模板模式：embed 或 disk
  templates_dir: templates   # 模板目录
```

### 环境变量支持

所有配置项均支持通过环境变量覆盖：

```bash
export HTTP_MODE=release
export HTTP_ADDR=:8080
export HTTP_STATIC_MODE=disk
export HTTP_TEMPLATES_MODE=embed
```

## 使用场景

### 场景 1：开发环境（推荐）

**配置：**
```yaml
http:
  mode: debug              # 开发模式
  static_mode: disk        # 从磁盘加载
  templates_mode: disk     # 从磁盘加载
```

**特点：**
- ✅ 修改静态文件后刷新浏览器即可看到效果
- ✅ 修改模板文件后刷新即可生效
- ✅ 无需重启服务器
- ⚠️ 性能相对较低（每次请求都读取磁盘）

**适用于：** 本地开发、调试

---

### 场景 2：生产环境 - Disk 模式 + 缓存

**配置：**
```yaml
http:
  mode: release            # 生产模式
  static_mode: disk        # 从磁盘加载（启用缓存）
  templates_mode: disk     # 从磁盘加载
```

**特点：**
- ✅ 首次访问读取文件并缓存到内存
- ✅ 后续请求直接返回缓存，响应速度快
- ✅ 可以在运行时替换文件（需重启生效）
- ⚠️ 内存占用会随文件数量增加

**适用于：** 需要在运行时更新静态资源的生产环境

---

### 场景 3：生产环境 - Embed 模式（推荐）

**配置：**
```yaml
http:
  mode: release            # 生产模式
  static_mode: embed       # 嵌入到二进制
  templates_mode: embed    # 嵌入到二进制
```

**特点：**
- ✅ 所有资源打包到可执行文件中
- ✅ 部署简单，无需额外文件
- ✅ 性能最优，无磁盘 IO
- ⚠️ 更新需要重新编译

**适用于：** 标准生产环境部署

---

### 场景 4：混合模式

**配置：**
```yaml
http:
  mode: release
  static_mode: embed       # 静态资源嵌入
  templates_mode: disk     # 模板从磁盘加载
```

**特点：**
- 静态资源（JS/CSS/图片）嵌入，性能最优
- 模板文件从磁盘加载，可灵活修改
- 平衡了性能和灵活性

**适用于：** 需要经常调整页面模板的场景

## 目录结构示例

```
project/
├── static/              # 静态资源根目录
│   ├── admin/           # 管理后台
│   │   ├── index.html
│   │   ├── css/
│   │   │   └── main.css
│   │   └── js/
│   │       └── app.js
│   └── assets/          # 公共资源
│       └── logo.png
└── templates/           # 模板目录
    ├── index.html
    ├── 404.html
    └── error.html
```

## URL 映射规则

### 文件路径映射

| 文件路径 | URL |
|---------|-----|
| `static/admin/index.html` | `/admin/index.html` |
| `static/admin/css/main.css` | `/admin/css/main.css` |
| `static/assets/logo.png` | `/assets/logo.png` |

### 目录访问

访问目录路径时，系统会自动尝试返回 `index.html` 或 `index.htm`：

| URL | 实际返回文件 |
|-----|------------|
| `/admin` | `/admin/index.html` |
| `/admin/` | `/admin/index.html` |

如果目录下不存在 index 文件，返回 404。

## 性能对比

| 模式 | 首次响应 | 后续响应 | 内存占用 | 热更新 |
|------|---------|---------|---------|--------|
| disk + debug | 较慢（读磁盘） | 较慢（读磁盘） | 低 | ✅ |
| disk + release | 较慢（读磁盘） | 极快（内存） | 中等 | ❌ |
| embed | 极快（内存） | 极快（内存） | 中等 | ❌ |

## 常见问题

### Q1: 如何在开发时使用热更新？

**A:** 将 `http.mode` 设置为 `debug` 或 `test`，并使用 `disk` 模式：

```yaml
http:
  mode: debug
  static_mode: disk
  templates_mode: disk
```

### Q2: 生产环境推荐使用哪种模式？

**A:** 推荐使用 embed 模式，部署最简单，性能最优：

```yaml
http:
  mode: release
  static_mode: embed
  templates_mode: embed
```

### Q3: 缓存何时清除？

**A:** 
- 内存缓存会一直保留直到进程重启
- 如需更新已缓存的文件，必须重启服务

### Q4: 如何验证当前使用的模式？

**A:** 启动服务后查看日志输出：

```
Disk 模式（实时读取）：每次请求从磁盘读取，支持热更新
Disk 模式（缓存启用）：首次读取后缓存到内存，提升性能
Embed 模式：使用 embed 驱动加载静态文件
```

### Q5: 静态资源路径配置错误会怎样？

**A:** 系统会记录错误日志并继续运行，但相关资源会返回 404。请确保：
- `static_dir` 指向的目录存在
- 目录中包含需要的文件
- embed 模式下，文件已正确嵌入（检查 `main.go` 的 `//go:embed` 指令）

## 编译说明

### Embed 模式编译

确保 `main.go` 中包含正确的 embed 指令：

```go
//go:embed templates/**
var templateFS embed.FS

//go:embed static/**
var staticFs embed.FS
```

编译时会自动将文件嵌入：

```bash
go build -o dwz-server
```

### Disk 模式运行

无需特殊编译，确保运行时目录结构正确即可：

```bash
./dwz-server
```

## 最佳实践

1. **开发环境**：使用 disk + debug 模式，享受热更新
2. **测试环境**：使用 disk + release 模式，测试缓存行为
3. **生产环境**：使用 embed + release 模式，部署简单性能好
4. **大量静态文件**：优先使用 embed 模式，避免频繁磁盘 IO
5. **频繁更新内容**：考虑使用 CDN 或外部静态资源服务

## 技术细节

### 驱动架构

系统使用驱动模式实现不同的加载策略：

```
StaticFileDriver (接口)
├── DiskStaticDriver  (磁盘驱动，支持缓存)
└── EmbedStaticDriver (嵌入驱动)
```

### 缓存实现

- 使用 `sync.Map` 实现并发安全的缓存
- 缓存键：文件完整路径
- 缓存值：文件字节内容
- 无过期时间，直到进程重启

### MIME 类型

系统自动根据文件扩展名设置正确的 Content-Type：

- `.html` → `text/html; charset=utf-8`
- `.css` → `text/css; charset=utf-8`
- `.js` → `application/javascript; charset=utf-8`
- `.png` → `image/png`
- 更多类型请参考 `mime_mapper.go`

## 相关文件

- 配置文件：`config/autoload/http.go`
- 静态处理器：`internal/pkg/http_server/impl/static_handler.go`
- 磁盘驱动：`internal/pkg/http_server/impl/static_handler/disk_static_driver.go`
- 嵌入驱动：`internal/pkg/http_server/impl/static_handler/embed_static_driver.go`
- MIME 映射：`internal/pkg/http_server/impl/static_handler/mime_mapper.go`

---

**版本**：v1.0  
**更新日期**：2026-01-05

