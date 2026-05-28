package model

import (
	"monorepo/pkg/db"
)

// PermissionMenu maps to permission_menu table.
type PermissionMenu struct {
	db.ModelBase
	ID            int64  `gorm:"column:id"`
	ParentID      int64  `gorm:"column:parent_id"`
	Name          string `gorm:"column:name"`
	RoutePath     string `gorm:"column:route_path"`
	ComponentPath string `gorm:"column:component_path"`
	MenuType      int32  `gorm:"column:menu_type"`
	PermissionKey string `gorm:"column:permission_key"`
	Sort          int32  `gorm:"column:sort"`
	Status        int32  `gorm:"column:status"`
	DeletedAt     int64  `gorm:"column:deleted_at"`
}

func (PermissionMenu) TableName() string {
	return "permission_menu"
}

// PermissionRole maps to permission_role table.
type PermissionRole struct {
	db.ModelBase
	ID        int64  `gorm:"column:id"`
	RoleName  string `gorm:"column:role_name"`
	RoleType  int32  `gorm:"column:role_type"`
	DeletedAt int64  `gorm:"column:deleted_at"`
}

func (PermissionRole) TableName() string {
	return "permission_role"
}

// PermissionRoleMenu maps to permission_role_menu table.
type PermissionRoleMenu struct {
	db.ModelBase
	ID     int64 `gorm:"column:id"`
	RoleID int64 `gorm:"column:role_id"`
	MenuID int64 `gorm:"column:menu_id"`
}

func (PermissionRoleMenu) TableName() string {
	return "permission_role_menu"
}

// PermissionRoleUser maps to permission_role_user table.
type PermissionRoleUser struct {
	db.ModelBase
	ID     int64 `gorm:"column:id"`
	RoleID int64 `gorm:"column:role_id"`
	UID    int32 `gorm:"column:uid"`
}

func (PermissionRoleUser) TableName() string {
	return "permission_role_user"
}
