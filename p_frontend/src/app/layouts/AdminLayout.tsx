import {
  AuditOutlined,
  CloseCircleOutlined,
  CloseOutlined,
  DeploymentUnitOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  LoadingOutlined,
  FolderOpenOutlined,
  FileTextOutlined,
  MoonOutlined,
  ExclamationCircleOutlined,
  HeartOutlined,
  SafetyCertificateOutlined,
  SettingOutlined,
  SolutionOutlined,
  SunOutlined,
  ApartmentOutlined,
  UsergroupAddOutlined,
  ProfileOutlined,
  KeyOutlined,
  MinusCircleOutlined,
} from '@ant-design/icons'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate, useRouterState } from '@tanstack/react-router'
import { Alert, Avatar, Breadcrumb, Dropdown, Form, Input, Layout, Menu, Modal, Select, Space, Switch, Tabs, Tooltip, Typography, Watermark, message } from 'antd'
import type { MenuProps } from 'antd'
import type { AppTabKey, AppTabItem } from '../../store/pageTabs'
import { useEffect, useMemo, useRef, useState } from 'react'
import type { ReactNode } from 'react'
import { HomePage } from '../../pages/HomePage'
import { OrganizationStructurePage } from '../../pages/organization/OrganizationStructurePage'
import { OrganizationMembersPage } from '../../pages/organization/OrganizationMembersPage'
import { OrganizationPositionsPage } from '../../pages/organization/OrganizationPositionsPage'
import { BusinessUsersPage } from '../../pages/business/BusinessUsersPage'
import { BusinessUserPunishmentsPage } from '../../pages/business/BusinessUserPunishmentsPage'
import { ResourceFilesPage } from '../../pages/resource/ResourceFilesPage'
import { PermissionRolesPage } from '../../pages/permission/PermissionRolesPage'
import { PermissionPoliciesPage } from '../../pages/permission/PermissionPoliciesPage'
import { SystemSettingsPage } from '../../pages/system/SystemSettingsPage'
import { SystemAuditLogsPage } from '../../pages/system/SystemAuditLogsPage'
import { SystemIpBlacklistPage } from '../../pages/system/SystemIpBlacklistPage'
import { SystemWarmTipsPage } from '../../pages/system/SystemWarmTipsPage'
import { SystemAlertBotsPage } from '../../pages/system/SystemAlertBotsPage'
import { useAuthStore } from '../../store/auth'
import { useI18nStore, type SupportedLocale } from '../../store/i18n'
import { usePageTabsStore } from '../../store/pageTabs'
import { useThemeStore } from '../../store/theme'
import { useUiSettingsStore } from '../../store/uiSettings'
import { UiButton } from '../../components/ui'
import { useI18n } from '../../i18n/messages'
import { getSessions, logout as logoutApi, logoutOthers as logoutOthersApi } from '../../services/api/auth'
import { getMyProfile, getPersonalSettings, getSystemSettings, updatePersonalSettings, type CurrentUserMenuItem, type CurrentUserProfile, type PersonalSettings } from '../../services/api/account'
import { SESSIONS_POLL_INTERVAL_MS } from '../../services/api/sessionPolling'
import { getEnabledSystemWarmTips } from '../../services/api/system'
import { personalSettingsKeys, systemSettingsKeys, warmTipKeys } from '../../services/api/queryKeys'
import {
  formatWarmTipRemainingTime,
  getWarmTipRemainingMs,
  getWarmTipRotationIndex,
} from './warmTip'

const { Header, Content, Sider } = Layout

const ROUTE_MENU_FALLBACK_TITLE: Record<AppTabKey, string> = {
  '/': '概览',
  '/organization/departments': '部门管理',
  '/organization/users': '用户列表',
  '/organization/positions': '岗位管理',
  '/business/users': '用户列表',
  '/business/user-punishments': '用户惩罚',
  '/resource/files': '文件管理',
  '/permission/role-permissions': '角色权限',
  '/permission/menu-permissions': '菜单权限',
  '/system/settings': '系统设置',
  '/system/audit-logs': '操作审计',
  '/system/ip-blacklist': 'IP黑名单',
  '/system/warm-tips': '关怀提示',
  '/system/alert-bots': '告警通知',
}

const ROUTE_MENU_ICON: Partial<Record<AppTabKey | string, ReactNode>> = {
  '/': <DashboardOutlined />,
  organization: <ApartmentOutlined />,
  '/organization/departments': <DeploymentUnitOutlined />,
  '/organization/users': <UsergroupAddOutlined />,
  '/organization/positions': <SolutionOutlined />,
  business: <DatabaseOutlined />,
  '/business/users': <UsergroupAddOutlined />,
  '/business/user-punishments': <ExclamationCircleOutlined />,
  resource: <FolderOpenOutlined />,
  '/resource/files': <FileTextOutlined />,
  permission: <SafetyCertificateOutlined />,
  '/permission/role-permissions': <ProfileOutlined />,
  '/permission/menu-permissions': <KeyOutlined />,
  system: <SettingOutlined />,
  '/system/settings': <SettingOutlined />,
  '/system/audit-logs': <AuditOutlined />,
  '/system/ip-blacklist': <ExclamationCircleOutlined />,
  '/system/warm-tips': <HeartOutlined />,
  '/system/alert-bots': <ExclamationCircleOutlined />,
}

const API_MENU_ICON: Record<string, ReactNode> = {
  audit: <AuditOutlined />,
  apartment: <ApartmentOutlined />,
  dashboard: <DashboardOutlined />,
  database: <DatabaseOutlined />,
  'deployment-unit': <DeploymentUnitOutlined />,
  'exclamation-circle': <ExclamationCircleOutlined />,
  'file-text': <FileTextOutlined />,
  'folder-open': <FolderOpenOutlined />,
  key: <KeyOutlined />,
  profile: <ProfileOutlined />,
  'safety-certificate': <SafetyCertificateOutlined />,
  setting: <SettingOutlined />,
  solution: <SolutionOutlined />,
  'usergroup-add': <UsergroupAddOutlined />,
  heart: <HeartOutlined />,
}

const WARM_TIP_INTERVAL_OPTIONS = [
  { value: 10, label: '10分钟' },
  { value: 60, label: '1小时' },
  { value: 360, label: '6小时' },
  { value: 720, label: '12小时' },
  { value: 1440, label: '1天' },
]

function resolveMenuKey(pathname: string): AppTabKey {
  if (pathname.startsWith('/organization/departments')) {
    return '/organization/departments'
  }

  if (pathname.startsWith('/organization/users')) {
    return '/organization/users'
  }

  if (pathname.startsWith('/organization/positions')) {
    return '/organization/positions'
  }

  if (pathname.startsWith('/business/users')) {
    return '/business/users'
  }

  if (pathname.startsWith('/business/user-punishments')) {
    return '/business/user-punishments'
  }

  if (pathname.startsWith('/resource/files')) {
    return '/resource/files'
  }

  if (pathname.startsWith('/permission/role-permissions')) {
    return '/permission/role-permissions'
  }

  if (pathname.startsWith('/permission/menu-permissions')) {
    return '/permission/menu-permissions'
  }

  if (pathname.startsWith('/system/settings')) {
    return '/system/settings'
  }

  if (pathname.startsWith('/system/audit-logs')) {
    return '/system/audit-logs'
  }

  if (pathname.startsWith('/system/ip-blacklist')) {
    return '/system/ip-blacklist'
  }

  if (pathname.startsWith('/system/warm-tips')) {
    return '/system/warm-tips'
  }

  if (pathname.startsWith('/system/alert-bots')) {
    return '/system/alert-bots'
  }

  return '/'
}

function resolveCrumbTitle(segment: string): string {
  const map: Record<string, string> = {
    organization: '组织管理',
    departments: '部门管理',
    users: '用户列表',
    positions: '岗位管理',
    business: '业务管理',
    'user-punishments': '用户惩罚',
    resource: '资源管理',
    files: '文件管理',
    permission: '权限管理',
    'role-permissions': '角色权限',
    'menu-permissions': '菜单权限',
    system: '系统管理',
    settings: '系统设置',
    'audit-logs': '操作审计',
    'ip-blacklist': 'IP黑名单',
    'warm-tips': '关怀提示',
    new: '新增',
    edit: '编辑',
  }

  return map[segment] ?? segment
}

function normalizeDirectoryKey(routePath: string, permissionKey: string): string {
  if (routePath) return routePath
  const root = permissionKey.split('.')[0]
  return root || permissionKey
}

function buildMenuTitleMap(items: CurrentUserMenuItem[]): Map<string, string> {
  const map = new Map<string, string>([['/', '概览']])
  const walk = (nodes: CurrentUserMenuItem[]) => {
    nodes.forEach((node) => {
      const key = normalizeDirectoryKey(node.routePath, node.permissionKey)
      if (key) {
        map.set(key, node.name)
      }
      walk(node.children)
    })
  }
  walk(items)
  return map
}

function resolveMenuTitle(key: AppTabKey, titleMap?: Map<string, string>): string {
  return titleMap?.get(key) || ROUTE_MENU_FALLBACK_TITLE[key]
}

function resolveMenuIcon(item: CurrentUserMenuItem, key: string): ReactNode {
  return (item.icon ? API_MENU_ICON[item.icon] : undefined) ?? ROUTE_MENU_ICON[key]
}

function resolveGreeting(hour: number): { text: string; emoji: string } {
  if (hour >= 5 && hour < 9) return { text: '早上好', emoji: '🌅' }
  if (hour >= 9 && hour < 12) return { text: '上午好', emoji: '☕' }
  if (hour >= 12 && hour < 17) return { text: '下午好', emoji: '🌤️' }
  if (hour >= 17 && hour < 19) return { text: '傍晚了', emoji: '🌇' }
  if (hour >= 19 && hour < 23) return { text: '晚上好', emoji: '🌙' }
  return { text: '深夜了', emoji: '✨' }
}

function resolvePathTitle(pathname: string, titleMap: Map<string, string>): string | undefined {
  if (titleMap.has(pathname)) return titleMap.get(pathname)
  const segments = pathname.split('/').filter(Boolean)
  while (segments.length > 0) {
    const key = `/${segments.join('/')}`
    const title = titleMap.get(key)
    if (title) return title
    segments.pop()
  }
  return undefined
}

function buildMenuItemsFromApi(items: CurrentUserMenuItem[], t: (text: string) => string): MenuProps['items'] {
  return items
    .map((item) => {
      const key = normalizeDirectoryKey(item.routePath, item.permissionKey)
      if (!key) return null
      const children = buildMenuItemsFromApi(item.children, t)
      if (!item.routePath && (!children || children.length === 0)) return null
      return {
        key,
        icon: resolveMenuIcon(item, key),
        label: t(item.name),
        children: children && children.length > 0 ? children : undefined,
      }
    })
    .filter(Boolean) as MenuProps['items']
}

function renderPageByTabKey(key: AppTabKey) {
  switch (key) {
    case '/':
      return <HomePage />
    case '/organization/departments':
      return <OrganizationStructurePage />
    case '/organization/users':
      return <OrganizationMembersPage />
    case '/organization/positions':
      return <OrganizationPositionsPage />
    case '/business/users':
      return <BusinessUsersPage />
    case '/business/user-punishments':
      return <BusinessUserPunishmentsPage />
    case '/resource/files':
      return <ResourceFilesPage />
    case '/permission/role-permissions':
      return <PermissionRolesPage />
    case '/permission/menu-permissions':
      return <PermissionPoliciesPage />
    case '/system/settings':
      return <SystemSettingsPage />
    case '/system/audit-logs':
      return <SystemAuditLogsPage />
    case '/system/ip-blacklist':
      return <SystemIpBlacklistPage />
    case '/system/warm-tips':
      return <SystemWarmTipsPage />
    case '/system/alert-bots':
      return <SystemAlertBotsPage />
    default:
      return <HomePage />
  }
}

export function AdminLayout() {
  const { t } = useI18n()
  const pathname = useRouterState({ select: (state) => state.location.pathname })
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const logout = useAuthStore((state) => state.logout)
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const currentUser = useAuthStore((state) => state.currentUser)
  const updateCurrentUserProfile = useAuthStore((state) => state.updateCurrentUserProfile)
  const locale = useI18nStore((state) => state.locale)
  const setLocale = useI18nStore((state) => state.setLocale)
  const themeMode = useThemeStore((state) => state.mode)
  const toggleThemeMode = useThemeStore((state) => state.toggleMode)
  const watermarkEnabled = useUiSettingsStore((state) => state.globalWatermarkEnabled)
  const currentUserName = useUiSettingsStore((state) => state.currentUserName)
  const watermarkFontSize = useUiSettingsStore((state) => state.globalWatermarkFontSize)
  const setCurrentUserBackgroundImage = useUiSettingsStore((state) => state.setCurrentUserBackgroundImage)
  const setGlobalBackgroundApplyEnabled = useUiSettingsStore((state) => state.setGlobalBackgroundApplyEnabled)
  const resetCurrentUserBackgroundImage = useUiSettingsStore((state) => state.resetCurrentUserBackgroundImage)
  const tabs = usePageTabsStore((state) => state.tabs)
  const touchTab = usePageTabsStore((state) => state.touchTab)
  const closeTab = usePageTabsStore((state) => state.closeTab)
  const closeOtherTabs = usePageTabsStore((state) => state.closeOtherTabs)
  const resetTabs = usePageTabsStore((state) => state.resetTabs)
  const [messageApi, contextHolder] = message.useMessage()
  const [modalApi, modalContextHolder] = Modal.useModal()
  const [personalSettingsForm] = Form.useForm<{
    limitSingleLogin: boolean
    backgroundImageUrl: string
    globalBackgroundApplyEnabled: boolean
    avatar: string
    warmTipIntervalMinutes: number
  }>()
  const [personalSettingsOpen, setPersonalSettingsOpen] = useState(false)
  const [personalSettingsLoading, setPersonalSettingsLoading] = useState(false)
  const [personalSettingsSaving, setPersonalSettingsSaving] = useState(false)
  const [localeSaving, setLocaleSaving] = useState(false)
  const [myInfoOpen, setMyInfoOpen] = useState(false)
  const [myInfoLoading, setMyInfoLoading] = useState(false)
  const [myProfile, setMyProfile] = useState<CurrentUserProfile | null>(null)
  const [themeSwitchFxMode, setThemeSwitchFxMode] = useState<'dark' | 'light' | null>(null)
  const [greetingHour, setGreetingHour] = useState(() => new Date().getHours())
  const [remainingTick, setRemainingTick] = useState(() => Date.now())
  const [warmTipAnimationKey, setWarmTipAnimationKey] = useState(0)
  const themeSwitchApplyTimerRef = useRef<number | null>(null)
  const themeSwitchFxTimerRef = useRef<number | null>(null)
  const watchedAvatar = Form.useWatch('avatar', personalSettingsForm)
  const systemSettingsQuery = useQuery({
    queryKey: systemSettingsKeys.detail,
    queryFn: getSystemSettings,
    enabled: isAuthenticated && Boolean(currentUser?.uid),
    staleTime: 60_000,
  })

  const allowedRouteSet = useMemo(() => {
    const routes = (currentUser?.menuRoutes ?? []).filter((item) => typeof item === 'string' && item.startsWith('/'))
    return new Set(routes)
  }, [currentUser?.menuRoutes])
  const profileMenuItems = useMemo(() => currentUser?.menuItems ?? [], [currentUser?.menuItems])
  const menuLoaded = Boolean(currentUser?.menuLoaded)
  const menuLoadError = currentUser?.menuLoadError ?? ''
  const greeting = useMemo(() => resolveGreeting(greetingHour), [greetingHour])
  const greetingUserName = (currentUser?.displayName || currentUser?.username || t('未登录用户')).trim()
  const hasProfileMenuItems = profileMenuItems.length > 0
  const menuUnavailable = menuLoaded && (!hasProfileMenuItems || Boolean(menuLoadError))
  const menuTitleMap = useMemo(() => buildMenuTitleMap(profileMenuItems), [profileMenuItems])

  const selectedKey = resolveMenuKey(pathname)
  const openKeys = useMemo(() => {
    if (pathname.startsWith('/organization')) {
      return ['organization']
    }
    if (pathname.startsWith('/permission')) {
      return ['permission']
    }
    if (pathname.startsWith('/business')) {
      return ['business']
    }
    if (pathname.startsWith('/resource')) {
      return ['resource']
    }
    if (pathname.startsWith('/system')) {
      return ['system']
    }
    return []
  }, [pathname])
  const breadcrumbItems = useMemo(() => {
    const segments = pathname.split('/').filter(Boolean)

    if (segments.length === 0) {
      return [{ title: t('首页') }]
    }

    return segments.map((segment, index) => {
      const pathTitle = resolvePathTitle(`/${segments.slice(0, index + 1).join('/')}`, menuTitleMap)
      return { title: t(pathTitle || resolveCrumbTitle(segment)) }
    })
  }, [menuTitleMap, pathname, t])
  const openedTabs = useMemo<AppTabItem[]>(
    () =>
      tabs.length > 0
        ? tabs
        : [
            {
              key: selectedKey,
              title: resolveMenuTitle(selectedKey, menuTitleMap),
            },
          ],
    [menuTitleMap, selectedKey, tabs],
  )
  const visibleTabs = useMemo(
    () =>
      openedTabs.filter((item) => {
        if (item.key === '/') return true
        return allowedRouteSet.has(item.key)
      }),
    [allowedRouteSet, openedTabs],
  )
  const activeTabTitle = useMemo(() => t(resolveMenuTitle(selectedKey, menuTitleMap)), [menuTitleMap, selectedKey, t])
  const menuItems = useMemo<MenuProps['items']>(() => {
    if (!hasProfileMenuItems) {
      return []
    }
    return [
      {
        key: '/',
        icon: ROUTE_MENU_ICON['/'],
        label: t('概览'),
      },
      ...(buildMenuItemsFromApi(profileMenuItems, t) ?? []),
    ]
  }, [hasProfileMenuItems, profileMenuItems, t])

  const getFallbackTabAfterClose = (key: AppTabKey): AppTabKey => {
    const remainingTabs = visibleTabs.filter((item) => item.key !== key)
    return remainingTabs[remainingTabs.length - 1]?.key ?? '/'
  }

  const closeCurrentTab = (key: AppTabKey) => {
    if (visibleTabs.length <= 1) return
    const fallbackKey = getFallbackTabAfterClose(key)
    closeTab(key)
    if (key === selectedKey) {
      void navigate({ to: fallbackKey })
    }
  }

  const closeOtherVisibleTabs = (key: AppTabKey) => {
    closeOtherTabs(key)
    if (key !== selectedKey) {
      void navigate({ to: key })
    }
  }

  const closeAllTabs = () => {
    resetTabs()
    if (selectedKey !== '/') {
      void navigate({ to: '/' })
    }
  }

  const copyWarmTipContent = async () => {
    const text = warmTipContent.trim()
    if (!text) return
    setWarmTipAnimationKey((current) => current + 1)
    try {
      await navigator.clipboard.writeText(text)
      void messageApi.success(t('已复制'))
    } catch {
      void messageApi.error(t('复制失败'))
    }
  }

  const renderTabLabel = (item: AppTabItem) => (
    <Dropdown
      trigger={['contextMenu']}
      menu={{
        items: [
          {
            key: 'close_current',
            icon: <CloseOutlined />,
            label: t('关闭当前'),
            disabled: visibleTabs.length <= 1,
          },
          {
            key: 'close_other',
            icon: <MinusCircleOutlined />,
            label: t('关闭其他'),
            disabled: visibleTabs.length <= 1,
          },
          {
            key: 'close_all',
            icon: <CloseCircleOutlined />,
            label: t('关闭全部'),
          },
        ],
        onClick: ({ key, domEvent }) => {
          domEvent.stopPropagation()
          if (key === 'close_current') {
            closeCurrentTab(item.key)
            return
          }
          if (key === 'close_other') {
            closeOtherVisibleTabs(item.key)
            return
          }
          if (key === 'close_all') {
            closeAllTabs()
          }
        },
      }}
    >
      <span onContextMenu={(event) => event.stopPropagation()}>{t(resolveMenuTitle(item.key, menuTitleMap))}</span>
    </Dropdown>
  )

  useEffect(() => {
    touchTab({
      key: selectedKey,
      title: t(resolveMenuTitle(selectedKey, menuTitleMap)),
    })
  }, [menuTitleMap, selectedKey, t, touchTab])

  useEffect(() => {
    if (!menuLoaded || menuUnavailable) {
      return
    }
    if (selectedKey === '/') {
      return
    }
    if (!allowedRouteSet.has(selectedKey)) {
      void navigate({ to: '/' })
    }
  }, [allowedRouteSet, menuLoaded, menuUnavailable, navigate, selectedKey])

  useEffect(() => {
    const siteName = systemSettingsQuery.data?.siteName?.trim() || 'XAdmin 管理后台'
    const pageTitle = activeTabTitle.trim()
    document.title = pageTitle ? `${siteName} - ${pageTitle}` : siteName
  }, [activeTabTitle, systemSettingsQuery.data?.siteName])

  useEffect(() => {
    return () => {
      if (themeSwitchApplyTimerRef.current) {
        window.clearTimeout(themeSwitchApplyTimerRef.current)
      }
      if (themeSwitchFxTimerRef.current) {
        window.clearTimeout(themeSwitchFxTimerRef.current)
      }
    }
  }, [])

  useEffect(() => {
    const timer = window.setInterval(() => {
      const now = Date.now()
      setGreetingHour((previous) => {
        const next = new Date(now).getHours()
        return previous === next ? previous : next
      })
      setRemainingTick(now)
    }, 60_000)
    return () => window.clearInterval(timer)
  }, [])

  const doLocalLogout = async () => {
    await queryClient.cancelQueries()
    logout()
    queryClient.clear()
    resetTabs()
    await navigate({ to: '/login', replace: true })
  }

  const doLogout = async () => {
    try {
      await logoutApi()
    } catch {
      // ignore network/auth edge cases, keep local logout deterministic
    }
    await doLocalLogout()
  }

  const toggleThemeWithFx = () => {
    const nextMode: 'dark' | 'light' = themeMode === 'dark' ? 'light' : 'dark'
    setThemeSwitchFxMode(nextMode)
    if (themeSwitchApplyTimerRef.current) {
      window.clearTimeout(themeSwitchApplyTimerRef.current)
    }
    themeSwitchApplyTimerRef.current = window.setTimeout(() => {
      toggleThemeMode()
      themeSwitchApplyTimerRef.current = null
    }, 220)
    if (themeSwitchFxTimerRef.current) {
      window.clearTimeout(themeSwitchFxTimerRef.current)
    }
    themeSwitchFxTimerRef.current = window.setTimeout(() => {
      setThemeSwitchFxMode(null)
      themeSwitchFxTimerRef.current = null
    }, 900)
  }

  const sessionsQuery = useQuery({
    queryKey: ['auth-sessions', currentUser?.uid],
    queryFn: () => getSessions('active'),
    enabled: isAuthenticated,
    staleTime: SESSIONS_POLL_INTERVAL_MS,
    refetchInterval: SESSIONS_POLL_INTERVAL_MS,
    refetchOnMount: false,
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
  })

  const personalSettingsQuery = useQuery({
    queryKey: personalSettingsKeys.detail(currentUser?.uid),
    queryFn: getPersonalSettings,
    enabled: isAuthenticated && Boolean(currentUser?.uid),
    staleTime: 60_000,
    refetchOnWindowFocus: false,
  })
  const warmTipsQuery = useQuery({
    queryKey: warmTipKeys.enabled,
    queryFn: getEnabledSystemWarmTips,
    enabled: isAuthenticated,
    staleTime: 300_000,
    refetchOnWindowFocus: false,
  })
  const warmTipItems = warmTipsQuery.data ?? []
  const warmTipIntervalMinutes = personalSettingsQuery.data?.warmTipIntervalMinutes ?? 1440
  const warmTipIndex = getWarmTipRotationIndex(remainingTick, warmTipIntervalMinutes, warmTipItems.length)
  const warmTip = warmTipIndex >= 0 ? warmTipItems[warmTipIndex] : currentUser?.warmTip ?? null
  const warmTipContent = warmTip ? (locale === 'en-US' ? warmTip.contentEn : warmTip.contentZh) : ''
  const warmTipRemainingMs = getWarmTipRemainingMs(remainingTick, warmTipIntervalMinutes)
  const warmTipRemainingText = formatWarmTipRemainingTime(warmTipRemainingMs, t)
  const warmTipLocaleClass = locale === 'en-US' ? 'admin-warm-tip__content--en' : 'admin-warm-tip__content--zh'

  useEffect(() => {
    if (!isAuthenticated) {
      resetCurrentUserBackgroundImage()
    }
  }, [isAuthenticated, resetCurrentUserBackgroundImage])

  useEffect(() => {
    if (personalSettingsQuery.data) {
      setCurrentUserBackgroundImage(personalSettingsQuery.data.backgroundImageUrl)
      setGlobalBackgroundApplyEnabled(personalSettingsQuery.data.globalBackgroundApplyEnabled)
      setLocale(personalSettingsQuery.data.locale)
    }
  }, [personalSettingsQuery.data, setCurrentUserBackgroundImage, setGlobalBackgroundApplyEnabled, setLocale])

  const persistPersonalSettings = async (settings: PersonalSettings) => {
    const updated = await updatePersonalSettings(
      settings.limitSingleLogin,
      settings.backgroundImageUrl,
      settings.locale,
      settings.globalBackgroundApplyEnabled,
      settings.avatar,
      settings.warmTipIntervalMinutes,
    )
    queryClient.setQueryData(personalSettingsKeys.detail(currentUser?.uid), updated)
    setCurrentUserBackgroundImage(updated.backgroundImageUrl)
    setGlobalBackgroundApplyEnabled(updated.globalBackgroundApplyEnabled)
    setLocale(updated.locale)
    updateCurrentUserProfile({ avatar: updated.avatar })
    return updated
  }

  const handleLocaleChange = (nextLocale: SupportedLocale) => {
    setLocaleSaving(true)
    void (async () => {
      const currentSettings: PersonalSettings = personalSettingsQuery.data ?? await queryClient.fetchQuery<PersonalSettings>({
        queryKey: personalSettingsKeys.detail(currentUser?.uid),
        queryFn: getPersonalSettings,
        staleTime: 60_000,
      })
      await persistPersonalSettings({
        ...currentSettings,
        locale: nextLocale,
      })
    })()
      .catch((error) => {
        const msg = error instanceof Error ? error.message : t('语言设置保存失败，请稍后重试')
        void messageApi.error(msg)
      })
      .finally(() => {
        setLocaleSaving(false)
      })
  }

  const otherOnlineCount =
    sessionsQuery.data
      ? sessionsQuery.data.filter((item) =>
          currentUser?.sessionId
            ? item.status === 'active' && item.sessionId !== currentUser.sessionId
            : item.status === 'active',
        ).length
      : 0

  const userActions: MenuProps['items'] = [
    {
      key: 'my_info',
      label: t('我的信息'),
    },
    {
      key: 'personal_settings',
      label: t('个人设置'),
    },
    {
      key: 'logout',
      label: t('退出登录'),
    },
    {
      key: 'logout_others',
      label: (
        <Space size={8}>
          <span>{t('注销其他会话')}</span>
          <Typography.Text type="secondary">({otherOnlineCount})</Typography.Text>
        </Space>
      ),
      disabled: otherOnlineCount < 1,
    },
  ]

  const shell = (
    <Layout className="admin-shell">
      {contextHolder}
      {modalContextHolder}
      <Sider width={232} className="admin-sider" breakpoint="lg" collapsedWidth={72}>
        <div className="brand">XAdmin</div>
        <Menu
          theme={themeMode === 'dark' ? 'dark' : 'light'}
          mode="inline"
          selectedKeys={[selectedKey]}
          defaultOpenKeys={openKeys}
          items={menuItems}
          disabled={!menuLoaded || menuUnavailable}
          onClick={({ key }) => {
            if (typeof key === 'string' && key.startsWith('/')) {
              void navigate({ to: key as AppTabKey })
            }
          }}
        />
      </Sider>
      <Layout>
        <Header className="admin-header">
          <Space size={18}>
            <div className="admin-warm-tip">
              <span className="admin-warm-tip__greeting">{t(greeting.text)}，{greetingUserName} {greeting.emoji}</span>
              {warmTipContent ? (
                <span className="admin-warm-tip__viewport">
                  <Tooltip title={warmTipRemainingText}>
                    <button
                      type="button"
                      className="admin-warm-tip__button"
                      onClick={() => void copyWarmTipContent()}
                      title={t('点击复制')}
                    >
                      <span
                        key={warmTipAnimationKey}
                        className={`admin-warm-tip__content ${warmTipLocaleClass}`}
                      >
                        <span>{warmTipContent}</span>
                        <span aria-hidden="true">{warmTipContent}</span>
                      </span>
                    </button>
                  </Tooltip>
                </span>
              ) : null}
            </div>
            <Select
              value={locale}
              popupMatchSelectWidth={false}
              style={{ width: 112, height: 36 }}
              loading={personalSettingsQuery.isFetching || localeSaving}
              disabled={personalSettingsQuery.isLoading || localeSaving}
              onChange={handleLocaleChange}
              options={[
                { value: 'zh-CN', label: t('中文') },
                { value: 'en-US', label: 'English' },
              ]}
            />
            <UiButton
              icon={themeMode === 'dark' ? <SunOutlined /> : <MoonOutlined />}
              onClick={toggleThemeWithFx}
              title={themeMode === 'dark' ? t('切换到日间模式') : t('切换到夜间模式')}
            />
            <Dropdown
              menu={{
                items: userActions,
                onClick: ({ key }) => {
                  if (key === 'my_info') {
                    setMyInfoOpen(true)
                    setMyInfoLoading(true)
                    void getMyProfile()
                      .then((profile) => {
                        setMyProfile(profile)
                      })
                      .finally(() => {
                        setMyInfoLoading(false)
                      })
                    return
                  }
                  if (key === 'logout') {
                    modalApi.confirm({
                      title: t('确认退出登录？'),
                      content: t('将退出当前设备登录状态。'),
                      onOk: async () => {
                        await doLogout()
                      },
                    })
                    return
                  }
                  if (key === 'personal_settings') {
                    setPersonalSettingsOpen(true)
                    setPersonalSettingsLoading(true)
                    void queryClient.fetchQuery({
                      queryKey: personalSettingsKeys.detail(currentUser?.uid),
                      queryFn: getPersonalSettings,
                      staleTime: 60_000,
                    })
                      .then((settings) => {
                        personalSettingsForm.setFieldsValue({
                          limitSingleLogin: settings.limitSingleLogin,
                          backgroundImageUrl: settings.backgroundImageUrl,
                          globalBackgroundApplyEnabled: settings.globalBackgroundApplyEnabled,
                          avatar: settings.avatar,
                          warmTipIntervalMinutes: settings.warmTipIntervalMinutes,
                        })
                        setCurrentUserBackgroundImage(settings.backgroundImageUrl)
                        setGlobalBackgroundApplyEnabled(settings.globalBackgroundApplyEnabled)
                      })
                      .finally(() => {
                        setPersonalSettingsLoading(false)
                      })
                    return
                  }
                  if (key === 'logout_others') {
                    if (otherOnlineCount < 1) {
                      return
                    }
                    modalApi.confirm({
                      title: t('确认注销其他会话？'),
                      content: t('将使当前账号在其他 {count} 处在线会话退出登录。', { count: otherOnlineCount }),
                      onOk: async () => {
                        try {
                          await logoutOthersApi()
                          void messageApi.success(t('已注销其他会话'))
                          void sessionsQuery.refetch()
                        } catch (error) {
                          const msg = error instanceof Error ? error.message : t('注销其他会话失败，请稍后重试')
                          await messageApi.error(msg)
                          throw error
                        }
                      },
                    })
                  }
                },
              }}
              trigger={['hover']}
            >
              <Space size={8} style={{ cursor: 'pointer' }}>
                <Avatar src={currentUser?.avatar || undefined}>
                  {(currentUser?.displayName || currentUser?.username || 'U').slice(0, 1).toUpperCase()}
                </Avatar>
                <Typography.Text>{currentUser?.displayName || currentUser?.username || t('未登录用户')}</Typography.Text>
                <span aria-hidden>▾</span>
              </Space>
            </Dropdown>
          </Space>
        </Header>
        <Content className="admin-content">
          <Tabs
            className="admin-panel-tabs"
            activeKey={selectedKey}
            hideAdd
            type="editable-card"
            onChange={(key) => void navigate({ to: key as AppTabKey })}
            onEdit={(targetKey, action) => {
              if (action !== 'remove' || typeof targetKey !== 'string') return
              const key = targetKey as AppTabKey
              closeCurrentTab(key)
            }}
            items={visibleTabs.map((item) => ({
              key: item.key,
              label: renderTabLabel(item),
              closable: visibleTabs.length > 1,
            }))}
          />
          <Breadcrumb items={breadcrumbItems} className="admin-breadcrumb" />
          {!menuLoaded ? (
            <div className="admin-panel-state">
              <Alert
                type="info"
                showIcon
                icon={<LoadingOutlined />}
                message={t('正在加载菜单权限')}
              />
            </div>
          ) : menuUnavailable ? (
            <div className="admin-panel-state">
              <Alert
                type="error"
                showIcon
                message={t('菜单权限加载失败')}
                description={menuLoadError || t('当前账号没有可用菜单权限，请联系管理员配置角色权限。')}
              />
            </div>
          ) : (
            <div className="admin-panel-pages">
              {visibleTabs.map((item) => (
                <div
                  key={item.key}
                  className="admin-panel-page"
                  style={{ display: item.key === selectedKey ? 'block' : 'none' }}
                >
                  {renderPageByTabKey(item.key)}
                </div>
              ))}
            </div>
          )}
        </Content>
        <Modal
          title={t('个人设置')}
          open={personalSettingsOpen}
          width={760}
          style={{ top: 32 }}
          styles={{ body: { maxHeight: 'calc(100vh - 180px)', overflowY: 'auto', paddingRight: 8 } }}
          onCancel={() => setPersonalSettingsOpen(false)}
          onOk={() => {
            setPersonalSettingsSaving(true)
            void personalSettingsForm
              .validateFields()
              .then(async (values) => {
                const updated = await persistPersonalSettings({
                  limitSingleLogin: Boolean(values.limitSingleLogin),
                  backgroundImageUrl: String(values.backgroundImageUrl || '').trim(),
                  locale,
                  globalBackgroundApplyEnabled: Boolean(values.globalBackgroundApplyEnabled),
                  avatar: String(values.avatar || '').trim(),
                  warmTipIntervalMinutes: Number(values.warmTipIntervalMinutes || 1440),
                })
                personalSettingsForm.setFieldValue('limitSingleLogin', updated.limitSingleLogin)
                personalSettingsForm.setFieldValue('backgroundImageUrl', updated.backgroundImageUrl)
                personalSettingsForm.setFieldValue('globalBackgroundApplyEnabled', updated.globalBackgroundApplyEnabled)
                personalSettingsForm.setFieldValue('avatar', updated.avatar)
                personalSettingsForm.setFieldValue('warmTipIntervalMinutes', updated.warmTipIntervalMinutes)
                void messageApi.success(t('个人设置保存成功'))
                setPersonalSettingsOpen(false)
              })
              .finally(() => {
                setPersonalSettingsSaving(false)
              })
          }}
          confirmLoading={personalSettingsSaving}
        >
          <Form form={personalSettingsForm} layout="vertical" className="personal-settings-form">
            <div className="personal-settings-section personal-settings-section--avatar">
              <Typography.Text strong>{t('头像设置')}</Typography.Text>
              <Space size={12} align="start" style={{ width: '100%', marginTop: 12 }}>
                <Avatar size={56} src={String(watchedAvatar || currentUser?.avatar || '') || undefined}>
                  {(currentUser?.displayName || currentUser?.username || 'U').slice(0, 1).toUpperCase()}
                </Avatar>
                <Form.Item
                  label={t('头像 URL')}
                  name="avatar"
                  rules={[{ max: 500, message: t('头像 URL 最多500字') }]}
                  style={{ flex: 1, marginBottom: 0 }}
                >
                  <Input placeholder="https://example.com/avatar.png" allowClear />
                </Form.Item>
              </Space>
            </div>
            <div className="personal-settings-grid">
              <div className="personal-settings-section">
                <Typography.Text strong>{t('登录会话')}</Typography.Text>
                <Form.Item
                  label={(
                    <Space direction="vertical" size={0}>
                      <span>{t('是否仅允许单点登录')}</span>
                      <Typography.Text type="secondary">
                        {t('仅影响当前账号，打开后将强制下线此账号其他会话。')}
                      </Typography.Text>
                    </Space>
                  )}
                  name="limitSingleLogin"
                  valuePropName="checked"
                  style={{ marginTop: 12, marginBottom: 0 }}
                >
                  <Switch loading={personalSettingsLoading} />
                </Form.Item>
              </div>
              <div className="personal-settings-section">
                <Typography.Text strong>{t('关怀提示')}</Typography.Text>
                <Form.Item
                  label={t('文案切换时间')}
                  name="warmTipIntervalMinutes"
                  style={{ marginTop: 12, marginBottom: 0 }}
                  rules={[{ required: true, message: t('请选择文案切换时间') }]}
                >
                  <Select
                    loading={personalSettingsLoading}
                    options={WARM_TIP_INTERVAL_OPTIONS.map((item) => ({ value: item.value, label: t(item.label) }))}
                  />
                </Form.Item>
              </div>
            </div>
            <div className="personal-settings-section">
              <Typography.Text strong>{t('全局背景')}</Typography.Text>
              <Form.Item label={t('全局背景图 URL')} name="backgroundImageUrl" style={{ marginTop: 12 }}>
                <Input placeholder="https://example.com/bg.avif" allowClear />
              </Form.Item>
              <Form.Item
                label={t('开启全局应用')}
                name="globalBackgroundApplyEnabled"
                valuePropName="checked"
                style={{ marginBottom: 0 }}
              >
                <Switch loading={personalSettingsLoading} />
              </Form.Item>
              <Typography.Text type="secondary">
                {t('开启后，左侧菜单和右侧列表等组件都将设置一定的透明度。')}
              </Typography.Text>
              <div style={{ marginTop: 12 }}>
                <Typography.Link
                  onClick={() => {
                    personalSettingsForm.setFieldValue('backgroundImageUrl', '')
                  }}
                >
                  {t('一键清空背景图')}
                </Typography.Link>
              </div>
              <div style={{ marginTop: 8 }}>
                <Typography.Text type="secondary">{t('保存后当前账号登录态立即生效，留空则使用纯色背景。')}</Typography.Text>
              </div>
            </div>
          </Form>
        </Modal>
        <Modal
          title={t('我的信息')}
          open={myInfoOpen}
          onCancel={() => setMyInfoOpen(false)}
          footer={null}
        >
          {myInfoLoading ? (
            <Typography.Text type="secondary">{t('加载中...')}</Typography.Text>
          ) : (
            <Space direction="vertical" size={12} style={{ width: '100%' }}>
              <Space size={12}>
                <Avatar size={56} src={myProfile?.avatar || currentUser?.avatar || undefined}>
                  {(myProfile?.displayName || currentUser?.displayName || currentUser?.username || 'U')
                    .slice(0, 1)
                    .toUpperCase()}
                </Avatar>
                <Space direction="vertical" size={0}>
                  <Typography.Text strong>{myProfile?.displayName || currentUser?.displayName || '-'}</Typography.Text>
                  <Typography.Text type="secondary">@{myProfile?.username || currentUser?.username || '-'}</Typography.Text>
                </Space>
              </Space>
              <Typography.Text>UID：{myProfile?.uid ?? currentUser?.uid ?? '-'}</Typography.Text>
              <Typography.Text>{t('邮箱：')}{myProfile?.email || '-'}</Typography.Text>
              <Typography.Text>{t('手机号：')}{myProfile?.phone || '-'}</Typography.Text>
            </Space>
          )}
        </Modal>
        {themeSwitchFxMode ? (
          <div className="theme-switch-fx" aria-hidden>
            <div className={`theme-switch-fx__panel theme-switch-fx__panel--${themeSwitchFxMode}`}>
              <div className="theme-switch-fx__halo" />
              <div className="theme-switch-fx__icon">
                {themeSwitchFxMode === 'dark' ? <MoonOutlined /> : <SunOutlined />}
              </div>
              <div className="theme-switch-fx__text">{themeSwitchFxMode === 'dark' ? t('关灯中') : t('开灯中')}</div>
            </div>
          </div>
        ) : null}
      </Layout>
    </Layout>
  )

  return (
    <Watermark
      content={[currentUserName]}
      gap={[120, 120]}
      offset={[24, 24]}
      font={{
        color: watermarkEnabled
          ? themeMode === 'dark'
            ? 'rgba(232, 238, 252, 0.14)'
            : 'rgba(22, 36, 71, 0.09)'
          : 'rgba(0, 0, 0, 0)',
        fontSize: watermarkFontSize,
      }}
    >
      {shell}
    </Watermark>
  )
}
