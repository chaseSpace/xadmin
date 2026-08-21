import { afterEach, describe, expect, it, vi } from 'vitest'
import { getMyProfile } from './account'
import { apiClient } from './client'

afterEach(() => vi.restoreAllMocks())

describe('getMyProfile permission context', () => {
  it('maps server-provided permission keys and super-admin identity', async () => {
    vi.spyOn(apiClient, 'get').mockResolvedValueOnce({
      data: {
        code: 200,
        data: {
          uid: 10001,
          username: 'renamed-root',
          display_name: 'Root',
          avatar: '',
          email: '',
          phone: '',
          menu_routes: [],
          menu_items: [],
          permission_keys: ['organization.positions.assign_roles'],
          is_super_admin: true,
        },
      },
    })

    const profile = await getMyProfile()

    expect(profile.isSuperAdmin).toBe(true)
    expect(profile.permissionKeys).toEqual(['organization.positions.assign_roles'])
    expect(profile.username).toBe('renamed-root')
  })
})
