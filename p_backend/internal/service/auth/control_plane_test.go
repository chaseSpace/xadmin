package auth

import (
	"context"
	"errors"
	"testing"

	xadmin "monorepo/proto/xadminpb"
)

var errDeniedByPolicy = errors.New("denied by authorization policy")

type authorizationStub struct{}

func (authorizationStub) EnsureCanManageUser(context.Context, int32, int32) error {
	return errDeniedByPolicy
}

func TestAccountTakeoverActionsRejectUnsafeTargetBeforeRepositoryAccess(t *testing.T) {
	svc := &service{authorization: authorizationStub{}}
	ctx := context.Background()

	if _, err := svc.ForceLogout(ctx, 2, &xadmin.AuthForceLogoutReq{TargetUid: 9}, "", "", ""); !errors.Is(err, errDeniedByPolicy) {
		t.Fatalf("expected force-logout target denial, got %v", err)
	}
	if _, err := svc.Deactivate(ctx, 2, &xadmin.AuthDeactivateReq{TargetUid: 9}, "", "", ""); !errors.Is(err, errDeniedByPolicy) {
		t.Fatalf("expected deactivation target denial, got %v", err)
	}
}
