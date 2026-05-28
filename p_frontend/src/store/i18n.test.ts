import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

type MockStorage = {
  getItem: (key: string) => string | null
  setItem: (key: string, value: string) => void
  removeItem: (key: string) => void
}

function createLocalStorage(): MockStorage {
  const store = new Map<string, string>()
  return {
    getItem: (key) => store.get(key) ?? null,
    setItem: (key, value) => {
      store.set(key, value)
    },
    removeItem: (key) => {
      store.delete(key)
    },
  }
}

describe('useI18nStore', () => {
  beforeEach(() => {
    vi.resetModules()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('persists locale changes and syncs from storage', async () => {
    const localStorage = createLocalStorage()
    vi.stubGlobal('window', { localStorage })

    const { LOCALE_STORAGE_KEY, useI18nStore } = await import('./i18n')

    useI18nStore.getState().setLocale('en-US')

    expect(localStorage.getItem(LOCALE_STORAGE_KEY)).toBe('en-US')
    expect(useI18nStore.getState().locale).toBe('en-US')

    localStorage.setItem(LOCALE_STORAGE_KEY, 'zh-CN')
    useI18nStore.getState().syncFromStorage()

    expect(useI18nStore.getState().locale).toBe('zh-CN')
  })
})
