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

// updateBatchSize 控制 UPDATE ... WHERE id IN (...) 的单批 ID 数量。
// 太小 => 单 IP 下 round-trip 变多；太大 => 单条 SQL 过长、可能触发
// MySQL max_allowed_packet / PG max_statement_len。500 是常见折中。
const updateBatchSize = 500

// backfillTable 对单张表执行扫全表 → 内存分桶 → 按主键 IN(...) 批量
// UPDATE，返回 (unique_ip 数, 累计 rows_affected)。
//
// 之所以不走 `UPDATE ... WHERE ip = ?` 的朴素写法：`ip` 列没有索引，每条
// UPDATE 都是全表扫描，在 6w unique IP × 百万行规模下会跑几个小时。走主
// 键批量：一次顺序全表扫拉出 (id, ip)（单次 IO）、在内存里按 ip 分桶，
// 最后 UPDATE 走 PK 索引 seek，总耗时能压到分钟级。
func backfillTable(db *gorm.DB, ipr ipRegionImpl.IPRegion, table string) (int, int64, error) {
	type row struct {
		ID uint64 `gorm:"column:id"`
		IP string `gorm:"column:ip"`
	}
	// 1. 一次性拉 (id, ip)。表大时这里是最大的内存开销：100 万行大约占
	//    ~30MB（8 字节 id + 平均 15 字节 ip string header + 内容），对现代
	//    服务器是可接受的常数级开销，比发数十万条 SQL 轻得多。
	var rows []row
	if err := db.
		Table(table).
		Select("id, ip").
		Where("ip <> ?", "").
		Where("ip IS NOT NULL").
		Find(&rows).Error; err != nil {
		return 0, 0, fmt.Errorf("select id,ip: %w", err)
	}

	// 2. 按 IP 在内存分桶。
	idsByIP := make(map[string][]uint64, len(rows)/2+1)
	for _, r := range rows {
		idsByIP[r.IP] = append(idsByIP[r.IP], r.ID)
	}

	// 3. 逐 IP 查 xdb + 按 PK IN (batch) 批量 UPDATE。单批 500 既避免 SQL
	//    过长，又足够让 PK 索引 seek 的批内开销摊薄掉。
	var totalAffected int64
	for ip, ids := range idsByIP {
		region := ipr.Lookup(ip)
		country := domain_validate.TruncateString(region.Country, 100)
		province := domain_validate.TruncateString(region.Province, 100)
		city := domain_validate.TruncateString(region.City, 100)
		isp := domain_validate.TruncateString(region.ISP, 100)

		for start := 0; start < len(ids); start += updateBatchSize {
			end := min(start+updateBatchSize, len(ids))
			chunk := ids[start:end]
			res := db.Exec(
				"UPDATE "+table+" SET country = ?, province = ?, city = ?, isp = ? WHERE id IN ?",
				country, province, city, isp, chunk,
			)
			if res.Error != nil {
				return len(idsByIP), totalAffected, fmt.Errorf("update ip=%s chunk=%d: %w", ip, start/updateBatchSize, res.Error)
			}
			totalAffected += res.RowsAffected
		}
	}
	return len(idsByIP), totalAffected, nil
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
