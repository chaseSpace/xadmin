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

type AlertSceneRow struct {
	ID             int64
	SceneKey       string
	BotID          int64
	ParseMode      string
	GroupName      string
	GroupID        string
	NotifyTemplate string
	CreatedAt      string
	UpdatedAt      string
}

func (r *Repo) ListAlertScenes(ctx context.Context, page *commpb.PageArgs, sort []*commpb.SortArgs, keyword string) ([]AlertSceneRow, int64, error) {
	query := r.db.WithContext(ctx).
		Model(&model.SystemAlertScene{}).
		Where("deleted_at = 0")
	if kw := strings.TrimSpace(keyword); kw != "" {
		like := "%" + kw + "%"
		query = query.Where("scene_key LIKE ? OR group_name LIKE ?", like, like)
	}
	var models []model.SystemAlertScene
	total, err := db.Paginate(query, page, sort, []string{"id", "scene_key", "group_name", "created_at", "updated_at"}, &models)
	if err != nil {
		return nil, 0, err
	}
	rows := make([]AlertSceneRow, 0, len(models))
	for _, m := range models {
		rows = append(rows, AlertSceneRow{
			ID:             m.ID,
			SceneKey:       m.SceneKey,
			BotID:          m.BotID,
			ParseMode:      m.ParseMode,
			GroupName:      m.GroupName,
			GroupID:        m.GroupID,
			NotifyTemplate: m.NotifyTemplate,
			CreatedAt:      m.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      m.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return rows, total, nil
}

func (r *Repo) SaveAlertScene(ctx context.Context, row *AlertSceneRow) error {
	if row.ID > 0 {
		ret := r.db.WithContext(ctx).
			Model(&model.SystemAlertScene{}).
			Where("id = ? AND deleted_at = 0", row.ID).
			Updates(map[string]any{
				"scene_key":       strings.TrimSpace(row.SceneKey),
				"bot_id":          row.BotID,
				"parse_mode":      strings.TrimSpace(row.ParseMode),
				"group_name":      strings.TrimSpace(row.GroupName),
				"group_id":        strings.TrimSpace(row.GroupID),
				"notify_template": strings.TrimSpace(row.NotifyTemplate),
			})
		return xerr.WrapDBUpdateMiss(ret, "告警场景不存在")
	}
	m := &model.SystemAlertScene{
		SceneKey:       strings.TrimSpace(row.SceneKey),
		BotID:          row.BotID,
		ParseMode:      strings.TrimSpace(row.ParseMode),
		GroupName:      strings.TrimSpace(row.GroupName),
		GroupID:        strings.TrimSpace(row.GroupID),
		NotifyTemplate: strings.TrimSpace(row.NotifyTemplate),
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return xerr.WrapDBE(err, "create alert scene")
	}
	row.ID = m.ID
	return nil
}

func (r *Repo) GetAlertScene(ctx context.Context, id int64) (*AlertSceneRow, error) {
	var m model.SystemAlertScene
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at = 0", id).First(&m).Error; err != nil {
		return nil, xerr.WrapDBNotFound(err, "告警场景不存在")
	}
	row := AlertSceneRow{
		ID: m.ID, SceneKey: m.SceneKey, BotID: m.BotID, ParseMode: m.ParseMode,
		GroupName: m.GroupName, GroupID: m.GroupID, NotifyTemplate: m.NotifyTemplate,
		CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: m.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	return &row, nil
}

func (r *Repo) DeleteAlertScene(ctx context.Context, id int64) error {
	ret := r.db.WithContext(ctx).
		Model(&model.SystemAlertScene{}).
		Where("id = ? AND deleted_at = 0", id).
		Update("deleted_at", time.Now().UnixMilli())
	return xerr.WrapDBUpdateMiss(ret, "告警场景不存在")
}

func (r *Repo) FindScenesByKey(ctx context.Context, sceneKey string) ([]AlertSceneRow, error) {
	var models []model.SystemAlertScene
	err := r.db.WithContext(ctx).
		Where("scene_key = ? AND deleted_at = 0", sceneKey).
		Find(&models).Error
	if err != nil {
		return nil, xerr.WrapDBE(err, "find scenes by key")
	}
	rows := make([]AlertSceneRow, 0, len(models))
	for _, m := range models {
		rows = append(rows, AlertSceneRow{
			ID: m.ID, SceneKey: m.SceneKey, BotID: m.BotID, ParseMode: m.ParseMode,
			GroupName: m.GroupName, GroupID: m.GroupID, NotifyTemplate: m.NotifyTemplate,
		})
	}
	return rows, nil
}
