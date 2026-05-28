package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"monorepo/internal/support/ipblacklist"

	"github.com/gofiber/fiber/v2"
)

type fakeIPBlacklistMatcher struct {
	blocked bool
	calls   int
}

func (f *fakeIPBlacklistMatcher) Match(_ string, _ time.Time) (ipblacklist.Match, bool) {
	f.calls++
	return ipblacklist.Match{ID: 1, IP: "127.0.0.1"}, f.blocked
}

func TestIPBlacklistBlocksActiveIP(t *testing.T) {
	matcher := &fakeIPBlacklistMatcher{blocked: true}
	app := fiber.New()
	app.Use(IPBlacklist(matcher))
	app.Get("/v1/account/profile", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/v1/account/profile", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected http status: %d", resp.StatusCode)
	}
	var body struct {
		Code int32 `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body failed: %v", err)
	}
	if body.Code != http.StatusForbidden {
		t.Fatalf("unexpected response code: %d", body.Code)
	}
	if matcher.calls != 1 {
		t.Fatalf("unexpected matcher calls: %d", matcher.calls)
	}
}

func TestIPBlacklistSkipsManagementRoutes(t *testing.T) {
	matcher := &fakeIPBlacklistMatcher{blocked: true}
	app := fiber.New()
	app.Use(IPBlacklist(matcher))
	app.Post("/v1/system/ip-blacklist/1/unblock", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodPost, "/v1/system/ip-blacklist/1/unblock", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("unexpected http status: %d", resp.StatusCode)
	}
	if matcher.calls != 0 {
		t.Fatalf("unexpected matcher calls: %d", matcher.calls)
	}
}
