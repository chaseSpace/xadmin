package assets

import (
	"monorepo/internal/service/assert"
	"monorepo/pkg/xerr"
	"monorepo/pkg/xfiber"
	"monorepo/proto/xadminpb"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const uploadFileField = "file"

type Handler struct {
	svc *assert.Service
}

func NewHandler() *Handler {
	return &Handler{
		svc: assert.NewService(),
	}
}

func RegisterRoutes(prefix string, parentGroup fiber.Router, authMW fiber.Handler, optionalAuthMW fiber.Handler) {
	group := parentGroup.Group(prefix)
	handler := NewHandler()
	group.Post("/UploadFile", authMW, handler.UploadFile)
	group.Get("/GetFile", optionalAuthMW, handler.GetFile)
}

func (h *Handler) UploadFile(c *fiber.Ctx) error {
	scene, err := strconv.ParseInt(c.FormValue("scene"), 10, 32)
	if err != nil || scene == 0 {
		return xfiber.StdResponse(c, nil, xerr.NewWithDetail(xerr.CodeBadRequest, "scene must be a valid enum value"))
	}

	req := &xadmin.UploadFileReq{Scene: xadmin.UploadScene(scene)}
	if err = req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}

	rsp, err := h.svc.UploadFile(c, req, uploadFileField)
	return xfiber.StdResponse(c, rsp, err)
}

func (h *Handler) GetFile(c *fiber.Ctx) error {
	req := &xadmin.GetFileReq{
		FileUrl:    strings.TrimSpace(c.Query("file_url")),
		IsDownload: c.QueryBool("is_download", false),
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}

	if err := h.svc.GetFile(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	return nil
}
