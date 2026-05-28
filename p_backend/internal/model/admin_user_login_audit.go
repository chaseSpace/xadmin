package model

import (
	"time"
)

type AdminUserLoginAudit struct {
	ID        int64     `gorm:"column:id"`
	UID       int32     `gorm:"column:uid"`
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

func (AdminUserLoginAudit) TableName() string {
	return "admin_user_login_audit"
}
