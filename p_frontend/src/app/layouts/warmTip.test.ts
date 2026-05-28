import { describe, expect, it } from 'vitest'
import { translateText } from '../../i18n/messages'
import {
  formatWarmTipRemainingTime,
  getWarmTipRemainingMs,
  getWarmTipRotationIndex,
  normalizeWarmTipIntervalMinutes,
} from './warmTip'

const t = (text: string, params?: Record<string, string | number>) => translateText(text, 'zh-CN', params)

describe('formatWarmTipRemainingTime', () => {
  it('does not round ten minutes up to eleven minutes', () => {
    expect(formatWarmTipRemainingTime(10 * 60_000 + 999, t)).toBe('距离下次切换约 10 分钟')
  })

  it('keeps a visible one minute hint for sub-minute remaining time', () => {
    expect(formatWarmTipRemainingTime(30_000, t)).toBe('距离下次切换约 1 分钟')
  })

  it('uses completed hours and days instead of rounding upward', () => {
    expect(formatWarmTipRemainingTime(90 * 60_000, t)).toBe('距离下次切换约 1 小时')
    expect(formatWarmTipRemainingTime(36 * 60 * 60_000, t)).toBe('距离下次切换约 1 天')
  })

  it('normalizes unsupported rotation intervals', () => {
    expect(normalizeWarmTipIntervalMinutes(10)).toBe(10)
    expect(normalizeWarmTipIntervalMinutes(15)).toBe(1440)
  })

  it('rotates locally by time bucket without api state', () => {
    expect(getWarmTipRotationIndex(0, 10, 3)).toBe(0)
    expect(getWarmTipRotationIndex(10 * 60_000, 10, 3)).toBe(1)
    expect(getWarmTipRotationIndex(30 * 60_000, 10, 3)).toBe(0)
    expect(getWarmTipRotationIndex(30 * 60_000, 10, 0)).toBe(-1)
  })

  it('computes remaining time inside the current local bucket', () => {
    expect(getWarmTipRemainingMs(0, 10)).toBe(10 * 60_000)
    expect(getWarmTipRemainingMs(9 * 60_000 + 1_000, 10)).toBe(59_000)
    expect(getWarmTipRemainingMs(10 * 60_000, 10)).toBe(10 * 60_000)
  })
})
