package test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	systemhandler "monorepo/internal/handler/system"
	"monorepo/internal/middleware"
	"monorepo/internal/model"
	xadmin "monorepo/proto/xadminpb"

	"github.com/gofiber/fiber/v2"
)

type mockSystemService struct{}

func (m *mockSystemService) ListAuditLogs(ctx context.Context, req *xadmin.SystemAuditLogsReq) (*xadmin.SystemAuditLogsResp, error) {
	_ = ctx
	_ = req
	return &xadmin.SystemAuditLogsResp{
		Total: 1,
		Page:  req.GetPage(),
		Items: []*xadmin.SystemAuditLogItem{
			{
				Id:        1,
				Uid:       10001,
				Actor:     "系统管理员",
				Action:    "login_success",
				Result:    "success",
				TraceId:   "trace-1",
				SourceIp:  "127.0.0.1",
				UserAgent: "Mozilla/5.0",
				Detail:    "",
				CreatedAt: "2026-04-24 08:00:00",
			},
		},
	}, nil
}

func (m *mockSystemService) GetAuditLog(ctx context.Context, req *xadmin.SystemAuditLogDetailReq) (*xadmin.SystemAuditLogDetailResp, error) {
	_ = ctx
	return &xadmin.SystemAuditLogDetailResp{
		Id:        req.GetId(),
		Uid:       10001,
		Actor:     "系统管理员",
		Action:    "login_success",
		Result:    "success",
		TraceId:   "trace-1",
		SourceIp:  "127.0.0.1",
		UserAgent: "Mozilla/5.0",
		Detail:    "",
		CreatedAt: "2026-04-24 08:00:00",
	}, nil
}

func (m *mockSystemService) ExportAuditLogs(ctx context.Context, req *xadmin.SystemAuditLogsReq) ([]byte, error) {
	_ = ctx
	_ = req
	return []byte("日志ID,用户UID\n1,10001\n"), nil
}

func (m *mockSystemService) ApplyAuditLogRetention(ctx context.Context, req *xadmin.SystemAuditLogRetentionReq) (*xadmin.SystemAuditLogRetentionResp, error) {
	_ = ctx
	return &xadmin.SystemAuditLogRetentionResp{
		Success:      req.GetConfirm(),
		RetainDays:   req.GetRetainDays(),
		ExpiredCount: 12,
		ValidCount:   34,
		CutoffAt:     "2026-04-16 00:00:00",
	}, nil
}

func (m *mockSystemService) ListIPBlacklist(ctx context.Context, req *xadmin.SystemIPBlacklistReq) (*xadmin.SystemIPBlacklistResp, error) {
	_ = ctx
	return &xadmin.SystemIPBlacklistResp{
		Total: 1,
		Page:  req.GetPage(),
		Items: []*xadmin.SystemIPBlacklistItem{
			{Id: 1, Ip: "127.0.0.1", BanType: "temp", StartAt: "2026-04-28 00:00:00", EndAt: "2026-04-29 00:00:00", Reason: "测试", Creator: "10001", Status: "active", HitCount: 1, UpdatedAt: "2026-04-28 00:00:00"},
		},
	}, nil
}

func (m *mockSystemService) ListIPBlacklistCreators(ctx context.Context) (*xadmin.SystemIPBlacklistCreatorsResp, error) {
	_ = ctx
	return &xadmin.SystemIPBlacklistCreatorsResp{Creators: []string{"admin", "operator"}}, nil
}

func (m *mockSystemService) CreateIPBlacklist(ctx context.Context, req *xadmin.SystemCreateIPBlacklistReq) (*xadmin.SystemActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.SystemActionResp{Success: true, Action: "create_ip_blacklist"}, nil
}

func (m *mockSystemService) UpdateIPBlacklist(ctx context.Context, req *xadmin.SystemUpdateIPBlacklistReq) (*xadmin.SystemActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.SystemActionResp{Success: true, Action: "update_ip_blacklist"}, nil
}

func (m *mockSystemService) UnblockIPBlacklist(ctx context.Context, req *xadmin.SystemIPBlacklistActionReq) (*xadmin.SystemActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.SystemActionResp{Success: true, Action: "unblock_ip_blacklist"}, nil
}

func (m *mockSystemService) DeleteIPBlacklist(ctx context.Context, req *xadmin.SystemIPBlacklistActionReq) (*xadmin.SystemActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.SystemActionResp{Success: true, Action: "delete_ip_blacklist"}, nil
}

func (m *mockSystemService) BatchUnblockIPBlacklist(ctx context.Context, req *xadmin.SystemIPBlacklistBatchUnblockReq) (*xadmin.SystemActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.SystemActionResp{Success: true, Action: "batch_unblock_ip_blacklist"}, nil
}

func (m *mockSystemService) ImportIPBlacklist(ctx context.Context, req *xadmin.SystemIPBlacklistImportReq, creator string) (*xadmin.SystemActionResp, error) {
	_ = ctx
	_ = req
	_ = creator
	return &xadmin.SystemActionResp{Success: true, Action: "import_ip_blacklist"}, nil
}

func (m *mockSystemService) ListWarmTips(ctx context.Context, req *xadmin.SystemWarmTipsReq) (*xadmin.SystemWarmTipsResp, error) {
	_ = ctx
	return &xadmin.SystemWarmTipsResp{
		Total: 1,
		Page:  req.GetPage(),
		Items: []*xadmin.SystemWarmTipItem{
			{Id: 1, TipType: "rest", ContentZh: "喝水伸展，眼睛休息一下", ContentEn: "Drink water and rest your eyes", Sort: 10, Status: 1, UpdatedAt: "2026-05-12 14:00:00"},
		},
	}, nil
}

func (m *mockSystemService) CreateWarmTip(ctx context.Context, req *xadmin.SystemCreateWarmTipReq) (*xadmin.SystemActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.SystemActionResp{Success: true, Action: "create_warm_tip"}, nil
}

func (m *mockSystemService) UpdateWarmTip(ctx context.Context, req *xadmin.SystemUpdateWarmTipReq) (*xadmin.SystemActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.SystemActionResp{Success: true, Action: "update_warm_tip"}, nil
}

func (m *mockSystemService) DeleteWarmTip(ctx context.Context, req *xadmin.SystemWarmTipActionReq) (*xadmin.SystemActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.SystemActionResp{Success: true, Action: "delete_warm_tip"}, nil
}

func (m *mockSystemService) ListAlertBots(ctx context.Context, req *xadmin.SystemAlertBotListReq) (*xadmin.SystemAlertBotListResp, error) {
	_ = ctx
	return &xadmin.SystemAlertBotListResp{
		Total: 1,
		Page:  req.GetPage(),
		Items: []*xadmin.SystemAlertBotItem{
			{Id: 1, Name: "Test Bot", Username: "test_bot", Token: "tok", BotType: "telegram", Enabled: true, LinkedSceneKeys: []string{"test"}, CreatedAt: "2026-05-23 00:00:00", UpdatedAt: "2026-05-23 00:00:00"},
		},
	}, nil
}

func (m *mockSystemService) SaveAlertBot(ctx context.Context, req *xadmin.SystemAlertBotSaveReq) (*xadmin.SystemActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.SystemActionResp{Success: true, Action: "create_alert_bot"}, nil
}

func (m *mockSystemService) DeleteAlertBot(ctx context.Context, req *xadmin.SystemAlertBotActionReq) (*xadmin.SystemActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.SystemActionResp{Success: true, Action: "delete_alert_bot"}, nil
}

func (m *mockSystemService) ListAlertScenes(ctx context.Context, req *xadmin.SystemAlertSceneListReq) (*xadmin.SystemAlertSceneListResp, error) {
	_ = ctx
	return &xadmin.SystemAlertSceneListResp{Total: 0, Page: req.GetPage(), Items: nil}, nil
}

func (m *mockSystemService) SaveAlertScene(ctx context.Context, req *xadmin.SystemAlertSceneSaveReq) (*xadmin.SystemActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.SystemActionResp{Success: true, Action: "create_alert_scene"}, nil
}

func (m *mockSystemService) DeleteAlertScene(ctx context.Context, req *xadmin.SystemAlertSceneActionReq) (*xadmin.SystemActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.SystemActionResp{Success: true, Action: "delete_alert_scene"}, nil
}

func (m *mockSystemService) TestSendAlertScene(ctx context.Context, req *xadmin.SystemAlertSceneTestSendReq) (*xadmin.SystemActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.SystemActionResp{Success: true, Action: "test_send"}, nil
}

func (m *mockSystemService) ListAlertTemplates(ctx context.Context, req *xadmin.SystemAlertTemplateListReq) (*xadmin.SystemAlertTemplateListResp, error) {
	_ = ctx
	return &xadmin.SystemAlertTemplateListResp{Total: 0, Page: req.GetPage(), Items: nil}, nil
}

func (m *mockSystemService) SaveAlertTemplate(ctx context.Context, req *xadmin.SystemAlertTemplateSaveReq) (*xadmin.SystemActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.SystemActionResp{Success: true, Action: "create_alert_template"}, nil
}

func (m *mockSystemService) DeleteAlertTemplate(ctx context.Context, req *xadmin.SystemAlertTemplateActionReq) (*xadmin.SystemActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.SystemActionResp{Success: true, Action: "delete_alert_template"}, nil
}

func setupSystemApp() *fiber.App {
	app := fiber.New()
	v1 := app.Group("/v1")
	handler := systemhandler.NewHandlerWithService(&mockSystemService{})
	group := v1.Group("/system")
	authMW := func(c *fiber.Ctx) error {
		c.Locals(middleware.AuthCtx{}, &middleware.AuthCtx{
			User:      &model.AdminUser{UID: 10001},
			SessionID: "s-1",
		})
		return c.Next()
	}
	group.Get("/audit-logs", authMW, handler.AuditLogs)
	group.Get("/audit-logs/export", authMW, handler.ExportAuditLogs)
	group.Get("/audit-logs/:id", authMW, handler.AuditLog)
	group.Post("/audit-logs/retention", authMW, handler.ApplyAuditLogRetention)
	group.Get("/ip-blacklist", authMW, handler.IPBlacklist)
	group.Get("/ip-blacklist/creators", authMW, handler.IPBlacklistCreators)
	group.Post("/ip-blacklist", authMW, handler.CreateIPBlacklist)
	group.Put("/ip-blacklist/:id", authMW, handler.UpdateIPBlacklist)
	group.Post("/ip-blacklist/:id/unblock", authMW, handler.UnblockIPBlacklist)
	group.Delete("/ip-blacklist/:id", authMW, handler.DeleteIPBlacklist)
	group.Post("/ip-blacklist/unblock-batch", authMW, handler.BatchUnblockIPBlacklist)
	group.Post("/ip-blacklist/import", authMW, handler.ImportIPBlacklist)
	group.Get("/warm-tips", authMW, handler.WarmTips)
	group.Post("/warm-tips", authMW, handler.CreateWarmTip)
	group.Put("/warm-tips/:id", authMW, handler.UpdateWarmTip)
	group.Delete("/warm-tips/:id", authMW, handler.DeleteWarmTip)
	group.Get("/alert-bots", authMW, handler.AlertBots)
	group.Post("/alert-bots", authMW, handler.SaveAlertBot)
	group.Delete("/alert-bots/:id", authMW, handler.DeleteAlertBot)
	group.Get("/alert-scenes", authMW, handler.AlertScenes)
	group.Post("/alert-scenes", authMW, handler.SaveAlertScene)
	group.Delete("/alert-scenes/:id", authMW, handler.DeleteAlertScene)
	group.Post("/alert-scenes/:id/test-send", authMW, handler.TestSendAlertScene)
	group.Get("/alert-templates", authMW, handler.AlertTemplates)
	group.Post("/alert-templates", authMW, handler.SaveAlertTemplate)
	group.Delete("/alert-templates/:id", authMW, handler.DeleteAlertTemplate)
	return app
}

func TestSystemAuditLogsAPI(t *testing.T) {
	app := setupSystemApp()
	resp := requestAccount(t, app, http.MethodGet, "/v1/system/audit-logs?page_no=1&page_size=10", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeAccountJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestSystemAuditLogDetailAPI(t *testing.T) {
	app := setupSystemApp()
	resp := requestAccount(t, app, http.MethodGet, "/v1/system/audit-logs/1", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeAccountJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestSystemAuditLogsExportAPI(t *testing.T) {
	app := setupSystemApp()
	resp := requestAccount(t, app, http.MethodGet, "/v1/system/audit-logs/export", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/csv") {
		t.Fatalf("unexpected content type: %s", ct)
	}
}

func TestSystemAuditLogsRetentionAPI(t *testing.T) {
	app := setupSystemApp()
	resp := requestAccount(t, app, http.MethodPost, "/v1/system/audit-logs/retention", map[string]any{
		"retain_days": 30,
		"confirm":     false,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeAccountJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestSystemIPBlacklistAPI(t *testing.T) {
	app := setupSystemApp()
	cases := []struct {
		method string
		path   string
		body   map[string]any
	}{
		{http.MethodGet, "/v1/system/ip-blacklist?page_no=1&page_size=10", nil},
		{http.MethodGet, "/v1/system/ip-blacklist/creators", nil},
		{http.MethodPost, "/v1/system/ip-blacklist", map[string]any{"ip": "1.1.1.1", "ban_type": 1, "reason": "test"}},
		{http.MethodPut, "/v1/system/ip-blacklist/1", map[string]any{"ban_type": 2, "reason": "test-update"}},
		{http.MethodPost, "/v1/system/ip-blacklist/1/unblock", map[string]any{}},
		{http.MethodDelete, "/v1/system/ip-blacklist/1", nil},
		{http.MethodPost, "/v1/system/ip-blacklist/unblock-batch", map[string]any{"ids": []int64{1, 2}}},
		{http.MethodPost, "/v1/system/ip-blacklist/import", map[string]any{"ips": []string{"2.2.2.2"}, "ban_type": 1, "duration_hours": 24}},
		{http.MethodPost, "/v1/system/ip-blacklist/import", map[string]any{"ips": []string{"2.2.2.3"}, "ban_type": 1, "end_at": "2026-05-01 18:30:00"}},
	}
	for _, tc := range cases {
		resp := requestAccount(t, app, tc.method, tc.path, tc.body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s %s status=%d", tc.method, tc.path, resp.StatusCode)
		}
	}
}

func TestSystemIPBlacklistStatusValuesAPI(t *testing.T) {
	app := setupSystemApp()
	resp := requestAccount(t, app, http.MethodGet, "/v1/system/ip-blacklist?page_no=1&page_size=10", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeAccountJSON(t, resp)
	data := body["data"].(map[string]any)
	items := data["items"].([]any)
	if len(items) == 0 {
		t.Fatal("expected ip blacklist item")
	}
	status := items[0].(map[string]any)["status"]
	switch status {
	case "active", "expired", "manual_inactive":
	default:
		t.Fatalf("unexpected ip blacklist status: %v", status)
	}
}

func TestSystemWarmTipsAPI(t *testing.T) {
	app := setupSystemApp()
	cases := []struct {
		method string
		path   string
		body   map[string]any
	}{
		{http.MethodGet, "/v1/system/warm-tips?page_no=1&page_size=10", nil},
		{http.MethodGet, "/v1/system/warm-tips?page_no=1&page_size=10&order_field=sort&order_type=asc", nil},
		{http.MethodPost, "/v1/system/warm-tips", map[string]any{"tip_type": "rest", "content_zh": "喝水伸展，眼睛休息一下", "content_en": "Drink water and rest your eyes", "sort": 10, "status": 1}},
		{http.MethodPut, "/v1/system/warm-tips/1", map[string]any{"tip_type": "positive", "content_zh": "今天也把重要的事推进一点", "content_en": "Move one important thing forward today", "sort": 20, "status": 1}},
		{http.MethodDelete, "/v1/system/warm-tips/1", nil},
	}
	for _, tc := range cases {
		resp := requestAccount(t, app, tc.method, tc.path, tc.body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s %s status=%d", tc.method, tc.path, resp.StatusCode)
		}
	}
}

func TestSystemAlertBotAPI(t *testing.T) {
	app := setupSystemApp()
	cases := []struct {
		method string
		path   string
		body   map[string]any
	}{
		{http.MethodGet, "/v1/system/alert-bots?page_no=1&page_size=10", nil},
		{http.MethodPost, "/v1/system/alert-bots", map[string]any{"name": "Bot1", "token": "tok1", "bot_type": "telegram", "enabled": true, "scene_keys": []string{"test"}}},
		{http.MethodDelete, "/v1/system/alert-bots/1", nil},
	}
	for _, tc := range cases {
		resp := requestAccount(t, app, tc.method, tc.path, tc.body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s %s status=%d", tc.method, tc.path, resp.StatusCode)
		}
	}
}
