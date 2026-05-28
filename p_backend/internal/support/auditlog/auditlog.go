package auditlog

import (
	"context"
	"strings"

	"monorepo/internal/model"
	authrepo "monorepo/internal/repo/auth"
	"monorepo/internal/support/requestmeta"
)

type Meta struct {
	UID       int32
	Action    string
	Result    string
	TraceID   string
	SourceIP  string
	Duration  string
	UserAgent string
	Detail    string
}

func Log(ctx context.Context, meta Meta) error {
	duration := strings.TrimSpace(meta.Duration)
	if duration == "" {
		duration = requestmeta.DurationString(ctx)
	}
	return authrepo.NewRepo().CreateAudit(ctx, &model.AdminUserLoginAudit{
		UID:       meta.UID,
		Action:    strings.TrimSpace(meta.Action),
		Result:    strings.TrimSpace(meta.Result),
		TraceID:   strings.TrimSpace(meta.TraceID),
		RequestID: requestmeta.RequestID(ctx),
		SourceIP:  strings.TrimSpace(meta.SourceIP),
		Duration:  duration,
		UserAgent: strings.TrimSpace(meta.UserAgent),
		Detail:    strings.TrimSpace(meta.Detail),
	})
}
