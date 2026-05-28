package system

import (
	"context"
	"strings"
	"time"

	"monorepo/internal/model"
	"monorepo/pkg/db"
	"monorepo/pkg/xerr"
	commpb "monorepo/proto/xadminpb/commpb"
	"monorepo/util"

	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

type AuditFilters struct {
	Actor     string
	Action    string
	Result    string
	TraceID   string
	RequestID string
	SourceIP  string
	Keyword   string
	CreatedAt *commpb.TimeRange
}

type AuditRow struct {
	ID        int64     `gorm:"column:id"`
	UID       int32     `gorm:"column:uid"`
	Actor     string    `gorm:"column:actor"`
	Action    string    `gorm:"column:action"`
	Result    string    `gorm:"column:result"`
	TraceID   string    `gorm:"column:trace_id"`
	RequestID string    `gorm:"column:request_id"`
	SourceIP  string    `gorm:"column:source_ip"`
	Duration  string    `gorm:"column:duration"`
	UserAgent string    `gorm:"column:user_agent"`
	Detail    string    `gorm:"column:detail"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func NewRepo() *Repo {
	return &Repo{db: db.GetDatabase()}
}

func NewRepoWithDB(database *gorm.DB) *Repo {
	return &Repo{db: database}
}

func (r *Repo) ListAuditLogs(ctx context.Context, page *commpb.PageArgs, sort []*commpb.SortArgs, filters AuditFilters) ([]AuditRow, int64, error) {
	query := r.buildAuditQuery(ctx, filters)
	rows := make([]AuditRow, 0)
	total, err := db.Paginate(query, page, sort, []string{
		"a.id",
		"a.uid",
		"actor",
		"a.action",
		"a.result",
		"a.trace_id",
		"a.source_ip",
		"a.created_at",
	}, &rows, db.PaginateArgs{
		AppendCreatedAtDesc: true,
	})
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repo) GetAuditLogByID(ctx context.Context, id int64) (*AuditRow, error) {
	var row AuditRow
	err := r.baseAuditSelect(r.db.WithContext(ctx)).
		Where("a.id = ?", id).
		First(&row).Error
	if err != nil {
		return nil, xerr.WrapDBNotFound(err, "audit log not found")
	}
	return &row, nil
}

func (r *Repo) ExportAuditLogs(ctx context.Context, filters AuditFilters, sort []*commpb.SortArgs) ([]AuditRow, error) {
	query := r.buildAuditQuery(ctx, filters)
	if len(sort) == 0 {
		sort = []*commpb.SortArgs{{OrderField: "a.created_at", OrderType: commpb.OrderType_OT_Desc}}
	}
	rows := make([]AuditRow, 0)
	if _, err := db.Paginate(query, &commpb.PageArgs{Pn: 1, Ps: 100000, IsDownload: true}, sort, []string{
		"a.id",
		"a.uid",
		"actor",
		"a.action",
		"a.result",
		"a.trace_id",
		"a.source_ip",
		"a.created_at",
	}, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repo) CountAuditLogsByCutoff(ctx context.Context, cutoff time.Time) (expiredCount int64, validCount int64, err error) {
	if err = r.db.WithContext(ctx).
		Model(&model.AdminUserLoginAudit{}).
		Where("created_at < ?", cutoff).
		Count(&expiredCount).Error; err != nil {
		return 0, 0, xerr.WrapDBE(err, "count expired audit logs")
	}
	if err = r.db.WithContext(ctx).
		Model(&model.AdminUserLoginAudit{}).
		Where("created_at >= ?", cutoff).
		Count(&validCount).Error; err != nil {
		return 0, 0, xerr.WrapDBE(err, "count valid audit logs")
	}
	return expiredCount, validCount, nil
}

func (r *Repo) DeleteAuditLogsBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < ?", cutoff).
		Delete(&model.AdminUserLoginAudit{})
	if result.Error != nil {
		return 0, xerr.WrapDBE(result.Error, "delete expired audit logs")
	}
	return result.RowsAffected, nil
}

func (r *Repo) CountAuditLogsSince(ctx context.Context, cutoff time.Time) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.AdminUserLoginAudit{}).
		Where("created_at >= ?", cutoff).
		Count(&count).Error; err != nil {
		return 0, xerr.WrapDBE(err, "count remaining audit logs")
	}
	return count, nil
}

func (r *Repo) buildAuditQuery(ctx context.Context, filters AuditFilters) *gorm.DB {
	query := r.baseAuditSelect(r.db.WithContext(ctx))
	if actor := strings.TrimSpace(filters.Actor); actor != "" {
		like := "%" + actor + "%"
		query = query.Where("u.username LIKE ? OR u.display_name LIKE ?", like, like)
	}
	if action := strings.TrimSpace(filters.Action); action != "" {
		query = query.Where("a.action LIKE ?", "%"+action+"%")
	}
	if result := strings.TrimSpace(filters.Result); result != "" {
		query = query.Where("a.result = ?", result)
	}
	if traceID := strings.TrimSpace(filters.TraceID); traceID != "" {
		query = query.Where("a.trace_id LIKE ?", "%"+traceID+"%")
	}
	if requestID := strings.TrimSpace(filters.RequestID); requestID != "" {
		query = query.Where("a.request_id LIKE ?", "%"+requestID+"%")
	}
	if sourceIP := strings.TrimSpace(filters.SourceIP); sourceIP != "" {
		query = query.Where("a.source_ip LIKE ?", "%"+sourceIP+"%")
	}
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("a.detail LIKE ? OR a.trace_id LIKE ? OR a.action LIKE ?", like, like, like)
	}
	if filters.CreatedAt != nil {
		if start := strings.TrimSpace(filters.CreatedAt.GetStartDt()); start != "" {
			query = query.Where("a.created_at >= ?", start)
		}
		if end := strings.TrimSpace(filters.CreatedAt.GetEndDt()); end != "" {
			query = query.Where("a.created_at <= ?", end)
		}
	}
	return query
}

func (r *Repo) baseAuditSelect(query *gorm.DB) *gorm.DB {
	return query.
		Table(model.AdminUserLoginAudit{}.TableName() + " a").
		Joins("LEFT JOIN admin_user u ON u.uid = a.uid AND u.deleted_at = 0").
		Select(`
			a.id,
			a.uid,
			COALESCE(NULLIF(TRIM(u.display_name), ''), NULLIF(TRIM(u.username), ''), '') AS actor,
			a.action,
			a.result,
			a.trace_id,
			a.source_ip,
			a.duration,
			a.user_agent,
			a.detail,
			a.request_id,
			a.created_at
		`)
}

func normalizeAuditSortArgs(input []*commpb.SortArgs) []*commpb.SortArgs {
	if len(input) == 0 {
		return nil
	}
	fieldMap := map[string]string{
		"id":         "a.id",
		"uid":        "a.uid",
		"actor":      "actor",
		"action":     "a.action",
		"result":     "a.result",
		"trace_id":   "a.trace_id",
		"source_ip":  "a.source_ip",
		"created_at": "a.created_at",
	}
	out := make([]*commpb.SortArgs, 0, len(input))
	for _, item := range input {
		if item == nil {
			continue
		}
		mapped, ok := fieldMap[item.GetOrderField()]
		if !ok {
			continue
		}
		out = append(out, &commpb.SortArgs{
			OrderField: mapped,
			OrderType:  item.GetOrderType(),
		})
	}
	return out
}

func NormalizeSortArgs(input []*commpb.SortArgs) []*commpb.SortArgs {
	return normalizeAuditSortArgs(input)
}

func ValidatedTimeRange(input *commpb.TimeRange) error {
	if input == nil {
		return nil
	}
	start := strings.TrimSpace(input.GetStartDt())
	end := strings.TrimSpace(input.GetEndDt())
	if start == "" || end == "" {
		return nil
	}
	return util.CheckTimeRange(input)
}
