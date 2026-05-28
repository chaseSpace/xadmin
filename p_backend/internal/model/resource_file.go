package model

import (
	"time"

	"monorepo/pkg/db"
)

type ResourceFile struct {
	db.ModelBase
	ID              int64      `gorm:"column:id"`
	FileType        string     `gorm:"column:file_type"`
	Name            string     `gorm:"column:name"`
	FileURL         string     `gorm:"column:file_url"`
	MimeType        string     `gorm:"column:mime_type"`
	Extension       string     `gorm:"column:extension"`
	SizeBytes       int64      `gorm:"column:size_bytes"`
	Remark          string     `gorm:"column:remark"`
	RequireAuth     bool       `gorm:"column:require_auth"`
	AccessMode      string     `gorm:"column:access_mode"`
	Exists          bool       `gorm:"column:exists"`
	ExistsCheckedAt *time.Time `gorm:"column:exists_checked_at"`
	LastAccessAt    *time.Time `gorm:"column:last_access_at"`
	AccessCount     int32      `gorm:"column:access_count"`
	CreatorUID      int32      `gorm:"column:creator_uid"`
	DeletedAt       int64      `gorm:"column:deleted_at"`
}

func (ResourceFile) TableName() string {
	return "resource_file"
}
