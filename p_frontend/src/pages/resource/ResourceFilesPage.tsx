import {
  AudioOutlined,
  CopyOutlined,
  DeleteOutlined,
  DownloadOutlined,
  EditOutlined,
  ClockCircleOutlined,
  FileImageOutlined,
  FileZipOutlined,
  FileTextOutlined,
  ExclamationCircleOutlined,
  QuestionCircleOutlined,
  ReloadOutlined,
  VideoCameraOutlined,
  UploadOutlined,
  LinkOutlined,
} from '@ant-design/icons'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Card,
  Col,
  Drawer,
  Form,
  Input,
  Modal,
  Popover,
  Row,
  Segmented,
  Select,
  Space,
  Spin,
  Switch,
  Table,
  Tag,
  Tooltip,
  Typography,
  Upload,
  message,
} from 'antd'
import type { UploadFile, UploadProps } from 'antd'
import type { Key } from 'react'
import { useMemo, useState } from 'react'
import { UiButton } from '../../components/ui'
import { useI18n } from '../../i18n/messages'
import {
  accessResourceFile,
  checkResourceFiles,
  deleteResourceFile,
  getResourceFiles,
  updateResourceFile,
  uploadResourceFile,
} from '../../services/api/resource'
import { useUiSettingsStore } from '../../store/uiSettings'
import { formatDateTime, normalizeTimezone } from '../../utils/timezone'

type ResourceFileType = 'image' | 'audio' | 'video' | 'document' | 'archive'
type ResourceFileAccessMode = 'preview' | 'download'
type ResourceFileFilter = ResourceFileType | 'all'
type ResourceFileExistsFilter = 'all' | 'exists' | 'missing'

type ResourceFileRecord = {
  id: number
  type: ResourceFileType
  name: string
  url: string
  sizeBytes: number
  uploadedAt: string
  lastVisitedAt: string
  visitCount: number
  remark: string
  exists: boolean | null
  requireAuth: boolean
  existsCheckedAt: string
  accessMode: ResourceFileAccessMode
}

type ResourceFileFormValues = {
  type: ResourceFileType
  name: string
  requireAuth: boolean
  accessMode: ResourceFileAccessMode
  remark?: string
}

const EMPTY_RESOURCE_FILES: ResourceFileRecord[] = []
const RESOURCE_FILES_TABLE_SCROLL_Y = 336

type DetectedResourceFile = {
  type: ResourceFileType
  label: string
  extension: string
  mime: string
}

const RESOURCE_FILE_TYPE_META: Record<
  ResourceFileType,
  {
    icon:
      | typeof FileImageOutlined
      | typeof AudioOutlined
      | typeof VideoCameraOutlined
      | typeof FileTextOutlined
      | typeof FileZipOutlined
    label: string
    limitMB: number
    description: string
    color: string
  }
> = {
  image: {
    icon: FileImageOutlined,
    label: '图片',
    limitMB: 5,
    description: '封面、截图、海报与素材图片。',
    color: 'blue',
  },
  audio: {
    icon: AudioOutlined,
    label: '语音',
    limitMB: 5,
    description: '语音留言、播报与音频素材。',
    color: 'gold',
  },
  video: {
    icon: VideoCameraOutlined,
    label: '视频',
    limitMB: 100,
    description: '短视频、课程片段与演示录像。',
    color: 'purple',
  },
  document: {
    icon: FileTextOutlined,
    label: '文档',
    limitMB: 20,
    description: 'PDF、Word、PPT 与附件文档。',
    color: 'green',
  },
  archive: {
    icon: FileZipOutlined,
    label: '压缩包',
    limitMB: 100,
    description: 'ZIP、RAR、7Z 与 GZ 压缩文件。',
    color: 'cyan',
  },
}

const SUPPORTED_RESOURCE_FILE_TYPES: Record<ResourceFileType, DetectedResourceFile[]> = {
  image: [
    { type: 'image', label: 'JPEG', extension: '.jpg', mime: 'image/jpeg' },
    { type: 'image', label: 'PNG', extension: '.png', mime: 'image/png' },
    { type: 'image', label: 'GIF', extension: '.gif', mime: 'image/gif' },
    { type: 'image', label: 'WebP', extension: '.webp', mime: 'image/webp' },
    { type: 'image', label: 'AVIF', extension: '.avif', mime: 'image/avif' },
  ],
  audio: [
    { type: 'audio', label: 'MP3', extension: '.mp3', mime: 'audio/mpeg' },
    { type: 'audio', label: 'WAV', extension: '.wav', mime: 'audio/wav' },
    { type: 'audio', label: 'OGG', extension: '.ogg', mime: 'audio/ogg' },
    { type: 'audio', label: 'FLAC', extension: '.flac', mime: 'audio/flac' },
    { type: 'audio', label: 'M4A', extension: '.m4a', mime: 'audio/mp4' },
  ],
  video: [
    { type: 'video', label: 'MP4', extension: '.mp4', mime: 'video/mp4' },
    { type: 'video', label: 'MOV', extension: '.mov', mime: 'video/quicktime' },
    { type: 'video', label: 'WebM', extension: '.webm', mime: 'video/webm' },
    { type: 'video', label: 'AVI', extension: '.avi', mime: 'video/x-msvideo' },
  ],
  document: [
    { type: 'document', label: 'PDF', extension: '.pdf', mime: 'application/pdf' },
    { type: 'document', label: 'DOC/XLS/PPT', extension: '.doc', mime: 'application/msword' },
    { type: 'document', label: 'TXT/MD', extension: '.txt', mime: 'text/plain' },
  ],
  archive: [
    { type: 'archive', label: 'ZIP', extension: '.zip', mime: 'application/zip' },
    { type: 'archive', label: 'RAR', extension: '.rar', mime: 'application/vnd.rar' },
    { type: 'archive', label: '7Z', extension: '.7z', mime: 'application/x-7z-compressed' },
    { type: 'archive', label: 'GZ', extension: '.gz', mime: 'application/gzip' },
  ],
}

function formatFileSize(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 B'
  if (bytes < 1024) return `${bytes} B`
  const kb = bytes / 1024
  if (kb < 1024) return `${kb.toFixed(kb < 10 ? 1 : 0)} KB`
  const mb = kb / 1024
  if (mb < 1024) return `${mb.toFixed(mb < 10 ? 1 : 0)} MB`
  const gb = mb / 1024
  return `${gb.toFixed(gb < 10 ? 1 : 0)} GB`
}

function fileTypeLabel(type: ResourceFileType): string {
  return RESOURCE_FILE_TYPE_META[type].label
}

function resourceFileTypeAccept(type: ResourceFileType): string {
  switch (type) {
    case 'image':
      return 'image/*'
    case 'audio':
      return 'audio/*'
    case 'video':
      return 'video/*'
    case 'document':
      return '.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.md'
    case 'archive':
      return '.zip,.rar,.7z,.gz'
    default:
      return '*/*'
  }
}

function isResourceValidityStale(checkedAt: string): boolean {
  if (!checkedAt) return false
  const checkedTime = new Date(checkedAt.replace(' ', 'T')).getTime()
  if (!Number.isFinite(checkedTime)) return false
  return Date.now() - checkedTime > 3 * 24 * 60 * 60 * 1000
}

function bytesToAscii(bytes: Uint8Array, start = 0, end = bytes.length): string {
  return Array.from(bytes.slice(start, end))
    .map((byte) => String.fromCharCode(byte))
    .join('')
}

function startsWithBytes(bytes: Uint8Array, signature: number[]): boolean {
  if (bytes.length < signature.length) return false
  return signature.every((byte, index) => bytes[index] === byte)
}

function detectResourceFileFromBytes(bytes: Uint8Array): DetectedResourceFile | null {
  const ascii12 = bytesToAscii(bytes, 0, 12)
  const brand = bytesToAscii(bytes, 4, 12)

  if (startsWithBytes(bytes, [0xff, 0xd8, 0xff])) return SUPPORTED_RESOURCE_FILE_TYPES.image[0]
  if (startsWithBytes(bytes, [0x89, 0x50, 0x4e, 0x47])) return SUPPORTED_RESOURCE_FILE_TYPES.image[1]
  if (ascii12.startsWith('GIF87a') || ascii12.startsWith('GIF89a')) return SUPPORTED_RESOURCE_FILE_TYPES.image[2]
  if (ascii12.startsWith('RIFF') && bytesToAscii(bytes, 8, 12) === 'WEBP') return SUPPORTED_RESOURCE_FILE_TYPES.image[3]
  if (brand.includes('avif') || brand.includes('avis')) return SUPPORTED_RESOURCE_FILE_TYPES.image[4]

  if (ascii12.startsWith('ID3') || (bytes[0] === 0xff && (bytes[1] & 0xe0) === 0xe0)) return SUPPORTED_RESOURCE_FILE_TYPES.audio[0]
  if (ascii12.startsWith('RIFF') && bytesToAscii(bytes, 8, 12) === 'WAVE') return SUPPORTED_RESOURCE_FILE_TYPES.audio[1]
  if (ascii12.startsWith('OggS')) return SUPPORTED_RESOURCE_FILE_TYPES.audio[2]
  if (ascii12.startsWith('fLaC')) return SUPPORTED_RESOURCE_FILE_TYPES.audio[3]
  if (brand.includes('M4A') || brand.includes('mp42')) return SUPPORTED_RESOURCE_FILE_TYPES.audio[4]

  if (brand.includes('mp4') || brand.includes('isom')) return SUPPORTED_RESOURCE_FILE_TYPES.video[0]
  if (brand.includes('qt  ')) return SUPPORTED_RESOURCE_FILE_TYPES.video[1]
  if (startsWithBytes(bytes, [0x1a, 0x45, 0xdf, 0xa3])) return SUPPORTED_RESOURCE_FILE_TYPES.video[2]
  if (ascii12.startsWith('RIFF') && bytesToAscii(bytes, 8, 12) === 'AVI ') return SUPPORTED_RESOURCE_FILE_TYPES.video[3]

  if (ascii12.startsWith('%PDF')) return SUPPORTED_RESOURCE_FILE_TYPES.document[0]
  if (startsWithBytes(bytes, [0xd0, 0xcf, 0x11, 0xe0])) return SUPPORTED_RESOURCE_FILE_TYPES.document[1]
  if (startsWithBytes(bytes, [0x50, 0x4b, 0x03, 0x04])) return SUPPORTED_RESOURCE_FILE_TYPES.archive[0]
  if (startsWithBytes(bytes, [0x52, 0x61, 0x72, 0x21, 0x1a, 0x07])) return SUPPORTED_RESOURCE_FILE_TYPES.archive[1]
  if (startsWithBytes(bytes, [0x37, 0x7a, 0xbc, 0xaf, 0x27, 0x1c])) return SUPPORTED_RESOURCE_FILE_TYPES.archive[2]
  if (startsWithBytes(bytes, [0x1f, 0x8b])) return SUPPORTED_RESOURCE_FILE_TYPES.archive[3]
  if (bytes.length > 0 && bytes.every((byte) => byte === 9 || byte === 10 || byte === 13 || (byte >= 32 && byte <= 126))) {
    return SUPPORTED_RESOURCE_FILE_TYPES.document[2]
  }

  return null
}

async function detectResourceFile(file: File): Promise<DetectedResourceFile | null> {
  const buffer = await file.slice(0, 512).arrayBuffer()
  return detectResourceFileFromBytes(new Uint8Array(buffer))
}

export function ResourceFilesPage() {
  const { t } = useI18n()
  const queryClient = useQueryClient()
  const [messageApi, contextHolder] = message.useMessage()
  const [modalApi, modalContextHolder] = Modal.useModal()
  const [filterForm] = Form.useForm<{ keyword?: string }>()
  const [editorForm] = Form.useForm<ResourceFileFormValues>()

  const [selectedType, setSelectedType] = useState<ResourceFileFilter>('all')
  const [existsFilter, setExistsFilter] = useState<ResourceFileExistsFilter>('all')
  const [keyword, setKeyword] = useState('')
  const [pageNo, setPageNo] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [queryVersion, setQueryVersion] = useState(0)
  const [selectedRowKeys, setSelectedRowKeys] = useState<Key[]>([])
  const [editorOpen, setEditorOpen] = useState(false)
  const [editingRecord, setEditingRecord] = useState<ResourceFileRecord | null>(null)
  const [previewRecord, setPreviewRecord] = useState<ResourceFileRecord | null>(null)
  const [uploadFileList, setUploadFileList] = useState<UploadFile[]>([])
  const [detectedUpload, setDetectedUpload] = useState<DetectedResourceFile | null>(null)
  const [checkingFiles, setCheckingFiles] = useState(false)
  const systemTimezone = useUiSettingsStore((state) => state.systemTimezone)
  const serverTimezone = useUiSettingsStore((state) => state.serverTimezone)
  const watchedEditorType = Form.useWatch('type', editorForm) as ResourceFileType | undefined
  const forcedDownloadType = watchedEditorType === 'document' || watchedEditorType === 'archive'
  const displayedTimezone = normalizeTimezone(systemTimezone)
  const sourceTimezone = normalizeTimezone(serverTimezone)

  const resourceFilesQuery = useQuery({
    queryKey: ['resource-files', selectedType, keyword, existsFilter, pageNo, pageSize, queryVersion],
    queryFn: () =>
      getResourceFiles(
        pageNo,
        pageSize,
        selectedType === 'all' ? undefined : selectedType,
        keyword,
        existsFilter === 'all' ? undefined : existsFilter === 'exists',
      ),
  })
  const records = resourceFilesQuery.data?.items ?? EMPTY_RESOURCE_FILES
  const totalRecords = resourceFilesQuery.data?.total ?? 0
  const totalSizeBytes = resourceFilesQuery.data?.totalSizeBytes ?? 0

  const filteredRecords = useMemo(() => records, [records])

  const refreshFiles = async () => {
    await queryClient.invalidateQueries({ queryKey: ['resource-files'] })
  }

  const displayFileUrl = (url: string, isDownload = false) => {
    if (!url) return ''
    if (url.startsWith('http') || url.startsWith('blob:') || url.startsWith('data:')) {
      return url
    }
    const params = new URLSearchParams({ file_url: url })
    if (isDownload) {
      params.set('is_download', 'true')
    }
    return `/v1/assets/GetFile?${params.toString()}`
  }

  const copyableFileUrl = (url: string, isDownload = false) => {
    const fileUrl = displayFileUrl(url, isDownload)
    if (!fileUrl || fileUrl.startsWith('http') || fileUrl.startsWith('blob:') || fileUrl.startsWith('data:')) {
      return fileUrl
    }
    return new URL(fileUrl, window.location.origin).toString()
  }

  const openEditor = (record?: ResourceFileRecord, nextType?: ResourceFileType) => {
    const targetType = record?.type ?? nextType ?? (selectedType === 'all' ? 'image' : selectedType)
    setEditingRecord(record ?? null)
    setUploadFileList([])
    setDetectedUpload(null)
    editorForm.resetFields()
    editorForm.setFieldsValue(
      record
        ? {
            type: record.type,
            name: record.name,
            requireAuth: record.requireAuth,
            accessMode: record.accessMode,
            remark: record.remark,
          }
        : {
            type: targetType,
            name: '',
            requireAuth: false,
            accessMode: targetType === 'document' || targetType === 'archive' ? 'download' : 'preview',
            remark: '',
          },
    )
    setEditorOpen(true)
  }

  const fileUploadProps: UploadProps = {
    fileList: uploadFileList,
    multiple: false,
    maxCount: 1,
    beforeUpload: async (file) => {
      const detected = await detectResourceFile(file)
      if (!detected) {
        setDetectedUpload(null)
        setUploadFileList([])
        void messageApi.error(t('不支持的文件类型'))
        return Upload.LIST_IGNORE
      }

      const limitMB = RESOURCE_FILE_TYPE_META[detected.type].limitMB
      const sizeMB = file.size / 1024 / 1024
      if (sizeMB > limitMB) {
        void messageApi.error(t('当前类型最大支持 {size}M', { size: limitMB }))
        return Upload.LIST_IGNORE
      }

      setUploadFileList([
        {
          uid: file.uid,
          name: file.name,
          status: 'done',
          size: file.size,
          type: file.type,
          originFileObj: file,
        },
      ])
      editorForm.setFieldsValue({
        name: file.name,
      } as Partial<ResourceFileFormValues>)
      setDetectedUpload(detected)
      if (editorForm.getFieldValue('type') !== detected.type) {
        void messageApi.error(t('文件真实类型与选择类型不一致'))
      }
      if (detected.type === 'document' || detected.type === 'archive') {
        editorForm.setFieldsValue({ accessMode: 'download' } as Partial<ResourceFileFormValues>)
      }
      return false
    },
    onRemove: () => {
      setUploadFileList([])
      setDetectedUpload(null)
      return true
    },
  }

  const saveMutation = useMutation({
    mutationFn: async (values: ResourceFileFormValues) => {
      const uploadFile = uploadFileList[0]?.originFileObj
      if (editingRecord) {
        await updateResourceFile(
          editingRecord.id,
          values.type,
          values.name.trim(),
          values.remark?.trim() || '',
          Boolean(values.requireAuth),
          values.accessMode,
        )
        return
      }
      if (!(uploadFile instanceof File)) {
        throw new Error(t('请上传文件'))
      }
      await uploadResourceFile(
        uploadFile,
        values.type,
        values.name.trim(),
        values.remark?.trim() || '',
        Boolean(values.requireAuth),
        values.accessMode,
      )
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['resource-files'] })
      await resourceFilesQuery.refetch()
      void messageApi.success(editingRecord ? t('编辑成功') : t('新增成功'))
      setEditorOpen(false)
      setEditingRecord(null)
      setUploadFileList([])
      setDetectedUpload(null)
    },
  })

  const deleteRecord = (record: ResourceFileRecord) => {
    modalApi.confirm({
      title: t('确认删除记录 {id}？', { id: record.id }),
      content: t('删除后可重新上传同名文件。'),
      okButtonProps: { danger: true },
      onOk: async () => {
        await deleteResourceFile(record.id)
        setSelectedRowKeys((current) => current.filter((key) => key !== record.id))
        void refreshFiles()
        void messageApi.success(t('删除成功'))
      },
    })
  }

  const batchDelete = () => {
    if (selectedRowKeys.length === 0) {
      void messageApi.warning(t('请先勾选需要删除的数据'))
      return
    }

    modalApi.confirm({
      title: t('确认批量删除 {count} 条记录？', { count: selectedRowKeys.length }),
      content: t('删除后可重新上传同名文件。'),
      okButtonProps: { danger: true },
      onOk: async () => {
        await Promise.all(selectedRowKeys.map((key) => deleteResourceFile(Number(key))))
        setSelectedRowKeys([])
        void refreshFiles()
        void messageApi.success(t('删除成功'))
      },
    })
  }

  const openResource = async (record: ResourceFileRecord) => {
    const access = await accessResourceFile(record.id)
    await refreshFiles()
    window.open(displayFileUrl(access.url || record.url, record.accessMode === 'download'), '_blank', 'noopener,noreferrer')
    await messageApi.success(t('访问成功'))
  }

  const copyResourceUrl = async (record: ResourceFileRecord) => {
    const url = copyableFileUrl(record.url)
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(url)
    } else {
      const textarea = document.createElement('textarea')
      textarea.value = url
      textarea.style.position = 'fixed'
      textarea.style.left = '-9999px'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }
    await messageApi.success(t('已复制地址'))
  }

  const batchCheckFiles = () => {
    modalApi.confirm({
      title: t('确认检测文件有效性？'),
      content: t('文件数量较多时检测会耗时，检测期间页面将暂时不可操作。'),
      onOk: async () => {
        setCheckingFiles(true)
        try {
          const result = await checkResourceFiles(selectedType === 'all' ? undefined : selectedType)
          void queryClient.invalidateQueries({ queryKey: ['resource-files'] })
          void resourceFilesQuery.refetch()
          void messageApi.success(
            t('检测完成：共 {total} 个，存在 {exists} 个，不存在 {missing} 个', {
              total: result.checkedCount,
              exists: result.existsCount,
              missing: result.missingCount,
            }),
          )
        } finally {
          setCheckingFiles(false)
        }
      },
    })
  }

  const typeOptions = useMemo(
    () => [
      { value: 'all', label: t('全部') },
      ...(Object.keys(RESOURCE_FILE_TYPE_META) as ResourceFileType[]).map((type) => ({
        value: type,
        label: t(fileTypeLabel(type)),
      })),
    ],
    [t],
  )
  const existsOptions = useMemo(
    () => [
      { value: 'all', label: t('全部') },
      { value: 'exists', label: t('存在') },
      { value: 'missing', label: t('不存在') },
    ],
    [t],
  )
  const querySummary = useMemo(() => {
    const parts = [
      `${t('类型')}：${selectedType === 'all' ? t('全部') : t(fileTypeLabel(selectedType))}`,
      `${t('文件有效性')}：${
        existsFilter === 'all'
          ? t('全部')
          : existsFilter === 'exists'
            ? t('存在')
            : t('不存在')
      }`,
      `${t('关键字')}：${keyword || t('全部')}`,
      `${t('文件总大小')}：${formatFileSize(totalSizeBytes)}`,
      `${t('分页')}：${pageNo} / ${pageSize}`,
    ]
    return parts.join('，')
  }, [existsFilter, keyword, pageNo, pageSize, selectedType, t, totalSizeBytes])

  const renderUploadedAt = (value: string) => {
    if (!value) return '-'
    const shouldShowTimezoneHint = displayedTimezone !== sourceTimezone
    const formattedValue = formatDateTime(value, displayedTimezone, sourceTimezone)
    return (
      <Space size={4} align="center" style={{ whiteSpace: 'nowrap' }}>
        <Typography.Text>{formattedValue}</Typography.Text>
        {shouldShowTimezoneHint ? (
          <Tooltip title={t('根据设置，当前使用的是 {timezone} 时区', { timezone: displayedTimezone })}>
            <ClockCircleOutlined style={{ color: 'var(--color-text-secondary)', fontSize: 14 }} />
          </Tooltip>
        ) : null}
      </Space>
    )
  }

  return (
    <>
      {contextHolder}
      {modalContextHolder}
      {checkingFiles ? (
        <div className="resource-files-blocking-mask">
          <Spin size="large" tip={t('正在检测文件有效性...')} />
        </div>
      ) : null}

      <Space direction="vertical" size={16} className="full-width resource-files-page table-scroll-page">
      <Card>
        <Space direction="vertical" size={14} className="full-width">
          <Typography.Text type="secondary">
            {t('通过服务器本地上传和管理各类文件，暂未支持三方存储。')}
          </Typography.Text>
          <Space wrap style={{ justifyContent: 'space-between', width: '100%' }}>
            <Segmented
              value={selectedType}
              options={typeOptions}
              onChange={(value) => {
                setSelectedType(value as ResourceFileFilter)
                setPageNo(1)
                setSelectedRowKeys([])
                setQueryVersion((current) => current + 1)
              }}
            />
            <Space wrap>
              <UiButton
                type="primary"
                icon={<UploadOutlined />}
                disabled={checkingFiles}
                onClick={() =>
                  openEditor(undefined, selectedType === 'all' ? 'image' : selectedType)
                }
              >
                {t('上传文件')}
              </UiButton>
              <UiButton
                danger
                disabled={checkingFiles || selectedRowKeys.length === 0}
                onClick={batchDelete}
                icon={<DeleteOutlined />}
              >
                {t('批量删除')}
              </UiButton>
              <UiButton
                disabled={checkingFiles}
                onClick={batchCheckFiles}
              >
                {t('批量检测文件有效性')}
              </UiButton>
              <UiButton
                icon={<ReloadOutlined />}
                disabled={checkingFiles}
                onClick={() => {
                  setSelectedType('all')
                  setKeyword('')
                  setPageNo(1)
                  filterForm.resetFields()
                  setSelectedRowKeys([])
                }}
              >
                {t('重置')}
              </UiButton>
            </Space>
          </Space>

          <Form
            form={filterForm}
            layout="inline"
            onFinish={(values: { keyword?: string }) => {
              setKeyword(values.keyword?.trim() || '')
              setPageNo(1)
              setQueryVersion((current) => current + 1)
            }}
            onKeyDown={(event) => {
              if (event.key === 'Enter') {
                event.preventDefault()
                filterForm.submit()
              }
            }}
          >
            <Form.Item label={t('关键字')} name="keyword">
              <Input
                placeholder={t('文件名 / 地址 / 备注')}
                allowClear
                onPressEnter={() => filterForm.submit()}
              />
            </Form.Item>
            <Form.Item label={t('文件有效性')}>
              <Select
                value={existsFilter}
                options={existsOptions}
                style={{ width: 120 }}
                onChange={(value) => {
                  setExistsFilter(value)
                  setPageNo(1)
                }}
              />
            </Form.Item>
            <Form.Item>
              <Space size={8}>
                <UiButton type="primary" htmlType="submit">
                  {t('查询')}
                </UiButton>
                <UiButton
                  onClick={() => {
                    filterForm.resetFields()
                    setKeyword('')
                    setExistsFilter('all')
                    setPageNo(1)
                    setQueryVersion((current) => current + 1)
                  }}
                >
                  {t('重置')}
                </UiButton>
              </Space>
            </Form.Item>
          </Form>
        </Space>
      </Card>

      <Card title={t('文件列表')} className="resource-files-list-card resource-files-list-card--scroll-only-body system-table-card table-scroll-region">
        <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 12 }}>
          {t('本次查询')}
          {'：'}
          {querySummary}
        </Typography.Text>
        <Table<ResourceFileRecord>
          rowKey="id"
          dataSource={filteredRecords}
          loading={resourceFilesQuery.isFetching}
          rowSelection={{
            selectedRowKeys,
            onChange: (keys) => setSelectedRowKeys(keys),
          }}
          pagination={{
            current: pageNo,
            pageSize,
            total: totalRecords,
            showSizeChanger: true,
            showTotal: (total) => t('共 {total} 条', { total }),
            onChange: (nextPage, nextPageSize) => {
              setPageNo(nextPage)
              setPageSize(nextPageSize)
              setSelectedRowKeys([])
            },
          }}
          scroll={{ x: 'max-content', y: RESOURCE_FILES_TABLE_SCROLL_Y }}
          columns={[
            {
              title: 'ID',
              dataIndex: 'id',
              width: 90,
            },
            {
              title: t('文件'),
              dataIndex: 'name',
              width: 240,
              render: (_, record) => {
                const Icon = RESOURCE_FILE_TYPE_META[record.type].icon
                return (
                  <Space align="start" size={10}>
                    <Icon style={{ fontSize: 18, marginTop: 3 }} />
                    <Space direction="vertical" size={2}>
                      <Tooltip title={record.name}>
                        <Typography.Text strong ellipsis style={{ maxWidth: 180 }}>
                          {record.name}
                        </Typography.Text>
                      </Tooltip>
                      <Typography.Text type="secondary">{record.remark}</Typography.Text>
                    </Space>
                  </Space>
                )
              },
            },
            {
              title: t('类型'),
              dataIndex: 'type',
              width: 120,
              render: (value: ResourceFileType) => (
                <Tag color={RESOURCE_FILE_TYPE_META[value].color}>{t(fileTypeLabel(value))}</Tag>
              ),
            },
            {
              title: t('大小'),
              dataIndex: 'sizeBytes',
              width: 120,
              render: (value: number) => formatFileSize(value),
            },
            {
              title: t('上传地址'),
              dataIndex: 'url',
              width: 360,
              ellipsis: true,
              render: (value: string, record) => (
                <Space size={6}>
                  <Typography.Text code ellipsis style={{ maxWidth: 260 }}>
                    {displayFileUrl(value)}
                  </Typography.Text>
                  <Tooltip title={t('复制地址')}>
                    <UiButton
                      type="text"
                      icon={<CopyOutlined />}
                      onClick={() => void copyResourceUrl(record)}
                    />
                  </Tooltip>
                  <Tooltip title={t('访问地址')}>
                    <UiButton
                      type="text"
                      icon={<LinkOutlined />}
                      onClick={() => void openResource(record)}
                    />
                  </Tooltip>
                </Space>
              ),
            },
            {
              title: t('上传时间'),
              dataIndex: 'uploadedAt',
              width: 170,
              render: (value: string) => renderUploadedAt(value),
            },
            {
              title: t('是否存在'),
              dataIndex: 'exists',
              width: 110,
              render: (value: boolean | null, record) => {
                const stale = isResourceValidityStale(record.existsCheckedAt)
                const tag =
                  value === null ? (
                    <Tag>{t('未知')}</Tag>
                  ) : (
                    <Tag color={value ? 'green' : 'red'}>{value ? t('存在') : t('不存在')}</Tag>
                  )
                return (
                  <Space size={4}>
                    {tag}
                    {stale ? (
                      <Tooltip title={t('距离上次检测时间超过3天，有效性未知')}>
                        <ExclamationCircleOutlined />
                      </Tooltip>
                    ) : null}
                  </Space>
                )
              },
            },
            {
              title: t('访问方式'),
              dataIndex: 'accessMode',
              width: 120,
              render: (value: ResourceFileAccessMode) => (
                <Tag color={value === 'download' ? 'gold' : 'default'}>
                  {value === 'download' ? t('直接下载') : t('预览模式')}
                </Tag>
              ),
            },
            {
              title: t('访问鉴权'),
              dataIndex: 'requireAuth',
              width: 110,
              render: (value: boolean) => (
                <Tag color={value ? 'blue' : 'default'}>{value ? t('开启') : t('关闭')}</Tag>
              ),
            },
            {
              title: t('上次访问时间'),
              dataIndex: 'lastVisitedAt',
              width: 170,
            },
            {
              title: t('访问次数'),
              dataIndex: 'visitCount',
              width: 110,
            },
            {
              title: t('操作'),
              width: 160,
              render: (_, record) => (
                <Space size={0}>
                  <UiButton type="link" icon={<EditOutlined />} onClick={() => openEditor(record)}>
                    {t('编辑')}
                  </UiButton>
                  <UiButton
                    type="link"
                    icon={<DeleteOutlined />}
                    danger
                    onClick={() => deleteRecord(record)}
                  >
                    {t('删除')}
                  </UiButton>
                </Space>
              ),
            },
          ]}
        />
      </Card>
      </Space>

      <Modal
        title={editingRecord ? t('编辑文件') : t('上传文件')}
        open={editorOpen}
        onCancel={() => {
          setEditorOpen(false)
          setEditingRecord(null)
          setUploadFileList([])
          setDetectedUpload(null)
        }}
        onOk={() => void editorForm.submit()}
        confirmLoading={saveMutation.isPending}
        width={720}
      >
        <Form
          form={editorForm}
          layout="vertical"
          onFinish={(values) => void saveMutation.mutateAsync(values)}
          initialValues={{
            type: selectedType === 'all' ? 'image' : selectedType,
            requireAuth: false,
            accessMode: 'preview',
          }}
        >
          <Row gutter={12}>
            <Col xs={24} md={12}>
              <Form.Item
                label={t('文件类型')}
                name="type"
                rules={[
                  { required: true, message: t('请选择文件类型') },
                  () => ({
                    validator: async (_, value) => {
                      if (uploadFileList.length === 0 || !detectedUpload) {
                        return
                      }
                      if (value !== detectedUpload.type) {
                        throw new Error(t('文件真实类型与选择类型不一致'))
                      }
                    },
                  }),
                ]}
              >
                <Select
                  options={(Object.keys(RESOURCE_FILE_TYPE_META) as ResourceFileType[]).map(
                    (type) => ({
                      value: type,
                      label: `${t(fileTypeLabel(type))}（${RESOURCE_FILE_TYPE_META[type].limitMB}M）`,
                    }),
                  )}
                  onChange={(value: ResourceFileType) => {
                    if (value === 'document' || value === 'archive') {
                      editorForm.setFieldsValue({ accessMode: 'download' })
                    }
                  }}
                />
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item
                label={t('文件名称')}
                name="name"
                rules={[{ required: true, message: t('请输入文件名称') }]}
              >
                <Input placeholder={t('请输入文件名称')} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={12}>
            <Col xs={24} md={12}>
              <Form.Item
                label={t('是否访问鉴权')}
                tooltip={{
                  title: t('开启后访问文件需要登录态；关闭后任何拿到地址的人都可以访问。'),
                  icon: <QuestionCircleOutlined />,
                }}
                name="requireAuth"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item
                label={t('访问方式')}
                tooltip={{
                  title: t('预览模式会在浏览器新窗口打开文件；直接下载会以附件方式保存到本地。'),
                  icon: <QuestionCircleOutlined />,
                }}
                name="accessMode"
                rules={[{ required: true, message: t('请选择访问方式') }]}
              >
                <Segmented
                  disabled={forcedDownloadType}
                  options={[
                    { value: 'preview', label: t('预览模式') },
                    { value: 'download', label: t('直接下载') },
                  ]}
                />
              </Form.Item>
              {forcedDownloadType ? (
                <Typography.Text type="secondary">
                  {t('文档和压缩包仅支持直接下载')}
                </Typography.Text>
              ) : null}
            </Col>
          </Row>

          <Form.Item label={t('支持类型')}>
            <Popover
              trigger="hover"
              placement="rightTop"
              content={
                <Space direction="vertical" size={8} style={{ maxWidth: 420 }}>
                  {(Object.keys(RESOURCE_FILE_TYPE_META) as ResourceFileType[]).map((type) => (
                    <Space key={type} wrap>
                      <Typography.Text type={watchedEditorType === type ? undefined : 'secondary'}>
                        {t(fileTypeLabel(type))}
                      </Typography.Text>
                      {SUPPORTED_RESOURCE_FILE_TYPES[type].map((item) => (
                        <Tag key={`${type}-${item.extension}`}>{item.label}</Tag>
                      ))}
                    </Space>
                  ))}
                </Space>
              }
            >
              <UiButton>{t('查看支持类型')}</UiButton>
            </Popover>
            {detectedUpload ? (
              <Typography.Text type="secondary" style={{ marginLeft: 12 }}>
                {t('已识别类型')}：{t(fileTypeLabel(detectedUpload.type))} / {detectedUpload.label}
              </Typography.Text>
            ) : null}
          </Form.Item>

          <Form.Item
            label={t('备注')}
            name="remark"
            rules={[{ max: 50, message: t('备注最多50字') }]}
          >
            <Input.TextArea rows={3} maxLength={50} showCount placeholder={t('请输入备注')} />
          </Form.Item>

          <Form.Item
            label={t('上传文件')}
            required={!editingRecord}
            validateStatus={!editingRecord && uploadFileList.length === 0 ? 'error' : undefined}
            help={!editingRecord && uploadFileList.length === 0 ? t('请上传文件') : undefined}
          >
            <Upload.Dragger
              {...fileUploadProps}
              accept={resourceFileTypeAccept(watchedEditorType ?? 'image')}
            >
              <p className="ant-upload-drag-icon">
                <UploadOutlined />
              </p>
              <p className="ant-upload-text">{t('点击或拖拽文件到这里')}</p>
              <p className="ant-upload-hint">
                {t('将读取文件头自动识别真实类型；图片 5M、语音 5M、视频 100M、文档 20M。')}
              </p>
            </Upload.Dragger>
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        title={previewRecord ? t('文件详情') : t('文件详情')}
        open={Boolean(previewRecord)}
        onClose={() => setPreviewRecord(null)}
        width={520}
      >
        {previewRecord ? (
          <Space direction="vertical" size={14} style={{ width: '100%' }}>
            <Card>
              <Space direction="vertical" size={10} style={{ width: '100%' }}>
                <Space align="center" size={10}>
                  <Tag color={RESOURCE_FILE_TYPE_META[previewRecord.type].color}>
                    {t(fileTypeLabel(previewRecord.type))}
                  </Tag>
                  <Typography.Text strong>{previewRecord.name}</Typography.Text>
                </Space>
                <Typography.Text type="secondary">{previewRecord.remark}</Typography.Text>
              </Space>
            </Card>
            <Card title={t('文件信息')} size="small">
              <Space direction="vertical" size={10} style={{ width: '100%' }}>
                <Typography.Text>
                  {t('文件大小')}
                  {'：'}
                  {formatFileSize(previewRecord.sizeBytes)}
                </Typography.Text>
                <Typography.Text>
                  {t('上传时间')}
                  {'：'}
                  {renderUploadedAt(previewRecord.uploadedAt)}
                </Typography.Text>
                <Typography.Text>
                  {t('是否存在')}
                  {'：'}
                  {previewRecord.exists === null
                    ? t('未知')
                    : previewRecord.exists
                      ? t('存在')
                      : t('不存在')}
                  {isResourceValidityStale(previewRecord.existsCheckedAt) ? (
                    <Tooltip title={t('距离上次检测时间超过3天，有效性未知')}>
                      <ExclamationCircleOutlined style={{ marginLeft: 6 }} />
                    </Tooltip>
                  ) : null}
                </Typography.Text>
                <Typography.Text>
                  {t('访问方式')}
                  {'：'}
                  {previewRecord.accessMode === 'download' ? t('直接下载') : t('预览模式')}
                </Typography.Text>
                <Typography.Text>
                  {t('上次访问时间')}
                  {'：'}
                  {previewRecord.lastVisitedAt}
                </Typography.Text>
                <Typography.Text>
                  {t('访问次数')}
                  {'：'}
                  {previewRecord.visitCount}
                </Typography.Text>
                <Space size={8} wrap>
                  <UiButton
                    icon={<CopyOutlined />}
                    onClick={() => void copyResourceUrl(previewRecord)}
                  >
                    {t('复制地址')}
                  </UiButton>
                  <UiButton
                    icon={previewRecord.accessMode === 'download' ? <DownloadOutlined /> : <LinkOutlined />}
                    onClick={() => void openResource(previewRecord)}
                  >
                    {t('访问地址')}
                  </UiButton>
                </Space>
              </Space>
            </Card>
            <Card title={t('上传地址')} size="small">
              <Typography.Paragraph
                copyable={{ text: copyableFileUrl(previewRecord.url) }}
                style={{ marginBottom: 0 }}
              >
                {displayFileUrl(previewRecord.url)}
              </Typography.Paragraph>
            </Card>
          </Space>
        ) : null}
      </Drawer>
    </>
  )
}
