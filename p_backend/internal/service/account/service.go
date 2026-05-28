package account

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"monorepo/config"
	"monorepo/internal/model"
	authrepo "monorepo/internal/repo/auth"
	"monorepo/internal/support/requestmeta"
	"monorepo/pkg/xerr"
	xadmin "monorepo/proto/xadminpb"
)

type Service interface {
	GetPersonalSettings(ctx context.Context, uid int32) (*xadmin.AuthPersonalSettingsResp, error)
	UpdatePersonalSettings(ctx context.Context, uid int32, sessionID string, req *xadmin.AuthUpdatePersonalSettingsReq, ip, userAgent, traceID string) (*xadmin.AuthPersonalSettingsResp, error)
	GetMyProfile(ctx context.Context, uid int32) (*xadmin.AuthMeProfileResp, error)
	GetSystemSettings(ctx context.Context) (*xadmin.AuthSystemSettingsResp, error)
	UpdateSystemSettings(ctx context.Context, req *xadmin.AuthUpdateSystemSettingsReq) (*xadmin.AuthSystemSettingsResp, error)
}

type service struct {
	repo *authrepo.Repo
}

type systemSettingsState struct {
	SiteName                string
	Locale                  string
	Timezone                string
	LoginLockThreshold      int32
	PasswordMinLength       int32
	SessionTimeoutMinutes   int32
	PasswordPolicy          []string
	GlobalWatermarkEnabled  bool
	GlobalWatermarkFontSize int32
}

var (
	systemSettingsMu sync.RWMutex
	systemSettings   = systemSettingsState{
		SiteName:                "XAdmin 管理后台",
		Locale:                  "zh-CN",
		Timezone:                "Asia/Shanghai",
		LoginLockThreshold:      5,
		PasswordMinLength:       8,
		SessionTimeoutMinutes:   30,
		PasswordPolicy:          []string{"uppercase", "number"},
		GlobalWatermarkEnabled:  false,
		GlobalWatermarkFontSize: 16,
	}
)

func NewService() Service {
	return &service{repo: authrepo.NewRepo()}
}

func NewServiceWithRepo(repo *authrepo.Repo) Service {
	return &service{repo: repo}
}

func currentServerTimezone() string {
	if loc := time.Local; loc != nil {
		if timezone := strings.TrimSpace(loc.String()); timezone != "" {
			return timezone
		}
	}
	return "Asia/Shanghai"
}

func (s *service) GetPersonalSettings(ctx context.Context, uid int32) (*xadmin.AuthPersonalSettingsResp, error) {
	if uid <= 0 {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "account.invalid_user")
	}
	setting, err := s.repo.GetOrCreatePersonalSetting(ctx, uid)
	if err != nil {
		return nil, err
	}
	user, err := s.repo.GetUserByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return &xadmin.AuthPersonalSettingsResp{
		LimitSingleLogin:             setting.LimitSingleLogin,
		BackgroundImageUrl:           strings.TrimSpace(setting.BackgroundImageURL),
		Locale:                       strings.TrimSpace(setting.Locale),
		GlobalBackgroundApplyEnabled: setting.GlobalBackgroundApplyEnabled,
		Avatar:                       strings.TrimSpace(user.Avatar),
		WarmTipIntervalMinutes:       normalizeWarmTipInterval(setting.WarmTipIntervalMinutes),
	}, nil
}

func (s *service) UpdatePersonalSettings(ctx context.Context, uid int32, sessionID string, req *xadmin.AuthUpdatePersonalSettingsReq, ip, userAgent, traceID string) (*xadmin.AuthPersonalSettingsResp, error) {
	if uid <= 0 {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "account.invalid_user")
	}
	if sessionID == "" {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "account.invalid_session")
	}

	enabled := req.GetLimitSingleLogin()
	backgroundImageURL := strings.TrimSpace(req.GetBackgroundImageUrl())
	locale := strings.TrimSpace(req.GetLocale())
	avatar := strings.TrimSpace(req.GetAvatar())
	globalBackgroundApplyEnabled := req.GetGlobalBackgroundApplyEnabled()
	warmTipIntervalMinutes := normalizeWarmTipInterval(req.GetWarmTipIntervalMinutes())
	hasGlobalBackgroundApplyEnabled := req.GlobalBackgroundApplyEnabled != nil
	var currentSetting *model.AdminUserPersonalSetting
	loadCurrentSetting := func() (*model.AdminUserPersonalSetting, error) {
		if currentSetting != nil {
			return currentSetting, nil
		}
		setting, err := s.repo.GetOrCreatePersonalSetting(ctx, uid)
		if err != nil {
			return nil, err
		}
		currentSetting = setting
		return setting, nil
	}
	if locale == "" {
		setting, err := loadCurrentSetting()
		if err != nil {
			return nil, err
		}
		locale = strings.TrimSpace(setting.Locale)
		if hasGlobalBackgroundApplyEnabled {
			globalBackgroundApplyEnabled = req.GetGlobalBackgroundApplyEnabled()
		} else {
			globalBackgroundApplyEnabled = setting.GlobalBackgroundApplyEnabled
		}
		if warmTipIntervalMinutes == 0 {
			warmTipIntervalMinutes = normalizeWarmTipInterval(setting.WarmTipIntervalMinutes)
		}
	}
	if locale == "" {
		locale = "zh-CN"
	}
	if !hasGlobalBackgroundApplyEnabled {
		setting, err := loadCurrentSetting()
		if err != nil {
			return nil, err
		}
		globalBackgroundApplyEnabled = setting.GlobalBackgroundApplyEnabled
		if warmTipIntervalMinutes == 0 {
			warmTipIntervalMinutes = normalizeWarmTipInterval(setting.WarmTipIntervalMinutes)
		}
	}
	if warmTipIntervalMinutes == 0 {
		setting, err := loadCurrentSetting()
		if err != nil {
			return nil, err
		}
		warmTipIntervalMinutes = normalizeWarmTipInterval(setting.WarmTipIntervalMinutes)
		if warmTipIntervalMinutes == 0 {
			warmTipIntervalMinutes = 1440
		}
	}
	if err := s.repo.UpdatePersonalSettings(ctx, uid, enabled, backgroundImageURL, locale, globalBackgroundApplyEnabled, warmTipIntervalMinutes, avatar); err != nil {
		return nil, err
	}
	if enabled {
		if err := s.repo.RevokeOtherSessionsByUID(ctx, uid, sessionID, "single_login_setting_enabled"); err != nil {
			return nil, err
		}
	}
	_ = s.audit(ctx, uid, "update_personal_settings", "success", ip, userAgent, traceID, "enabled="+strconv.FormatBool(enabled))
	return &xadmin.AuthPersonalSettingsResp{
		LimitSingleLogin:             enabled,
		BackgroundImageUrl:           backgroundImageURL,
		Locale:                       locale,
		GlobalBackgroundApplyEnabled: globalBackgroundApplyEnabled,
		Avatar:                       avatar,
		WarmTipIntervalMinutes:       warmTipIntervalMinutes,
	}, nil
}

func normalizeWarmTipInterval(value int32) int32 {
	switch value {
	case 10, 60, 360, 720, 1440:
		return value
	default:
		return 0
	}
}

func (s *service) GetMyProfile(ctx context.Context, uid int32) (*xadmin.AuthMeProfileResp, error) {
	if uid <= 0 {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "account.invalid_user")
	}
	user, err := s.repo.GetUserByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	menuRoutes, err := s.repo.ListEnabledMenuRoutesByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	menuRows, err := s.repo.ListEnabledMenusByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	setting, err := s.repo.GetOrCreatePersonalSetting(ctx, uid)
	if err != nil {
		return nil, err
	}
	intervalMinutes := normalizeWarmTipInterval(setting.WarmTipIntervalMinutes)
	if intervalMinutes == 0 {
		intervalMinutes = 1440
	}
	warmTip, err := s.repo.GetEnabledWarmTip(ctx, int64(uid)+time.Now().Unix()/int64(intervalMinutes*60))
	if err != nil {
		return nil, err
	}
	return &xadmin.AuthMeProfileResp{
		Uid:         user.UID,
		Username:    strings.TrimSpace(user.Username),
		DisplayName: strings.TrimSpace(user.DisplayName),
		Avatar:      strings.TrimSpace(user.Avatar),
		Email:       strings.TrimSpace(user.Email),
		Phone:       strings.TrimSpace(user.Phone),
		MenuRoutes:  menuRoutes,
		MenuItems:   buildAuthMenuTree(menuRows),
		WarmTip:     buildAuthWarmTip(warmTip),
	}, nil
}

func buildAuthWarmTip(row *authrepo.WarmTipRow) *xadmin.AuthWarmTip {
	if row == nil {
		return nil
	}
	return &xadmin.AuthWarmTip{
		Id:        row.ID,
		TipType:   strings.TrimSpace(row.TipType),
		ContentZh: strings.TrimSpace(row.ContentZh),
		ContentEn: strings.TrimSpace(row.ContentEn),
	}
}

func buildAuthMenuTree(rows []authrepo.MenuItemRow) []*xadmin.AuthMenuItem {
	nodesByID := make(map[int64]*xadmin.AuthMenuItem, len(rows))
	roots := make([]*xadmin.AuthMenuItem, 0, len(rows))
	for _, row := range rows {
		nodesByID[row.ID] = &xadmin.AuthMenuItem{
			Id:            row.ID,
			ParentId:      row.ParentID,
			Name:          strings.TrimSpace(row.Name),
			RoutePath:     strings.TrimSpace(row.RoutePath),
			PermissionKey: strings.TrimSpace(row.PermissionKey),
			Sort:          row.Sort,
			Children:      []*xadmin.AuthMenuItem{},
			Icon:          resolveAuthMenuIcon(row.RoutePath, row.PermissionKey),
		}
	}
	for _, row := range rows {
		node := nodesByID[row.ID]
		if node == nil {
			continue
		}
		parent := nodesByID[row.ParentID]
		if row.ParentID > 0 && parent != nil {
			parent.Children = append(parent.Children, node)
			continue
		}
		roots = append(roots, node)
	}
	return roots
}

func resolveAuthMenuIcon(routePath, permissionKey string) string {
	key := strings.TrimSpace(routePath)
	if key == "" {
		key = strings.Split(strings.TrimSpace(permissionKey), ".")[0]
	}
	switch key {
	case "/":
		return "dashboard"
	case "organization":
		return "apartment"
	case "/organization/departments":
		return "deployment-unit"
	case "/organization/users", "/business/users":
		return "usergroup-add"
	case "/organization/positions":
		return "solution"
	case "business":
		return "database"
	case "/business/user-punishments", "/system/ip-blacklist":
		return "exclamation-circle"
	case "resource":
		return "folder-open"
	case "/resource/files":
		return "file-text"
	case "permission":
		return "safety-certificate"
	case "/permission/role-permissions":
		return "profile"
	case "/permission/menu-permissions":
		return "key"
	case "system", "/system/settings":
		return "setting"
	case "/system/audit-logs":
		return "audit"
	case "/system/warm-tips":
		return "heart"
	default:
		return ""
	}
}

func (s *service) GetSystemSettings(ctx context.Context) (*xadmin.AuthSystemSettingsResp, error) {
	_ = ctx
	systemSettingsMu.RLock()
	defer systemSettingsMu.RUnlock()
	return &xadmin.AuthSystemSettingsResp{
		SiteName:                systemSettings.SiteName,
		Locale:                  systemSettings.Locale,
		Timezone:                systemSettings.Timezone,
		ServerTimezone:          currentServerTimezone(),
		LoginLockThreshold:      systemSettings.LoginLockThreshold,
		PasswordMinLength:       systemSettings.PasswordMinLength,
		SessionTimeoutMinutes:   systemSettings.SessionTimeoutMinutes,
		PasswordPolicy:          append([]string{}, systemSettings.PasswordPolicy...),
		GlobalWatermarkEnabled:  systemSettings.GlobalWatermarkEnabled,
		GlobalWatermarkFontSize: systemSettings.GlobalWatermarkFontSize,
	}, nil
}

func (s *service) UpdateSystemSettings(ctx context.Context, req *xadmin.AuthUpdateSystemSettingsReq) (*xadmin.AuthSystemSettingsResp, error) {
	_ = ctx
	timeoutMinutes := req.GetSessionTimeoutMinutes()
	cfg := config.GetConfig()
	if cfg != nil {
		minutes := int64(timeoutMinutes)
		hours := minutes / 60
		if minutes%60 != 0 {
			hours += 1
		}
		if hours <= 0 {
			hours = 1
		}
		cfg.App.Auth.TokenTTLMinutes = minutes
		cfg.App.Auth.TokenTTLHours = hours
	}
	systemSettingsMu.Lock()
	systemSettings = systemSettingsState{
		SiteName:                truncateSystemSiteName(strings.TrimSpace(req.GetSiteName())),
		Locale:                  strings.TrimSpace(req.GetLocale()),
		Timezone:                strings.TrimSpace(req.GetTimezone()),
		LoginLockThreshold:      req.GetLoginLockThreshold(),
		PasswordMinLength:       req.GetPasswordMinLength(),
		SessionTimeoutMinutes:   timeoutMinutes,
		PasswordPolicy:          append([]string{}, req.GetPasswordPolicy()...),
		GlobalWatermarkEnabled:  req.GetGlobalWatermarkEnabled(),
		GlobalWatermarkFontSize: req.GetGlobalWatermarkFontSize(),
	}
	systemSettingsMu.Unlock()
	return &xadmin.AuthSystemSettingsResp{
		SiteName:                systemSettings.SiteName,
		Locale:                  systemSettings.Locale,
		Timezone:                systemSettings.Timezone,
		ServerTimezone:          currentServerTimezone(),
		LoginLockThreshold:      systemSettings.LoginLockThreshold,
		PasswordMinLength:       systemSettings.PasswordMinLength,
		SessionTimeoutMinutes:   systemSettings.SessionTimeoutMinutes,
		PasswordPolicy:          append([]string{}, systemSettings.PasswordPolicy...),
		GlobalWatermarkEnabled:  systemSettings.GlobalWatermarkEnabled,
		GlobalWatermarkFontSize: systemSettings.GlobalWatermarkFontSize,
	}, nil
}

func truncateSystemSiteName(value string) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= 20 {
		return string(runes)
	}
	return string(runes[:20])
}

func (s *service) audit(ctx context.Context, uid int32, action, result, ip, userAgent, traceID, detail string) error {
	return s.repo.CreateAudit(ctx, &model.AdminUserLoginAudit{
		UID:       uid,
		Action:    action,
		Result:    result,
		TraceID:   traceID,
		SourceIP:  ip,
		Duration:  requestmeta.DurationString(ctx),
		UserAgent: userAgent,
		Detail:    detail,
	})
}
