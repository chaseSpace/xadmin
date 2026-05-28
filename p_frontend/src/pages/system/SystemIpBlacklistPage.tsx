import { useMutation, useQuery } from '@tanstack/react-query'
import { Card, DatePicker, Form, Input, Modal, Select, Space, Table, Tag, Typography, message } from 'antd'
import type { TablePaginationConfig } from 'antd'
import type { SorterResult } from 'antd/es/table/interface'
import { useMemo, useState } from 'react'
import type { Key } from 'react'
import { UiButton } from '../../components/ui'
import { useI18n } from '../../i18n/messages'
import {
  batchUnblockSystemIPBlacklist,
  createSystemIPBlacklist,
  deleteSystemIPBlacklist,
  getSystemIPBlacklist,
  getSystemIPBlacklistCreators,
  importSystemIPBlacklist,
  unblockSystemIPBlacklist,
  type SystemIPBlacklistFilters,
  type SystemIPBlacklistItem,
} from '../../services/api/system'
import { useUiSettingsStore } from '../../store/uiSettings'
import { formatDateTime } from '../../utils/timezone'

type FilterFormValues = {
  keyword?: string
  status?: 'active' | 'inactive'
  creator?: string
}

type DateTimePickerValue = { format: (template: string) => string; valueOf: () => number }
type DateTimeFormValue = string | DateTimePickerValue

type EditorFormValues = {
  ip: string
  durationHours?: number | 'custom'
  customEndAt?: DateTimeFormValue
  reason: string
}

type ImportFormValues = {
  ips: string
  durationHours?: number | 'custom'
  customEndAt?: DateTimeFormValue
}

const TEMP_BAN_DURATION_OPTIONS: Array<{ value: number | 'custom'; label: string }> = [
  { value: 1, label: '1 小时' },
  { value: 6, label: '6 小时' },
  { value: 24, label: '24 小时' },
  { value: 72, label: '72 小时' },
  { value: 168, label: '7 天' },
  { value: 720, label: '30 天' },
  { value: 'custom', label: '自定义时间' },
]

function formatDateTimeString(date: Date): string {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  const hh = String(date.getHours()).padStart(2, '0')
  const mm = String(date.getMinutes()).padStart(2, '0')
  const ss = String(date.getSeconds()).padStart(2, '0')
  return `${y}-${m}-${d} ${hh}:${mm}:${ss}`
}

function numberRange(start: number, end: number): number[] {
  return Array.from({ length: Math.max(end - start, 0) }, (_unused, index) => start + index)
}

function todayDateString(): string {
  return formatDateTimeString(new Date()).slice(0, 10)
}

function addHoursFromNow(hours: number): string {
  return formatDateTimeString(new Date(Date.now() + hours * 60 * 60 * 1000))
}

function normalizeDateTimeLocal(value?: DateTimeFormValue): string {
  if (value && typeof value !== 'string') {
    return value.format('YYYY-MM-DD HH:mm:ss')
  }
  const input = (value || '').trim()
  if (!input) return ''
  const normalized = input.replace('T', ' ')
  return normalized.length === 16 ? `${normalized}:00` : normalized
}

function isFutureDateTime(value?: DateTimeFormValue): boolean {
  if (value && typeof value !== 'string') {
    return Number.isFinite(value.valueOf()) && value.valueOf() > Date.now()
  }
  const normalized = normalizeDateTimeLocal(value)
  if (!normalized) return false
  const ms = Date.parse(normalized.replace(' ', 'T'))
  return Number.isFinite(ms) && ms > Date.now()
}

function disabledPastDate(value?: DateTimePickerValue | null): boolean {
  return Boolean(value && value.format('YYYY-MM-DD') < todayDateString())
}

function disabledPastTime(value?: DateTimePickerValue | null) {
  if (!value || value.format('YYYY-MM-DD') !== todayDateString()) {
    return {}
  }
  const now = new Date()
  const currentHour = now.getHours()
  const currentMinute = now.getMinutes()
  const currentSecond = now.getSeconds()
  return {
    disabledHours: () => numberRange(0, currentHour),
    disabledMinutes: (selectedHour: number) => {
      if (selectedHour < currentHour) return numberRange(0, 60)
      if (selectedHour === currentHour) return numberRange(0, currentMinute)
      return []
    },
    disabledSeconds: (selectedHour: number, selectedMinute: number) => {
      if (selectedHour < currentHour || (selectedHour === currentHour && selectedMinute < currentMinute)) {
        return numberRange(0, 60)
      }
      if (selectedHour === currentHour && selectedMinute === currentMinute) {
        return numberRange(0, currentSecond + 1)
      }
      return []
    },
  }
}

const IPV4_SEGMENT = '(25[0-5]|2[0-4]\\d|1\\d\\d|[1-9]?\\d)'
const IPV4_REGEX = new RegExp(`^${IPV4_SEGMENT}\\.${IPV4_SEGMENT}\\.${IPV4_SEGMENT}\\.${IPV4_SEGMENT}$`)

function isValidIPAddress(value?: string): boolean {
  const input = (value || '').trim()
  if (!input || input.includes('/')) return false
  if (IPV4_REGEX.test(input)) return true
  if (!input.includes(':')) return false
  try {
    const parsed = new URL(`http://[${input}]/`)
    return parsed.hostname.toLowerCase() === `[${input.toLowerCase()}]`
  } catch {
    return false
  }
}

function parseImportIPs(value?: string): string[] {
  return (value || '')
    .split('\n')
    .map((item) => item.trim())
    .filter(Boolean)
}

function resolveBlacklistStatusKind(row: SystemIPBlacklistItem, nowMs: number): SystemIPBlacklistItem['status'] {
  if (row.status === 'manual_inactive') return 'manual_inactive'
  if (row.status === 'expired') return 'expired'
  const endAtMs = Date.parse(row.endAt.replace(' ', 'T'))
  if (Number.isFinite(endAtMs) && endAtMs < nowMs) return 'expired'
  return 'active'
}

function formatDurationSpan(startAt: string, endAt: string): string {
  const startMs = Date.parse(startAt.replace(' ', 'T'))
  const endMs = Date.parse(endAt.replace(' ', 'T'))
  if (!Number.isFinite(startMs) || !Number.isFinite(endMs) || endMs <= startMs) {
    return '0小时'
  }
  let remainingHours = Math.max(1, Math.ceil((endMs - startMs) / (60 * 60 * 1000)))
  const units = [
    { hours: 24 * 365, label: '年' },
    { hours: 24 * 30, label: '月' },
    { hours: 24, label: '日' },
    { hours: 1, label: '时' },
  ]
  const parts: string[] = []
  for (const unit of units) {
    if (parts.length >= 2) break
    const value = Math.floor(remainingHours / unit.hours)
    if (value <= 0) continue
    parts.push(`${value}${unit.label}`)
    remainingHours -= value * unit.hours
  }
  return parts.length > 0 ? parts.join('') : '1时'
}

export function SystemIpBlacklistPage() {
  const { t } = useI18n()
  const systemTimezone = useUiSettingsStore((state) => state.systemTimezone)
  const [messageApi, contextHolder] = message.useMessage()
  const [modalApi, modalContextHolder] = Modal.useModal()
  const [filterForm] = Form.useForm<FilterFormValues>()
  const [editorForm] = Form.useForm<EditorFormValues>()
  const [importForm] = Form.useForm<ImportFormValues>()

  const [filters, setFilters] = useState<SystemIPBlacklistFilters>({})
  const [searchNonce, setSearchNonce] = useState(0)
  const [pageNo, setPageNo] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [orderField, setOrderField] = useState<'id' | 'ip' | 'ban_type' | 'status' | 'hit_count' | 'updated_at' | undefined>()
  const [orderType, setOrderType] = useState<'asc' | 'desc' | undefined>()
  const [selectedRowKeys, setSelectedRowKeys] = useState<Key[]>([])
  const [editorOpen, setEditorOpen] = useState(false)
  const [importOpen, setImportOpen] = useState(false)
  const [renderNowMs] = useState(() => Date.now())

  const listQuery = useQuery({
    queryKey: ['system-ip-blacklist', pageNo, pageSize, orderField, orderType, filters, searchNonce],
    queryFn: () => getSystemIPBlacklist(pageNo, pageSize, orderField, orderType, filters),
  })

  const creatorsQuery = useQuery({
    queryKey: ['system-ip-blacklist-creators'],
    queryFn: getSystemIPBlacklistCreators,
  })

  const listData = listQuery.data?.items ?? []

  const creatorOptions = useMemo(
    () =>
      (creatorsQuery.data || []).map((creator) => ({
        value: creator,
        label: creator,
      })),
    [creatorsQuery.data],
  )

  const saveMutation = useMutation({
    mutationFn: async (values: EditorFormValues) => {
      const durationHours = typeof values.durationHours === 'number' ? values.durationHours : undefined
      const customEndAt = values.durationHours === 'custom' ? normalizeDateTimeLocal(values.customEndAt) : undefined
      await createSystemIPBlacklist({
        ip: values.ip.trim(),
        endAt: customEndAt || (durationHours ? addHoursFromNow(durationHours) : ''),
        reason: values.reason.trim(),
      })
    },
    onSuccess: async () => {
      void messageApi.success(t('新增成功'))
      setEditorOpen(false)
      await Promise.all([listQuery.refetch(), creatorsQuery.refetch()])
    },
  })

  const unblockMutation = useMutation({
    mutationFn: async (id: number) => unblockSystemIPBlacklist(id),
    onSuccess: async () => {
      void messageApi.success(t('IP已主动失效'))
      await Promise.all([listQuery.refetch(), creatorsQuery.refetch()])
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => deleteSystemIPBlacklist(id),
    onSuccess: async () => {
      void messageApi.success(t('删除成功'))
      await Promise.all([listQuery.refetch(), creatorsQuery.refetch()])
    },
  })

  const batchUnblockMutation = useMutation({
    mutationFn: async (ids: number[]) => batchUnblockSystemIPBlacklist(ids),
    onSuccess: async () => {
      setSelectedRowKeys([])
      void messageApi.success(t('批量主动失效成功'))
      await Promise.all([listQuery.refetch(), creatorsQuery.refetch()])
    },
  })

  const importMutation = useMutation({
    mutationFn: async (values: ImportFormValues) => {
      const ips = parseImportIPs(values.ips)
      const durationHours = typeof values.durationHours === 'number' ? values.durationHours : undefined
      const customEndAt = values.durationHours === 'custom' ? normalizeDateTimeLocal(values.customEndAt) : undefined
      await importSystemIPBlacklist({
        ips,
        durationHours,
        endAt: customEndAt,
      })
    },
    onSuccess: async () => {
      setImportOpen(false)
      importForm.resetFields()
      void messageApi.success(t('导入成功'))
      await Promise.all([listQuery.refetch(), creatorsQuery.refetch()])
    },
  })

  const applyFilters = (values: FilterFormValues) => {
    setFilters({
      keyword: values.keyword?.trim() || undefined,
      status: values.status,
      creator: values.creator,
    })
    setPageNo(1)
    setSearchNonce((value) => value + 1)
  }

  const editorDurationMode = Form.useWatch('durationHours', editorForm)
  const importDurationMode = Form.useWatch('durationHours', importForm)

  const getRowStatus = (row: SystemIPBlacklistItem): { label: string; color: 'green' | 'orange' | 'default'; canUnblock: boolean } => {
    const status = resolveBlacklistStatusKind(row, renderNowMs)
    if (status === 'expired') {
      return { label: t('到期失效'), color: 'orange', canUnblock: false }
    }
    if (status === 'manual_inactive') {
      return { label: t('主动失效'), color: 'default', canUnblock: false }
    }
    return { label: t('生效'), color: 'green', canUnblock: true }
  }

  return (
    <>
      {contextHolder}
      {modalContextHolder}

      <Space direction="vertical" size={16} className="full-width system-ip-blacklist-page table-scroll-page">
        <Space wrap>
          <UiButton
            type="primary"
            onClick={() => {
              editorForm.resetFields()
              editorForm.setFieldsValue({ durationHours: 24 })
              setEditorOpen(true)
            }}
          >
            {t('新增黑名单IP')}
          </UiButton>
          <UiButton
            loading={batchUnblockMutation.isPending}
            onClick={() => {
              if (selectedRowKeys.length === 0) {
                void messageApi.warning(t('请先选择需要主动失效的数据'))
                return
              }
              void modalApi.confirm({
                title: t('确认批量主动失效 {count} 条记录？', { count: selectedRowKeys.length }),
                onOk: async () => {
                  const ids = selectedRowKeys.map((key) => Number(key)).filter((value) => Number.isFinite(value) && value > 0)
                  await batchUnblockMutation.mutateAsync(ids)
                },
              })
            }}
          >
            {t('批量主动失效')}
          </UiButton>
          <UiButton
            onClick={() => {
              importForm.resetFields()
              importForm.setFieldsValue({ durationHours: 24 })
              setImportOpen(true)
            }}
          >
            {t('导入黑名单')}
          </UiButton>
        </Space>

        <Card>
          <Form
            form={filterForm}
            layout="inline"
            style={{ rowGap: 12 }}
            onFinish={applyFilters}
            onKeyDown={(event) => {
              if (event.key === 'Enter') {
                event.preventDefault()
                filterForm.submit()
              }
            }}
          >
            <Form.Item label={t('关键字')} name="keyword">
              <Input placeholder={t('请输入IP地址')} allowClear onPressEnter={() => filterForm.submit()} />
            </Form.Item>
            <Form.Item label={t('状态')} name="status">
              <Select
                style={{ width: 140 }}
                allowClear
                options={[
                  { value: 'active', label: t('生效') },
                  { value: 'inactive', label: t('失效') },
                ]}
              />
            </Form.Item>
            <Form.Item label={t('创建人')} name="creator">
              <Select style={{ width: 180 }} allowClear options={creatorOptions} />
            </Form.Item>
            <Form.Item>
              <Space size={8}>
                <UiButton type="primary" onClick={() => filterForm.submit()}>
                  {t('查询')}
                </UiButton>
                <UiButton
                  onClick={() => {
                    filterForm.resetFields()
                    setFilters({})
                    setPageNo(1)
                    setSearchNonce((value) => value + 1)
                  }}
                >
                  {t('重置')}
                </UiButton>
              </Space>
            </Form.Item>
          </Form>
        </Card>

        <Card title={t('封禁记录')} className="compact-table-card system-table-card table-scroll-region">
          <Table<SystemIPBlacklistItem>
            rowKey="id"
            loading={listQuery.isLoading || listQuery.isFetching}
            rowSelection={{
              selectedRowKeys,
              onChange: (keys) => setSelectedRowKeys(keys),
              getCheckboxProps: (row) => ({ disabled: !getRowStatus(row).canUnblock }),
            }}
            dataSource={listData}
            scroll={{ x: 1680, y: 392 }}
            pagination={{
              current: pageNo,
              pageSize,
              total: listQuery.data?.total ?? 0,
              showSizeChanger: true,
              showTotal: (total) => t('共 {total} 条', { total }),
            }}
            onChange={(
              pagination: TablePaginationConfig,
              _tableFilters,
              sorter: SorterResult<SystemIPBlacklistItem> | SorterResult<SystemIPBlacklistItem>[],
            ) => {
              if (pagination.current) setPageNo(pagination.current)
              if (pagination.pageSize && pagination.pageSize !== pageSize) {
                setPageSize(pagination.pageSize)
                setPageNo(1)
              }
              const single = Array.isArray(sorter) ? sorter[0] : sorter
              if (!single || !single.field || !single.order) {
                setOrderField(undefined)
                setOrderType(undefined)
                return
              }
              const mapper: Record<string, 'id' | 'ip' | 'ban_type' | 'status' | 'hit_count' | 'updated_at'> = {
                id: 'id',
                ip: 'ip',
                banType: 'ban_type',
                status: 'status',
                hitCount: 'hit_count',
                updatedAt: 'updated_at',
              }
              const field = mapper[String(single.field)]
              if (!field) {
                setOrderField(undefined)
                setOrderType(undefined)
                return
              }
              setOrderField(field)
              setOrderType(single.order === 'ascend' ? 'asc' : 'desc')
            }}
            columns={[
              { title: t('记录ID'), dataIndex: 'id', sorter: true, width: 88 },
              { title: t('创建人'), dataIndex: 'creator', width: 120 },
              { title: t('IP地址'), dataIndex: 'ip', sorter: true, width: 140 },
              { title: t('封禁时长'), width: 112, render: (_, row) => formatDurationSpan(row.startAt, row.endAt) },
              { title: t('封禁开始'), dataIndex: 'startAt', width: 156, render: (value: string) => formatDateTime(value, systemTimezone) },
              { title: t('封禁结束'), dataIndex: 'endAt', width: 156, render: (value: string) => formatDateTime(value, systemTimezone) },
              {
                title: t('状态'),
                dataIndex: 'status',
                sorter: true,
                width: 96,
                fixed: 'left',
                render: (_value: SystemIPBlacklistItem['status'], row) => {
                  const status = getRowStatus(row)
                  return <Tag color={status.color}>{status.label}</Tag>
                },
              },
              { title: t('封禁原因'), dataIndex: 'reason', width: 180, ellipsis: true },
              { title: t('命中次数'), dataIndex: 'hitCount', sorter: true, width: 100 },
              { title: t('更新时间'), dataIndex: 'updatedAt', sorter: true, width: 170, fixed: 'right', render: (value: string) => formatDateTime(value, systemTimezone) },
              {
                title: t('操作'),
                fixed: 'right',
                width: 150,
                render: (_, row) => (
                  <Space size={0}>
                    {getRowStatus(row).canUnblock ? (
                      <UiButton
                        type="link"
                        loading={unblockMutation.isPending}
                        onClick={() => {
                          void modalApi.confirm({
                            title: t('确认主动失效IP {ip}？', { ip: row.ip }),
                            onOk: async () => {
                              await unblockMutation.mutateAsync(row.id)
                            },
                          })
                        }}
                      >
                        {t('主动失效')}
                      </UiButton>
                    ) : (
                      <Typography.Text type="secondary">-</Typography.Text>
                    )}
                    <UiButton
                      type="link"
                      danger
                      loading={deleteMutation.isPending}
                      onClick={() => {
                        void modalApi.confirm({
                          title: t('确认删除IP {ip}？', { ip: row.ip }),
                          onOk: async () => {
                            await deleteMutation.mutateAsync(row.id)
                          },
                        })
                      }}
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
        title={t('新增黑名单IP')}
        open={editorOpen}
        onCancel={() => setEditorOpen(false)}
        onOk={() => editorForm.submit()}
        confirmLoading={saveMutation.isPending}
      >
        <Form form={editorForm} layout="vertical" onFinish={(values) => void saveMutation.mutateAsync(values)}>
          <Form.Item
            label={t('IP地址')}
            name="ip"
            rules={[
              { required: true, message: t('请输入IP地址') },
              {
                validator: async (_, value) => {
                  if (isValidIPAddress(value)) return
                  throw new Error(t('请输入合法IP地址'))
                },
              },
            ]}
          >
            <Input placeholder={t('例如 1.2.3.4')} />
          </Form.Item>
          <Form.Item label={t('封禁时长')} name="durationHours" rules={[{ required: true, message: t('请选择封禁时长') }]}>
            <Select
              options={TEMP_BAN_DURATION_OPTIONS.map((item) => ({
                value: item.value,
                label: t(item.label),
              }))}
            />
          </Form.Item>
          {editorDurationMode === 'custom' ? (
            <Form.Item
              label={t('自定义时间点')}
              name="customEndAt"
              rules={[
                { required: true, message: t('请选择自定义时间点') },
                {
                  validator: async (_, value) => {
                    if (isFutureDateTime(value)) return
                    throw new Error(t('自定义时间点必须晚于当前时间'))
                  },
                },
              ]}
            >
              <DatePicker
                showTime={{ hideDisabledOptions: true }}
                disabledDate={disabledPastDate}
                disabledTime={disabledPastTime}
                format="YYYY-MM-DD HH:mm:ss"
                style={{ width: '100%' }}
              />
            </Form.Item>
          ) : null}
          <Form.Item label={t('封禁原因')} name="reason" rules={[{ required: true, message: t('请输入封禁原因') }]}>
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={t('导入黑名单')}
        open={importOpen}
        onCancel={() => setImportOpen(false)}
        onOk={() => importForm.submit()}
        confirmLoading={importMutation.isPending}
      >
        <Form form={importForm} layout="vertical" onFinish={(values) => void importMutation.mutateAsync(values)}>
          <Form.Item label={t('封禁时长')} name="durationHours" rules={[{ required: true, message: t('请选择封禁时长') }]}>
            <Select
              options={TEMP_BAN_DURATION_OPTIONS.map((item) => ({
                value: item.value,
                label: t(item.label),
              }))}
            />
          </Form.Item>
          {importDurationMode === 'custom' ? (
            <Form.Item
              label={t('自定义时间点')}
              name="customEndAt"
              rules={[
                { required: true, message: t('请选择自定义时间点') },
                {
                  validator: async (_, value) => {
                    if (isFutureDateTime(value)) return
                    throw new Error(t('自定义时间点必须晚于当前时间'))
                  },
                },
              ]}
            >
              <DatePicker
                showTime={{ hideDisabledOptions: true }}
                disabledDate={disabledPastDate}
                disabledTime={disabledPastTime}
                format="YYYY-MM-DD HH:mm:ss"
                style={{ width: '100%' }}
              />
            </Form.Item>
          ) : null}
          <Form.Item
            label={t('请粘贴 IP 列表（每行一个）：')}
            name="ips"
            rules={[
              { required: true, message: t('请输入IP地址') },
              {
                validator: async (_, value) => {
                  const ips = parseImportIPs(value)
                  if (ips.length > 0 && ips.every(isValidIPAddress)) return
                  throw new Error(t('请输入合法IP地址'))
                },
              },
            ]}
          >
            <Input.TextArea rows={6} placeholder={'120.12.8.91\n223.66.19.42'} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
