import { describe, expect, it } from 'vitest'
import { formatDateTime } from './timezone'

describe('formatDateTime', () => {
  it('falls back from Local to the default system timezone', () => {
    expect(formatDateTime('2026-05-18 12:00:00', 'Asia/Shanghai', 'Local')).toBe('2026-05-18 12:00:00')
  })

  it('converts between source and display timezones', () => {
    expect(formatDateTime('2026-05-18 12:00:00', 'America/New_York', 'Asia/Shanghai')).toBe(
      '2026-05-18 00:00:00',
    )
  })
})
