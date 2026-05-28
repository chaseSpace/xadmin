import { describe, expect, it } from 'vitest'
import { createAdminUser, getAdminUserDetail, updateAdminUser } from './adminUsers'

describe('admin users CRUD flow', () => {
  it('creates and updates a user via mutation functions', async () => {
    const created = await createAdminUser({
      name: 'Demo User',
      role: 'Operator',
      status: 'active',
    })

    expect(created.id).toMatch(/^u_/)
    expect(created.name).toBe('Demo User')

    const updated = await updateAdminUser(created.id, {
      name: 'Demo User Updated',
      role: 'Auditor',
      status: 'disabled',
    })

    expect(updated.name).toBe('Demo User Updated')
    expect(updated.role).toBe('Auditor')
    expect(updated.status).toBe('disabled')

    const detail = await getAdminUserDetail(created.id)
    expect(detail.name).toBe('Demo User Updated')
  })
})
