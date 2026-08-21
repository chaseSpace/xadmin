import { afterEach, describe, expect, it, vi } from 'vitest'
import { apiClient } from './client'
import { getPermissionMenu, getPermissionRoles } from './permission'

afterEach(() => vi.restoreAllMocks())

describe('permission security metadata', () => {
  it('maps stable role identity and server capability flags', async () => {
    vi.spyOn(apiClient, 'get').mockResolvedValueOnce({
      data: {
        code: 200,
        data: {
          total: '1',
          page: { pn: 1, ps: 10 },
          items: [
            {
              id: 1,
              role_name: '已改名的根角色',
              role_type: 'system',
              role_code: 'super_admin',
              is_protected: true,
              can_manage: false,
              can_assign_menus: false,
              users: 1,
              updated_at: '',
            },
          ],
        },
      },
    })

    const page = await getPermissionRoles()

    expect(page.items[0]).toMatchObject({
      roleCode: 'super_admin',
      isProtected: true,
      canManage: false,
      canAssignMenus: false,
    })
  })

  it('maps delegable menu metadata', async () => {
    vi.spyOn(apiClient, 'get').mockResolvedValueOnce({
      data: {
        code: 200,
        data: {
          id: 44,
          parent_id: 9,
          name: '角色权限-分配菜单',
          route_path: '',
          component_path: '',
          menu_type: 'button',
          permission_key: 'permission.roles.assign_menus',
          is_delegable: false,
          sort: 114,
          status: 'enabled',
          updated_at: '',
        },
      },
    })

    const menu = await getPermissionMenu(44)

    expect(menu.isDelegable).toBe(false)
  })
})
