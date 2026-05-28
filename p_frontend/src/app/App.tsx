import { QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from '@tanstack/react-router'
import enUS from 'antd/locale/en_US'
import zhCN from 'antd/locale/zh_CN'
import { ConfigProvider, Modal, message } from 'antd'
import { useEffect, useMemo, useRef } from 'react'
import { router } from './router'
import { createUnauthorizedHandler } from './unauthorized'
import { useAuthStore } from '../store/auth'
import { useI18nStore } from '../store/i18n'
import { useThemeStore } from '../store/theme'
import { useUiSettingsStore } from '../store/uiSettings'
import { getAppTheme } from '../styles/theme'
import { registerApiErrorHandler, registerUnauthorizedHandler, resetUnauthorizedState } from '../services/api/client'
import { createAppQueryClient } from '../services/api/queryClient'
import { getMyProfile, getSystemSettings } from '../services/api/account'
import { getAccessToken } from '../services/auth/token'
import { cacheBackgroundImage, getCachedBackgroundObjectURL } from '../utils/backgroundCache'

let profileBootstrapTokenDone = ''
let profileBootstrapTokenInFlight = ''
let profileBootstrapInFlightPromise: Promise<void> | null = null
let systemSettingsBootstrapTokenDone = ''
let systemSettingsBootstrapTokenInFlight = ''
let systemSettingsBootstrapInFlightPromise: Promise<void> | null = null
export function App() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const logout = useAuthStore((state) => state.logout)
  const syncFromStorage = useAuthStore((state) => state.syncFromStorage)
  const updateCurrentUserProfile = useAuthStore((state) => state.updateCurrentUserProfile)
  const locale = useI18nStore((state) => state.locale)
  const syncLocaleFromStorage = useI18nStore((state) => state.syncFromStorage)
  const themeMode = useThemeStore((state) => state.mode)
  const syncThemeFromStorage = useThemeStore((state) => state.syncFromStorage)
  const syncUiSettingsFromStorage = useUiSettingsStore((state) => state.syncFromStorage)
  const currentUserBackgroundImage = useUiSettingsStore((state) => state.currentUserBackgroundImage)
  const globalBackgroundApplyEnabled = useUiSettingsStore((state) => state.globalBackgroundApplyEnabled)
  const setGlobalWatermarkEnabled = useUiSettingsStore((state) => state.setGlobalWatermarkEnabled)
  const setGlobalWatermarkFontSize = useUiSettingsStore((state) => state.setGlobalWatermarkFontSize)
  const setSystemTimezone = useUiSettingsStore((state) => state.setSystemTimezone)
  const setServerTimezone = useUiSettingsStore((state) => state.setServerTimezone)

  const queryClient = useMemo(() => createAppQueryClient(), [])
  const appTheme = useMemo(() => getAppTheme(themeMode), [themeMode])
  const antdLocale = useMemo(() => (locale === 'en-US' ? enUS : zhCN), [locale])
  const unauthorizedRedirectingRef = useRef(false)
  const [modalApi, modalContextHolder] = Modal.useModal()
  const [messageApi, messageContextHolder] = message.useMessage()

  useEffect(() => {
    registerUnauthorizedHandler(
      createUnauthorizedHandler({
        unauthorizedRedirectingRef,
        showConfirm: modalApi.confirm,
        logout,
        clearQueryClient: () => queryClient.clear(),
        navigateToLogin: () =>
          router.navigate({
            to: '/login',
            search: {
              redirect: router.state.location.href,
              reason: 'expired',
            },
          }),
      }),
    )
  }, [logout, modalApi, queryClient])

  useEffect(() => {
    registerApiErrorHandler((msg) => {
      void messageApi.error(msg)
    })
  }, [messageApi])

  useEffect(() => {
    if (!isAuthenticated) {
      unauthorizedRedirectingRef.current = false
      resetUnauthorizedState()
    }
  }, [isAuthenticated])

  useEffect(() => {
    const handleStorageChange = (event: StorageEvent) => {
      if (event.key === 'xadmin_access_token') {
        syncFromStorage()
      }
      if (event.key === 'xadmin_user_profile') {
        syncFromStorage()
      }
      if (event.key === 'xadmin_theme_mode') {
        syncThemeFromStorage()
      }
      if (event.key === 'xadmin_global_watermark_enabled') {
        syncUiSettingsFromStorage()
      }
      if (event.key === 'xadmin_global_watermark_font_size') {
        syncUiSettingsFromStorage()
      }
      if (event.key === 'xadmin_system_timezone') {
        syncUiSettingsFromStorage()
      }
      if (event.key === 'xadmin_system_timezone:server') {
        syncUiSettingsFromStorage()
      }
      if (event.key === 'xadmin_global_background_apply_enabled') {
        syncUiSettingsFromStorage()
      }
      if (event.key === 'xadmin_locale') {
        syncLocaleFromStorage()
      }
    }

    window.addEventListener('storage', handleStorageChange)
    return () => {
      window.removeEventListener('storage', handleStorageChange)
    }
  }, [syncFromStorage, syncLocaleFromStorage, syncThemeFromStorage, syncUiSettingsFromStorage])

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', themeMode)
  }, [themeMode])

  useEffect(() => {
    document.documentElement.setAttribute('data-bg-apply', globalBackgroundApplyEnabled ? 'on' : 'off')
  }, [globalBackgroundApplyEnabled])

  useEffect(() => {
    if (!isAuthenticated) {
      return
    }
    const currentToken = getAccessToken()
    if (!currentToken) {
      return
    }
    if (profileBootstrapTokenDone === currentToken) {
      return
    }
    if (profileBootstrapTokenInFlight === currentToken && profileBootstrapInFlightPromise) {
      return
    }

    profileBootstrapTokenInFlight = currentToken
    profileBootstrapInFlightPromise = getMyProfile()
      .then((profile) => {
        updateCurrentUserProfile({
          username: profile.username,
          displayName: profile.displayName,
          avatar: profile.avatar,
          menuRoutes: profile.menuRoutes,
          menuItems: profile.menuItems,
          warmTip: profile.warmTip,
          menuLoaded: true,
          menuLoadError: '',
        })
        profileBootstrapTokenDone = currentToken
      })
      .catch((error) => {
        const message = error instanceof Error ? error.message : '菜单加载失败'
        updateCurrentUserProfile({
          menuRoutes: [],
          menuItems: [],
          menuLoaded: true,
          menuLoadError: message,
        })
      })
      .finally(() => {
        if (profileBootstrapTokenInFlight === currentToken) {
          profileBootstrapTokenInFlight = ''
          profileBootstrapInFlightPromise = null
        }
      })
  }, [isAuthenticated, updateCurrentUserProfile])

  useEffect(() => {
    if (!isAuthenticated) {
      return
    }
    const currentToken = getAccessToken()
    if (!currentToken) {
      return
    }
    if (systemSettingsBootstrapTokenDone === currentToken) {
      return
    }
    if (systemSettingsBootstrapTokenInFlight === currentToken && systemSettingsBootstrapInFlightPromise) {
      return
    }

    systemSettingsBootstrapTokenInFlight = currentToken
    systemSettingsBootstrapInFlightPromise = getSystemSettings()
      .then((settings) => {
        setGlobalWatermarkEnabled(settings.globalWatermarkEnabled)
        setGlobalWatermarkFontSize(settings.globalWatermarkFontSize)
        setSystemTimezone(settings.timezone)
        setServerTimezone(settings.serverTimezone)
        systemSettingsBootstrapTokenDone = currentToken
      })
      .catch(() => {
        // keep storage fallback when bootstrap request fails
      })
      .finally(() => {
        if (systemSettingsBootstrapTokenInFlight === currentToken) {
          systemSettingsBootstrapTokenInFlight = ''
          systemSettingsBootstrapInFlightPromise = null
        }
      })
  }, [isAuthenticated, setGlobalWatermarkEnabled, setGlobalWatermarkFontSize, setServerTimezone, setSystemTimezone])

  useEffect(() => {
    let disposed = false
    let objectURLToRevoke = ''

    const applyBackground = (imageURL: string) => {
      document.documentElement.style.setProperty('--xadmin-global-bg-image', `url("${imageURL}")`)
    }

    if (!currentUserBackgroundImage) {
      document.documentElement.style.removeProperty('--xadmin-global-bg-image')
      return () => {
        if (objectURLToRevoke) {
          URL.revokeObjectURL(objectURLToRevoke)
        }
      }
    }

    const cachedObjectURL = getCachedBackgroundObjectURL(currentUserBackgroundImage)
    if (cachedObjectURL) {
      objectURLToRevoke = cachedObjectURL
      applyBackground(cachedObjectURL)
    } else {
      applyBackground(currentUserBackgroundImage)
    }

    void cacheBackgroundImage(currentUserBackgroundImage)
      .then(() => {
        if (disposed) return
        const refreshed = getCachedBackgroundObjectURL(currentUserBackgroundImage)
        if (!refreshed) return
        if (objectURLToRevoke) {
          URL.revokeObjectURL(objectURLToRevoke)
        }
        objectURLToRevoke = refreshed
        applyBackground(refreshed)
      })
      .catch(() => {
        // keep direct URL fallback when image cannot be cached due to CORS/quota
      })

    return () => {
      disposed = true
      if (objectURLToRevoke) {
        URL.revokeObjectURL(objectURLToRevoke)
      }
    }
  }, [currentUserBackgroundImage])

  return (
    <ConfigProvider theme={appTheme} locale={antdLocale}>
      {modalContextHolder}
      {messageContextHolder}
      <QueryClientProvider client={queryClient}>
        <RouterProvider
          router={router}
          context={{
            auth: {
              isAuthenticated,
            },
          }}
        />
      </QueryClientProvider>
    </ConfigProvider>
  )
}
