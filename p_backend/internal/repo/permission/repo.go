package permission

import (
	"context"
	"strings"
	"time"

	"monorepo/internal/model"
	"monorepo/pkg/db"
	"monorepo/pkg/xerr"
	commpb "monorepo/proto/xadminpb/commpb"

	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

type MenuFilters struct {
	Keyword  string
	Status   *int32
	MenuType *int32
	MenuIDs  []int64
	Deleted  *bool
}

type MenuRow struct {
	ID            int64      `gorm:"column:id"`
	ParentID      int64      `gorm:"column:parent_id"`
	Name          string     `gorm:"column:name"`
	RoutePath     string     `gorm:"column:route_path"`
	ComponentPath string     `gorm:"column:component_path"`
	MenuType      int32      `gorm:"column:menu_type"`
	PermissionKey string     `gorm:"column:permission_key"`
	Sort          int32      `gorm:"column:sort"`
	Status        int32      `gorm:"column:status"`
	UpdatedAt     *time.Time `gorm:"column:updated_at"`
	DeletedAt     int64      `gorm:"column:deleted_at"`
}

type RoleFilters struct {
	Keyword  string
	RoleType *int32
}

type RoleRow struct {
	ID        int64      `gorm:"column:id"`
	RoleName  string     `gorm:"column:role_name"`
	RoleType  int32      `gorm:"column:role_type"`
	Users     int32      `gorm:"column:users"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
}

const roleUsersAggregateJoinSQL = `
LEFT JOIN (
  SELECT opr.role_id, COUNT(DISTINCT u.uid) AS users
  FROM organization_position_role opr
  INNER JOIN admin_user u ON u.position_id = opr.position_id AND u.deleted_at = 0
  GROUP BY role_id
) ru ON ru.role_id = r.id
`

func NewRepo() *Repo {
	return &Repo{db: db.GetDatabase()}
}

func NewRepoWithDB(database *gorm.DB) *Repo {
	return &Repo{db: database}
}

func (r *Repo) ListMenus(ctx context.Context, page *commpb.PageArgs, sort []*commpb.SortArgs, filters MenuFilters) ([]MenuRow, int64, error) {
	rows := make([]MenuRow, 0, page.GetPs())
	query := r.db.WithContext(ctx).
		Table("permission_menu m").
		Select("m.id,m.parent_id,m.name,m.route_path,m.component_path,m.menu_type,m.permission_key,m.sort,m.status,m.updated_at,m.deleted_at")
	query = applyMenuFilters(query, filters)
	if len(sort) == 0 {
		query = query.Order("m.sort asc, m.id asc")
	}
	total, err := db.Paginate(
		query,
		page,
		sort,
		[]string{"m.id", "m.name", "m.menu_type", "m.sort", "m.status", "m.updated_at"},
		&rows,
		db.PaginateArgs{AppendCreatedAtDesc: false},
	)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repo) ListAllMenus(ctx context.Context) ([]MenuRow, error) {
	rows := make([]MenuRow, 0, 64)
	err := r.db.WithContext(ctx).
		Table("permission_menu m").
		Select("m.id,m.parent_id,m.name,m.route_path,m.component_path,m.menu_type,m.permission_key,m.sort,m.status,m.updated_at").
		Where("m.deleted_at = 0").
		Order("m.sort asc, m.id asc").
		Find(&rows).Error
	if err != nil {
		return nil, xerr.WrapDBE(err, "list all permission menus")
	}
	return rows, nil
}

func (r *Repo) GetMenuByID(ctx context.Context, id int64) (*MenuRow, error) {
	var row MenuRow
	err := r.db.WithContext(ctx).
		Table("permission_menu m").
		Select("m.id,m.parent_id,m.name,m.route_path,m.component_path,m.menu_type,m.permission_key,m.sort,m.status,m.updated_at,m.deleted_at").
		Where("m.id = ? AND m.deleted_at = 0", id).
		First(&row).Error
	if err != nil {
		return nil, xerr.WrapDBNotFound(err, "permission menu not found")
	}
	return &row, nil
}

func (r *Repo) GetMenuByIDIncludingDeleted(ctx context.Context, id int64) (*MenuRow, error) {
	var row MenuRow
	err := r.db.WithContext(ctx).
		Table("permission_menu m").
		Select("m.id,m.parent_id,m.name,m.route_path,m.component_path,m.menu_type,m.permission_key,m.sort,m.status,m.updated_at,m.deleted_at").
		Where("m.id = ?", id).
		First(&row).Error
	if err != nil {
		return nil, xerr.WrapDBNotFound(err, "permission menu not found")
	}
	return &row, nil
}

func (r *Repo) CreateMenu(ctx context.Context, data *model.PermissionMenu) error {
	return xerr.WrapDBDuplicate(
		r.db.WithContext(ctx).Create(data).Error,
		"permission menu already exists",
	)
}

func (r *Repo) UpdateMenuByID(ctx context.Context, id int64, updates map[string]any) error {
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Model(&model.PermissionMenu{}).
			Where("id = ? AND deleted_at = 0", id).
			Updates(updates).Error,
		"update permission menu",
	)
}

func (r *Repo) HasMenuChildren(ctx context.Context, parentID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.PermissionMenu{}).
		Where("parent_id = ? AND deleted_at = 0", parentID).
		Count(&count).Error
	if err != nil {
		return false, xerr.WrapDBE(err, "count menu children")
	}
	return count > 0, nil
}

func (r *Repo) SoftDeleteMenuByID(ctx context.Context, id int64) error {
	deletedAt := time.Now().Unix()
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Model(&model.PermissionMenu{}).
			Where("id = ? AND deleted_at = 0", id).
			Update("deleted_at", deletedAt).Error,
		"soft delete permission menu",
	)
}

func (r *Repo) HardDeleteMenuByID(ctx context.Context, id int64) error {
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Where("id = ? AND deleted_at <> 0", id).
			Delete(&model.PermissionMenu{}).Error,
		"hard delete permission menu",
	)
}

func (r *Repo) ListRoles(ctx context.Context, page *commpb.PageArgs, sort []*commpb.SortArgs, filters RoleFilters) ([]RoleRow, int64, error) {
	rows := make([]RoleRow, 0, page.GetPs())
	query := r.db.WithContext(ctx).
		Table("permission_role r").
		Select(`
r.id,
r.role_name,
r.role_type,
COALESCE(ru.users, 0) AS users,
r.updated_at
`).
		Joins(roleUsersAggregateJoinSQL).
		Where("r.deleted_at = 0")
	query = applyRoleFilters(query, filters)
	if len(sort) == 0 {
		query = query.Order("r.created_at desc")
	}
	total, err := db.Paginate(
		query,
		page,
		sort,
		[]string{"r.id", "r.role_name", "r.role_type", "r.updated_at", "users"},
		&rows,
		db.PaginateArgs{AppendCreatedAtDesc: false},
	)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repo) GetRoleByID(ctx context.Context, id int64) (*RoleRow, error) {
	var row RoleRow
	err := r.db.WithContext(ctx).
		Table("permission_role r").
		Select(`
r.id,
r.role_name,
r.role_type,
COALESCE(ru.users, 0) AS users,
r.updated_at
`).
		Joins(roleUsersAggregateJoinSQL).
		Where("r.id = ? AND r.deleted_at = 0", id).
		First(&row).Error
	if err != nil {
		return nil, xerr.WrapDBNotFound(err, "permission role not found")
	}
	return &row, nil
}

func (r *Repo) CreateRole(ctx context.Context, data *model.PermissionRole) error {
	return xerr.WrapDBDuplicate(
		r.db.WithContext(ctx).Create(data).Error,
		"角色名称已存在",
	)
}

func (r *Repo) UpdateRoleByID(ctx context.Context, id int64, updates map[string]any) error {
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Model(&model.PermissionRole{}).
			Where("id = ? AND deleted_at = 0", id).
			Updates(updates).Error,
		"update permission role",
	)
}

func (r *Repo) SoftDeleteRoleByID(ctx context.Context, id int64) error {
	deletedAt := time.Now().Unix()
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Model(&model.PermissionRole{}).
			Where("id = ? AND deleted_at = 0", id).
			Update("deleted_at", deletedAt).Error,
		"soft delete permission role",
	)
}

func (r *Repo) HasRoleUser(ctx context.Context, roleID int64, uid int32) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("organization_position_role opr").
		Joins("INNER JOIN admin_user u ON u.position_id = opr.position_id AND u.deleted_at = 0").
		Where("opr.role_id = ? AND u.uid = ?", roleID, uid).
		Count(&count).Error
	if err != nil {
		return false, xerr.WrapDBE(err, "query role user relation")
	}
	return count > 0, nil
}

func (r *Repo) ListRoleMenuIDs(ctx context.Context, roleID int64) ([]int64, error) {
	ids := make([]int64, 0, 32)
	err := r.db.WithContext(ctx).
		Table("permission_role_menu").
		Where("role_id = ?", roleID).
		Order("menu_id asc").
		Pluck("menu_id", &ids).Error
	if err != nil {
		return nil, xerr.WrapDBE(err, "list role menu ids")
	}
	return ids, nil
}

func (r *Repo) ExpandMenuIDsWithAncestors(ctx context.Context, menuIDs []int64) ([]int64, error) {
	uniqueMenuIDs := normalizeMenuIDs(menuIDs)
	if len(uniqueMenuIDs) == 0 {
		return []int64{}, nil
	}
	rows, err := r.ListAllMenus(ctx)
	if err != nil {
		return nil, err
	}
	parentByID := make(map[int64]int64, len(rows))
	for _, row := range rows {
		parentByID[row.ID] = row.ParentID
	}
	expanded := make([]int64, 0, len(uniqueMenuIDs))
	seen := make(map[int64]struct{}, len(rows))
	for _, menuID := range uniqueMenuIDs {
		for currentID := menuID; currentID > 0; currentID = parentByID[currentID] {
			if _, exists := parentByID[currentID]; !exists {
				break
			}
			if _, exists := seen[currentID]; exists {
				continue
			}
			seen[currentID] = struct{}{}
			expanded = append(expanded, currentID)
		}
	}
	return expanded, nil
}

func (r *Repo) ReplaceRoleMenus(ctx context.Context, roleID int64, menuIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.PermissionRoleMenu{}).Error; err != nil {
			return xerr.WrapDBE(err, "clear role menus")
		}
		if len(menuIDs) == 0 {
			return nil
		}
		uniqueMenuIDs := normalizeMenuIDs(menuIDs)
		if len(uniqueMenuIDs) == 0 {
			return nil
		}
		items := make([]*model.PermissionRoleMenu, 0, len(uniqueMenuIDs))
		for _, menuID := range uniqueMenuIDs {
			items = append(items, &model.PermissionRoleMenu{RoleID: roleID, MenuID: menuID})
		}
		if err := tx.Create(items).Error; err != nil {
			return xerr.WrapDBE(err, "insert role menus")
		}
		return nil
	})
}

func normalizeMenuIDs(menuIDs []int64) []int64 {
	uniqueMenuIDs := make([]int64, 0, len(menuIDs))
	seen := make(map[int64]struct{}, len(menuIDs))
	for _, menuID := range menuIDs {
		if menuID <= 0 {
			continue
		}
		if _, ok := seen[menuID]; ok {
			continue
		}
		seen[menuID] = struct{}{}
		uniqueMenuIDs = append(uniqueMenuIDs, menuID)
	}
	return uniqueMenuIDs
}

func applyMenuFilters(query *gorm.DB, filters MenuFilters) *gorm.DB {
	if filters.Deleted != nil && *filters.Deleted {
		query = query.Where("m.deleted_at <> 0")
	} else {
		query = query.Where("m.deleted_at = 0")
	}
	if kw := strings.TrimSpace(filters.Keyword); kw != "" {
		query = query.Where("(m.name LIKE ? OR m.route_path LIKE ? OR m.permission_key LIKE ?)", "%"+kw+"%", "%"+kw+"%", "%"+kw+"%")
	}
	if filters.Status != nil {
		query = query.Where("m.status = ?", *filters.Status)
	}
	if filters.MenuType != nil {
		query = query.Where("m.menu_type = ?", *filters.MenuType)
	}
	if len(filters.MenuIDs) > 0 {
		query = query.Where("m.id IN ?", filters.MenuIDs)
	}
	return query
}

func applyRoleFilters(query *gorm.DB, filters RoleFilters) *gorm.DB {
	if kw := strings.TrimSpace(filters.Keyword); kw != "" {
		query = query.Where("r.role_name LIKE ?", "%"+kw+"%")
	}
	if filters.RoleType != nil {
		query = query.Where("r.role_type = ?", *filters.RoleType)
	}
	return query
}
