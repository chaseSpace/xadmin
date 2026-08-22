import { apiClient } from './client'

type PersonalSettingsApiResponse = {
  code: number
  message: string
  data: {
    limit_single_login: boolean
    background_image_url: string
    locale: string
    global_background_apply_enabled: boolean
    avatar: string
    warm_tip_interval_minutes: number
  }
}

export type PersonalSettings = {
  limitSingleLogin: boolean
  backgroundImageUrl: string
  locale: string
  globalBackgroundApplyEnabled: boolean
  avatar: string
  warmTipIntervalMinutes: number
}

export async function getPersonalSettings(): Promise<PersonalSettings> {
  const response = await apiClient.get<PersonalSettingsApiResponse>('/account/me/settings')
  return {
    limitSingleLogin: Boolean(response.data.data.limit_single_login),
    backgroundImageUrl: String(response.data.data.background_image_url || ''),
    locale: String(response.data.data.locale || 'zh-CN'),
    globalBackgroundApplyEnabled: Boolean(response.data.data.global_background_apply_enabled),
    avatar: String(response.data.data.avatar || ''),
    warmTipIntervalMinutes: normalizeWarmTipInterval(response.data.data.warm_tip_interval_minutes),
  }
}

export async function updatePersonalSettings(
  limitSingleLogin: boolean,
  backgroundImageUrl: string,
  locale: string,
  globalBackgroundApplyEnabled?: boolean,
  avatar?: string,
  warmTipIntervalMinutes?: number,
): Promise<PersonalSettings> {
  const payload: Record<string, unknown> = {
    limit_single_login: limitSingleLogin,
    background_image_url: backgroundImageUrl,
    locale,
    avatar: avatar ?? '',
    warm_tip_interval_minutes: normalizeWarmTipInterval(warmTipIntervalMinutes),
  }
  if (typeof globalBackgroundApplyEnabled === 'boolean') {
    payload.global_background_apply_enabled = globalBackgroundApplyEnabled
  }
  const response = await apiClient.post<PersonalSettingsApiResponse>(
    '/account/me/settings',
    payload,
  )
  return {
    limitSingleLogin: Boolean(response.data.data.limit_single_login),
    backgroundImageUrl: String(response.data.data.background_image_url || ''),
    locale: String(response.data.data.locale || 'zh-CN'),
    globalBackgroundApplyEnabled: Boolean(response.data.data.global_background_apply_enabled),
    avatar: String(response.data.data.avatar || ''),
    warmTipIntervalMinutes: normalizeWarmTipInterval(response.data.data.warm_tip_interval_minutes),
  }
}

function normalizeWarmTipInterval(value: unknown): number {
  const minutes = Number(value)
  return [10, 60, 360, 720, 1440].includes(minutes) ? minutes : 1440
}

type MeProfileApiResponse = {
  code: number
  message: string
  data: {
    uid: number
    username: string
    display_name: string
    avatar: string
    email: string
    phone: string
    menu_routes?: string[]
    menu_items?: CurrentUserMenuApiItem[]
    warm_tip?: CurrentUserWarmTipApiItem
    permission_keys?: string[]
    is_super_admin?: boolean
    department?: CurrentUserProfileRelationApiItem | null
    position?: CurrentUserProfileRelationApiItem | null
    roles?: CurrentUserProfileRelationApiItem[]
  }
}

type CurrentUserProfileRelationApiItem = {
  id: number | string
  name: string
  code: string
}

type CurrentUserWarmTipApiItem = {
  id: number
  tip_type: string
  content_zh: string
  content_en: string
}

type CurrentUserMenuApiItem = {
  id: number
  parent_id: number
  name: string
  route_path: string
  permission_key: string
  sort: number
  icon?: string
  children?: CurrentUserMenuApiItem[]
}

export type CurrentUserMenuItem = {
  id: number
  parentId: number
  name: string
  routePath: string
  permissionKey: string
  sort: number
  icon: string
  children: CurrentUserMenuItem[]
}

export type CurrentUserProfile = {
  uid: number
  username: string
  displayName: string
  avatar: string
  email: string
  phone: string
  menuRoutes: string[]
  menuItems: CurrentUserMenuItem[]
  warmTip: CurrentUserWarmTip | null
  permissionKeys: string[]
  isSuperAdmin: boolean
  department: CurrentUserProfileRelation | null
  position: CurrentUserProfileRelation | null
  roles: CurrentUserProfileRelation[]
}

export type CurrentUserProfileRelation = {
  id: number
  name: string
  code: string
}

export type CurrentUserWarmTip = {
  id: number
  tipType: string
  contentZh: string
  contentEn: string
}

function mapCurrentUserWarmTip(
  item: CurrentUserWarmTipApiItem | undefined,
): CurrentUserWarmTip | null {
  if (!item || typeof item !== 'object') {
    return null
  }
  return {
    id: Number(item.id),
    tipType: String(item.tip_type || ''),
    contentZh: String(item.content_zh || ''),
    contentEn: String(item.content_en || ''),
  }
}

function mapCurrentUserProfileRelation(
  item: CurrentUserProfileRelationApiItem | null | undefined,
): CurrentUserProfileRelation | null {
  if (!item || typeof item !== 'object') {
    return null
  }
  const id = Number(item.id)
  if (!Number.isFinite(id) || id <= 0) {
    return null
  }
  return {
    id,
    name: String(item.name || '').trim(),
    code: String(item.code || '').trim(),
  }
}

function mapCurrentUserMenuItems(
  items: CurrentUserMenuApiItem[] | undefined,
): CurrentUserMenuItem[] {
  if (!Array.isArray(items)) {
    return []
  }
  return items.map((item) => ({
    id: Number(item.id),
    parentId: Number(item.parent_id),
    name: String(item.name || ''),
    routePath: String(item.route_path || ''),
    permissionKey: String(item.permission_key || ''),
    sort: Number(item.sort || 0),
    icon: String(item.icon || ''),
    children: mapCurrentUserMenuItems(item.children),
  }))
}

export async function getMyProfile(): Promise<CurrentUserProfile> {
  const response = await apiClient.get<MeProfileApiResponse>('/account/me/profile')
  return {
    uid: response.data.data.uid,
    username: response.data.data.username,
    displayName: response.data.data.display_name,
    avatar: response.data.data.avatar,
    email: response.data.data.email,
    phone: response.data.data.phone,
    menuRoutes: Array.isArray(response.data.data.menu_routes)
      ? response.data.data.menu_routes.filter(
          (item) => typeof item === 'string' && item.startsWith('/'),
        )
      : [],
    menuItems: mapCurrentUserMenuItems(response.data.data.menu_items),
    warmTip: mapCurrentUserWarmTip(response.data.data.warm_tip),
    permissionKeys: Array.isArray(response.data.data.permission_keys)
      ? response.data.data.permission_keys.map(String).filter(Boolean)
      : [],
    isSuperAdmin: Boolean(response.data.data.is_super_admin),
    department: mapCurrentUserProfileRelation(response.data.data.department),
    position: mapCurrentUserProfileRelation(response.data.data.position),
    roles: Array.isArray(response.data.data.roles)
      ? response.data.data.roles
          .map(mapCurrentUserProfileRelation)
          .filter((item): item is CurrentUserProfileRelation => item !== null)
      : [],
  }
}

export type SystemSettings = {
  siteName: string
  locale: string
  timezone: string
  serverTimezone: string
  loginLockThreshold: number
  passwordMinLength: number
  sessionTimeout: number
  passwordPolicy: string[]
  globalWatermarkEnabled: boolean
  globalWatermarkFontSize: number
}

type SystemSettingsApiResponse = {
  code: number
  message: string
  data: {
    site_name: string
    locale: string
    timezone: string
    server_timezone: string
    login_lock_threshold: number
    password_min_length: number
    session_timeout_minutes: number
    password_policy: string[]
    global_watermark_enabled: boolean
    global_watermark_font_size: number
  }
}

export async function getSystemSettings(): Promise<SystemSettings> {
  const response = await apiClient.get<SystemSettingsApiResponse>('/account/system/settings')
  return {
    siteName: String(response.data.data.site_name || ''),
    locale: String(response.data.data.locale || 'zh-CN'),
    timezone: String(response.data.data.timezone || 'Asia/Shanghai'),
    serverTimezone: String(response.data.data.server_timezone || 'Local'),
    loginLockThreshold: Number(response.data.data.login_lock_threshold || 5),
    passwordMinLength: Number(response.data.data.password_min_length || 8),
    sessionTimeout: Number(response.data.data.session_timeout_minutes || 30),
    passwordPolicy: Array.isArray(response.data.data.password_policy)
      ? response.data.data.password_policy
      : [],
    globalWatermarkEnabled: Boolean(response.data.data.global_watermark_enabled),
    globalWatermarkFontSize: Number(response.data.data.global_watermark_font_size || 16),
  }
}

export async function updateSystemSettings(settings: SystemSettings): Promise<SystemSettings> {
  const response = await apiClient.post<SystemSettingsApiResponse>('/account/system/settings', {
    site_name: settings.siteName,
    locale: settings.locale,
    timezone: settings.timezone,
    login_lock_threshold: settings.loginLockThreshold,
    password_min_length: settings.passwordMinLength,
    session_timeout_minutes: settings.sessionTimeout,
    password_policy: settings.passwordPolicy,
    global_watermark_enabled: settings.globalWatermarkEnabled,
    global_watermark_font_size: settings.globalWatermarkFontSize,
  })
  return {
    siteName: String(response.data.data.site_name || ''),
    locale: String(response.data.data.locale || 'zh-CN'),
    timezone: String(response.data.data.timezone || 'Asia/Shanghai'),
    serverTimezone: String(response.data.data.server_timezone || 'Local'),
    loginLockThreshold: Number(response.data.data.login_lock_threshold || 5),
    passwordMinLength: Number(response.data.data.password_min_length || 8),
    sessionTimeout: Number(response.data.data.session_timeout_minutes || 30),
    passwordPolicy: Array.isArray(response.data.data.password_policy)
      ? response.data.data.password_policy
      : [],
    globalWatermarkEnabled: Boolean(response.data.data.global_watermark_enabled),
    globalWatermarkFontSize: Number(response.data.data.global_watermark_font_size || 16),
  }
}
