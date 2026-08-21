import { describe, expect, it } from 'vitest'
import type { AuthUserProfile } from '../store/auth'
import { hasPermission, permissionKeys } from './permissions'

const user: AuthUserProfile = {
  uid: 10002,
  username: 'operator',
  displayName: 'Operator',
  avatar: '',
  sessionId: 'session',
  menuRoutes: [],
  menuItems: [],
  permissionKeys: [permissionKeys.usersEditProfile],
  isSuperAdmin: false,
  warmTip: null,
  menuLoaded: true,
  menuLoadError: '',
}

describe('hasPermission', () => {
  it('uses server-provided permission keys instead of names', () => {
    expect(hasPermission(user, permissionKeys.usersEditProfile)).toBe(true)
    expect(hasPermission(user, permissionKeys.positionsAssignRoles)).toBe(false)
  })
})
