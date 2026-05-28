import { create } from 'zustand'

export type ThemeMode = 'light' | 'dark'

const THEME_STORAGE_KEY = 'xadmin_theme_mode'

function getInitialThemeMode(): ThemeMode {
  if (typeof window === 'undefined') return 'light'
  const saved = window.localStorage.getItem(THEME_STORAGE_KEY)
  return saved === 'dark' ? 'dark' : 'light'
}

type ThemeState = {
  mode: ThemeMode
  setMode: (mode: ThemeMode) => void
  toggleMode: () => void
  syncFromStorage: () => void
}

export const useThemeStore = create<ThemeState>((set, get) => ({
  mode: getInitialThemeMode(),
  setMode: (mode) => {
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(THEME_STORAGE_KEY, mode)
    }
    set({ mode })
  },
  toggleMode: () => {
    const nextMode: ThemeMode = get().mode === 'light' ? 'dark' : 'light'
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(THEME_STORAGE_KEY, nextMode)
    }
    set({ mode: nextMode })
  },
  syncFromStorage: () => {
    set({ mode: getInitialThemeMode() })
  },
}))
