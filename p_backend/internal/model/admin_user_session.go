package model

import (
	"time"

	"monorepo/pkg/db"
)

type AdminUserSession struct {
	db.ModelBase
	ID            int64      `gorm:"column:id"`
	SessionID     string     `gorm:"column:session_id"`
	UID           int32      `gorm:"column:uid"`
	TokenHash     string     `gorm:"column:token_hash"`
	Status        string     `gorm:"column:status"`
	LoginIP       string     `gorm:"column:login_ip"`
	UserAgent     string     `gorm:"column:user_agent"`
	LastSeenAt    *time.Time `gorm:"column:last_seen_at"`
	ExpiredAt     time.Time  `gorm:"column:expired_at"`
	RevokedAt     *time.Time `gorm:"column:revoked_at"`
	RevokedReason string     `gorm:"column:revoked_reason"`
}

func (AdminUserSession) TableName() string {
	return "admin_user_session"
}
