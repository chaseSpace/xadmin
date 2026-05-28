package auth

import (
	"context"
	"strconv"
	"strings"
	"time"

	"monorepo/config"
	"monorepo/internal/model"
	authrepo "monorepo/internal/repo/auth"
	"monorepo/internal/support/requestmeta"
	"monorepo/internal/support/timefmt"
	"monorepo/pkg/auth"
	"monorepo/pkg/consts"
	"monorepo/pkg/xerr"
	xadmin "monorepo/proto/xadminpb"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Login(ctx context.Context, req *xadmin.AuthLoginReq, ip, userAgent, traceID string) (*xadmin.AuthLoginResp, error)
	Logout(ctx context.Context, uid int32, sessionID, ip, userAgent, traceID string) (*xadmin.AuthActionResp, error)
	LogoutOthers(ctx context.Context, uid int32, sessionID, ip, userAgent, traceID string) (*xadmin.AuthActionResp, error)
	ForceLogout(ctx context.Context, operatorUID int32, req *xadmin.AuthForceLogoutReq, ip, userAgent, traceID string) (*xadmin.AuthActionResp, error)
	Deactivate(ctx context.Context, operatorUID int32, req *xadmin.AuthDeactivateReq, ip, userAgent, traceID string) (*xadmin.AuthActionResp, error)
	ListSessions(ctx context.Context, uid int32, req *xadmin.AuthSessionsReq) (*xadmin.AuthSessionsResp, error)
	IsSessionActive(ctx context.Context, uid int32, sessionID, tokenHash string) (bool, error)
}

type service struct {
	repo *authrepo.Repo
}

func NewService() Service {
	return &service{repo: authrepo.NewRepo()}
}

func NewServiceWithRepo(repo *authrepo.Repo) Service {
	return &service{repo: repo}
}

func (s *service) Login(ctx context.Context, req *xadmin.AuthLoginReq, ip, userAgent, traceID string) (*xadmin.AuthLoginResp, error) {
	req.Username = strings.TrimSpace(req.GetUsername())
	if req.GetUsername() == "" || req.GetPassword() == "" {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "auth.credentials_required")
	}

	user, err := s.repo.GetActiveUserByUsername(ctx, req.GetUsername())
	if err != nil {
		_ = s.audit(ctx, 0, "login_failed", "failed", ip, userAgent, traceID, "user not found")
		return nil, xerr.NewBiz(xerr.CodeUnauthorized, "auth.invalid_credentials")
	}

	if user.Status == consts.UserStatusDisabled || user.Status == consts.UserStatusDeactivated {
		detail := "account disabled"
		bizCode := "auth.account_disabled"
		if user.Status == consts.UserStatusDeactivated {
			detail = "account deactivated"
			bizCode = "auth.account_deactivated"
		}
		_ = s.audit(ctx, user.UID, "login_failed", "failed", ip, userAgent, traceID, detail)
		return nil, xerr.NewBiz(xerr.CodeForbidden, bizCode)
	}

	if user.DepartmentID > 0 {
		departmentStatus, exists, err := s.repo.GetDepartmentStatusByID(ctx, user.DepartmentID)
		if err != nil {
			return nil, err
		}
		if !exists || departmentStatus == consts.DepartmentStatusDisabled {
			_ = s.audit(ctx, user.UID, "login_failed", "failed", ip, userAgent, traceID, "department disabled")
			return nil, xerr.NewBiz(xerr.CodeForbidden, "auth.department_disabled")
		}
	}
	if user.PositionID > 0 {
		positionStatus, exists, err := s.repo.GetPositionStatusByID(ctx, user.PositionID)
		if err != nil {
			return nil, err
		}
		if !exists || positionStatus == consts.PositionStatusDisabled {
			_ = s.audit(ctx, user.UID, "login_failed", "failed", ip, userAgent, traceID, "position disabled")
			return nil, xerr.NewBiz(xerr.CodeForbidden, "auth.position_disabled")
		}
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.GetPassword())); err != nil {
		_ = s.audit(ctx, user.UID, "login_failed", "failed", ip, userAgent, traceID, "password mismatch")
		return nil, xerr.NewBiz(xerr.CodeUnauthorized, "auth.invalid_credentials")
	}

	limitSingleLogin, err := s.repo.GetSingleLoginSetting(ctx, user.UID)
	if err != nil {
		return nil, err
	}
	if limitSingleLogin {
		if err = s.repo.RevokeAllSessionsByUID(ctx, user.UID, "single_login_limit"); err != nil {
			return nil, err
		}
	}

	sessionID := uuid.NewString()
	token, err := auth.IssueTokenSignature(ctx, user.UID, sessionID)
	if err != nil {
		return nil, err
	}

	tokenHash := auth.HashToken(token)
	expAt := time.Now().Add(tokenTTL(config.GetConfig().App.Auth.TokenTTLMinutes, config.GetConfig().App.Auth.TokenTTLHours))
	now := time.Now()
	if err = s.repo.CreateSession(ctx, &model.AdminUserSession{
		SessionID:     sessionID,
		UID:           user.UID,
		TokenHash:     tokenHash,
		Status:        consts.SessionStatusActive,
		LoginIP:       ip,
		UserAgent:     userAgent,
		LastSeenAt:    &now,
		ExpiredAt:     expAt,
		RevokedReason: "",
	}); err != nil {
		return nil, err
	}

	_ = s.repo.UpdateLoginMeta(ctx, user.UID, ip, now)
	_ = s.audit(ctx, user.UID, "login_success", "success", ip, userAgent, traceID, "")

	return &xadmin.AuthLoginResp{
		AccessToken: token,
		ExpiresAt:   timefmt.RFC3339(expAt),
		Uid:         user.UID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Avatar:      user.Avatar,
		SessionId:   sessionID,
	}, nil
}

func (s *service) Logout(ctx context.Context, uid int32, sessionID, ip, userAgent, traceID string) (*xadmin.AuthActionResp, error) {
	if uid <= 0 || strings.TrimSpace(sessionID) == "" {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "auth.params_missing")
	}
	if err := s.repo.RevokeSession(ctx, uid, sessionID, "logout"); err != nil {
		return nil, err
	}
	_ = s.audit(ctx, uid, "logout", "success", ip, userAgent, traceID, "")
	return &xadmin.AuthActionResp{
		Success: true,
		Action:  "logout",
	}, nil
}

func (s *service) LogoutOthers(ctx context.Context, uid int32, sessionID, ip, userAgent, traceID string) (*xadmin.AuthActionResp, error) {
	if uid <= 0 || strings.TrimSpace(sessionID) == "" {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "auth.params_invalid")
	}
	if err := s.repo.RevokeOtherSessionsByUID(ctx, uid, sessionID, "logout_others"); err != nil {
		return nil, err
	}
	_ = s.audit(ctx, uid, "logout_others", "success", ip, userAgent, traceID, "")
	return &xadmin.AuthActionResp{
		Success: true,
		Action:  "logout_others",
	}, nil
}

func (s *service) ForceLogout(ctx context.Context, operatorUID int32, req *xadmin.AuthForceLogoutReq, ip, userAgent, traceID string) (*xadmin.AuthActionResp, error) {
	targetUID := req.GetTargetUid()
	if operatorUID <= 0 || targetUID <= 0 {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "auth.target_invalid")
	}
	if operatorUID == targetUID {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "auth.cannot_force_logout_self")
	}
	targetUser, err := s.repo.GetUserByUID(ctx, targetUID)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(strings.TrimSpace(targetUser.Username), "admin") {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "auth.admin_protected")
	}
	if err := s.repo.RevokeAllSessionsByUID(ctx, targetUID, "force_logout"); err != nil {
		return nil, err
	}
	_ = s.audit(ctx, targetUID, "force_logout", "success", ip, userAgent, traceID, "operator_uid="+int32ToString(operatorUID))
	return &xadmin.AuthActionResp{
		Success: true,
		Action:  "force_logout",
	}, nil
}

func (s *service) Deactivate(ctx context.Context, operatorUID int32, req *xadmin.AuthDeactivateReq, ip, userAgent, traceID string) (*xadmin.AuthActionResp, error) {
	targetUID := req.GetTargetUid()
	if operatorUID <= 0 || targetUID <= 0 {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "auth.target_invalid")
	}
	if operatorUID == targetUID {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "auth.cannot_deactivate_self")
	}
	targetUser, err := s.repo.GetUserByUID(ctx, targetUID)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(strings.TrimSpace(targetUser.Username), "admin") {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "auth.admin_protected")
	}
	if err := s.repo.DeactivateUser(ctx, targetUID); err != nil {
		return nil, err
	}
	if err := s.repo.RevokeAllSessionsByUID(ctx, targetUID, "deactivate"); err != nil {
		return nil, err
	}
	_ = s.audit(ctx, targetUID, "deactivate", "success", ip, userAgent, traceID, "operator_uid="+int32ToString(operatorUID))
	return &xadmin.AuthActionResp{
		Success: true,
		Action:  "deactivate",
	}, nil
}

func (s *service) ListSessions(ctx context.Context, uid int32, req *xadmin.AuthSessionsReq) (*xadmin.AuthSessionsResp, error) {
	if uid <= 0 {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "account.invalid_user")
	}
	status, err := sessionStatusToDB(req.GetStatus())
	if err != nil {
		return nil, err
	}
	limit := int(req.GetPageSize())
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.repo.ListSessionsByUID(ctx, uid, status, limit)
	if err != nil {
		return nil, err
	}
	result := make([]*xadmin.AuthSessionItem, 0, len(rows))
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
		result = append(result, item)
	}
	return &xadmin.AuthSessionsResp{
		Items: result,
	}, nil
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

func (s *service) IsSessionActive(ctx context.Context, uid int32, sessionID, tokenHash string) (bool, error) {
	if uid <= 0 || sessionID == "" || tokenHash == "" {
		return false, xerr.NewBiz(xerr.CodeUnauthorized, "auth.session_invalid")
	}
	return s.repo.IsSessionActive(ctx, uid, sessionID, tokenHash)
}

func (s *service) audit(ctx context.Context, uid int32, action, result, ip, userAgent, traceID, detail string) error {
	return s.repo.CreateAudit(ctx, &model.AdminUserLoginAudit{
		UID:       uid,
		Action:    action,
		Result:    result,
		TraceID:   traceID,
		SourceIP:  ip,
		Duration:  requestmeta.DurationString(ctx),
		UserAgent: userAgent,
		Detail:    detail,
	})
}

func int32ToString(v int32) string {
	return strconv.FormatInt(int64(v), 10)
}

func tokenTTL(ttlMinutes int64, ttlHours int64) time.Duration {
	if ttlMinutes > 0 {
		return time.Duration(ttlMinutes) * time.Minute
	}
	if ttlHours <= 0 {
		return 24 * time.Hour
	}
	return time.Duration(ttlHours) * time.Hour
}
