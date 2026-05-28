import { create } from 'zustand'

const WATERMARK_STORAGE_KEY = 'xadmin_global_watermark_enabled'
const WATERMARK_FONT_SIZE_STORAGE_KEY = 'xadmin_global_watermark_font_size'
const SYSTEM_TIMEZONE_STORAGE_KEY = 'xadmin_system_timezone'
const SERVER_TIMEZONE_STORAGE_KEY = 'xadmin_system_timezone:server'
const GLOBAL_BACKGROUND_APPLY_ENABLED_STORAGE_KEY = 'xadmin_global_background_apply_enabled'
const DEFAULT_WATERMARK_FONT_SIZE = 16
export const DEFAULT_SYSTEM_TIMEZONE = 'Asia/Shanghai'

function getInitialWatermarkEnabled(): boolean {
  if (typeof window === 'undefined') return false
  return window.localStorage.getItem(WATERMARK_STORAGE_KEY) === '1'
}

function getInitialWatermarkFontSize(): number {
  if (typeof window === 'undefined') return DEFAULT_WATERMARK_FONT_SIZE

  const value = Number(window.localStorage.getItem(WATERMARK_FONT_SIZE_STORAGE_KEY))
  if (Number.isNaN(value)) return DEFAULT_WATERMARK_FONT_SIZE
  return Math.min(32, Math.max(12, value))
}

function getInitialSystemTimezone(): string {
  if (typeof window === 'undefined') return DEFAULT_SYSTEM_TIMEZONE
  const value = window.localStorage.getItem(SYSTEM_TIMEZONE_STORAGE_KEY)?.trim()
  return value || DEFAULT_SYSTEM_TIMEZONE
}

function getInitialServerTimezone(): string {
  if (typeof window === 'undefined') return DEFAULT_SYSTEM_TIMEZONE
  const value = window.localStorage.getItem(SERVER_TIMEZONE_STORAGE_KEY)?.trim()
  return value || DEFAULT_SYSTEM_TIMEZONE
}

function getInitialGlobalBackgroundApplyEnabled(): boolean {
  if (typeof window === 'undefined') return true
  const value = window.localStorage.getItem(GLOBAL_BACKGROUND_APPLY_ENABLED_STORAGE_KEY)
  if (value === null) return true
  return value === '1'
}

type UiSettingsState = {
  currentUserName: string
  globalWatermarkEnabled: boolean
  globalWatermarkFontSize: number
  currentUserBackgroundImage: string
  globalBackgroundApplyEnabled: boolean
  systemTimezone: string
  serverTimezone: string
  warmTipLastChangedAt: number
  setGlobalWatermarkEnabled: (enabled: boolean) => void
  setGlobalWatermarkFontSize: (fontSize: number) => void
  setCurrentUserBackgroundImage: (backgroundImage: string) => void
  resetCurrentUserBackgroundImage: () => void
  setGlobalBackgroundApplyEnabled: (enabled: boolean) => void
  setSystemTimezone: (timezone: string) => void
  setServerTimezone: (timezone: string) => void
  setWarmTipLastChangedAt: (changedAt: number) => void
  syncFromStorage: () => void
}

export const useUiSettingsStore = create<UiSettingsState>((set) => ({
  currentUserName: 'admin',
  globalWatermarkEnabled: getInitialWatermarkEnabled(),
  globalWatermarkFontSize: getInitialWatermarkFontSize(),
  currentUserBackgroundImage: '',
  globalBackgroundApplyEnabled: getInitialGlobalBackgroundApplyEnabled(),
  systemTimezone: getInitialSystemTimezone(),
  serverTimezone: getInitialServerTimezone(),
  warmTipLastChangedAt: Date.now(),
  setGlobalWatermarkEnabled: (enabled) => {
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(WATERMARK_STORAGE_KEY, enabled ? '1' : '0')
    }
    set({ globalWatermarkEnabled: enabled })
  },
  setGlobalWatermarkFontSize: (fontSize) => {
    const nextFontSize = Math.min(32, Math.max(12, Math.round(fontSize)))
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(WATERMARK_FONT_SIZE_STORAGE_KEY, String(nextFontSize))
    }
    set({ globalWatermarkFontSize: nextFontSize })
  },
  setCurrentUserBackgroundImage: (backgroundImage) => {
    set({ currentUserBackgroundImage: backgroundImage.trim() })
  },
  resetCurrentUserBackgroundImage: () => {
    set({ currentUserBackgroundImage: '' })
  },
  setGlobalBackgroundApplyEnabled: (enabled) => {
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(GLOBAL_BACKGROUND_APPLY_ENABLED_STORAGE_KEY, enabled ? '1' : '0')
    }
    set({ globalBackgroundApplyEnabled: enabled })
  },
  setSystemTimezone: (timezone) => {
    const nextTimezone = timezone.trim() || DEFAULT_SYSTEM_TIMEZONE
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(SYSTEM_TIMEZONE_STORAGE_KEY, nextTimezone)
    }
    set({ systemTimezone: nextTimezone })
  },
  setServerTimezone: (timezone) => {
    const nextTimezone = timezone.trim() || DEFAULT_SYSTEM_TIMEZONE
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(SERVER_TIMEZONE_STORAGE_KEY, nextTimezone)
    }
    set({ serverTimezone: nextTimezone })
  },
  setWarmTipLastChangedAt: (changedAt) => {
    set({ warmTipLastChangedAt: changedAt > 0 ? changedAt : Date.now() })
  },
  syncFromStorage: () => {
    set({
      globalWatermarkEnabled: getInitialWatermarkEnabled(),
      globalWatermarkFontSize: getInitialWatermarkFontSize(),
      globalBackgroundApplyEnabled: getInitialGlobalBackgroundApplyEnabled(),
      systemTimezone: getInitialSystemTimezone(),
      serverTimezone: getInitialServerTimezone(),
    })
  },
}))
