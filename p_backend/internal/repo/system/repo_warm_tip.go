package system

import (
	"context"
	"strings"

	"monorepo/internal/model"
	"monorepo/pkg/db"
	"monorepo/pkg/xerr"
	commpb "monorepo/proto/xadminpb/commpb"

	"gorm.io/gorm"
)

var warmTipOrderFieldMap = map[string]string{
	"id":         "t.id",
	"tip_type":   "t.tip_type",
	"sort":       "t.sort",
	"status":     "t.status",
	"updated_at": "t.updated_at",
}

func (r *Repo) ListWarmTips(ctx context.Context, page *commpb.PageArgs, sortArgs []*commpb.SortArgs, filters WarmTipFilters) ([]WarmTipRow, int64, error) {
	rows := make([]WarmTipRow, 0, page.GetPs())
	query := r.db.WithContext(ctx).
		Table("account_warm_tip t").
		Select(`
t.id,
t.tip_type,
t.content_zh,
t.content_en,
t.sort,
t.status,
TO_CHAR(t.updated_at, 'YYYY-MM-DD HH24:MI:SS') AS updated_at
`).
		Where("t.deleted_at = 0")
	query = applyWarmTipFilters(query, filters)
	if len(sortArgs) == 0 {
		query = query.Order("t.sort asc, t.id asc")
	}
	sortArgs = mapWarmTipSortArgs(sortArgs)
	total, err := db.Paginate(
		query,
		page,
		sortArgs,
		[]string{"t.id", "t.tip_type", "t.sort", "t.status", "t.updated_at"},
		&rows,
		db.PaginateArgs{AppendCreatedAtDesc: false},
	)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func mapWarmTipSortArgs(input []*commpb.SortArgs) []*commpb.SortArgs {
	if len(input) == 0 {
		return input
	}
	output := make([]*commpb.SortArgs, 0, len(input))
	for _, item := range input {
		if item == nil {
			continue
		}
		next := &commpb.SortArgs{
			OrderField: strings.TrimSpace(item.GetOrderField()),
			OrderType:  item.GetOrderType(),
		}
		if mapped, ok := warmTipOrderFieldMap[next.OrderField]; ok {
			next.OrderField = mapped
		}
		output = append(output, next)
	}
	return output
}

func (r *Repo) CreateWarmTip(ctx context.Context, row *WarmTipRow) error {
	if row == nil {
		return xerr.NewWithDetail(xerr.CodeParamError, "warm tip row is empty")
	}
	data := &model.AccountWarmTip{
		TipType:   strings.TrimSpace(row.TipType),
		ContentZh: strings.TrimSpace(row.ContentZh),
		ContentEn: strings.TrimSpace(row.ContentEn),
		Sort:      row.Sort,
		Status:    row.Status,
		DeletedAt: 0,
	}
	if err := xerr.WrapDBE(r.db.WithContext(ctx).Create(data).Error, "create warm tip"); err != nil {
		return err
	}
	row.ID = data.ID
	return nil
}

func (r *Repo) UpdateWarmTip(ctx context.Context, row *WarmTipRow) error {
	if row == nil {
		return xerr.NewWithDetail(xerr.CodeParamError, "warm tip row is empty")
	}
	ret := r.db.WithContext(ctx).
		Model(&model.AccountWarmTip{}).
		Where("id = ? AND deleted_at = 0", row.ID).
		Updates(map[string]any{
			"tip_type":   strings.TrimSpace(row.TipType),
			"content_zh": strings.TrimSpace(row.ContentZh),
			"content_en": strings.TrimSpace(row.ContentEn),
			"sort":       row.Sort,
			"status":     row.Status,
		})
	return xerr.WrapDBUpdateMiss(ret, "关怀提示不存在")
}

func (r *Repo) DeleteWarmTip(ctx context.Context, id int64) error {
	ret := r.db.WithContext(ctx).
		Model(&model.AccountWarmTip{}).
		Where("id = ? AND deleted_at = 0", id).
		Update("deleted_at", id)
	return xerr.WrapDBUpdateMiss(ret, "关怀提示不存在")
}

func applyWarmTipFilters(query *gorm.DB, filters WarmTipFilters) *gorm.DB {
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("(t.content_zh LIKE ? OR t.content_en LIKE ?)", like, like)
	}
	if tipType := strings.TrimSpace(filters.TipType); tipType != "" {
		query = query.Where("t.tip_type = ?", tipType)
	}
	if filters.Status != nil {
		query = query.Where("t.status = ?", *filters.Status)
	}
	return query
}
