import { useMutation, useQuery } from '@tanstack/react-query'
import { QuestionCircleOutlined } from '@ant-design/icons'
import { Card, DatePicker, Drawer, Form, Input, InputNumber, Modal, Select, Space, Table, Tag, Tooltip, Typography, message } from 'antd'
import type { TablePaginationConfig } from 'antd'
import type { SorterResult } from 'antd/es/table/interface'
import { useMemo, useState } from 'react'
import { UiButton } from '../../components/ui'
import { useI18n } from '../../i18n/messages'
import {
  applySystemAuditLogRetention,
  createSystemIPBlacklist,
  exportSystemAuditLogs,
  getSystemAuditLog,
  getSystemAuditLogs,
  type SystemAuditLog,
  type SystemAuditLogFilters,
} from '../../services/api/system'
import { useUiSettingsStore } from '../../store/uiSettings'
import { formatDateTime, toTimezoneDateTimeString } from '../../utils/timezone'

type FilterFormValues = {
  actor?: string
  action?: string
  result?: 'success' | 'failed'
  traceId?: string
  sourceIp?: string
  keyword?: string
  time?: [{ toDate?: () => Date }, { toDate?: () => Date }]
}

type BlacklistFormValues = {
  ip: string
  endAt?: DateTimeFormValue
  reason: string
}

type DateTimePickerValue = { format: (template: string) => string; valueOf: () => number }
type DateTimeFormValue = string | DateTimePickerValue
const DEFAULT_AUDIT_RETAIN_DAYS = 30

const AUDIT_ACTION_LABELS: Record<string, string> = {
  login_success: '登录成功',
  login_failed: '登录失败',
  logout: '退出登录',
  logout_others: '注销其他会话',
  force_logout: '强制下线',
  deactivate: '注销账号',
  update_personal_settings: '更新个人设置',
  auth_failed: '鉴权失败',
  view_my_profile: '查看个人资料',
  view_personal_settings: '查看个人设置',
  view_system_settings: '查看系统设置',
  view_sessions: '查看我的会话列表',
  view_department_tree: '查看部门树',
  view_department_detail: '查看部门详情',
  view_position_list: '查看岗位列表',
  view_position_detail: '查看岗位详情',
  view_user_list: '查看用户列表',
  view_user_sessions: '查看用户会话',
  export_users: '导出用户列表',
  view_menu_tree: '查看菜单树',
  view_menu_list: '查看菜单列表',
  view_menu_detail: '查看菜单详情',
  view_role_list: '查看角色列表',
  view_role_detail: '查看角色详情',
  view_role_menus: '查看角色菜单',
  view_audit_logs: '查看审计日志列表',
  view_audit_log_detail: '查看审计日志详情',
  export_audit_logs: '导出审计日志',
  view_ip_blacklist: '查看IP黑名单',
  view_warm_tips: '查看关怀提示',
  view_alert_bots: '查看告警机器人',
  view_alert_scenes: '查看告警场景',
  view_alert_templates: '查看告警模板',
  apply_audit_log_retention: '清理审计日志',
  create_ip_blacklist: '新增IP黑名单',
  update_ip_blacklist: '编辑IP黑名单',
  unblock_ip_blacklist: '解封IP黑名单',
  delete_ip_blacklist: '删除IP黑名单',
  batch_unblock_ip_blacklist: '批量解封IP黑名单',
  import_ip_blacklist: '导入IP黑名单',
  create_warm_tip: '新增关怀提示',
  update_warm_tip: '编辑关怀提示',
  delete_warm_tip: '删除关怀提示',
  save_alert_bot: '保存告警机器人',
  delete_alert_bot: '删除告警机器人',
  save_alert_scene: '保存告警场景',
  delete_alert_scene: '删除告警场景',
  test_send_alert_scene: '测试发送告警',
  save_alert_template: '保存告警模板',
  delete_alert_template: '删除告警模板',
  update_system_settings: '更新系统设置',
  view_files: '查看文件列表',
  upload_file: '上传文件',
  update_file: '编辑文件',
  delete_file: '删除文件',
  create_menu: '新增菜单',
  update_menu: '编辑菜单',
  update_menu_status: '更新菜单状态',
  delete_menu: '删除菜单',
  sync_menus: '同步菜单',
  create_role: '新增角色',
  update_role: '编辑角色',
  delete_role: '删除角色',
  update_role_menus: '更新角色菜单',
}

function auditActionLabel(action: string): string {
  return AUDIT_ACTION_LABELS[action] || action || '-'
}

function auditResultLabel(result: 'success' | 'failed'): string {
  return result === 'failed' ? '失败' : '成功'
}

function numberRange(start: number, end: number): number[] {
  return Array.from({ length: Math.max(end - start, 0) }, (_unused, index) => start + index)
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

function todayDateString(): string {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
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

export function SystemAuditLogsPage() {
  const { t } = useI18n()
  const [messageApi, contextHolder] = message.useMessage()
  const [modalApi, modalContextHolder] = Modal.useModal()
  const [filterForm] = Form.useForm<FilterFormValues>()
  const [blacklistForm] = Form.useForm<BlacklistFormValues>()
  const [detailOpen, setDetailOpen] = useState(false)
  const [blacklistOpen, setBlacklistOpen] = useState(false)
  const [selected, setSelected] = useState<SystemAuditLog | null>(null)
  const [pageNo, setPageNo] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [filters, setFilters] = useState<SystemAuditLogFilters>({})
  const [queryTrigger, setQueryTrigger] = useState(0)
  const [orderField, setOrderField] = useState<'id' | 'uid' | 'actor' | 'action' | 'result' | 'trace_id' | 'source_ip' | 'created_at' | undefined>()
  const [orderType, setOrderType] = useState<'asc' | 'desc' | undefined>()
  const [exporting, setExporting] = useState(false)
  const [retainDays, setRetainDays] = useState(DEFAULT_AUDIT_RETAIN_DAYS)
  const [retainDraftDays, setRetainDraftDays] = useState(DEFAULT_AUDIT_RETAIN_DAYS)
  const [editingRetention, setEditingRetention] = useState(false)
  const systemTimezone = useUiSettingsStore((state) => state.systemTimezone)
  const serverTimezone = useUiSettingsStore((state) => state.serverTimezone)

  const actionOptions = useMemo(
    () =>
      Object.entries(AUDIT_ACTION_LABELS).map(([value, label]) => ({
        value,
        label: t(label),
      })),
    [t],
  )

  const auditLogsQuery = useQuery({
    queryKey: ['system-audit-logs', pageNo, pageSize, orderField, orderType, filters, queryTrigger],
    queryFn: () => getSystemAuditLogs(pageNo, pageSize, orderField, orderType, filters),
  })
  const auditLogs = auditLogsQuery.data?.items ?? []

  const detailMutation = useMutation({
    mutationFn: async (id: number) => getSystemAuditLog(id),
    onSuccess: (detail) => {
      setSelected(detail)
      setDetailOpen(true)
    },
  })

  const blacklistMutation = useMutation({
    mutationFn: async (values: BlacklistFormValues) => {
      await createSystemIPBlacklist({
        ip: values.ip.trim(),
        endAt: normalizeDateTimeLocal(values.endAt),
        reason: values.reason.trim(),
        creator: '',
      })
    },
    onSuccess: () => {
      setBlacklistOpen(false)
      blacklistForm.resetFields()
      void messageApi.success(t('拉黑IP成功'))
    },
  })

  const retentionMutation = useMutation({
    mutationFn: async ({ days, confirm }: { days: number; confirm: boolean }) => applySystemAuditLogRetention(days, confirm),
  })

  const applyFilters = (values: FilterFormValues) => {
    setFilters({
      actor: values.actor?.trim() || undefined,
      action: values.action?.trim() || undefined,
      result: values.result,
      traceId: values.traceId?.trim() || undefined,
      requestId: values.requestId?.trim() || undefined,
      sourceIp: values.sourceIp?.trim() || undefined,
      keyword: values.keyword?.trim() || undefined,
      createdFrom: toTimezoneDateTimeString(values.time?.[0], systemTimezone),
      createdTo: toTimezoneDateTimeString(values.time?.[1], systemTimezone),
    })
    setPageNo(1)
    setQueryTrigger((current) => current + 1)
  }

  const handleTableChange = (
    pagination: TablePaginationConfig,
    _tableFilters: Record<string, unknown>,
    sorter: SorterResult<SystemAuditLog> | SorterResult<SystemAuditLog>[],
  ) => {
    const nextSorter = Array.isArray(sorter) ? sorter[0] : sorter
    setPageNo(pagination.current || 1)
    setPageSize(pagination.pageSize || 10)
    if (!nextSorter?.field || !nextSorter.order) {
      setOrderField(undefined)
      setOrderType(undefined)
      return
    }
    setOrderField(nextSorter.field as 'id' | 'uid' | 'actor' | 'action' | 'result' | 'trace_id' | 'source_ip' | 'created_at')
    setOrderType(nextSorter.order === 'ascend' ? 'asc' : 'desc')
  }

  const handleExport = async () => {
    setExporting(true)
    try {
      const blob = await exportSystemAuditLogs(orderField, orderType, filters)
      const url = window.URL.createObjectURL(blob)
      const anchor = document.createElement('a')
      anchor.href = url
      anchor.download = `system-audit-logs-${Date.now()}.csv`
      document.body.appendChild(anchor)
      anchor.click()
      document.body.removeChild(anchor)
      window.URL.revokeObjectURL(url)
      void messageApi.success(t('导出日志成功'))
    } finally {
      setExporting(false)
    }
  }

  const handleRetentionSubmit = async () => {
    const nextDays = Math.max(1, Math.min(3650, Math.round(retainDraftDays || DEFAULT_AUDIT_RETAIN_DAYS)))
    const preview = await retentionMutation.mutateAsync({ days: nextDays, confirm: false })
    if (preview.expiredCount > 0) {
      void modalApi.confirm({
        title: t('确认保存保留策略？'),
        content: t('本次操作将删除{expired}条过期数据，剩余{valid}条有效数据，确定要保存吗', {
          expired: preview.expiredCount,
          valid: preview.validCount,
        }),
        onOk: async () => {
          const result = await retentionMutation.mutateAsync({ days: nextDays, confirm: true })
          setRetainDays(result.retainDays)
          setRetainDraftDays(result.retainDays)
          setEditingRetention(false)
          setQueryTrigger((current) => current + 1)
          void messageApi.success(t('审计日志保留策略已保存'))
        },
      })
      return
    }
    setRetainDays(preview.retainDays)
    setRetainDraftDays(preview.retainDays)
    setEditingRetention(false)
    void messageApi.success(t('审计日志保留策略已保存'))
  }

  return (
    <>
      {contextHolder}
      {modalContextHolder}

      <Space direction="vertical" size={16} className="full-width table-scroll-page">
        <Space wrap>
          <UiButton type="primary" loading={exporting} onClick={() => void handleExport()}>
            {t('导出日志')}
          </UiButton>
          <Space size={8} wrap>
            <Typography.Text>{t('保留')}</Typography.Text>
            <InputNumber
              min={1}
              max={3650}
              precision={0}
              value={editingRetention ? retainDraftDays : retainDays}
              onChange={(value) => setRetainDraftDays(Number(value || DEFAULT_AUDIT_RETAIN_DAYS))}
              disabled={!editingRetention}
            />
            <Typography.Text>{t('日内的数据')}</Typography.Text>
            {editingRetention ? (
              <>
                <UiButton
                  onClick={() => {
                    setRetainDraftDays(retainDays)
                    setEditingRetention(false)
                  }}
                >
                  {t('取消')}
                </UiButton>
                <UiButton type="primary" loading={retentionMutation.isPending} onClick={() => void handleRetentionSubmit()}>
                  {t('确认')}
                </UiButton>
              </>
            ) : (
              <UiButton
                onClick={() => {
                  setRetainDraftDays(retainDays)
                  setEditingRetention(true)
                }}
              >
                {t('修改')}
              </UiButton>
            )}
          </Space>
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
            <Form.Item label={t('操作人')} name="actor">
              <Input placeholder={t('请输入操作人')} allowClear onPressEnter={() => filterForm.submit()} />
            </Form.Item>
            <Form.Item label={t('审计动作')} name="action">
              <Select
                allowClear
                showSearch
                style={{ width: 220 }}
                placeholder={t('请选择审计动作')}
                options={actionOptions}
                optionFilterProp="label"
              />
            </Form.Item>
            <Form.Item label={t('结果')} name="result">
              <Select
                style={{ width: 140 }}
                allowClear
                options={[
                  { value: 'success', label: t('成功') },
                  { value: 'failed', label: t('失败') },
                ]}
              />
            </Form.Item>
            <Form.Item label="TraceID" name="traceId">
              <Input placeholder={t('请输入 TraceID')} allowClear onPressEnter={() => filterForm.submit()} />
            </Form.Item>
            <Form.Item label="RequestID" name="requestId">
              <Input placeholder={t('请输入 RequestID')} allowClear onPressEnter={() => filterForm.submit()} />
            </Form.Item>
            <Form.Item label={t('请求IP')} name="sourceIp">
              <Input placeholder={t('请输入请求IP')} allowClear onPressEnter={() => filterForm.submit()} />
            </Form.Item>
            <Form.Item label={t('关键字')} name="keyword">
              <Input placeholder={t('详情/动作/TraceID')} allowClear onPressEnter={() => filterForm.submit()} />
            </Form.Item>
            <Form.Item label={t('时间范围')} name="time">
              <DatePicker.RangePicker showTime />
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
                    setQueryTrigger((current) => current + 1)
                  }}
                >
                  {t('重置')}
                </UiButton>
              </Space>
            </Form.Item>
          </Form>
        </Card>

        <Card className="compact-table-card system-table-card table-scroll-region">
          <Table<SystemAuditLog>
            rowKey="id"
            loading={auditLogsQuery.isLoading}
            dataSource={auditLogs}
            onChange={handleTableChange}
            pagination={{
              current: pageNo,
              pageSize,
              total: auditLogsQuery.data?.total ?? 0,
              showSizeChanger: true,
              showTotal: (total) => t('共 {total} 条', { total }),
            }}
            scroll={{ x: 1430, y: 392 }}
            columns={[
            { title: t('日志ID'), dataIndex: 'id', sorter: true, width: 92 },
            { title: t('操作人'), dataIndex: 'actor', sorter: true, width: 140, render: (value: string) => value || '-' },
            {
              title: t('审计动作'),
              dataIndex: 'action',
              width: 190,
              render: (value: string) => t(auditActionLabel(value)),
            },
            {
              title: t('结果'),
              dataIndex: 'result',
              sorter: true,
              width: 88,
              render: (value: SystemAuditLog['result']) => (
                <Tag color={value === 'success' ? 'green' : 'red'}>{t(auditResultLabel(value))}</Tag>
              ),
            },
            { title: t('请求IP'), dataIndex: 'sourceIp', sorter: true, width: 150, render: (value: string) => value || '-' },
            { title: t('耗时'), dataIndex: 'duration', width: 110, render: (value: string) => value || '-' },
            { title: <Space size={4}>TraceID<Tooltip title={t('由前端生成，用于关联同一用户操作的前后端请求链路')}><QuestionCircleOutlined style={{ color: 'var(--color-text-secondary)', fontSize: 12 }} /></Tooltip></Space>, dataIndex: 'traceId', width: 200, ellipsis: true, render: (value: string) => value || '-' },
            { title: <Space size={4}>RequestID<Tooltip title={t('由后端生成，唯一标识每次服务端请求，不可伪造')}><QuestionCircleOutlined style={{ color: 'var(--color-text-secondary)', fontSize: 12 }} /></Tooltip></Space>, dataIndex: 'requestId', width: 200, ellipsis: true, render: (value: string) => value || '-' },
            { title: t('用户代理'), dataIndex: 'userAgent', width: 200, ellipsis: true, render: (value: string) => value || '-' },
            { title: t('详情'), dataIndex: 'detail', width: 220, ellipsis: true, render: (value: string) => value || '-' },
            {
              title: t('操作时间'),
              dataIndex: 'createdAt',
              sorter: true,
              width: 180,
              fixed: 'right',
              render: (value: string) => formatDateTime(value, systemTimezone, serverTimezone),
            },
            {
              title: t('操作'),
              width: 180,
              fixed: 'right',
              render: (_, row) => (
                <Space size={0}>
                  <UiButton
                    type="link"
                    loading={detailMutation.isPending && selected?.id === row.id}
                    onClick={() => {
                      setSelected(row)
                      detailMutation.mutate(row.id)
                    }}
                  >
                    {t('查看详情')}
                  </UiButton>
                  <UiButton
                    type="link"
                    danger
                    onClick={() => {
                      if (!row.sourceIp) {
                        void messageApi.warning(t('该记录无请求IP'))
                        return
                      }
                      blacklistForm.setFieldsValue({
                        ip: row.sourceIp,
                        endAt: undefined,
                        reason: t('审计页面拉黑'),
                      })
                      setBlacklistOpen(true)
                    }}
                  >
                    {t('拉黑IP')}
                  </UiButton>
                </Space>
              ),
            },
            ]}
          />
        </Card>
      </Space>

      <Drawer title={t('审计日志详情')} open={detailOpen} onClose={() => setDetailOpen(false)} width={560}>
        <Typography.Paragraph>{t('日志ID：')}{selected?.id ?? '-'}</Typography.Paragraph>
        <Typography.Paragraph>{t('用户UID：')}{selected?.uid ?? '-'}</Typography.Paragraph>
        <Typography.Paragraph>{t('操作人：')}{selected?.actor || '-'}</Typography.Paragraph>
        <Typography.Paragraph>{t('审计动作：')}{selected ? t(auditActionLabel(selected.action)) : '-'}</Typography.Paragraph>
        <Typography.Paragraph>{t('结果：')}{selected ? t(auditResultLabel(selected.result)) : '-'}</Typography.Paragraph>
        <Typography.Paragraph>{t('请求IP：')}{selected?.sourceIp || '-'}</Typography.Paragraph>
        <Typography.Paragraph>{t('耗时：')}{selected?.duration || '-'}</Typography.Paragraph>
        <Typography.Paragraph>TraceID：{selected?.traceId || '-'}</Typography.Paragraph>
        <Typography.Paragraph>{t('用户代理：')}{selected?.userAgent || '-'}</Typography.Paragraph>
        <Typography.Paragraph>{t('详情：')}{selected?.detail || '-'}</Typography.Paragraph>
        <Typography.Paragraph>{t('操作时间：')}{formatDateTime(selected?.createdAt, systemTimezone, serverTimezone)}</Typography.Paragraph>
      </Drawer>

      <Modal
        title={t('拉黑IP设置')}
        open={blacklistOpen}
        onCancel={() => setBlacklistOpen(false)}
        onOk={() => blacklistForm.submit()}
        confirmLoading={blacklistMutation.isPending}
      >
        <Form form={blacklistForm} layout="vertical" onFinish={(values) => void blacklistMutation.mutateAsync(values)}>
          <Form.Item label={t('IP地址')} name="ip" rules={[{ required: true, message: t('请输入IP地址') }]}>
            <Input disabled />
          </Form.Item>
          <Form.Item
            label={t('失效时间')}
            name="endAt"
            rules={[
              {
                validator: async (_, value) => {
                  if (!value || isFutureDateTime(value)) return
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
          <Form.Item label={t('封禁原因')} name="reason" rules={[{ required: true, message: t('请输入封禁原因') }]}>
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
