package authorization

import (
	"context"

	"monorepo/pkg/consts"
	"monorepo/pkg/db"
	"monorepo/pkg/xerr"

	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo() *Repo {
	return &Repo{db: db.GetDatabase()}
}

func NewRepoWithDB(database *gorm.DB) *Repo {
	return &Repo{db: database}
}

func (r *Repo) IsSuperAdmin(ctx context.Context, uid int32) (bool, error) {
	if uid <= 0 {
		return false, nil
	}
	var count int64
	err := r.db.WithContext(ctx).
		Table("admin_user u").
		Joins("INNER JOIN organization_position_role opr ON opr.position_id = u.position_id").
		Joins("INNER JOIN permission_role r ON r.id = opr.role_id AND r.deleted_at = 0").
		Where("u.uid = ? AND u.deleted_at = 0 AND r.role_code = ? AND r.is_protected = TRUE", uid, consts.PermissionRoleCodeSuperAdmin).
		Count(&count).Error
	if err != nil {
		return false, xerr.WrapDBE(err, "check super admin")
	}
	return count > 0, nil
}

func (r *Repo) UserHasProtectedRole(ctx context.Context, uid int32) (bool, error) {
	if uid <= 0 {
		return false, nil
	}
	var count int64
	err := r.db.WithContext(ctx).
		Table("admin_user u").
		Joins("INNER JOIN organization_position_role opr ON opr.position_id = u.position_id").
		Joins("INNER JOIN permission_role r ON r.id = opr.role_id AND r.deleted_at = 0").
		Where("u.uid = ? AND u.deleted_at = 0 AND r.is_protected = TRUE", uid).
		Count(&count).Error
	if err != nil {
		return false, xerr.WrapDBE(err, "check protected user role")
	}
	return count > 0, nil
}

func (r *Repo) RoleIsProtected(ctx context.Context, roleID int64) (bool, error) {
	if roleID <= 0 {
		return false, nil
	}
	var count int64
	err := r.db.WithContext(ctx).
		Table("permission_role").
		Where("id = ? AND deleted_at = 0 AND is_protected = TRUE", roleID).
		Count(&count).Error
	if err != nil {
		return false, xerr.WrapDBE(err, "check protected role")
	}
	return count > 0, nil
}

func (r *Repo) PositionHasProtectedRole(ctx context.Context, positionID int64) (bool, error) {
	if positionID <= 0 {
		return false, nil
	}
	var count int64
	err := r.db.WithContext(ctx).
		Table("organization_position_role opr").
		Joins("INNER JOIN permission_role r ON r.id = opr.role_id AND r.deleted_at = 0").
		Where("opr.position_id = ? AND r.is_protected = TRUE", positionID).
		Count(&count).Error
	if err != nil {
		return false, xerr.WrapDBE(err, "check protected position role")
	}
	return count > 0, nil
}

func (r *Repo) RoleIDsContainProtected(ctx context.Context, roleIDs []int64) (bool, error) {
	if len(roleIDs) == 0 {
		return false, nil
	}
	var count int64
	err := r.db.WithContext(ctx).
		Table("permission_role").
		Where("id IN ? AND deleted_at = 0 AND is_protected = TRUE", roleIDs).
		Count(&count).Error
	if err != nil {
		return false, xerr.WrapDBE(err, "check protected role ids")
	}
	return count > 0, nil
}

func (r *Repo) GetUserPositionInfo(ctx context.Context, uid int32) (int64, int32, error) {
	if uid <= 0 {
		return 0, 0, nil
	}
	var row struct {
		PositionID     int64 `gorm:"column:position_id"`
		ManagementRank int32 `gorm:"column:management_rank"`
	}
	err := r.db.WithContext(ctx).
		Table("admin_user u").
		Select("u.position_id, COALESCE(p.management_rank, 0) AS management_rank").
		Joins("LEFT JOIN organization_position p ON p.id = u.position_id AND p.deleted_at = 0").
		Where("u.uid = ? AND u.deleted_at = 0", uid).
		Scan(&row).Error
	if err != nil {
		return 0, 0, xerr.WrapDBE(err, "get user position info")
	}
	return row.PositionID, row.ManagementRank, nil
}

func (r *Repo) GetPositionManagementRank(ctx context.Context, positionID int64) (int32, error) {
	if positionID <= 0 {
		return 0, nil
	}
	var managementRank int32
	err := r.db.WithContext(ctx).
		Table("organization_position").
		Where("id = ? AND deleted_at = 0", positionID).
		Pluck("management_rank", &managementRank).Error
	if err != nil {
		return 0, xerr.WrapDBE(err, "get position management rank")
	}
	return managementRank, nil
}

func (r *Repo) ListEffectivePermissionKeysByUID(ctx context.Context, uid int32, delegableOnly bool) ([]string, error) {
	if uid <= 0 {
		return []string{}, nil
	}
	query := r.db.WithContext(ctx).
		Table("permission_menu m").
		Distinct("m.permission_key").
		Joins("INNER JOIN permission_role_menu prm ON prm.menu_id = m.id").
		Joins("INNER JOIN permission_role r ON r.id = prm.role_id AND r.deleted_at = 0").
		Joins("INNER JOIN organization_position_role opr ON opr.role_id = r.id").
		Joins("INNER JOIN admin_user u ON u.position_id = opr.position_id AND u.deleted_at = 0").
		Where("u.uid = ? AND m.deleted_at = 0 AND m.status = ? AND m.permission_key <> ''", uid, consts.PermissionStatusEnabled)
	if delegableOnly {
		query = query.Where("m.is_delegable = TRUE")
	}
	keys := make([]string, 0, 32)
	if err := query.Order("m.permission_key ASC").Pluck("m.permission_key", &keys).Error; err != nil {
		return nil, xerr.WrapDBE(err, "list user effective permissions")
	}
	return keys, nil
}

func (r *Repo) ListEffectivePermissionKeysByPositionID(ctx context.Context, positionID int64) ([]string, error) {
	if positionID <= 0 {
		return []string{}, nil
	}
	keys := make([]string, 0, 32)
	err := r.db.WithContext(ctx).
		Table("permission_menu m").
		Distinct("m.permission_key").
		Joins("INNER JOIN permission_role_menu prm ON prm.menu_id = m.id").
		Joins("INNER JOIN permission_role r ON r.id = prm.role_id AND r.deleted_at = 0").
		Joins("INNER JOIN organization_position_role opr ON opr.role_id = r.id").
		Where("opr.position_id = ? AND m.deleted_at = 0 AND m.status = ? AND m.permission_key <> ''", positionID, consts.PermissionStatusEnabled).
		Order("m.permission_key ASC").
		Pluck("m.permission_key", &keys).Error
	if err != nil {
		return nil, xerr.WrapDBE(err, "list position effective permissions")
	}
	return keys, nil
}

func (r *Repo) ListEffectivePermissionKeysByRoleID(ctx context.Context, roleID int64) ([]string, error) {
	if roleID <= 0 {
		return []string{}, nil
	}
	keys := make([]string, 0, 32)
	err := r.db.WithContext(ctx).
		Table("permission_menu m").
		Distinct("m.permission_key").
		Joins("INNER JOIN permission_role_menu prm ON prm.menu_id = m.id").
		Where("prm.role_id = ? AND m.deleted_at = 0 AND m.status = ? AND m.permission_key <> ''", roleID, consts.PermissionStatusEnabled).
		Order("m.permission_key ASC").
		Pluck("m.permission_key", &keys).Error
	if err != nil {
		return nil, xerr.WrapDBE(err, "list role effective permissions")
	}
	return keys, nil
}

func (r *Repo) ListEffectivePermissionKeysByUIDForMiddleware(ctx context.Context, uid int32) (map[string]struct{}, error) {
	keys, err := r.ListEffectivePermissionKeysByUID(ctx, uid, false)
	if err != nil {
		return nil, err
	}
	result := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		result[key] = struct{}{}
	}
	return result, nil
}
