package auth

import (
	"context"
	"strings"
	"time"

	"monorepo/internal/model"
	"monorepo/pkg/consts"
	"monorepo/pkg/db"
	"monorepo/pkg/xerr"

	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

type MenuItemRow struct {
	ID            int64  `gorm:"column:id"`
	ParentID      int64  `gorm:"column:parent_id"`
	Name          string `gorm:"column:name"`
	RoutePath     string `gorm:"column:route_path"`
	PermissionKey string `gorm:"column:permission_key"`
	Sort          int32  `gorm:"column:sort"`
}

type WarmTipRow struct {
	ID        int64  `gorm:"column:id"`
	TipType   string `gorm:"column:tip_type"`
	ContentZh string `gorm:"column:content_zh"`
	ContentEn string `gorm:"column:content_en"`
}

func NewRepo() *Repo {
	return &Repo{db: db.GetDatabase()}
}

func NewRepoWithDB(database *gorm.DB) *Repo {
	return &Repo{db: database}
}

func (r *Repo) GetActiveUserByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	var user model.AdminUser
	err := r.db.WithContext(ctx).
		Where("username = ? AND deleted_at = 0", username).
		First(&user).Error
	if err != nil {
		return nil, xerr.WrapDBNotFound(err, "user not found")
	}
	return &user, nil
}

func (r *Repo) GetDepartmentStatusByID(ctx context.Context, departmentID int64) (int32, bool, error) {
	if departmentID <= 0 {
		return 0, false, nil
	}
	var row struct {
		Status int32 `gorm:"column:status"`
	}
	err := r.db.WithContext(ctx).
		Table("organization_department").
		Select("status").
		Where("id = ? AND deleted_at = 0", departmentID).
		First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, xerr.WrapDBE(err, "query department status")
	}
	return row.Status, true, nil
}

func (r *Repo) GetPositionStatusByID(ctx context.Context, positionID int64) (int32, bool, error) {
	if positionID <= 0 {
		return 0, false, nil
	}
	var row struct {
		Status int32 `gorm:"column:status"`
	}
	err := r.db.WithContext(ctx).
		Table("organization_position").
		Select("status").
		Where("id = ? AND deleted_at = 0", positionID).
		First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, xerr.WrapDBE(err, "query position status")
	}
	return row.Status, true, nil
}

func (r *Repo) GetUserByUID(ctx context.Context, uid int32) (*model.AdminUser, error) {
	var user model.AdminUser
	err := r.db.WithContext(ctx).
		Where("uid = ? AND deleted_at = 0", uid).
		First(&user).Error
	if err != nil {
		return nil, xerr.WrapDBNotFound(err, "user not found")
	}
	return &user, nil
}

func (r *Repo) UpdateLoginMeta(ctx context.Context, uid int32, ip string, loginAt time.Time) error {
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Model(&model.AdminUser{}).
			Where("uid = ?", uid).
			Updates(map[string]any{
				"last_login_at": loginAt,
				"last_login_ip": ip,
			}).Error,
		"update user login metadata",
	)
}

func (r *Repo) GetSingleLoginSetting(ctx context.Context, uid int32) (bool, error) {
	setting, err := r.GetOrCreatePersonalSetting(ctx, uid)
	if err != nil {
		return false, err
	}
	return setting.LimitSingleLogin, nil
}

func (r *Repo) UpdateSingleLoginSetting(ctx context.Context, uid int32, enable bool) error {
	setting, err := r.GetOrCreatePersonalSetting(ctx, uid)
	if err != nil {
		return err
	}
	return r.UpdatePersonalSettings(ctx, uid, enable, setting.BackgroundImageURL, setting.Locale, setting.GlobalBackgroundApplyEnabled, setting.WarmTipIntervalMinutes, "")
}

func (r *Repo) UpdatePersonalSettings(ctx context.Context, uid int32, enable bool, backgroundImageURL, locale string, globalBackgroundApplyEnabled bool, warmTipIntervalMinutes int32, avatar string) error {
	if _, err := r.GetOrCreatePersonalSetting(ctx, uid); err != nil {
		return err
	}
	return xerr.WrapDBE(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Model(&model.AdminUserPersonalSetting{}).
			Where("uid = ?", uid).
			Updates(map[string]any{
				"limit_single_login":              enable,
				"background_image_url":            backgroundImageURL,
				"locale":                          locale,
				"global_background_apply_enabled": globalBackgroundApplyEnabled,
				"warm_tip_interval_minutes":       warmTipIntervalMinutes,
			}).Error; err != nil {
			return err
		}
		return tx.
			Model(&model.AdminUser{}).
			Where("uid = ? AND deleted_at = 0", uid).
			Update("avatar", strings.TrimSpace(avatar)).Error
	}), "update personal settings")
}

func (r *Repo) GetOrCreatePersonalSetting(ctx context.Context, uid int32) (*model.AdminUserPersonalSetting, error) {
	var setting model.AdminUserPersonalSetting
	err := r.db.WithContext(ctx).
		Where("uid = ?", uid).
		First(&setting).Error
	if err == nil {
		return &setting, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, xerr.WrapDBE(err, "query personal setting")
	}

	defaultSetting := &model.AdminUserPersonalSetting{
		UID:                          uid,
		LimitSingleLogin:             false,
		BackgroundImageURL:           "",
		Locale:                       "zh-CN",
		GlobalBackgroundApplyEnabled: false,
		WarmTipIntervalMinutes:       1440,
	}
	if createErr := r.db.WithContext(ctx).Create(defaultSetting).Error; createErr != nil {
		return nil, xerr.WrapDBE(createErr, "create personal setting")
	}
	return defaultSetting, nil
}

func (r *Repo) CreateSession(ctx context.Context, session *model.AdminUserSession) error {
	return xerr.WrapDBDuplicate(
		r.db.WithContext(ctx).Create(session).Error,
		"session already exists",
	)
}

func (r *Repo) GetSessionByID(ctx context.Context, sessionID string) (*model.AdminUserSession, error) {
	var session model.AdminUserSession
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		First(&session).Error
	if err != nil {
		return nil, xerr.WrapDBNotFound(err, "session not found")
	}
	return &session, nil
}

func (r *Repo) IsSessionActive(ctx context.Context, uid int32, sessionID, tokenHash string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.AdminUserSession{}).
		Where("uid = ? AND session_id = ? AND token_hash = ? AND status = ? AND expired_at > ?",
			uid, sessionID, tokenHash, consts.SessionStatusActive, time.Now()).
		Count(&count).Error
	if err != nil {
		return false, xerr.WrapDBE(err, "query active session")
	}
	return count > 0, nil
}

func (r *Repo) RevokeSession(ctx context.Context, uid int32, sessionID, reason string) error {
	now := time.Now()
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Model(&model.AdminUserSession{}).
			Where("uid = ? AND session_id = ? AND status = ?", uid, sessionID, consts.SessionStatusActive).
			Updates(map[string]any{
				"status":         consts.SessionStatusRevoked,
				"revoked_at":     now,
				"revoked_reason": reason,
			}).Error,
		"revoke session",
	)
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
		"revoke all sessions",
	)
}

func (r *Repo) RevokeOtherSessionsByUID(ctx context.Context, uid int32, keepSessionID, reason string) error {
	now := time.Now()
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Model(&model.AdminUserSession{}).
			Where("uid = ? AND session_id <> ? AND status = ?", uid, keepSessionID, consts.SessionStatusActive).
			Updates(map[string]any{
				"status":         consts.SessionStatusRevoked,
				"revoked_at":     now,
				"revoked_reason": reason,
			}).Error,
		"revoke other sessions",
	)
}

func (r *Repo) DeactivateUser(ctx context.Context, uid int32) error {
	now := time.Now()
	return xerr.WrapDBE(
		r.db.WithContext(ctx).
			Model(&model.AdminUser{}).
			Where("uid = ?", uid).
			Updates(map[string]any{
				"status":         consts.UserStatusDeactivated,
				"deactivated_at": now,
			}).Error,
		"deactivate user",
	)
}

func (r *Repo) ListSessionsByUID(ctx context.Context, uid int32, status string, limit int) ([]model.AdminUserSession, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows := make([]model.AdminUserSession, 0, limit)
	query := r.db.WithContext(ctx).Where("uid = ?", uid)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at desc").Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, xerr.WrapDBE(err, "list sessions")
	}
	return rows, nil
}

func (r *Repo) CreateAudit(ctx context.Context, audit *model.AdminUserLoginAudit) error {
	return xerr.WrapDBE(r.db.WithContext(ctx).Create(audit).Error, "create auth audit")
}

func (r *Repo) ListEnabledMenuRoutesByUID(ctx context.Context, uid int32) ([]string, error) {
	if uid <= 0 {
		return []string{}, nil
	}
	routes := make([]string, 0, 32)
	err := r.db.WithContext(ctx).Raw(`
SELECT x.route_path
FROM (
  SELECT
    m.route_path,
    MIN(m.sort) AS min_sort,
    MIN(m.id) AS min_id
  FROM permission_menu m
  INNER JOIN permission_role_menu prm ON prm.menu_id = m.id
  INNER JOIN permission_role r ON r.id = prm.role_id AND r.deleted_at = 0
  INNER JOIN (
    SELECT opr.role_id
    FROM admin_user u
    INNER JOIN organization_position_role opr ON opr.position_id = u.position_id
    WHERE u.uid = ? AND u.deleted_at = 0
  ) ur ON ur.role_id = prm.role_id
  WHERE m.deleted_at = 0
    AND m.status = ?
    AND m.route_path <> ''
  GROUP BY m.route_path
) x
ORDER BY x.min_sort ASC, x.min_id ASC
`, uid, consts.PermissionStatusEnabled).Scan(&routes).Error
	if err != nil {
		return nil, xerr.WrapDBE(err, "list enabled menu routes by uid")
	}
	return routes, nil
}

func (r *Repo) ListEnabledMenusByUID(ctx context.Context, uid int32) ([]MenuItemRow, error) {
	if uid <= 0 {
		return []MenuItemRow{}, nil
	}
	rows := make([]MenuItemRow, 0, 32)
	err := r.db.WithContext(ctx).Raw(`
SELECT
  m.id,
  m.parent_id,
  m.name,
  m.route_path,
  m.permission_key,
  m.sort
FROM permission_menu m
INNER JOIN permission_role_menu prm ON prm.menu_id = m.id
INNER JOIN permission_role r ON r.id = prm.role_id AND r.deleted_at = 0
INNER JOIN (
  SELECT opr.role_id
  FROM admin_user u
  INNER JOIN organization_position_role opr ON opr.position_id = u.position_id
  WHERE u.uid = ? AND u.deleted_at = 0
) ur ON ur.role_id = prm.role_id
WHERE m.deleted_at = 0
  AND m.status = ?
  AND m.menu_type IN (?, ?)
GROUP BY m.id, m.parent_id, m.name, m.route_path, m.permission_key, m.sort
ORDER BY m.sort ASC, m.id ASC
`, uid, consts.PermissionStatusEnabled, consts.PermissionMenuTypeDirectory, consts.PermissionMenuTypeMenu).Scan(&rows).Error
	if err != nil {
		return nil, xerr.WrapDBE(err, "list enabled menus by uid")
	}
	return rows, nil
}

func (r *Repo) GetEnabledWarmTip(ctx context.Context, seed int64) (*WarmTipRow, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.AccountWarmTip{}).
		Where("deleted_at = 0 AND status = ?", consts.PermissionStatusEnabled).
		Count(&count).Error; err != nil {
		return nil, xerr.WrapDBE(err, "count enabled warm tips")
	}
	if count <= 0 {
		return nil, nil
	}
	offset := int(seed)
	if offset < 0 {
		offset = -offset
	}
	offset = offset % int(count)
	var row WarmTipRow
	err := r.db.WithContext(ctx).
		Model(&model.AccountWarmTip{}).
		Select("id, tip_type, content_zh, content_en").
		Where("deleted_at = 0 AND status = ?", consts.PermissionStatusEnabled).
		Order("sort ASC, id ASC").
		Offset(offset).
		Limit(1).
		Scan(&row).Error
	if err != nil {
		return nil, xerr.WrapDBE(err, "get enabled warm tip")
	}
	if row.ID <= 0 {
		return nil, nil
	}
	return &row, nil
}
