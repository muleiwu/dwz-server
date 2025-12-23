package migration

import (
	"fmt"

	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

type Migration struct {
	Helper    interfaces.HelperInterface
	Migration []any
}

func (receiver *Migration) Run() error {

	if receiver.Helper.GetDatabase() == nil {
		return fmt.Errorf("[db migration err: database is nil]")
	}

	autoInstall := receiver.Helper.GetEnv().GetString("AUTO_INSTALL", "")

	if !receiver.Helper.GetInstalled().IsInstalled() && autoInstall != "install" {
		return nil
	}

	if !receiver.Helper.GetInstalled().IsInstalled() && autoInstall == "install" {
		//installService := service.NewInitInstallService(receiver.Helper)
		//installService.AutoInstall(receiver.Migration)
		return nil
	}

	if len(receiver.Migration) > 0 {
		err := receiver.Helper.GetDatabase().AutoMigrate(receiver.Migration...)
		if err != nil {
			return fmt.Errorf("[db migration err:%s]", err.Error())
		}

		receiver.Helper.GetLogger().Info(fmt.Sprintf("[db migration success: %d models migrated]", len(receiver.Migration)))

		// 修复空字符串的 token 和 app_id 为 NULL（解决唯一索引冲突问题）
		receiver.fixEmptyTokenFields()
	}
	return nil
}

// fixEmptyTokenFields 将空字符串的 token 和 app_id 字段更新为 NULL
// 这是为了解决签名认证类型的 Token 不需要 token 字段，但唯一索引会对空字符串生效的问题
func (receiver *Migration) fixEmptyTokenFields() {
	db := receiver.Helper.GetDatabase()

	// 将空字符串的 token 更新为 NULL
	result := db.Exec("UPDATE user_tokens SET token = NULL WHERE token = ''")
	if result.Error != nil {
		receiver.Helper.GetLogger().Warn(fmt.Sprintf("[migration] 修复空 token 字段失败: %s", result.Error.Error()))
	} else if result.RowsAffected > 0 {
		receiver.Helper.GetLogger().Info(fmt.Sprintf("[migration] 已将 %d 条空 token 记录更新为 NULL", result.RowsAffected))
	}

	// 将空字符串的 app_id 更新为 NULL
	result = db.Exec("UPDATE user_tokens SET app_id = NULL WHERE app_id = ''")
	if result.Error != nil {
		receiver.Helper.GetLogger().Warn(fmt.Sprintf("[migration] 修复空 app_id 字段失败: %s", result.Error.Error()))
	} else if result.RowsAffected > 0 {
		receiver.Helper.GetLogger().Info(fmt.Sprintf("[migration] 已将 %d 条空 app_id 记录更新为 NULL", result.RowsAffected))
	}
}
