package model

import "time"

type SystemAlertBot struct {
	ID        int64     `gorm:"column:id"`
	Name      string    `gorm:"column:name"`
	Username  string    `gorm:"column:username"`
	Token     string    `gorm:"column:token"`
	BotType   string    `gorm:"column:bot_type"`
	Enabled   bool      `gorm:"column:enabled"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	DeletedAt int64     `gorm:"column:deleted_at"`
}

func (SystemAlertBot) TableName() string {
	return "system_alert_bot"
}
