package model

import "monorepo/pkg/db"

// OrganizationPositionRole maps to organization_position_role table.
type OrganizationPositionRole struct {
	db.ModelBase
	ID         int64 `gorm:"column:id"`
	PositionID int64 `gorm:"column:position_id"`
	RoleID     int64 `gorm:"column:role_id"`
}

func (OrganizationPositionRole) TableName() string {
	return "organization_position_role"
}
