// backfill_ip_region 在后台协程里把 click_statistics 与
// ab_test_click_statistics 两张表所有历史行的 country/province/city/isp
// 按当前 ip2region xdb 重新计算一遍。
//
// 设计要点：
//  1. 放后台：本任务对 HTTP 启动不是阻塞前置，移到 goroutine 里让服务立刻
//     可对外，避免大表回填期间 /health 都报 down。
//  2. 幂等标记文件 ./config/ip_region_backfill.done：存在即跳过。遵循仓库
//     已有的 install.lock / install_admin.json 文件标记惯例；单机部署够用。
//     多实例并发触发时会重复跑但结果收敛（每条 UPDATE 自包含且可重入）。
//  3. 主键批量 UPDATE：依赖 0014 迁移新加的 idx_*_ip，先扫全表拿 (id, ip)
//     再在内存按 IP 分桶，按 WHERE id IN (…) 批量更新，走 PK / ip 索引 seek。
//  4. Noop / xdb 未就绪时直接放弃本轮，不写标记，下次启动重试。
package migration

import (
	"errors"
	"fmt"
	"os"
	"time"

	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/helper"
	"cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/domain_validate"
	ipRegionImpl "cnb.cool/mliev/dwz/dwz-server/v2/pkg/service/ip_region/impl"
	"gorm.io/gorm"
)

const (
	backfillMarkerPath      = "./config/ip_region_backfill.done"
	backfillUpdateBatchSize = 500
)

// 需要回填 region 四列的表。字段结构一致，走同一段逻辑。
var backfillTargets = []string{
	"click_statistics",
	"ab_test_click_statistics",
}

// StartBackfillIPRegion 启动一个后台协程做历史归属地回填，立即返回。
// 调用方一般是 Migration.Run() 尾部 —— 那时 goose 迁移（含 0014 的 ip 索引）
// 已全部就绪，IPRegion assembly 也早已初始化。
func StartBackfillIPRegion() {
	go runBackfillIPRegion()
}

func runBackfillIPRegion() {
	h := helper.GetHelper()
	logger := h.GetLogger()

	if _, err := os.Stat(backfillMarkerPath); err == nil {
		logger.Info("[ip_region_backfill] 已存在 " + backfillMarkerPath + "，跳过回填")
		return
	} else if !errors.Is(err, os.ErrNotExist) {
		logger.Warn("[ip_region_backfill] 读取标记文件失败，仍尝试回填: " + err.Error())
	}

	db := h.GetDatabase()
	if db == nil {
		logger.Warn("[ip_region_backfill] 数据库不可用，放弃")
		return
	}
	ipr := h.GetIPRegion()
	if _, isNoop := ipr.(ipRegionImpl.Noop); isNoop {
		logger.Warn("[ip_region_backfill] IPRegion 尚未就绪（Noop），放弃，下次启动再试")
		return
	}

	logger.Info("[ip_region_backfill] 后台回填开始")
	startedAt := time.Now()
	for _, table := range backfillTargets {
		uniqueIPs, affected, err := backfillIPRegionTable(db, ipr, table)
		if err != nil {
			logger.Error(fmt.Sprintf("[ip_region_backfill] %s 失败: %s", table, err.Error()))
			return
		}
		logger.Info(fmt.Sprintf(
			"[ip_region_backfill] %s 完成: unique_ips=%d, rows_affected=%d",
			table, uniqueIPs, affected,
		))
	}

	if err := os.WriteFile(backfillMarkerPath, []byte(time.Now().Format(time.RFC3339)), 0o644); err != nil {
		logger.Warn("[ip_region_backfill] 写入标记文件失败，下次启动仍会重跑: " + err.Error())
	}
	logger.Info(fmt.Sprintf("[ip_region_backfill] 全部完成，耗时 %s", time.Since(startedAt)))
}

// backfillIPRegionTable 扫一张表，按 IP 分桶，按 PK 批量 UPDATE。
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
