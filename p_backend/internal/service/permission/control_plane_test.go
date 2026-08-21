package permission

import (
	"context"
	"errors"
	"testing"

	xadmin "monorepo/proto/xadminpb"
)

var errDeniedByPolicy = errors.New("denied by authorization policy")

type authorizationStub struct {
	superAdminErr error
	manageRoleErr error
}

func (s authorizationStub) EnsureSuperAdmin(context.Context, int32) error {
	return s.superAdminErr
}

func (s authorizationStub) EnsureCanManageRole(context.Context, int32, int64) error {
	return s.manageRoleErr
}

func TestControlPlaneMutationsRejectDirectServiceBypass(t *testing.T) {
	svc := &service{authorization: authorizationStub{superAdminErr: errDeniedByPolicy}}
	ctx := context.Background()

	tests := map[string]func() error{
		"menu permission schema": func() error {
			_, err := svc.UpdateMenu(ctx, 2, &xadmin.PermissionUpdateMenuReq{Id: 1})
			return err
		},
		"role menu binding": func() error {
			_, err := svc.UpdateRoleMenus(ctx, 2, &xadmin.PermissionUpdateRoleMenusReq{RoleId: 2})
			return err
		},
	}
	for name, run := range tests {
		t.Run(name, func(t *testing.T) {
			if err := run(); !errors.Is(err, errDeniedByPolicy) {
				t.Fatalf("expected authorization denial before repository access, got %v", err)
			}
		})
	}
}

func TestRoleProfileMutationRejectsEqualOrHigherTarget(t *testing.T) {
	svc := &service{authorization: authorizationStub{manageRoleErr: errDeniedByPolicy}}
	_, err := svc.UpdateRole(context.Background(), 2, &xadmin.PermissionUpdateRoleReq{Id: 9})
	if !errors.Is(err, errDeniedByPolicy) {
		t.Fatalf("expected role target denial before repository access, got %v", err)
	}
}
