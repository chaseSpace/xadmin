package account

import (
	"testing"

	authrepo "monorepo/internal/repo/auth"
)

func TestBuildAuthProfileRelations(t *testing.T) {
	department, position, roles := buildAuthProfileRelations([]authrepo.ProfileRelationRow{
		{
			DepartmentID:   10,
			DepartmentName: " 技术部 ",
			DepartmentCode: " tech ",
			PositionID:     20,
			PositionName:   " 平台工程师 ",
			PositionCode:   " platform_engineer ",
			RoleID:         30,
			RoleName:       " 运维管理员 ",
			RoleCode:       " ops_admin ",
		},
		{
			DepartmentID:   10,
			DepartmentName: "技术部",
			DepartmentCode: "tech",
			PositionID:     20,
			PositionName:   "平台工程师",
			PositionCode:   "platform_engineer",
			RoleID:         31,
			RoleName:       "审计员",
			RoleCode:       "auditor",
		},
	})

	if department == nil || department.GetId() != 10 || department.GetName() != "技术部" || department.GetCode() != "tech" {
		t.Fatalf("unexpected department: %+v", department)
	}
	if position == nil || position.GetId() != 20 || position.GetName() != "平台工程师" || position.GetCode() != "platform_engineer" {
		t.Fatalf("unexpected position: %+v", position)
	}
	if len(roles) != 2 || roles[0].GetName() != "运维管理员" || roles[1].GetCode() != "auditor" {
		t.Fatalf("unexpected roles: %+v", roles)
	}
}

func TestBuildAuthProfileRelationsWithoutAssignment(t *testing.T) {
	department, position, roles := buildAuthProfileRelations(nil)
	if department != nil || position != nil || len(roles) != 0 {
		t.Fatalf("unexpected empty relations: department=%+v position=%+v roles=%+v", department, position, roles)
	}
}
