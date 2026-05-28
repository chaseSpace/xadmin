package model

import "time"

type SystemAlertTemplate struct {
	ID        int64     `gorm:"column:id"`
	BotType   string    `gorm:"column:bot_type"`
	Name      string    `gorm:"column:name"`
	ParseMode string    `gorm:"column:parse_mode"`
	Content   string    `gorm:"column:content"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	DeletedAt int64     `gorm:"column:deleted_at"`
}

func (SystemAlertTemplate) TableName() string {
	return "system_alert_template"
}
