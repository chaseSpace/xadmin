import { apiClient } from './client'

export type ResourceFileType = 'image' | 'audio' | 'video' | 'document' | 'archive'
export type ResourceFileAccessMode = 'preview' | 'download'

type ResourceFileApiItem = {
  id: number
  file_type: string
  name: string
  file_url: string
  mime_type: string
  extension: string
  size_bytes: number
  remark: string
  uploaded_at: string
  last_access_at: string
  access_count: number
  exists: boolean | null
  require_auth: boolean
  exists_checked_at: string
  access_mode: string
}

type ResourceFilesApiResp = {
  code: number
  message: string
  data: {
    items: ResourceFileApiItem[]
    total: number
    total_size_bytes: number
    page: {
      pn: number
      ps: number
    }
  }
}

type ResourceUploadApiResp = {
  code: number
  message: string
  data: {
    item: ResourceFileApiItem
  }
}

type ResourceAccessApiResp = {
  code: number
  message: string
  data: {
    file_url: string
    access_count: number
    last_access_at: string
  }
}

type ResourceCheckFilesApiResp = {
  code: number
  message: string
  data: {
    checked_count: number
    exists_count: number
    missing_count: number
  }
}

export type ResourceFile = {
  id: number
  type: ResourceFileType
  name: string
  url: string
  mimeType: string
  extension: string
  sizeBytes: number
  remark: string
  uploadedAt: string
  lastVisitedAt: string
  visitCount: number
  exists: boolean | null
  requireAuth: boolean
  existsCheckedAt: string
  accessMode: ResourceFileAccessMode
}

export type ResourceFilesPage = {
  total: number
  totalSizeBytes: number
  pn: number
  ps: number
  items: ResourceFile[]
}

export type ResourceCheckFilesResult = {
  checkedCount: number
  existsCount: number
  missingCount: number
}

function normalizeResourceType(type: string): ResourceFileType {
  if (type === 'audio' || type === 'video' || type === 'document' || type === 'archive') return type
  return 'image'
}

function normalizeAccessMode(accessMode: string): ResourceFileAccessMode {
  return accessMode === 'download' ? 'download' : 'preview'
}

function resourceFileTypeValue(fileType: ResourceFileType): number {
  switch (fileType) {
    case 'audio':
      return 2
    case 'video':
      return 3
    case 'document':
      return 4
    case 'archive':
      return 5
    default:
      return 1
  }
}

function mapResourceFile(item: ResourceFileApiItem): ResourceFile {
  return {
    id: Number(item.id) || 0,
    type: normalizeResourceType(String(item.file_type || '')),
    name: String(item.name || ''),
    url: String(item.file_url || ''),
    mimeType: String(item.mime_type || ''),
    extension: String(item.extension || ''),
    sizeBytes: Number(item.size_bytes) || 0,
    remark: String(item.remark || ''),
    uploadedAt: String(item.uploaded_at || ''),
    lastVisitedAt: String(item.last_access_at || ''),
    visitCount: Number(item.access_count) || 0,
    exists: typeof item.exists === 'boolean' ? item.exists : null,
    requireAuth: Boolean(item.require_auth),
    existsCheckedAt: String(item.exists_checked_at || ''),
    accessMode: normalizeAccessMode(String(item.access_mode || '')),
  }
}

export async function getResourceFiles(
  pageNo = 1,
  pageSize = 10,
  fileType?: ResourceFileType,
  keyword?: string,
  exists?: boolean,
): Promise<ResourceFilesPage> {
  const response = await apiClient.get<ResourceFilesApiResp>('/resource/files', {
    params: {
      page_no: pageNo,
      page_size: pageSize,
      file_type: fileType,
      keyword,
      exists,
    },
  })
  return {
    total: Number(response.data.data.total || 0),
    totalSizeBytes: Number(response.data.data.total_size_bytes) || 0,
    pn: Number(response.data.data.page?.pn || pageNo),
    ps: Number(response.data.data.page?.ps || pageSize),
    items: (response.data.data.items || []).map(mapResourceFile),
  }
}

export async function uploadResourceFile(
  file: File,
  fileType: ResourceFileType,
  name: string,
  remark: string,
  requireAuth: boolean,
  accessMode: ResourceFileAccessMode,
): Promise<ResourceFile> {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('file_type', fileType)
  formData.append('name', name)
  formData.append('remark', remark)
  formData.append('require_auth', String(requireAuth))
  formData.append('access_mode', accessMode)
  const response = await apiClient.post<ResourceUploadApiResp>('/resource/files', formData)
  return mapResourceFile(response.data.data.item)
}

export async function updateResourceFile(
  id: number,
  fileType: ResourceFileType,
  name: string,
  remark: string,
  requireAuth: boolean,
  accessMode: ResourceFileAccessMode,
): Promise<void> {
  await apiClient.put(`/resource/files/${id}`, {
    file_type: resourceFileTypeValue(fileType),
    name,
    remark,
    require_auth: requireAuth,
    access_mode: accessMode === 'download' ? 2 : 1,
  })
}

export async function deleteResourceFile(id: number): Promise<void> {
  await apiClient.delete(`/resource/files/${id}`)
}

export async function accessResourceFile(id: number): Promise<{ url: string; visitCount: number; lastVisitedAt: string }> {
  const response = await apiClient.post<ResourceAccessApiResp>(`/resource/files/${id}/access`, {})
  return {
    url: String(response.data.data.file_url || ''),
    visitCount: Number(response.data.data.access_count) || 0,
    lastVisitedAt: String(response.data.data.last_access_at || ''),
  }
}

export async function checkResourceFiles(fileType?: ResourceFileType): Promise<ResourceCheckFilesResult> {
  const response = await apiClient.post<ResourceCheckFilesApiResp>(
    '/resource/files/check',
    {},
    {
      params: {
        file_type: fileType,
      },
      timeout: 120_000,
    },
  )
  return {
    checkedCount: Number(response.data.data.checked_count) || 0,
    existsCount: Number(response.data.data.exists_count) || 0,
    missingCount: Number(response.data.data.missing_count) || 0,
  }
}
