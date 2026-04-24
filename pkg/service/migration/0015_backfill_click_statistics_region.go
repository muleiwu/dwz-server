// 0015_backfill_click_statistics_region 把 click_statistics 与
// ab_test_click_statistics 两张表所有历史行的 country/province/city/isp
// 按当前 ip2region xdb 重新计算一遍。
//
// 设计要点：
//  1. **goose 跟踪，异步执行**：up() 走一次性校验后立刻 `go runBackfill()`
//     返回 nil —— goose 把 0015 记账入 goose_db_version，跨 k8s pod 的共享
//     DB 天然保证「整个集群只有一个 pod 真正跑回填」。相比文件标记方案，不
//     依赖单机 FS，适合多副本部署。
//  2. **不阻塞 HTTP 启动**：大表回填可能分钟级甚至更久，留给后台协程做。
//  3. **依赖 0014 的 ip 索引**：先扫全表拿 (id, ip)，内存里按 IP 分桶，
//     按 PK IN (…) 批量 UPDATE；PK / ip 索引都 seek，不再全表扫。
//  4. **IPRegion 未就绪 → up 直接报错**：此时不让 goose 记账，下次启动重试；
//     否则一旦标记 applied，xdb 再就绪也不会触发回填。
//  5. **up 返回后协程才启动**：所以即使协程中途失败，goose 已经记账。这是
//     一次性任务的权衡 —— 失败恢复需要运维手工
//     `DELETE FROM goose_db_version WHERE version_id=15;` 触发重跑。协程本
//     身天然幂等（DISTINCT ip → Lookup → UPDATE），重跑结果收敛。
package migration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/domain_validate"
	ipRegionImpl "cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/ip_region/impl"
	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

const backfillUpdateBatchSize = 500

// 需要回填 region 四列的表。字段结构一致，走同一段逻辑。
var backfillTargets = []string{
	"click_statistics",
	"ab_test_click_statistics",
}

func init() {
	goose.AddNamedMigrationNoTxContext(
		"0015_backfill_click_statistics_region.go",
		upBackfillClickRegion,
		downBackfillClickRegion,
	)
}

// upBackfillClickRegion 做前置校验后把真正的回填放到后台协程。
// 返回 nil 表示「迁移已被受理」—— goose 会立即把 0015 标记 applied，不等
// 后台协程完成。
func upBackfillClickRegion(_ context.Context, _ *sql.DB) error {
	h := helper.GetHelper()
	logger := h.GetLogger()

	if h.GetDatabase() == nil {
		return errors.New("[migration 0015] 数据库不可用")
	}
	if _, isNoop := h.GetIPRegion().(ipRegionImpl.Noop); isNoop {
		// 不记账 → 下次启动再试。
		return errors.New("[migration 0015] IPRegion 尚未就绪（Noop），请确认 xdb 加载成功后重启以触发回填")
	}

	logger.Info("[migration 0015] 已受理，后台协程执行回填（不阻塞启动）")
	go runBackfillAsync()
	return nil
}

// runBackfillAsync 是后台协程的入口。即使崩了也不会影响主流程；失败信息走
// error 日志，由运维检查 + 必要时手工重置 goose_db_version 触发重跑。
func runBackfillAsync() {
	h := helper.GetHelper()
	logger := h.GetLogger()
	db := h.GetDatabase()
	ipr := h.GetIPRegion()

	// 二次防御：协程启动的瞬间 container 理论上仍然稳定，但 container reload
	// (SIGHUP) 期间有极小窗口会让 DB/IPRegion 落到 nil / Noop。拿到空值就退。
	if db == nil {
		logger.Error("[migration 0015] 协程启动时 DB 不可用，放弃")
		return
	}
	if _, isNoop := ipr.(ipRegionImpl.Noop); isNoop {
		logger.Error("[migration 0015] 协程启动时 IPRegion 是 Noop，放弃")
		return
	}

	logger.Info("[migration 0015] 后台回填开始")
	startedAt := time.Now()
	for _, table := range backfillTargets {
		uniqueIPs, affected, err := backfillIPRegionTable(db, ipr, table)
		if err != nil {
			logger.Error(fmt.Sprintf("[migration 0015] %s 失败: %s", table, err.Error()))
			return
		}
		logger.Info(fmt.Sprintf(
			"[migration 0015] %s 完成: unique_ips=%d, rows_affected=%d",
			table, uniqueIPs, affected,
		))
	}
	logger.Info(fmt.Sprintf("[migration 0015] 全部完成，耗时 %s", time.Since(startedAt)))
}

// backfillIPRegionTable 扫一张表，按 IP 分桶，按 PK IN (batch) 批量 UPDATE。
// 返回 (unique_ip 数, 累计 rows_affected)。
func backfillIPRegionTable(db *gorm.DB, ipr ipRegionImpl.IPRegion, table string) (int, int64, error) {
	type row struct {
		ID uint64 `gorm:"column:id"`
		IP string `gorm:"column:ip"`
	}
	var rows []row
	if err := db.
		Table(table).
		Select("id, ip").
		Where("ip <> ?", "").
		Where("ip IS NOT NULL").
		Find(&rows).Error; err != nil {
		return 0, 0, fmt.Errorf("select id,ip: %w", err)
	}

	idsByIP := make(map[string][]uint64, len(rows)/2+1)
	for _, r := range rows {
		idsByIP[r.IP] = append(idsByIP[r.IP], r.ID)
	}

	var totalAffected int64
	for ip, ids := range idsByIP {
		region := ipr.Lookup(ip)
		country := domain_validate.TruncateString(region.Country, 100)
		province := domain_validate.TruncateString(region.Province, 100)
		city := domain_validate.TruncateString(region.City, 100)
		isp := domain_validate.TruncateString(region.ISP, 100)

		for start := 0; start < len(ids); start += backfillUpdateBatchSize {
			end := min(start+backfillUpdateBatchSize, len(ids))
			chunk := ids[start:end]
			res := db.Exec(
				"UPDATE "+table+" SET country = ?, province = ?, city = ?, isp = ? WHERE id IN ?",
				country, province, city, isp, chunk,
			)
			if res.Error != nil {
				return len(idsByIP), totalAffected, fmt.Errorf("update ip=%s chunk=%d: %w", ip, start/backfillUpdateBatchSize, res.Error)
			}
			totalAffected += res.RowsAffected
		}
	}
	return len(idsByIP), totalAffected, nil
}

// downBackfillClickRegion 把两张表的 region 四列置空，满足 goose 回滚契约。
// 同步执行即可 —— DOWN 是一条全表 UPDATE，不至于拖太久。
func downBackfillClickRegion(_ context.Context, _ *sql.DB) error {
	h := helper.GetHelper()
	db := h.GetDatabase()
	if db == nil {
		return errors.New("[migration 0015] 数据库不可用")
	}
	for _, table := range backfillTargets {
		if err := db.Exec(
			"UPDATE " + table + " SET country = '', province = '', city = '', isp = ''",
		).Error; err != nil {
			return fmt.Errorf("[migration 0015] 回滚 %s 失败: %w", table, err)
		}
	}
	return nil
}
