import { create } from 'zustand'

export type SupportedLocale = 'zh-CN' | 'en-US'

export const DEFAULT_LOCALE: SupportedLocale = 'zh-CN'
export const LOCALE_STORAGE_KEY = 'xadmin_locale'

export function normalizeLocale(locale: string | null | undefined): SupportedLocale {
  return locale === 'en-US' ? 'en-US' : DEFAULT_LOCALE
}

function getInitialLocale(): SupportedLocale {
  if (typeof window === 'undefined') return DEFAULT_LOCALE
  return normalizeLocale(window.localStorage.getItem(LOCALE_STORAGE_KEY))
}

type I18nState = {
  locale: SupportedLocale
  setLocale: (locale: string) => void
  syncFromStorage: () => void
}

export const useI18nStore = create<I18nState>((set) => ({
  locale: getInitialLocale(),
  setLocale: (locale) => {
    const nextLocale = normalizeLocale(locale)
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(LOCALE_STORAGE_KEY, nextLocale)
    }
    set({ locale: nextLocale })
  },
  syncFromStorage: () => {
    set({ locale: getInitialLocale() })
  },
}))
