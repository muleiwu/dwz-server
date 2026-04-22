# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**木雷短网址 (Muleiwu URL Shortener)** is an enterprise-grade URL shortening service written in Go. It supports multi-domain short links, A/B testing, user management, real-time analytics, and can run in two deployment modes:

- **Standalone Mode**: SQLite + in-memory cache, zero external dependencies
- **Production Mode**: MySQL/PostgreSQL + Redis for high concurrency

## Development Commands

```bash
# Run development server
go run main.go

# Run tests
go test ./...

# Run specific test (tests live under pkg/helper/ and test/)
go test ./pkg/helper/ -run TestGenerateAndVerifySignature

# Build binary
go build -o dwz-server main.go

# Run with race detection
go run -race main.go

# Health check
curl http://localhost:8080/health
```

## Go Version

Go 1.25.0 (see go.mod:3)

## Architecture

The application follows a clean architecture pattern with dependency injection:

### Entry Point
- `main.go` - Embeds `templates/**` and `static/**` via `//go:embed`, passes them under keys `templates` and `web.static`, then calls `cmd.Start()`
- `cmd/run.go` - Runs an 8-step init: get CE assemblies → inject EE configs → run CE assemblies → run EE assemblies → store static FS → set version → store EE hooks → start servers
- `cmd/options.go` - `StartOptions` defines the CE/EE boundary (see CE/EE section below)

### Dependency Injection (Assembly Pattern)

Services are initialized in strict dependency order via `config.Assembly` (`config/assembly.go`):
1. Env (environment variables)
2. Config (code defaults + env vars, merged with any `ExtraConfigs` from EE)
3. Logger (Zap)
4. Version
5. Installed (installation check — gates migrations)
6. Database (MySQL/PostgreSQL/SQLite)
7. Redis
8. Cache (Redis/memory)

Each assembly lives under `pkg/service/<name>/assembly/` and implements `interfaces.AssemblyInterface`.

Servers are started via `config.Server` (`config/server.go`):
1. Migration (`pkg/service/migration`) — runs `GORM.AutoMigrate` only when `Installed.IsInstalled()` is true (or `AUTO_INSTALL=install`); merges `ee.extra_models` from config
2. IDGenerator (`pkg/service/id_generator`) — distributed ID generation
3. HttpServer (`pkg/service/http_server`) — Gin HTTP server

### CE / EE Extension Points

The binary in this repo is the Community Edition. Enterprise Edition consumers import `cmd` and pass extra hooks through `StartOptions`:
- `ExtraAssembly` — additional services to initialize
- `ExtraConfigs` — additional config providers merged into the CE Config assembly
- `ExtraRoutes` — additional Gin routes (stashed at `ee.extra_routes`)
- `ExtraModels` — additional GORM models (stashed at `ee.extra_models`, picked up by Migration)
- `ExtraMigrationsFS` — EE-specific migration embed FS

When editing CE code, preserve these hooks: they are nil-safe but consumed by reflection/config keys and silently skipping them breaks EE builds.

### Directory Structure

- `app/controller/` - HTTP request handlers (REST API endpoints)
- `app/service/` - Business logic layer
- `app/dao/` - Data access objects (GORM operations)
- `app/model/` - GORM data models
- `app/dto/` - Data transfer objects
- `app/middleware/` - Gin middleware (auth, CORS, operation logging, install check)
- `app/constants/` - Application constants
- `config/autoload/` - Router + per-concern config providers (base, cache, database, http, id_generator, jwt, middleware, migration, redis, router, static_fs); each implements `InitConfig(helper) map[string]any` and is merged into the global config
- `config/assembly.go` - Service assembly initialization
- `config/server.go` - Server initialization
- `pkg/helper/` - Global DI container (`GetHelper()`), signature/JWT helpers, tests
- `pkg/interfaces/` - Interface definitions for dependency inversion (HelperInterface, AssemblyInterface, ServerInterface, InitConfig, etc.)
- `pkg/service/<name>/` - Each service has an `assembly/` (init) and optionally `service/` (runtime) subpackage — covers cache, config, database, domain_validate, env, http_server, id_generator, installed, logger, migration, redis, version
- `templates/` - Embedded HTML templates (error pages, installation)
- `static/` - Embedded admin web UI build files (populated by the `admin-webui` git submodule build)
- `admin-webui/` - Git submodule pointing at the separate [dwz-admin-webui](https://cnb.cool/mliev/dwz/dwz-admin-webui) repo; run `git submodule update --init --recursive` after clone

### Key Interfaces

The codebase uses interface-based dependency injection defined in `pkg/interfaces/`:
- `AssemblyInterface` - For service initialization (run once at startup)
- `ServerInterface` - For server startup (run in order after assemblies)
- `HelperInterface` - Global DI container providing access to logger, config, cache, database, redis, env, installed, version, id_generator
- `InitConfig` - Implemented by each `config/autoload/*.go` provider; returns a `map[string]any` merged into the global Viper config

### Routing

Routes are defined in `config/autoload/router.go` using Gin framework + `github.com/jxskiss/ginregex` for the short-code regex route:
- `/health`, `/health/simple` - Health checks (no auth)
- `/install/index`, `/api/v1/install/*` - Installation endpoints (no auth)
- `/api/v1/auth/login`, legacy `/api/v1/login` - Login (no auth, operation logged)
- `/api/v1/*` - Protected API endpoints (require auth via `AuthMiddleware` + `OperationLogMiddleware`): `short_links`, `domains`, `ab_tests`, `users`, `profile`, `tokens`, `logs`, `click_statistics`, `ab_test_click_statistics`, `statistics`
- `GET /:code` (regex `[a-zA-Z0-9\-_.]+`) - Short code redirect
- `GET /preview/:code` - Short link preview
- `InstallMiddleware` runs globally and redirects API traffic to install flow when `config/install.lock` is absent

### Authentication

- JWT-based authentication via `middleware.AuthMiddleware()`
- Tokens validated via `Authorization: Bearer <token>` header
- Login endpoint: `POST /api/v1/auth/login`
- Operation logging middleware tracks user actions

### Database Support

Multi-database support via GORM with driver selection:
- MySQL: `gorm.io/driver/mysql`
- PostgreSQL: `gorm.io/driver/postgres`
- SQLite: `github.com/glebarez/sqlite` (pure Go, no CGO)

Database driver is configured via `DATABASE_DRIVER` environment variable or config.

### Caching Strategy

- Redis: `github.com/redis/go-redis/v9` (production)
- Memory: `github.com/patrickmn/go-cache` (standalone)

Cache driver selected via `CACHE_DRIVER` environment variable.

### ID Generation

- Redis-based distributed ID generation (production)
- Local fallback for standalone mode
- Configured via `ID_GENERATOR_DRIVER`

### Static Assets

Templates and static files are embedded in the binary using `//go:embed`:
- `templates/**` - HTML templates
- `static/**` - Admin web UI files
- Passed to services via `static.fs` config key

## Common Patterns

### Controller Pattern

Controllers receive dependencies via `HttpDeps` struct and use `deps.WrapHandler()` to wrap handlers:

```go
func (controller MyController) MyHandler(c *gin.Context, deps *impl.HttpDeps) {
    // Access services via deps
    service := deps.GetShortLinkService()
    logger := deps.GetLogger()
    // ...
}
```

### Service Layer Pattern

Services access database and other dependencies via the global Helper:

```go
helper.GetLogger()
helper.GetConfig()
helper.GetDatabase()
helper.GetCache()
```

### Adding New Features

1. Add model in `app/model/`
2. Add DAO in `app/dao/`
3. Add service in `app/service/`
4. Add DTO in `app/dto/`
5. Add controller in `app/controller/`
6. Add routes in `config/autoload/router.go`
7. Register service provider if needed

## Configuration

Configuration uses Viper with environment variable override support. Key environment variables:

- `SERVER_MODE` - debug/release/test
- `SERVER_ADDR` - bind address (default :8080)
- `DATABASE_DRIVER` - mysql/postgresql/sqlite
- `CACHE_DRIVER` - redis/local
- `ID_GENERATOR_DRIVER` - redis/local
- `JWT_SECRET` - JWT signing secret
- `SHORTLINK_DOMAIN` - base domain for short links

See README.md for full configuration options.

## Testing

Tests use standard Go testing. Existing tests live at:
- `pkg/helper/signature_helper_test.go`
- `pkg/helper/jwt_helper_test.go`
- `test/cache_test.go`

Run tests with coverage:
```bash
go test -cover ./...
```

## Build & Deployment Notes

- `Dockerfile` is a two-stage build: builds `admin-webui` (pnpm) first, then compiles the Go binary. The admin UI must be present as a submodule for docker builds to succeed.
- `.goreleaser.yaml` handles multi-arch release builds; see `.build/Dockerfile.goreleaser`.
- `.cnb.yml` defines CI pipelines (CNB platform) for amd64/arm64/loong64 image builds.
- `config/install.lock` is created on first successful install; deleting it re-triggers the install flow and disables migrations until reinstalled.