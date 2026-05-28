package xi18n

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type ctxKey struct{}

// Entry holds translations for one biz error code.
type Entry struct {
	Zh string
	En string
}

var (
	mu   sync.RWMutex
	dict = map[string]Entry{} // bizCode -> Entry
)

// Register registers a biz error code with zh/en messages.
func Register(bizCode, zh, en string) {
	mu.Lock()
	dict[bizCode] = Entry{Zh: zh, En: en}
	mu.Unlock()
}

// WithLang stores the language in context.
func WithLang(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, ctxKey{}, lang)
}

// Lang returns the language from context, defaults to "zh".
func Lang(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKey{}).(string); ok && v != "" {
		return v
	}
	return "zh"
}

// Msg returns the translated message for a biz code, with optional format args.
func Msg(ctx context.Context, bizCode string, args ...any) string {
	mu.RLock()
	entry, ok := dict[bizCode]
	mu.RUnlock()
	if !ok {
		return bizCode
	}
	lang := normLang(Lang(ctx))
	tpl := entry.Zh
	if lang == "en" {
		tpl = entry.En
	}
	if len(args) > 0 {
		return fmt.Sprintf(tpl, args...)
	}
	return tpl
}

// MsgZh returns the Chinese message (for logging/fallback).
func MsgZh(bizCode string, args ...any) string {
	mu.RLock()
	entry, ok := dict[bizCode]
	mu.RUnlock()
	if !ok {
		return bizCode
	}
	if len(args) > 0 {
		return fmt.Sprintf(entry.Zh, args...)
	}
	return entry.Zh
}

func normLang(lang string) string {
	lang = strings.ToLower(lang)
	if strings.HasPrefix(lang, "en") {
		return "en"
	}
	return "zh"
}
