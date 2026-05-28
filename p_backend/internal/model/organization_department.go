package model

import (
	"monorepo/pkg/db"
)

// OrganizationDepartment maps to organization_department table.
type OrganizationDepartment struct {
	db.ModelBase
	ID          int64  `gorm:"column:id"`
	ParentID    int64  `gorm:"column:parent_id"`
	Name        string `gorm:"column:name"`
	Code        string `gorm:"column:code"`
	Status      int32  `gorm:"column:status"`
	MemberCount int32  `gorm:"column:member_count"`
	Sort        int32  `gorm:"column:sort"`
	DeletedAt   int64  `gorm:"column:deleted_at"`
}

func (OrganizationDepartment) TableName() string {
	return "organization_department"
}
