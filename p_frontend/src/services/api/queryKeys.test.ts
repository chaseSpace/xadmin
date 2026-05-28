import { describe, expect, it } from 'vitest'
import { adminUserKeys, personalSettingsKeys, systemSettingsKeys, warmTipKeys, WARM_TIP_ENABLED_PAGE_SIZE } from './queryKeys'

describe('adminUserKeys', () => {
  it('builds stable list and detail keys', () => {
    expect(adminUserKeys.list({ page: 1, pageSize: 20 })).toEqual([
      'adminUsers',
      'list',
      { page: 1, pageSize: 20 },
    ])

    expect(adminUserKeys.detail('u_1001')).toEqual(['adminUsers', 'detail', 'u_1001'])
  })
})

describe('warmTipKeys', () => {
  it('keeps stable keys for enabled warm tip cache', () => {
    expect(WARM_TIP_ENABLED_PAGE_SIZE).toBe(200)
    expect(warmTipKeys.enabled).toEqual(['systemWarmTips', 'enabled'])
  })
})

describe('personalSettingsKeys', () => {
  it('keeps settings detail cache scoped by user id', () => {
    expect(personalSettingsKeys.detail(1001)).toEqual(['authPersonalSettings', 1001])
  })
})

describe('systemSettingsKeys', () => {
  it('keeps one shared system settings cache key', () => {
    expect(systemSettingsKeys.detail).toEqual(['system-settings'])
  })
})
