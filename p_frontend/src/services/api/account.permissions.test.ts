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
          department: { id: '10', name: '技术部', code: 'tech' },
          position: { id: '20', name: '平台工程师', code: 'platform_engineer' },
          roles: [
            { id: '30', name: '运维管理员', code: 'ops_admin' },
            { id: '31', name: '审计员', code: 'auditor' },
          ],
        },
      },
    })

    const profile = await getMyProfile()

    expect(profile.isSuperAdmin).toBe(true)
    expect(profile.permissionKeys).toEqual(['organization.positions.assign_roles'])
    expect(profile.username).toBe('renamed-root')
    expect(profile.department).toEqual({ id: 10, name: '技术部', code: 'tech' })
    expect(profile.position).toEqual({ id: 20, name: '平台工程师', code: 'platform_engineer' })
    expect(profile.roles.map((item) => item.name)).toEqual(['运维管理员', '审计员'])
  })

  it('uses empty organization relations when the account has no assignment', async () => {
    vi.spyOn(apiClient, 'get').mockResolvedValueOnce({
      data: {
        code: 200,
        data: {
          uid: 10002,
          username: 'unassigned',
          display_name: 'Unassigned',
          avatar: '',
          email: '',
          phone: '',
          menu_routes: [],
          menu_items: [],
          permission_keys: [],
          is_super_admin: false,
          department: null,
          position: null,
          roles: [],
        },
      },
    })

    const profile = await getMyProfile()

    expect(profile.department).toBeNull()
    expect(profile.position).toBeNull()
    expect(profile.roles).toEqual([])
  })
})
