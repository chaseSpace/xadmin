const REMEMBERED_LOGIN_CREDENTIALS_KEY = 'xadmin_remembered_login_credentials'

export type RememberedLoginCredentials = {
  username: string
  password: string
}

function isRememberedLoginCredentials(value: unknown): value is RememberedLoginCredentials {
  if (!value || typeof value !== 'object') return false
  const credentials = value as Partial<RememberedLoginCredentials>
  return typeof credentials.username === 'string' && typeof credentials.password === 'string'
}

export function getRememberedLoginCredentials(): RememberedLoginCredentials | null {
  if (typeof window === 'undefined') return null

  const rawCredentials = window.localStorage.getItem(REMEMBERED_LOGIN_CREDENTIALS_KEY)
  if (!rawCredentials) return null

  try {
    const credentials: unknown = JSON.parse(rawCredentials)
    return isRememberedLoginCredentials(credentials) ? credentials : null
  } catch {
    return null
  }
}

export function saveRememberedLoginCredentials(credentials: RememberedLoginCredentials): void {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(REMEMBERED_LOGIN_CREDENTIALS_KEY, JSON.stringify(credentials))
}

export function clearRememberedLoginCredentials(): void {
  if (typeof window === 'undefined') return
  window.localStorage.removeItem(REMEMBERED_LOGIN_CREDENTIALS_KEY)
}
