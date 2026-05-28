package test

import (
	"context"
	"net/http"
	"testing"

	permissionhandler "monorepo/internal/handler/permission"
	"monorepo/internal/middleware"
	"monorepo/internal/model"
	xadmin "monorepo/proto/xadminpb"

	"github.com/gofiber/fiber/v2"
)

type mockPermissionService struct {
	lastMenusReq *xadmin.PermissionMenusReq
}

func (m *mockPermissionService) ListMenus(ctx context.Context, req *xadmin.PermissionMenusReq) (*xadmin.PermissionMenusResp, error) {
	_ = ctx
	m.lastMenusReq = req
	return &xadmin.PermissionMenusResp{Total: 1, Items: []*xadmin.PermissionMenuItem{{Id: 1, Name: "菜单权限", MenuType: "menu", Status: "enabled"}}}, nil
}
func (m *mockPermissionService) GetMenuTree(ctx context.Context) (*xadmin.PermissionMenuTreeResp, error) {
	_ = ctx
	return &xadmin.PermissionMenuTreeResp{Items: []*xadmin.PermissionMenuNode{{Id: 1, Name: "权限管理"}}}, nil
}
func (m *mockPermissionService) GetMenu(ctx context.Context, req *xadmin.PermissionMenuDetailReq) (*xadmin.PermissionMenuItem, error) {
	_ = ctx
	return &xadmin.PermissionMenuItem{Id: req.GetId(), Name: "菜单权限", MenuType: "menu", Status: "enabled"}, nil
}
func (m *mockPermissionService) CreateMenu(ctx context.Context, req *xadmin.PermissionCreateMenuReq) (*xadmin.PermissionActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.PermissionActionResp{Success: true, Action: "create_menu"}, nil
}
func (m *mockPermissionService) UpdateMenu(ctx context.Context, req *xadmin.PermissionUpdateMenuReq) (*xadmin.PermissionActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.PermissionActionResp{Success: true, Action: "update_menu"}, nil
}
func (m *mockPermissionService) UpdateMenuStatus(ctx context.Context, req *xadmin.PermissionUpdateMenuStatusReq) (*xadmin.PermissionActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.PermissionActionResp{Success: true, Action: "update_menu_status"}, nil
}
func (m *mockPermissionService) DeleteMenu(ctx context.Context, req *xadmin.PermissionDeleteMenuReq) (*xadmin.PermissionActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.PermissionActionResp{Success: true, Action: "delete_menu"}, nil
}
func (m *mockPermissionService) SyncMenus(ctx context.Context) (*xadmin.PermissionActionResp, error) {
	_ = ctx
	return &xadmin.PermissionActionResp{Success: true, Action: "sync_menus"}, nil
}
func (m *mockPermissionService) ListRoles(ctx context.Context, req *xadmin.PermissionRolesReq) (*xadmin.PermissionRolesResp, error) {
	_ = ctx
	_ = req
	return &xadmin.PermissionRolesResp{Total: 1, Items: []*xadmin.PermissionRoleItem{{Id: 1, RoleName: "超级管理员", RoleType: "system"}}}, nil
}
func (m *mockPermissionService) GetRole(ctx context.Context, req *xadmin.PermissionRoleDetailReq) (*xadmin.PermissionRoleItem, error) {
	_ = ctx
	return &xadmin.PermissionRoleItem{Id: req.GetId(), RoleName: "超级管理员", RoleType: "system"}, nil
}
func (m *mockPermissionService) CreateRole(ctx context.Context, req *xadmin.PermissionCreateRoleReq) (*xadmin.PermissionActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.PermissionActionResp{Success: true, Action: "create_role"}, nil
}
func (m *mockPermissionService) UpdateRole(ctx context.Context, req *xadmin.PermissionUpdateRoleReq) (*xadmin.PermissionActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.PermissionActionResp{Success: true, Action: "update_role"}, nil
}
func (m *mockPermissionService) DeleteRole(ctx context.Context, operatorUID int32, req *xadmin.PermissionDeleteRoleReq) (*xadmin.PermissionActionResp, error) {
	_ = ctx
	_ = operatorUID
	_ = req
	return &xadmin.PermissionActionResp{Success: true, Action: "delete_role"}, nil
}
func (m *mockPermissionService) GetRoleMenus(ctx context.Context, req *xadmin.PermissionRoleMenusReq) (*xadmin.PermissionRoleMenusResp, error) {
	_ = ctx
	_ = req
	return &xadmin.PermissionRoleMenusResp{MenuIds: []int64{1, 2, 3}}, nil
}
func (m *mockPermissionService) UpdateRoleMenus(ctx context.Context, req *xadmin.PermissionUpdateRoleMenusReq) (*xadmin.PermissionActionResp, error) {
	_ = ctx
	_ = req
	return &xadmin.PermissionActionResp{Success: true, Action: "update_role_menus"}, nil
}
func setupPermissionAppWithMock(svc *mockPermissionService) *fiber.App {
	app := fiber.New()
	v1 := app.Group("/v1")
	handler := permissionhandler.NewHandlerWithService(svc)
	group := v1.Group("/permission")
	authMW := func(c *fiber.Ctx) error {
		c.Locals(middleware.AuthCtx{}, &middleware.AuthCtx{User: &model.AdminUser{UID: 10001}, SessionID: "s-1"})
		return c.Next()
	}
	group.Get("/menus/tree", authMW, handler.MenuTree)
	group.Get("/menus", authMW, handler.Menus)
	group.Get("/menus/:id", authMW, handler.Menu)
	group.Post("/menus", authMW, handler.CreateMenu)
	group.Put("/menus/:id", authMW, handler.UpdateMenu)
	group.Post("/menus/:id/status", authMW, handler.UpdateMenuStatus)
	group.Delete("/menus/:id", authMW, handler.DeleteMenu)
	group.Post("/menus/sync", authMW, handler.SyncMenus)
	group.Get("/roles", authMW, handler.Roles)
	group.Get("/roles/:id", authMW, handler.Role)
	group.Post("/roles", authMW, handler.CreateRole)
	group.Put("/roles/:id", authMW, handler.UpdateRole)
	group.Delete("/roles/:id", authMW, handler.DeleteRole)
	group.Get("/roles/:id/menus", authMW, handler.RoleMenus)
	group.Post("/roles/:id/menus", authMW, handler.UpdateRoleMenus)
	return app
}

func setupPermissionApp() *fiber.App {
	return setupPermissionAppWithMock(&mockPermissionService{})
}

func TestPermissionMenusAPI(t *testing.T) {
	app := setupPermissionApp()
	resp := request(t, app, http.MethodGet, "/v1/permission/menus?page_no=1&page_size=10", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestPermissionMenusAPIWithTreeNodeID(t *testing.T) {
	svc := &mockPermissionService{}
	app := setupPermissionAppWithMock(svc)
	resp := request(t, app, http.MethodGet, "/v1/permission/menus?page_no=1&page_size=10&tree_node_id=5", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if svc.lastMenusReq == nil {
		t.Fatalf("expected list menus request captured")
	}
	if svc.lastMenusReq.GetTreeNodeId() != 5 {
		t.Fatalf("expected tree_node_id=5, got=%d", svc.lastMenusReq.GetTreeNodeId())
	}
}

func TestPermissionMenusAPIWithSortField(t *testing.T) {
	svc := &mockPermissionService{}
	app := setupPermissionAppWithMock(svc)
	resp := request(t, app, http.MethodGet, "/v1/permission/menus?page_no=1&page_size=10&order_field=sort&order_type=asc", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if svc.lastMenusReq == nil {
		t.Fatalf("expected list menus request captured")
	}
	sortArgs := svc.lastMenusReq.GetSort()
	if len(sortArgs) != 1 {
		t.Fatalf("expected one sort arg, got=%d", len(sortArgs))
	}
	if sortArgs[0].GetOrderField() != "sort" {
		t.Fatalf("expected order_field=sort, got=%s", sortArgs[0].GetOrderField())
	}
}

func TestPermissionMenusAPIWithDeletedFilter(t *testing.T) {
	svc := &mockPermissionService{}
	app := setupPermissionAppWithMock(svc)
	resp := request(t, app, http.MethodGet, "/v1/permission/menus?page_no=1&page_size=10&deleted=yes", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if svc.lastMenusReq == nil {
		t.Fatalf("expected list menus request captured")
	}
	if svc.lastMenusReq.GetDeleted() != xadmin.PermissionMenuDeletedFilter_PERMISSION_MENU_DELETED_FILTER_YES {
		t.Fatalf("expected deleted=yes, got=%s", svc.lastMenusReq.GetDeleted().String())
	}
}

func TestPermissionMenuActionsAPI(t *testing.T) {
	app := setupPermissionApp()
	for _, tc := range []struct {
		method string
		path   string
		body   map[string]any
	}{
		{http.MethodPost, "/v1/permission/menus", map[string]any{"parent_id": 0, "name": "测试菜单", "menu_type": 2}},
		{http.MethodPut, "/v1/permission/menus/1", map[string]any{"name": "测试菜单", "menu_type": 2}},
		{http.MethodPost, "/v1/permission/menus/1/status", map[string]any{"enabled": true}},
		{http.MethodDelete, "/v1/permission/menus/1", nil},
		{http.MethodPost, "/v1/permission/menus/sync", map[string]any{}},
	} {
		resp := request(t, app, tc.method, tc.path, tc.body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s %s status=%d", tc.method, tc.path, resp.StatusCode)
		}
	}
}

func TestPermissionRolesAPI(t *testing.T) {
	app := setupPermissionApp()
	for _, tc := range []struct {
		method string
		path   string
		body   map[string]any
	}{
		{http.MethodGet, "/v1/permission/roles?page_no=1&page_size=10", nil},
		{http.MethodGet, "/v1/permission/roles/1", nil},
		{http.MethodPost, "/v1/permission/roles", map[string]any{"role_name": "测试角色", "role_type": 2}},
		{http.MethodPut, "/v1/permission/roles/1", map[string]any{"role_name": "测试角色", "role_type": 2}},
		{http.MethodDelete, "/v1/permission/roles/1", nil},
		{http.MethodGet, "/v1/permission/roles/1/menus", nil},
		{http.MethodPost, "/v1/permission/roles/1/menus", map[string]any{"menu_ids": []int64{1, 2}}},
	} {
		resp := request(t, app, tc.method, tc.path, tc.body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s %s status=%d", tc.method, tc.path, resp.StatusCode)
		}
	}
}
