type TranslateFn = (text: string, params?: Record<string, string | number>) => string

export const WARM_TIP_ENABLED_PAGE_SIZE = 200

export function formatWarmTipRemainingTime(remainingMs: number, t: TranslateFn): string {
  const totalMinutes = Math.max(0, Math.floor(remainingMs / 60_000))
  if (totalMinutes >= 1440) {
    return t('距离下次切换约 {count} 天', { count: Math.max(1, Math.floor(totalMinutes / 1440)) })
  }
  if (totalMinutes >= 60) {
    return t('距离下次切换约 {count} 小时', { count: Math.max(1, Math.floor(totalMinutes / 60)) })
  }
  return t('距离下次切换约 {count} 分钟', { count: Math.max(1, totalMinutes) })
}

export function normalizeWarmTipIntervalMinutes(value: number): number {
  return [10, 60, 360, 720, 1440].includes(value) ? value : 1440
}

export function getWarmTipRotationIndex(now: number, intervalMinutes: number, total: number): number {
  if (total <= 0) return -1
  const normalizedInterval = normalizeWarmTipIntervalMinutes(intervalMinutes)
  const bucket = Math.floor(Math.max(0, now) / (normalizedInterval * 60_000))
  return bucket % total
}

export function getWarmTipRemainingMs(now: number, intervalMinutes: number): number {
  const normalizedInterval = normalizeWarmTipIntervalMinutes(intervalMinutes)
  const intervalMs = normalizedInterval * 60_000
  const elapsed = Math.max(0, now) % intervalMs
  return elapsed === 0 ? intervalMs : intervalMs - elapsed
}
