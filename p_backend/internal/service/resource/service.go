package resource

import (
	"context"
	"mime/multipart"
	"strings"
	"sync"
	"time"

	"monorepo/internal/repo/assert"
	resourcerepo "monorepo/internal/repo/resource"
	"monorepo/pkg/xerr"
	xadmin "monorepo/proto/xadminpb"
	commpb "monorepo/proto/xadminpb/commpb"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Service interface {
	ListFiles(ctx context.Context, req *xadmin.ResourceFilesReq) (*xadmin.ResourceFilesResp, error)
	UploadFile(ctx context.Context, uid int32, name string, remark string, requireAuth bool, accessMode string, requestedFileType string, fileHeader *multipart.FileHeader) (*xadmin.ResourceUploadFileResp, error)
	UpdateFile(ctx context.Context, req *xadmin.ResourceUpdateFileReq) (*xadmin.ResourceActionResp, error)
	DeleteFile(ctx context.Context, req *xadmin.ResourceFileActionReq) (*xadmin.ResourceActionResp, error)
	AccessFile(ctx context.Context, req *xadmin.ResourceFileActionReq) (*xadmin.ResourceFileAccessResp, error)
	CheckFiles(ctx context.Context, req *xadmin.ResourceCheckFilesReq) (*xadmin.ResourceCheckFilesResp, error)
}

type service struct {
	repo       *resourcerepo.Repo
	assertRepo *assert.Repo
}

var checkFilesMu sync.Mutex

func NewService() Service {
	return &service{repo: resourcerepo.NewRepo(), assertRepo: assert.NewRepo()}
}

func (s *service) ListFiles(ctx context.Context, req *xadmin.ResourceFilesReq) (*xadmin.ResourceFilesResp, error) {
	page := req.GetPage()
	if page == nil {
		page = &commpb.PageArgs{Pn: 1, Ps: 10}
	}
	if page.GetPn() <= 0 {
		page.Pn = 1
	}
	if page.GetPs() <= 0 {
		page.Ps = 10
	}
	rows, total, totalSizeBytes, err := s.repo.ListFiles(ctx, page, req.GetSort(), resourcerepo.FileFilters{
		Keyword:  strings.TrimSpace(req.GetKeyword()),
		FileType: resourceTypeToDB(req.GetFileType()),
		Exists:   req.Exists,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*xadmin.ResourceFileItem, 0, len(rows))
	for _, row := range rows {
		row = s.refreshFileExists(ctx, row)
		items = append(items, s.mapFileItem(ctx, row))
	}
	return &xadmin.ResourceFilesResp{Items: items, Total: total, TotalSizeBytes: totalSizeBytes, Page: &commpb.PageArgs{Pn: page.GetPn(), Ps: page.GetPs()}}, nil
}

func (s *service) UploadFile(ctx context.Context, uid int32, name string, remark string, requireAuth bool, accessMode string, requestedFileType string, fileHeader *multipart.FileHeader) (*xadmin.ResourceUploadFileResp, error) {
	if fileHeader == nil {
		return nil, xerr.NewWithDetail(xerr.CodeBadRequest, "missing upload file")
	}
	saved, err := s.assertRepo.SaveResourceUploadedFileMeta(ctx, uid, fileHeader, requestedFileType)
	if err != nil {
		return nil, err
	}
	if requested := strings.TrimSpace(requestedFileType); requested != "" && requested != saved.ResourceType {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "res.type_mismatch")
	}
	displayName := strings.TrimSpace(name)
	if displayName == "" {
		displayName = strings.TrimSpace(fileHeader.Filename)
	}
	row := &resourcerepo.FileRow{
		FileType:    saved.ResourceType,
		Name:        displayName,
		FileURL:     saved.URL,
		MimeType:    saved.ContentType,
		Extension:   saved.Extension,
		SizeBytes:   saved.Size,
		Remark:      strings.TrimSpace(remark),
		RequireAuth: requireAuth,
		AccessMode:  normalizeAccessMode(accessMode),
		CreatorUID:  uid,
	}
	if len([]rune(row.Remark)) > 50 {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "res.remark_too_long")
	}
	if err := s.repo.CreateFile(ctx, row); err != nil {
		return nil, err
	}
	return &xadmin.ResourceUploadFileResp{Item: s.mapFileItem(ctx, *row)}, nil
}

func (s *service) UpdateFile(ctx context.Context, req *xadmin.ResourceUpdateFileReq) (*xadmin.ResourceActionResp, error) {
	if len([]rune(strings.TrimSpace(req.GetRemark()))) > 50 {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "res.remark_too_long")
	}
	fileType := resourceTypeToDB(req.GetFileType())
	if fileType == "" {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "res.select_type")
	}
	if err := s.repo.UpdateFile(ctx, req.GetId(), fileType, req.GetName(), req.GetRemark(), req.GetRequireAuth(), accessModeToDB(req.GetAccessMode())); err != nil {
		return nil, err
	}
	return &xadmin.ResourceActionResp{Success: true, Action: "update"}, nil
}

func (s *service) DeleteFile(ctx context.Context, req *xadmin.ResourceFileActionReq) (*xadmin.ResourceActionResp, error) {
	if err := s.repo.DeleteFile(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &xadmin.ResourceActionResp{Success: true, Action: "delete"}, nil
}

func (s *service) AccessFile(ctx context.Context, req *xadmin.ResourceFileActionReq) (*xadmin.ResourceFileAccessResp, error) {
	row, err := s.repo.GetFile(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	go func(id int64) {
		_ = s.repo.MarkAccess(context.Background(), id)
	}(req.GetId())
	return &xadmin.ResourceFileAccessResp{
		FileUrl:      row.FileURL,
		AccessCount:  row.AccessCount,
		LastAccessAt: formatTime(row.LastAccessAt),
	}, nil
}

func (s *service) CheckFiles(ctx context.Context, req *xadmin.ResourceCheckFilesReq) (*xadmin.ResourceCheckFilesResp, error) {
	if !checkFilesMu.TryLock() {
		return nil, xerr.NewBiz(xerr.CodeConflict, "res.check_in_progress")
	}
	defer checkFilesMu.Unlock()

	rows, err := s.repo.ListAllFiles(ctx, resourceTypeToDB(req.GetFileType()))
	if err != nil {
		return nil, err
	}
	var existsCount int64
	var missingCount int64
	for _, row := range rows {
		exists := s.assertRepo.FileExists(ctx, row.FileURL)
		if exists {
			existsCount++
		} else {
			missingCount++
		}
		if err := s.repo.UpdateFileExists(ctx, row.ID, exists); err != nil {
			return nil, err
		}
	}
	return &xadmin.ResourceCheckFilesResp{
		CheckedCount: int64(len(rows)),
		ExistsCount:  existsCount,
		MissingCount: missingCount,
	}, nil
}

func (s *service) mapFileItem(ctx context.Context, row resourcerepo.FileRow) *xadmin.ResourceFileItem {
	return &xadmin.ResourceFileItem{
		Id:              row.ID,
		FileType:        strings.TrimSpace(row.FileType),
		Name:            strings.TrimSpace(row.Name),
		FileUrl:         strings.TrimSpace(row.FileURL),
		MimeType:        strings.TrimSpace(row.MimeType),
		Extension:       strings.TrimSpace(row.Extension),
		SizeBytes:       row.SizeBytes,
		Remark:          strings.TrimSpace(row.Remark),
		UploadedAt:      row.CreatedAt.Format("2006-01-02 15:04:05"),
		LastAccessAt:    formatTime(row.LastAccessAt),
		AccessCount:     row.AccessCount,
		Exists:          fileExistsValue(row.Exists, row.ExistsCheckedAt),
		RequireAuth:     row.RequireAuth,
		ExistsCheckedAt: formatTime(row.ExistsCheckedAt),
		AccessMode:      normalizeAccessMode(row.AccessMode),
	}
}

func (s *service) refreshFileExists(ctx context.Context, row resourcerepo.FileRow) resourcerepo.FileRow {
	exists := s.assertRepo.FileExists(ctx, row.FileURL)
	if exists != row.Exists {
		if err := s.repo.UpdateFileExists(ctx, row.ID, exists); err != nil {
			return row
		}
	}
	now := time.Now()
	row.Exists = exists
	row.ExistsCheckedAt = &now
	return row
}

func fileExistsValue(exists bool, checkedAt *time.Time) *wrapperspb.BoolValue {
	if checkedAt == nil || time.Since(*checkedAt) > 24*time.Hour {
		return nil
	}
	return wrapperspb.Bool(exists)
}

func formatTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func resourceTypeToDB(fileType xadmin.ResourceFileType) string {
	switch fileType {
	case xadmin.ResourceFileType_RESOURCE_FILE_TYPE_IMAGE:
		return "image"
	case xadmin.ResourceFileType_RESOURCE_FILE_TYPE_AUDIO:
		return "audio"
	case xadmin.ResourceFileType_RESOURCE_FILE_TYPE_VIDEO:
		return "video"
	case xadmin.ResourceFileType_RESOURCE_FILE_TYPE_DOCUMENT:
		return "document"
	case xadmin.ResourceFileType_RESOURCE_FILE_TYPE_ARCHIVE:
		return "archive"
	default:
		return ""
	}
}

func accessModeToDB(accessMode xadmin.ResourceFileAccessMode) string {
	switch accessMode {
	case xadmin.ResourceFileAccessMode_RESOURCE_FILE_ACCESS_MODE_DOWNLOAD:
		return "download"
	default:
		return "preview"
	}
}

func normalizeAccessMode(accessMode string) string {
	if strings.EqualFold(strings.TrimSpace(accessMode), "download") {
		return "download"
	}
	return "preview"
}
