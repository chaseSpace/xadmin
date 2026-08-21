package organization

import (
	"context"
	"errors"
	"testing"

	xadmin "monorepo/proto/xadminpb"
)

var errDeniedByPolicy = errors.New("denied by authorization policy")

type authorizationStub struct {
	superAdminErr     error
	manageUserErr     error
	assignPositionErr error
	managePositionErr error
}

func (s authorizationStub) EnsureSuperAdmin(context.Context, int32) error {
	return s.superAdminErr
}

func (s authorizationStub) EnsureCanManageUser(context.Context, int32, int32) error {
	return s.manageUserErr
}

func (s authorizationStub) EnsureCanAssignPosition(context.Context, int32, int64) error {
	return s.assignPositionErr
}

func (s authorizationStub) EnsureCanManagePosition(context.Context, int32, int64) error {
	return s.managePositionErr
}

func TestPositionRoleBindingRejectsDirectServiceBypass(t *testing.T) {
	svc := &service{authorization: authorizationStub{superAdminErr: errDeniedByPolicy}}
	_, err := svc.UpdatePositionRoles(context.Background(), 2, &xadmin.OrganizationUpdatePositionRolesReq{Id: 1, RoleIds: []int64{1}})
	if !errors.Is(err, errDeniedByPolicy) {
		t.Fatalf("expected super-admin denial before repository access, got %v", err)
	}
}

func TestUserPrivilegeMutationsRejectUnsafeTargetsBeforeRepositoryAccess(t *testing.T) {
	ctx := context.Background()

	tests := map[string]func() error{
		"self or protected position change": func() error {
			svc := &service{authorization: authorizationStub{manageUserErr: errDeniedByPolicy}}
			_, err := svc.AssignUserPosition(ctx, 2, &xadmin.OrganizationAssignUserPositionReq{Uid: 2, PositionId: 1})
			return err
		},
		"equal or higher password reset": func() error {
			svc := &service{authorization: authorizationStub{manageUserErr: errDeniedByPolicy}}
			_, err := svc.ResetPassword(ctx, 2, &xadmin.OrganizationResetPasswordReq{Uid: 9})
			return err
		},
		"new account in protected position": func() error {
			svc := &service{authorization: authorizationStub{assignPositionErr: errDeniedByPolicy}}
			_, err := svc.CreateUser(ctx, 2, &xadmin.OrganizationCreateUserReq{PositionId: 9})
			return err
		},
		"batch transfer containing unsafe target": func() error {
			svc := &service{authorization: authorizationStub{manageUserErr: errDeniedByPolicy}}
			_, err := svc.BatchTransferUsers(ctx, 2, &xadmin.OrganizationBatchTransferUsersReq{Uids: []int32{2}, PositionId: 9})
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
