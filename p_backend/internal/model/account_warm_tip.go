package model

import "monorepo/pkg/db"

type AccountWarmTip struct {
	db.ModelBase
	ID        int64  `gorm:"column:id"`
	TipType   string `gorm:"column:tip_type"`
	ContentZh string `gorm:"column:content_zh"`
	ContentEn string `gorm:"column:content_en"`
	Sort      int32  `gorm:"column:sort"`
	Status    int32  `gorm:"column:status"`
	DeletedAt int64  `gorm:"column:deleted_at"`
}

func (AccountWarmTip) TableName() string {
	return "account_warm_tip"
}
