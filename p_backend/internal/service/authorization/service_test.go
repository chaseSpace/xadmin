package authorization

import (
	"context"
	"testing"
)

type stubRepository struct {
	superUsers           map[int32]bool
	protectedUsers       map[int32]bool
	protectedRoles       map[int64]bool
	protectedPositions   map[int64]bool
	userPositions        map[int32]int64
	userManagementRanks  map[int32]int32
	userPermissions      map[int32][]string
	delegablePermissions map[int32][]string
	positionRanks        map[int64]int32
	positionPermissions  map[int64][]string
	rolePermissions      map[int64][]string
}

func (r *stubRepository) IsSuperAdmin(_ context.Context, uid int32) (bool, error) {
	return r.superUsers[uid], nil
}

func (r *stubRepository) UserHasProtectedRole(_ context.Context, uid int32) (bool, error) {
	return r.protectedUsers[uid], nil
}

func (r *stubRepository) RoleIsProtected(_ context.Context, roleID int64) (bool, error) {
	return r.protectedRoles[roleID], nil
}

func (r *stubRepository) PositionHasProtectedRole(_ context.Context, positionID int64) (bool, error) {
	return r.protectedPositions[positionID], nil
}

func (r *stubRepository) GetUserPositionInfo(_ context.Context, uid int32) (int64, int32, error) {
	return r.userPositions[uid], r.userManagementRanks[uid], nil
}

func (r *stubRepository) GetPositionManagementRank(_ context.Context, positionID int64) (int32, error) {
	return r.positionRanks[positionID], nil
}

func (r *stubRepository) ListEffectivePermissionKeysByUID(_ context.Context, uid int32, delegableOnly bool) ([]string, error) {
	if delegableOnly {
		return r.delegablePermissions[uid], nil
	}
	return r.userPermissions[uid], nil
}

func (r *stubRepository) ListEffectivePermissionKeysByPositionID(_ context.Context, positionID int64) ([]string, error) {
	return r.positionPermissions[positionID], nil
}

func (r *stubRepository) ListEffectivePermissionKeysByRoleID(_ context.Context, roleID int64) ([]string, error) {
	return r.rolePermissions[roleID], nil
}

func newPolicyService() *Service {
	return NewServiceWithRepo(&stubRepository{
		superUsers:           map[int32]bool{1: true},
		protectedUsers:       map[int32]bool{9: true},
		protectedRoles:       map[int64]bool{90: true},
		protectedPositions:   map[int64]bool{900: true},
		userPositions:        map[int32]int64{1: 100, 2: 200, 3: 300, 4: 400, 9: 900},
		userManagementRanks:  map[int32]int32{1: 100, 2: 80, 3: 20, 4: 80, 9: 20},
		userPermissions:      map[int32][]string{2: {"read", "write", "manage"}, 3: {"audit"}, 4: {"read"}, 9: {"root"}},
		delegablePermissions: map[int32][]string{2: {"read", "write"}},
		positionRanks:        map[int64]int32{100: 100, 200: 80, 300: 20, 400: 20, 500: 80, 900: 100},
		positionPermissions:  map[int64][]string{200: {"read", "write", "manage"}, 300: {"read"}, 400: {"read", "reset_password"}, 500: {"read"}, 900: {"root"}},
		rolePermissions:      map[int64][]string{30: {"read"}, 40: {"read", "write", "manage"}, 90: {"root"}},
	})
}

func TestEnsureCanManageUserUsesManagementRank(t *testing.T) {
	svc := newPolicyService()
	ctx := context.Background()

	for name, targetUID := range map[string]int32{
		"self":                  2,
		"protected":             9,
		"equal management rank": 4,
	} {
		t.Run(name, func(t *testing.T) {
			if err := svc.EnsureCanManageUser(ctx, 2, targetUID); err == nil {
				t.Fatal("expected target management to be denied")
			}
		})
	}

	if err := svc.EnsureCanManageUser(ctx, 2, 3); err != nil {
		t.Fatalf("expected lower-rank target with a different permission domain to be allowed: %v", err)
	}
	if err := svc.EnsureCanManageUser(ctx, 1, 9); err != nil {
		t.Fatalf("expected super admin to manage protected target: %v", err)
	}
}

func TestEnsureCanAssignPositionUsesDelegablePermissions(t *testing.T) {
	svc := newPolicyService()
	ctx := context.Background()

	if err := svc.EnsureCanAssignPosition(ctx, 2, 300); err != nil {
		t.Fatalf("expected delegable subset to be allowed: %v", err)
	}
	if err := svc.EnsureCanAssignPosition(ctx, 2, 400); err == nil {
		t.Fatal("expected non-delegable permission grant to be denied")
	}
	if err := svc.EnsureCanAssignPosition(ctx, 2, 500); err == nil {
		t.Fatal("expected equal-rank position assignment to be denied")
	}
	if err := svc.EnsureCanAssignPosition(ctx, 2, 900); err == nil {
		t.Fatal("expected protected position assignment to be denied")
	}
}

func TestEnsureCanManagePositionUsesRankAndRoleUsesPermissionScope(t *testing.T) {
	svc := newPolicyService()
	ctx := context.Background()

	if err := svc.EnsureCanManagePosition(ctx, 2, 200); err == nil {
		t.Fatal("expected own position mutation to be denied")
	}
	if err := svc.EnsureCanManagePosition(ctx, 2, 300); err != nil {
		t.Fatalf("expected lower-rank position to be allowed: %v", err)
	}
	if err := svc.EnsureCanManagePosition(ctx, 2, 500); err == nil {
		t.Fatal("expected equal-rank position to be denied")
	}
	if err := svc.EnsureCanManagePosition(ctx, 2, 900); err == nil {
		t.Fatal("expected protected position to be denied")
	}
	if err := svc.EnsureCanManageRole(ctx, 2, 30); err != nil {
		t.Fatalf("expected lower-scope role to be allowed: %v", err)
	}
	if err := svc.EnsureCanManageRole(ctx, 2, 40); err == nil {
		t.Fatal("expected equal-scope role to be denied")
	}
	if err := svc.EnsureCanManageRole(ctx, 2, 90); err == nil {
		t.Fatal("expected protected role to be denied")
	}
	if err := svc.EnsureCanManagePosition(ctx, 1, 100); err == nil {
		t.Fatal("expected super admin own-position mutation to be denied")
	}
}

func TestCapabilityFlagsMirrorEnforcementRules(t *testing.T) {
	ordinary := &OperatorScope{
		UID:                  2,
		PositionID:           20,
		ManagementRank:       80,
		Permissions:          []string{"read", "write", "manage"},
		DelegablePermissions: []string{"read", "write"},
	}
	superAdmin := &OperatorScope{UID: 1, PositionID: 10, SuperAdmin: true}

	if !CanManageUserFromScope(ordinary, 3, false, 20) {
		t.Fatal("expected lower-rank user to be manageable")
	}
	if CanManageUserFromScope(ordinary, 2, false, 20) {
		t.Fatal("expected self-management to be denied")
	}
	if CanManageUserFromScope(ordinary, 4, false, 80) {
		t.Fatal("expected equal-rank user to be denied")
	}
	if CanManagePositionFromScope(ordinary, 30, true, 20) {
		t.Fatal("expected protected position management to be denied")
	}
	if !CanAssignPositionFromScope(ordinary, 30, false, 20, []string{"read", "write"}) {
		t.Fatal("expected delegable position to be assignable")
	}
	if CanAssignPositionFromScope(ordinary, 40, false, 20, []string{"read", "reset_password"}) {
		t.Fatal("expected non-delegable position to be denied")
	}
	if CanAssignPositionFromScope(ordinary, 50, false, 80, []string{"read"}) {
		t.Fatal("expected equal-rank position assignment to be denied")
	}
	if !CanManageRoleFromScope(superAdmin, true, []string{"root"}) {
		t.Fatal("expected super admin capability to bypass protected targets")
	}
}
