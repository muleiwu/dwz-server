# 配置说明

首次安装完成后，主配置文件位于 `config/config.yaml`（安装向导生成）。仓库根目录的 [config.yaml.example](../config.yaml.example) 列出了全部可配置项；**同名环境变量会覆盖 YAML 配置**（规则：YAML 键中的 `.` 替换为 `_` 并转大写，如 `database.dbname` → `DATABASE_DBNAME`）。

## 常用配置

| YAML 键 | 环境变量 | 说明 |
| --- | --- | --- |
| `http.addr` | `HTTP_ADDR` | HTTP 监听地址，默认 `:8080` |
| `http.mode` | `HTTP_MODE` | `debug`、`release` 或 `test` |
| `database.driver` | `DATABASE_DRIVER` | `sqlite`、`mysql` 或 `postgresql` |
| `database.filepath` | `DATABASE_FILEPATH` | SQLite 文件路径（默认 `./config/sqlite.db`） |
| `database.host` | `DATABASE_HOST` | MySQL/PostgreSQL 主机 |
| `database.port` | `DATABASE_PORT` | 数据库端口 |
| `database.dbname` | `DATABASE_DBNAME` | 数据库名 |
| `database.username` | `DATABASE_USERNAME` | 数据库用户名 |
| `database.password` | `DATABASE_PASSWORD` | 数据库密码 |
| `redis.host` | `REDIS_HOST` | Redis 主机（默认 `localhost:6379` 对应的 host） |
| `redis.port` | `REDIS_PORT` | Redis 端口 |
| `redis.password` | `REDIS_PASSWORD` | Redis 密码 |
| `redis.db` | `REDIS_DB` | Redis DB 编号 |
| `cache.driver` | `CACHE_DRIVER` | `memory`（默认）、`redis` 或 `none` |
| `id_generator.driver` | `ID_GENERATOR_DRIVER` | `local`（默认）或 `redis` |
| `jwt.secret` | `JWT_SECRET` | JWT 签名密钥，生产环境必须妥善保管（首启自动生成） |
| `jwt.expire_hours` | `JWT_EXPIRE_HOURS` | 登录 Token 有效时长 |

## 驱动选择与部署模式

| 场景 | 数据库 | 缓存 | ID 发号器 |
| --- | --- | --- | --- |
| 轻量单机 | SQLite | memory | local |
| 标准生产 | MySQL / PostgreSQL | redis | redis |

- 多实例部署必须让各实例共用数据库，并使用 `redis` 发号器，避免本地计数器在实例间冲突。
- 配置了 `redis` 缓存但 Redis 不可用时，服务会回退到 `memory` 缓存。

## Docker 注意事项

容器内填写数据库/Redis 主机时，应使用 Compose 服务名而不是 `localhost`。例如数据库服务名为 `mysql`，则 `database.host` 填 `mysql`。

## 相关文档

- [config.yaml.example](../config.yaml.example) — 全量配置项示例
- [静态资源与模板说明](./static-templates.md) — `embed` / `disk` 模式与开发热更新
