package system

import (
	"strconv"
	"strings"

	"monorepo/internal/middleware"
	systemsvc "monorepo/internal/service/system"
	"monorepo/internal/support/auditlog"
	"monorepo/pkg/xfiber"
	xadmin "monorepo/proto/xadminpb"
	commpb "monorepo/proto/xadminpb/commpb"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/encoding/protojson"
	proto2 "google.golang.org/protobuf/proto"
)

type Handler struct {
	svc systemsvc.Service
}

func NewHandler() *Handler {
	return &Handler{svc: systemsvc.NewService()}
}

func NewHandlerWithService(svc systemsvc.Service) *Handler {
	return &Handler{svc: svc}
}

func RegisterRoutes(prefix string, parent fiber.Router, authMW fiber.Handler) {
	handler := NewHandler()
	group := parent.Group(prefix, authMW)

	edit := middleware.RequirePermission

	group.Get("/audit-logs", handler.AuditLogs)
	group.Get("/audit-logs/export", handler.ExportAuditLogs)
	group.Get("/audit-logs/:id", handler.AuditLog)
	group.Post("/audit-logs/retention", edit("system.settings.edit"), handler.ApplyAuditLogRetention)
	group.Get("/ip-blacklist", handler.IPBlacklist)
	group.Get("/ip-blacklist/creators", handler.IPBlacklistCreators)
	group.Post("/ip-blacklist", edit("system.ip_blacklist.edit"), handler.CreateIPBlacklist)
	group.Put("/ip-blacklist/:id", edit("system.ip_blacklist.edit"), handler.UpdateIPBlacklist)
	group.Post("/ip-blacklist/:id/unblock", edit("system.ip_blacklist.edit"), handler.UnblockIPBlacklist)
	group.Delete("/ip-blacklist/:id", edit("system.ip_blacklist.delete"), handler.DeleteIPBlacklist)
	group.Post("/ip-blacklist/unblock-batch", edit("system.ip_blacklist.edit"), handler.BatchUnblockIPBlacklist)
	group.Post("/ip-blacklist/import", edit("system.ip_blacklist.edit"), handler.ImportIPBlacklist)
	group.Get("/warm-tips", handler.WarmTips)
	group.Post("/warm-tips", edit("system.warm_tips.edit"), handler.CreateWarmTip)
	group.Put("/warm-tips/:id", edit("system.warm_tips.edit"), handler.UpdateWarmTip)
	group.Delete("/warm-tips/:id", edit("system.warm_tips.delete"), handler.DeleteWarmTip)
	group.Get("/alert-bots", handler.AlertBots)
	group.Post("/alert-bots", edit("system.alert_bots.edit"), handler.SaveAlertBot)
	group.Delete("/alert-bots/:id", edit("system.alert_bots.delete"), handler.DeleteAlertBot)
	group.Get("/alert-scenes", handler.AlertScenes)
	group.Post("/alert-scenes", edit("system.alert_bots.edit"), handler.SaveAlertScene)
	group.Delete("/alert-scenes/:id", edit("system.alert_bots.delete"), handler.DeleteAlertScene)
	group.Post("/alert-scenes/:id/test-send", edit("system.alert_bots.edit"), handler.TestSendAlertScene)
	group.Get("/alert-templates", handler.AlertTemplates)
	group.Post("/alert-templates", edit("system.alert_bots.edit"), handler.SaveAlertTemplate)
	group.Delete("/alert-templates/:id", edit("system.alert_bots.delete"), handler.DeleteAlertTemplate)
}

func (h *Handler) AuditLogs(c *fiber.Ctx) error {
	req, err := parseAuditLogsReq(c, false)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.ListAuditLogs(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_audit_logs",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) ExportAuditLogs(c *fiber.Ctx) error {
	req, err := parseAuditLogsReq(c, true)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	data, err := h.svc.ExportAuditLogs(c.UserContext(), req)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	_ = auditlog.Log(c.UserContext(), auditlog.Meta{
		UID:       middleware.GetUID(c),
		Action:    "export_audit_logs",
		Result:    "success",
		TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
		SourceIP:  c.IP(),
		UserAgent: c.Get("User-Agent"),
	})
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=system_audit_logs.csv")
	return c.Status(fiber.StatusOK).Send(data)
}

func (h *Handler) AuditLog(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.SystemAuditLogDetailReq{Id: id}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.GetAuditLog(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_audit_log_detail",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) ApplyAuditLogRetention(c *fiber.Ctx) error {
	req := &xadmin.SystemAuditLogRetentionReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.ApplyAuditLogRetention(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "apply_audit_log_retention",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) IPBlacklist(c *fiber.Ctx) error {
	page, sort, err := xfiber.ParsePageSortQuery(c, 1, 10)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.SystemIPBlacklistReq{Page: page, Sort: sort}
	req.Keyword = strings.TrimSpace(c.Query("keyword"))
	req.Creator = strings.TrimSpace(c.Query("creator"))
	switch strings.ToLower(strings.TrimSpace(c.Query("status"))) {
	case "active", "1":
		req.Status = xadmin.SystemIPBanStatusFilter_SYSTEM_IP_BAN_STATUS_FILTER_ACTIVE
	case "inactive", "0":
		req.Status = xadmin.SystemIPBanStatusFilter_SYSTEM_IP_BAN_STATUS_FILTER_INACTIVE
	}
	switch strings.ToLower(strings.TrimSpace(c.Query("ban_type"))) {
	case "temp", "1":
		req.BanType = xadmin.SystemIPBanType_SYSTEM_IP_BAN_TYPE_TEMP
	case "permanent", "2":
		req.BanType = xadmin.SystemIPBanType_SYSTEM_IP_BAN_TYPE_PERMANENT
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.ListIPBlacklist(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_ip_blacklist",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) IPBlacklistCreators(c *fiber.Ctx) error {
	resp, err := h.svc.ListIPBlacklistCreators(c.UserContext())
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) CreateIPBlacklist(c *fiber.Ctx) error {
	req := &xadmin.SystemCreateIPBlacklistReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	if strings.TrimSpace(req.GetCreator()) == "" {
		req.Creator = strconv.FormatInt(int64(middleware.GetUID(c)), 10)
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.CreateIPBlacklist(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "create_ip_blacklist",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "ip=" + req.GetIp(),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) UpdateIPBlacklist(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.SystemUpdateIPBlacklistReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req.Id = id
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.UpdateIPBlacklist(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "update_ip_blacklist",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) UnblockIPBlacklist(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.SystemIPBlacklistActionReq{Id: id}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.UnblockIPBlacklist(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "unblock_ip_blacklist",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) DeleteIPBlacklist(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.SystemIPBlacklistActionReq{Id: id}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.DeleteIPBlacklist(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "delete_ip_blacklist",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) BatchUnblockIPBlacklist(c *fiber.Ctx) error {
	req := &xadmin.SystemIPBlacklistBatchUnblockReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.BatchUnblockIPBlacklist(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "batch_unblock_ip_blacklist",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) ImportIPBlacklist(c *fiber.Ctx) error {
	req := &xadmin.SystemIPBlacklistImportReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.ImportIPBlacklist(c.UserContext(), req, strconv.FormatInt(int64(middleware.GetUID(c)), 10))
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "import_ip_blacklist",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) WarmTips(c *fiber.Ctx) error {
	page, sort, err := xfiber.ParsePageSortQuery(c, 1, 10)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.SystemWarmTipsReq{Page: page, Sort: sort}
	req.Keyword = strings.TrimSpace(c.Query("keyword"))
	req.TipType = strings.TrimSpace(c.Query("tip_type"))
	switch strings.ToLower(strings.TrimSpace(c.Query("status"))) {
	case "enabled", "1":
		req.Status = xadmin.SystemWarmTipStatusFilter_SYSTEM_WARM_TIP_STATUS_FILTER_ENABLED
	case "disabled", "0":
		req.Status = xadmin.SystemWarmTipStatusFilter_SYSTEM_WARM_TIP_STATUS_FILTER_DISABLED
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.ListWarmTips(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_warm_tips",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) CreateWarmTip(c *fiber.Ctx) error {
	req := &xadmin.SystemCreateWarmTipReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.CreateWarmTip(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "create_warm_tip",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) UpdateWarmTip(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.SystemUpdateWarmTipReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req.Id = id
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.UpdateWarmTip(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "update_warm_tip",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) DeleteWarmTip(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.SystemWarmTipActionReq{Id: id}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.DeleteWarmTip(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "delete_warm_tip",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func parseAuditLogsReq(c *fiber.Ctx, isDownload bool) (*xadmin.SystemAuditLogsReq, error) {
	page, sort, err := xfiber.ParsePageSortQuery(c, 1, 10)
	if err != nil {
		return nil, err
	}
	page.IsDownload = isDownload
	req := &xadmin.SystemAuditLogsReq{
		Page: page,
		Sort: sort,
		CreatedAt: &commpb.TimeRange{
			StartDt: strings.TrimSpace(c.Query("created_from")),
			EndDt:   strings.TrimSpace(c.Query("created_to")),
		},
	}
	req.Actor = strings.TrimSpace(c.Query("actor"))
	req.Action = strings.TrimSpace(c.Query("action"))
	req.TraceId = strings.TrimSpace(c.Query("trace_id"))
	req.RequestId = strings.TrimSpace(c.Query("request_id"))
	req.SourceIp = strings.TrimSpace(c.Query("source_ip"))
	req.Keyword = strings.TrimSpace(c.Query("keyword"))
	switch strings.ToLower(strings.TrimSpace(c.Query("result"))) {
	case "success", "1":
		req.Result = xadmin.SystemAuditResultFilter_SYSTEM_AUDIT_RESULT_FILTER_SUCCESS
	case "failed", "0":
		req.Result = xadmin.SystemAuditResultFilter_SYSTEM_AUDIT_RESULT_FILTER_FAILED
	}
	if req.CreatedAt.GetStartDt() == "" && req.CreatedAt.GetEndDt() == "" {
		req.CreatedAt = nil
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return req, nil
}

func parseProtoRequest(c *fiber.Ctx, req proto2.Message) error {
	body := c.Body()
	if len(body) == 0 {
		return nil
	}
	return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(body, req)
}

func (h *Handler) AlertBots(c *fiber.Ctx) error {
	page, sort, err := xfiber.ParsePageSortQuery(c, 1, 10)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.SystemAlertBotListReq{Page: page, Sort: sort}
	req.Keyword = strings.TrimSpace(c.Query("keyword"))
	req.BotType = strings.TrimSpace(c.Query("bot_type"))
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.ListAlertBots(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_alert_bots",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) SaveAlertBot(c *fiber.Ctx) error {
	req := &xadmin.SystemAlertBotSaveReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.SaveAlertBot(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "save_alert_bot",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) DeleteAlertBot(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.SystemAlertBotActionReq{Id: id}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.DeleteAlertBot(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "delete_alert_bot",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) AlertScenes(c *fiber.Ctx) error {
	page, sort, err := xfiber.ParsePageSortQuery(c, 1, 10)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.SystemAlertSceneListReq{Page: page, Sort: sort}
	req.Keyword = strings.TrimSpace(c.Query("keyword"))
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.ListAlertScenes(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_alert_scenes",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) SaveAlertScene(c *fiber.Ctx) error {
	req := &xadmin.SystemAlertSceneSaveReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.SaveAlertScene(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "save_alert_scene",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) DeleteAlertScene(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.SystemAlertSceneActionReq{Id: id}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.DeleteAlertScene(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "delete_alert_scene",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) TestSendAlertScene(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.SystemAlertSceneTestSendReq{Id: id}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req.Id = id
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.TestSendAlertScene(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "test_send_alert_scene",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) AlertTemplates(c *fiber.Ctx) error {
	page, sort, err := xfiber.ParsePageSortQuery(c, 1, 10)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.SystemAlertTemplateListReq{Page: page, Sort: sort}
	req.Keyword = strings.TrimSpace(c.Query("keyword"))
	req.BotType = strings.TrimSpace(c.Query("bot_type"))
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.ListAlertTemplates(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_alert_templates",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) SaveAlertTemplate(c *fiber.Ctx) error {
	req := &xadmin.SystemAlertTemplateSaveReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.SaveAlertTemplate(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "save_alert_template",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) DeleteAlertTemplate(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.SystemAlertTemplateActionReq{Id: id}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.DeleteAlertTemplate(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "delete_alert_template",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}
