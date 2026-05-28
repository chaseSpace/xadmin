package system

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/netip"
	"strings"
	"time"

	systemrepo "monorepo/internal/repo/system"
	"monorepo/internal/support/alertbot"
	"monorepo/internal/support/ipblacklist"
	"monorepo/internal/support/timefmt"
	"monorepo/pkg/xerr"
	xadmin "monorepo/proto/xadminpb"
	commpb "monorepo/proto/xadminpb/commpb"
)

type Service interface {
	ListAuditLogs(ctx context.Context, req *xadmin.SystemAuditLogsReq) (*xadmin.SystemAuditLogsResp, error)
	GetAuditLog(ctx context.Context, req *xadmin.SystemAuditLogDetailReq) (*xadmin.SystemAuditLogDetailResp, error)
	ExportAuditLogs(ctx context.Context, req *xadmin.SystemAuditLogsReq) ([]byte, error)
	ApplyAuditLogRetention(ctx context.Context, req *xadmin.SystemAuditLogRetentionReq) (*xadmin.SystemAuditLogRetentionResp, error)
	ListIPBlacklist(ctx context.Context, req *xadmin.SystemIPBlacklistReq) (*xadmin.SystemIPBlacklistResp, error)
	ListIPBlacklistCreators(ctx context.Context) (*xadmin.SystemIPBlacklistCreatorsResp, error)
	CreateIPBlacklist(ctx context.Context, req *xadmin.SystemCreateIPBlacklistReq) (*xadmin.SystemActionResp, error)
	UpdateIPBlacklist(ctx context.Context, req *xadmin.SystemUpdateIPBlacklistReq) (*xadmin.SystemActionResp, error)
	UnblockIPBlacklist(ctx context.Context, req *xadmin.SystemIPBlacklistActionReq) (*xadmin.SystemActionResp, error)
	DeleteIPBlacklist(ctx context.Context, req *xadmin.SystemIPBlacklistActionReq) (*xadmin.SystemActionResp, error)
	BatchUnblockIPBlacklist(ctx context.Context, req *xadmin.SystemIPBlacklistBatchUnblockReq) (*xadmin.SystemActionResp, error)
	ImportIPBlacklist(ctx context.Context, req *xadmin.SystemIPBlacklistImportReq, creator string) (*xadmin.SystemActionResp, error)
	ListWarmTips(ctx context.Context, req *xadmin.SystemWarmTipsReq) (*xadmin.SystemWarmTipsResp, error)
	CreateWarmTip(ctx context.Context, req *xadmin.SystemCreateWarmTipReq) (*xadmin.SystemActionResp, error)
	UpdateWarmTip(ctx context.Context, req *xadmin.SystemUpdateWarmTipReq) (*xadmin.SystemActionResp, error)
	DeleteWarmTip(ctx context.Context, req *xadmin.SystemWarmTipActionReq) (*xadmin.SystemActionResp, error)
	ListAlertBots(ctx context.Context, req *xadmin.SystemAlertBotListReq) (*xadmin.SystemAlertBotListResp, error)
	SaveAlertBot(ctx context.Context, req *xadmin.SystemAlertBotSaveReq) (*xadmin.SystemActionResp, error)
	DeleteAlertBot(ctx context.Context, req *xadmin.SystemAlertBotActionReq) (*xadmin.SystemActionResp, error)
	ListAlertScenes(ctx context.Context, req *xadmin.SystemAlertSceneListReq) (*xadmin.SystemAlertSceneListResp, error)
	SaveAlertScene(ctx context.Context, req *xadmin.SystemAlertSceneSaveReq) (*xadmin.SystemActionResp, error)
	DeleteAlertScene(ctx context.Context, req *xadmin.SystemAlertSceneActionReq) (*xadmin.SystemActionResp, error)
	TestSendAlertScene(ctx context.Context, req *xadmin.SystemAlertSceneTestSendReq) (*xadmin.SystemActionResp, error)
	ListAlertTemplates(ctx context.Context, req *xadmin.SystemAlertTemplateListReq) (*xadmin.SystemAlertTemplateListResp, error)
	SaveAlertTemplate(ctx context.Context, req *xadmin.SystemAlertTemplateSaveReq) (*xadmin.SystemActionResp, error)
	DeleteAlertTemplate(ctx context.Context, req *xadmin.SystemAlertTemplateActionReq) (*xadmin.SystemActionResp, error)
}

type service struct {
	repo *systemrepo.Repo
}

func NewService() Service {
	return &service{repo: systemrepo.NewRepo()}
}

func NewServiceWithRepo(repo *systemrepo.Repo) Service {
	return &service{repo: repo}
}

func LoadIPBlacklistStore(ctx context.Context) error {
	return (&service{repo: systemrepo.NewRepo()}).syncIPBlacklistStore(ctx)
}

func (s *service) ListAuditLogs(ctx context.Context, req *xadmin.SystemAuditLogsReq) (*xadmin.SystemAuditLogsResp, error) {
	page := req.GetPage()
	if page == nil {
		page = &commpb.PageArgs{Pn: 1, Ps: 10}
	}
	if page.GetPn() <= 0 {
		page.Pn = 1
	}
	if page.GetPs() <= 0 {
		page.Ps = 10
	}
	if err := systemrepo.ValidatedTimeRange(req.GetCreatedAt()); err != nil {
		return nil, err
	}
	rows, total, err := s.repo.ListAuditLogs(ctx, page, systemrepo.NormalizeSortArgs(req.GetSort()), buildAuditFilters(req))
	if err != nil {
		return nil, err
	}
	items := make([]*xadmin.SystemAuditLogItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapAuditItem(row))
	}
	return &xadmin.SystemAuditLogsResp{
		Items: items,
		Total: total,
		Page:  &commpb.PageArgs{Pn: page.GetPn(), Ps: page.GetPs()},
	}, nil
}

func (s *service) GetAuditLog(ctx context.Context, req *xadmin.SystemAuditLogDetailReq) (*xadmin.SystemAuditLogDetailResp, error) {
	row, err := s.repo.GetAuditLogByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return mapAuditDetail(row), nil
}

func (s *service) ExportAuditLogs(ctx context.Context, req *xadmin.SystemAuditLogsReq) ([]byte, error) {
	if err := systemrepo.ValidatedTimeRange(req.GetCreatedAt()); err != nil {
		return nil, err
	}
	rows, err := s.repo.ExportAuditLogs(ctx, buildAuditFilters(req), systemrepo.NormalizeSortArgs(req.GetSort()))
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(nil)
	writer := csv.NewWriter(buf)
	_ = writer.Write([]string{"日志ID", "用户UID", "操作人", "审计动作", "结果", "请求IP", "耗时", "TraceID", "RequestID", "详情", "用户代理", "操作时间"})
	for _, row := range rows {
		_ = writer.Write([]string{
			fmt.Sprintf("%d", row.ID),
			fmt.Sprintf("%d", row.UID),
			strings.TrimSpace(row.Actor),
			auditActionCN(row.Action),
			auditResultCN(row.Result),
			strings.TrimSpace(row.SourceIP),
			strings.TrimSpace(row.Duration),
			strings.TrimSpace(row.TraceID),
			strings.TrimSpace(row.RequestID),
			strings.TrimSpace(row.Detail),
			strings.TrimSpace(row.UserAgent),
			timefmt.DateTime(row.CreatedAt),
		})
	}
	writer.Flush()
	return buf.Bytes(), nil
}

func (s *service) ApplyAuditLogRetention(ctx context.Context, req *xadmin.SystemAuditLogRetentionReq) (*xadmin.SystemAuditLogRetentionResp, error) {
	retainDays := req.GetRetainDays()
	cutoff := time.Now().AddDate(0, 0, -int(retainDays))
	expiredCount, validCount, err := s.repo.CountAuditLogsByCutoff(ctx, cutoff)
	if err != nil {
		return nil, err
	}
	if !req.GetConfirm() || expiredCount <= 0 {
		return &xadmin.SystemAuditLogRetentionResp{
			Success:      req.GetConfirm() && expiredCount == 0,
			RetainDays:   retainDays,
			ExpiredCount: expiredCount,
			ValidCount:   validCount,
			CutoffAt:     timefmt.DateTime(cutoff),
		}, nil
	}
	deletedCount, err := s.repo.DeleteAuditLogsBefore(ctx, cutoff)
	if err != nil {
		return nil, err
	}
	remainingCount, err := s.repo.CountAuditLogsSince(ctx, cutoff)
	if err != nil {
		return nil, err
	}
	return &xadmin.SystemAuditLogRetentionResp{
		Success:      true,
		RetainDays:   retainDays,
		ExpiredCount: deletedCount,
		ValidCount:   remainingCount,
		CutoffAt:     timefmt.DateTime(cutoff),
	}, nil
}

func buildAuditFilters(req *xadmin.SystemAuditLogsReq) systemrepo.AuditFilters {
	return systemrepo.AuditFilters{
		Actor:     strings.TrimSpace(req.GetActor()),
		Action:    strings.TrimSpace(req.GetAction()),
		Result:    auditResultToDB(req.GetResult()),
		TraceID:   strings.TrimSpace(req.GetTraceId()),
		RequestID: strings.TrimSpace(req.GetRequestId()),
		SourceIP:  strings.TrimSpace(req.GetSourceIp()),
		Keyword:   strings.TrimSpace(req.GetKeyword()),
		CreatedAt: req.GetCreatedAt(),
	}
}

func mapAuditItem(row systemrepo.AuditRow) *xadmin.SystemAuditLogItem {
	return &xadmin.SystemAuditLogItem{
		Id:        row.ID,
		Uid:       row.UID,
		Actor:     strings.TrimSpace(row.Actor),
		Action:    strings.TrimSpace(row.Action),
		Result:    strings.TrimSpace(row.Result),
		TraceId:   strings.TrimSpace(row.TraceID),
		RequestId: strings.TrimSpace(row.RequestID),
		SourceIp:  strings.TrimSpace(row.SourceIP),
		Duration:  strings.TrimSpace(row.Duration),
		UserAgent: strings.TrimSpace(row.UserAgent),
		Detail:    strings.TrimSpace(row.Detail),
		CreatedAt: timefmt.DateTime(row.CreatedAt),
	}
}

func mapAuditDetail(row *systemrepo.AuditRow) *xadmin.SystemAuditLogDetailResp {
	return &xadmin.SystemAuditLogDetailResp{
		Id:        row.ID,
		Uid:       row.UID,
		Actor:     strings.TrimSpace(row.Actor),
		Action:    strings.TrimSpace(row.Action),
		Result:    strings.TrimSpace(row.Result),
		TraceId:   strings.TrimSpace(row.TraceID),
		RequestId: strings.TrimSpace(row.RequestID),
		SourceIp:  strings.TrimSpace(row.SourceIP),
		Duration:  strings.TrimSpace(row.Duration),
		UserAgent: strings.TrimSpace(row.UserAgent),
		Detail:    strings.TrimSpace(row.Detail),
		CreatedAt: timefmt.DateTime(row.CreatedAt),
	}
}

func auditResultToDB(result xadmin.SystemAuditResultFilter) string {
	switch result {
	case xadmin.SystemAuditResultFilter_SYSTEM_AUDIT_RESULT_FILTER_SUCCESS:
		return "success"
	case xadmin.SystemAuditResultFilter_SYSTEM_AUDIT_RESULT_FILTER_FAILED:
		return "failed"
	default:
		return ""
	}
}

func auditResultCN(result string) string {
	if strings.EqualFold(strings.TrimSpace(result), "failed") {
		return "失败"
	}
	return "成功"
}

func auditActionCN(action string) string {
	switch strings.TrimSpace(action) {
	case "login_success":
		return "登录成功"
	case "login_failed":
		return "登录失败"
	case "logout":
		return "退出登录"
	case "logout_others":
		return "注销其他会话"
	case "force_logout":
		return "强制下线"
	case "deactivate":
		return "注销账号"
	case "update_personal_settings":
		return "更新个人设置"
	case "auth_failed":
		return "鉴权失败"
	default:
		if strings.TrimSpace(action) == "" {
			return "未知"
		}
		return action
	}
}

func RequirePage(req *xadmin.SystemAuditLogsReq) error {
	if req == nil {
		return xerr.NewBiz(xerr.CodeBadRequest, "sys.params_required")
	}
	return nil
}

func (s *service) ListWarmTips(ctx context.Context, req *xadmin.SystemWarmTipsReq) (*xadmin.SystemWarmTipsResp, error) {
	page := req.GetPage()
	if page == nil {
		page = &commpb.PageArgs{Pn: 1, Ps: 10}
	}
	if page.GetPn() <= 0 {
		page.Pn = 1
	}
	if page.GetPs() <= 0 {
		page.Ps = 10
	}
	rows, total, err := s.repo.ListWarmTips(ctx, page, req.GetSort(), buildWarmTipFilters(req))
	if err != nil {
		return nil, err
	}
	items := make([]*xadmin.SystemWarmTipItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapWarmTipItem(row))
	}
	return &xadmin.SystemWarmTipsResp{
		Items: items,
		Total: total,
		Page:  &commpb.PageArgs{Pn: page.GetPn(), Ps: page.GetPs()},
	}, nil
}

func (s *service) CreateWarmTip(ctx context.Context, req *xadmin.SystemCreateWarmTipReq) (*xadmin.SystemActionResp, error) {
	tipType, err := normalizeWarmTipType(req.GetTipType())
	if err != nil {
		return nil, err
	}
	if err := validateWarmTipText(req.GetContentZh(), req.GetContentEn()); err != nil {
		return nil, err
	}
	status := req.GetStatus()
	if status != 0 && status != 1 {
		status = 1
	}
	if err := s.repo.CreateWarmTip(ctx, &systemrepo.WarmTipRow{
		TipType:   tipType,
		ContentZh: strings.TrimSpace(req.GetContentZh()),
		ContentEn: strings.TrimSpace(req.GetContentEn()),
		Sort:      req.GetSort(),
		Status:    status,
	}); err != nil {
		return nil, err
	}
	return &xadmin.SystemActionResp{Success: true, Action: "create_warm_tip"}, nil
}

func (s *service) UpdateWarmTip(ctx context.Context, req *xadmin.SystemUpdateWarmTipReq) (*xadmin.SystemActionResp, error) {
	tipType, err := normalizeWarmTipType(req.GetTipType())
	if err != nil {
		return nil, err
	}
	if err := validateWarmTipText(req.GetContentZh(), req.GetContentEn()); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateWarmTip(ctx, &systemrepo.WarmTipRow{
		ID:        req.GetId(),
		TipType:   tipType,
		ContentZh: strings.TrimSpace(req.GetContentZh()),
		ContentEn: strings.TrimSpace(req.GetContentEn()),
		Sort:      req.GetSort(),
		Status:    req.GetStatus(),
	}); err != nil {
		return nil, err
	}
	return &xadmin.SystemActionResp{Success: true, Action: "update_warm_tip"}, nil
}

func (s *service) DeleteWarmTip(ctx context.Context, req *xadmin.SystemWarmTipActionReq) (*xadmin.SystemActionResp, error) {
	if err := s.repo.DeleteWarmTip(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &xadmin.SystemActionResp{Success: true, Action: "delete_warm_tip"}, nil
}

func mapWarmTipItem(row systemrepo.WarmTipRow) *xadmin.SystemWarmTipItem {
	return &xadmin.SystemWarmTipItem{
		Id:        row.ID,
		TipType:   strings.TrimSpace(row.TipType),
		ContentZh: strings.TrimSpace(row.ContentZh),
		ContentEn: strings.TrimSpace(row.ContentEn),
		Sort:      row.Sort,
		Status:    row.Status,
		UpdatedAt: strings.TrimSpace(row.UpdatedAt),
	}
}

func (s *service) ListIPBlacklist(ctx context.Context, req *xadmin.SystemIPBlacklistReq) (*xadmin.SystemIPBlacklistResp, error) {
	page := req.GetPage()
	if page == nil {
		page = &commpb.PageArgs{Pn: 1, Ps: 10}
	}
	if page.GetPn() <= 0 {
		page.Pn = 1
	}
	if page.GetPs() <= 0 {
		page.Ps = 10
	}
	rows, total, err := s.repo.ListIPBlacklist(ctx, page, req.GetSort(), buildIPBlacklistFilters(req))
	if err != nil {
		return nil, err
	}
	items := make([]*xadmin.SystemIPBlacklistItem, 0, len(rows))
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	hitDeltas := ipblacklist.DefaultStore().HitDeltas(ids)
	for _, row := range rows {
		status := resolveIPBlacklistStatus(row.Status, row.EndAt)
		items = append(items, &xadmin.SystemIPBlacklistItem{
			Id:        row.ID,
			Ip:        strings.TrimSpace(row.IP),
			BanType:   strings.TrimSpace(row.BanType),
			StartAt:   strings.TrimSpace(row.StartAt),
			EndAt:     strings.TrimSpace(row.EndAt),
			Reason:    strings.TrimSpace(row.Reason),
			Creator:   strings.TrimSpace(row.Creator),
			Status:    status,
			HitCount:  row.HitCount + hitDeltas[row.ID],
			UpdatedAt: strings.TrimSpace(row.UpdatedAt),
		})
	}
	return &xadmin.SystemIPBlacklistResp{
		Items: items,
		Total: total,
		Page:  &commpb.PageArgs{Pn: page.GetPn(), Ps: page.GetPs()},
	}, nil
}

func (s *service) ListIPBlacklistCreators(ctx context.Context) (*xadmin.SystemIPBlacklistCreatorsResp, error) {
	creators, err := s.repo.ListIPBlacklistCreators(ctx)
	if err != nil {
		return nil, err
	}
	return &xadmin.SystemIPBlacklistCreatorsResp{Creators: creators}, nil
}

func resolveIPBlacklistStatus(rawStatus string, rawEndAt string) string {
	if strings.TrimSpace(rawStatus) != "active" {
		return "manual_inactive"
	}
	endAt := parseServiceDateTime(rawEndAt)
	if !endAt.IsZero() && endAt.Before(time.Now()) {
		return "expired"
	}
	return "active"
}

func parseServiceDateTime(raw string) time.Time {
	input := strings.TrimSpace(raw)
	if input == "" {
		return time.Time{}
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if value, err := time.ParseInLocation(layout, input, time.Local); err == nil {
			return value
		}
	}
	return time.Time{}
}

func (s *service) CreateIPBlacklist(ctx context.Context, req *xadmin.SystemCreateIPBlacklistReq) (*xadmin.SystemActionResp, error) {
	ip := strings.TrimSpace(req.GetIp())
	if !isValidBlacklistIP(ip) {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "sys.ip_invalid")
	}
	row := &systemrepo.IPBlacklistRow{
		IP:         ip,
		BanType:    mapIPBanTypeToText(req.GetBanType()),
		StartAt:    strings.TrimSpace(req.GetStartAt()),
		EndAt:      strings.TrimSpace(req.GetEndAt()),
		Reason:     strings.TrimSpace(req.GetReason()),
		Creator:    strings.TrimSpace(req.GetCreator()),
		Status:     "active",
		HitCount:   0,
		LastAction: "manual_ban",
	}
	if row.StartAt == "" {
		row.StartAt = time.Now().Format(time.DateTime)
	}
	if strings.TrimSpace(row.BanType) == "" {
		row.BanType = "temp"
	}
	if row.EndAt == "" {
		row.EndAt = time.Now().Add(24 * time.Hour).Format(time.DateTime)
	}
	if row.Creator == "" {
		row.Creator = "system"
	}
	if err := s.repo.CreateIPBlacklist(ctx, row); err != nil {
		return nil, err
	}
	s.addIPBlacklistStoreEntry(row)
	return &xadmin.SystemActionResp{Success: true, Action: "create_ip_blacklist"}, nil
}

func (s *service) UpdateIPBlacklist(ctx context.Context, req *xadmin.SystemUpdateIPBlacklistReq) (*xadmin.SystemActionResp, error) {
	if err := s.repo.UpdateIPBlacklist(ctx, req.GetId(), mapIPBanTypeToText(req.GetBanType()), strings.TrimSpace(req.GetEndAt()), strings.TrimSpace(req.GetReason())); err != nil {
		return nil, err
	}
	row, err := s.repo.GetIPBlacklistEntry(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	s.syncIPBlacklistStoreEntry(row)
	return &xadmin.SystemActionResp{Success: true, Action: "update_ip_blacklist"}, nil
}

func (s *service) UnblockIPBlacklist(ctx context.Context, req *xadmin.SystemIPBlacklistActionReq) (*xadmin.SystemActionResp, error) {
	if err := s.repo.UnblockIPBlacklist(ctx, req.GetId()); err != nil {
		return nil, err
	}
	row, err := s.repo.GetIPBlacklistEntry(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	s.deleteIPBlacklistStoreEntriesByIDs(row.ID)
	return &xadmin.SystemActionResp{Success: true, Action: "unblock_ip_blacklist"}, nil
}

func (s *service) DeleteIPBlacklist(ctx context.Context, req *xadmin.SystemIPBlacklistActionReq) (*xadmin.SystemActionResp, error) {
	if err := s.repo.DeleteIPBlacklist(ctx, req.GetId()); err != nil {
		return nil, err
	}
	s.deleteIPBlacklistStoreEntriesByIDs(req.GetId())
	return &xadmin.SystemActionResp{Success: true, Action: "delete_ip_blacklist"}, nil
}

func (s *service) BatchUnblockIPBlacklist(ctx context.Context, req *xadmin.SystemIPBlacklistBatchUnblockReq) (*xadmin.SystemActionResp, error) {
	if err := s.repo.BatchUnblockIPBlacklist(ctx, req.GetIds()); err != nil {
		return nil, err
	}
	s.deleteIPBlacklistStoreEntriesByIDs(req.GetIds()...)
	return &xadmin.SystemActionResp{Success: true, Action: "batch_unblock_ip_blacklist"}, nil
}

func (s *service) ImportIPBlacklist(ctx context.Context, req *xadmin.SystemIPBlacklistImportReq, creator string) (*xadmin.SystemActionResp, error) {
	ips, err := normalizeBlacklistIPs(req.GetIps())
	if err != nil {
		return nil, err
	}
	if err := s.repo.ImportIPBlacklist(ctx, ips, mapIPBanTypeToText(req.GetBanType()), req.GetDurationHours(), strings.TrimSpace(req.GetEndAt()), creator); err != nil {
		return nil, err
	}
	if err := s.addIPBlacklistStoreEntriesByIPs(ctx, ips); err != nil {
		return nil, err
	}
	return &xadmin.SystemActionResp{Success: true, Action: "import_ip_blacklist"}, nil
}

func (s *service) addIPBlacklistStoreEntry(row *systemrepo.IPBlacklistRow) {
	if row == nil {
		return
	}
	s.syncIPBlacklistStoreEntry(row)
}

func (s *service) syncIPBlacklistStoreEntry(row *systemrepo.IPBlacklistRow) {
	if row == nil {
		return
	}
	if strings.TrimSpace(row.Status) != "active" {
		s.deleteIPBlacklistStoreEntriesByIDs(row.ID)
		return
	}
	ipblacklist.DefaultStore().Add(ipblacklist.Entry{
		ID:      row.ID,
		IP:      strings.TrimSpace(row.IP),
		StartAt: parseServiceDateTime(row.StartAt),
		EndAt:   parseServiceDateTime(row.EndAt),
	})
}

func (s *service) addIPBlacklistStoreEntriesByIPs(ctx context.Context, ips []string) error {
	rows, err := s.repo.ListActiveIPBlacklistEntriesByIPs(ctx, ips)
	if err != nil {
		return err
	}
	entries := make([]ipblacklist.Entry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, ipblacklist.Entry{
			ID:      row.ID,
			IP:      strings.TrimSpace(row.IP),
			StartAt: parseServiceDateTime(row.StartAt),
			EndAt:   parseServiceDateTime(row.EndAt),
		})
	}
	ipblacklist.DefaultStore().AddMany(entries)
	return nil
}

func (s *service) deleteIPBlacklistStoreEntriesByIDs(ids ...int64) {
	ipblacklist.DefaultStore().DeleteByID(ids...)
}

func (s *service) syncIPBlacklistStore(ctx context.Context) error {
	rows, err := s.repo.ListActiveIPBlacklistEntries(ctx)
	if err != nil {
		return err
	}
	entries := make([]ipblacklist.Entry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, ipblacklist.Entry{
			ID:      row.ID,
			IP:      strings.TrimSpace(row.IP),
			StartAt: parseServiceDateTime(row.StartAt),
			EndAt:   parseServiceDateTime(row.EndAt),
		})
	}
	ipblacklist.DefaultStore().Replace(entries)
	return nil
}

func normalizeBlacklistIPs(raw []string) ([]string, error) {
	ips := make([]string, 0, len(raw))
	for _, item := range raw {
		ip := strings.TrimSpace(item)
		if ip == "" {
			continue
		}
		if !isValidBlacklistIP(ip) {
			return nil, xerr.NewBiz(xerr.CodeBadRequest, "sys.ip_invalid")
		}
		ips = append(ips, ip)
	}
	if len(ips) == 0 {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "sys.ip_required")
	}
	return ips, nil
}

func isValidBlacklistIP(raw string) bool {
	input := strings.TrimSpace(raw)
	if input == "" || strings.Contains(input, "/") {
		return false
	}
	_, err := netip.ParseAddr(input)
	return err == nil
}

func buildIPBlacklistFilters(req *xadmin.SystemIPBlacklistReq) systemrepo.IPBlacklistFilters {
	filters := systemrepo.IPBlacklistFilters{
		Keyword: strings.TrimSpace(req.GetKeyword()),
		Creator: strings.TrimSpace(req.GetCreator()),
	}
	switch req.GetStatus() {
	case xadmin.SystemIPBanStatusFilter_SYSTEM_IP_BAN_STATUS_FILTER_ACTIVE:
		filters.Status = "active"
	case xadmin.SystemIPBanStatusFilter_SYSTEM_IP_BAN_STATUS_FILTER_INACTIVE:
		filters.Status = "inactive"
	}
	switch req.GetBanType() {
	case xadmin.SystemIPBanType_SYSTEM_IP_BAN_TYPE_TEMP:
		filters.BanType = "temp"
	case xadmin.SystemIPBanType_SYSTEM_IP_BAN_TYPE_PERMANENT:
		filters.BanType = "permanent"
	}
	return filters
}

func buildWarmTipFilters(req *xadmin.SystemWarmTipsReq) systemrepo.WarmTipFilters {
	filters := systemrepo.WarmTipFilters{
		Keyword: strings.TrimSpace(req.GetKeyword()),
		TipType: strings.TrimSpace(req.GetTipType()),
	}
	switch req.GetStatus() {
	case xadmin.SystemWarmTipStatusFilter_SYSTEM_WARM_TIP_STATUS_FILTER_ENABLED:
		v := int32(1)
		filters.Status = &v
	case xadmin.SystemWarmTipStatusFilter_SYSTEM_WARM_TIP_STATUS_FILTER_DISABLED:
		v := int32(0)
		filters.Status = &v
	}
	return filters
}

func validateWarmTipText(contentZh, contentEn string) error {
	if !isWarmTipWordCountValid(strings.TrimSpace(contentZh)) {
		return xerr.NewBiz(xerr.CodeBadRequest, "sys.warm_tip_zh_len")
	}
	if !isWarmTipWordCountValid(strings.TrimSpace(contentEn)) {
		return xerr.NewBiz(xerr.CodeBadRequest, "sys.warm_tip_en_len")
	}
	return nil
}

func normalizeWarmTipType(input string) (string, error) {
	tipType := strings.TrimSpace(input)
	switch tipType {
	case "rest", "positive", "quote", "line":
		return tipType, nil
	default:
		return "", xerr.NewBiz(xerr.CodeBadRequest, "sys.warm_tip_type_invalid")
	}
}

func isWarmTipWordCountValid(input string) bool {
	if input == "" {
		return false
	}
	fields := strings.Fields(input)
	if len(fields) > 1 {
		return len(fields) >= 3 && len(fields) <= 20
	}
	runes := []rune(input)
	return len(runes) >= 3 && len(runes) <= 40
}

func mapIPBanTypeToText(input xadmin.SystemIPBanType) string {
	switch input {
	case xadmin.SystemIPBanType_SYSTEM_IP_BAN_TYPE_PERMANENT:
		return "permanent"
	case xadmin.SystemIPBanType_SYSTEM_IP_BAN_TYPE_TEMP:
		return "temp"
	default:
		return ""
	}
}

func (s *service) ListAlertBots(ctx context.Context, req *xadmin.SystemAlertBotListReq) (*xadmin.SystemAlertBotListResp, error) {
	page := req.GetPage()
	if page == nil {
		page = &commpb.PageArgs{Pn: 1, Ps: 10}
	}
	if page.GetPn() <= 0 {
		page.Pn = 1
	}
	if page.GetPs() <= 0 {
		page.Ps = 10
	}
	rows, total, err := s.repo.ListAlertBots(ctx, page, req.GetSort(), req.GetKeyword(), req.GetBotType())
	if err != nil {
		return nil, err
	}
	botIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		botIDs = append(botIDs, row.ID)
	}
	sceneMap, _ := s.repo.ListSceneKeysByBotIDs(ctx, botIDs)
	items := make([]*xadmin.SystemAlertBotItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &xadmin.SystemAlertBotItem{
			Id:              row.ID,
			Name:            row.Name,
			Username:        row.Username,
			Token:           row.Token,
			BotType:         row.BotType,
			Enabled:         row.Enabled,
			LinkedSceneKeys: sceneMap[row.ID],
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}
	return &xadmin.SystemAlertBotListResp{
		Items: items,
		Total: total,
		Page:  &commpb.PageArgs{Pn: page.GetPn(), Ps: page.GetPs()},
	}, nil
}

func (s *service) SaveAlertBot(ctx context.Context, req *xadmin.SystemAlertBotSaveReq) (*xadmin.SystemActionResp, error) {
	row := &systemrepo.AlertBotRow{
		ID:       req.GetId(),
		Name:     req.GetName(),
		Username: req.GetUsername(),
		Token:    req.GetToken(),
		BotType:  req.GetBotType(),
		Enabled:  req.GetEnabled(),
	}
	if err := s.repo.SaveAlertBot(ctx, row); err != nil {
		return nil, err
	}
	action := "create_alert_bot"
	if req.GetId() > 0 {
		action = "update_alert_bot"
	}
	return &xadmin.SystemActionResp{Success: true, Action: action}, nil
}

func (s *service) DeleteAlertBot(ctx context.Context, req *xadmin.SystemAlertBotActionReq) (*xadmin.SystemActionResp, error) {
	count, err := s.repo.CountScenesByBotID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, xerr.NewBiz(xerr.CodeParamError, "sys.bot_linked")
	}
	if err := s.repo.DeleteAlertBot(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &xadmin.SystemActionResp{Success: true, Action: "delete_alert_bot"}, nil
}

func (s *service) ListAlertScenes(ctx context.Context, req *xadmin.SystemAlertSceneListReq) (*xadmin.SystemAlertSceneListResp, error) {
	page := req.GetPage()
	if page == nil {
		page = &commpb.PageArgs{Pn: 1, Ps: 10}
	}
	if page.GetPn() <= 0 {
		page.Pn = 1
	}
	if page.GetPs() <= 0 {
		page.Ps = 10
	}
	rows, total, err := s.repo.ListAlertScenes(ctx, page, req.GetSort(), req.GetKeyword())
	if err != nil {
		return nil, err
	}
	items := make([]*xadmin.SystemAlertSceneItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &xadmin.SystemAlertSceneItem{
			Id:             row.ID,
			SceneKey:       row.SceneKey,
			BotId:          row.BotID,
			ParseMode:      row.ParseMode,
			GroupName:      row.GroupName,
			GroupId:        row.GroupID,
			NotifyTemplate: row.NotifyTemplate,
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
		})
	}
	return &xadmin.SystemAlertSceneListResp{
		Items: items,
		Total: total,
		Page:  &commpb.PageArgs{Pn: page.GetPn(), Ps: page.GetPs()},
	}, nil
}

func (s *service) SaveAlertScene(ctx context.Context, req *xadmin.SystemAlertSceneSaveReq) (*xadmin.SystemActionResp, error) {
	row := &systemrepo.AlertSceneRow{
		ID:             req.GetId(),
		SceneKey:       req.GetSceneKey(),
		BotID:          req.GetBotId(),
		ParseMode:      req.GetParseMode(),
		GroupName:      req.GetGroupName(),
		GroupID:        req.GetGroupId(),
		NotifyTemplate: req.GetNotifyTemplate(),
	}
	if err := s.repo.SaveAlertScene(ctx, row); err != nil {
		return nil, err
	}
	action := "create_alert_scene"
	if req.GetId() > 0 {
		action = "update_alert_scene"
	}
	return &xadmin.SystemActionResp{Success: true, Action: action}, nil
}

func (s *service) DeleteAlertScene(ctx context.Context, req *xadmin.SystemAlertSceneActionReq) (*xadmin.SystemActionResp, error) {
	if err := s.repo.DeleteAlertScene(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &xadmin.SystemActionResp{Success: true, Action: "delete_alert_scene"}, nil
}

func (s *service) TestSendAlertScene(ctx context.Context, req *xadmin.SystemAlertSceneTestSendReq) (*xadmin.SystemActionResp, error) {
	scene, err := s.repo.GetAlertScene(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if scene.BotID <= 0 {
		return nil, xerr.NewBiz(xerr.CodeParamError, "sys.scene_no_bot")
	}
	bot, err := s.repo.GetAlertBot(ctx, scene.BotID)
	if err != nil {
		return nil, xerr.NewBiz(xerr.CodeParamError, "sys.bot_not_found")
	}
	if !bot.Enabled {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "sys.bot_disabled")
	}
	msg := scene.NotifyTemplate
	for k, v := range req.GetVariables() {
		msg = strings.ReplaceAll(msg, "{"+k+"}", v)
	}
	var sendErr error
	switch bot.BotType {
	case "telegram":
		if scene.GroupID == "" {
			return nil, xerr.NewBiz(xerr.CodeParamError, "sys.tg_group_required")
		}
		tgBot := &alertbot.TelegramBot{Name: bot.Name, Username: scene.GroupID, Token: bot.Token, ParseMode: scene.ParseMode}
		sendErr = tgBot.Send(ctx, msg)
	case "feishu":
		fsBot := &alertbot.FeishuBot{Name: bot.Name, Token: bot.Token, ParseMode: scene.ParseMode}
		sendErr = fsBot.Send(ctx, msg)
	default:
		return nil, xerr.NewBiz(xerr.CodeParamError, "sys.send_unsupported")
	}
	if sendErr != nil {
		return nil, xerr.NewWithDetail(xerr.CodeInternalError, "send failed: %s", sendErr.Error())
	}
	return &xadmin.SystemActionResp{Success: true, Action: "test_send"}, nil
}

func (s *service) ListAlertTemplates(ctx context.Context, req *xadmin.SystemAlertTemplateListReq) (*xadmin.SystemAlertTemplateListResp, error) {
	page := req.GetPage()
	if page == nil {
		page = &commpb.PageArgs{Pn: 1, Ps: 10}
	}
	if page.GetPn() <= 0 {
		page.Pn = 1
	}
	if page.GetPs() <= 0 {
		page.Ps = 10
	}
	rows, total, err := s.repo.ListAlertTemplates(ctx, page, req.GetSort(), req.GetKeyword(), req.GetBotType())
	if err != nil {
		return nil, err
	}
	items := make([]*xadmin.SystemAlertTemplateItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &xadmin.SystemAlertTemplateItem{
			Id: row.ID, BotType: row.BotType, Name: row.Name, ParseMode: row.ParseMode, Content: row.Content,
			CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		})
	}
	return &xadmin.SystemAlertTemplateListResp{Items: items, Total: total, Page: &commpb.PageArgs{Pn: page.GetPn(), Ps: page.GetPs()}}, nil
}

func (s *service) SaveAlertTemplate(ctx context.Context, req *xadmin.SystemAlertTemplateSaveReq) (*xadmin.SystemActionResp, error) {
	row := &systemrepo.AlertTemplateRow{
		ID: req.GetId(), BotType: req.GetBotType(), Name: req.GetName(), ParseMode: req.GetParseMode(), Content: req.GetContent(),
	}
	if err := s.repo.SaveAlertTemplate(ctx, row); err != nil {
		return nil, err
	}
	action := "create_alert_template"
	if req.GetId() > 0 {
		action = "update_alert_template"
	}
	return &xadmin.SystemActionResp{Success: true, Action: action}, nil
}

func (s *service) DeleteAlertTemplate(ctx context.Context, req *xadmin.SystemAlertTemplateActionReq) (*xadmin.SystemActionResp, error) {
	if err := s.repo.DeleteAlertTemplate(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &xadmin.SystemActionResp{Success: true, Action: "delete_alert_template"}, nil
}
