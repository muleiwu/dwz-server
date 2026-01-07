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

# Run specific test
go test ./internal/helper/signature_helper_test.go

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
- `main.go` - Embeds templates and static files, calls `cmd.Start()`
- `cmd/run.go` - Initializes services via Assembly pattern, starts servers

### Dependency Injection (Assembly Pattern)

Services are initialized in strict dependency order via `config.Assembly` (config/assembly.go:20-32):
1. Env (environment variables)
2. Config (code defaults + env vars)
3. Logger (Zap)
4. Version
5. Installed (installation check)
6. Database (MySQL/PostgreSQL/SQLite)
7. Redis
8. Cache (Redis/memory)

Servers are started via `config.Server` (config/server.go:15-28):
1. Migration (database migrations)
2. IDGenerator (distributed ID generation)
3. HttpServer (Gin HTTP server)

### Directory Structure

- `app/controller/` - HTTP request handlers (REST API endpoints)
- `app/service/` - Business logic layer
- `app/dao/` - Data access objects (GORM operations)
- `app/model/` - GORM data models
- `app/dto/` - Data transfer objects
- `app/middleware/` - Gin middleware (auth, CORS, operation logging, install check)
- `app/constants/` - Application constants
- `config/autoload/` - Router configuration and dependency providers
- `config/assembly.go` - Service assembly initialization
- `config/server.go` - Server initialization
- `internal/helper/` - Helper utilities and DI container
- `internal/interfaces/` - Interface definitions for dependency inversion
- `internal/pkg/` - Internal packages (cache, database, http_server, etc.)
- `templates/` - Embedded HTML templates (error pages, installation)
- `static/` - Embedded admin web UI build files

### Key Interfaces

The codebase uses interface-based dependency injection defined in `internal/interfaces/`:
- `AssemblyInterface` - For service initialization
- `ServerInterface` - For server startup
- `HelperInterface` - Global DI container providing access to logger, config, cache, database, etc.

### Routing

Routes are defined in `config/autoload/router.go` using Gin framework:
- `/health`, `/health/simple` - Health checks (no auth)
- `/api/v1/install/*` - Installation endpoints (no auth)
- `/api/v1/auth/login` - Login (no auth, operation logged)
- `/api/v1/*` - Protected API endpoints (require auth via `AuthMiddleware`)
- Short code redirect: `GET /:code` (regex route for [a-zA-Z0-9\-_.]+)

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

Tests use standard Go testing. Example test file: `internal/helper/signature_helper_test.go`

Run tests with coverage:
```bash
go test -cover ./...
```