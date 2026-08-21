import { describe, expect, it } from 'vitest'
import { buildRoleLabels } from './RoleSummary'

describe('buildRoleLabels', () => {
  it('keeps role order, removes duplicates and falls back to role ids', () => {
    expect(buildRoleLabels(['管理员', '', '管理员', '审计员'], [1, 2, 3, 4])).toEqual([
      '管理员',
      '角色#2',
      '审计员',
    ])
  })
})
