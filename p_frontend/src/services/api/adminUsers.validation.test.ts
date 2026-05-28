import { describe, expect, it } from 'vitest'
import { createAdminUser, updateAdminUser } from './adminUsers'

describe('admin user mutations validation', () => {
  it('rejects invalid create payload before request', async () => {
    await expect(
      createAdminUser({
        name: '',
        role: 'Super Admin',
        status: 'active',
      }),
    ).rejects.toMatchObject({
      name: 'ZodError',
    })
  })

  it('rejects invalid update payload before request', async () => {
    await expect(
      updateAdminUser('u_1001', {
        name: 'Valid Name',
        role: 'invalid-role' as 'Super Admin',
        status: 'active',
      }),
    ).rejects.toMatchObject({
      name: 'ZodError',
    })
  })
})
