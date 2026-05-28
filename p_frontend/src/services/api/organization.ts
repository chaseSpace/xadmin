import { apiClient } from './client'

type OrganizationUsersApiData = {
  total: string
  page: {
    pn: number
    ps: number
  }
  items: Array<{
    uid: number
    username: string
    display_name: string
    avatar: string
    email: string
    phone: string
    account_status: 'active' | 'disabled' | 'deactivated'
    online_status: 'online' | 'offline'
    active_session_count: number
    last_login_ip: string
    last_login_at: string
    department_id: number | string
    department_name: string
    position_id: number | string
    position_name: string
    role_names?: string[]
  }>
}

type OrganizationUsersApiResponse = {
  code: number
  message: string
  data: OrganizationUsersApiData
}

type OrganizationActionApiResponse = {
  code: number
  message: string
  data: {
    success: boolean
    action: string
    temp_password?: string
  }
}

type OrganizationImportApiResponse = {
  code: number
  message: string
  data: {
    success: boolean
    action: string
  }
}

export type OrganizationUser = {
  uid: number
  username: string
  displayName: string
  avatar: string
  email: string
  phone: string
  accountStatus: 'active' | 'disabled' | 'deactivated'
  onlineStatus: 'online' | 'offline'
  activeSessionCount: number
  lastLoginIp: string
  lastLoginAt: string
  departmentId: number
  departmentName: string
  positionId: number
  positionName: string
  roleNames: string[]
}

export type OrganizationUsersPage = {
  total: number
  pn: number
  ps: number
  items: OrganizationUser[]
}

export type OrganizationUsersFilters = {
  keyword?: string
  phone?: string
  status?: 'active' | 'disabled' | 'deactivated'
  departmentId?: number
  positionId?: number
  createdFrom?: string
  createdTo?: string
}

export type OrganizationUserStatus = 0 | 1 | 2

export type CreateOrganizationUserPayload = {
  username: string
  password: string
  displayName: string
  email: string
  phone: string
  status: OrganizationUserStatus
  departmentId: number
  positionId: number
}

export type UpdateOrganizationUserPayload = {
  displayName: string
  avatar: string
  email: string
  phone: string
  status: OrganizationUserStatus
  departmentId: number
  positionId: number
}

export type ImportOrganizationUserItem = {
  username: string
  password: string
  displayName: string
  email: string
  phone: string
  status: OrganizationUserStatus
}

export async function getOrganizationUsers(
  pageNo = 1,
  pageSize = 10,
  orderField?: 'uid' | 'username' | 'display_name' | 'status' | 'active_session_count' | 'last_login_at',
  orderType?: 'asc' | 'desc',
  filters?: OrganizationUsersFilters,
): Promise<OrganizationUsersPage> {
  const response = await apiClient.get<OrganizationUsersApiResponse>('/organization/users', {
    params: {
      page_no: pageNo,
      page_size: pageSize,
      order_field: orderField,
      order_type: orderType,
      keyword: filters?.keyword,
      phone: filters?.phone,
      status: filters?.status,
      department_id: filters?.departmentId,
      position_id: filters?.positionId,
      created_from: filters?.createdFrom,
      created_to: filters?.createdTo,
    },
  })
  return {
    total: Number(response.data.data.total || 0),
    pn: response.data.data.page.pn,
    ps: response.data.data.page.ps,
    items: response.data.data.items.map((item) => ({
      uid: item.uid,
      username: item.username,
      displayName: item.display_name,
      avatar: item.avatar,
      email: item.email,
      phone: item.phone,
      accountStatus: item.account_status,
      onlineStatus: item.online_status,
      activeSessionCount: item.active_session_count,
      lastLoginIp: item.last_login_ip,
      lastLoginAt: item.last_login_at,
      departmentId: Number(item.department_id || 0),
      departmentName: item.department_name,
      positionId: Number(item.position_id || 0),
      positionName: item.position_name,
      roleNames: Array.isArray(item.role_names) ? item.role_names.map(String).filter(Boolean) : [],
    })),
  }
}

type OrganizationUserSessionsApiResponse = {
  code: number
  message: string
  data: {
    items: Array<{
      session_id: string
      status: string
      login_ip: string
      user_agent: string
      last_seen_at: string
      expired_at: string
      revoked_at: string
      revoked_reason: string
    }>
  }
}

export type OrganizationUserSessionItem = {
  sessionId: string
  status: string
  loginIp: string
  userAgent: string
  lastSeenAt: string
  expiredAt: string
  revokedAt: string
  revokedReason: string
}

export async function getOrganizationUserSessions(
  uid: number,
  status?: 'active' | 'revoked' | 'expired',
  pageSize = 10,
): Promise<OrganizationUserSessionItem[]> {
  const response = await apiClient.get<OrganizationUserSessionsApiResponse>(`/organization/users/${uid}/sessions`, {
    params: {
      status,
      page_size: pageSize,
    },
  })
  return response.data.data.items.map((item) => ({
    sessionId: item.session_id,
    status: item.status,
    loginIp: item.login_ip,
    userAgent: item.user_agent,
    lastSeenAt: item.last_seen_at,
    expiredAt: item.expired_at,
    revokedAt: item.revoked_at,
    revokedReason: item.revoked_reason,
  }))
}

export async function createOrganizationUser(payload: CreateOrganizationUserPayload): Promise<void> {
  await apiClient.post<OrganizationActionApiResponse>('/organization/users', {
    username: payload.username,
    password: payload.password,
    display_name: payload.displayName,
    email: payload.email,
    phone: payload.phone,
    status: payload.status,
    department_id: payload.departmentId,
    position_id: payload.positionId,
  })
}

export async function updateOrganizationUser(uid: number, payload: UpdateOrganizationUserPayload): Promise<void> {
  await apiClient.put<OrganizationActionApiResponse>(`/organization/users/${uid}`, {
    display_name: payload.displayName,
    avatar: payload.avatar,
    email: payload.email,
    phone: payload.phone,
    status: payload.status,
    department_id: payload.departmentId,
    position_id: payload.positionId,
  })
}

export async function deleteOrganizationUser(uid: number): Promise<void> {
  await apiClient.delete<OrganizationActionApiResponse>(`/organization/users/${uid}`)
}

export async function resetOrganizationUserPassword(uid: number): Promise<string> {
  const response = await apiClient.post<OrganizationActionApiResponse>(`/organization/users/${uid}/reset_password`, {})
  return response.data.data.temp_password ?? ''
}

export async function importOrganizationUsers(items: ImportOrganizationUserItem[]): Promise<void> {
  await apiClient.post<OrganizationImportApiResponse>('/organization/users/import', {
    items: items.map((item) => ({
      username: item.username,
      password: item.password,
      display_name: item.displayName,
      email: item.email,
      phone: item.phone,
      status: item.status,
    })),
  })
}

export async function exportOrganizationUsers(filters?: OrganizationUsersFilters): Promise<Blob> {
  const response = await apiClient.get<Blob>('/organization/users/export', {
    params: {
      keyword: filters?.keyword,
      phone: filters?.phone,
      status: filters?.status,
      department_id: filters?.departmentId,
      position_id: filters?.positionId,
      created_from: filters?.createdFrom,
      created_to: filters?.createdTo,
    },
    responseType: 'blob',
  })
  return response.data
}

type OrganizationDepartmentApiItem = {
  id: number | string
  parent_id: number | string
  name: string
  code: string
  status: 'enabled' | 'disabled'
  member_count: number | string
  position_count?: number | string
  updated_at: string
  children?: OrganizationDepartmentApiItem[]
}

type OrganizationDepartmentTreeApiResponse = {
  code: number
  message: string
  data: {
    items: OrganizationDepartmentApiItem[]
  }
}

type OrganizationDepartmentDetailApiResponse = {
  code: number
  message: string
  data: OrganizationDepartmentApiItem
}

export type OrganizationDepartment = {
  id: number
  parentId: number
  name: string
  code: string
  status: 'enabled' | 'disabled'
  memberCount: number
  positionCount: number
  updatedAt: string
  children: OrganizationDepartment[]
}

export type OrganizationDepartmentStats = {
  directMemberCount: number
  directPositionCount: number
  totalMemberCount: number
  totalPositionCount: number
}

export type CreateOrganizationDepartmentPayload = {
  parentId: number
  name: string
  code: string
}

export type UpdateOrganizationDepartmentPayload = {
  name: string
  code: string
}

function mapDepartmentItem(item: OrganizationDepartmentApiItem): OrganizationDepartment {
  const id = Number(item.id)
  const parentId = Number(item.parent_id)
  const memberCount = Number(item.member_count)
  const positionCount = Number(item.position_count)
  return {
    id: Number.isFinite(id) ? id : 0,
    parentId: Number.isFinite(parentId) ? parentId : 0,
    name: item.name,
    code: item.code,
    status: item.status,
    memberCount: Number.isFinite(memberCount) ? memberCount : 0,
    positionCount: Number.isFinite(positionCount) ? positionCount : 0,
    updatedAt: item.updated_at,
    children: (item.children ?? []).map(mapDepartmentItem),
  }
}

export function computeDepartmentStats(node: OrganizationDepartment): OrganizationDepartmentStats {
  const directMemberCount = Number.isFinite(node.memberCount) ? node.memberCount : 0
  const directPositionCount = Number.isFinite(node.positionCount) ? node.positionCount : 0
  let totalMemberCount = directMemberCount
  let totalPositionCount = directPositionCount
  for (const child of node.children) {
    const childStats = computeDepartmentStats(child)
    totalMemberCount += childStats.totalMemberCount
    totalPositionCount += childStats.totalPositionCount
  }
  return {
    directMemberCount,
    directPositionCount,
    totalMemberCount,
    totalPositionCount,
  }
}

export async function getOrganizationDepartmentsTree(): Promise<OrganizationDepartment[]> {
  const response = await apiClient.get<OrganizationDepartmentTreeApiResponse>('/organization/departments/tree')
  return response.data.data.items.map(mapDepartmentItem)
}

export async function getOrganizationDepartment(id: number): Promise<OrganizationDepartment> {
  const response = await apiClient.get<OrganizationDepartmentDetailApiResponse>(`/organization/departments/${id}`)
  return mapDepartmentItem(response.data.data)
}

export async function createOrganizationDepartment(payload: CreateOrganizationDepartmentPayload): Promise<void> {
  await apiClient.post<OrganizationActionApiResponse>('/organization/departments', {
    parent_id: payload.parentId,
    name: payload.name,
    code: payload.code,
  })
}

export async function updateOrganizationDepartment(id: number, payload: UpdateOrganizationDepartmentPayload): Promise<void> {
  await apiClient.put<OrganizationActionApiResponse>(`/organization/departments/${id}`, {
    name: payload.name,
    code: payload.code,
  })
}

export async function toggleOrganizationDepartmentStatus(id: number, enabled: boolean): Promise<void> {
  await apiClient.post<OrganizationActionApiResponse>(`/organization/departments/${id}/status`, {
    enabled,
  })
}

export async function deleteOrganizationDepartment(id: number, force = false): Promise<void> {
  await apiClient.delete<OrganizationActionApiResponse>(`/organization/departments/${id}`, {
    params: { force: force ? 'true' : undefined },
  })
}

type OrganizationPositionApiItem = {
  id: number | string
  name: string
  code: string
  department_id: number | string
  department_name: string
  level: string
  hc: number | string
  staffed: number | string
  related_count?: number | string
  status: 'enabled' | 'disabled'
  updated_at: string
  role_ids?: Array<number | string>
  role_names?: string[]
}

type OrganizationPositionsApiResponse = {
  code: number
  message: string
  data: {
    total: string
    page: {
      pn: number
      ps: number
    }
    items: OrganizationPositionApiItem[]
  }
}

type OrganizationPositionDetailApiResponse = {
  code: number
  message: string
  data: OrganizationPositionApiItem
}

export type OrganizationPosition = {
  id: number
  name: string
  code: string
  departmentId: number
  departmentName: string
  level: string
  hc: number
  staffed: number
  relatedCount: number
  status: 'enabled' | 'disabled'
  updatedAt: string
  roleIds: number[]
  roleNames: string[]
}

export type OrganizationPositionsPage = {
  total: number
  pn: number
  ps: number
  items: OrganizationPosition[]
}

export type OrganizationPositionFilters = {
  keyword?: string
  departmentId?: number
  level?: string
  status?: 'enabled' | 'disabled'
}

export type CreateOrganizationPositionPayload = {
  name: string
  code: string
  departmentId: number
  level: string
  hc: number
  staffed: number
  roleIds: number[]
}

export type UpdateOrganizationPositionPayload = {
  name: string
  code: string
  departmentId: number
  level: string
  hc: number
  staffed: number
  status: 'enabled' | 'disabled'
  roleIds: number[]
}

function mapPositionItem(item: OrganizationPositionApiItem): OrganizationPosition {
  const id = Number(item.id)
  const departmentId = Number(item.department_id)
  const hc = Number(item.hc)
  const staffed = Number(item.staffed)
  const relatedCount = Number(item.related_count)
  return {
    id: Number.isFinite(id) ? id : 0,
    name: item.name,
    code: item.code,
    departmentId: Number.isFinite(departmentId) ? departmentId : 0,
    departmentName: item.department_name,
    level: item.level,
    hc: Number.isFinite(hc) ? hc : 0,
    staffed: Number.isFinite(staffed) ? staffed : 0,
    relatedCount: Number.isFinite(relatedCount) ? relatedCount : Number.isFinite(hc) ? hc : 0,
    status: item.status,
    updatedAt: item.updated_at,
    roleIds: (item.role_ids ?? []).map((id) => Number(id)).filter((id) => Number.isFinite(id) && id > 0),
    roleNames: (item.role_names ?? []).map((name) => String(name).trim()).filter((name) => name.length > 0),
  }
}

export async function getOrganizationPositions(
  pageNo = 1,
  pageSize = 10,
  orderField?: 'id' | 'name' | 'level' | 'status' | 'updated_at',
  orderType?: 'asc' | 'desc',
  filters?: OrganizationPositionFilters,
): Promise<OrganizationPositionsPage> {
  const response = await apiClient.get<OrganizationPositionsApiResponse>('/organization/positions', {
    params: {
      page_no: pageNo,
      page_size: pageSize,
      order_field: orderField,
      order_type: orderType,
      keyword: filters?.keyword,
      department_id: filters?.departmentId,
      level: filters?.level,
      status: filters?.status,
    },
  })
  return {
    total: Number(response.data.data.total || 0),
    pn: response.data.data.page.pn,
    ps: response.data.data.page.ps,
    items: response.data.data.items.map(mapPositionItem),
  }
}

export async function getOrganizationPosition(id: number): Promise<OrganizationPosition> {
  const response = await apiClient.get<OrganizationPositionDetailApiResponse>(`/organization/positions/${id}`)
  return mapPositionItem(response.data.data)
}

export async function createOrganizationPosition(payload: CreateOrganizationPositionPayload): Promise<void> {
  await apiClient.post<OrganizationActionApiResponse>('/organization/positions', {
    name: payload.name,
    code: payload.code,
    department_id: payload.departmentId,
    level: payload.level,
    hc: payload.hc,
    staffed: payload.staffed,
    role_ids: payload.roleIds,
  })
}

export async function updateOrganizationPosition(id: number, payload: UpdateOrganizationPositionPayload): Promise<void> {
  await apiClient.put<OrganizationActionApiResponse>(`/organization/positions/${id}`, {
    name: payload.name,
    code: payload.code,
    department_id: payload.departmentId,
    level: payload.level,
    hc: payload.hc,
    staffed: payload.staffed,
    status: payload.status === 'enabled' ? 1 : 0,
    role_ids: payload.roleIds,
  })
}

export async function toggleOrganizationPositionStatus(id: number, enabled: boolean): Promise<void> {
  await apiClient.post<OrganizationActionApiResponse>(`/organization/positions/${id}/status`, {
    enabled,
  })
}

export async function deleteOrganizationPosition(id: number): Promise<void> {
  await apiClient.delete<OrganizationActionApiResponse>(`/organization/positions/${id}`)
}

export async function batchTransferOrganizationUsers(uids: number[], departmentId: number, positionId: number): Promise<void> {
  await apiClient.post<OrganizationActionApiResponse>('/organization/users/transfer-position', {
    uids,
    department_id: departmentId,
    position_id: positionId,
  })
}
