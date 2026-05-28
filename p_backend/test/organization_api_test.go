package test

import (
	"context"
	"net/http"
	"testing"

	organizationhandler "monorepo/internal/handler/organization"
	"monorepo/internal/middleware"
	"monorepo/internal/model"
	"monorepo/pkg/xerr"
	xadmin "monorepo/proto/xadminpb"

	"github.com/gofiber/fiber/v2"
)

type mockOrganizationService struct{}

func (m *mockOrganizationService) GetDepartmentsTree(ctx context.Context) (*xadmin.OrganizationDepartmentsTreeResp, error) {
	_ = ctx
	return &xadmin.OrganizationDepartmentsTreeResp{
		Items: []*xadmin.OrganizationDepartmentItem{
			{
				Id:          1,
				ParentId:    0,
				Name:        "集团总部",
				Code:        "HQ",
				Status:      "enabled",
				MemberCount: 0,
			},
		},
	}, nil
}

func (m *mockOrganizationService) GetDepartment(ctx context.Context, req *xadmin.OrganizationDepartmentDetailReq) (*xadmin.OrganizationDepartmentItem, error) {
	_ = ctx
	return &xadmin.OrganizationDepartmentItem{
		Id:          req.GetId(),
		ParentId:    0,
		Name:        "集团总部",
		Code:        "HQ",
		Status:      "enabled",
		MemberCount: 0,
	}, nil
}

func (m *mockOrganizationService) CreateDepartment(ctx context.Context, req *xadmin.OrganizationCreateDepartmentReq) (*xadmin.OrganizationActionResp, error) {
	_ = ctx
	if req.GetParentId() == 4 {
		return nil, xerr.NewWithDetail(xerr.CodeBadRequest, "最多只支持4级部门")
	}
	return &xadmin.OrganizationActionResp{Success: true, Action: "create_department"}, nil
}

func (m *mockOrganizationService) UpdateDepartment(ctx context.Context, req *xadmin.OrganizationUpdateDepartmentReq) (*xadmin.OrganizationActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.OrganizationActionResp{Success: true, Action: "update_department"}, nil
}

func (m *mockOrganizationService) UpdateDepartmentStatus(ctx context.Context, req *xadmin.OrganizationUpdateDepartmentStatusReq) (*xadmin.OrganizationActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.OrganizationActionResp{Success: true, Action: "update_department_status"}, nil
}

func (m *mockOrganizationService) DeleteDepartment(ctx context.Context, req *xadmin.OrganizationDeleteDepartmentReq) (*xadmin.OrganizationActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.OrganizationActionResp{Success: true, Action: "delete_department"}, nil
}

func (m *mockOrganizationService) ListPositions(ctx context.Context, req *xadmin.OrganizationPositionsReq) (*xadmin.OrganizationPositionsResp, error) {
	_ = ctx
	_ = req
	return &xadmin.OrganizationPositionsResp{
		Total: 1,
		Items: []*xadmin.OrganizationPositionItem{
			{
				Id:             1,
				Name:           "前端工程师",
				Code:           "POS-FE",
				DepartmentId:   1,
				DepartmentName: "集团总部",
				Level:          "P5",
				Hc:             5,
				Staffed:        4,
				Status:         "enabled",
			},
		},
	}, nil
}

func (m *mockOrganizationService) GetPosition(ctx context.Context, req *xadmin.OrganizationPositionDetailReq) (*xadmin.OrganizationPositionItem, error) {
	_ = ctx
	return &xadmin.OrganizationPositionItem{
		Id:             req.GetId(),
		Name:           "前端工程师",
		Code:           "POS-FE",
		DepartmentId:   1,
		DepartmentName: "集团总部",
		Level:          "P5",
		Hc:             5,
		Staffed:        4,
		Status:         "enabled",
	}, nil
}

func (m *mockOrganizationService) CreatePosition(ctx context.Context, req *xadmin.OrganizationCreatePositionReq) (*xadmin.OrganizationActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.OrganizationActionResp{Success: true, Action: "create_position"}, nil
}

func (m *mockOrganizationService) UpdatePosition(ctx context.Context, req *xadmin.OrganizationUpdatePositionReq) (*xadmin.OrganizationActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.OrganizationActionResp{Success: true, Action: "update_position"}, nil
}

func (m *mockOrganizationService) UpdatePositionStatus(ctx context.Context, req *xadmin.OrganizationUpdatePositionStatusReq) (*xadmin.OrganizationActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.OrganizationActionResp{Success: true, Action: "update_position_status"}, nil
}

func (m *mockOrganizationService) DeletePosition(ctx context.Context, req *xadmin.OrganizationDeletePositionReq) (*xadmin.OrganizationActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.OrganizationActionResp{Success: true, Action: "delete_position"}, nil
}

func (m *mockOrganizationService) ListUsers(ctx context.Context, req *xadmin.OrganizationUsersReq) (*xadmin.OrganizationUsersResp, error) {
	_ = ctx
	_ = req
	return &xadmin.OrganizationUsersResp{
		Total: 1,
		Items: []*xadmin.OrganizationUserItem{
			{
				Uid:                10001,
				Username:           "admin",
				DisplayName:        "系统管理员",
				AccountStatus:      "active",
				OnlineStatus:       "online",
				ActiveSessionCount: 2,
				LastLoginIp:        "127.0.0.1",
				LastLoginAt:        "2026-04-20T12:00:00Z",
			},
		},
	}, nil
}

func (m *mockOrganizationService) ListUserSessions(ctx context.Context, req *xadmin.OrganizationUserSessionsReq) (*xadmin.OrganizationUserSessionsResp, error) {
	_ = ctx
	_ = req
	return &xadmin.OrganizationUserSessionsResp{
		Items: []*xadmin.AuthSessionItem{
			{
				SessionId: "s-1",
				Status:    "active",
				LoginIp:   "127.0.0.1",
			},
		},
	}, nil
}

func (m *mockOrganizationService) CreateUser(ctx context.Context, req *xadmin.OrganizationCreateUserReq) (*xadmin.OrganizationActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.OrganizationActionResp{Success: true, Action: "create_user"}, nil
}

func (m *mockOrganizationService) UpdateUser(ctx context.Context, req *xadmin.OrganizationUpdateUserReq) (*xadmin.OrganizationActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.OrganizationActionResp{Success: true, Action: "update_user"}, nil
}

func (m *mockOrganizationService) BatchTransferUsers(ctx context.Context, req *xadmin.OrganizationBatchTransferUsersReq) (*xadmin.OrganizationActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.OrganizationActionResp{Success: true, Action: "batch_transfer_users"}, nil
}

func (m *mockOrganizationService) DeleteUser(ctx context.Context, req *xadmin.OrganizationDeleteUserReq) (*xadmin.OrganizationActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.OrganizationActionResp{Success: true, Action: "delete_user"}, nil
}

func (m *mockOrganizationService) ResetPassword(ctx context.Context, req *xadmin.OrganizationResetPasswordReq) (*xadmin.OrganizationActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.OrganizationActionResp{Success: true, Action: "reset_password"}, nil
}

func (m *mockOrganizationService) ImportUsers(ctx context.Context, req *xadmin.OrganizationImportUsersReq) (*xadmin.OrganizationActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.OrganizationActionResp{Success: true, Action: "import_users"}, nil
}

func (m *mockOrganizationService) ExportUsers(ctx context.Context, req *xadmin.OrganizationUsersReq) ([]byte, error) {
	_ = ctx
	_ = req
	return []byte("uid,username\n10001,admin\n"), nil
}

func setupOrganizationApp() *fiber.App {
	app := fiber.New()
	v1 := app.Group("/v1")

	handler := organizationhandler.NewHandlerWithService(&mockOrganizationService{})
	group := v1.Group("/organization")
	authMW := func(c *fiber.Ctx) error {
		c.Locals(middleware.AuthCtx{}, &middleware.AuthCtx{
			User:      &model.AdminUser{UID: 10001},
			SessionID: "s-1",
		})
		return c.Next()
	}

	group.Get("/departments/tree", authMW, handler.DepartmentsTree)
	group.Get("/departments/:id", authMW, handler.Department)
	group.Post("/departments", authMW, handler.CreateDepartment)
	group.Put("/departments/:id", authMW, handler.UpdateDepartment)
	group.Post("/departments/:id/status", authMW, handler.UpdateDepartmentStatus)
	group.Delete("/departments/:id", authMW, handler.DeleteDepartment)
	group.Get("/positions", authMW, handler.Positions)
	group.Get("/positions/:id", authMW, handler.Position)
	group.Post("/positions", authMW, handler.CreatePosition)
	group.Put("/positions/:id", authMW, handler.UpdatePosition)
	group.Post("/positions/:id/status", authMW, handler.UpdatePositionStatus)
	group.Delete("/positions/:id", authMW, handler.DeletePosition)

	group.Get("/users", authMW, handler.Users)
	group.Post("/users", authMW, handler.CreateUser)
	group.Post("/users/import", authMW, handler.ImportUsers)
	group.Get("/users/export", authMW, handler.ExportUsers)
	group.Delete("/users/:uid", authMW, handler.DeleteUser)
	group.Put("/users/:uid", authMW, handler.UpdateUser)
	group.Post("/users/transfer-position", authMW, handler.BatchTransferUsers)
	group.Post("/users/:uid/reset_password", authMW, handler.ResetPassword)
	group.Get("/users/:uid/sessions", authMW, handler.UserSessions)
	return app
}

func TestOrganizationDepartmentsTreeAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodGet, "/v1/organization/departments/tree", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestOrganizationDepartmentDetailAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodGet, "/v1/organization/departments/1", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestOrganizationCreateDepartmentAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodPost, "/v1/organization/departments", map[string]any{
		"parent_id": 0,
		"name":      "产品中心",
		"code":      "DEPT-PROD",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestOrganizationCreateDepartmentLevelLimitAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodPost, "/v1/organization/departments", map[string]any{
		"parent_id": 4,
		"name":      "五级部门",
		"code":      "DEPT-L5",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) == 200 {
		t.Fatalf("expected level limit error, got success")
	}
}

func TestOrganizationUpdateDepartmentAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodPut, "/v1/organization/departments/1", map[string]any{
		"name": "产品中心",
		"code": "DEPT-PROD",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestOrganizationUpdateDepartmentStatusAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodPost, "/v1/organization/departments/1/status", map[string]any{
		"enabled": true,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestOrganizationDeleteDepartmentAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodDelete, "/v1/organization/departments/1", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestOrganizationPositionsAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodGet, "/v1/organization/positions?page_no=1&page_size=10", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestOrganizationPositionDetailAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodGet, "/v1/organization/positions/1", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestOrganizationCreatePositionAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodPost, "/v1/organization/positions", map[string]any{
		"name":          "前端工程师",
		"code":          "POS-FE",
		"department_id": 1,
		"level":         "P5",
		"hc":            5,
		"staffed":       3,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestOrganizationUpdatePositionAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodPut, "/v1/organization/positions/1", map[string]any{
		"name":          "前端工程师",
		"code":          "POS-FE",
		"department_id": 1,
		"level":         "P6",
		"hc":            6,
		"staffed":       4,
		"status":        1,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestOrganizationUpdatePositionStatusAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodPost, "/v1/organization/positions/1/status", map[string]any{
		"enabled": true,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestOrganizationDeletePositionAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodDelete, "/v1/organization/positions/1", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestOrganizationUsersAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodGet, "/v1/organization/users?pn=1&ps=1", nil)
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
	if !ok || len(items) != 1 {
		t.Fatalf("unexpected items: %v", data["items"])
	}
	if data["total"] != "1" {
		t.Fatalf("unexpected total: %v", data["total"])
	}
	first, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected first item type: %T", items[0])
	}
	if first["online_status"] != "online" {
		t.Fatalf("unexpected online_status: %v", first["online_status"])
	}
	if first["active_session_count"].(float64) != 2 {
		t.Fatalf("unexpected active_session_count: %v", first["active_session_count"])
	}
}

func TestOrganizationCreateUserAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodPost, "/v1/organization/users", map[string]any{
		"username":     "tester",
		"password":     "123456",
		"display_name": "测试用户",
		"status":       1,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestOrganizationUpdateUserAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodPut, "/v1/organization/users/10001", map[string]any{
		"display_name": "新名字",
		"status":       1,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestOrganizationDeleteUserAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodDelete, "/v1/organization/users/10001", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestOrganizationResetPasswordAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodPost, "/v1/organization/users/10001/reset_password", map[string]any{})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}

func TestOrganizationUserSessionsAPI(t *testing.T) {
	app := setupOrganizationApp()
	resp := request(t, app, http.MethodGet, "/v1/organization/users/10001/sessions?page_size=10&status=active", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	body := decodeJSON(t, resp)
	if body["code"].(float64) != 200 {
		t.Fatalf("unexpected code: %v", body["code"])
	}
}
