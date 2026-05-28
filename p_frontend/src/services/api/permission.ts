import { apiClient } from './client'

export type PermissionOrderType = 'asc' | 'desc'

export type PermissionMenuType = 'directory' | 'menu' | 'button'
export type PermissionStatus = 'enabled' | 'disabled'
export type PermissionDeletedFilter = 'yes' | 'no'

export type PermissionMenu = {
  id: number
  parentId: number
  name: string
  routePath: string
  componentPath: string
  menuType: PermissionMenuType
  permissionKey: string
  sort: number
  status: PermissionStatus
  deleted: boolean
  deletedAt: string
  updatedAt: string
}

export type PermissionMenuTreeNode = {
  id: number
  parentId: number
  name: string
  children: PermissionMenuTreeNode[]
}

type PermissionMenuTreeApiNode = {
  id: number
  parent_id: number
  name: string
  children?: PermissionMenuTreeApiNode[]
}

export type PermissionMenusFilters = {
  keyword?: string
  status?: PermissionStatus
  deleted?: PermissionDeletedFilter
  menuType?: PermissionMenuType
  treeNodeId?: number
}

export type PermissionMenusPage = {
  total: number
  pn: number
  ps: number
  items: PermissionMenu[]
}

export type CreatePermissionMenuPayload = {
  parentId: number
  name: string
  routePath: string
  componentPath: string
  menuType: PermissionMenuType
  permissionKey: string
  sort: number
}

export type UpdatePermissionMenuPayload = CreatePermissionMenuPayload

type PermissionApiResponse<T> = {
  code: number
  message: string
  data: T
}

function mapMenuType(type: PermissionMenuType): number {
  if (type === 'directory') return 1
  if (type === 'menu') return 2
  return 3
}

function mapRoleType(type: PermissionRoleType): number {
  if (type === 'system') return 1
  return 2
}

function mapMenu(item: {
  id: number
  parent_id: number
  name: string
  route_path: string
  component_path: string
  menu_type: string
  permission_key: string
  sort: number
  status: string
  deleted?: boolean
  deleted_at?: string
  updated_at: string
}): PermissionMenu {
  const menuType: PermissionMenuType = item.menu_type === 'directory' ? 'directory' : item.menu_type === 'button' ? 'button' : 'menu'
  const status: PermissionStatus = item.status === 'disabled' ? 'disabled' : 'enabled'
  return {
    id: item.id,
    parentId: item.parent_id,
    name: item.name,
    routePath: item.route_path,
    componentPath: item.component_path,
    menuType,
    permissionKey: item.permission_key,
    sort: item.sort,
    status,
    deleted: Boolean(item.deleted),
    deletedAt: String(item.deleted_at || ''),
    updatedAt: item.updated_at,
  }
}

export async function getPermissionMenuTree(): Promise<PermissionMenuTreeNode[]> {
  const response = await apiClient.get<PermissionApiResponse<{ items: PermissionMenuTreeApiNode[] }>>('/permission/menus/tree')
  const walk = (items: PermissionMenuTreeApiNode[]): PermissionMenuTreeNode[] =>
    items.map((item) => ({
      id: item.id,
      parentId: item.parent_id,
      name: item.name,
      children: walk(item.children ?? []),
    }))
  return walk(response.data.data.items ?? [])
}

export async function getPermissionMenus(
  pageNo = 1,
  pageSize = 10,
  orderField?: 'id' | 'name' | 'menu_type' | 'sort' | 'status' | 'updated_at',
  orderType?: PermissionOrderType,
  filters?: PermissionMenusFilters,
): Promise<PermissionMenusPage> {
  const response = await apiClient.get<PermissionApiResponse<{
    total: number | string
    page?: { pn: number; ps: number }
    items: Array<{
      id: number
      parent_id: number
      name: string
      route_path: string
      component_path: string
      menu_type: string
      permission_key: string
      sort: number
      status: string
      deleted?: boolean
      deleted_at?: string
      updated_at: string
    }>
  }>>('/permission/menus', {
    params: {
      page_no: pageNo,
      page_size: pageSize,
      order_field: orderField,
      order_type: orderType,
      keyword: filters?.keyword,
      status: filters?.status,
      deleted: filters?.deleted,
      menu_type: filters?.menuType,
      tree_node_id: filters?.treeNodeId,
    },
  })
  return {
    total: Number(response.data.data.total || 0),
    pn: response.data.data.page?.pn ?? pageNo,
    ps: response.data.data.page?.ps ?? pageSize,
    items: (response.data.data.items ?? []).map(mapMenu),
  }
}

export async function getPermissionMenu(id: number): Promise<PermissionMenu> {
  const response = await apiClient.get<PermissionApiResponse<{
    id: number
    parent_id: number
    name: string
    route_path: string
    component_path: string
    menu_type: string
    permission_key: string
    sort: number
    status: string
    deleted?: boolean
    deleted_at?: string
    updated_at: string
  }>>(`/permission/menus/${id}`)
  return mapMenu(response.data.data)
}

export async function createPermissionMenu(payload: CreatePermissionMenuPayload): Promise<void> {
  await apiClient.post<PermissionApiResponse<{ success: boolean; action: string }>>('/permission/menus', {
    parent_id: payload.parentId,
    name: payload.name,
    route_path: payload.routePath,
    component_path: payload.componentPath,
    menu_type: mapMenuType(payload.menuType),
    permission_key: payload.permissionKey,
    sort: payload.sort,
  })
}

export async function updatePermissionMenu(id: number, payload: UpdatePermissionMenuPayload): Promise<void> {
  await apiClient.put<PermissionApiResponse<{ success: boolean; action: string }>>(`/permission/menus/${id}`, {
    parent_id: payload.parentId,
    name: payload.name,
    route_path: payload.routePath,
    component_path: payload.componentPath,
    menu_type: mapMenuType(payload.menuType),
    permission_key: payload.permissionKey,
    sort: payload.sort,
  })
}

export async function updatePermissionMenuStatus(id: number, enabled: boolean): Promise<void> {
  await apiClient.post<PermissionApiResponse<{ success: boolean; action: string }>>(`/permission/menus/${id}/status`, {
    enabled,
  })
}

export async function deletePermissionMenu(id: number): Promise<void> {
  await apiClient.delete<PermissionApiResponse<{ success: boolean; action: string }>>(`/permission/menus/${id}`)
}

export async function syncPermissionMenus(): Promise<void> {
  await apiClient.post<PermissionApiResponse<{ success: boolean; action: string }>>('/permission/menus/sync', {})
}

export type PermissionRoleType = 'system' | 'custom'

export type PermissionRole = {
  id: number
  roleName: string
  roleType: PermissionRoleType
  users: number
  updatedAt: string
}

export type PermissionRolesFilters = {
  keyword?: string
  roleType?: PermissionRoleType
}

export type PermissionRolesPage = {
  total: number
  pn: number
  ps: number
  items: PermissionRole[]
}

export type CreatePermissionRolePayload = {
  roleName: string
  roleType: PermissionRoleType
}

export type UpdatePermissionRolePayload = CreatePermissionRolePayload

function mapRole(item: {
  id: number | string
  role_name: string
  role_type: string
  users: number | string
  updated_at: string
}): PermissionRole {
  const id = Number(item.id)
  const users = Number(item.users)
  return {
    id: Number.isFinite(id) ? id : 0,
    roleName: item.role_name,
    roleType: item.role_type === 'system' ? 'system' : 'custom',
    users: Number.isFinite(users) ? users : 0,
    updatedAt: item.updated_at,
  }
}

export async function getPermissionRoles(
  pageNo = 1,
  pageSize = 10,
  orderField?: 'id' | 'role_name' | 'role_type' | 'updated_at' | 'users',
  orderType?: PermissionOrderType,
  filters?: PermissionRolesFilters,
): Promise<PermissionRolesPage> {
  const response = await apiClient.get<PermissionApiResponse<{
    total: number | string
    page?: { pn: number; ps: number }
    items: Array<{
      id: number
      role_name: string
      role_type: string
      users: number
      updated_at: string
    }>
  }>>('/permission/roles', {
    params: {
      page_no: pageNo,
      page_size: pageSize,
      order_field: orderField,
      order_type: orderType,
      keyword: filters?.keyword,
      role_type: filters?.roleType,
    },
  })
  return {
    total: Number(response.data.data.total || 0),
    pn: response.data.data.page?.pn ?? pageNo,
    ps: response.data.data.page?.ps ?? pageSize,
    items: (response.data.data.items ?? []).map(mapRole),
  }
}

export async function getPermissionRole(id: number): Promise<PermissionRole> {
  const response = await apiClient.get<PermissionApiResponse<{
    id: number
    role_name: string
    role_type: string
    users: number
    updated_at: string
  }>>(`/permission/roles/${id}`)
  return mapRole(response.data.data)
}

export async function createPermissionRole(payload: CreatePermissionRolePayload): Promise<void> {
  await apiClient.post<PermissionApiResponse<{ success: boolean; action: string }>>('/permission/roles', {
    role_name: payload.roleName,
    role_type: mapRoleType(payload.roleType),
  })
}

export async function updatePermissionRole(id: number, payload: UpdatePermissionRolePayload): Promise<void> {
  await apiClient.put<PermissionApiResponse<{ success: boolean; action: string }>>(`/permission/roles/${id}`, {
    role_name: payload.roleName,
    role_type: mapRoleType(payload.roleType),
  })
}

export async function deletePermissionRole(id: number): Promise<void> {
  await apiClient.delete<PermissionApiResponse<{ success: boolean; action: string }>>(`/permission/roles/${id}`)
}

export async function getPermissionRoleMenus(roleId: number): Promise<number[]> {
  const response = await apiClient.get<PermissionApiResponse<{ menu_ids: number[] }>>(`/permission/roles/${roleId}/menus`)
  return response.data.data.menu_ids ?? []
}

export async function updatePermissionRoleMenus(roleId: number, menuIds: number[]): Promise<void> {
  await apiClient.post<PermissionApiResponse<{ success: boolean; action: string }>>(`/permission/roles/${roleId}/menus`, {
    role_id: roleId,
    menu_ids: menuIds,
  })
}
