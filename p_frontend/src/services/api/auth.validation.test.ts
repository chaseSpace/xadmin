import { describe, expect, it } from 'vitest'
import { login } from './auth'

describe('auth validation', () => {
  it('rejects invalid login payload before request', async () => {
    await expect(
      login({ username: 'a', password: '123' }),
    ).rejects.toMatchObject({
      name: 'ZodError',
    })
  })
})
