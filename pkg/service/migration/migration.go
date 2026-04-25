package migration

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"time"

	// 触发 migrations 包加载 —— 包内的 Go 迁移（如 0015）需要在
	// goose.Up 之前完成 goose.AddNamedMigration* 注册。
	_ "cnb.cool/mliev/dwz/dwz-server/v2/migrations"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/install_bootstrap"
	"github.com/pressly/goose/v3"
)

// baselineLegacyVersion 是迁移系统切换到 goose 之前已经存在的最大迁移版本号。
// 早期版本通过 GORM AutoMigrate 创建业务表,这些库里既有 users / short_links 等
// 业务表,也没有 goose_db_version 记录。当我们引入 0010 起的新增 SQL 后,
// goose 会从 0001 开始重跑并撞库("table already exists")。
//
// baselineLegacyVersions() 在启动期检测这种"遗留 schema"状态,把 1..本常量
// 标记为 applied,使 goose 只去运行真正没跑过的迁移。
//
// 后续如果要把"自适应窗口"扩到更高版本(例如 12),需要确保 1..12 的迁移在
// 当时确实可被 AutoMigrate / 历史部署完整建出。除非有这种历史包袱,新增的
// 迁移版本不应抬高这个数值——让 goose 正常运行才是默认路径。
const baselineLegacyVersion int64 = 9

// legacySentinelTable 是用于判断"这是一个遗留 schema 库"的探针表。
// 选 users 是因为它在 0007 引入并被所有后续业务依赖,几乎不可能被删除。
const legacySentinelTable = "users"

// dialectMap normalises driver names to goose dialect names. The "memory"
// driver is in-memory SQLite, which still uses the sqlite3 dialect but its
// migration tracking table is wiped on restart — that is fine for ephemeral
// dev/test environments.
var dialectMap = map[string]string{
	"mysql":      "mysql",
	"postgresql": "postgres",
	"sqlite":     "sqlite3",
	"memory":     "sqlite3",
}

// dirByDriver chooses the migration files subdirectory. SQLite-style files
// also work for the in-memory driver.
var dirByDriver = map[string]string{
	"mysql":      "migrations/mysql",
	"postgresql": "migrations/postgresql",
	"sqlite":     "migrations/sqlite",
	"memory":     "migrations/sqlite",
}

// Migration is the dwz Server that runs goose-based schema migrations.
// It reads a base embed.FS plus any EE-supplied embed.FS and applies them
// in dialect-specific subdirectories, gated by the install state.
type Migration struct {
	BaseFS embed.FS
}

func (m *Migration) Run() error {
	h := helper.GetHelper()
	logger := h.GetLogger()

	if installed := h.GetInstalled(); installed == nil || !installed.IsInstalled() {
		logger.Info("[migration] 应用未安装，跳过迁移")
		return nil
	}

	db := h.GetDatabase()
	if db == nil {
		return fmt.Errorf("[migration] 数据库连接获取失败")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("[migration] 获取 sql.DB 失败: %w", err)
	}

	driver := h.GetConfig().GetString("database.driver", "mysql")
	dialect, ok := dialectMap[driver]
	if !ok {
		return fmt.Errorf("[migration] 不支持的数据库驱动: %s", driver)
	}
	dir, ok := dirByDriver[driver]
	if !ok {
		return fmt.Errorf("[migration] 缺少 %s 驱动的迁移目录映射", driver)
	}

	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("[migration] 设置方言失败: %w", err)
	}
	goose.SetBaseFS(m.BaseFS)
	goose.SetLogger(gooseLogger{})

	if err := m.baselineLegacyVersions(); err != nil {
		// baseline 失败不直接挂掉:如果是新装库,后续 goose.Up 仍能正确执行;
		// 如果是遗留库,goose.Up 会再报一次 "table already exists",到时再排查。
		logger.Warn("[migration] baseline 失败: " + err.Error())
	}

	if err := goose.Up(sqlDB, dir); err != nil {
		return fmt.Errorf("[migration] CE 迁移失败: %w", err)
	}
	logger.Info(fmt.Sprintf("[migration] CE 迁移完成 (driver=%s, dir=%s)", driver, dir))

	// EE additional migrations FS (also in dialect subdirectory).
	if extra := h.GetConfig().Get("ee.extra_migrations_fs", nil); extra != nil {
		if eeFS, ok := extra.(embed.FS); ok && hasEntries(eeFS, dir) {
			goose.SetBaseFS(eeFS)
			if err := goose.Up(sqlDB, dir); err != nil {
				return fmt.Errorf("[migration] EE 迁移失败: %w", err)
			}
			logger.Info("[migration] EE 迁移完成")
			// reset to CE FS for any later operations
			goose.SetBaseFS(m.BaseFS)
		}
	}

	m.fixEmptyTokenFields()

	// One-shot post-install bootstrap: when /api/v1/install ran and dropped
	// the admin payload at config/install_admin.json, create the user now
	// that the schema is in place.
	if err := install_bootstrap.Consume(); err != nil {
		logger.Warn("[migration] 创建初始管理员失败: " + err.Error())
	}
	return nil
}

func (m *Migration) Stop() error { return nil }

// fixEmptyTokenFields normalises empty-string token / app_id values to NULL
// so the unique indexes on user_tokens don't collide for signature-auth rows.
func (m *Migration) fixEmptyTokenFields() {
	h := helper.GetHelper()
	db := h.GetDatabase()
	logger := h.GetLogger()

	if r := db.Exec("UPDATE user_tokens SET token = NULL WHERE token = ''"); r.Error != nil {
		logger.Warn(fmt.Sprintf("[migration] 修复空 token 字段失败: %s", r.Error.Error()))
	} else if r.RowsAffected > 0 {
		logger.Info(fmt.Sprintf("[migration] 已将 %d 条空 token 记录更新为 NULL", r.RowsAffected))
	}
	if r := db.Exec("UPDATE user_tokens SET app_id = NULL WHERE app_id = ''"); r.Error != nil {
		logger.Warn(fmt.Sprintf("[migration] 修复空 app_id 字段失败: %s", r.Error.Error()))
	} else if r.RowsAffected > 0 {
		logger.Info(fmt.Sprintf("[migration] 已将 %d 条空 app_id 记录更新为 NULL", r.RowsAffected))
	}
}

// baselineLegacyVersions handles the AutoMigrate → goose transition. When the
// database has business tables but goose has no record of them, this writes
// version rows 1..baselineLegacyVersion as already-applied so goose only runs
// the genuinely new SQL files.
//
// Conditions for action (all must hold):
//  1. goose_db_version exists (created by EnsureDBVersion if needed) and the
//     current applied version is 0 — i.e., nothing beyond the bootstrap row.
//  2. The legacy sentinel table (users) exists, indicating the schema was
//     previously provisioned by some non-goose path.
//
// On a fresh install neither condition (2) holds, and goose runs the full
// migration set normally. On a healthy goose-managed install (1) is false and
// we leave it alone.
func (m *Migration) baselineLegacyVersions() error {
	h := helper.GetHelper()
	logger := h.GetLogger()
	db := h.GetDatabase()
	if db == nil {
		return fmt.Errorf("数据库连接不可用")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取 sql.DB 失败: %w", err)
	}

	// 1) 创建 goose_db_version(若不存在)并读取当前版本。
	currentVersion, err := goose.EnsureDBVersion(sqlDB)
	if err != nil {
		// ErrNoNextVersion 表示版本表里全部记录都标为 rollback。这是异常状态,
		// 不再插入 baseline 以免破坏现状,交由运维介入。
		if errors.Is(err, goose.ErrNoNextVersion) {
			logger.Warn("[migration] goose_db_version 状态异常,跳过 baseline")
			return nil
		}
		return fmt.Errorf("EnsureDBVersion: %w", err)
	}
	if currentVersion > 0 {
		// goose 已有记录,正常路径。
		return nil
	}

	// 2) 仅在确实存在遗留业务表时才写 baseline,避免误改 fresh install。
	if !db.Migrator().HasTable(legacySentinelTable) {
		return nil
	}

	// 3) 检查 1..baselineLegacyVersion 是否已经有任何一行存在(防止重复运行)。
	var existing int64
	if err := db.Table(goose.TableName()).
		Where("version_id BETWEEN ? AND ?", 1, baselineLegacyVersion).
		Count(&existing).Error; err != nil {
		return fmt.Errorf("检查现有 baseline 失败: %w", err)
	}
	if existing > 0 {
		// 已经部分写入过,谨慎起见不补写,提示运维确认。
		logger.Warn(fmt.Sprintf(
			"[migration] goose_db_version 中已存在 %d 条 1..%d 区间的记录,跳过 baseline",
			existing, baselineLegacyVersion,
		))
		return nil
	}

	// 4) 写入 baseline 行。
	now := time.Now()
	for v := int64(1); v <= baselineLegacyVersion; v++ {
		err := db.Exec(
			"INSERT INTO "+goose.TableName()+" (version_id, is_applied, tstamp) VALUES (?, ?, ?)",
			v, true, now,
		).Error
		if err != nil {
			return fmt.Errorf("写入 baseline version=%d 失败: %w", v, err)
		}
	}
	logger.Info(fmt.Sprintf(
		"[migration] 检测到遗留 schema (%s 表已存在,goose 无记录),已为 1..%d 写入 baseline",
		legacySentinelTable, baselineLegacyVersion,
	))
	return nil
}

// hasEntries returns true when fsys contains at least one regular file under
// dir. This avoids invoking goose against an empty subtree (which it would
// surface as an error).
func hasEntries(fsys embed.FS, dir string) bool {
	found := false
	_ = fs.WalkDir(fsys, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return fs.SkipAll
			}
			return err
		}
		if !d.IsDir() {
			found = true
			return fs.SkipAll
		}
		return nil
	})
	return found
}

// gooseLogger bridges goose's logger contract onto helper.GetLogger().
type gooseLogger struct{}

func (gooseLogger) Fatal(v ...any)               { helper.GetHelper().GetLogger().Fatal(fmt.Sprint(v...)) }
func (gooseLogger) Fatalf(f string, v ...any)    { helper.GetHelper().GetLogger().Fatal(fmt.Sprintf(f, v...)) }
func (gooseLogger) Print(v ...any)               { helper.GetHelper().GetLogger().Info(fmt.Sprint(v...)) }
func (gooseLogger) Println(v ...any)             { helper.GetHelper().GetLogger().Info(fmt.Sprint(v...)) }
func (gooseLogger) Printf(f string, v ...any)    { helper.GetHelper().GetLogger().Info(fmt.Sprintf(f, v...)) }
