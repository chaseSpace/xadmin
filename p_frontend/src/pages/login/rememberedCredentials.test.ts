import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

type MockStorage = {
  getItem: (key: string) => string | null
  setItem: (key: string, value: string) => void
  removeItem: (key: string) => void
}

function createLocalStorage(): MockStorage {
  const storage = new Map<string, string>()
  return {
    getItem: (key) => storage.get(key) ?? null,
    setItem: (key, value) => {
      storage.set(key, value)
    },
    removeItem: (key) => {
      storage.delete(key)
    },
  }
}

describe('remembered login credentials', () => {
  beforeEach(() => {
    vi.resetModules()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('persists, reads, and clears an explicitly remembered login', async () => {
    const localStorage = createLocalStorage()
    vi.stubGlobal('window', { localStorage })

    const credentials = await import('./rememberedCredentials')
    credentials.saveRememberedLoginCredentials({ username: 'admin', password: '123456' })

    expect(credentials.getRememberedLoginCredentials()).toEqual({
      username: 'admin',
      password: '123456',
    })

    credentials.clearRememberedLoginCredentials()
    expect(credentials.getRememberedLoginCredentials()).toBeNull()
  })

  it('ignores malformed stored values', async () => {
    const localStorage = createLocalStorage()
    localStorage.setItem('xadmin_remembered_login_credentials', '{not-json')
    vi.stubGlobal('window', { localStorage })

    const { getRememberedLoginCredentials } = await import('./rememberedCredentials')
    expect(getRememberedLoginCredentials()).toBeNull()
  })
})
