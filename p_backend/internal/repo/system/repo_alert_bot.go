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

type AlertBotRow struct {
	ID        int64
	Name      string
	Username  string
	Token     string
	BotType   string
	Enabled   bool
	CreatedAt string
	UpdatedAt string
}

func (r *Repo) ListAlertBots(ctx context.Context, page *commpb.PageArgs, sort []*commpb.SortArgs, keyword, botType string) ([]AlertBotRow, int64, error) {
	query := r.db.WithContext(ctx).
		Model(&model.SystemAlertBot{}).
		Where("deleted_at = 0")
	if kw := strings.TrimSpace(keyword); kw != "" {
		like := "%" + kw + "%"
		query = query.Where("name LIKE ? OR username LIKE ?", like, like)
	}
	if bt := strings.TrimSpace(botType); bt != "" {
		query = query.Where("bot_type = ?", bt)
	}
	var models []model.SystemAlertBot
	total, err := db.Paginate(query, page, sort, []string{"id", "name", "bot_type", "enabled", "created_at", "updated_at"}, &models)
	if err != nil {
		return nil, 0, err
	}
	rows := make([]AlertBotRow, 0, len(models))
	for _, m := range models {
		rows = append(rows, AlertBotRow{
			ID: m.ID, Name: m.Name, Username: m.Username, Token: m.Token,
			BotType: m.BotType, Enabled: m.Enabled,
			CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: m.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return rows, total, nil
}

func (r *Repo) GetAlertBot(ctx context.Context, id int64) (*AlertBotRow, error) {
	var m model.SystemAlertBot
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at = 0", id).First(&m).Error; err != nil {
		return nil, xerr.WrapDBNotFound(err, "告警机器人不存在")
	}
	row := AlertBotRow{
		ID: m.ID, Name: m.Name, Username: m.Username, Token: m.Token,
		BotType: m.BotType, Enabled: m.Enabled,
		CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: m.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	return &row, nil
}

func (r *Repo) SaveAlertBot(ctx context.Context, row *AlertBotRow) error {
	if row.ID > 0 {
		ret := r.db.WithContext(ctx).
			Model(&model.SystemAlertBot{}).
			Where("id = ? AND deleted_at = 0", row.ID).
			Updates(map[string]any{
				"name":     strings.TrimSpace(row.Name),
				"username": strings.TrimSpace(row.Username),
				"token":    strings.TrimSpace(row.Token),
				"bot_type": strings.TrimSpace(row.BotType),
				"enabled":  row.Enabled,
			})
		return xerr.WrapDBUpdateMiss(ret, "告警机器人不存在")
	}
	m := &model.SystemAlertBot{
		Name:     strings.TrimSpace(row.Name),
		Username: strings.TrimSpace(row.Username),
		Token:    strings.TrimSpace(row.Token),
		BotType:  strings.TrimSpace(row.BotType),
		Enabled:  row.Enabled,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return xerr.WrapDBE(err, "create alert bot")
	}
	row.ID = m.ID
	return nil
}

func (r *Repo) DeleteAlertBot(ctx context.Context, id int64) error {
	ret := r.db.WithContext(ctx).
		Model(&model.SystemAlertBot{}).
		Where("id = ? AND deleted_at = 0", id).
		Update("deleted_at", time.Now().UnixMilli())
	return xerr.WrapDBUpdateMiss(ret, "告警机器人不存在")
}

func (r *Repo) CountScenesByBotID(ctx context.Context, botID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.SystemAlertScene{}).
		Where("bot_id = ? AND deleted_at = 0", botID).
		Count(&count).Error
	if err != nil {
		return 0, xerr.WrapDBE(err, "count scenes by bot_id")
	}
	return count, nil
}

func (r *Repo) ListSceneKeysByBotIDs(ctx context.Context, botIDs []int64) (map[int64][]string, error) {
	if len(botIDs) == 0 {
		return nil, nil
	}
	var rows []struct {
		BotID    int64  `gorm:"column:bot_id"`
		SceneKey string `gorm:"column:scene_key"`
	}
	err := r.db.WithContext(ctx).
		Model(&model.SystemAlertScene{}).
		Select("bot_id, scene_key").
		Where("bot_id IN ? AND deleted_at = 0", botIDs).
		Find(&rows).Error
	if err != nil {
		return nil, xerr.WrapDBE(err, "list scene keys by bot ids")
	}
	result := make(map[int64][]string)
	for _, r := range rows {
		result[r.BotID] = append(result[r.BotID], r.SceneKey)
	}
	return result, nil
}
