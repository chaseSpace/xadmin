package permission

import (
	"context"
	"testing"
	"time"

	"monorepo/internal/model"
	commpb "monorepo/proto/xadminpb/commpb"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func openPermissionRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "host=127.0.0.1 user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("skip permission repo test: open postgres failed: %v", err)
	}
	return db
}

func resetPermissionRepoTestTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqls := []string{
		`DROP TABLE IF EXISTS organization_position_role`,
		`DROP TABLE IF EXISTS admin_user`,
		`DROP TABLE IF EXISTS permission_role_menu`,
		`DROP TABLE IF EXISTS permission_menu`,
		`DROP TABLE IF EXISTS permission_role`,
		`CREATE TABLE permission_menu (
			id BIGINT PRIMARY KEY,
			parent_id BIGINT NOT NULL DEFAULT 0,
			name VARCHAR(64) NOT NULL,
			route_path VARCHAR(255) NOT NULL DEFAULT '',
			component_path VARCHAR(255) NOT NULL DEFAULT '',
			menu_type INT NOT NULL,
			permission_key VARCHAR(128) NOT NULL DEFAULT '',
			sort INT NOT NULL DEFAULT 0,
			status INT NOT NULL DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at BIGINT NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE permission_role (
			id BIGINT PRIMARY KEY,
			role_name VARCHAR(64) NOT NULL,
			role_type INT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at BIGINT NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE permission_role_menu (
			id BIGINT PRIMARY KEY,
			role_id BIGINT NOT NULL,
			menu_id BIGINT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE admin_user (
			id BIGINT PRIMARY KEY,
			uid INT NOT NULL,
			position_id BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at BIGINT NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE organization_position_role (
			id BIGINT PRIMARY KEY,
			position_id BIGINT NOT NULL,
			role_id BIGINT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatalf("exec sql failed: %v, sql=%s", err, sql)
		}
	}
}

func TestListRolesCountsPositionUsersDistinctly(t *testing.T) {
	db := openPermissionRepoTestDB(t)
	resetPermissionRepoTestTables(t, db)

	ctx := context.Background()
	repo := NewRepoWithDB(db)

	roles := []model.PermissionRole{
		{ID: 1, RoleName: "组织管理员", RoleType: 2},
		{ID: 2, RoleName: "审计员", RoleType: 2},
	}
	for _, role := range roles {
		if err := db.WithContext(ctx).Table(role.TableName()).Create(&role).Error; err != nil {
			t.Fatalf("create role failed: %v", err)
		}
	}

	users := []model.AdminUser{
		{ID: 1, UID: 10001, PositionID: 10, DeletedAt: 0},
		{ID: 2, UID: 10002, PositionID: 10, DeletedAt: 0},
		{ID: 3, UID: 10003, PositionID: 11, DeletedAt: 0},
		{ID: 4, UID: 10004, PositionID: 10, DeletedAt: 123},
		{ID: 5, UID: 10005, PositionID: 12, DeletedAt: 0},
	}
	for _, user := range users {
		if err := db.WithContext(ctx).Table(user.TableName()).Create(&user).Error; err != nil {
			t.Fatalf("create user failed: %v", err)
		}
	}

	positionRoles := []model.OrganizationPositionRole{
		{ID: 1, PositionID: 10, RoleID: 1},
		{ID: 2, PositionID: 11, RoleID: 2},
		{ID: 3, PositionID: 10, RoleID: 2},
	}
	for _, item := range positionRoles {
		if err := db.WithContext(ctx).Table(item.TableName()).Create(&item).Error; err != nil {
			t.Fatalf("create organization_position_role failed: %v", err)
		}
	}

	rows, total, err := repo.ListRoles(ctx, &commpb.PageArgs{Pn: 1, Ps: 10}, nil, RoleFilters{})
	if err != nil {
		t.Fatalf("ListRoles failed: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total=2, got=%d", total)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got=%d", len(rows))
	}

	roleUsersCount := map[int64]int32{}
	for _, row := range rows {
		roleUsersCount[row.ID] = row.Users
	}
	if roleUsersCount[1] != 2 {
		t.Fatalf("expected role 1 users=2, got=%d", roleUsersCount[1])
	}
	if roleUsersCount[2] != 3 {
		t.Fatalf("expected role 2 users=3, got=%d", roleUsersCount[2])
	}

	role, err := repo.GetRoleByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetRoleByID failed: %v", err)
	}
	if role.Users != 2 {
		t.Fatalf("expected role detail users=2, got=%d", role.Users)
	}
}

func TestExpandMenuIDsWithAncestors(t *testing.T) {
	db := openPermissionRepoTestDB(t)
	resetPermissionRepoTestTables(t, db)

	ctx := context.Background()
	repo := NewRepoWithDB(db)

	menus := []model.PermissionMenu{
		{ID: 1, ParentID: 0, Name: "权限管理", MenuType: 1, PermissionKey: "permission.root", Sort: 10, Status: 1, DeletedAt: 0},
		{ID: 2, ParentID: 1, Name: "菜单权限", RoutePath: "/permission/menu-permissions", MenuType: 2, PermissionKey: "permission.menus.view", Sort: 20, Status: 1, DeletedAt: 0},
		{ID: 3, ParentID: 2, Name: "编辑按钮", MenuType: 3, PermissionKey: "permission.menus.edit", Sort: 30, Status: 1, DeletedAt: 0},
		{ID: 4, ParentID: 0, Name: "系统设置", RoutePath: "/system/settings", MenuType: 2, PermissionKey: "system.settings.view", Sort: 40, Status: 1, DeletedAt: 123},
	}
	for _, menu := range menus {
		if err := db.WithContext(ctx).Table(menu.TableName()).Create(&menu).Error; err != nil {
			t.Fatalf("create menu failed: %v", err)
		}
	}

	ids, err := repo.ExpandMenuIDsWithAncestors(ctx, []int64{2, 2, 3, 4, 999, -1})
	if err != nil {
		t.Fatalf("ExpandMenuIDsWithAncestors failed: %v", err)
	}
	expected := []int64{2, 1, 3}
	if len(ids) != len(expected) {
		t.Fatalf("expected ids=%v, got=%v", expected, ids)
	}
	for index, expectedID := range expected {
		if ids[index] != expectedID {
			t.Fatalf("expected ids=%v, got=%v", expected, ids)
		}
	}
}

func TestListMenusFiltersDeletedRows(t *testing.T) {
	db := openPermissionRepoTestDB(t)
	resetPermissionRepoTestTables(t, db)

	ctx := context.Background()
	repo := NewRepoWithDB(db)
	now := time.Now()

	menus := []model.PermissionMenu{
		{ID: 1, ParentID: 0, Name: "可见菜单", MenuType: 2, PermissionKey: "menu.visible", Sort: 10, Status: 1, DeletedAt: 0},
		{ID: 2, ParentID: 0, Name: "已删菜单", MenuType: 2, PermissionKey: "menu.deleted", Sort: 20, Status: 1, DeletedAt: now.Add(-2 * time.Hour).Unix()},
	}
	for _, menu := range menus {
		menu.CreatedAt = now
		menu.UpdatedAt = now
		if err := db.WithContext(ctx).Table(menu.TableName()).Create(&menu).Error; err != nil {
			t.Fatalf("create menu failed: %v", err)
		}
	}

	notDeleted := false
	rows, total, err := repo.ListMenus(ctx, &commpb.PageArgs{Pn: 1, Ps: 10}, nil, MenuFilters{Deleted: &notDeleted})
	if err != nil {
		t.Fatalf("ListMenus not deleted failed: %v", err)
	}
	if total != 1 || len(rows) != 1 || rows[0].ID != 1 {
		t.Fatalf("expected only not deleted menu, total=%d rows=%v", total, rows)
	}

	deleted := true
	rows, total, err = repo.ListMenus(ctx, &commpb.PageArgs{Pn: 1, Ps: 10}, nil, MenuFilters{Deleted: &deleted})
	if err != nil {
		t.Fatalf("ListMenus deleted failed: %v", err)
	}
	if total != 1 || len(rows) != 1 || rows[0].ID != 2 || rows[0].DeletedAt == 0 {
		t.Fatalf("expected only deleted menu, total=%d rows=%v", total, rows)
	}
}

func TestHardDeleteMenuByIDRemovesOnlyDeletedRows(t *testing.T) {
	db := openPermissionRepoTestDB(t)
	resetPermissionRepoTestTables(t, db)

	ctx := context.Background()
	repo := NewRepoWithDB(db)
	now := time.Now()

	menus := []model.PermissionMenu{
		{ID: 1, ParentID: 0, Name: "可见菜单", MenuType: 2, PermissionKey: "menu.visible", Sort: 10, Status: 1, DeletedAt: 0},
		{ID: 2, ParentID: 0, Name: "已删菜单", MenuType: 2, PermissionKey: "menu.deleted", Sort: 20, Status: 1, DeletedAt: now.Add(-2 * time.Hour).Unix()},
	}
	for _, menu := range menus {
		menu.CreatedAt = now
		menu.UpdatedAt = now
		if err := db.WithContext(ctx).Table(menu.TableName()).Create(&menu).Error; err != nil {
			t.Fatalf("create menu failed: %v", err)
		}
	}

	if err := repo.HardDeleteMenuByID(ctx, 1); err != nil {
		t.Fatalf("HardDeleteMenuByID visible failed: %v", err)
	}
	if _, err := repo.GetMenuByIDIncludingDeleted(ctx, 1); err != nil {
		t.Fatalf("visible menu should not be hard deleted: %v", err)
	}

	if err := repo.HardDeleteMenuByID(ctx, 2); err != nil {
		t.Fatalf("HardDeleteMenuByID deleted failed: %v", err)
	}
	if _, err := repo.GetMenuByIDIncludingDeleted(ctx, 2); err == nil {
		t.Fatalf("deleted menu should be hard deleted")
	}
}
