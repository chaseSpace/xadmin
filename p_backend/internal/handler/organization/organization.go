package organization

import (
	"strconv"
	"strings"

	"monorepo/internal/middleware"
	organizationrepo "monorepo/internal/repo/organization"
	organizationsvc "monorepo/internal/service/organization"
	"monorepo/internal/support/auditlog"
	"monorepo/pkg/consts"
	"monorepo/pkg/xerr"
	"monorepo/pkg/xfiber"
	xadmin "monorepo/proto/xadminpb"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/encoding/protojson"
	proto2 "google.golang.org/protobuf/proto"
)

type Handler struct {
	svc  organizationsvc.Service
	repo *organizationrepo.Repo
}

func NewHandler() *Handler {
	return &Handler{svc: organizationsvc.NewService(), repo: organizationrepo.NewRepo()}
}

func NewHandlerWithService(svc organizationsvc.Service) *Handler {
	return &Handler{svc: svc, repo: organizationrepo.NewRepo()}
}

func RegisterRoutes(prefix string, parent fiber.Router, authMW fiber.Handler) {
	group := parent.Group(prefix)
	handler := NewHandler()
	edit := middleware.RequirePermission

	group.Get("/departments/tree", authMW, handler.DepartmentsTree)
	group.Get("/departments/:id", authMW, handler.Department)
	group.Post("/departments", authMW, edit("organization.departments.edit"), handler.CreateDepartment)
	group.Put("/departments/:id", authMW, edit("organization.departments.edit"), handler.UpdateDepartment)
	group.Post("/departments/:id/status", authMW, edit("organization.departments.edit"), handler.UpdateDepartmentStatus)
	group.Delete("/departments/:id", authMW, edit("organization.departments.delete"), handler.DeleteDepartment)
	group.Get("/positions", authMW, handler.Positions)
	group.Get("/positions/:id", authMW, handler.Position)
	group.Post("/positions", authMW, edit("organization.positions.edit"), handler.CreatePosition)
	group.Put("/positions/:id", authMW, edit("organization.positions.edit"), handler.UpdatePosition)
	group.Post("/positions/:id/status", authMW, edit("organization.positions.edit"), handler.UpdatePositionStatus)
	group.Delete("/positions/:id", authMW, edit("organization.positions.delete"), handler.DeletePosition)

	group.Get("/users", authMW, handler.Users)
	group.Post("/users", authMW, edit("organization.users.edit"), handler.CreateUser)
	group.Post("/users/transfer-position", authMW, edit("organization.users.edit"), handler.BatchTransferUsers)
	group.Post("/users/import", authMW, edit("organization.users.edit"), handler.ImportUsers)
	group.Get("/users/export", authMW, handler.ExportUsers)
	group.Delete("/users/:uid", authMW, edit("organization.users.delete"), handler.DeleteUser)
	group.Put("/users/:uid", authMW, edit("organization.users.edit"), handler.UpdateUser)
	group.Post("/users/:uid/reset_password", authMW, edit("organization.users.edit"), handler.ResetPassword)
	group.Get("/users/:uid/sessions", authMW, handler.UserSessions)
}

func (h *Handler) Positions(c *fiber.Ctx) error {
	page, sort, err := xfiber.ParsePageSortQuery(c, 1, 10)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.OrganizationPositionsReq{Page: page, Sort: sort}
	req.Keyword = strings.TrimSpace(c.Query("keyword"))
	req.Level = strings.TrimSpace(c.Query("level"))
	if departmentIDRaw := strings.TrimSpace(c.Query("department_id")); departmentIDRaw != "" {
		departmentID, err := strconv.ParseInt(departmentIDRaw, 10, 64)
		if err != nil {
			return xfiber.StdResponse(c, nil, err)
		}
		req.DepartmentId = departmentID
	}
	switch strings.ToLower(strings.TrimSpace(c.Query("status"))) {
	case "enabled", "1":
		req.Status = xadmin.OrganizationPositionFilterStatus_ORGANIZATION_POSITION_FILTER_STATUS_ENABLED
	case "disabled", "0":
		req.Status = xadmin.OrganizationPositionFilterStatus_ORGANIZATION_POSITION_FILTER_STATUS_DISABLED
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.ListPositions(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_position_list",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) Position(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.OrganizationPositionDetailReq{Id: id}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.GetPosition(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_position_detail",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) CreatePosition(c *fiber.Ctx) error {
	req := &xadmin.OrganizationCreatePositionReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.CreatePosition(c.UserContext(), req)
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) UpdatePosition(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.OrganizationUpdatePositionReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req.Id = id
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.UpdatePosition(c.UserContext(), req)
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) UpdatePositionStatus(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.OrganizationUpdatePositionStatusReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req.Id = id
	if !req.GetEnabled() {
		if authUser := middleware.GetUserEntity(c); authUser != nil && authUser.User != nil {
			current, err := h.repo.GetUserByUID(c.UserContext(), authUser.User.UID)
			if err == nil && current.PositionID == id {
				return xfiber.StdResponse(c, nil, xerr.NewBiz(xerr.CodeBadRequest, "org.cannot_disable_own_position"))
			}
		}
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.UpdatePositionStatus(c.UserContext(), req)
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) DeletePosition(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.OrganizationDeletePositionReq{Id: id}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.DeletePosition(c.UserContext(), req)
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) DepartmentsTree(c *fiber.Ctx) error {
	resp, err := h.svc.GetDepartmentsTree(c.UserContext())
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_department_tree",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) Department(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.OrganizationDepartmentDetailReq{Id: id}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.GetDepartment(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_department_detail",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "id=" + strconv.FormatInt(id, 10),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) CreateDepartment(c *fiber.Ctx) error {
	req := &xadmin.OrganizationCreateDepartmentReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.CreateDepartment(c.UserContext(), req)
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) UpdateDepartment(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.OrganizationUpdateDepartmentReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req.Id = id
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.UpdateDepartment(c.UserContext(), req)
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) UpdateDepartmentStatus(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.OrganizationUpdateDepartmentStatusReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req.Id = id
	if !req.GetEnabled() {
		if authUser := middleware.GetUserEntity(c); authUser != nil && authUser.User != nil {
			current, err := h.repo.GetUserByUID(c.UserContext(), authUser.User.UID)
			if err == nil && current.DepartmentID == id {
				return xfiber.StdResponse(c, nil, xerr.NewBiz(xerr.CodeBadRequest, "org.cannot_disable_own_department"))
			}
		}
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.UpdateDepartmentStatus(c.UserContext(), req)
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) DeleteDepartment(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.OrganizationDeleteDepartmentReq{Id: id, Force: strings.EqualFold(strings.TrimSpace(c.Query("force")), "true") || strings.TrimSpace(c.Query("force")) == "1"}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.DeleteDepartment(c.UserContext(), req)
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) BatchTransferUsers(c *fiber.Ctx) error {
	req := &xadmin.OrganizationBatchTransferUsersReq{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(c.Body(), req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.BatchTransferUsers(c.UserContext(), req)
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) Users(c *fiber.Ctx) error {
	page, sort, err := xfiber.ParsePageSortQuery(c, 1, 10)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.OrganizationUsersReq{Page: page, Sort: sort}
	req.Keyword = strings.TrimSpace(c.Query("keyword"))
	req.Phone = strings.TrimSpace(c.Query("phone"))
	req.CreatedFrom = strings.TrimSpace(c.Query("created_from"))
	req.CreatedTo = strings.TrimSpace(c.Query("created_to"))
	if departmentIDRaw := strings.TrimSpace(c.Query("department_id")); departmentIDRaw != "" {
		departmentID, err := strconv.ParseInt(departmentIDRaw, 10, 64)
		if err != nil {
			return xfiber.StdResponse(c, nil, err)
		}
		req.DepartmentId = departmentID
	}
	if positionIDRaw := strings.TrimSpace(c.Query("position_id")); positionIDRaw != "" {
		positionID, err := strconv.ParseInt(positionIDRaw, 10, 64)
		if err != nil {
			return xfiber.StdResponse(c, nil, err)
		}
		req.PositionId = positionID
	}
	switch strings.ToLower(strings.TrimSpace(c.Query("status"))) {
	case "active", "1":
		req.Status = xadmin.OrganizationUserFilterStatus_ORGANIZATION_USER_FILTER_STATUS_ACTIVE
	case "disabled", "0":
		req.Status = xadmin.OrganizationUserFilterStatus_ORGANIZATION_USER_FILTER_STATUS_DISABLED
	case "deactivated", "2":
		req.Status = xadmin.OrganizationUserFilterStatus_ORGANIZATION_USER_FILTER_STATUS_DEACTIVATED
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.ListUsers(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_user_list",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) CreateUser(c *fiber.Ctx) error {
	req := &xadmin.OrganizationCreateUserReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.CreateUser(c.UserContext(), req)
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) ImportUsers(c *fiber.Ctx) error {
	req := &xadmin.OrganizationImportUsersReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.ImportUsers(c.UserContext(), req)
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) ExportUsers(c *fiber.Ctx) error {
	page, sort, err := xfiber.ParsePageSortQuery(c, 1, 10)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.OrganizationUsersReq{Page: page, Sort: sort}
	req.Keyword = strings.TrimSpace(c.Query("keyword"))
	req.Phone = strings.TrimSpace(c.Query("phone"))
	req.CreatedFrom = strings.TrimSpace(c.Query("created_from"))
	req.CreatedTo = strings.TrimSpace(c.Query("created_to"))
	if departmentIDRaw := strings.TrimSpace(c.Query("department_id")); departmentIDRaw != "" {
		departmentID, err := strconv.ParseInt(departmentIDRaw, 10, 64)
		if err != nil {
			return xfiber.StdResponse(c, nil, err)
		}
		req.DepartmentId = departmentID
	}
	if positionIDRaw := strings.TrimSpace(c.Query("position_id")); positionIDRaw != "" {
		positionID, err := strconv.ParseInt(positionIDRaw, 10, 64)
		if err != nil {
			return xfiber.StdResponse(c, nil, err)
		}
		req.PositionId = positionID
	}
	switch strings.ToLower(strings.TrimSpace(c.Query("status"))) {
	case "active", "1":
		req.Status = xadmin.OrganizationUserFilterStatus_ORGANIZATION_USER_FILTER_STATUS_ACTIVE
	case "disabled", "0":
		req.Status = xadmin.OrganizationUserFilterStatus_ORGANIZATION_USER_FILTER_STATUS_DISABLED
	case "deactivated", "2":
		req.Status = xadmin.OrganizationUserFilterStatus_ORGANIZATION_USER_FILTER_STATUS_DEACTIVATED
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	data, err := h.svc.ExportUsers(c.UserContext(), req)
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	_ = auditlog.Log(c.UserContext(), auditlog.Meta{
		UID:       middleware.GetUID(c),
		Action:    "export_users",
		Result:    "success",
		TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
		SourceIP:  c.IP(),
		UserAgent: c.Get("User-Agent"),
	})
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=organization_users.csv")
	return c.Status(fiber.StatusOK).Send(data)
}

func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	uid, err := strconv.Atoi(strings.TrimSpace(c.Params("uid")))
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.OrganizationUpdateUserReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req.Uid = int32(uid)
	if authUser := middleware.GetUserEntity(c); authUser != nil && authUser.User != nil {
		if req.GetUid() == authUser.User.UID && (req.GetStatus() == consts.UserStatusDisabled || req.GetStatus() == consts.UserStatusDeactivated) {
			return xfiber.StdResponse(c, nil, xerr.NewBiz(xerr.CodeBadRequest, "org.cannot_disable_self"))
		}
	}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.UpdateUser(c.UserContext(), req)
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	uid, err := strconv.Atoi(strings.TrimSpace(c.Params("uid")))
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.OrganizationDeleteUserReq{Uid: int32(uid)}
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.DeleteUser(c.UserContext(), req)
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	uid, err := strconv.Atoi(strings.TrimSpace(c.Params("uid")))
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.OrganizationResetPasswordReq{}
	if err := parseProtoRequest(c, req); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req.Uid = int32(uid)
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.ResetPassword(c.UserContext(), req)
	return xfiber.StdResponse(c, resp, err)
}

func (h *Handler) UserSessions(c *fiber.Ctx) error {
	uid, err := strconv.Atoi(strings.TrimSpace(c.Params("uid")))
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req := &xadmin.OrganizationUserSessionsReq{Uid: int32(uid), PageSize: 10}
	if pageSizeRaw := strings.TrimSpace(c.Query("page_size")); pageSizeRaw != "" {
		pageSize, err := strconv.Atoi(pageSizeRaw)
		if err != nil {
			return xfiber.StdResponse(c, nil, err)
		}
		req.PageSize = int32(pageSize)
	}
	status, err := parseSessionStatusQuery(c.Query("status"))
	if err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	req.Status = status
	if err := req.Validate(); err != nil {
		return xfiber.StdResponse(c, nil, err)
	}
	resp, err := h.svc.ListUserSessions(c.UserContext(), req)
	if err == nil {
		_ = auditlog.Log(c.UserContext(), auditlog.Meta{
			UID:       middleware.GetUID(c),
			Action:    "view_user_sessions",
			Result:    "success",
			TraceID:   strings.TrimSpace(c.Get("X-Trace-ID")),
			SourceIP:  c.IP(),
			UserAgent: c.Get("User-Agent"),
			Detail:    "target_uid=" + strconv.Itoa(uid),
		})
	}
	return xfiber.StdResponse(c, resp, err)
}

func parseSessionStatusQuery(raw string) (xadmin.AuthSessionStatus, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return xadmin.AuthSessionStatus_AUTH_SESSION_STATUS_UNSPECIFIED, nil
	case consts.SessionStatusActive:
		return xadmin.AuthSessionStatus_AUTH_SESSION_STATUS_ACTIVE, nil
	case consts.SessionStatusRevoked:
		return xadmin.AuthSessionStatus_AUTH_SESSION_STATUS_REVOKED, nil
	case consts.SessionStatusExpired:
		return xadmin.AuthSessionStatus_AUTH_SESSION_STATUS_EXPIRED, nil
	default:
		return xadmin.AuthSessionStatus_AUTH_SESSION_STATUS_UNSPECIFIED, xerr.NewBiz(xerr.CodeBadRequest, "account.invalid_status")
	}
}

func parseProtoRequest(c *fiber.Ctx, req proto2.Message) error {
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(c.Body(), req); err != nil {
		return err
	}
	return nil
}
