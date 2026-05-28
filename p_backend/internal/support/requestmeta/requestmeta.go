package requestmeta

import (
	"context"
	"time"
)

type startTimeKey struct{}
type requestIDKey struct{}

func WithStartTime(ctx context.Context, start time.Time) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if start.IsZero() {
		return ctx
	}
	return context.WithValue(ctx, startTimeKey{}, start)
}

func WithRequestID(ctx context.Context, rid string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, requestIDKey{}, rid)
}

func RequestID(ctx context.Context) string {
	v, _ := ctx.Value(requestIDKey{}).(string)
	return v
}

func DurationString(ctx context.Context) string {
	start, ok := ctx.Value(startTimeKey{}).(time.Time)
	if !ok || start.IsZero() {
		return ""
	}
	duration := time.Since(start)
	if duration < 0 {
		return ""
	}
	return duration.String()
}
