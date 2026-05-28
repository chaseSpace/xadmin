package auth

import (
	"strconv"
	"strings"

	"monorepo/internal/middleware"
	authsvc "monorepo/internal/service/auth"
	"monorepo/internal/support/auditlog"
	"monorepo/pkg/consts"
	"monorepo/pkg/xerr"
	"monorepo/pkg/xfiber"
	xadmin "monorepo/proto/xadminpb"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/encoding/protojson"
	proto2 "google.golang.org/protobuf/proto"
)

type Handler struct {
	svc authsvc.Service
}

func NewHandler() *Handler {
	return &Handler{svc: authsvc.NewService()}
}

func NewHandlerWithService(svc authsvc.Service) *Handler {
	return &Handler{svc: svc}
}

func RegisterRoutes(prefix string, parent fiber.Router, authMW fiber.Handler) {
	handler := NewHandler()
	group := parent.Group(prefix)

	group.Post("/login", handler.Login)
	group.Post("/logout", authMW, handler.Logout)
	group.Post("/logout_others", authMW, handler.LogoutOthers)
	group.Post("/force_logout", authMW, handler.ForceLogout)
	group.Post("/deactivate", authMW, handler.Deactivate)
	group.Get("/sessions", authMW, handler.Sessions)
}

func (h *Handler) Login(c *fiber.Ctx) error {
	req := &xadmin.AuthLoginReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}

	resp, err := h.svc.Login(
		c.UserContext(),
		req,
		c.IP(),
		c.Get("User-Agent"),
		traceIDFromCtx(c),
	)
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	uid := middleware.GetUID(c)
	sessionID := middleware.GetSessionID(c)
	resp, err := h.svc.Logout(c.UserContext(), uid, sessionID, c.IP(), c.Get("User-Agent"), traceIDFromCtx(c))
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) LogoutOthers(c *fiber.Ctx) error {
	uid := middleware.GetUID(c)
	sessionID := middleware.GetSessionID(c)
	resp, err := h.svc.LogoutOthers(c.UserContext(), uid, sessionID, c.IP(), c.Get("User-Agent"), traceIDFromCtx(c))
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) ForceLogout(c *fiber.Ctx) error {
	req := &xadmin.AuthForceLogoutReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.ForceLogout(c.UserContext(), middleware.GetUID(c), req, c.IP(), c.Get("User-Agent"), traceIDFromCtx(c))
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) Deactivate(c *fiber.Ctx) error {
	req := &xadmin.AuthDeactivateReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.Deactivate(c.UserContext(), middleware.GetUID(c), req, c.IP(), c.Get("User-Agent"), traceIDFromCtx(c))
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) Sessions(c *fiber.Ctx) error {
	req := &xadmin.AuthSessionsReq{PageSize: 20}
	if pageSizeRaw := strings.TrimSpace(c.Query("page_size")); pageSizeRaw != "" {
		pageSize, err := strconv.Atoi(pageSizeRaw)
		if err != nil {
			return xfiber.StdResponse(c, nil, err)
		}
		req.PageSize = int32(pageSize)
	}
	status, err := parseSessionStatusQuery(c.Query("status"))
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req.Status = status
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.ListSessions(c.UserContext(), middleware.GetUID(c), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_sessions",
			Result:    "success",
			TraceID:   traceIDFromCtx(c),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "status=" + req.GetStatus().String(),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func parseSessionStatusQuery(raw string) (xadmin.AuthSessionStatus, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return xadmin.AuthSessionStatus_AUTH_SESSION_STATUS_UNSPECIFIED, nil
	case consts.SessionStatusActive:
		return xadmin.AuthSessionStatus_AUTH_SESSION_STATUS_ACTIVE, nil
	case consts.SessionStatusRevoked:
		return xadmin.AuthSessionStatus_AUTH_SESSION_STATUS_REVOKED, nil
	case consts.SessionStatusExpired:
		return xadmin.AuthSessionStatus_AUTH_SESSION_STATUS_EXPIRED, nil
	default:
		return xadmin.AuthSessionStatus_AUTH_SESSION_STATUS_UNSPECIFIED, xerr.NewBiz(xerr.CodeBadRequest, "account.invalid_status")
	}
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
