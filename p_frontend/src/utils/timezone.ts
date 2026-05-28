import { DEFAULT_SYSTEM_TIMEZONE } from '../store/uiSettings'

type DateInput = Date | string | number | null | undefined
type PickerLike = {
  toDate?: () => Date
  toISOString?: () => string
  format?: (pattern: string) => string
}

export function normalizeTimezone(timezone?: string): string {
  const candidate = timezone?.trim()
  if (!candidate || candidate.toLowerCase() === 'local') {
    return DEFAULT_SYSTEM_TIMEZONE
  }
  try {
    new Intl.DateTimeFormat('sv-SE', { timeZone: candidate })
    return candidate
  } catch {
    return DEFAULT_SYSTEM_TIMEZONE
  }
}

function getServerTimezone(): string {
  if (typeof window === 'undefined') return DEFAULT_SYSTEM_TIMEZONE
  const value = window.localStorage.getItem('xadmin_system_timezone:server')?.trim()
  return value || DEFAULT_SYSTEM_TIMEZONE
}

function formatParts(date: Date, timezone?: string) {
  const formatter = new Intl.DateTimeFormat('sv-SE', {
    timeZone: normalizeTimezone(timezone),
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  })
  const parts = formatter.formatToParts(date)
  const read = (type: Intl.DateTimeFormatPartTypes) => parts.find((item) => item.type === type)?.value || '00'
  return {
    year: read('year'),
    month: read('month'),
    day: read('day'),
    hour: read('hour'),
    minute: read('minute'),
    second: read('second'),
  }
}

function timezoneOffsetMinutes(date: Date, timezone?: string): number {
  const parts = formatParts(date, timezone)
  const asUtc = Date.UTC(
    Number(parts.year),
    Number(parts.month) - 1,
    Number(parts.day),
    Number(parts.hour),
    Number(parts.minute),
    Number(parts.second),
  )
  return (asUtc - date.getTime()) / 60000
}

function parseTimezoneDateString(value: string, timezone?: string): Date | null {
  const input = value.trim()
  if (!input) return null
  const match = input.match(
    /^(\d{4})-(\d{2})-(\d{2})(?:[ T](\d{2}):(\d{2})(?::(\d{2}))?)?$/,
  )
  if (!match) return null

  const year = Number(match[1])
  const month = Number(match[2])
  const day = Number(match[3])
  const hour = Number(match[4] || '00')
  const minute = Number(match[5] || '00')
  const second = Number(match[6] || '00')
  const utcGuess = new Date(Date.UTC(year, month - 1, day, hour, minute, second))
  const offset = timezoneOffsetMinutes(utcGuess, timezone)
  const date = new Date(utcGuess.getTime() - offset * 60000)
  return Number.isNaN(date.getTime()) ? null : date
}

function toDate(value: DateInput, sourceTimezone?: string): Date | null {
  if (value instanceof Date) {
    return Number.isNaN(value.getTime()) ? null : value
  }
  if (typeof value === 'number') {
    const date = new Date(value)
    return Number.isNaN(date.getTime()) ? null : date
  }
  if (typeof value === 'string') {
    const input = value.trim()
    if (!input) return null
    if (/[zZ]$|[+-]\d{2}:\d{2}$/.test(input)) {
      const date = new Date(input)
      return Number.isNaN(date.getTime()) ? null : date
    }
    return parseTimezoneDateString(input, sourceTimezone)
  }
  return null
}

export function formatDateTime(value: DateInput, timezone?: string, sourceTimezone?: string): string {
  const date = toDate(value, sourceTimezone || getServerTimezone())
  if (!date) return '-'
  const parts = formatParts(date, timezone)
  return `${parts.year}-${parts.month}-${parts.day} ${parts.hour}:${parts.minute}:${parts.second}`
}

export function toTimezoneDateTimeString(value: PickerLike | DateInput, timezone?: string): string | undefined {
  if (!value) return undefined
  const date =
    typeof value === 'object' && value !== null && 'toDate' in value && typeof value.toDate === 'function'
      ? value.toDate()
      : toDate(value as DateInput)
  if (!date || Number.isNaN(date.getTime())) {
    return undefined
  }
  return formatDateTime(date, timezone)
}
