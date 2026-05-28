export type AdminUserListParams = {
  page: number
  pageSize: number
}

export const adminUserKeys = {
  all: ['adminUsers'] as const,
  list: (params: AdminUserListParams) => ['adminUsers', 'list', params] as const,
  detail: (id: string) => ['adminUsers', 'detail', id] as const,
}

export const WARM_TIP_ENABLED_PAGE_SIZE = 200

export const warmTipKeys = {
  enabled: ['systemWarmTips', 'enabled'] as const,
}

export const systemSettingsKeys = {
  detail: ['system-settings'] as const,
}

export const personalSettingsKeys = {
  detail: (uid?: number) => ['authPersonalSettings', uid] as const,
}
