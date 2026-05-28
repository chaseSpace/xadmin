package permission

import (
	"context"
	"fmt"
	"strings"
	"time"

	"monorepo/internal/middleware"
	"monorepo/internal/model"
	permissionrepo "monorepo/internal/repo/permission"
	"monorepo/internal/support/timefmt"
	"monorepo/pkg/consts"
	"monorepo/pkg/xerr"
	xadmin "monorepo/proto/xadminpb"
	commpb "monorepo/proto/xadminpb/commpb"
)

type Service interface {
	ListMenus(ctx context.Context, req *xadmin.PermissionMenusReq) (*xadmin.PermissionMenusResp, error)
	GetMenuTree(ctx context.Context) (*xadmin.PermissionMenuTreeResp, error)
	GetMenu(ctx context.Context, req *xadmin.PermissionMenuDetailReq) (*xadmin.PermissionMenuItem, error)
	CreateMenu(ctx context.Context, req *xadmin.PermissionCreateMenuReq) (*xadmin.PermissionActionResp, error)
	UpdateMenu(ctx context.Context, req *xadmin.PermissionUpdateMenuReq) (*xadmin.PermissionActionResp, error)
	UpdateMenuStatus(ctx context.Context, req *xadmin.PermissionUpdateMenuStatusReq) (*xadmin.PermissionActionResp, error)
	DeleteMenu(ctx context.Context, req *xadmin.PermissionDeleteMenuReq) (*xadmin.PermissionActionResp, error)
	SyncMenus(ctx context.Context) (*xadmin.PermissionActionResp, error)

	ListRoles(ctx context.Context, req *xadmin.PermissionRolesReq) (*xadmin.PermissionRolesResp, error)
	GetRole(ctx context.Context, req *xadmin.PermissionRoleDetailReq) (*xadmin.PermissionRoleItem, error)
	CreateRole(ctx context.Context, req *xadmin.PermissionCreateRoleReq) (*xadmin.PermissionActionResp, error)
	UpdateRole(ctx context.Context, req *xadmin.PermissionUpdateRoleReq) (*xadmin.PermissionActionResp, error)
	DeleteRole(ctx context.Context, operatorUID int32, req *xadmin.PermissionDeleteRoleReq) (*xadmin.PermissionActionResp, error)
	GetRoleMenus(ctx context.Context, req *xadmin.PermissionRoleMenusReq) (*xadmin.PermissionRoleMenusResp, error)
	UpdateRoleMenus(ctx context.Context, req *xadmin.PermissionUpdateRoleMenusReq) (*xadmin.PermissionActionResp, error)
}

type service struct {
	repo *permissionrepo.Repo
}

func NewService() Service {
	return &service{repo: permissionrepo.NewRepo()}
}

func NewServiceWithRepo(repo *permissionrepo.Repo) Service {
	return &service{repo: repo}
}

func (s *service) ListMenus(ctx context.Context, req *xadmin.PermissionMenusReq) (*xadmin.PermissionMenusResp, error) {
	page := req.GetPage()
	if page == nil {
		page = &commpb.PageArgs{Pn: 1, Ps: 10}
	}
	if page.GetPn() <= 0 {
		page.Pn = 1
	}
	if page.GetPs() <= 0 {
		page.Ps = 10
	}
	filters := buildMenuFilters(req)
	if req.GetTreeNodeId() > 0 {
		allRows, err := s.repo.ListAllMenus(ctx)
		if err != nil {
			return nil, err
		}
		subtreeMenuIDs := collectMenuSubtreeIDs(allRows, req.GetTreeNodeId())
		if len(subtreeMenuIDs) == 0 {
			return &xadmin.PermissionMenusResp{
				Items: []*xadmin.PermissionMenuItem{},
				Total: 0,
				Page:  &commpb.PageArgs{Pn: page.GetPn(), Ps: page.GetPs()},
			}, nil
		}
		filters.MenuIDs = subtreeMenuIDs
	}
	rows, total, err := s.repo.ListMenus(ctx, page, normalizeMenuSortArgs(req.GetSort()), filters)
	if err != nil {
		return nil, err
	}
	items := make([]*xadmin.PermissionMenuItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapMenuRow(&row))
	}
	return &xadmin.PermissionMenusResp{
		Items: items,
		Total: total,
		Page:  &commpb.PageArgs{Pn: page.GetPn(), Ps: page.GetPs()},
	}, nil
}

func (s *service) GetMenuTree(ctx context.Context) (*xadmin.PermissionMenuTreeResp, error) {
	rows, err := s.repo.ListAllMenus(ctx)
	if err != nil {
		return nil, err
	}
	nodes := make(map[int64]*xadmin.PermissionMenuNode, len(rows))
	childrenMap := make(map[int64][]*xadmin.PermissionMenuNode, len(rows))
	roots := make([]*xadmin.PermissionMenuNode, 0, 8)
	for _, row := range rows {
		node := &xadmin.PermissionMenuNode{Id: row.ID, ParentId: row.ParentID, Name: row.Name}
		nodes[row.ID] = node
		childrenMap[row.ParentID] = append(childrenMap[row.ParentID], node)
	}
	for _, row := range rows {
		nodes[row.ID].Children = childrenMap[row.ID]
		if row.ParentID == 0 {
			roots = append(roots, nodes[row.ID])
		}
	}
	return &xadmin.PermissionMenuTreeResp{Items: roots}, nil
}

func (s *service) GetMenu(ctx context.Context, req *xadmin.PermissionMenuDetailReq) (*xadmin.PermissionMenuItem, error) {
	row, err := s.repo.GetMenuByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return mapMenuRow(row), nil
}

func (s *service) CreateMenu(ctx context.Context, req *xadmin.PermissionCreateMenuReq) (*xadmin.PermissionActionResp, error) {
	if req.GetParentId() > 0 {
		if _, err := s.repo.GetMenuByID(ctx, req.GetParentId()); err != nil {
			return nil, err
		}
	}
	menuType, err := menuTypeProtoToDB(req.GetMenuType())
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateMenu(ctx, &model.PermissionMenu{
		ParentID:      req.GetParentId(),
		Name:          strings.TrimSpace(req.GetName()),
		RoutePath:     strings.TrimSpace(req.GetRoutePath()),
		ComponentPath: strings.TrimSpace(req.GetComponentPath()),
		MenuType:      menuType,
		PermissionKey: strings.TrimSpace(req.GetPermissionKey()),
		Sort:          req.GetSort(),
		Status:        consts.PermissionStatusEnabled,
	}); err != nil {
		return nil, err
	}
	return &xadmin.PermissionActionResp{Success: true, Action: "create_menu"}, nil
}

func (s *service) UpdateMenu(ctx context.Context, req *xadmin.PermissionUpdateMenuReq) (*xadmin.PermissionActionResp, error) {
	if _, err := s.repo.GetMenuByID(ctx, req.GetId()); err != nil {
		return nil, err
	}
	menuType, err := menuTypeProtoToDB(req.GetMenuType())
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateMenuByID(ctx, req.GetId(), map[string]any{
		"name":           strings.TrimSpace(req.GetName()),
		"route_path":     strings.TrimSpace(req.GetRoutePath()),
		"component_path": strings.TrimSpace(req.GetComponentPath()),
		"menu_type":      menuType,
		"permission_key": strings.TrimSpace(req.GetPermissionKey()),
		"sort":           req.GetSort(),
	}); err != nil {
		return nil, err
	}
	return &xadmin.PermissionActionResp{Success: true, Action: "update_menu"}, nil
}

func (s *service) UpdateMenuStatus(ctx context.Context, req *xadmin.PermissionUpdateMenuStatusReq) (*xadmin.PermissionActionResp, error) {
	if _, err := s.repo.GetMenuByID(ctx, req.GetId()); err != nil {
		return nil, err
	}
	status := consts.PermissionStatusDisabled
	if req.GetEnabled() {
		status = consts.PermissionStatusEnabled
	}
	if err := s.repo.UpdateMenuByID(ctx, req.GetId(), map[string]any{"status": status}); err != nil {
		return nil, err
	}
	return &xadmin.PermissionActionResp{Success: true, Action: "update_menu_status"}, nil
}

func (s *service) DeleteMenu(ctx context.Context, req *xadmin.PermissionDeleteMenuReq) (*xadmin.PermissionActionResp, error) {
	row, err := s.repo.GetMenuByIDIncludingDeleted(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if row.DeletedAt != 0 {
		now := time.Now().Unix()
		if now-row.DeletedAt < int64(time.Hour/time.Second) {
			remainingSeconds := int64(time.Hour/time.Second) - (now - row.DeletedAt)
			return nil, xerr.NewBiz(xerr.CodeBadRequest, "perm.delete_cooldown", formatRemainingDuration(remainingSeconds))
		}
		if err := s.repo.HardDeleteMenuByID(ctx, req.GetId()); err != nil {
			return nil, err
		}
		return &xadmin.PermissionActionResp{Success: true, Action: "hard_delete_menu"}, nil
	}
	hasChild, err := s.repo.HasMenuChildren(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if hasChild {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "perm.has_children")
	}
	if err := s.repo.SoftDeleteMenuByID(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &xadmin.PermissionActionResp{Success: true, Action: "delete_menu"}, nil
}

func (s *service) SyncMenus(ctx context.Context) (*xadmin.PermissionActionResp, error) {
	_ = ctx
	return &xadmin.PermissionActionResp{Success: true, Action: "sync_menus"}, nil
}

func (s *service) ListRoles(ctx context.Context, req *xadmin.PermissionRolesReq) (*xadmin.PermissionRolesResp, error) {
	page := req.GetPage()
	if page == nil {
		page = &commpb.PageArgs{Pn: 1, Ps: 10}
	}
	if page.GetPn() <= 0 {
		page.Pn = 1
	}
	if page.GetPs() <= 0 {
		page.Ps = 10
	}
	rows, total, err := s.repo.ListRoles(ctx, page, normalizeRoleSortArgs(req.GetSort()), buildRoleFilters(req))
	if err != nil {
		return nil, err
	}
	items := make([]*xadmin.PermissionRoleItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapRoleRow(&row))
	}
	return &xadmin.PermissionRolesResp{Items: items, Total: total, Page: &commpb.PageArgs{Pn: page.GetPn(), Ps: page.GetPs()}}, nil
}

func (s *service) GetRole(ctx context.Context, req *xadmin.PermissionRoleDetailReq) (*xadmin.PermissionRoleItem, error) {
	row, err := s.repo.GetRoleByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return mapRoleRow(row), nil
}

func (s *service) CreateRole(ctx context.Context, req *xadmin.PermissionCreateRoleReq) (*xadmin.PermissionActionResp, error) {
	roleType, err := roleTypeProtoToDB(req.GetRoleType())
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateRole(ctx, &model.PermissionRole{
		RoleName: strings.TrimSpace(req.GetRoleName()),
		RoleType: roleType,
	}); err != nil {
		return nil, err
	}
	return &xadmin.PermissionActionResp{Success: true, Action: "create_role"}, nil
}

func (s *service) UpdateRole(ctx context.Context, req *xadmin.PermissionUpdateRoleReq) (*xadmin.PermissionActionResp, error) {
	role, err := s.repo.GetRoleByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if isSystemRole(role) {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "perm.system_role_immutable")
	}
	roleType, err := roleTypeProtoToDB(req.GetRoleType())
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateRoleByID(ctx, req.GetId(), map[string]any{
		"role_name": strings.TrimSpace(req.GetRoleName()),
		"role_type": roleType,
	}); err != nil {
		return nil, err
	}
	return &xadmin.PermissionActionResp{Success: true, Action: "update_role"}, nil
}

func (s *service) DeleteRole(ctx context.Context, operatorUID int32, req *xadmin.PermissionDeleteRoleReq) (*xadmin.PermissionActionResp, error) {
	role, err := s.repo.GetRoleByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if isSystemRole(role) {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "perm.system_role_undeletable")
	}
	if operatorUID > 0 {
		bound, err := s.repo.HasRoleUser(ctx, req.GetId(), operatorUID)
		if err != nil {
			return nil, err
		}
		if bound {
			return nil, xerr.NewBiz(xerr.CodeBadRequest, "perm.role_affects_self")
		}
	}
	if err := s.repo.SoftDeleteRoleByID(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &xadmin.PermissionActionResp{Success: true, Action: "delete_role"}, nil
}

func (s *service) GetRoleMenus(ctx context.Context, req *xadmin.PermissionRoleMenusReq) (*xadmin.PermissionRoleMenusResp, error) {
	if _, err := s.repo.GetRoleByID(ctx, req.GetRoleId()); err != nil {
		return nil, err
	}
	ids, err := s.repo.ListRoleMenuIDs(ctx, req.GetRoleId())
	if err != nil {
		return nil, err
	}
	return &xadmin.PermissionRoleMenusResp{MenuIds: ids}, nil
}

func (s *service) UpdateRoleMenus(ctx context.Context, req *xadmin.PermissionUpdateRoleMenusReq) (*xadmin.PermissionActionResp, error) {
	role, err := s.repo.GetRoleByID(ctx, req.GetRoleId())
	if err != nil {
		return nil, err
	}
	if isRootAdminRole(role) {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "perm.superadmin_immutable")
	}
	menuIDs, err := s.repo.ExpandMenuIDsWithAncestors(ctx, req.GetMenuIds())
	if err != nil {
		return nil, err
	}
	if err := s.repo.ReplaceRoleMenus(ctx, req.GetRoleId(), menuIDs); err != nil {
		return nil, err
	}
	middleware.InvalidateAllPermissionCache()
	return &xadmin.PermissionActionResp{Success: true, Action: "update_role_menus"}, nil
}

func mapMenuRow(row *permissionrepo.MenuRow) *xadmin.PermissionMenuItem {
	item := &xadmin.PermissionMenuItem{
		Id:            row.ID,
		ParentId:      row.ParentID,
		Name:          row.Name,
		RoutePath:     row.RoutePath,
		ComponentPath: row.ComponentPath,
		MenuType:      menuTypeText(row.MenuType),
		PermissionKey: row.PermissionKey,
		Sort:          row.Sort,
		Status:        enabledStatusText(row.Status),
	}
	if row.UpdatedAt != nil {
		item.UpdatedAt = timefmt.RFC3339Ptr(row.UpdatedAt)
	}
	if row.DeletedAt != 0 {
		item.Deleted = true
		item.DeletedAt = timefmt.RFC3339Unix(row.DeletedAt)
	}
	return item
}

func mapRoleRow(row *permissionrepo.RoleRow) *xadmin.PermissionRoleItem {
	item := &xadmin.PermissionRoleItem{
		Id:       row.ID,
		RoleName: row.RoleName,
		RoleType: roleTypeText(row.RoleType),
		Users:    row.Users,
	}
	if row.UpdatedAt != nil {
		item.UpdatedAt = timefmt.RFC3339Ptr(row.UpdatedAt)
	}
	return item
}

func enabledStatusText(status int32) string {
	if status == consts.PermissionStatusEnabled {
		return "enabled"
	}
	return "disabled"
}

func menuTypeText(menuType int32) string {
	switch menuType {
	case consts.PermissionMenuTypeDirectory:
		return "directory"
	case consts.PermissionMenuTypeMenu:
		return "menu"
	case consts.PermissionMenuTypeButton:
		return "button"
	default:
		return "menu"
	}
}

func roleTypeText(roleType int32) string {
	if roleType == consts.PermissionRoleTypeSystem {
		return "system"
	}
	return "custom"
}

func menuTypeProtoToDB(menuType xadmin.PermissionMenuType) (int32, error) {
	switch menuType {
	case xadmin.PermissionMenuType_PERMISSION_MENU_TYPE_DIRECTORY:
		return consts.PermissionMenuTypeDirectory, nil
	case xadmin.PermissionMenuType_PERMISSION_MENU_TYPE_MENU:
		return consts.PermissionMenuTypeMenu, nil
	case xadmin.PermissionMenuType_PERMISSION_MENU_TYPE_BUTTON:
		return consts.PermissionMenuTypeButton, nil
	default:
		return 0, xerr.NewBiz(xerr.CodeBadRequest, "perm.invalid_menu_type")
	}
}

func roleTypeProtoToDB(roleType xadmin.PermissionRoleType) (int32, error) {
	switch roleType {
	case xadmin.PermissionRoleType_PERMISSION_ROLE_TYPE_SYSTEM:
		return consts.PermissionRoleTypeSystem, nil
	case xadmin.PermissionRoleType_PERMISSION_ROLE_TYPE_CUSTOM:
		return consts.PermissionRoleTypeCustom, nil
	default:
		return 0, xerr.NewBiz(xerr.CodeBadRequest, "perm.invalid_role_type")
	}
}

func normalizeMenuSortArgs(input []*commpb.SortArgs) []*commpb.SortArgs {
	if len(input) == 0 {
		return nil
	}
	fieldMap := map[string]string{"id": "m.id", "name": "m.name", "menu_type": "m.menu_type", "sort": "m.sort", "status": "m.status", "updated_at": "m.updated_at", "created_at": "m.created_at"}
	out := make([]*commpb.SortArgs, 0, len(input))
	for _, item := range input {
		if item == nil {
			continue
		}
		mapped, ok := fieldMap[item.GetOrderField()]
		if !ok {
			continue
		}
		out = append(out, &commpb.SortArgs{OrderField: mapped, OrderType: item.GetOrderType()})
	}
	return out
}

func normalizeRoleSortArgs(input []*commpb.SortArgs) []*commpb.SortArgs {
	if len(input) == 0 {
		return nil
	}
	fieldMap := map[string]string{"id": "r.id", "role_name": "r.role_name", "role_type": "r.role_type", "updated_at": "r.updated_at", "created_at": "r.created_at", "users": "users"}
	out := make([]*commpb.SortArgs, 0, len(input))
	for _, item := range input {
		if item == nil {
			continue
		}
		mapped, ok := fieldMap[item.GetOrderField()]
		if !ok {
			continue
		}
		out = append(out, &commpb.SortArgs{OrderField: mapped, OrderType: item.GetOrderType()})
	}
	return out
}

func buildMenuFilters(req *xadmin.PermissionMenusReq) permissionrepo.MenuFilters {
	var statusPtr *int32
	switch req.GetStatus() {
	case xadmin.PermissionMenuFilterStatus_PERMISSION_MENU_FILTER_STATUS_ENABLED:
		v := consts.PermissionStatusEnabled
		statusPtr = &v
	case xadmin.PermissionMenuFilterStatus_PERMISSION_MENU_FILTER_STATUS_DISABLED:
		v := consts.PermissionStatusDisabled
		statusPtr = &v
	}
	var menuTypePtr *int32
	switch req.GetMenuType() {
	case xadmin.PermissionMenuType_PERMISSION_MENU_TYPE_DIRECTORY:
		v := consts.PermissionMenuTypeDirectory
		menuTypePtr = &v
	case xadmin.PermissionMenuType_PERMISSION_MENU_TYPE_MENU:
		v := consts.PermissionMenuTypeMenu
		menuTypePtr = &v
	case xadmin.PermissionMenuType_PERMISSION_MENU_TYPE_BUTTON:
		v := consts.PermissionMenuTypeButton
		menuTypePtr = &v
	}
	deleted := false
	if req.GetDeleted() == xadmin.PermissionMenuDeletedFilter_PERMISSION_MENU_DELETED_FILTER_YES {
		deleted = true
	}
	return permissionrepo.MenuFilters{
		Keyword:  strings.TrimSpace(req.GetKeyword()),
		Status:   statusPtr,
		MenuType: menuTypePtr,
		Deleted:  &deleted,
	}
}

func formatRemainingDuration(seconds int64) string {
	if seconds <= 0 {
		return "0分钟"
	}
	minutes := (seconds + 59) / 60
	if minutes < 60 {
		return fmt.Sprintf("%d分钟", minutes)
	}
	return fmt.Sprintf("%d小时%d分钟", minutes/60, minutes%60)
}

func collectMenuSubtreeIDs(rows []permissionrepo.MenuRow, rootID int64) []int64 {
	if rootID <= 0 || len(rows) == 0 {
		return nil
	}
	children := make(map[int64][]int64, len(rows))
	exists := false
	for _, row := range rows {
		if row.ID == rootID {
			exists = true
		}
		children[row.ParentID] = append(children[row.ParentID], row.ID)
	}
	if !exists {
		return nil
	}
	result := make([]int64, 0, 16)
	queue := []int64{rootID}
	seen := map[int64]struct{}{}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if _, ok := seen[current]; ok {
			continue
		}
		seen[current] = struct{}{}
		result = append(result, current)
		queue = append(queue, children[current]...)
	}
	return result
}

func buildRoleFilters(req *xadmin.PermissionRolesReq) permissionrepo.RoleFilters {
	var roleTypePtr *int32
	switch req.GetRoleType() {
	case xadmin.PermissionRoleType_PERMISSION_ROLE_TYPE_SYSTEM:
		v := consts.PermissionRoleTypeSystem
		roleTypePtr = &v
	case xadmin.PermissionRoleType_PERMISSION_ROLE_TYPE_CUSTOM:
		v := consts.PermissionRoleTypeCustom
		roleTypePtr = &v
	}
	return permissionrepo.RoleFilters{Keyword: strings.TrimSpace(req.GetKeyword()), RoleType: roleTypePtr}
}

func isRootAdminRole(role *permissionrepo.RoleRow) bool {
	if role == nil {
		return false
	}
	return role.RoleType == consts.PermissionRoleTypeSystem && strings.TrimSpace(role.RoleName) == "超级管理员"
}

func isSystemRole(role *permissionrepo.RoleRow) bool {
	if role == nil {
		return false
	}
	return role.RoleType == consts.PermissionRoleTypeSystem
}
