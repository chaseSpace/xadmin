import { afterEach, describe, expect, it, vi } from 'vitest'
import { apiClient } from './client'
import {
  assignOrganizationUserPosition,
  createOrganizationPosition,
  getOrganizationPositions,
  getOrganizationUsers,
  updateOrganizationPosition,
  updateOrganizationPositionRoles,
  updateOrganizationUserProfile,
  updateOrganizationUserStatus,
} from './organization'

afterEach(() => vi.restoreAllMocks())

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

describe('organization privilege mutation payloads', () => {
  it('sends management rank only through position creation', async () => {
    const postSpy = vi.spyOn(apiClient, 'post').mockResolvedValueOnce({ data: { code: 200 } })

    await createOrganizationPosition({
      name: '测试主管',
      code: 'POS-TEST-MANAGER',
      departmentId: 3,
      level: 'M1',
      managementRank: 60,
      hc: 1,
      staffed: 0,
    })

    expect(postSpy).toHaveBeenCalledWith('/organization/positions', {
      name: '测试主管',
      code: 'POS-TEST-MANAGER',
      department_id: 3,
      level: 'M1',
      management_rank: 60,
      hc: 1,
      staffed: 0,
    })
  })

  it('keeps profile updates free of position and status fields', async () => {
    const putSpy = vi.spyOn(apiClient, 'put').mockResolvedValueOnce({ data: { code: 200 } })

    await updateOrganizationUserProfile(10002, {
      displayName: '测试用户',
      avatar: '',
      email: 'user@example.com',
      phone: '13800000000',
    })

    expect(putSpy).toHaveBeenCalledWith('/organization/users/10002/profile', {
      display_name: '测试用户',
      avatar: '',
      email: 'user@example.com',
      phone: '13800000000',
    })
  })

  it('uses dedicated user position and status endpoints', async () => {
    const postSpy = vi.spyOn(apiClient, 'post').mockResolvedValue({ data: { code: 200 } })

    await assignOrganizationUserPosition(10002, { departmentId: 3, positionId: 8 })
    await updateOrganizationUserStatus(10002, 0)

    expect(postSpy).toHaveBeenNthCalledWith(1, '/organization/users/10002/position', {
      department_id: 3,
      position_id: 8,
    })
    expect(postSpy).toHaveBeenNthCalledWith(2, '/organization/users/10002/status', { status: 0 })
  })

  it('separates position metadata from role assignment', async () => {
    const putSpy = vi.spyOn(apiClient, 'put').mockResolvedValueOnce({ data: { code: 200 } })
    const postSpy = vi.spyOn(apiClient, 'post').mockResolvedValueOnce({ data: { code: 200 } })

    await updateOrganizationPosition(7, {
      name: '测试岗位',
      code: 'POS-TEST',
      departmentId: 3,
      level: 'P5',
      hc: 2,
      staffed: 1,
      status: 'enabled',
    })
    await updateOrganizationPositionRoles(7, [2, 3])

    expect(putSpy).toHaveBeenCalledWith('/organization/positions/7', {
      name: '测试岗位',
      code: 'POS-TEST',
      department_id: 3,
      level: 'P5',
      hc: 2,
      staffed: 1,
      status: 1,
    })
    expect(postSpy).toHaveBeenCalledWith('/organization/positions/7/roles', { role_ids: [2, 3] })
  })
})

describe('getOrganizationPositions', () => {
  it('maps the read-only management rank', async () => {
    vi.spyOn(apiClient, 'get').mockResolvedValueOnce({
      data: {
        code: 200,
        message: 'ok',
        data: {
          total: '1',
          page: { pn: 1, ps: 10 },
          items: [
            {
              id: 7,
              name: '测试主管',
              code: 'POS-TEST-MANAGER',
              department_id: 3,
              department_name: '测试部',
              level: 'M1',
              management_rank: 60,
              hc: 1,
              staffed: 0,
              related_count: 0,
              status: 'enabled',
              updated_at: '2026-08-22T12:00:00Z',
            },
          ],
        },
      },
    })

    const result = await getOrganizationPositions()
    expect(result.items[0]?.managementRank).toBe(60)
  })

  it('requests nearest-parent position inheritance for user assignment', async () => {
    const getSpy = vi.spyOn(apiClient, 'get').mockResolvedValueOnce({
      data: {
        code: 200,
        message: 'ok',
        data: { total: '0', page: { pn: 1, ps: 200 }, items: [] },
      },
    })

    await getOrganizationPositions(1, 200, undefined, undefined, {
      departmentId: 12,
      inheritParent: true,
    })

    expect(getSpy).toHaveBeenCalledWith('/organization/positions', {
      params: {
        page_no: 1,
        page_size: 200,
        order_field: undefined,
        order_type: undefined,
        keyword: undefined,
        department_id: 12,
        inherit_parent: true,
        level: undefined,
        status: undefined,
      },
    })
  })
})
