import { describe, expect, it, vi } from 'vitest'
import { apiClient } from './client'
import { getOrganizationUsers } from './organization'

describe('getOrganizationUsers', () => {
  it('maps organization users payload to frontend model', async () => {
    const getSpy = vi.spyOn(apiClient, 'get').mockResolvedValueOnce({
      data: {
        code: 200,
        message: 'ok',
        data: {
          total: '1',
          page: {
            pn: 2,
            ps: 50,
          },
          items: [
            {
              uid: 10001,
              username: 'admin',
              display_name: '系统管理员',
              avatar: '',
              email: 'admin@example.com',
              phone: '13800000000',
              account_status: 'active',
              online_status: 'online',
              active_session_count: 2,
              last_login_ip: '127.0.0.1',
              last_login_at: '2026-04-20T12:00:00Z',
            },
          ],
        },
      },
    })

    const users = await getOrganizationUsers(2, 50, 'last_login_at', 'desc')
    expect(getSpy).toHaveBeenCalledWith('/organization/users', {
      params: {
        page_no: 2,
        page_size: 50,
        order_field: 'last_login_at',
        order_type: 'desc',
      },
    })
    expect(users.total).toBe(1)
    expect(users.pn).toBe(2)
    expect(users.ps).toBe(50)
    expect(users.items[0]).toMatchObject({
      uid: 10001,
      username: 'admin',
      displayName: '系统管理员',
      accountStatus: 'active',
      onlineStatus: 'online',
      activeSessionCount: 2,
    })
  })
})
