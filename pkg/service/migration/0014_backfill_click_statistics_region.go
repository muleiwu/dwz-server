// 0014_backfill_click_statistics_region 把 click_statistics 与
// ab_test_click_statistics 两张表所有历史行的 country/province/city/isp
// 按当前 ip2region xdb 重新计算一遍。
//
// 为什么是 Go 迁移而不是 SQL：回填需要在 Go 侧逐 IP 调 IPRegion.Lookup，
// SQL 表达不出来。goose 的版本追踪表 goose_db_version 天然提供「一次性
// 执行」语义 —— 迁移成功记账后永远不会再触发，失败则下次启动重试。
//
// 为什么是 NoTx：逐 IP UPDATE 可能触达数十万条数据，单事务在 MySQL binlog
// 与 Postgres vacuum 上都不划算。迁移天然幂等（DISTINCT IP → 查 xdb →
// UPDATE），失败重跑会收敛到同一状态。
package migration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/domain_validate"
	ipRegionImpl "cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/ip_region/impl"
	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

// 需要回填 region 四列的表。字段结构一致，走同一段回填逻辑即可。
var backfillTargets = []string{
	"click_statistics",
	"ab_test_click_statistics",
}

func init() {
	// 用 Named 变体让 goose 从显式文件名解析版本号（0014），避免依赖编译时
	// runtime.Caller 拿到的绝对路径 —— 不同 CI / GOPATH 下路径会抖动。
	goose.AddNamedMigrationNoTxContext(
		"0014_backfill_click_statistics_region.go",
		upBackfillClickRegion,
		downBackfillClickRegion,
	)
}

// upBackfillClickRegion 扫两张表、按 DISTINCT IP 去重查 xdb、批量 UPDATE。
// 参数里的 *sql.DB 是 goose 递进来的原始连接，我们转而用 helper 暴露的
// *gorm.DB 获得跨方言 ? 占位符与既有日志 / IPRegion 通道。
func upBackfillClickRegion(_ context.Context, _ *sql.DB) error {
	h := helper.GetHelper()
	logger := h.GetLogger()
	db := h.GetDatabase()
	if db == nil {
		return errors.New("[migration 0014] 数据库连接不可用")
	}

	ipr := h.GetIPRegion()
	if _, isNoop := ipr.(ipRegionImpl.Noop); isNoop {
		return errors.New("[migration 0014] IPRegion 尚未就绪（Noop 实现），拒绝回填以免清空所有 region 字段；请确认 xdb 加载成功后重启")
	}

	for _, table := range backfillTargets {
		uniqueIPs, affected, err := backfillTable(db, ipr, table)
		if err != nil {
			return fmt.Errorf("[migration 0014] 回填 %s 失败: %w", table, err)
		}
		logger.Info(fmt.Sprintf(
			"[migration 0014] %s 回填完成: unique_ips=%d, rows_affected=%d",
			table, uniqueIPs, affected,
		))
	}
	return nil
}

// backfillTable 对单张表执行 DISTINCT IP → Lookup → UPDATE，返回
// (unique_ip 数, 累计 rows_affected)。
func backfillTable(db *gorm.DB, ipr ipRegionImpl.IPRegion, table string) (int, int64, error) {
	// 1. 拉去重后的 IP 集合。表可能巨大但 DISTINCT ip 的基数通常只有几千到
	//    几万，可以一次性吞内存。
	var ips []string
	if err := db.
		Table(table).
		Distinct("ip").
		Where("ip <> ?", "").
		Where("ip IS NOT NULL").
		Pluck("ip", &ips).Error; err != nil {
		return 0, 0, fmt.Errorf("select distinct ip: %w", err)
	}

	// 2. 逐 IP 查 xdb + UPDATE。失败直接中断当前 table，上层返回错误 ——
	//    goose 不会写 applied，下次启动重试。
	var totalAffected int64
	for _, ip := range ips {
		region := ipr.Lookup(ip)
		res := db.Exec(
			"UPDATE "+table+" SET country = ?, province = ?, city = ?, isp = ? WHERE ip = ?",
			domain_validate.TruncateString(region.Country, 100),
			domain_validate.TruncateString(region.Province, 100),
			domain_validate.TruncateString(region.City, 100),
			domain_validate.TruncateString(region.ISP, 100),
			ip,
		)
		if res.Error != nil {
			return len(ips), totalAffected, fmt.Errorf("update ip=%s: %w", ip, res.Error)
		}
		totalAffected += res.RowsAffected
	}
	return len(ips), totalAffected, nil
}

// downBackfillClickRegion 把两张表的 region 四列统统置空，满足 goose 回滚
// 契约。实际运维里几乎不会用到。
func downBackfillClickRegion(_ context.Context, _ *sql.DB) error {
	h := helper.GetHelper()
	db := h.GetDatabase()
	if db == nil {
		return errors.New("[migration 0014] 数据库连接不可用")
	}
	for _, table := range backfillTargets {
		if err := db.Exec(
			"UPDATE " + table + " SET country = '', province = '', city = '', isp = ''",
		).Error; err != nil {
			return fmt.Errorf("[migration 0014] 回滚 %s 失败: %w", table, err)
		}
	}
	return nil
}
