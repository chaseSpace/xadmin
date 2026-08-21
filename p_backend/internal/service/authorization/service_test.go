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
	userPermissions      map[int32][]string
	delegablePermissions map[int32][]string
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

func (r *stubRepository) GetUserPositionID(_ context.Context, uid int32) (int64, error) {
	return r.userPositions[uid], nil
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
		userPermissions:      map[int32][]string{2: {"read", "write", "manage"}, 3: {"read"}, 4: {"read", "write", "manage"}, 9: {"root"}},
		delegablePermissions: map[int32][]string{2: {"read", "write"}},
		positionPermissions:  map[int64][]string{200: {"read", "write", "manage"}, 300: {"read"}, 400: {"read", "reset_password"}, 900: {"root"}},
		rolePermissions:      map[int64][]string{30: {"read"}, 40: {"read", "write", "manage"}, 90: {"root"}},
	})
}

func TestEnsureCanManageUserEnforcesTargetHierarchy(t *testing.T) {
	svc := newPolicyService()
	ctx := context.Background()

	for name, targetUID := range map[string]int32{
		"self":              2,
		"protected":         9,
		"equal permissions": 4,
	} {
		t.Run(name, func(t *testing.T) {
			if err := svc.EnsureCanManageUser(ctx, 2, targetUID); err == nil {
				t.Fatal("expected target management to be denied")
			}
		})
	}

	if err := svc.EnsureCanManageUser(ctx, 2, 3); err != nil {
		t.Fatalf("expected lower-privilege target to be allowed: %v", err)
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
	if err := svc.EnsureCanAssignPosition(ctx, 2, 900); err == nil {
		t.Fatal("expected protected position assignment to be denied")
	}
}

func TestEnsureCanManagePositionAndRoleRequireStrictlyLowerScope(t *testing.T) {
	svc := newPolicyService()
	ctx := context.Background()

	if err := svc.EnsureCanManagePosition(ctx, 2, 200); err == nil {
		t.Fatal("expected own position mutation to be denied")
	}
	if err := svc.EnsureCanManagePosition(ctx, 2, 300); err != nil {
		t.Fatalf("expected lower-scope position to be allowed: %v", err)
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
}
