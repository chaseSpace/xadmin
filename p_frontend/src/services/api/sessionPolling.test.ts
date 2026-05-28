import { describe, expect, it } from 'vitest'
import { SESSIONS_POLL_INTERVAL_MS } from './sessionPolling'

describe('SESSIONS_POLL_INTERVAL_MS', () => {
  it('uses one-minute polling interval', () => {
    expect(SESSIONS_POLL_INTERVAL_MS).toBe(60_000)
  })
})

