import { apiClient } from './client'
import { WARM_TIP_ENABLED_PAGE_SIZE } from './queryKeys'

type SystemAuditLogApiItem = {
  id: number
  uid: number
  actor: string
  action: string
  result: string
  trace_id: string
  request_id: string
  source_ip: string
  duration: string
  user_agent: string
  detail: string
  created_at: string
}

type SystemAuditLogsApiResponse = {
  code: number
  message: string
  data: {
    items: SystemAuditLogApiItem[]
    total: number
    page: {
      pn: number
      ps: number
    }
  }
}

type SystemAuditLogDetailApiResponse = {
  code: number
  message: string
  data: SystemAuditLogApiItem
}

export type SystemAuditLog = {
  id: number
  uid: number
  actor: string
  action: string
  result: 'success' | 'failed'
  traceId: string
  requestId: string
  sourceIp: string
  duration: string
  userAgent: string
  detail: string
  createdAt: string
}

export type SystemAuditLogFilters = {
  actor?: string
  action?: string
  result?: 'success' | 'failed'
  traceId?: string
  requestId?: string
  sourceIp?: string
  keyword?: string
  createdFrom?: string
  createdTo?: string
}

export type SystemAuditLogsPage = {
  total: number
  pn: number
  ps: number
  items: SystemAuditLog[]
}

type SystemAuditRetentionApiResponse = {
  code: number
  message: string
  data: {
    success: boolean
    retain_days: number
    expired_count: number
    valid_count: number
    cutoff_at: string
  }
}

export type SystemAuditRetentionResult = {
  success: boolean
  retainDays: number
  expiredCount: number
  validCount: number
  cutoffAt: string
}

function mapAuditLog(item: SystemAuditLogApiItem): SystemAuditLog {
  return {
    id: Number(item.id) || 0,
    uid: Number(item.uid) || 0,
    actor: String(item.actor || ''),
    action: String(item.action || ''),
    result: item.result === 'failed' ? 'failed' : 'success',
    traceId: String(item.trace_id || ''),
    requestId: String(item.request_id || ''),
    sourceIp: String(item.source_ip || ''),
    duration: String(item.duration || ''),
    userAgent: String(item.user_agent || ''),
    detail: String(item.detail || ''),
    createdAt: String(item.created_at || ''),
  }
}

export async function getSystemAuditLogs(
  pageNo = 1,
  pageSize = 10,
  orderField?: 'id' | 'uid' | 'actor' | 'action' | 'result' | 'trace_id' | 'source_ip' | 'created_at',
  orderType?: 'asc' | 'desc',
  filters?: SystemAuditLogFilters,
): Promise<SystemAuditLogsPage> {
  const response = await apiClient.get<SystemAuditLogsApiResponse>('/system/audit-logs', {
    params: {
      page_no: pageNo,
      page_size: pageSize,
      order_field: orderField,
      order_type: orderType,
      actor: filters?.actor,
      action: filters?.action,
      result: filters?.result,
      trace_id: filters?.traceId,
      request_id: filters?.requestId,
      source_ip: filters?.sourceIp,
      keyword: filters?.keyword,
      created_from: filters?.createdFrom,
      created_to: filters?.createdTo,
    },
  })
  return {
    total: Number(response.data.data.total || 0),
    pn: Number(response.data.data.page?.pn || pageNo),
    ps: Number(response.data.data.page?.ps || pageSize),
    items: (response.data.data.items || []).map(mapAuditLog),
  }
}

export async function getSystemAuditLog(id: number): Promise<SystemAuditLog> {
  const response = await apiClient.get<SystemAuditLogDetailApiResponse>(`/system/audit-logs/${id}`)
  return mapAuditLog(response.data.data)
}

export async function exportSystemAuditLogs(
  orderField?: 'id' | 'uid' | 'actor' | 'action' | 'result' | 'trace_id' | 'source_ip' | 'created_at',
  orderType?: 'asc' | 'desc',
  filters?: SystemAuditLogFilters,
): Promise<Blob> {
  const response = await apiClient.get('/system/audit-logs/export', {
    params: {
      page_no: 1,
      page_size: 10,
      order_field: orderField,
      order_type: orderType,
      actor: filters?.actor,
      action: filters?.action,
      result: filters?.result,
      trace_id: filters?.traceId,
      request_id: filters?.requestId,
      source_ip: filters?.sourceIp,
      keyword: filters?.keyword,
      created_from: filters?.createdFrom,
      created_to: filters?.createdTo,
    },
    responseType: 'blob',
  })
  return response.data as Blob
}

export async function applySystemAuditLogRetention(retainDays: number, confirm = false): Promise<SystemAuditRetentionResult> {
  const response = await apiClient.post<SystemAuditRetentionApiResponse>('/system/audit-logs/retention', {
    retain_days: retainDays,
    confirm,
  })
  return {
    success: Boolean(response.data.data.success),
    retainDays: Number(response.data.data.retain_days || retainDays),
    expiredCount: Number(response.data.data.expired_count || 0),
    validCount: Number(response.data.data.valid_count || 0),
    cutoffAt: String(response.data.data.cutoff_at || ''),
  }
}

type SystemActionResp = {
  code: number
  message: string
  data: {
    success: boolean
    action: string
  }
}

type SystemIPBlacklistApiItem = {
  id: number
  ip: string
  ban_type: string
  start_at: string
  end_at: string
  reason: string
  creator: string
  status: string
  hit_count: number
  updated_at: string
}

type SystemIPBlacklistApiResp = {
  code: number
  message: string
  data: {
    items: SystemIPBlacklistApiItem[]
    total: number
    page: {
      pn: number
      ps: number
    }
  }
}

export type SystemIPBlacklistItem = {
  id: number
  ip: string
  banType: 'temp' | 'permanent'
  startAt: string
  endAt: string
  reason: string
  creator: string
  status: 'active' | 'expired' | 'manual_inactive'
  hitCount: number
  updatedAt: string
}

export type SystemIPBlacklistPage = {
  total: number
  pn: number
  ps: number
  items: SystemIPBlacklistItem[]
}

export type SystemIPBlacklistFilters = {
  keyword?: string
  status?: 'active' | 'inactive'
  banType?: 'temp' | 'permanent'
  creator?: string
}

export type CreateSystemIPBlacklistPayload = {
  ip: string
  startAt?: string
  endAt?: string
  reason: string
  creator?: string
}

export type UpdateSystemIPBlacklistPayload = {
  banType: 'temp' | 'permanent'
  endAt?: string
  reason: string
}

function mapIPBlacklistItem(item: SystemIPBlacklistApiItem): SystemIPBlacklistItem {
  return {
    id: Number(item.id) || 0,
    ip: String(item.ip || ''),
    banType: item.ban_type === 'permanent' ? 'permanent' : 'temp',
    startAt: String(item.start_at || ''),
    endAt: String(item.end_at || ''),
    reason: String(item.reason || ''),
    creator: String(item.creator || ''),
    status: item.status === 'expired' ? 'expired' : item.status === 'manual_inactive' || item.status === 'inactive' ? 'manual_inactive' : 'active',
    hitCount: Number(item.hit_count) || 0,
    updatedAt: String(item.updated_at || ''),
  }
}

function mapBanTypeToNum(type: 'temp' | 'permanent'): number {
  return type === 'permanent' ? 2 : 1
}

export async function getSystemIPBlacklist(
  pageNo = 1,
  pageSize = 10,
  orderField?: 'id' | 'ip' | 'ban_type' | 'status' | 'hit_count' | 'updated_at',
  orderType?: 'asc' | 'desc',
  filters?: SystemIPBlacklistFilters,
): Promise<SystemIPBlacklistPage> {
  const response = await apiClient.get<SystemIPBlacklistApiResp>('/system/ip-blacklist', {
    params: {
      page_no: pageNo,
      page_size: pageSize,
      order_field: orderField,
      order_type: orderType,
      keyword: filters?.keyword,
      status: filters?.status,
      ban_type: filters?.banType,
      creator: filters?.creator,
    },
  })
  return {
    total: Number(response.data.data.total || 0),
    pn: Number(response.data.data.page?.pn || pageNo),
    ps: Number(response.data.data.page?.ps || pageSize),
    items: (response.data.data.items || []).map(mapIPBlacklistItem),
  }
}

export async function createSystemIPBlacklist(payload: CreateSystemIPBlacklistPayload): Promise<void> {
  await apiClient.post<SystemActionResp>('/system/ip-blacklist', {
    ip: payload.ip,
    start_at: payload.startAt || '',
    end_at: payload.endAt || '',
    reason: payload.reason,
    creator: payload.creator || '',
  })
}

export async function updateSystemIPBlacklist(id: number, payload: UpdateSystemIPBlacklistPayload): Promise<void> {
  await apiClient.put<SystemActionResp>(`/system/ip-blacklist/${id}`, {
    ban_type: mapBanTypeToNum(payload.banType),
    end_at: payload.endAt || '',
    reason: payload.reason,
  })
}

export async function unblockSystemIPBlacklist(id: number): Promise<void> {
  await apiClient.post<SystemActionResp>(`/system/ip-blacklist/${id}/unblock`, {})
}

export async function deleteSystemIPBlacklist(id: number): Promise<void> {
  await apiClient.delete<SystemActionResp>(`/system/ip-blacklist/${id}`)
}

export async function batchUnblockSystemIPBlacklist(ids: number[]): Promise<void> {
  await apiClient.post<SystemActionResp>('/system/ip-blacklist/unblock-batch', {
    ids,
  })
}

export type ImportSystemIPBlacklistPayload = {
  ips: string[]
  durationHours?: number
  endAt?: string
}

export async function importSystemIPBlacklist(payload: ImportSystemIPBlacklistPayload): Promise<void> {
  await apiClient.post<SystemActionResp>('/system/ip-blacklist/import', {
    ips: payload.ips,
    duration_hours: payload.durationHours,
    end_at: payload.endAt,
  })
}

export async function getSystemIPBlacklistCreators(): Promise<string[]> {
  const response = await apiClient.get<{ code: number; message: string; data: { creators: string[] } }>('/system/ip-blacklist/creators')
  return response.data.data.creators || []
}

type SystemWarmTipApiItem = {
  id: number
  tip_type: string
  content_zh: string
  content_en: string
  sort: number
  status: number
  updated_at: string
}

type SystemWarmTipsApiResp = {
  code: number
  message: string
  data: {
    items: SystemWarmTipApiItem[]
    total: number
    page: {
      pn: number
      ps: number
    }
  }
}

export type SystemWarmTipType = 'rest' | 'positive' | 'quote' | 'line'

export type SystemWarmTipItem = {
  id: number
  tipType: SystemWarmTipType
  contentZh: string
  contentEn: string
  sort: number
  status: 0 | 1
  updatedAt: string
}

export type SystemWarmTipsPage = {
  total: number
  pn: number
  ps: number
  items: SystemWarmTipItem[]
}

export type SystemWarmTipFilters = {
  keyword?: string
  tipType?: SystemWarmTipType
  status?: 'enabled' | 'disabled'
}

export type SaveSystemWarmTipPayload = {
  tipType: SystemWarmTipType
  contentZh: string
  contentEn: string
  sort: number
  status: 0 | 1
}

function mapWarmTipItem(item: SystemWarmTipApiItem): SystemWarmTipItem {
  const tipType = ['rest', 'positive', 'quote', 'line'].includes(item.tip_type) ? item.tip_type as SystemWarmTipType : 'rest'
  return {
    id: Number(item.id) || 0,
    tipType,
    contentZh: String(item.content_zh || ''),
    contentEn: String(item.content_en || ''),
    sort: Number(item.sort) || 0,
    status: Number(item.status) === 0 ? 0 : 1,
    updatedAt: String(item.updated_at || ''),
  }
}

export async function getSystemWarmTips(
  pageNo = 1,
  pageSize = 10,
  orderField?: 'id' | 'tip_type' | 'sort' | 'status' | 'updated_at',
  orderType?: 'asc' | 'desc',
  filters?: SystemWarmTipFilters,
): Promise<SystemWarmTipsPage> {
  const response = await apiClient.get<SystemWarmTipsApiResp>('/system/warm-tips', {
    params: {
      page_no: pageNo,
      page_size: pageSize,
      order_field: orderField,
      order_type: orderType,
      keyword: filters?.keyword,
      tip_type: filters?.tipType,
      status: filters?.status,
    },
  })
  return {
    total: Number(response.data.data.total || 0),
    pn: Number(response.data.data.page?.pn || pageNo),
    ps: Number(response.data.data.page?.ps || pageSize),
    items: (response.data.data.items || []).map(mapWarmTipItem),
  }
}

export async function getEnabledSystemWarmTips(): Promise<SystemWarmTipItem[]> {
  const page = await getSystemWarmTips(1, WARM_TIP_ENABLED_PAGE_SIZE, 'sort', 'asc', { status: 'enabled' })
  return page.items
}

export async function createSystemWarmTip(payload: SaveSystemWarmTipPayload): Promise<void> {
  await apiClient.post<SystemActionResp>('/system/warm-tips', {
    tip_type: payload.tipType,
    content_zh: payload.contentZh,
    content_en: payload.contentEn,
    sort: payload.sort,
    status: payload.status,
  })
}

export async function updateSystemWarmTip(id: number, payload: SaveSystemWarmTipPayload): Promise<void> {
  await apiClient.put<SystemActionResp>(`/system/warm-tips/${id}`, {
    tip_type: payload.tipType,
    content_zh: payload.contentZh,
    content_en: payload.contentEn,
    sort: payload.sort,
    status: payload.status,
  })
}

export async function deleteSystemWarmTip(id: number): Promise<void> {
  await apiClient.delete<SystemActionResp>(`/system/warm-tips/${id}`)
}

// ===== Alert Bots =====

export type SystemAlertBotItem = {
  id: number
  name: string
  username: string
  token: string
  botType: string
  enabled: boolean
  linkedSceneKeys: string[]
  createdAt: string
  updatedAt: string
}

export type SystemAlertBotPage = {
  total: number
  pn: number
  ps: number
  items: SystemAlertBotItem[]
}

export type SaveSystemAlertBotPayload = {
  id?: number
  name: string
  username: string
  token: string
  botType: string
  enabled: boolean
}

export async function getSystemAlertBots(
  pageNo = 1,
  pageSize = 10,
  orderField?: string,
  orderType?: 'asc' | 'desc',
  keyword?: string,
  botType?: string,
): Promise<SystemAlertBotPage> {
  const response = await apiClient.get<{ code: number; message: string; data: { items: any[]; total: number; page: { pn: number; ps: number } } }>('/system/alert-bots', {
    params: { page_no: pageNo, page_size: pageSize, order_field: orderField, order_type: orderType, keyword, bot_type: botType },
  })
  const data = response.data.data
  return {
    total: Number(data.total || 0),
    pn: Number(data.page?.pn || pageNo),
    ps: Number(data.page?.ps || pageSize),
    items: (data.items || []).map((item: any) => ({
      id: Number(item.id) || 0,
      name: String(item.name || ''),
      username: String(item.username || ''),
      token: String(item.token || ''),
      botType: String(item.bot_type || 'telegram'),
      enabled: Boolean(item.enabled),
      linkedSceneKeys: Array.isArray(item.linked_scene_keys) ? item.linked_scene_keys : [],
      createdAt: String(item.created_at || ''),
      updatedAt: String(item.updated_at || ''),
    })),
  }
}

export async function saveSystemAlertBot(payload: SaveSystemAlertBotPayload): Promise<void> {
  await apiClient.post<SystemActionResp>('/system/alert-bots', {
    id: payload.id || 0,
    name: payload.name,
    username: payload.username,
    token: payload.token,
    bot_type: payload.botType,
    enabled: payload.enabled,
  })
}

export async function deleteSystemAlertBot(id: number): Promise<void> {
  await apiClient.delete<SystemActionResp>(`/system/alert-bots/${id}`)
}

// ===== Alert Scenes =====

export type SystemAlertSceneItem = {
  id: number
  sceneKey: string
  botId: number
  parseMode: string
  groupName: string
  groupId: string
  notifyTemplate: string
  createdAt: string
  updatedAt: string
}

export type SystemAlertScenePage = {
  total: number
  pn: number
  ps: number
  items: SystemAlertSceneItem[]
}

export type SaveSystemAlertScenePayload = {
  id?: number
  sceneKey: string
  botId: number
  parseMode: string
  groupName: string
  groupId: string
  notifyTemplate: string
}

export async function getSystemAlertScenes(
  pageNo = 1,
  pageSize = 10,
  orderField?: string,
  orderType?: 'asc' | 'desc',
  keyword?: string,
): Promise<SystemAlertScenePage> {
  const response = await apiClient.get<{ code: number; message: string; data: { items: any[]; total: number; page: { pn: number; ps: number } } }>('/system/alert-scenes', {
    params: { page_no: pageNo, page_size: pageSize, order_field: orderField, order_type: orderType, keyword },
  })
  const data = response.data.data
  return {
    total: Number(data.total || 0),
    pn: Number(data.page?.pn || pageNo),
    ps: Number(data.page?.ps || pageSize),
    items: (data.items || []).map((item: any) => ({
      id: Number(item.id) || 0,
      sceneKey: String(item.scene_key || ''),
      botId: Number(item.bot_id) || 0,
      parseMode: String(item.parse_mode || ''),
      groupName: String(item.group_name || ''),
      groupId: String(item.group_id || ''),
      notifyTemplate: String(item.notify_template || ''),
      createdAt: String(item.created_at || ''),
      updatedAt: String(item.updated_at || ''),
    })),
  }
}

export async function saveSystemAlertScene(payload: SaveSystemAlertScenePayload): Promise<void> {
  await apiClient.post<SystemActionResp>('/system/alert-scenes', {
    id: payload.id || 0,
    scene_key: payload.sceneKey,
    bot_id: payload.botId,
    parse_mode: payload.parseMode,
    group_name: payload.groupName,
    group_id: payload.groupId,
    notify_template: payload.notifyTemplate,
  })
}

export async function deleteSystemAlertScene(id: number): Promise<void> {
  await apiClient.delete<SystemActionResp>(`/system/alert-scenes/${id}`)
}

export async function testSendAlertScene(id: number, variables: Record<string, string>): Promise<void> {
  await apiClient.post<SystemActionResp>(`/system/alert-scenes/${id}/test-send`, { variables })
}

// ===== Alert Templates =====

export type SystemAlertTemplateItem = {
  id: number
  botType: string
  name: string
  parseMode: string
  content: string
  createdAt: string
  updatedAt: string
}

export type SystemAlertTemplatePage = {
  total: number
  pn: number
  ps: number
  items: SystemAlertTemplateItem[]
}

export type SaveSystemAlertTemplatePayload = {
  id?: number
  botType: string
  name: string
  parseMode: string
  content: string
}

export async function getSystemAlertTemplates(
  pageNo = 1, pageSize = 10, orderField?: string, orderType?: 'asc' | 'desc', keyword?: string, botType?: string,
): Promise<SystemAlertTemplatePage> {
  const response = await apiClient.get<{ code: number; message: string; data: { items: any[]; total: number; page: { pn: number; ps: number } } }>('/system/alert-templates', {
    params: { page_no: pageNo, page_size: pageSize, order_field: orderField, order_type: orderType, keyword, bot_type: botType },
  })
  const data = response.data.data
  return {
    total: Number(data.total || 0),
    pn: Number(data.page?.pn || pageNo),
    ps: Number(data.page?.ps || pageSize),
    items: (data.items || []).map((item: any) => ({
      id: Number(item.id) || 0,
      botType: String(item.bot_type || 'telegram'),
      name: String(item.name || ''),
      parseMode: String(item.parse_mode || ''),
      content: String(item.content || ''),
      createdAt: String(item.created_at || ''),
      updatedAt: String(item.updated_at || ''),
    })),
  }
}

export async function saveSystemAlertTemplate(payload: SaveSystemAlertTemplatePayload): Promise<void> {
  await apiClient.post<SystemActionResp>('/system/alert-templates', {
    id: payload.id || 0, bot_type: payload.botType, name: payload.name, parse_mode: payload.parseMode, content: payload.content,
  })
}

export async function deleteSystemAlertTemplate(id: number): Promise<void> {
  await apiClient.delete<SystemActionResp>(`/system/alert-templates/${id}`)
}
