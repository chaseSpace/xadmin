package permission

import (
	"strconv"
	"strings"

	"monorepo/internal/middleware"
	permissionsvc "monorepo/internal/service/permission"
	"monorepo/internal/support/auditlog"
	"monorepo/pkg/xfiber"
	xadmin "monorepo/proto/xadminpb"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/encoding/protojson"
	proto2 "google.golang.org/protobuf/proto"
)

type Handler struct {
	svc permissionsvc.Service
}

func NewHandler() *Handler {
	return &Handler{svc: permissionsvc.NewService()}
}

func NewHandlerWithService(svc permissionsvc.Service) *Handler {
	return &Handler{svc: svc}
}

func RegisterRoutes(prefix string, parent fiber.Router, authMW fiber.Handler) {
	handler := NewHandler()
	group := parent.Group(prefix, authMW)
	edit := middleware.RequirePermission

	group.Get("/menus/tree", handler.MenuTree)
	group.Get("/menus", handler.Menus)
	group.Get("/menus/:id", handler.Menu)
	group.Post("/menus", edit("permission.menus.edit"), handler.CreateMenu)
	group.Put("/menus/:id", edit("permission.menus.edit"), handler.UpdateMenu)
	group.Post("/menus/:id/status", edit("permission.menus.edit"), handler.UpdateMenuStatus)
	group.Delete("/menus/:id", edit("permission.menus.delete"), handler.DeleteMenu)
	group.Post("/menus/sync", edit("permission.menus.edit"), handler.SyncMenus)

	group.Get("/roles", handler.Roles)
	group.Get("/roles/:id", handler.Role)
	group.Post("/roles", edit("permission.roles.edit"), handler.CreateRole)
	group.Put("/roles/:id", edit("permission.roles.edit"), handler.UpdateRole)
	group.Delete("/roles/:id", edit("permission.roles.delete"), handler.DeleteRole)
	group.Get("/roles/:id/menus", handler.RoleMenus)
	group.Post("/roles/:id/menus", edit("permission.roles.edit"), handler.UpdateRoleMenus)
}

func (h *Handler) MenuTree(c *fiber.Ctx) error {
	resp, err := h.svc.GetMenuTree(c.UserContext())
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_menu_tree",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) Menus(c *fiber.Ctx) error {
	page, sort, err := xfiber.ParsePageSortQuery(c, 1, 10)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.PermissionMenusReq{Page: page, Sort: sort}
	req.Keyword = strings.TrimSpace(c.Query("keyword"))
	switch strings.ToLower(strings.TrimSpace(c.Query("status"))) {
	case "enabled", "1":
		req.Status = xadmin.PermissionMenuFilterStatus_PERMISSION_MENU_FILTER_STATUS_ENABLED
	case "disabled", "0":
		req.Status = xadmin.PermissionMenuFilterStatus_PERMISSION_MENU_FILTER_STATUS_DISABLED
	}
	switch strings.ToLower(strings.TrimSpace(c.Query("menu_type"))) {
	case "directory", "1":
		req.MenuType = xadmin.PermissionMenuType_PERMISSION_MENU_TYPE_DIRECTORY
	case "menu", "2":
		req.MenuType = xadmin.PermissionMenuType_PERMISSION_MENU_TYPE_MENU
	case "button", "3":
		req.MenuType = xadmin.PermissionMenuType_PERMISSION_MENU_TYPE_BUTTON
	}
	switch strings.ToLower(strings.TrimSpace(c.Query("deleted"))) {
	case "yes", "true", "1":
		req.Deleted = xadmin.PermissionMenuDeletedFilter_PERMISSION_MENU_DELETED_FILTER_YES
	case "no", "false", "0", "":
		req.Deleted = xadmin.PermissionMenuDeletedFilter_PERMISSION_MENU_DELETED_FILTER_NO
	}
	if treeNodeIDRaw := strings.TrimSpace(c.Query("tree_node_id")); treeNodeIDRaw != "" {
		treeNodeID, parseErr := strconv.ParseInt(treeNodeIDRaw, 10, 64)
		if parseErr != nil {
			return xfiber.StdResponse(c, nil, parseErr)
		}
		req.TreeNodeId = treeNodeID
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.ListMenus(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_menu_list",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) Menu(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.PermissionMenuDetailReq{Id: id}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.GetMenu(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_menu_detail",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) CreateMenu(c *fiber.Ctx) error {
	req := &xadmin.PermissionCreateMenuReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.CreateMenu(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "create_menu",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) UpdateMenu(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.PermissionUpdateMenuReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req.Id = id
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.UpdateMenu(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "update_menu",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) UpdateMenuStatus(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.PermissionUpdateMenuStatusReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req.Id = id
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.UpdateMenuStatus(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "update_menu_status",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) DeleteMenu(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.PermissionDeleteMenuReq{Id: id}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.DeleteMenu(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "delete_menu",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) SyncMenus(c *fiber.Ctx) error {
	resp, err := h.svc.SyncMenus(c.UserContext())
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "sync_menus",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) Roles(c *fiber.Ctx) error {
	page, sort, err := xfiber.ParsePageSortQuery(c, 1, 10)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.PermissionRolesReq{Page: page, Sort: sort}
	req.Keyword = strings.TrimSpace(c.Query("keyword"))
	switch strings.ToLower(strings.TrimSpace(c.Query("role_type"))) {
	case "system", "1":
		req.RoleType = xadmin.PermissionRoleType_PERMISSION_ROLE_TYPE_SYSTEM
	case "custom", "2":
		req.RoleType = xadmin.PermissionRoleType_PERMISSION_ROLE_TYPE_CUSTOM
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.ListRoles(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_role_list",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) Role(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.PermissionRoleDetailReq{Id: id}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.GetRole(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_role_detail",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) CreateRole(c *fiber.Ctx) error {
	req := &xadmin.PermissionCreateRoleReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.CreateRole(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "create_role",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) UpdateRole(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.PermissionUpdateRoleReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req.Id = id
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.UpdateRole(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "update_role",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) DeleteRole(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.PermissionDeleteRoleReq{Id: id}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.DeleteRole(c.UserContext(), middleware.GetUID(c), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "delete_role",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) RoleMenus(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.PermissionRoleMenusReq{RoleId: id}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.GetRoleMenus(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_role_menus",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "role_id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) UpdateRoleMenus(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.PermissionUpdateRoleMenusReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req.RoleId = id
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.UpdateRoleMenus(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "update_role_menus",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "role_id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func parseProtoRequest(c *fiber.Ctx, req proto2.Message) error {
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(c.Body(), req); err != nil {
		return err
	}
	return nil
}
