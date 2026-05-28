package middleware

import (
	"context"
	"sync"

	"monorepo/pkg/consts"
	"monorepo/pkg/db"
	"monorepo/pkg/xerr"
	"monorepo/pkg/xfiber"

	"github.com/gofiber/fiber/v2"
)

// permissionCache stores user permission_keys in memory with TTL.
var permissionCache = &permCache{
	data: make(map[int32]*permEntry),
}

type permEntry struct {
	keys map[string]struct{}
}

type permCache struct {
	mu   sync.RWMutex
	data map[int32]*permEntry
}

func (c *permCache) Get(uid int32) (map[string]struct{}, bool) {
	c.mu.RLock()
	entry, ok := c.data[uid]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return entry.keys, true
}

func (c *permCache) Set(uid int32, keys map[string]struct{}) {
	c.mu.Lock()
	c.data[uid] = &permEntry{keys: keys}
	c.mu.Unlock()
}

// InvalidateAllPermissionCache clears all cached permissions (e.g. after role menu changes).
func InvalidateAllPermissionCache() {
	permissionCache.mu.Lock()
	permissionCache.data = make(map[int32]*permEntry)
	permissionCache.mu.Unlock()
}

// loadUserPermissionKeys queries all permission_keys for a user via role chain.
func loadUserPermissionKeys(ctx context.Context, uid int32) (map[string]struct{}, error) {
	var keys []string
	err := db.GetDatabase().WithContext(ctx).Raw(`
SELECT DISTINCT m.permission_key
FROM permission_menu m
INNER JOIN permission_role_menu prm ON prm.menu_id = m.id
INNER JOIN permission_role r ON r.id = prm.role_id AND r.deleted_at = 0
INNER JOIN (
  SELECT opr.role_id
  FROM admin_user u
  INNER JOIN organization_position_role opr ON opr.position_id = u.position_id
  WHERE u.uid = ? AND u.deleted_at = 0
) ur ON ur.role_id = prm.role_id
WHERE m.deleted_at = 0 AND m.status = ? AND m.permission_key <> ''
`, uid, consts.PermissionStatusEnabled).Scan(&keys).Error
	if err != nil {
		return nil, err
	}
	set := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		set[k] = struct{}{}
	}
	return set, nil
}

// RequirePermission returns a middleware that checks if the user has ALL specified permission keys.
func RequirePermission(keys ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid := GetUID(c)
		if uid == 0 {
			return xfiber.StdResponse(c, nil, xerr.NewBiz(xerr.CodeUnauthorized, "auth.not_logged_in"))
		}

		userKeys, ok := permissionCache.Get(uid)
		if !ok {
			var err error
			userKeys, err = loadUserPermissionKeys(c.UserContext(), uid)
			if err != nil {
				return xfiber.StdResponse(c, nil, xerr.NewBiz(xerr.CodeInternalError, "auth.permission_load_failed"))
			}
			permissionCache.Set(uid, userKeys)
		}

		for _, key := range keys {
			if _, has := userKeys[key]; !has {
				return xfiber.StdResponse(c, nil, xerr.NewBiz(xerr.CodeForbidden, "auth.no_permission"))
			}
		}
		return c.Next()
	}
}
