package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"monorepo/internal/bootstrap"
	accounthandler "monorepo/internal/handler/account"
	"monorepo/internal/middleware"
	"monorepo/internal/model"
	xadmin "monorepo/proto/xadminpb"

	"github.com/gofiber/fiber/v2"
)

type mockAccountService struct{}

func (m *mockAccountService) GetPersonalSettings(ctx context.Context, uid int32) (*xadmin.AuthPersonalSettingsResp, error) {
	return &xadmin.AuthPersonalSettingsResp{
		LimitSingleLogin:             false,
		BackgroundImageUrl:           "",
		Locale:                       "en-US",
		GlobalBackgroundApplyEnabled: true,
		WarmTipIntervalMinutes:       1440,
	}, nil
}

func (m *mockAccountService) UpdatePersonalSettings(ctx context.Context, uid int32, sessionID string, req *xadmin.AuthUpdatePersonalSettingsReq, ip, userAgent, traceID string) (*xadmin.AuthPersonalSettingsResp, error) {
	return &xadmin.AuthPersonalSettingsResp{
		LimitSingleLogin:             req.GetLimitSingleLogin(),
		BackgroundImageUrl:           req.GetBackgroundImageUrl(),
		Locale:                       req.GetLocale(),
		GlobalBackgroundApplyEnabled: req.GetGlobalBackgroundApplyEnabled(),
		WarmTipIntervalMinutes:       req.GetWarmTipIntervalMinutes(),
	}, nil
}

func (m *mockAccountService) GetMyProfile(ctx context.Context, uid int32) (*xadmin.AuthMeProfileResp, error) {
	return &xadmin.AuthMeProfileResp{
		Uid:         uid,
		Username:    "admin",
		DisplayName: "系统管理员",
		Avatar:      "",
		Email:       "admin@example.com",
		Phone:       "13800000000",
		MenuRoutes:  []string{"/business/users", "/business/user-punishments"},
		MenuItems: []*xadmin.AuthMenuItem{
			{
				Id:            5,
				ParentId:      0,
				Name:          "业务管理",
				RoutePath:     "",
				PermissionKey: "business.root",
				Sort:          50,
				Icon:          "database",
				Children: []*xadmin.AuthMenuItem{
					{
						Id:            6,
						ParentId:      5,
						Name:          "用户列表（demo）",
						RoutePath:     "/business/users",
						PermissionKey: "business.users.view",
						Sort:          60,
						Icon:          "usergroup-add",
						Children:      []*xadmin.AuthMenuItem{},
					},
				},
			},
		},
		WarmTip: &xadmin.AuthWarmTip{
			Id:        3,
			TipType:   "positive",
			ContentZh: "今天也把重要的事推进一点",
			ContentEn: "Move one important thing forward today",
		},
	}, nil
}

func (m *mockAccountService) GetSystemSettings(ctx context.Context) (*xadmin.AuthSystemSettingsResp, error) {
	return &xadmin.AuthSystemSettingsResp{
		SiteName:                "XAdmin 管理后台",
		Locale:                  "zh-CN",
		Timezone:                "Asia/Shanghai",
		ServerTimezone:          "Asia/Shanghai",
		LoginLockThreshold:      5,
		PasswordMinLength:       8,
		SessionTimeoutMinutes:   30,
		PasswordPolicy:          []string{"uppercase", "number"},
		GlobalWatermarkEnabled:  false,
		GlobalWatermarkFontSize: 16,
	}, nil
}

func (m *mockAccountService) UpdateSystemSettings(ctx context.Context, req *xadmin.AuthUpdateSystemSettingsReq) (*xadmin.AuthSystemSettingsResp, error) {
	return &xadmin.AuthSystemSettingsResp{
		SiteName:                req.GetSiteName(),
		Locale:                  req.GetLocale(),
		Timezone:                req.GetTimezone(),
		ServerTimezone:          "Asia/Shanghai",
		LoginLockThreshold:      req.GetLoginLockThreshold(),
		PasswordMinLength:       req.GetPasswordMinLength(),
		SessionTimeoutMinutes:   req.GetSessionTimeoutMinutes(),
		PasswordPolicy:          req.GetPasswordPolicy(),
		GlobalWatermarkEnabled:  req.GetGlobalWatermarkEnabled(),
		GlobalWatermarkFontSize: req.GetGlobalWatermarkFontSize(),
	}, nil
}

func setupAccountApp() *fiber.App {
	app := fiber.New()
	v1 := app.Group("/v1")

	handler := accounthandler.NewHandlerWithService(&mockAccountService{})
	group := v1.Group("/account")

	authMW := func(c *fiber.Ctx) error {
		c.Locals(middleware.AuthCtx{}, &middleware.AuthCtx{
			User:      &model.AdminUser{UID: 10001},
			SessionID: "s-1",
		})
		return c.Next()
	}

	group.Get("/me/profile", authMW, handler.GetMyProfile)
	group.Get("/me/settings", authMW, handler.GetPersonalSettings)
	group.Post("/me/settings", authMW, handler.UpdatePersonalSettings)
	group.Get("/system/settings", authMW, handler.GetSystemSettings)
	group.Post("/system/settings", authMW, handler.UpdateSystemSettings)
	return app
}

func requestAccount(t *testing.T, app *fiber.App, method, path string, body any) *http.Response {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body failed: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func decodeAccountJSON(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()

	var got map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	return got
}

func TestGetAccountPersonalSettingsAPI(t *testing.T) {
	app := setupAccountApp()
	resp := requestAccount(t, app, http.MethodGet, "/v1/account/me/settings", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeAccountJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
	data := body["data"].(map[string]any)
	if data["locale"] != "en-US" {
		t.Fatalf("unexpected locale: %v", data["locale"])
	}
	if data["global_background_apply_enabled"] != true {
		t.Fatalf("unexpected global_background_apply_enabled: %v", data["global_background_apply_enabled"])
	}
	if data["warm_tip_interval_minutes"] != float64(1440) {
		t.Fatalf("unexpected warm_tip_interval_minutes: %v", data["warm_tip_interval_minutes"])
	}
}

func TestUpdateAccountPersonalSettingsAPI(t *testing.T) {
	app := setupAccountApp()
	resp := requestAccount(t, app, http.MethodPost, "/v1/account/me/settings", map[string]any{
		"limit_single_login":              true,
		"background_image_url":            "https://example.com/bg.avif",
		"locale":                          "en-US",
		"global_background_apply_enabled": false,
		"warm_tip_interval_minutes":       360,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeAccountJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
	data := body["data"].(map[string]any)
	if data["locale"] != "en-US" {
		t.Fatalf("unexpected locale: %v", data["locale"])
	}
	if data["global_background_apply_enabled"] != false {
		t.Fatalf("unexpected global_background_apply_enabled: %v", data["global_background_apply_enabled"])
	}
	if data["warm_tip_interval_minutes"] != float64(360) {
		t.Fatalf("unexpected warm_tip_interval_minutes: %v", data["warm_tip_interval_minutes"])
	}
}

func TestGetAccountSystemSettingsAPI(t *testing.T) {
	app := setupAccountApp()
	resp := requestAccount(t, app, http.MethodGet, "/v1/account/system/settings", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeAccountJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
	data := body["data"].(map[string]any)
	if data["server_timezone"] != "Asia/Shanghai" {
		t.Fatalf("unexpected server_timezone: %v", data["server_timezone"])
	}
}

func TestUpdateAccountSystemSettingsAPI(t *testing.T) {
	app := setupAccountApp()
	resp := requestAccount(t, app, http.MethodPost, "/v1/account/system/settings", map[string]any{
		"site_name":                  "XAdmin 管理后台",
		"locale":                     "zh-CN",
		"timezone":                   "Asia/Shanghai",
		"login_lock_threshold":       5,
		"password_min_length":        8,
		"session_timeout_minutes":    30,
		"password_policy":            []string{"uppercase", "number"},
		"alert_email_enabled":        true,
		"alert_receiver":             "ops@example.com",
		"webhook_url":                "https://example.com/webhook",
		"global_watermark_enabled":   false,
		"global_watermark_font_size": 16,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeAccountJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
	data := body["data"].(map[string]any)
	if data["server_timezone"] != "Asia/Shanghai" {
		t.Fatalf("unexpected server_timezone: %v", data["server_timezone"])
	}
}

func TestProgramTimezoneInitialization(t *testing.T) {
	original := time.Local
	defer func() {
		time.Local = original
	}()

	bootstrap.InitProgramTimezone("Asia/Shanghai")
	if got := time.Local.String(); got != "Asia/Shanghai" {
		t.Fatalf("unexpected local timezone: %v", got)
	}
}

func TestGetAccountMyProfileAPI(t *testing.T) {
	app := setupAccountApp()
	resp := requestAccount(t, app, http.MethodGet, "/v1/account/me/profile", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeAccountJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
	data := body["data"].(map[string]any)
	routes := data["menu_routes"].([]any)
	if len(routes) != 2 || routes[0] != "/business/users" {
		t.Fatalf("unexpected menu_routes: %v", routes)
	}
	items := data["menu_items"].([]any)
	if len(items) != 1 {
		t.Fatalf("unexpected menu_items length: %d", len(items))
	}
	root := items[0].(map[string]any)
	if root["name"] != "业务管理" {
		t.Fatalf("unexpected root menu name: %v", root["name"])
	}
	if root["icon"] != "database" {
		t.Fatalf("unexpected root menu icon: %v", root["icon"])
	}
	children := root["children"].([]any)
	if len(children) != 1 {
		t.Fatalf("unexpected children length: %d", len(children))
	}
	child := children[0].(map[string]any)
	if child["name"] != "用户列表（demo）" || child["route_path"] != "/business/users" {
		t.Fatalf("unexpected child menu: %v", child)
	}
	if child["icon"] != "usergroup-add" {
		t.Fatalf("unexpected child menu icon: %v", child["icon"])
	}
	warmTip := data["warm_tip"].(map[string]any)
	if warmTip["tip_type"] != "positive" || warmTip["content_zh"] != "今天也把重要的事推进一点" {
		t.Fatalf("unexpected warm_tip: %v", warmTip)
	}
}
