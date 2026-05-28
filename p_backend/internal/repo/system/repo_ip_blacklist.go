package system

import (
	"context"
	"strings"
	"time"

	"monorepo/internal/model"
	"monorepo/pkg/db"
	"monorepo/pkg/xerr"
	commpb "monorepo/proto/xadminpb/commpb"

	"gorm.io/gorm"
)

func (r *Repo) ListIPBlacklist(ctx context.Context, page *commpb.PageArgs, sortArgs []*commpb.SortArgs, filters IPBlacklistFilters) ([]IPBlacklistRow, int64, error) {
	rows := make([]IPBlacklistRow, 0, page.GetPs())
	query := r.db.WithContext(ctx).
		Table("system_ip_blacklist b").
		Joins("LEFT JOIN admin_user u ON u.uid = CASE WHEN b.creator ~ '^[0-9]+$' THEN CAST(b.creator AS BIGINT) ELSE NULL END AND u.deleted_at = 0").
		Select(`
b.id,
b.ip,
b.ban_type,
TO_CHAR(b.start_at, 'YYYY-MM-DD HH24:MI:SS') AS start_at,
TO_CHAR(b.end_at, 'YYYY-MM-DD HH24:MI:SS') AS end_at,
b.reason,
COALESCE(NULLIF(TRIM(u.username), ''), b.creator) AS creator,
b.status,
b.hit_count,
b.last_action,
TO_CHAR(b.updated_at, 'YYYY-MM-DD HH24:MI:SS') AS updated_at
`).
		Where("b.deleted_at = 0")
	query = applyIPBlacklistFilters(query, filters)
	if len(sortArgs) == 0 {
		query = query.Order("b.created_at desc")
	}
	total, err := db.Paginate(
		query,
		page,
		sortArgs,
		[]string{"b.id", "b.ip", "b.ban_type", "b.status", "b.hit_count", "b.updated_at"},
		&rows,
		db.PaginateArgs{AppendCreatedAtDesc: false},
	)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repo) ListActiveIPBlacklistEntries(ctx context.Context) ([]IPBlacklistRow, error) {
	rows := make([]IPBlacklistRow, 0)
	err := r.db.WithContext(ctx).
		Model(&model.SystemIPBlacklist{}).
		Select("id, ip, TO_CHAR(start_at, 'YYYY-MM-DD HH24:MI:SS') AS start_at, TO_CHAR(end_at, 'YYYY-MM-DD HH24:MI:SS') AS end_at").
		Where("status = ? AND deleted_at = 0", "active").
		Find(&rows).Error
	if err != nil {
		return nil, xerr.WrapDBE(err, "load active ip blacklist")
	}
	return rows, nil
}

func (r *Repo) ListActiveIPBlacklistEntriesByIPs(ctx context.Context, ips []string) ([]IPBlacklistRow, error) {
	normalizedIPs := make([]string, 0, len(ips))
	for _, ip := range ips {
		trimmed := strings.TrimSpace(ip)
		if trimmed != "" {
			normalizedIPs = append(normalizedIPs, trimmed)
		}
	}
	if len(normalizedIPs) == 0 {
		return []IPBlacklistRow{}, nil
	}
	rows := make([]IPBlacklistRow, 0, len(normalizedIPs))
	err := r.db.WithContext(ctx).
		Model(&model.SystemIPBlacklist{}).
		Select("id, ip, status, TO_CHAR(start_at, 'YYYY-MM-DD HH24:MI:SS') AS start_at, TO_CHAR(end_at, 'YYYY-MM-DD HH24:MI:SS') AS end_at").
		Where("ip IN ? AND status = ? AND deleted_at = 0", normalizedIPs, "active").
		Find(&rows).Error
	if err != nil {
		return nil, xerr.WrapDBE(err, "load active ip blacklist by ips")
	}
	return rows, nil
}

func (r *Repo) GetIPBlacklistEntry(ctx context.Context, id int64) (*IPBlacklistRow, error) {
	var row IPBlacklistRow
	err := r.db.WithContext(ctx).
		Model(&model.SystemIPBlacklist{}).
		Select("id, ip, status, TO_CHAR(start_at, 'YYYY-MM-DD HH24:MI:SS') AS start_at, TO_CHAR(end_at, 'YYYY-MM-DD HH24:MI:SS') AS end_at").
		Where("id = ? AND deleted_at = 0", id).
		First(&row).Error
	if err != nil {
		return nil, xerr.WrapDBNotFound(err, "IP 黑名单记录不存在")
	}
	return &row, nil
}

func (r *Repo) CreateIPBlacklist(ctx context.Context, row *IPBlacklistRow) error {
	if row == nil {
		return xerr.NewWithDetail(xerr.CodeParamError, "ip blacklist row is empty")
	}
	startAt := parseIPBlacklistDateTime(row.StartAt, time.Now())
	endAt := parseIPBlacklistDateTime(row.EndAt, time.Now().Add(24*time.Hour))
	data := &model.SystemIPBlacklist{
		IP:         strings.TrimSpace(row.IP),
		BanType:    strings.TrimSpace(row.BanType),
		StartAt:    startAt,
		EndAt:      endAt,
		Reason:     strings.TrimSpace(row.Reason),
		Creator:    strings.TrimSpace(row.Creator),
		Status:     strings.TrimSpace(row.Status),
		HitCount:   row.HitCount,
		LastAction: strings.TrimSpace(row.LastAction),
		DeletedAt:  0,
	}
	if err := xerr.WrapDBDuplicate(r.db.WithContext(ctx).Create(data).Error, "IP 黑名单记录已存在"); err != nil {
		return err
	}
	row.ID = data.ID
	row.StartAt = data.StartAt.Format("2006-01-02 15:04:05")
	row.EndAt = data.EndAt.Format("2006-01-02 15:04:05")
	return nil
}

func (r *Repo) UpdateIPBlacklist(ctx context.Context, id int64, banType, endAt, reason string) error {
	updates := map[string]any{
		"last_action": "manual_update",
	}
	if strings.TrimSpace(banType) != "" {
		updates["ban_type"] = strings.TrimSpace(banType)
	}
	if strings.TrimSpace(endAt) != "" {
		updates["end_at"] = parseIPBlacklistDateTime(endAt, time.Now().Add(24*time.Hour))
	}
	if strings.TrimSpace(reason) != "" {
		updates["reason"] = strings.TrimSpace(reason)
	}
	ret := r.db.WithContext(ctx).
		Model(&model.SystemIPBlacklist{}).
		Where("id = ? AND deleted_at = 0", id).
		Updates(updates)
	return xerr.WrapDBUpdateMiss(ret, "IP 黑名单记录不存在")
}

func (r *Repo) UnblockIPBlacklist(ctx context.Context, id int64) error {
	ret := r.db.WithContext(ctx).
		Model(&model.SystemIPBlacklist{}).
		Where("id = ? AND deleted_at = 0", id).
		Updates(map[string]any{
			"status":      "inactive",
			"last_action": "manual_unblock",
		})
	return xerr.WrapDBUpdateMiss(ret, "IP 黑名单记录不存在")
}

func (r *Repo) BatchUnblockIPBlacklist(ctx context.Context, ids []int64) error {
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if err := r.UnblockIPBlacklist(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) IncrementHitCounts(ctx context.Context, deltas map[int64]int32) error {
	for id, delta := range deltas {
		if id <= 0 || delta <= 0 {
			continue
		}
		if err := r.db.WithContext(ctx).
			Model(&model.SystemIPBlacklist{}).
			Where("id = ? AND deleted_at = 0", id).
			Update("hit_count", gorm.Expr("hit_count + ?", delta)).Error; err != nil {
			return xerr.WrapDBE(err, "increment ip blacklist hit count")
		}
	}
	return nil
}

func (r *Repo) DeleteIPBlacklist(ctx context.Context, id int64) error {
	ret := r.db.WithContext(ctx).
		Model(&model.SystemIPBlacklist{}).
		Where("id = ? AND deleted_at = 0", id).
		Updates(map[string]any{
			"deleted_at":  time.Now().UnixMilli(),
			"last_action": "manual_delete",
		})
	return xerr.WrapDBUpdateMiss(ret, "IP 黑名单记录不存在")
}

func (r *Repo) ImportIPBlacklist(ctx context.Context, ips []string, banType string, durationHours int32, customEndAtRaw string, creator string) error {
	normalizedBanType := strings.TrimSpace(banType)
	if normalizedBanType == "" || normalizedBanType == "unspecified" {
		normalizedBanType = "temp"
	}
	if durationHours <= 0 {
		durationHours = 24
	}
	now := time.Now()
	endAtValue := now.Add(time.Duration(durationHours) * time.Hour)
	customEndAt := parseIPBlacklistDateTime(strings.TrimSpace(customEndAtRaw), time.Time{})
	if !customEndAt.IsZero() {
		endAtValue = customEndAt
	}
	if normalizedBanType == "permanent" {
		endAtValue = parseIPBlacklistDateTime("2099-12-31 23:59:59", now.Add(24*time.Hour))
	}
	rows := make([]*model.SystemIPBlacklist, 0, len(ips))
	for _, ip := range ips {
		trimmed := strings.TrimSpace(ip)
		if trimmed == "" {
			continue
		}
		rows = append(rows, &model.SystemIPBlacklist{
			IP:         trimmed,
			BanType:    normalizedBanType,
			StartAt:    now,
			EndAt:      endAtValue,
			Reason:     "导入拉黑",
			Creator:    strings.TrimSpace(creator),
			Status:     "active",
			HitCount:   0,
			LastAction: "import",
			DeletedAt:  0,
		})
	}
	if len(rows) == 0 {
		return nil
	}
	return xerr.WrapDBE(
		r.db.WithContext(ctx).Create(rows).Error,
		"import ip blacklist",
	)
}

func (r *Repo) ListIPBlacklistCreators(ctx context.Context) ([]string, error) {
	var creators []string
	err := r.db.WithContext(ctx).
		Table("system_ip_blacklist b").
		Joins("LEFT JOIN admin_user u ON u.uid = CASE WHEN b.creator ~ '^[0-9]+$' THEN CAST(b.creator AS BIGINT) ELSE NULL END AND u.deleted_at = 0").
		Where("b.deleted_at = 0").
		Distinct("COALESCE(NULLIF(TRIM(u.username), ''), b.creator)").
		Pluck("COALESCE(NULLIF(TRIM(u.username), ''), b.creator)", &creators).Error
	if err != nil {
		return nil, xerr.WrapDBE(err, "list ip blacklist creators")
	}
	filtered := creators[:0]
	for _, c := range creators {
		if strings.TrimSpace(c) != "" {
			filtered = append(filtered, strings.TrimSpace(c))
		}
	}
	return filtered, nil
}

func applyIPBlacklistFilters(query *gorm.DB, filters IPBlacklistFilters) *gorm.DB {
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("(b.ip LIKE ? OR b.reason LIKE ?)", like, like)
	}
	if status := strings.TrimSpace(filters.Status); status != "" {
		now := time.Now()
		if status == "active" {
			query = query.Where("b.status = ? AND b.end_at >= ?", status, now)
		} else {
			query = query.Where("(b.status <> ? OR b.end_at < ?)", "active", now)
		}
	}
	if banType := strings.TrimSpace(filters.BanType); banType != "" {
		query = query.Where("b.ban_type = ?", banType)
	}
	if creator := strings.TrimSpace(filters.Creator); creator != "" {
		query = query.Where("(b.creator = ? OR u.username = ? OR u.display_name = ?)", creator, creator, creator)
	}
	return query
}

func parseIPBlacklistDateTime(raw string, fallback time.Time) time.Time {
	input := strings.TrimSpace(raw)
	if input == "" {
		return fallback
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if value, err := time.ParseInLocation(layout, input, time.Local); err == nil {
			return value
		}
	}
	return fallback
}
