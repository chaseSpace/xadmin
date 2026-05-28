import { describe, expect, it, vi } from 'vitest'
import { createUnauthorizedHandler } from './unauthorized'

describe('createUnauthorizedHandler', () => {
  it('opens confirm dialog once and redirects only after confirming', async () => {
    const unauthorizedRedirectingRef = { current: false }
    const showConfirm = vi.fn<
      (config: {
        title: string
        content: string
        okText: string
        cancelText: string
        onOk: () => void | Promise<void>
        onCancel: () => void
      }) => void
    >()
    const logout = vi.fn()
    const clearQueryClient = vi.fn()
    const navigateToLogin = vi.fn<() => Promise<void>>().mockResolvedValue()

    const handler = createUnauthorizedHandler({
      unauthorizedRedirectingRef,
      showConfirm,
      logout,
      clearQueryClient,
      navigateToLogin,
    })

    handler()
    handler()

    expect(showConfirm).toHaveBeenCalledTimes(1)
    expect(logout).not.toHaveBeenCalled()
    expect(clearQueryClient).not.toHaveBeenCalled()
    expect(navigateToLogin).not.toHaveBeenCalled()

    const config = showConfirm.mock.calls[0]?.[0]
    if (!config) {
      throw new Error('confirm config should exist')
    }

    await config.onOk()

    expect(logout).toHaveBeenCalledTimes(1)
    expect(clearQueryClient).toHaveBeenCalledTimes(1)
    expect(navigateToLogin).toHaveBeenCalledTimes(1)
    expect(unauthorizedRedirectingRef.current).toBe(false)
  })
})

