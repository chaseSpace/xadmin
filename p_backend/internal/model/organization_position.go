package model

import (
	"monorepo/pkg/db"
)

// OrganizationPosition maps to organization_position table.
type OrganizationPosition struct {
	db.ModelBase
	ID           int64  `gorm:"column:id"`
	Name         string `gorm:"column:name"`
	Code         string `gorm:"column:code"`
	DepartmentID int64  `gorm:"column:department_id"`
	Level        string `gorm:"column:level"`
	Hc           int32  `gorm:"column:hc"`
	Staffed      int32  `gorm:"column:staffed"`
	Status       int32  `gorm:"column:status"`
	Sort         int32  `gorm:"column:sort"`
	DeletedAt    int64  `gorm:"column:deleted_at"`
}

func (OrganizationPosition) TableName() string {
	return "organization_position"
}
