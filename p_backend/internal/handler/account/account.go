package account

import (
	"strings"

	"monorepo/internal/middleware"
	accountsvc "monorepo/internal/service/account"
	"monorepo/internal/support/auditlog"
	"monorepo/pkg/xfiber"
	xadmin "monorepo/proto/xadminpb"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/encoding/protojson"
	proto2 "google.golang.org/protobuf/proto"
)

type Handler struct {
	svc accountsvc.Service
}

func NewHandler() *Handler {
	return &Handler{svc: accountsvc.NewService()}
}

func NewHandlerWithService(svc accountsvc.Service) *Handler {
	return &Handler{svc: svc}
}

func RegisterRoutes(prefix string, parent fiber.Router, authMW fiber.Handler) {
	handler := NewHandler()
	group := parent.Group(prefix, authMW)

	group.Get("/me/profile", handler.GetMyProfile)
	group.Get("/me/settings", handler.GetPersonalSettings)
	group.Post("/me/settings", handler.UpdatePersonalSettings)
	group.Get("/system/settings", handler.GetSystemSettings)
	group.Post("/system/settings", handler.UpdateSystemSettings)
}

func (h *Handler) GetMyProfile(c *fiber.Ctx) error {
	resp, err := h.svc.GetMyProfile(c.UserContext(), middleware.GetUID(c))
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_my_profile",
			Result:    "success",
			TraceID:   traceIDFromCtx(c),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) GetPersonalSettings(c *fiber.Ctx) error {
	resp, err := h.svc.GetPersonalSettings(c.UserContext(), middleware.GetUID(c))
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_personal_settings",
			Result:    "success",
			TraceID:   traceIDFromCtx(c),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) UpdatePersonalSettings(c *fiber.Ctx) error {
	req := &xadmin.AuthUpdatePersonalSettingsReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.UpdatePersonalSettings(
		c.UserContext(),
		middleware.GetUID(c),
		middleware.GetSessionID(c),
		req,
		c.IP(),
		c.Get("User-Agent"),
		traceIDFromCtx(c),
	)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "update_personal_settings",
			Result:    "success",
			TraceID:   traceIDFromCtx(c),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) GetSystemSettings(c *fiber.Ctx) error {
	resp, err := h.svc.GetSystemSettings(c.UserContext())
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_system_settings",
			Result:    "success",
			TraceID:   traceIDFromCtx(c),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) UpdateSystemSettings(c *fiber.Ctx) error {
	req := &xadmin.AuthUpdateSystemSettingsReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.UpdateSystemSettings(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "update_system_settings",
			Result:    "success",
			TraceID:   traceIDFromCtx(c),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func traceIDFromCtx(c *fiber.Ctx) string {
	return strings.TrimSpace(c.Get("X-Trace-ID"))
}

func parseProtoRequest(c *fiber.Ctx, req proto2.Message) error {
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(c.Body(), req); err != nil {
		return err
	}
	if v, ok := req.(interface{ Validate() error }); ok {
		return v.Validate()
	}
	return nil
}
