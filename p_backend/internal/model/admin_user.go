package model

import (
	"time"

	"monorepo/pkg/db"
)

// AdminUser maps to admin_user table.
type AdminUser struct {
	db.ModelBase
	ID               int64      `gorm:"column:id"`
	UID              int32      `gorm:"column:uid"`
	Username         string     `gorm:"column:username"`
	PasswordHash     string     `gorm:"column:password_hash"`
	DisplayName      string     `gorm:"column:display_name"`
	Avatar           string     `gorm:"column:avatar"`
	Email            string     `gorm:"column:email"`
	Phone            string     `gorm:"column:phone"`
	Status           int32      `gorm:"column:status"`
	DepartmentID     int64      `gorm:"column:department_id"`
	PositionID       int64      `gorm:"column:position_id"`
	LimitSingleLogin bool       `gorm:"column:limit_single_login"`
	DeactivatedAt    *time.Time `gorm:"column:deactivated_at"`
	LastLoginAt      *time.Time `gorm:"column:last_login_at"`
	LastLoginIP      string     `gorm:"column:last_login_ip"`
	DeletedAt        int64      `gorm:"column:deleted_at"`
}

func (AdminUser) TableName() string {
	return "admin_user"
}
