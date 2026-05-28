package organization

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/csv"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"monorepo/internal/model"
	organizationrepo "monorepo/internal/repo/organization"
	"monorepo/internal/support/timefmt"
	"monorepo/pkg/consts"
	"monorepo/pkg/xerr"
	xadmin "monorepo/proto/xadminpb"
	commpb "monorepo/proto/xadminpb/commpb"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	GetDepartmentsTree(ctx context.Context) (*xadmin.OrganizationDepartmentsTreeResp, error)
	GetDepartment(ctx context.Context, req *xadmin.OrganizationDepartmentDetailReq) (*xadmin.OrganizationDepartmentItem, error)
	CreateDepartment(ctx context.Context, req *xadmin.OrganizationCreateDepartmentReq) (*xadmin.OrganizationActionResp, error)
	UpdateDepartment(ctx context.Context, req *xadmin.OrganizationUpdateDepartmentReq) (*xadmin.OrganizationActionResp, error)
	UpdateDepartmentStatus(ctx context.Context, req *xadmin.OrganizationUpdateDepartmentStatusReq) (*xadmin.OrganizationActionResp, error)
	DeleteDepartment(ctx context.Context, req *xadmin.OrganizationDeleteDepartmentReq) (*xadmin.OrganizationActionResp, error)
	ListPositions(ctx context.Context, req *xadmin.OrganizationPositionsReq) (*xadmin.OrganizationPositionsResp, error)
	GetPosition(ctx context.Context, req *xadmin.OrganizationPositionDetailReq) (*xadmin.OrganizationPositionItem, error)
	CreatePosition(ctx context.Context, req *xadmin.OrganizationCreatePositionReq) (*xadmin.OrganizationActionResp, error)
	UpdatePosition(ctx context.Context, req *xadmin.OrganizationUpdatePositionReq) (*xadmin.OrganizationActionResp, error)
	UpdatePositionStatus(ctx context.Context, req *xadmin.OrganizationUpdatePositionStatusReq) (*xadmin.OrganizationActionResp, error)
	DeletePosition(ctx context.Context, req *xadmin.OrganizationDeletePositionReq) (*xadmin.OrganizationActionResp, error)

	ListUsers(ctx context.Context, req *xadmin.OrganizationUsersReq) (*xadmin.OrganizationUsersResp, error)
	ListUserSessions(ctx context.Context, req *xadmin.OrganizationUserSessionsReq) (*xadmin.OrganizationUserSessionsResp, error)
	CreateUser(ctx context.Context, req *xadmin.OrganizationCreateUserReq) (*xadmin.OrganizationActionResp, error)
	UpdateUser(ctx context.Context, req *xadmin.OrganizationUpdateUserReq) (*xadmin.OrganizationActionResp, error)
	BatchTransferUsers(ctx context.Context, req *xadmin.OrganizationBatchTransferUsersReq) (*xadmin.OrganizationActionResp, error)
	ResetPassword(ctx context.Context, req *xadmin.OrganizationResetPasswordReq) (*xadmin.OrganizationActionResp, error)
	DeleteUser(ctx context.Context, req *xadmin.OrganizationDeleteUserReq) (*xadmin.OrganizationActionResp, error)
	ImportUsers(ctx context.Context, req *xadmin.OrganizationImportUsersReq) (*xadmin.OrganizationActionResp, error)
	ExportUsers(ctx context.Context, req *xadmin.OrganizationUsersReq) ([]byte, error)
}

type service struct {
	repo *organizationrepo.Repo
}

func NewService() Service {
	return &service{repo: organizationrepo.NewRepo()}
}

func NewServiceWithRepo(repo *organizationrepo.Repo) Service {
	return &service{repo: repo}
}

func (s *service) GetDepartmentsTree(ctx context.Context) (*xadmin.OrganizationDepartmentsTreeResp, error) {
	rows, err := s.repo.ListDepartments(ctx)
	if err != nil {
		return nil, err
	}

	childrenMap := make(map[int64][]*xadmin.OrganizationDepartmentItem, len(rows))
	nodes := make(map[int64]*xadmin.OrganizationDepartmentItem, len(rows))
	roots := make([]*xadmin.OrganizationDepartmentItem, 0, 8)

	for _, row := range rows {
		item := mapDepartmentRow(&row)
		nodes[row.ID] = item
		childrenMap[row.ParentID] = append(childrenMap[row.ParentID], item)
	}
	for _, row := range rows {
		if row.ParentID == 0 {
			roots = append(roots, nodes[row.ID])
		}
	}
	for _, row := range rows {
		nodes[row.ID].Children = childrenMap[row.ID]
	}

	return &xadmin.OrganizationDepartmentsTreeResp{Items: roots}, nil
}

func (s *service) GetDepartment(ctx context.Context, req *xadmin.OrganizationDepartmentDetailReq) (*xadmin.OrganizationDepartmentItem, error) {
	row, err := s.repo.GetDepartmentByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return mapDepartmentRow(row), nil
}

func (s *service) CreateDepartment(ctx context.Context, req *xadmin.OrganizationCreateDepartmentReq) (*xadmin.OrganizationActionResp, error) {
	if req.GetParentId() > 0 {
		parent, err := s.repo.GetDepartmentByID(ctx, req.GetParentId())
		if err != nil {
			return nil, err
		}
		if s.departmentLevel(ctx, parent.ID) >= 4 {
			return nil, xerr.NewBiz(xerr.CodeBadRequest, "org.max_depth")
		}
	}
	if err := s.repo.CreateDepartment(ctx, &model.OrganizationDepartment{
		ParentID: req.GetParentId(),
		Name:     strings.TrimSpace(req.GetName()),
		Code:     strings.TrimSpace(req.GetCode()),
		Status:   consts.DepartmentStatusEnabled,
	}); err != nil {
		return nil, err
	}
	return &xadmin.OrganizationActionResp{Success: true, Action: "create_department"}, nil
}

func (s *service) departmentLevel(ctx context.Context, departmentID int64) int {
	level := 0
	currentID := departmentID
	visited := make(map[int64]struct{}, 8)
	for currentID > 0 {
		if _, ok := visited[currentID]; ok {
			break
		}
		visited[currentID] = struct{}{}
		row, err := s.repo.GetDepartmentByID(ctx, currentID)
		if err != nil || row == nil {
			break
		}
		level += 1
		currentID = row.ParentID
	}
	return level
}

func (s *service) UpdateDepartment(ctx context.Context, req *xadmin.OrganizationUpdateDepartmentReq) (*xadmin.OrganizationActionResp, error) {
	if _, err := s.repo.GetDepartmentByID(ctx, req.GetId()); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateDepartmentByID(ctx, req.GetId(), map[string]any{
		"name": strings.TrimSpace(req.GetName()),
		"code": strings.TrimSpace(req.GetCode()),
	}); err != nil {
		return nil, err
	}
	return &xadmin.OrganizationActionResp{Success: true, Action: "update_department"}, nil
}

func (s *service) UpdateDepartmentStatus(ctx context.Context, req *xadmin.OrganizationUpdateDepartmentStatusReq) (*xadmin.OrganizationActionResp, error) {
	if _, err := s.repo.GetDepartmentByID(ctx, req.GetId()); err != nil {
		return nil, err
	}
	status := consts.DepartmentStatusDisabled
	if req.GetEnabled() {
		status = consts.DepartmentStatusEnabled
	}
	if err := s.repo.UpdateDepartmentByID(ctx, req.GetId(), map[string]any{
		"status": status,
	}); err != nil {
		return nil, err
	}
	if status == consts.DepartmentStatusDisabled {
		if err := s.repo.RevokeSessionsByDepartmentID(ctx, req.GetId(), "department_disabled"); err != nil {
			return nil, err
		}
	}
	return &xadmin.OrganizationActionResp{Success: true, Action: "update_department_status"}, nil
}

func (s *service) DeleteDepartment(ctx context.Context, req *xadmin.OrganizationDeleteDepartmentReq) (*xadmin.OrganizationActionResp, error) {
	if _, err := s.repo.GetDepartmentByID(ctx, req.GetId()); err != nil {
		return nil, err
	}
	hasChild, err := s.repo.HasChildDepartment(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	positionCount, err := s.repo.CountPositionsByDepartmentID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	userCount, err := s.repo.CountUsersByDepartmentID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if userCount > 0 {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "org.has_members", userCount)
	}
	if (hasChild || positionCount > 0) && !req.GetForce() {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "org.has_children_force")
	}
	if err := s.repo.SoftDeleteDepartmentByID(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &xadmin.OrganizationActionResp{Success: true, Action: "delete_department"}, nil
}

func (s *service) ListPositions(ctx context.Context, req *xadmin.OrganizationPositionsReq) (*xadmin.OrganizationPositionsResp, error) {
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
	rows, total, err := s.repo.ListPositions(ctx, page, normalizePositionSortArgs(req.GetSort()), buildPositionFilters(req))
	if err != nil {
		return nil, err
	}
	items := make([]*xadmin.OrganizationPositionItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapPositionRow(&row))
	}
	return &xadmin.OrganizationPositionsResp{
		Items: items,
		Total: total,
		Page: &commpb.PageArgs{
			Pn: page.GetPn(),
			Ps: page.GetPs(),
		},
	}, nil
}

func (s *service) GetPosition(ctx context.Context, req *xadmin.OrganizationPositionDetailReq) (*xadmin.OrganizationPositionItem, error) {
	row, err := s.repo.GetPositionByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return mapPositionRow(row), nil
}

func (s *service) CreatePosition(ctx context.Context, req *xadmin.OrganizationCreatePositionReq) (*xadmin.OrganizationActionResp, error) {
	if _, err := s.repo.GetDepartmentByID(ctx, req.GetDepartmentId()); err != nil {
		return nil, err
	}
	roleIDs := normalizeRoleIDs(req.GetRoleIds())
	if len(roleIDs) > 0 {
		count, err := s.repo.CountValidRolesByIDs(ctx, roleIDs)
		if err != nil {
			return nil, err
		}
		if count != int64(len(roleIDs)) {
			return nil, xerr.NewBiz(xerr.CodeBadRequest, "org.role_not_found")
		}
	}
	position := &model.OrganizationPosition{
		Name:         strings.TrimSpace(req.GetName()),
		Code:         strings.TrimSpace(req.GetCode()),
		DepartmentID: req.GetDepartmentId(),
		Level:        strings.TrimSpace(req.GetLevel()),
		Hc:           req.GetHc(),
		Staffed:      req.GetStaffed(),
		Status:       consts.PositionStatusEnabled,
	}
	if err := s.repo.CreatePosition(ctx, position); err != nil {
		return nil, err
	}
	if err := s.repo.SyncPositionRoles(ctx, position.ID, roleIDs); err != nil {
		return nil, err
	}
	return &xadmin.OrganizationActionResp{Success: true, Action: "create_position"}, nil
}

func (s *service) UpdatePosition(ctx context.Context, req *xadmin.OrganizationUpdatePositionReq) (*xadmin.OrganizationActionResp, error) {
	if _, err := s.repo.GetPositionByID(ctx, req.GetId()); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetDepartmentByID(ctx, req.GetDepartmentId()); err != nil {
		return nil, err
	}
	roleIDs := normalizeRoleIDs(req.GetRoleIds())
	if len(roleIDs) > 0 {
		count, err := s.repo.CountValidRolesByIDs(ctx, roleIDs)
		if err != nil {
			return nil, err
		}
		if count != int64(len(roleIDs)) {
			return nil, xerr.NewBiz(xerr.CodeBadRequest, "org.role_not_found")
		}
	}
	if err := s.repo.UpdatePositionByID(ctx, req.GetId(), map[string]any{
		"name":          strings.TrimSpace(req.GetName()),
		"code":          strings.TrimSpace(req.GetCode()),
		"department_id": req.GetDepartmentId(),
		"level":         strings.TrimSpace(req.GetLevel()),
		"hc":            req.GetHc(),
		"staffed":       req.GetStaffed(),
		"status":        req.GetStatus(),
	}); err != nil {
		return nil, err
	}
	if err := s.repo.SyncPositionRoles(ctx, req.GetId(), roleIDs); err != nil {
		return nil, err
	}
	return &xadmin.OrganizationActionResp{Success: true, Action: "update_position"}, nil
}

func (s *service) UpdatePositionStatus(ctx context.Context, req *xadmin.OrganizationUpdatePositionStatusReq) (*xadmin.OrganizationActionResp, error) {
	if _, err := s.repo.GetPositionByID(ctx, req.GetId()); err != nil {
		return nil, err
	}
	status := consts.PositionStatusDisabled
	if req.GetEnabled() {
		status = consts.PositionStatusEnabled
	}
	if err := s.repo.UpdatePositionByID(ctx, req.GetId(), map[string]any{
		"status": status,
	}); err != nil {
		return nil, err
	}
	if status == consts.PositionStatusDisabled {
		if err := s.repo.RevokeSessionsByPositionID(ctx, req.GetId(), "position_disabled"); err != nil {
			return nil, err
		}
	}
	return &xadmin.OrganizationActionResp{Success: true, Action: "update_position_status"}, nil
}

func (s *service) DeletePosition(ctx context.Context, req *xadmin.OrganizationDeletePositionReq) (*xadmin.OrganizationActionResp, error) {
	if _, err := s.repo.GetPositionByID(ctx, req.GetId()); err != nil {
		return nil, err
	}
	if err := s.repo.SoftDeletePositionByID(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &xadmin.OrganizationActionResp{Success: true, Action: "delete_position"}, nil
}

func (s *service) ListUsers(ctx context.Context, req *xadmin.OrganizationUsersReq) (*xadmin.OrganizationUsersResp, error) {
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

	rows, total, err := s.repo.ListUsers(ctx, page, normalizeSortArgs(req.GetSort()), buildUserFilters(req))
	if err != nil {
		return nil, err
	}

	items := make([]*xadmin.OrganizationUserItem, 0, len(rows))
	for _, row := range rows {
		item := &xadmin.OrganizationUserItem{
			Uid:                row.UID,
			Username:           row.Username,
			DisplayName:        row.DisplayName,
			Avatar:             row.Avatar,
			Email:              row.Email,
			Phone:              row.Phone,
			AccountStatus:      userStatusText(row.Status),
			OnlineStatus:       onlineStatusText(row.ActiveSessionCount),
			ActiveSessionCount: row.ActiveSessionCount,
			LastLoginIp:        row.LastLoginIP,
			DepartmentId:       row.DepartmentID,
			DepartmentName:     row.DepartmentName,
			PositionId:         row.PositionID,
			PositionName:       row.PositionName,
			RoleNames:          parseCSVString(row.RoleNamesCSV),
		}
		if row.LastLoginAt != nil {
			item.LastLoginAt = timefmt.RFC3339Ptr(row.LastLoginAt)
		}
		items = append(items, item)
	}
	return &xadmin.OrganizationUsersResp{
		Items: items,
		Total: total,
		Page: &commpb.PageArgs{
			Pn: page.GetPn(),
			Ps: page.GetPs(),
		},
	}, nil
}

func mapDepartmentRow(row *organizationrepo.DepartmentRow) *xadmin.OrganizationDepartmentItem {
	item := &xadmin.OrganizationDepartmentItem{
		Id:            row.ID,
		ParentId:      row.ParentID,
		Name:          row.Name,
		Code:          row.Code,
		Status:        departmentStatusText(row.Status),
		MemberCount:   row.MemberCount,
		PositionCount: row.PositionCount,
	}
	if row.UpdatedAt != nil {
		item.UpdatedAt = timefmt.RFC3339Ptr(row.UpdatedAt)
	}
	return item
}

func mapPositionRow(row *organizationrepo.PositionRow) *xadmin.OrganizationPositionItem {
	item := &xadmin.OrganizationPositionItem{
		Id:             row.ID,
		Name:           row.Name,
		Code:           row.Code,
		DepartmentId:   row.DepartmentID,
		DepartmentName: row.DepartmentName,
		Level:          row.Level,
		Hc:             row.Hc,
		Staffed:        row.Staffed,
		RelatedCount:   row.RelatedCount,
		Status:         positionStatusText(row.Status),
		RoleIds:        parseCSVInt64(row.RoleIDsCSV),
		RoleNames:      parseCSVString(row.RoleNamesCSV),
	}
	if row.UpdatedAt != nil {
		item.UpdatedAt = timefmt.RFC3339Ptr(row.UpdatedAt)
	}
	return item
}

func departmentStatusText(status int32) string {
	if status == consts.DepartmentStatusEnabled {
		return "enabled"
	}
	return "disabled"
}

func positionStatusText(status int32) string {
	if status == consts.PositionStatusEnabled {
		return "enabled"
	}
	return "disabled"
}

func (s *service) ListUserSessions(ctx context.Context, req *xadmin.OrganizationUserSessionsReq) (*xadmin.OrganizationUserSessionsResp, error) {
	status, err := sessionStatusToDB(req.GetStatus())
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListSessionsByUID(ctx, req.GetUid(), status, int(req.GetPageSize()))
	if err != nil {
		return nil, err
	}
	items := make([]*xadmin.AuthSessionItem, 0, len(rows))
	for _, row := range rows {
		item := &xadmin.AuthSessionItem{
			SessionId:     row.SessionID,
			Status:        row.Status,
			LoginIp:       row.LoginIP,
			UserAgent:     row.UserAgent,
			ExpiredAt:     timefmt.RFC3339(row.ExpiredAt),
			RevokedReason: row.RevokedReason,
		}
		if row.LastSeenAt != nil {
			item.LastSeenAt = timefmt.RFC3339Ptr(row.LastSeenAt)
		}
		if row.RevokedAt != nil {
			item.RevokedAt = timefmt.RFC3339Ptr(row.RevokedAt)
		}
		items = append(items, item)
	}
	return &xadmin.OrganizationUserSessionsResp{Items: items}, nil
}

func (s *service) CreateUser(ctx context.Context, req *xadmin.OrganizationCreateUserReq) (*xadmin.OrganizationActionResp, error) {
	if req.GetDepartmentId() > 0 {
		if _, err := s.repo.GetDepartmentByID(ctx, req.GetDepartmentId()); err != nil {
			return nil, err
		}
	}
	if req.GetPositionId() > 0 {
		if req.GetDepartmentId() <= 0 {
			return nil, xerr.NewBiz(xerr.CodeBadRequest, "org.position_requires_department")
		}
		position, err := s.repo.GetPositionByID(ctx, req.GetPositionId())
		if err != nil {
			return nil, err
		}
		if position.DepartmentID != req.GetDepartmentId() {
			return nil, xerr.NewBiz(xerr.CodeBadRequest, "org.position_not_in_department")
		}
		if position.Status != consts.PositionStatusEnabled {
			return nil, xerr.NewBiz(xerr.CodeBadRequest, "org.position_disabled")
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, xerr.NewWithError(xerr.CodeInternalError, err, "hash password")
	}
	uid, err := s.repo.NextUID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateUser(ctx, &model.AdminUser{
		UID:          uid,
		Username:     strings.TrimSpace(req.GetUsername()),
		PasswordHash: string(hash),
		DisplayName:  strings.TrimSpace(req.GetDisplayName()),
		Email:        strings.TrimSpace(req.GetEmail()),
		Phone:        strings.TrimSpace(req.GetPhone()),
		Status:       req.GetStatus(),
		DepartmentID: req.GetDepartmentId(),
		PositionID:   req.GetPositionId(),
	}); err != nil {
		return nil, err
	}
	return &xadmin.OrganizationActionResp{Success: true, Action: "create_user"}, nil
}

func (s *service) UpdateUser(ctx context.Context, req *xadmin.OrganizationUpdateUserReq) (*xadmin.OrganizationActionResp, error) {
	user, err := s.repo.GetUserByUID(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	if user.Status == consts.UserStatusDeactivated {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "org.account_deactivated_readonly")
	}
	if req.GetDepartmentId() > 0 {
		if _, err := s.repo.GetDepartmentByID(ctx, req.GetDepartmentId()); err != nil {
			return nil, err
		}
	}
	if req.GetPositionId() > 0 {
		if req.GetDepartmentId() <= 0 {
			return nil, xerr.NewBiz(xerr.CodeBadRequest, "org.position_requires_department")
		}
		position, err := s.repo.GetPositionByID(ctx, req.GetPositionId())
		if err != nil {
			return nil, err
		}
		if position.DepartmentID != req.GetDepartmentId() {
			return nil, xerr.NewBiz(xerr.CodeBadRequest, "org.position_not_in_department")
		}
		if position.Status != consts.PositionStatusEnabled && position.ID != user.PositionID {
			return nil, xerr.NewBiz(xerr.CodeBadRequest, "org.position_disabled")
		}
	}

	updates := map[string]any{
		"display_name":  strings.TrimSpace(req.GetDisplayName()),
		"avatar":        strings.TrimSpace(req.GetAvatar()),
		"email":         strings.TrimSpace(req.GetEmail()),
		"phone":         strings.TrimSpace(req.GetPhone()),
		"status":        req.GetStatus(),
		"department_id": req.GetDepartmentId(),
		"position_id":   req.GetPositionId(),
	}
	if req.GetStatus() == consts.UserStatusDeactivated {
		updates["deactivated_at"] = time.Now()
	} else {
		updates["deactivated_at"] = nil
	}
	if err := s.repo.UpdateUserByUID(ctx, req.GetUid(), updates); err != nil {
		return nil, err
	}
	if req.GetStatus() == consts.UserStatusDisabled || req.GetStatus() == consts.UserStatusDeactivated {
		if err := s.repo.RevokeAllSessionsByUID(ctx, req.GetUid(), "user_disabled_or_deactivated"); err != nil {
			return nil, err
		}
	}
	return &xadmin.OrganizationActionResp{Success: true, Action: "update_user"}, nil
}

func (s *service) BatchTransferUsers(ctx context.Context, req *xadmin.OrganizationBatchTransferUsersReq) (*xadmin.OrganizationActionResp, error) {
	if _, err := s.repo.GetDepartmentByID(ctx, req.GetDepartmentId()); err != nil {
		return nil, err
	}
	position, err := s.repo.GetPositionByID(ctx, req.GetPositionId())
	if err != nil {
		return nil, err
	}
	if position.DepartmentID != req.GetDepartmentId() {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "org.position_not_in_department")
	}
	if position.Status != consts.PositionStatusEnabled {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "org.position_disabled")
	}
	if err := s.repo.BatchUpdateUsersPosition(ctx, req.GetUids(), req.GetDepartmentId(), req.GetPositionId()); err != nil {
		return nil, err
	}
	return &xadmin.OrganizationActionResp{Success: true, Action: "batch_transfer_users"}, nil
}

func (s *service) ResetPassword(ctx context.Context, req *xadmin.OrganizationResetPasswordReq) (*xadmin.OrganizationActionResp, error) {
	user, err := s.repo.GetUserByUID(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(strings.TrimSpace(user.Username), "admin") {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "org.admin_no_reset_pwd")
	}
	password, err := generateRandomPassword(12)
	if err != nil {
		return nil, xerr.NewWithError(xerr.CodeInternalError, err, "generate random password")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, xerr.NewWithError(xerr.CodeInternalError, err, "hash reset password")
	}
	if err := s.repo.UpdateUserByUID(ctx, req.GetUid(), map[string]any{
		"password_hash": string(hash),
	}); err != nil {
		return nil, err
	}
	if err := s.repo.RevokeAllSessionsByUID(ctx, req.GetUid(), "reset_password"); err != nil {
		return nil, err
	}
	return &xadmin.OrganizationActionResp{
		Success:      true,
		Action:       "reset_password",
		TempPassword: password,
	}, nil
}

func (s *service) DeleteUser(ctx context.Context, req *xadmin.OrganizationDeleteUserReq) (*xadmin.OrganizationActionResp, error) {
	user, err := s.repo.GetUserByUID(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(strings.TrimSpace(user.Username), "admin") {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "org.admin_no_delete")
	}
	if user.Status != consts.UserStatusDeactivated {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "org.only_delete_deactivated")
	}
	if user.DeactivatedAt == nil || user.DeactivatedAt.After(time.Now().AddDate(0, -3, 0)) {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "org.delete_cooldown_3m")
	}
	if err := s.repo.SoftDeleteUserByUID(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &xadmin.OrganizationActionResp{Success: true, Action: "delete_user"}, nil
}

func (s *service) ImportUsers(ctx context.Context, req *xadmin.OrganizationImportUsersReq) (*xadmin.OrganizationActionResp, error) {
	for _, item := range req.GetItems() {
		if item == nil {
			continue
		}
		password := strings.TrimSpace(item.GetPassword())
		if password == "" {
			password = "Reset@123456"
		}
		if _, err := s.CreateUser(ctx, &xadmin.OrganizationCreateUserReq{
			Username:    strings.TrimSpace(item.GetUsername()),
			Password:    password,
			DisplayName: strings.TrimSpace(item.GetDisplayName()),
			Email:       strings.TrimSpace(item.GetEmail()),
			Phone:       strings.TrimSpace(item.GetPhone()),
			Status:      item.GetStatus(),
		}); err != nil {
			// Continue import for best effort; duplicates should not stop whole request.
			continue
		}
	}
	return &xadmin.OrganizationActionResp{Success: true, Action: "import_users"}, nil
}

func (s *service) ExportUsers(ctx context.Context, req *xadmin.OrganizationUsersReq) ([]byte, error) {
	rows, err := s.repo.ExportUsers(ctx, buildUserFilters(req))
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(nil)
	writer := csv.NewWriter(buf)
	_ = writer.Write([]string{"uid", "username", "display_name", "email", "phone", "账号状态", "在线状态", "last_login_at", "last_login_ip"})
	for _, row := range rows {
		lastLoginAt := ""
		if row.LastLoginAt != nil {
			lastLoginAt = timefmt.RFC3339Ptr(row.LastLoginAt)
		}
		_ = writer.Write([]string{
			fmt.Sprintf("%d", row.UID),
			row.Username,
			row.DisplayName,
			row.Email,
			row.Phone,
			userStatusCN(row.Status),
			onlineStatusCN(row.ActiveSessionCount),
			lastLoginAt,
			row.LastLoginIP,
		})
	}
	writer.Flush()
	return buf.Bytes(), nil
}

func onlineStatusText(activeSessionCount int32) string {
	if activeSessionCount > 0 {
		return "online"
	}
	return "offline"
}

func onlineStatusCN(activeSessionCount int32) string {
	if activeSessionCount > 0 {
		return "在线"
	}
	return "离线"
}

func userStatusText(status int32) string {
	switch status {
	case consts.UserStatusActive:
		return "active"
	case consts.UserStatusDisabled:
		return "disabled"
	case consts.UserStatusDeactivated:
		return "deactivated"
	default:
		return "disabled"
	}
}

func userStatusCN(status int32) string {
	switch status {
	case consts.UserStatusActive:
		return "启用"
	case consts.UserStatusDisabled:
		return "停用"
	case consts.UserStatusDeactivated:
		return "已注销"
	default:
		return "未知"
	}
}

func normalizeSortArgs(input []*commpb.SortArgs) []*commpb.SortArgs {
	if len(input) == 0 {
		return nil
	}
	fieldMap := map[string]string{
		"uid":                  "u.uid",
		"username":             "u.username",
		"display_name":         "u.display_name",
		"status":               "u.status",
		"active_session_count": "active_session_count",
		"last_login_at":        "u.last_login_at",
		"created_at":           "u.created_at",
	}
	out := make([]*commpb.SortArgs, 0, len(input))
	for _, item := range input {
		if item == nil {
			continue
		}
		mapped, ok := fieldMap[item.GetOrderField()]
		if !ok {
			continue
		}
		out = append(out, &commpb.SortArgs{
			OrderField: mapped,
			OrderType:  item.GetOrderType(),
		})
	}
	return out
}

func normalizePositionSortArgs(input []*commpb.SortArgs) []*commpb.SortArgs {
	if len(input) == 0 {
		return nil
	}
	fieldMap := map[string]string{
		"id":         "p.id",
		"name":       "p.name",
		"level":      "p.level",
		"status":     "p.status",
		"updated_at": "p.updated_at",
		"created_at": "p.created_at",
	}
	out := make([]*commpb.SortArgs, 0, len(input))
	for _, item := range input {
		if item == nil {
			continue
		}
		mapped, ok := fieldMap[item.GetOrderField()]
		if !ok {
			continue
		}
		out = append(out, &commpb.SortArgs{
			OrderField: mapped,
			OrderType:  item.GetOrderType(),
		})
	}
	return out
}

func buildUserFilters(req *xadmin.OrganizationUsersReq) organizationrepo.UserFilters {
	var statusPtr *int32
	switch req.GetStatus() {
	case xadmin.OrganizationUserFilterStatus_ORGANIZATION_USER_FILTER_STATUS_ACTIVE:
		v := consts.UserStatusActive
		statusPtr = &v
	case xadmin.OrganizationUserFilterStatus_ORGANIZATION_USER_FILTER_STATUS_DISABLED:
		v := consts.UserStatusDisabled
		statusPtr = &v
	case xadmin.OrganizationUserFilterStatus_ORGANIZATION_USER_FILTER_STATUS_DEACTIVATED:
		v := consts.UserStatusDeactivated
		statusPtr = &v
	}
	var createdFrom *time.Time
	if v := strings.TrimSpace(req.GetCreatedFrom()); v != "" {
		if t, err := parseDateTimeFilter(v); err == nil {
			createdFrom = &t
		}
	}
	var createdTo *time.Time
	if v := strings.TrimSpace(req.GetCreatedTo()); v != "" {
		if t, err := parseDateTimeFilter(v); err == nil {
			createdTo = &t
		}
	}
	var departmentID *int64
	if req.GetDepartmentId() > 0 {
		v := req.GetDepartmentId()
		departmentID = &v
	}
	var positionID *int64
	if req.GetPositionId() > 0 {
		v := req.GetPositionId()
		positionID = &v
	}
	return organizationrepo.UserFilters{
		Keyword:      strings.TrimSpace(req.GetKeyword()),
		Phone:        strings.TrimSpace(req.GetPhone()),
		Status:       statusPtr,
		DepartmentID: departmentID,
		PositionID:   positionID,
		CreatedFrom:  createdFrom,
		CreatedTo:    createdTo,
	}
}

func parseDateTimeFilter(input string) (time.Time, error) {
	layouts := []string{
		time.DateTime,
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if value, err := time.ParseInLocation(layout, strings.TrimSpace(input), time.Local); err == nil {
			return value, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid datetime: %s", input)
}

func buildPositionFilters(req *xadmin.OrganizationPositionsReq) organizationrepo.PositionFilters {
	var departmentID *int64
	if req.GetDepartmentId() > 0 {
		v := req.GetDepartmentId()
		departmentID = &v
	}
	var statusPtr *int32
	switch req.GetStatus() {
	case xadmin.OrganizationPositionFilterStatus_ORGANIZATION_POSITION_FILTER_STATUS_ENABLED:
		v := consts.PositionStatusEnabled
		statusPtr = &v
	case xadmin.OrganizationPositionFilterStatus_ORGANIZATION_POSITION_FILTER_STATUS_DISABLED:
		v := consts.PositionStatusDisabled
		statusPtr = &v
	}
	return organizationrepo.PositionFilters{
		Keyword:      strings.TrimSpace(req.GetKeyword()),
		DepartmentID: departmentID,
		Level:        strings.TrimSpace(req.GetLevel()),
		Status:       statusPtr,
	}
}

func sessionStatusToDB(status xadmin.AuthSessionStatus) (string, error) {
	switch status {
	case xadmin.AuthSessionStatus_AUTH_SESSION_STATUS_UNSPECIFIED:
		return "", nil
	case xadmin.AuthSessionStatus_AUTH_SESSION_STATUS_ACTIVE:
		return consts.SessionStatusActive, nil
	case xadmin.AuthSessionStatus_AUTH_SESSION_STATUS_REVOKED:
		return consts.SessionStatusRevoked, nil
	case xadmin.AuthSessionStatus_AUTH_SESSION_STATUS_EXPIRED:
		return consts.SessionStatusExpired, nil
	default:
		return "", xerr.NewBiz(xerr.CodeBadRequest, "account.invalid_status")
	}
}

func generateRandomPassword(length int) (string, error) {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789!@#$%^&*"
	if length <= 0 {
		length = 12
	}
	buf := make([]byte, length)
	randBytes := make([]byte, length)
	if _, err := rand.Read(randBytes); err != nil {
		return "", err
	}
	for i := range buf {
		buf[i] = chars[int(randBytes[i])%len(chars)]
	}
	return string(buf), nil
}

func normalizeRoleIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	uniq := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		uniq[id] = struct{}{}
	}
	out := make([]int64, 0, len(uniq))
	for id := range uniq {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func parseCSVInt64(raw string) []int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		v, err := strconv.ParseInt(part, 10, 64)
		if err != nil || v <= 0 {
			continue
		}
		out = append(out, v)
	}
	return normalizeRoleIDs(out)
}

func parseCSVString(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}
