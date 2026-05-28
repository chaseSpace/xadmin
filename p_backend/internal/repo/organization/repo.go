package organization

import (
	"context"
	"fmt"
	"strings"
	"time"

	"monorepo/internal/model"
	"monorepo/pkg/consts"
	"monorepo/pkg/db"
	"monorepo/pkg/xerr"
	commpb "monorepo/proto/xadminpb/commpb"

	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

type UserFilters struct {
	Keyword      string
	Phone        string
	Status       *int32
	DepartmentID *int64
	PositionID   *int64
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
}

type UserRow struct {
	UID                int32      `gorm:"column:uid"`
	Username           string     `gorm:"column:username"`
	DisplayName        string     `gorm:"column:display_name"`
	Avatar             string     `gorm:"column:avatar"`
	Email              string     `gorm:"column:email"`
	Phone              string     `gorm:"column:phone"`
	Status             int32      `gorm:"column:status"`
	DepartmentID       int64      `gorm:"column:department_id"`
	DepartmentName     string     `gorm:"column:department_name"`
	PositionID         int64      `gorm:"column:position_id"`
	PositionName       string     `gorm:"column:position_name"`
	RoleNamesCSV       string     `gorm:"column:role_names_csv"`
	LastLoginAt        *time.Time `gorm:"column:last_login_at"`
	LastLoginIP        string     `gorm:"column:last_login_ip"`
	ActiveSessionCount int32      `gorm:"column:active_session_count"`
}

type DepartmentRow struct {
	ID            int64      `gorm:"column:id"`
	ParentID      int64      `gorm:"column:parent_id"`
	Name          string     `gorm:"column:name"`
	Code          string     `gorm:"column:code"`
	Status        int32      `gorm:"column:status"`
	MemberCount   int32      `gorm:"column:member_count"`
	PositionCount int32      `gorm:"column:position_count"`
	UpdatedAt     *time.Time `gorm:"column:updated_at"`
}

type PositionFilters struct {
	Keyword      string
	DepartmentID *int64
	Level        string
	Status       *int32
}

type PositionRow struct {
	ID             int64      `gorm:"column:id"`
	Name           string     `gorm:"column:name"`
	Code           string     `gorm:"column:code"`
	DepartmentID   int64      `gorm:"column:department_id"`
	DepartmentName string     `gorm:"column:department_name"`
	Level          string     `gorm:"column:level"`
	Hc             int32      `gorm:"column:hc"`
	Staffed        int32      `gorm:"column:staffed"`
	RelatedCount   int32      `gorm:"column:related_count"`
	Status         int32      `gorm:"column:status"`
	UpdatedAt      *time.Time `gorm:"column:updated_at"`
	RoleIDsCSV     string     `gorm:"column:role_ids_csv"`
	RoleNamesCSV   string     `gorm:"column:role_names_csv"`
}

func NewRepo() *Repo {
	return &Repo{db: db.GetDatabase()}
}

func NewRepoWithDB(database *gorm.DB) *Repo {
	return &Repo{db: database}
}

func (r *Repo) ListUsers(ctx context.Context, page *commpb.PageArgs, sort []*commpb.SortArgs, filters UserFilters) ([]UserRow, int64, error) {
	rows := make([]UserRow, 0, page.GetPs())
	query := r.baseUsersQuery(ctx).
		Where("u.deleted_at = 0")
	query = applyUserFilters(query, filters)
	if len(sort) == 0 {
		// Avoid ambiguous `created_at` when joined tables contain same column.
		query = query.Order("u.created_at desc")
	}

	total, err := db.Paginate(
		query,
		page,
		sort,
		[]string{"u.uid", "u.username", "u.display_name", "u.status", "u.department_id", "u.position_id", "active_session_count", "u.last_login_at", "u.created_at"},
		&rows,
		db.PaginateArgs{AppendCreatedAtDesc: false},
	)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repo) ExportUsers(ctx context.Context, filters UserFilters) ([]UserRow, error) {
	rows := make([]UserRow, 0, 256)
	query := r.baseUsersQuery(ctx).
		Where("u.deleted_at = 0")
	query = applyUserFilters(query, filters)
	err := query.Order("u.created_at desc").Limit(5000).Find(&rows).Error
	if err != nil {
		return nil, xerr.WrapDBE(err, "export users")
	}
	return rows, nil
}

func (r *Repo) ListSessionsByUID(ctx context.Context, uid int32, status string, limit int) ([]model.AdminUserSession, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	rows := make([]model.AdminUserSession, 0, limit)
	query := r.db.WithContext(ctx).Where("uid = ?", uid)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at desc").Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, xerr.WrapDBE(err, "list user sessions")
	}
	return rows, nil
}

func (r *Repo) RevokeAllSessionsByUID(ctx context.Context, uid int32, reason string) error {
	now := time.Now()
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Model(&model.AdminUserSession{}).
			Where("uid = ? AND status = ?", uid, consts.SessionStatusActive).
			Updates(map[string]any{
				"status":         consts.SessionStatusRevoked,
				"revoked_at":     now,
				"revoked_reason": reason,
			}).Error,
		"revoke all user sessions",
	)
}

func (r *Repo) RevokeSessionsByDepartmentID(ctx context.Context, departmentID int64, reason string) error {
	now := time.Now()
	subQuery := r.db.WithContext(ctx).
		Model(&model.AdminUser{}).
		Select("uid").
		Where("department_id = ? AND deleted_at = 0", departmentID)
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Model(&model.AdminUserSession{}).
			Where("uid IN (?) AND status = ?", subQuery, consts.SessionStatusActive).
			Updates(map[string]any{
				"status":         consts.SessionStatusRevoked,
				"revoked_at":     now,
				"revoked_reason": reason,
			}).Error,
		"revoke department sessions",
	)
}

func (r *Repo) RevokeSessionsByPositionID(ctx context.Context, positionID int64, reason string) error {
	now := time.Now()
	subQuery := r.db.WithContext(ctx).
		Model(&model.AdminUser{}).
		Select("uid").
		Where("position_id = ? AND deleted_at = 0", positionID)
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Model(&model.AdminUserSession{}).
			Where("uid IN (?) AND status = ?", subQuery, consts.SessionStatusActive).
			Updates(map[string]any{
				"status":         consts.SessionStatusRevoked,
				"revoked_at":     now,
				"revoked_reason": reason,
			}).Error,
		"revoke position sessions",
	)
}

func (r *Repo) baseUsersQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("admin_user u").
		Select(`
u.uid,
  u.username,
  u.display_name,
  u.avatar,
  u.email,
  u.phone,
  u.status,
  u.department_id,
  COALESCE(d.name, '') AS department_name,
  u.position_id,
  COALESCE(p.name, '') AS position_name,
  COALESCE(pr.role_names_csv, '') AS role_names_csv,
  u.last_login_at,
  u.last_login_ip,
  COALESCE(s.active_session_count, 0) AS active_session_count
`).
		Joins("LEFT JOIN organization_department d ON d.id = u.department_id AND d.deleted_at = 0").
		Joins("LEFT JOIN organization_position p ON p.id = u.position_id AND p.deleted_at = 0").
		Joins(`
LEFT JOIN (
  SELECT opr.position_id, STRING_AGG(r.role_name, ',' ORDER BY r.id) AS role_names_csv
  FROM organization_position_role opr
  JOIN permission_role r ON r.id = opr.role_id AND r.deleted_at = 0
  GROUP BY opr.position_id
) pr ON pr.position_id = u.position_id
`).
		Joins(`
LEFT JOIN (
  SELECT uid, COUNT(1) AS active_session_count
  FROM admin_user_session
  WHERE status = ? AND expired_at > ?
  GROUP BY uid
) s ON s.uid = u.uid
`, consts.SessionStatusActive, time.Now())
}

func applyUserFilters(query *gorm.DB, filters UserFilters) *gorm.DB {
	if kw := strings.TrimSpace(filters.Keyword); kw != "" {
		query = query.Where("(u.username LIKE ? OR u.display_name LIKE ?)", "%"+kw+"%", "%"+kw+"%")
	}
	if phone := strings.TrimSpace(filters.Phone); phone != "" {
		query = query.Where("u.phone LIKE ?", "%"+phone+"%")
	}
	if filters.Status != nil {
		query = query.Where("u.status = ?", *filters.Status)
	}
	if filters.DepartmentID != nil {
		query = query.Where("u.department_id = ?", *filters.DepartmentID)
	}
	if filters.PositionID != nil {
		query = query.Where("u.position_id = ?", *filters.PositionID)
	}
	if filters.CreatedFrom != nil {
		query = query.Where("u.created_at >= ?", *filters.CreatedFrom)
	}
	if filters.CreatedTo != nil {
		query = query.Where("u.created_at <= ?", *filters.CreatedTo)
	}
	return query
}

func (r *Repo) CreateUser(ctx context.Context, user *model.AdminUser) error {
	return xerr.WrapDBDuplicate(
		r.db.WithContext(ctx).Create(user).Error,
		"user already exists",
	)
}

func (r *Repo) UpdateUserByUID(ctx context.Context, uid int32, updates map[string]any) error {
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Model(&model.AdminUser{}).
			Where("uid = ? AND deleted_at = 0", uid).
			Updates(updates).Error,
		"update organization user",
	)
}

func (r *Repo) BatchUpdateUsersPosition(ctx context.Context, uids []int32, departmentID int64, positionID int64) error {
	if len(uids) == 0 {
		return nil
	}
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Model(&model.AdminUser{}).
			Where("uid IN ? AND deleted_at = 0", uids).
			Updates(map[string]any{
				"department_id": departmentID,
				"position_id":   positionID,
			}).Error,
		"batch transfer organization users",
	)
}

func (r *Repo) SoftDeleteUserByUID(ctx context.Context, uid int32) error {
	deletedAt := time.Now().Unix()
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Model(&model.AdminUser{}).
			Where("uid = ? AND deleted_at = 0", uid).
			Update("deleted_at", deletedAt).Error,
		"soft delete organization user",
	)
}

func (r *Repo) GetUserByUID(ctx context.Context, uid int32) (*model.AdminUser, error) {
	var user model.AdminUser
	err := r.db.WithContext(ctx).
		Where("uid = ? AND deleted_at = 0", uid).
		First(&user).Error
	if err != nil {
		return nil, xerr.WrapDBNotFound(err, "organization user not found")
	}
	return &user, nil
}

func (r *Repo) NextUID(ctx context.Context) (int32, error) {
	var maxUID int32
	err := r.db.WithContext(ctx).
		Model(&model.AdminUser{}).
		Where("deleted_at = 0").
		Select("COALESCE(MAX(uid), 10000)").
		Scan(&maxUID).Error
	if err != nil {
		return 0, xerr.WrapDBE(err, "query max uid")
	}
	return maxUID + 1, nil
}

func (r *Repo) ListDepartments(ctx context.Context) ([]DepartmentRow, error) {
	rows := make([]DepartmentRow, 0, 64)
	err := r.db.WithContext(ctx).
		Table("organization_department d").
		Select(`
d.id,
d.parent_id,
d.name,
d.code,
d.status,
COUNT(DISTINCT u.id) AS member_count,
COUNT(DISTINCT p.id) AS position_count,
d.updated_at
`).
		Joins("LEFT JOIN admin_user u ON u.department_id = d.id AND u.deleted_at = 0").
		Joins("LEFT JOIN organization_position p ON p.department_id = d.id AND p.deleted_at = 0").
		Where("d.deleted_at = 0").
		Group("d.id, d.parent_id, d.name, d.code, d.status, d.updated_at").
		Order("d.sort asc, d.id asc").
		Find(&rows).Error
	if err != nil {
		return nil, xerr.WrapDBE(err, "list departments")
	}
	return rows, nil
}

func (r *Repo) GetDepartmentByID(ctx context.Context, id int64) (*DepartmentRow, error) {
	var row DepartmentRow
	err := r.db.WithContext(ctx).
		Table("organization_department d").
		Select(`
d.id,
d.parent_id,
d.name,
d.code,
d.status,
COUNT(DISTINCT u.id) AS member_count,
COUNT(DISTINCT p.id) AS position_count,
d.updated_at
`).
		Joins("LEFT JOIN admin_user u ON u.department_id = d.id AND u.deleted_at = 0").
		Joins("LEFT JOIN organization_position p ON p.department_id = d.id AND p.deleted_at = 0").
		Where("d.id = ? AND d.deleted_at = 0", id).
		Group("d.id, d.parent_id, d.name, d.code, d.status, d.updated_at").
		First(&row).Error
	if err != nil {
		return nil, xerr.WrapDBNotFound(err, "organization department not found")
	}
	return &row, nil
}

func (r *Repo) CreateDepartment(ctx context.Context, department *model.OrganizationDepartment) error {
	return xerr.WrapDBDuplicate(
		r.db.WithContext(ctx).Create(department).Error,
		"organization department already exists",
	)
}

func (r *Repo) UpdateDepartmentByID(ctx context.Context, id int64, updates map[string]any) error {
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Model(&model.OrganizationDepartment{}).
			Where("id = ? AND deleted_at = 0", id).
			Updates(updates).Error,
		"update organization department",
	)
}

func (r *Repo) HasChildDepartment(ctx context.Context, parentID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.OrganizationDepartment{}).
		Where("parent_id = ? AND deleted_at = 0", parentID).
		Count(&count).Error
	if err != nil {
		return false, xerr.WrapDBE(err, "count child departments")
	}
	return count > 0, nil
}

func (r *Repo) CountPositionsByDepartmentID(ctx context.Context, departmentID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.OrganizationPosition{}).
		Where("department_id = ? AND deleted_at = 0", departmentID).
		Count(&count).Error
	if err != nil {
		return 0, xerr.WrapDBE(err, "count department positions")
	}
	return count, nil
}

func (r *Repo) CountUsersByDepartmentID(ctx context.Context, departmentID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.AdminUser{}).
		Where("department_id = ? AND deleted_at = 0", departmentID).
		Count(&count).Error
	if err != nil {
		return 0, xerr.WrapDBE(err, "count department users")
	}
	return count, nil
}

func (r *Repo) SoftDeleteDepartmentByID(ctx context.Context, id int64) error {
	deletedAt := time.Now().Unix()
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Model(&model.OrganizationDepartment{}).
			Where("id = ? AND deleted_at = 0", id).
			Update("deleted_at", deletedAt).Error,
		"soft delete organization department",
	)
}

func (r *Repo) ListPositions(
	ctx context.Context,
	page *commpb.PageArgs,
	sort []*commpb.SortArgs,
	filters PositionFilters,
) ([]PositionRow, int64, error) {
	rows := make([]PositionRow, 0, page.GetPs())
	query := r.db.WithContext(ctx).
		Table("organization_position p").
		Select(`
p.id,
p.name,
p.code,
p.department_id,
COALESCE(d.name, '') as department_name,
p.level,
COUNT(DISTINCT u.id) AS hc,
COUNT(DISTINCT CASE WHEN u.status = ? THEN u.id END) AS staffed,
COUNT(DISTINCT u.id) AS related_count,
p.status,
p.updated_at,
COALESCE(pr.role_ids_csv, '') as role_ids_csv,
COALESCE(pr.role_names_csv, '') as role_names_csv
`, consts.UserStatusActive).
		Joins("LEFT JOIN organization_department d ON d.id = p.department_id AND d.deleted_at = 0").
		Joins("LEFT JOIN admin_user u ON u.position_id = p.id AND u.deleted_at = 0", consts.UserStatusActive).
		Joins(`
LEFT JOIN (
  SELECT pr.position_id,
         STRING_AGG(pr.role_id::text, ',' ORDER BY pr.role_id) AS role_ids_csv,
         STRING_AGG(r.role_name, ',' ORDER BY pr.role_id) AS role_names_csv
  FROM organization_position_role pr
  INNER JOIN permission_role r ON r.id = pr.role_id AND r.deleted_at = 0
  GROUP BY pr.position_id
) pr ON pr.position_id = p.id
`).
		Where("p.deleted_at = 0").
		Group("p.id, p.name, p.code, p.department_id, d.name, p.level, p.status, p.updated_at, pr.role_ids_csv, pr.role_names_csv")
	query = applyPositionFilters(query, filters)
	if len(sort) == 0 {
		// Avoid ambiguous `created_at` when joined tables contain same column.
		query = query.Order("p.created_at desc")
	}

	total, err := db.Paginate(
		query,
		page,
		sort,
		[]string{"p.id", "p.name", "p.level", "p.status", "p.updated_at"},
		&rows,
		db.PaginateArgs{AppendCreatedAtDesc: false},
	)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repo) GetPositionByID(ctx context.Context, id int64) (*PositionRow, error) {
	var row PositionRow
	err := r.db.WithContext(ctx).
		Table("organization_position p").
		Select(`
p.id,
p.name,
p.code,
p.department_id,
COALESCE(d.name, '') as department_name,
p.level,
COUNT(DISTINCT u.id) AS hc,
COUNT(DISTINCT CASE WHEN u.status = ? THEN u.id END) AS staffed,
COUNT(DISTINCT u.id) AS related_count,
p.status,
p.updated_at,
COALESCE(pr.role_ids_csv, '') as role_ids_csv,
COALESCE(pr.role_names_csv, '') as role_names_csv
`, consts.UserStatusActive).
		Joins("LEFT JOIN organization_department d ON d.id = p.department_id AND d.deleted_at = 0").
		Joins("LEFT JOIN admin_user u ON u.position_id = p.id AND u.deleted_at = 0", consts.UserStatusActive).
		Joins(`
LEFT JOIN (
  SELECT pr.position_id,
         STRING_AGG(pr.role_id::text, ',' ORDER BY pr.role_id) AS role_ids_csv,
         STRING_AGG(r.role_name, ',' ORDER BY pr.role_id) AS role_names_csv
  FROM organization_position_role pr
  INNER JOIN permission_role r ON r.id = pr.role_id AND r.deleted_at = 0
  GROUP BY pr.position_id
) pr ON pr.position_id = p.id
`).
		Where("p.id = ? AND p.deleted_at = 0", id).
		Group("p.id, p.name, p.code, p.department_id, d.name, p.level, p.status, p.updated_at, pr.role_ids_csv, pr.role_names_csv").
		First(&row).Error
	if err != nil {
		return nil, xerr.WrapDBNotFound(err, "organization position not found")
	}
	return &row, nil
}

func (r *Repo) CreatePosition(ctx context.Context, position *model.OrganizationPosition) error {
	return xerr.WrapDBDuplicate(
		r.db.WithContext(ctx).Create(position).Error,
		"organization position already exists",
	)
}

func (r *Repo) UpdatePositionByID(ctx context.Context, id int64, updates map[string]any) error {
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Model(&model.OrganizationPosition{}).
			Where("id = ? AND deleted_at = 0", id).
			Updates(updates).Error,
		"update organization position",
	)
}

func (r *Repo) SoftDeletePositionByID(ctx context.Context, id int64) error {
	deletedAt := time.Now().Unix()
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Model(&model.OrganizationPosition{}).
			Where("id = ? AND deleted_at = 0", id).
			Update("deleted_at", deletedAt).Error,
		"soft delete organization position",
	)
}

func (r *Repo) CountValidRolesByIDs(ctx context.Context, roleIDs []int64) (int64, error) {
	if len(roleIDs) == 0 {
		return 0, nil
	}
	var count int64
	err := r.db.WithContext(ctx).
		Table("permission_role").
		Where("id IN ? AND deleted_at = 0", roleIDs).
		Count(&count).Error
	if err != nil {
		return 0, xerr.WrapDBE(err, "count valid roles")
	}
	return count, nil
}

func (r *Repo) SyncPositionRoles(ctx context.Context, positionID int64, roleIDs []int64) error {
	return xerr.WrapDBE(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("position_id = ?", positionID).Delete(&model.OrganizationPositionRole{}).Error; err != nil {
			return fmt.Errorf("delete position roles: %w", err)
		}
		if len(roleIDs) == 0 {
			return nil
		}
		rows := make([]model.OrganizationPositionRole, 0, len(roleIDs))
		for _, roleID := range roleIDs {
			rows = append(rows, model.OrganizationPositionRole{
				PositionID: positionID,
				RoleID:     roleID,
			})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return fmt.Errorf("insert position roles: %w", err)
		}
		return nil
	}), "sync position roles")
}

func applyPositionFilters(query *gorm.DB, filters PositionFilters) *gorm.DB {
	if kw := strings.TrimSpace(filters.Keyword); kw != "" {
		query = query.Where("(p.name LIKE ? OR p.code LIKE ?)", "%"+kw+"%", "%"+kw+"%")
	}
	if filters.DepartmentID != nil {
		query = query.Where("p.department_id = ?", *filters.DepartmentID)
	}
	if level := strings.TrimSpace(filters.Level); level != "" {
		query = query.Where("p.level = ?", level)
	}
	if filters.Status != nil {
		query = query.Where("p.status = ?", *filters.Status)
	}
	return query
}
