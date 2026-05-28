package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authhandler "monorepo/internal/handler/auth"
	"monorepo/internal/middleware"
	"monorepo/internal/model"
	xadmin "monorepo/proto/xadminpb"

	"github.com/gofiber/fiber/v2"
)

type mockAuthService struct{}

func (m *mockAuthService) Login(ctx context.Context, req *xadmin.AuthLoginReq, ip, userAgent, traceID string) (*xadmin.AuthLoginResp, error) {
	return &xadmin.AuthLoginResp{
		AccessToken: "mock-token",
		ExpiresAt:   time.Now().Add(time.Hour).Format(time.RFC3339),
		Uid:         10001,
		Username:    req.GetUsername(),
		DisplayName: "管理员",
	}, nil
}

func (m *mockAuthService) Logout(ctx context.Context, uid int32, sessionID, ip, userAgent, traceID string) (*xadmin.AuthActionResp, error) {
	return &xadmin.AuthActionResp{Success: true, Action: "logout"}, nil
}

func (m *mockAuthService) LogoutOthers(ctx context.Context, uid int32, sessionID, ip, userAgent, traceID string) (*xadmin.AuthActionResp, error) {
	return &xadmin.AuthActionResp{Success: true, Action: "logout_others"}, nil
}

func (m *mockAuthService) ForceLogout(ctx context.Context, operatorUID int32, req *xadmin.AuthForceLogoutReq, ip, userAgent, traceID string) (*xadmin.AuthActionResp, error) {
	return &xadmin.AuthActionResp{Success: true, Action: "force_logout"}, nil
}

func (m *mockAuthService) Deactivate(ctx context.Context, operatorUID int32, req *xadmin.AuthDeactivateReq, ip, userAgent, traceID string) (*xadmin.AuthActionResp, error) {
	return &xadmin.AuthActionResp{Success: true, Action: "deactivate"}, nil
}

func (m *mockAuthService) ListSessions(ctx context.Context, uid int32, req *xadmin.AuthSessionsReq) (*xadmin.AuthSessionsResp, error) {
	items := []*xadmin.AuthSessionItem{
		{
			SessionId: "s-1",
			Status:    "active",
			LoginIp:   "127.0.0.1",
			UserAgent: "go-test",
			ExpiredAt: time.Now().Add(time.Hour).Format(time.RFC3339),
		},
		{
			SessionId: "s-2",
			Status:    "revoked",
			LoginIp:   "127.0.0.2",
			UserAgent: "go-test",
			ExpiredAt: time.Now().Add(time.Hour).Format(time.RFC3339),
		},
	}
	if req.GetStatus() == xadmin.AuthSessionStatus_AUTH_SESSION_STATUS_UNSPECIFIED {
		return &xadmin.AuthSessionsResp{Items: items}, nil
	}
	expectStatus := ""
	switch req.GetStatus() {
	case xadmin.AuthSessionStatus_AUTH_SESSION_STATUS_ACTIVE:
		expectStatus = "active"
	case xadmin.AuthSessionStatus_AUTH_SESSION_STATUS_REVOKED:
		expectStatus = "revoked"
	case xadmin.AuthSessionStatus_AUTH_SESSION_STATUS_EXPIRED:
		expectStatus = "expired"
	default:
		return &xadmin.AuthSessionsResp{Items: []*xadmin.AuthSessionItem{}}, nil
	}
	filtered := make([]*xadmin.AuthSessionItem, 0, len(items))
	for _, item := range items {
		if item.GetStatus() == expectStatus {
			filtered = append(filtered, item)
		}
	}
	return &xadmin.AuthSessionsResp{Items: filtered}, nil
}

func (m *mockAuthService) IsSessionActive(ctx context.Context, uid int32, sessionID, tokenHash string) (bool, error) {
	return true, nil
}

func setupAuthApp() *fiber.App {
	app := fiber.New()
	v1 := app.Group("/v1")

	handler := authhandler.NewHandlerWithService(&mockAuthService{})
	group := v1.Group("/auth")

	authMW := func(c *fiber.Ctx) error {
		c.Locals(middleware.AuthCtx{}, &middleware.AuthCtx{
			User:      &model.AdminUser{UID: 10001},
			SessionID: "s-1",
		})
		return c.Next()
	}

	group.Post("/login", handler.Login)
	group.Post("/logout", authMW, handler.Logout)
	group.Post("/logout_others", authMW, handler.LogoutOthers)
	group.Post("/force_logout", authMW, handler.ForceLogout)
	group.Post("/deactivate", authMW, handler.Deactivate)
	group.Get("/sessions", authMW, handler.Sessions)
	return app
}

func request(t *testing.T, app *fiber.App, method, path string, body any) *http.Response {
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

func decodeJSON(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()

	var got map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	return got
}

func TestAuthLoginAPI(t *testing.T) {
	app := setupAuthApp()
	resp := request(t, app, http.MethodPost, "/v1/auth/login", map[string]any{
		"username": "admin",
		"password": "secret",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestAuthLogoutAPI(t *testing.T) {
	app := setupAuthApp()
	resp := request(t, app, http.MethodPost, "/v1/auth/logout", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestAuthLogoutOthersAPI(t *testing.T) {
	app := setupAuthApp()
	resp := request(t, app, http.MethodPost, "/v1/auth/logout_others", map[string]any{})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestAuthForceLogoutAPI(t *testing.T) {
	app := setupAuthApp()
	resp := request(t, app, http.MethodPost, "/v1/auth/force_logout", map[string]any{
		"target_uid": 10002,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestAuthDeactivateAPI(t *testing.T) {
	app := setupAuthApp()
	resp := request(t, app, http.MethodPost, "/v1/auth/deactivate", map[string]any{
		"target_uid": 10002,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestAuthSessionsAPI(t *testing.T) {
	app := setupAuthApp()
	resp := request(t, app, http.MethodGet, "/v1/auth/sessions", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestAuthSessionsAPIWithStatusFilter(t *testing.T) {
	app := setupAuthApp()
	resp := request(t, app, http.MethodGet, "/v1/auth/sessions?status=active", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected data: %T", body["data"])
	}
	items, ok := data["items"].([]any)
	if !ok {
		t.Fatalf("unexpected items: %T", data["items"])
	}
	if len(items) != 1 {
		t.Fatalf("unexpected filtered length: %d", len(items))
	}
	first, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected item shape: %T", items[0])
	}
	if first["status"] != "active" {
		t.Fatalf("unexpected status: %v", first["status"])
	}
}

func TestAuthSessionsAPIWithInvalidStatusFilter(t *testing.T) {
	app := setupAuthApp()
	resp := request(t, app, http.MethodGet, "/v1/auth/sessions?status=unknown", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) == 200 {
		t.Fatalf("expected bad request code, got: %v", body["code"])
	}
}
