package migration

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"cnb.cool/mliev/dwz/dwz-server/pkg/helper"
	"cnb.cool/mliev/dwz/dwz-server/pkg/service/install_bootstrap"
	"github.com/pressly/goose/v3"
)

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
