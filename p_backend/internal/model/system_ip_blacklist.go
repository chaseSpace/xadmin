package model

import (
	"time"

	"monorepo/pkg/db"
)

type SystemIPBlacklist struct {
	db.ModelBase
	ID         int64     `gorm:"column:id"`
	IP         string    `gorm:"column:ip"`
	BanType    string    `gorm:"column:ban_type"`
	StartAt    time.Time `gorm:"column:start_at"`
	EndAt      time.Time `gorm:"column:end_at"`
	Reason     string    `gorm:"column:reason"`
	Creator    string    `gorm:"column:creator"`
	Status     string    `gorm:"column:status"`
	HitCount   int32     `gorm:"column:hit_count"`
	LastAction string    `gorm:"column:last_action"`
	DeletedAt  int64     `gorm:"column:deleted_at"`
}

func (SystemIPBlacklist) TableName() string {
	return "system_ip_blacklist"
}
