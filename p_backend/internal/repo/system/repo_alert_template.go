package system

import (
	"context"
	"strings"
	"time"

	"monorepo/internal/model"
	"monorepo/pkg/db"
	"monorepo/pkg/xerr"
	commpb "monorepo/proto/xadminpb/commpb"
)

type AlertTemplateRow struct {
	ID        int64
	BotType   string
	Name      string
	ParseMode string
	Content   string
	CreatedAt string
	UpdatedAt string
}

func (r *Repo) ListAlertTemplates(ctx context.Context, page *commpb.PageArgs, sort []*commpb.SortArgs, keyword, botType string) ([]AlertTemplateRow, int64, error) {
	query := r.db.WithContext(ctx).
		Model(&model.SystemAlertTemplate{}).
		Where("deleted_at = 0")
	if kw := strings.TrimSpace(keyword); kw != "" {
		like := "%" + kw + "%"
		query = query.Where("name LIKE ? OR content LIKE ?", like, like)
	}
	if bt := strings.TrimSpace(botType); bt != "" {
		query = query.Where("bot_type = ?", bt)
	}
	var models []model.SystemAlertTemplate
	total, err := db.Paginate(query, page, sort, []string{"id", "bot_type", "name", "parse_mode", "created_at", "updated_at"}, &models)
	if err != nil {
		return nil, 0, err
	}
	rows := make([]AlertTemplateRow, 0, len(models))
	for _, m := range models {
		rows = append(rows, AlertTemplateRow{
			ID: m.ID, BotType: m.BotType, Name: m.Name, ParseMode: m.ParseMode, Content: m.Content,
			CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: m.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return rows, total, nil
}

func (r *Repo) SaveAlertTemplate(ctx context.Context, row *AlertTemplateRow) error {
	if row.ID > 0 {
		ret := r.db.WithContext(ctx).
			Model(&model.SystemAlertTemplate{}).
			Where("id = ? AND deleted_at = 0", row.ID).
			Updates(map[string]any{
				"bot_type":   strings.TrimSpace(row.BotType),
				"name":       strings.TrimSpace(row.Name),
				"parse_mode": strings.TrimSpace(row.ParseMode),
				"content":    strings.TrimSpace(row.Content),
			})
		return xerr.WrapDBUpdateMiss(ret, "告警模板不存在")
	}
	m := &model.SystemAlertTemplate{
		BotType:   strings.TrimSpace(row.BotType),
		Name:      strings.TrimSpace(row.Name),
		ParseMode: strings.TrimSpace(row.ParseMode),
		Content:   strings.TrimSpace(row.Content),
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return xerr.WrapDBE(err, "create alert template")
	}
	row.ID = m.ID
	return nil
}

func (r *Repo) DeleteAlertTemplate(ctx context.Context, id int64) error {
	ret := r.db.WithContext(ctx).
		Model(&model.SystemAlertTemplate{}).
		Where("id = ? AND deleted_at = 0", id).
		Update("deleted_at", time.Now().UnixMilli())
	return xerr.WrapDBUpdateMiss(ret, "告警模板不存在")
}
