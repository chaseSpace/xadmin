package system

type IPBlacklistFilters struct {
	Keyword string
	Status  string
	BanType string
	Creator string
}

type IPBlacklistRow struct {
	ID         int64
	IP         string
	BanType    string
	StartAt    string
	EndAt      string
	Reason     string
	Creator    string
	Status     string
	HitCount   int32
	LastAction string
	UpdatedAt  string
}

type WarmTipFilters struct {
	Keyword string
	TipType string
	Status  *int32
}

type WarmTipRow struct {
	ID        int64
	TipType   string
	ContentZh string
	ContentEn string
	Sort      int32
	Status    int32
	UpdatedAt string
}
