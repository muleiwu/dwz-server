# 项目开发规范文档

## 1. 项目概述

本项目是一个基于 Gin 框架的 Go Web 应用，采用分层架构设计。


## 2. 技术栈

- **语言**: Go 1.23.2
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL/PostgreSQL
- **缓存**: Redis
- **配置管理**: Viper
- **日志**: Zap
- **HTTP客户端**: go-resty

## 3. 项目目录结构规范

```
go-web/
├── main.go                 # 程序入口
├── go.mod                  # Go模块定义
├── go.sum                  # 依赖版本锁定
├── README.md               # 项目说明文档
├── .gitignore              # Git忽略文件
├── .goreleaser.yaml        # 发布配置
├── config.yaml.example     # 配置文件示例
├── cmd/                    # 命令行入口
│   └── run.go
├── app/                    # 应用核心代码
│   ├── controller/         # 控制器层（处理HTTP请求）
│   ├── service/            # 服务层（业务逻辑）
│   ├── dao/                # 数据访问层
│   ├── model/              # 数据模型
│   ├── dto/                # 数据传输对象
│   └── middleware/         # 中间件
├── router/                 # 路由配置
├── config/                 # 配置管理
├── helper/                 # 基础支持服务
└── util/                   # 工具函数
```

## 4. 代码格式化规范

### 4.1 自动格式化
- **必须使用 `gofmt`** 格式化所有 Go 代码
- 推荐在IDE中配置保存时自动运行 `gofmt`
- 使用制表符（tab）进行缩进，而非空格
- 无需担心行长度限制，如需换行则增加额外缩进

```bash
# 格式化单个文件
gofmt -w filename.go

# 格式化整个项目
gofmt -w .

# 或使用 go fmt
go fmt ./...
```

### 4.2 导入规范
- 标准库导入放在最前面
- 第三方库导入在中间
- 本项目包导入在最后
- 各组之间用空行分隔

```go
import (
    "fmt"
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "your-project/app/model"
    "your-project/config"
)
```

## 5. 命名规范

### 5.1 包名规范
- **全部小写，单个单词，无下划线**
- 简洁明了，避免缩写（除非是广为人知的缩写）
- 目录名与包名保持一致
- 不要使用复数形式

```go
// 好的包名
package user
package auth
package http

// 避免的包名  
package users
package user_service
package userService
```

### 5.2 文件名规范
- **使用小写字母和下划线（snake_case）**
- 文件名应该清晰表达文件用途
- 避免与包名重复

```go
// 推荐的文件名
user_controller.go
auth_service.go
database_config.go
```

### 5.3 变量和函数命名
- **导出的**变量/函数：MixedCaps（首字母大写）
- **未导出的**变量/函数：mixedCaps（首字母小写）
- **不使用下划线**
- **常量**：使用 MixedCaps，可配合 iota

```go
// 导出的标识符
func ProcessRequest() error { }
var DefaultTimeout = 30 * time.Second

// 未导出的标识符
func processInternal() error { }
var maxRetries = 3

// 常量定义
const (
    StatusPending Status = iota
    StatusProcessing
    StatusCompleted
    StatusFailed
)
```

### 5.4 接口命名
- 单方法接口通常以 -er 结尾
- 接口名应该描述行为而非数据

```go
type Reader interface {
    Read([]byte) (int, error)
}

type Writer interface {
    Write([]byte) (int, error)
}

type UserService interface {
    CreateUser(user *User) error
    GetUser(id int) (*User, error)
}
```

### 5.5 Getter/Setter 命名
- **Getter 方法不使用 Get 前缀**
- **Setter 方法使用 Set 前缀**

```go
type User struct {
    name string
    age  int
}

// Getter - 不使用 Get 前缀
func (u *User) Name() string {
    return u.name
}

// Setter - 使用 Set 前缀
func (u *User) SetName(name string) {
    u.name = name
}
```

## 6. 代码组织规范

### 6.1 分层架构
```
HTTP请求 -> Controller -> Service -> Repository -> Database
          ↓
        DTO/Model
```

### 6.2 各层职责
- **Controller**: 处理HTTP请求，参数验证，调用Service
- **Service**: 业务逻辑处理，事务管理
- **Repository/DAO**: 数据访问，数据库操作
- **Model**: 数据模型定义
- **DTO**: 数据传输对象，API输入输出

### 6.3 接口设计
- **优先使用小接口**，遵循单一职责原则
- **通过接口组合构建复杂功能**
- **在使用处定义接口，而非实现处**

```go
// 小接口
type UserReader interface {
    GetUser(id int) (*User, error)
}

type UserWriter interface {
    CreateUser(user *User) error
    UpdateUser(user *User) error
}

// 接口组合
type UserRepository interface {
    UserReader
    UserWriter
}
```

## 7. 错误处理规范

### 7.1 错误处理原则
- **显式处理每个错误**，不要忽略
- **使用多返回值**，最后一个返回值为 error
- **错误信息应该描述失败的原因和上下文**
- **在调用栈中添加上下文信息**

```go
// 好的错误处理
func GetUser(id int) (*User, error) {
    if id <= 0 {
        return nil, fmt.Errorf("invalid user id: %d", id)
    }
    
    user, err := userRepo.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user %d: %w", id, err)
    }
    
    return user, nil
}

// 调用时的错误处理
user, err := GetUser(123)
if err != nil {
    log.Printf("get user failed: %v", err)
    return err
}
```

### 7.2 自定义错误类型
```go
// 定义错误类型
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}

// 使用错误类型
func ValidateUser(user *User) error {
    if user.Name == "" {
        return &ValidationError{
            Field:   "name",
            Message: "name cannot be empty",
        }
    }
    return nil
}
```

### 7.3 统一响应格式
```go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// 错误码定义使用 iota
const (
    CodeSuccess = iota
    CodeBadRequest
    CodeUnauthorized
    CodeNotFound
    CodeInternalError
)
```

## 8. 并发规范

### 8.1 Goroutine 使用
- **谨慎使用 goroutine**，避免泄露
- **使用 channel 进行通信**，而非共享内存
- **正确处理 goroutine 的生命周期**

```go
// 好的 goroutine 使用模式
func ProcessRequests(ctx context.Context, requests <-chan Request) {
    for {
        select {
        case req := <-requests:
            go func(r Request) {
                defer recover() // 防止 panic 导致程序崩溃
                processRequest(r)
            }(req)
        case <-ctx.Done():
            return
        }
    }
}
```

### 8.2 Channel 使用
- **优先使用有缓冲的 channel**
- **发送方负责关闭 channel**
- **使用 select 处理多个 channel**

```go
// 工作池模式
func WorkerPool(jobs <-chan Job, results chan<- Result) {
    const numWorkers = 3
    
    for i := 0; i < numWorkers; i++ {
        go func() {
            for job := range jobs {
                result := processJob(job)
                results <- result
            }
        }()
    }
}
```

## 9. 文档和注释规范

### 9.1 包文档
- **每个包都应该有包注释**
- **包注释应该以包名开头**

```go
// Package user provides user management functionality.
// It includes user creation, authentication, and profile management.
package user
```

### 9.2 函数和方法文档
- **导出的函数必须有文档注释**
- **注释应该以函数名开头**
- **描述函数的作用、参数和返回值**

```go
// CreateUser creates a new user with the given information.
// It returns the created user and any error encountered.
// The user's password will be hashed before storage.
func CreateUser(name, email, password string) (*User, error) {
    // implementation
}
```

### 9.3 类型文档
```go
// User represents a user in the system.
// It contains basic user information and authentication data.
type User struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"-"`
}
```

## 10. 数据库规范

### 10.1 Model定义
```go
// User represents a user entity in the database.
type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Username  string         `gorm:"uniqueIndex;size:50" json:"username"`
    Password  string         `gorm:"size:100" json:"-"`
    CreatedAt time.Time      `json:"createdAt"`
    UpdatedAt time.Time      `json:"updatedAt"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for User model.
func (User) TableName() string {
    return "users"
}
```

### 10.2 Repository 模式
```go
// UserRepository defines the interface for user data access.
type UserRepository interface {
    Create(user *User) error
    GetByID(id uint) (*User, error)
    GetByUsername(username string) (*User, error)
    Update(user *User) error
    Delete(id uint) error
}

// userRepository implements UserRepository interface.
type userRepository struct {
    db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance.
func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}
```

## 11. API设计规范

### 11.1 RESTful API
- GET: 获取资源
- POST: 创建资源
- PUT: 更新资源（全量）
- PATCH: 更新资源（部分）
- DELETE: 删除资源

### 11.2 URL规范
```
/api/v1/users          # 用户列表
/api/v1/users/{id}     # 特定用户
/api/v1/users/{id}/posts # 用户的文章
```

### 11.3 Controller 实现
```go
// UserController handles user-related HTTP requests.
type UserController struct {
    userService UserService
}

// NewUserController creates a new UserController instance.
func NewUserController(userService UserService) *UserController {
    return &UserController{
        userService: userService,
    }
}

// GetUser handles GET /api/v1/users/{id} requests.
func (c *UserController) GetUser(ctx *gin.Context) {
    id, err := strconv.Atoi(ctx.Param("id"))
    if err != nil {
        ctx.JSON(http.StatusBadRequest, Response{
            Code:    CodeBadRequest,
            Message: "invalid user id",
        })
        return
    }
    
    user, err := c.userService.GetUser(id)
    if err != nil {
        ctx.JSON(http.StatusInternalError, Response{
            Code:    CodeInternalError,
            Message: err.Error(),
        })
        return
    }
    
    ctx.JSON(http.StatusOK, Response{
        Code: CodeSuccess,
        Data: user,
    })
}
```

## 12. 测试规范

### 12.1 测试文件组织
- 测试文件以 `_test.go` 结尾
- 测试函数以 `Test` 开头
- 基准测试函数以 `Benchmark` 开头
- 示例函数以 `Example` 开头

### 12.2 单元测试
```go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        input   *User
        want    *User
        wantErr bool
    }{
        {
            name: "valid user",
            input: &User{
                Username: "testuser",
                Password: "password123",
            },
            want: &User{
                ID:       1,
                Username: "testuser",
            },
            wantErr: false,
        },
        // more test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## 13. 配置管理规范

### 13.1 配置结构定义
```go
// Config represents the application configuration.
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
}

// ServerConfig represents server configuration.
type ServerConfig struct {
    Port int    `mapstructure:"port"`
    Mode string `mapstructure:"mode"`
}
```

## 14. 日志规范

### 14.1 结构化日志
```go
// 使用结构化日志记录关键信息
logger.Info("user created successfully",
    zap.String("username", user.Username),
    zap.Int("userID", user.ID),
    zap.Duration("duration", time.Since(start)),
)

// 错误日志应包含足够的上下文
logger.Error("failed to create user",
    zap.String("username", username),
    zap.Error(err),
    zap.String("trace_id", traceID),
)
```

## 15. 健康检查和监控

### 15.1 健康检查实现
```go
// HealthController handles health check requests.
type HealthController struct {
    db    *gorm.DB
    redis *redis.Client
}

// Check performs comprehensive health check.
func (h *HealthController) Check(ctx *gin.Context) {
    status := gin.H{
        "status":    "ok",
        "timestamp": time.Now(),
    }
    
    // 检查数据库连接
    if err := h.checkDatabase(); err != nil {
        status["database"] = "error: " + err.Error()
        ctx.JSON(http.StatusServiceUnavailable, status)
        return
    }
    status["database"] = "ok"
    
    // 检查 Redis 连接
    if err := h.checkRedis(); err != nil {
        status["redis"] = "error: " + err.Error()
        ctx.JSON(http.StatusServiceUnavailable, status)
        return
    }
    status["redis"] = "ok"
    
    ctx.JSON(http.StatusOK, status)
}
```

## 16. 部署和构建规范

### 16.1 构建脚本
```bash
#!/bin/bash
# build.sh

# 设置构建变量
VERSION=$(git describe --tags --always)
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT=$(git rev-parse HEAD)

# 构建二进制文件
go build -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.Commit=${COMMIT}" -o bin/app main.go
```

### 16.2 版本信息
```go
// 在 main.go 中定义版本信息
var (
    Version   = "dev"
    BuildTime = "unknown"
    Commit    = "unknown"
)

func printVersion() {
    fmt.Printf("Version: %s\n", Version)
    fmt.Printf("Build Time: %s\n", BuildTime)
    fmt.Printf("Commit: %s\n", Commit)
}
```

# 短网址服务 (DWZ Server)

一个功能完整的短网址服务，基于Go语言开发，支持多域名、访问统计、自定义短代码等功能。

## 主要功能

- ✅ **短网址生成**: 支持自动生成和自定义短代码
- ✅ **多域名支持**: 可配置多个短网址域名
- ✅ **访问统计**: 详细的点击统计和分析
- ✅ **过期时间**: 支持设置短网址过期时间
- ✅ **批量操作**: 支持批量创建短网址
- ✅ **缓存机制**: 使用Redis缓存提高性能
- ✅ **RESTful API**: 完整的CRUD操作接口
- ✅ **数据库支持**: 支持MySQL和PostgreSQL

## 技术栈

- **后端框架**: Gin (Go Web Framework)
- **数据库ORM**: GORM
- **缓存**: Redis
- **数据库**: MySQL/PostgreSQL
- **日志**: Zap
- **配置**: Viper

## 项目结构

```
dwz-server/
├── app/
│   ├── controller/         # 控制器层
│   ├── dao/               # 数据访问层
│   ├── dto/               # 数据传输对象
│   ├── middleware/        # 中间件
│   ├── model/            # 数据模型
│   └── service/          # 业务逻辑层
├── cmd/                  # 启动命令
├── config/               # 配置相关
├── constants/            # 常量定义
├── docs/                 # 文档
├── helper/               # 工具类
├── router/               # 路由配置
├── util/                 # 工具函数
├── test/                 # 测试文件
├── config.yaml           # 配置文件
├── main.go              # 程序入口
└── test_api.sh          # API测试脚本
```

## 快速开始

### 1. 环境要求

- Go 1.19+
- MySQL 5.7+ 或 PostgreSQL 12+
- Redis 6.0+

### 2. 克隆项目

```bash
git clone <repository-url>
cd dwz-server
```

### 3. 安装依赖

```bash
go mod tidy
```

### 4. 配置数据库

创建数据库：
```sql
CREATE DATABASE dwz_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 5. 配置文件

复制并修改配置文件：
```bash
cp config.yaml.example config.yaml
```

修改 `config.yaml` 中的数据库和Redis连接信息：
```yaml
database:
  host: localhost
  port: 3306
  username: your_username
  password: your_password
  database: dwz_db

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
```

### 6. 启动服务

```bash
go run main.go
```

或者编译后运行：
```bash
go build -o dwz-server
./dwz-server
```

服务将在 `http://localhost:8080` 启动。

### 7. 验证服务

运行测试脚本验证服务是否正常：
```bash
./test_api.sh
```

或者手动测试健康检查接口：
```bash
curl http://localhost:8080/health
```

## API 使用

### 创建短网址

```bash
curl -X POST http://localhost:8080/api/v1/shortlinks \
  -H "Content-Type: application/json" \
  -d '{
    "original_url": "https://www.example.com",
    "custom_code": "example",
    "title": "Example Website"
  }'
```

### 访问短网址

```bash
curl -L http://localhost:8080/example
```

### 获取统计信息

```bash
curl http://localhost:8080/api/v1/shortlinks/1/statistics?days=30
```

详细的API文档请参考：[API文档](docs/API.md)

## Docker 部署

### 使用 Docker Compose

```bash
# 启动所有服务（包括MySQL和Redis）
docker-compose up -d

# 仅启动应用服务
docker-compose up -d dwz-server
```

### 使用 Dockerfile

```bash
# 构建镜像
docker build -t dwz-server .

# 运行容器
docker run -d \
  --name dwz-server \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  dwz-server
```

## 配置说明

### 主要配置项

```yaml
# 服务器配置
server:
  mode: release          # 运行模式: debug/release/test
  addr: ":8080"         # 监听地址

# 短链接配置
shortlink:
  domain: "http://localhost:8080"  # 默认域名
  length: 6                        # 短代码长度
  custom_length: true              # 是否允许自定义长度

# 限流配置
rate_limit:
  requests_per_minute: 100  # 每分钟请求数限制
  burst: 10                 # 突发请求数
```

### 环境变量

也可以通过环境变量覆盖配置：

```bash
export SERVER_ADDR=":9000"
export DATABASE_HOST="mysql.example.com"
export REDIS_HOST="redis.example.com"
```

## 性能优化

### 1. 缓存策略

- 短网址信息缓存1小时
- 支持缓存预热
- 使用Redis集群提高可用性

### 2. 数据库优化

- 为短代码和域名创建复合索引
- 统计数据按日期分区
- 使用读写分离

### 3. 负载均衡

- 支持水平扩展
- 无状态设计
- 使用Nginx进行负载均衡

## 监控和日志

### 日志配置

```yaml
log:
  level: info                    # 日志级别
  filename: logs/app.log        # 日志文件
  max_size: 100                 # 文件大小限制(MB)
  max_age: 30                   # 保留天数
  max_backups: 10               # 备份文件数
  compress: true                # 是否压缩
```

### 监控指标

- 请求QPS
- 响应时间
- 错误率
- 缓存命中率
- 数据库连接数

## 开发指南

### 添加新功能

1. 在 `app/model/` 中定义数据模型
2. 在 `app/dao/` 中实现数据访问层
3. 在 `app/service/` 中实现业务逻辑
4. 在 `app/controller/` 中实现控制器
5. 在 `router/` 中注册路由

### 代码规范

- 使用 `gofmt` 格式化代码
- 使用 `golint` 检查代码规范
- 编写单元测试
- 添加必要的注释

### 测试

```bash
# 运行单元测试
go test ./...

# 运行集成测试
go test -tags=integration ./test/

# 运行API测试
./test_api.sh
```

## 部署建议

### 生产环境

1. 使用 `release` 模式
2. 配置HTTPS
3. 设置防火墙规则
4. 使用专用数据库用户
5. 定期备份数据
6. 监控系统资源

### 安全考虑

1. 设置强密码
2. 限制API访问频率
3. 使用HTTPS传输
4. 定期更新依赖
5. 监控异常访问

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查数据库服务是否启动
   - 验证连接参数
   - 检查网络连通性

2. **Redis连接失败**
   - 检查Redis服务状态
   - 验证连接配置
   - 检查内存使用情况

3. **短网址无法访问**
   - 检查数据库中的记录
   - 验证缓存状态
   - 查看应用日志

### 日志分析

```bash
# 查看应用日志
tail -f logs/app.log

# 查看错误日志
grep ERROR logs/app.log

# 查看访问统计
grep "请求处理" logs/app.log | grep "GET /"
```

## 许可证

本项目采用 MIT 许可证。详情请参阅 [LICENSE](LICENSE) 文件。

## 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 支持

如果你觉得这个项目有用，请给个 ⭐️ Star！

如有问题，请通过以下方式联系：

- 提交 Issue
- 发送邮件到：admin@example.com

## 更新日志

### v1.0.0 (2024-01-01)

- ✅ 基础短网址功能
- ✅ 多域名支持
- ✅ 访问统计
- ✅ RESTful API
- ✅ 缓存机制
- ✅ 批量操作