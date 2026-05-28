package resource

import (
	"context"
	"strings"
	"time"

	"monorepo/internal/model"
	"monorepo/pkg/db"
	"monorepo/pkg/xerr"
	commpb "monorepo/proto/xadminpb/commpb"

	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

type FileFilters struct {
	Keyword  string
	FileType string
	Exists   *bool
}

type FileRow struct {
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
	CreatedAt       time.Time  `gorm:"column:created_at"`
	LastAccessAt    *time.Time `gorm:"column:last_access_at"`
	AccessCount     int32      `gorm:"column:access_count"`
	CreatorUID      int32      `gorm:"column:creator_uid"`
}

const fileExistsFilterCondition = `"exists" = ? AND exists_checked_at IS NOT NULL AND exists_checked_at >= ?`
const fileKeywordFilterCondition = "(name LIKE ? OR remark LIKE ? OR file_url LIKE ?)"

func NewRepo() *Repo {
	return &Repo{db: db.GetDatabase()}
}

func NewRepoWithDB(database *gorm.DB) *Repo {
	return &Repo{db: database}
}

func (r *Repo) ListFiles(ctx context.Context, page *commpb.PageArgs, sort []*commpb.SortArgs, filters FileFilters) ([]FileRow, int64, int64, error) {
	rows := make([]FileRow, 0, page.GetPs())
	query := r.db.WithContext(ctx).
		Model(&model.ResourceFile{}).
		Where("deleted_at = 0")
	query = applyFileFilters(query, filters)

	var totalSizeBytes int64
	if err := query.Session(&gorm.Session{}).Select("COALESCE(SUM(size_bytes), 0)").Scan(&totalSizeBytes).Error; err != nil {
		return nil, 0, 0, xerr.WrapDBE(err, "sum resource file size")
	}

	if len(sort) == 0 {
		query = query.Order("created_at desc")
	}
	total, err := db.Paginate(query, page, sort, []string{"id", "name", "file_type", "size_bytes", "created_at", "last_access_at", "access_count"}, &rows)
	if err != nil {
		return nil, 0, 0, err
	}
	return rows, total, totalSizeBytes, nil
}

func applyFileFilters(query *gorm.DB, filters FileFilters) *gorm.DB {
	if fileType := strings.TrimSpace(filters.FileType); fileType != "" {
		query = query.Where("file_type = ?", fileType)
	}
	if filters.Exists != nil {
		query = query.Where(fileExistsFilterCondition, *filters.Exists, time.Now().Add(-24*time.Hour))
	}
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where(fileKeywordFilterCondition, like, like, like)
	}
	return query
}

func (r *Repo) CreateFile(ctx context.Context, row *FileRow) error {
	data := &model.ResourceFile{
		FileType:    strings.TrimSpace(row.FileType),
		Name:        strings.TrimSpace(row.Name),
		FileURL:     strings.TrimSpace(row.FileURL),
		MimeType:    strings.TrimSpace(row.MimeType),
		Extension:   strings.TrimSpace(row.Extension),
		SizeBytes:   row.SizeBytes,
		Remark:      strings.TrimSpace(row.Remark),
		RequireAuth: row.RequireAuth,
		AccessMode:  strings.TrimSpace(row.AccessMode),
		Exists:      true,
		CreatorUID:  row.CreatorUID,
		DeletedAt:   0,
		AccessCount: 0,
	}
	if err := xerr.WrapDBDuplicate(r.db.WithContext(ctx).Create(data).Error, "文件记录已存在"); err != nil {
		return err
	}
	row.ID = data.ID
	row.CreatedAt = data.CreatedAt
	return nil
}

func (r *Repo) UpdateFile(ctx context.Context, id int64, fileType string, name string, remark string, requireAuth bool, accessMode string) error {
	ret := r.db.WithContext(ctx).
		Model(&model.ResourceFile{}).
		Where("id = ? AND deleted_at = 0", id).
		Updates(map[string]any{
			"file_type":    strings.TrimSpace(fileType),
			"name":         strings.TrimSpace(name),
			"remark":       strings.TrimSpace(remark),
			"require_auth": requireAuth,
			"access_mode":  strings.TrimSpace(accessMode),
		})
	return xerr.WrapDBUpdateMiss(ret, "文件记录不存在")
}

func (r *Repo) DeleteFile(ctx context.Context, id int64) error {
	ret := r.db.WithContext(ctx).
		Model(&model.ResourceFile{}).
		Where("id = ? AND deleted_at = 0", id).
		Update("deleted_at", time.Now().Unix())
	return xerr.WrapDBUpdateMiss(ret, "文件记录不存在")
}

func (r *Repo) GetFile(ctx context.Context, id int64) (*FileRow, error) {
	var row FileRow
	err := r.db.WithContext(ctx).
		Model(&model.ResourceFile{}).
		Where("id = ? AND deleted_at = 0", id).
		First(&row).Error
	if err != nil {
		return nil, xerr.WrapDBNotFound(err, "文件记录不存在")
	}
	return &row, nil
}

func (r *Repo) GetFileByURL(ctx context.Context, fileURL string) (*FileRow, error) {
	var row FileRow
	err := r.db.WithContext(ctx).
		Model(&model.ResourceFile{}).
		Where("file_url = ? AND deleted_at = 0", strings.TrimSpace(fileURL)).
		First(&row).Error
	if err != nil {
		return nil, xerr.WrapDBNotFound(err, "文件记录不存在")
	}
	return &row, nil
}

func (r *Repo) MarkAccess(ctx context.Context, id int64) error {
	now := time.Now()
	ret := r.db.WithContext(ctx).
		Model(&model.ResourceFile{}).
		Where("id = ? AND deleted_at = 0", id).
		Updates(map[string]any{
			"last_access_at": now,
			"access_count":   gorm.Expr("access_count + 1"),
		})
	return xerr.WrapDBUpdateMiss(ret, "文件记录不存在")
}

func (r *Repo) ListAllFiles(ctx context.Context, fileType string) ([]FileRow, error) {
	rows := make([]FileRow, 0)
	query := r.db.WithContext(ctx).
		Model(&model.ResourceFile{}).
		Where("deleted_at = 0")
	if fileType := strings.TrimSpace(fileType); fileType != "" {
		query = query.Where("file_type = ?", fileType)
	}
	err := query.Order("id asc").Find(&rows).Error
	if err != nil {
		return nil, xerr.WrapDBE(err, "query resource files")
	}
	return rows, nil
}

func (r *Repo) UpdateFileExists(ctx context.Context, id int64, exists bool) error {
	ret := r.db.WithContext(ctx).
		Model(&model.ResourceFile{}).
		Where("id = ? AND deleted_at = 0", id).
		Updates(map[string]any{
			"exists":            exists,
			"exists_checked_at": time.Now(),
		})
	return xerr.WrapDBUpdateMiss(ret, "文件记录不存在")
}
