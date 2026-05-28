package model

import "monorepo/pkg/db"

type AdminUserPersonalSetting struct {
	db.ModelBase
	ID                           int64  `gorm:"column:id"`
	UID                          int32  `gorm:"column:uid"`
	LimitSingleLogin             bool   `gorm:"column:limit_single_login"`
	BackgroundImageURL           string `gorm:"column:background_image_url"`
	Locale                       string `gorm:"column:locale"`
	GlobalBackgroundApplyEnabled bool   `gorm:"column:global_background_apply_enabled"`
	WarmTipIntervalMinutes       int32  `gorm:"column:warm_tip_interval_minutes"`
}

func (AdminUserPersonalSetting) TableName() string {
	return "admin_user_personal_setting"
}
