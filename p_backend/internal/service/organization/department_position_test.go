package organization

import (
	"context"
	"testing"

	organizationrepo "monorepo/internal/repo/organization"
)

func TestResolveAssignablePositionDepartmentIDUsesNearestAncestor(t *testing.T) {
	departments := map[int64]*organizationrepo.DepartmentRow{
		1:  {ID: 1, ParentID: 0, PositionCount: 4},
		2:  {ID: 2, ParentID: 1, PositionCount: 2},
		12: {ID: 12, ParentID: 2, PositionCount: 0},
	}
	lookup := func(_ context.Context, id int64) (*organizationrepo.DepartmentRow, error) {
		return departments[id], nil
	}

	got, err := resolveAssignablePositionDepartmentID(context.Background(), 12, lookup)
	if err != nil {
		t.Fatalf("resolve assignable position department failed: %v", err)
	}
	if got != 2 {
		t.Fatalf("expected nearest ancestor department 2, got %d", got)
	}
}

func TestResolveAssignablePositionDepartmentIDKeepsDirectDepartment(t *testing.T) {
	lookup := func(_ context.Context, id int64) (*organizationrepo.DepartmentRow, error) {
		return &organizationrepo.DepartmentRow{ID: id, ParentID: 1, PositionCount: 2}, nil
	}

	got, err := resolveAssignablePositionDepartmentID(context.Background(), 2, lookup)
	if err != nil {
		t.Fatalf("resolve assignable position department failed: %v", err)
	}
	if got != 2 {
		t.Fatalf("expected direct department 2, got %d", got)
	}
}
