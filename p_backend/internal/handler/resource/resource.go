package resource

import (
	"strconv"
	"strings"

	"monorepo/internal/middleware"
	resourcesvc "monorepo/internal/service/resource"
	"monorepo/internal/support/auditlog"
	"monorepo/pkg/xerr"
	"monorepo/pkg/xfiber"
	xadmin "monorepo/proto/xadminpb"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/encoding/protojson"
	proto2 "google.golang.org/protobuf/proto"
)

const uploadFileField = "file"

type Handler struct {
	svc resourcesvc.Service
}

func NewHandler() *Handler {
	return &Handler{svc: resourcesvc.NewService()}
}

func RegisterRoutes(prefix string, parent fiber.Router, authMW fiber.Handler) {
	handler := NewHandler()
	group := parent.Group(prefix, authMW)
	edit := middleware.RequirePermission

	group.Get("/files", handler.Files)
	group.Post("/files", edit("resource.files.upload"), handler.UploadFile)
	group.Put("/files/:id", edit("resource.files.upload"), handler.UpdateFile)
	group.Delete("/files/:id", edit("resource.files.delete"), handler.DeleteFile)
	group.Post("/files/:id/access", handler.AccessFile)
	group.Post("/files/check", handler.CheckFiles)
}

func (h *Handler) Files(c *fiber.Ctx) error {
	page, sort, err := xfiber.ParsePageSortQuery(c, 1, 10)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.ResourceFilesReq{
		Page:    page,
		Sort:    sort,
		Keyword: strings.TrimSpace(c.Query("keyword")),
	}
	req.FileType = parseResourceFileType(c.Query("file_type"))
	switch strings.ToLower(strings.TrimSpace(c.Query("exists"))) {
	case "true", "1":
		exists := true
		req.Exists = &exists
	case "false", "0":
		exists := false
		req.Exists = &exists
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.ListFiles(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_files",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) UploadFile(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile(uploadFileField)
	if err != nil {
		return xfiber.StdResponse(c, nil, xerr.NewWithDetail(xerr.CodeBadRequest, "missing upload file"))
	}
	resp, err := h.svc.UploadFile(
		c.UserContext(),
		middleware.GetUID(c),
		strings.TrimSpace(c.FormValue("name")),
		strings.TrimSpace(c.FormValue("remark")),
		c.FormValue("require_auth") == "true" || c.FormValue("require_auth") == "1",
		strings.ToLower(strings.TrimSpace(c.FormValue("access_mode"))),
		strings.ToLower(strings.TrimSpace(c.FormValue("file_type"))),
		fileHeader,
	)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "upload_file",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) UpdateFile(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.ResourceUpdateFileReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req.Id = id
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.UpdateFile(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "update_file",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) DeleteFile(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.ResourceFileActionReq{Id: id}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.DeleteFile(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "delete_file",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) AccessFile(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.ResourceFileActionReq{Id: id}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.AccessFile(c.UserContext(), req)
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) CheckFiles(c *fiber.Ctx) error {
	req := &xadmin.ResourceCheckFilesReq{FileType: parseResourceFileType(c.Query("file_type"))}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.CheckFiles(c.UserContext(), req)
	return xfiber.StdResponse(c, resp, err)
}

func parseResourceFileType(value string) xadmin.ResourceFileType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "image", "1":
		return xadmin.ResourceFileType_RESOURCE_FILE_TYPE_IMAGE
	case "audio", "2":
		return xadmin.ResourceFileType_RESOURCE_FILE_TYPE_AUDIO
	case "video", "3":
		return xadmin.ResourceFileType_RESOURCE_FILE_TYPE_VIDEO
	case "document", "4":
		return xadmin.ResourceFileType_RESOURCE_FILE_TYPE_DOCUMENT
	case "archive", "5":
		return xadmin.ResourceFileType_RESOURCE_FILE_TYPE_ARCHIVE
	default:
		return xadmin.ResourceFileType_RESOURCE_FILE_TYPE_UNSPECIFIED
	}
}

func parseProtoRequest(c *fiber.Ctx, req proto2.Message) error {
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(c.Body(), req); err != nil {
		return err
	}
	return nil
}
