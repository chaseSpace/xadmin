package authorization

import (
	"context"

	authorizationrepo "monorepo/internal/repo/authorization"
	"monorepo/pkg/xerr"
)

type Repository interface {
	IsSuperAdmin(ctx context.Context, uid int32) (bool, error)
	UserHasProtectedRole(ctx context.Context, uid int32) (bool, error)
	RoleIsProtected(ctx context.Context, roleID int64) (bool, error)
	PositionHasProtectedRole(ctx context.Context, positionID int64) (bool, error)
	GetUserPositionID(ctx context.Context, uid int32) (int64, error)
	ListEffectivePermissionKeysByUID(ctx context.Context, uid int32, delegableOnly bool) ([]string, error)
	ListEffectivePermissionKeysByPositionID(ctx context.Context, positionID int64) ([]string, error)
	ListEffectivePermissionKeysByRoleID(ctx context.Context, roleID int64) ([]string, error)
}

type Service struct {
	repo Repository
}

func NewService() *Service {
	return &Service{repo: authorizationrepo.NewRepo()}
}

func NewServiceWithRepo(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) IsSuperAdmin(ctx context.Context, uid int32) (bool, error) {
	return s.repo.IsSuperAdmin(ctx, uid)
}

func (s *Service) IsProtectedRole(ctx context.Context, roleID int64) (bool, error) {
	return s.repo.RoleIsProtected(ctx, roleID)
}

func (s *Service) EnsureSuperAdmin(ctx context.Context, uid int32) error {
	ok, err := s.repo.IsSuperAdmin(ctx, uid)
	if err != nil {
		return err
	}
	if !ok {
		return xerr.NewBiz(xerr.CodeForbidden, "auth.superadmin_required")
	}
	return nil
}

func (s *Service) EnsureCanManageUser(ctx context.Context, operatorUID, targetUID int32) error {
	if operatorUID <= 0 || targetUID <= 0 {
		return xerr.NewBiz(xerr.CodeUnauthorized, "auth.not_logged_in")
	}
	if operatorUID == targetUID {
		return xerr.NewBiz(xerr.CodeForbidden, "auth.self_privilege_change_forbidden")
	}
	if super, err := s.repo.IsSuperAdmin(ctx, operatorUID); err != nil {
		return err
	} else if super {
		return nil
	}
	protected, err := s.repo.UserHasProtectedRole(ctx, targetUID)
	if err != nil {
		return err
	}
	if protected {
		return xerr.NewBiz(xerr.CodeForbidden, "auth.protected_target")
	}
	operatorPermissions, err := s.repo.ListEffectivePermissionKeysByUID(ctx, operatorUID, false)
	if err != nil {
		return err
	}
	targetPermissions, err := s.repo.ListEffectivePermissionKeysByUID(ctx, targetUID, false)
	if err != nil {
		return err
	}
	if !isStrictSubset(targetPermissions, operatorPermissions) {
		return xerr.NewBiz(xerr.CodeForbidden, "auth.target_scope_exceeded")
	}
	return nil
}

func (s *Service) EnsureCanAssignPosition(ctx context.Context, operatorUID int32, positionID int64) error {
	if operatorUID <= 0 {
		return xerr.NewBiz(xerr.CodeUnauthorized, "auth.not_logged_in")
	}
	if super, err := s.repo.IsSuperAdmin(ctx, operatorUID); err != nil {
		return err
	} else if super {
		return nil
	}
	protected, err := s.repo.PositionHasProtectedRole(ctx, positionID)
	if err != nil {
		return err
	}
	if protected {
		return xerr.NewBiz(xerr.CodeForbidden, "auth.protected_target")
	}
	delegablePermissions, err := s.repo.ListEffectivePermissionKeysByUID(ctx, operatorUID, true)
	if err != nil {
		return err
	}
	positionPermissions, err := s.repo.ListEffectivePermissionKeysByPositionID(ctx, positionID)
	if err != nil {
		return err
	}
	if !isSubset(positionPermissions, delegablePermissions) {
		return xerr.NewBiz(xerr.CodeForbidden, "auth.grant_scope_exceeded")
	}
	return nil
}

func (s *Service) EnsureCanManagePosition(ctx context.Context, operatorUID int32, positionID int64) error {
	if operatorUID <= 0 {
		return xerr.NewBiz(xerr.CodeUnauthorized, "auth.not_logged_in")
	}
	if super, err := s.repo.IsSuperAdmin(ctx, operatorUID); err != nil {
		return err
	} else if super {
		return nil
	}
	operatorPositionID, err := s.repo.GetUserPositionID(ctx, operatorUID)
	if err != nil {
		return err
	}
	if operatorPositionID == positionID {
		return xerr.NewBiz(xerr.CodeForbidden, "auth.self_privilege_change_forbidden")
	}
	protected, err := s.repo.PositionHasProtectedRole(ctx, positionID)
	if err != nil {
		return err
	}
	if protected {
		return xerr.NewBiz(xerr.CodeForbidden, "auth.protected_target")
	}
	operatorPermissions, err := s.repo.ListEffectivePermissionKeysByUID(ctx, operatorUID, false)
	if err != nil {
		return err
	}
	positionPermissions, err := s.repo.ListEffectivePermissionKeysByPositionID(ctx, positionID)
	if err != nil {
		return err
	}
	if !isStrictSubset(positionPermissions, operatorPermissions) {
		return xerr.NewBiz(xerr.CodeForbidden, "auth.target_scope_exceeded")
	}
	return nil
}

func (s *Service) EnsureCanManageRole(ctx context.Context, operatorUID int32, roleID int64) error {
	if err := s.EnsureCanManageRoleTarget(ctx, operatorUID, roleID); err != nil {
		return err
	}
	return nil
}

func (s *Service) EnsureCanManageRoleTarget(ctx context.Context, operatorUID int32, roleID int64) error {
	if operatorUID <= 0 {
		return xerr.NewBiz(xerr.CodeUnauthorized, "auth.not_logged_in")
	}
	if super, err := s.repo.IsSuperAdmin(ctx, operatorUID); err != nil {
		return err
	} else if super {
		return nil
	}
	protected, err := s.repo.RoleIsProtected(ctx, roleID)
	if err != nil {
		return err
	}
	if protected {
		return xerr.NewBiz(xerr.CodeForbidden, "auth.protected_target")
	}
	operatorPermissions, err := s.repo.ListEffectivePermissionKeysByUID(ctx, operatorUID, false)
	if err != nil {
		return err
	}
	rolePermissions, err := s.repo.ListEffectivePermissionKeysByRoleID(ctx, roleID)
	if err != nil {
		return err
	}
	if !isStrictSubset(rolePermissions, operatorPermissions) {
		return xerr.NewBiz(xerr.CodeForbidden, "auth.target_scope_exceeded")
	}
	return nil
}

func isSubset(candidate, scope []string) bool {
	set := make(map[string]struct{}, len(scope))
	for _, key := range scope {
		if key != "" {
			set[key] = struct{}{}
		}
	}
	for _, key := range candidate {
		if key == "" {
			continue
		}
		if _, ok := set[key]; !ok {
			return false
		}
	}
	return true
}

func isStrictSubset(candidate, scope []string) bool {
	if !isSubset(candidate, scope) {
		return false
	}
	candidateSet := make(map[string]struct{}, len(candidate))
	scopeSet := make(map[string]struct{}, len(scope))
	for _, key := range candidate {
		if key != "" {
			candidateSet[key] = struct{}{}
		}
	}
	for _, key := range scope {
		if key != "" {
			scopeSet[key] = struct{}{}
		}
	}
	return len(candidateSet) < len(scopeSet)
}
