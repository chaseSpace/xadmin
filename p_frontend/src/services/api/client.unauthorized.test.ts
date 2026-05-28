import { describe, expect, it, vi } from 'vitest'
import {
  notifyUnauthorized,
  registerUnauthorizedHandler,
  resetUnauthorizedStateForTest,
} from './client'

describe('unauthorized notification', () => {
  it('replays pending unauthorized event after handler registration', () => {
    resetUnauthorizedStateForTest()
    const handler = vi.fn()

    notifyUnauthorized()
    expect(handler).not.toHaveBeenCalled()

    registerUnauthorizedHandler(handler)
    expect(handler).toHaveBeenCalledTimes(1)
  })
})

