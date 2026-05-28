import axios from 'axios'
import { getAccessToken } from '../auth/token'
import { normalizeApiError } from './error'
import { useI18nStore } from '../../store/i18n'

declare module 'axios' {
  export interface AxiosRequestConfig {
    skipGlobalErrorTip?: boolean
    skipUnauthorizedHandler?: boolean
  }

  export interface InternalAxiosRequestConfig {
    skipGlobalErrorTip?: boolean
    skipUnauthorizedHandler?: boolean
  }
}

let onUnauthorized: (() => void) | null = null
let hasPendingUnauthorized = false
let unauthorizedFrozen = false
let onApiError: ((message: string) => void) | null = null

function createTraceID(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return `trace-${Date.now()}-${Math.random().toString(16).slice(2, 10)}`
}

export function registerUnauthorizedHandler(handler: () => void) {
  onUnauthorized = handler
  if (hasPendingUnauthorized) {
    hasPendingUnauthorized = false
    onUnauthorized()
  }
}

export function notifyUnauthorized() {
  if (unauthorizedFrozen) {
    return
  }
  unauthorizedFrozen = true
  if (onUnauthorized) {
    onUnauthorized()
    return
  }
  hasPendingUnauthorized = true
}

export function registerApiErrorHandler(handler: (message: string) => void) {
  onApiError = handler
}

function notifyApiError(message: string) {
  if (!message) return
  if (onApiError) {
    onApiError(message)
  }
}

export function resetUnauthorizedStateForTest() {
  onUnauthorized = null
  hasPendingUnauthorized = false
  unauthorizedFrozen = false
}

export function resetUnauthorizedState() {
  hasPendingUnauthorized = false
  unauthorizedFrozen = false
}

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/v1',
  timeout: 10_000,
})

apiClient.interceptors.request.use((config) => {
  const token = getAccessToken()

  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  if (!config.headers['X-Trace-ID']) {
    config.headers['X-Trace-ID'] = createTraceID()
  }
  config.headers['Accept-Language'] = useI18nStore.getState().locale

  return config
})

apiClient.interceptors.response.use(
  (response) => {
    const payload = response.data as { code?: unknown } | undefined
    const skipGlobalErrorTip = response.config.skipGlobalErrorTip === true
    const skipUnauthorizedHandler = response.config.skipUnauthorizedHandler === true
    if (payload && typeof payload === 'object' && 'code' in payload) {
      const code = payload.code
      if (typeof code === 'number' && code !== 200) {
        const msg =
          typeof (response.data as { message?: unknown })?.message === 'string'
            ? String((response.data as { message?: unknown }).message)
            : `请求失败（code=${code}）`
        if (!skipGlobalErrorTip && !(code === 401 && unauthorizedFrozen)) {
          notifyApiError(msg)
        }
        if (code === 401 && !skipUnauthorizedHandler) {
          notifyUnauthorized()
        }
        return Promise.reject({
          response: {
            status: code,
            data: response.data,
          },
        })
      }
    }
    return response
  },
  (error: unknown) => {
    const normalized = normalizeApiError(error)
    const maybeConfig =
      error && typeof error === 'object' && 'config' in error
        ? (error as { config?: { skipGlobalErrorTip?: boolean; skipUnauthorizedHandler?: boolean } }).config
        : undefined
    const skipGlobalErrorTip = maybeConfig?.skipGlobalErrorTip === true
    const skipUnauthorizedHandler = maybeConfig?.skipUnauthorizedHandler === true

    if (!skipGlobalErrorTip && !(normalized.status === 401 && unauthorizedFrozen)) {
      notifyApiError(normalized.message || '请求失败，请稍后重试')
    }

    if (normalized.status === 401 && !skipUnauthorizedHandler) {
      notifyUnauthorized()
    }

    return Promise.reject(normalized)
  },
)
