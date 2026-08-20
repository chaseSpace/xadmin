import { describe, expect, it, vi } from 'vitest'
import {
  notifyUnauthorized,
  registerApiErrorHandler,
  registerUnauthorizedHandler,
  resetUnauthorizedStateForTest,
  apiClient,
} from './client'

vi.mock('../auth/token', () => ({
  getAccessToken: () => null,
}))

describe('unauthorized notification', () => {
  it('replays pending unauthorized event after handler registration', () => {
    resetUnauthorizedStateForTest()
    const handler = vi.fn()

    notifyUnauthorized()
    expect(handler).not.toHaveBeenCalled()

    registerUnauthorizedHandler(handler)
    expect(handler).toHaveBeenCalledTimes(1)
  })

  it('does not show a global error message for a business 401 response', async () => {
    resetUnauthorizedStateForTest()
    const apiErrorHandler = vi.fn()
    const unauthorizedHandler = vi.fn()
    registerApiErrorHandler(apiErrorHandler)
    registerUnauthorizedHandler(unauthorizedHandler)

    await expect(
      apiClient.get('/test-business-401', {
        adapter: async (config) => ({
          data: { code: 401, message: '登录已过期' },
          status: 200,
          statusText: 'OK',
          headers: {},
          config,
        }),
      }),
    ).rejects.toBeDefined()

    expect(apiErrorHandler).not.toHaveBeenCalled()
    expect(unauthorizedHandler).toHaveBeenCalledTimes(1)
  })

  it('does not show a global error message for an HTTP 401 response', async () => {
    resetUnauthorizedStateForTest()
    const apiErrorHandler = vi.fn()
    const unauthorizedHandler = vi.fn()
    registerApiErrorHandler(apiErrorHandler)
    registerUnauthorizedHandler(unauthorizedHandler)

    await expect(
      apiClient.get('/test-http-401', {
        adapter: async (config) =>
          Promise.reject({
            config,
            response: {
              data: { message: 'Unauthorized' },
              status: 401,
              statusText: 'Unauthorized',
              headers: {},
              config,
            },
          }),
      }),
    ).rejects.toBeDefined()

    expect(apiErrorHandler).not.toHaveBeenCalled()
    expect(unauthorizedHandler).toHaveBeenCalledTimes(1)
  })
})
