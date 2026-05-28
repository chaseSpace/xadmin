package model

import "time"

type SystemAlertScene struct {
	ID             int64     `gorm:"column:id"`
	SceneKey       string    `gorm:"column:scene_key"`
	BotID          int64     `gorm:"column:bot_id"`
	ParseMode      string    `gorm:"column:parse_mode"`
	GroupName      string    `gorm:"column:group_name"`
	GroupID        string    `gorm:"column:group_id"`
	NotifyTemplate string    `gorm:"column:notify_template"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
	DeletedAt      int64     `gorm:"column:deleted_at"`
}

func (SystemAlertScene) TableName() string {
	return "system_alert_scene"
}
