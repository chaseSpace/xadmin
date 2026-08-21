import { describe, expect, it } from 'vitest'
import type { AppTabItem } from './pageTabs'
import { reorderTabs } from './pageTabOrder'

const tabs: AppTabItem[] = [
  { key: '/', title: '概览' },
  { key: '/organization/users', title: '用户列表' },
  { key: '/organization/positions', title: '岗位管理' },
]

describe('reorderTabs', () => {
  it('moves a trailing tab to the first position', () => {
    const result = reorderTabs(tabs, '/organization/positions', '/')

    expect(result.map((item) => item.key)).toEqual([
      '/organization/positions',
      '/',
      '/organization/users',
    ])
  })

  it('moves the first tab to the final position', () => {
    const result = reorderTabs(tabs, '/', '/organization/positions')

    expect(result.map((item) => item.key)).toEqual([
      '/organization/users',
      '/organization/positions',
      '/',
    ])
  })

  it('keeps the original state when the drop does not change order', () => {
    expect(reorderTabs(tabs, '/', '/')).toBe(tabs)
  })
})
