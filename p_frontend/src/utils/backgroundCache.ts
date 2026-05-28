const BG_CACHE_INDEX_KEY = 'xadmin_bg_cache_index_v1'
const BG_CACHE_ENTRY_PREFIX = 'xadmin_bg_cache_entry_v1:'
const MAX_ENTRY_LENGTH = 1_500_000
const MAX_CACHE_ENTRIES = 6

type CacheIndexItem = {
  key: string
  updatedAt: number
}

function hashString(input: string): string {
  let hash = 0
  for (let i = 0; i < input.length; i += 1) {
    hash = (hash << 5) - hash + input.charCodeAt(i)
    hash |= 0
  }
  return String(Math.abs(hash))
}

function buildEntryKey(url: string): string {
  return `${BG_CACHE_ENTRY_PREFIX}${hashString(url)}`
}

function readIndex(): CacheIndexItem[] {
  if (typeof window === 'undefined') return []
  try {
    const raw = window.localStorage.getItem(BG_CACHE_INDEX_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw) as CacheIndexItem[]
    if (!Array.isArray(parsed)) return []
    return parsed.filter((item) => item && typeof item.key === 'string' && typeof item.updatedAt === 'number')
  } catch {
    return []
  }
}

function writeIndex(items: CacheIndexItem[]): void {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(BG_CACHE_INDEX_KEY, JSON.stringify(items))
  } catch {
    // ignore quota errors
  }
}

function trimCache(index: CacheIndexItem[]): CacheIndexItem[] {
  if (typeof window === 'undefined') return index
  if (index.length <= MAX_CACHE_ENTRIES) return index
  const sorted = [...index].sort((a, b) => b.updatedAt - a.updatedAt)
  const keep = sorted.slice(0, MAX_CACHE_ENTRIES)
  const keepSet = new Set(keep.map((item) => item.key))
  for (const item of sorted) {
    if (!keepSet.has(item.key)) {
      window.localStorage.removeItem(item.key)
    }
  }
  return keep
}

function dataUrlToBlob(dataUrl: string): Blob {
  const parts = dataUrl.split(',')
  const meta = parts[0] ?? ''
  const base64 = parts[1] ?? ''
  const contentType = (meta.match(/data:(.*?);base64/) || [])[1] || 'image/png'
  const binary = window.atob(base64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i += 1) {
    bytes[i] = binary.charCodeAt(i)
  }
  return new Blob([bytes], { type: contentType })
}

async function blobToDataURL(blob: Blob): Promise<string> {
  return await new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(reader.error)
    reader.readAsDataURL(blob)
  })
}

export function getCachedBackgroundDataURL(url: string): string {
  if (typeof window === 'undefined') return ''
  const key = buildEntryKey(url)
  return window.localStorage.getItem(key) || ''
}

export function getCachedBackgroundObjectURL(url: string): string {
  if (typeof window === 'undefined') return ''
  const dataUrl = getCachedBackgroundDataURL(url)
  if (!dataUrl) return ''
  try {
    return URL.createObjectURL(dataUrlToBlob(dataUrl))
  } catch {
    return ''
  }
}

function touchCacheKey(key: string): void {
  const index = readIndex().filter((item) => item.key !== key)
  index.unshift({ key, updatedAt: Date.now() })
  writeIndex(trimCache(index))
}

export async function cacheBackgroundImage(url: string): Promise<string> {
  if (typeof window === 'undefined') return ''
  const trimmed = url.trim()
  if (!trimmed) return ''

  const key = buildEntryKey(trimmed)
  const cached = window.localStorage.getItem(key)
  if (cached) {
    touchCacheKey(key)
    return cached
  }

  const response = await fetch(trimmed, { credentials: 'omit' })
  if (!response.ok) {
    throw new Error(`background fetch failed: ${response.status}`)
  }

  const blob = await response.blob()
  const dataUrl = await blobToDataURL(blob)
  if (!dataUrl || dataUrl.length > MAX_ENTRY_LENGTH) {
    return ''
  }

  window.localStorage.setItem(key, dataUrl)
  touchCacheKey(key)
  return dataUrl
}
