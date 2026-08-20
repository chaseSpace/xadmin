import type { Key } from 'react'
import type { PermissionMenuTreeNode } from '../../services/api/permission'

export function getDescendantKeys(nodes: PermissionMenuTreeNode[], parentId: string): string[] {
  const keys: string[] = []
  const walk = (list: PermissionMenuTreeNode[]) => {
    for (const node of list) {
      keys.push(String(node.id))
      walk(node.children)
    }
  }
  const find = (list: PermissionMenuTreeNode[]): PermissionMenuTreeNode | undefined => {
    for (const node of list) {
      if (String(node.id) === parentId) return node
      const found = find(node.children)
      if (found) return found
    }
    return undefined
  }
  const target = find(nodes)
  if (target) walk(target.children)
  return keys
}

export function getAncestorKeys(nodes: PermissionMenuTreeNode[], childId: string): string[] {
  const find = (list: PermissionMenuTreeNode[], ancestors: string[]): string[] | undefined => {
    for (const node of list) {
      if (String(node.id) === childId) return ancestors
      const found = find(node.children, [...ancestors, String(node.id)])
      if (found) return found
    }
    return undefined
  }

  return find(nodes, []) ?? []
}

export function applyRoleMenuCheckChange(
  nodes: PermissionMenuTreeNode[],
  previousKeys: readonly Key[],
  nextKeys: readonly Key[],
): string[] {
  const previous = previousKeys.map(String)
  const next = nextKeys.map(String)
  const result = new Set(next)

  for (const key of next.filter((item) => !previous.includes(item))) {
    result.add(key)
    for (const ancestor of getAncestorKeys(nodes, key)) result.add(ancestor)
    for (const descendant of getDescendantKeys(nodes, key)) result.add(descendant)
  }

  for (const key of previous.filter((item) => !next.includes(item))) {
    result.delete(key)
    for (const descendant of getDescendantKeys(nodes, key)) result.delete(descendant)
  }

  return [...result]
}
