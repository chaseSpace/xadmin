import { describe, expect, it } from 'vitest'
import type { PermissionMenuTreeNode } from '../../services/api/permission'
import { applyRoleMenuCheckChange } from './roleMenuTree'

const menuTree: PermissionMenuTreeNode[] = [
  {
    id: 1,
    parentId: 0,
    name: '系统管理',
    permissionKey: 'system.root',
    isDelegable: true,
    children: [
      {
        id: 2,
        parentId: 1,
        name: '用户管理',
        permissionKey: 'system.users.view',
        isDelegable: true,
        children: [
          {
            id: 3,
            parentId: 2,
            name: '查看用户',
            permissionKey: 'system.users.view',
            isDelegable: true,
            children: [],
          },
          {
            id: 4,
            parentId: 2,
            name: '编辑用户',
            permissionKey: 'system.users.edit',
            isDelegable: false,
            children: [],
          },
        ],
      },
    ],
  },
  { id: 5, parentId: 0, name: '首页', permissionKey: 'home.view', isDelegable: true, children: [] },
]

describe('applyRoleMenuCheckChange', () => {
  it('checks every ancestor when a child node is checked', () => {
    const result = applyRoleMenuCheckChange(menuTree, [], ['3'])

    expect(new Set(result)).toEqual(new Set(['1', '2', '3']))
  })

  it('checks descendants and ancestors when a branch node is checked', () => {
    const result = applyRoleMenuCheckChange(menuTree, [], ['2'])

    expect(new Set(result)).toEqual(new Set(['1', '2', '3', '4']))
  })

  it('unchecks every descendant when a branch node is unchecked', () => {
    const result = applyRoleMenuCheckChange(menuTree, ['1', '2', '3', '4'], ['1', '3', '4'])

    expect(result).toEqual(['1'])
  })

  it('keeps ancestors checked when one child is unchecked', () => {
    const result = applyRoleMenuCheckChange(menuTree, ['1', '2', '3', '4'], ['1', '2', '4'])

    expect(new Set(result)).toEqual(new Set(['1', '2', '4']))
  })
})
