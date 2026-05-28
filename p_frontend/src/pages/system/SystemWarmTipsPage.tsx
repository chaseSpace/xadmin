import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Card, Form, Input, InputNumber, Modal, Select, Space, Table, Tag, message } from 'antd'
import type { TablePaginationConfig } from 'antd'
import type { SorterResult } from 'antd/es/table/interface'
import { useState } from 'react'
import { UiButton } from '../../components/ui'
import { useI18n } from '../../i18n/messages'
import {
  createSystemWarmTip,
  deleteSystemWarmTip,
  getSystemWarmTips,
  updateSystemWarmTip,
  type SaveSystemWarmTipPayload,
  type SystemWarmTipFilters,
  type SystemWarmTipItem,
  type SystemWarmTipType,
} from '../../services/api/system'
import { useUiSettingsStore } from '../../store/uiSettings'
import { formatDateTime } from '../../utils/timezone'
import { warmTipKeys } from '../../services/api/queryKeys'

type FilterFormValues = {
  keyword?: string
  tipType?: SystemWarmTipType
  status?: 'enabled' | 'disabled'
}

type EditorFormValues = SaveSystemWarmTipPayload

const TIP_TYPE_OPTIONS: Array<{ value: SystemWarmTipType; label: string }> = [
  { value: 'rest', label: '休息提示' },
  { value: 'positive', label: '正能量' },
  { value: 'quote', label: '名人名言' },
  { value: 'line', label: '经典台词' },
]

function isWarmTipLengthValid(value: string): boolean {
  const input = value.trim()
  if (!input) return false
  const words = input.split(/\s+/).filter(Boolean)
  if (words.length > 1) return words.length >= 3 && words.length <= 20
  const count = Array.from(input).length
  return count >= 3 && count <= 40
}

export function SystemWarmTipsPage() {
  const { t } = useI18n()
  const systemTimezone = useUiSettingsStore((state) => state.systemTimezone)
  const queryClient = useQueryClient()
  const [messageApi, contextHolder] = message.useMessage()
  const [modalApi, modalContextHolder] = Modal.useModal()
  const [filterForm] = Form.useForm<FilterFormValues>()
  const [editorForm] = Form.useForm<EditorFormValues>()
  const [filters, setFilters] = useState<SystemWarmTipFilters>({})
  const [queryTrigger, setQueryTrigger] = useState(0)
  const [pageNo, setPageNo] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [orderField, setOrderField] = useState<'id' | 'tip_type' | 'sort' | 'status' | 'updated_at' | undefined>('sort')
  const [orderType, setOrderType] = useState<'asc' | 'desc' | undefined>('asc')
  const [editingRow, setEditingRow] = useState<SystemWarmTipItem | null>(null)
  const [editorOpen, setEditorOpen] = useState(false)

  const listQuery = useQuery({
    queryKey: ['system-warm-tips', pageNo, pageSize, orderField, orderType, filters, queryTrigger],
    queryFn: () => getSystemWarmTips(pageNo, pageSize, orderField, orderType, filters),
  })
  const warmTips = listQuery.data?.items ?? []

  const saveMutation = useMutation({
    mutationFn: async (values: EditorFormValues) => {
      const payload: SaveSystemWarmTipPayload = {
        tipType: values.tipType,
        contentZh: values.contentZh.trim(),
        contentEn: values.contentEn.trim(),
        sort: Number(values.sort) || 0,
        status: Number(values.status) === 0 ? 0 : 1,
      }
      if (editingRow) {
        await updateSystemWarmTip(editingRow.id, payload)
        return
      }
      await createSystemWarmTip(payload)
    },
    onSuccess: async () => {
      void messageApi.success(editingRow ? t('编辑成功') : t('新增成功'))
      setEditorOpen(false)
      setEditingRow(null)
      await queryClient.invalidateQueries({ queryKey: warmTipKeys.enabled })
      await listQuery.refetch()
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => deleteSystemWarmTip(id),
    onSuccess: async () => {
      void messageApi.success(t('删除成功'))
      await queryClient.invalidateQueries({ queryKey: warmTipKeys.enabled })
      await listQuery.refetch()
    },
  })

  const applyFilters = (values: FilterFormValues) => {
    setFilters({
      keyword: values.keyword?.trim() || undefined,
      tipType: values.tipType,
      status: values.status,
    })
    setPageNo(1)
    setQueryTrigger((current) => current + 1)
  }

  const openCreateModal = () => {
    setEditingRow(null)
    editorForm.resetFields()
    editorForm.setFieldsValue({ tipType: 'rest', status: 1, sort: 100 })
    setEditorOpen(true)
  }

  const openEditModal = (row: SystemWarmTipItem) => {
    setEditingRow(row)
    editorForm.setFieldsValue({
      tipType: row.tipType,
      contentZh: row.contentZh,
      contentEn: row.contentEn,
      sort: row.sort,
      status: row.status,
    })
    setEditorOpen(true)
  }

  return (
    <>
      {contextHolder}
      {modalContextHolder}
      <Space direction="vertical" size={16} className="full-width table-scroll-page">
        <Space wrap>
          <UiButton type="primary" onClick={openCreateModal}>
            {t('新增关怀提示')}
          </UiButton>
        </Space>

        <Card>
          <Form form={filterForm} layout="inline" style={{ rowGap: 12 }} onFinish={applyFilters}>
            <Form.Item label={t('关键字')} name="keyword">
              <Input placeholder={t('中文 / English')} allowClear onPressEnter={() => filterForm.submit()} />
            </Form.Item>
            <Form.Item label={t('提示类型')} name="tipType">
              <Select style={{ width: 140 }} allowClear options={TIP_TYPE_OPTIONS.map((item) => ({ value: item.value, label: t(item.label) }))} />
            </Form.Item>
            <Form.Item label={t('状态')} name="status">
              <Select
                style={{ width: 140 }}
                allowClear
                options={[
                  { value: 'enabled', label: t('启用') },
                  { value: 'disabled', label: t('禁用') },
                ]}
              />
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

        <Card title={t('关怀提示')} className="compact-table-card system-table-card table-scroll-region">
          <Table<SystemWarmTipItem>
            rowKey="id"
            loading={listQuery.isLoading || listQuery.isFetching}
            dataSource={warmTips}
            scroll={{ x: 1180, y: 392 }}
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
              sorter: SorterResult<SystemWarmTipItem> | SorterResult<SystemWarmTipItem>[],
            ) => {
              if (pagination.current) setPageNo(pagination.current)
              if (pagination.pageSize && pagination.pageSize !== pageSize) {
                setPageSize(pagination.pageSize)
                setPageNo(1)
              }
              const single = Array.isArray(sorter) ? sorter[0] : sorter
              if (!single || !single.field || !single.order) {
                setOrderField('sort')
                setOrderType('asc')
                return
              }
              const mapper: Record<string, 'id' | 'tip_type' | 'sort' | 'status' | 'updated_at'> = {
                id: 'id',
                tipType: 'tip_type',
                sort: 'sort',
                status: 'status',
                updatedAt: 'updated_at',
              }
              const field = mapper[String(single.field)]
              setOrderField(field || 'sort')
              setOrderType(single.order === 'ascend' ? 'asc' : 'desc')
            }}
            columns={[
            { title: t('记录ID'), dataIndex: 'id', sorter: true, width: 88 },
            {
              title: t('提示类型'),
              dataIndex: 'tipType',
              sorter: true,
              width: 112,
              render: (value: SystemWarmTipType) => t(TIP_TYPE_OPTIONS.find((item) => item.value === value)?.label || value),
            },
            { title: t('中文文案'), dataIndex: 'contentZh', ellipsis: true, width: 260 },
            { title: t('英文文案'), dataIndex: 'contentEn', ellipsis: true, width: 280 },
            { title: t('排序'), dataIndex: 'sort', sorter: true, width: 90 },
            {
              title: t('状态'),
              dataIndex: 'status',
              sorter: true,
              width: 90,
              render: (value: SystemWarmTipItem['status']) => value === 1 ? <Tag color="green">{t('启用')}</Tag> : <Tag>{t('禁用')}</Tag>,
            },
            { title: t('更新时间'), dataIndex: 'updatedAt', sorter: true, width: 170, render: (value: string) => formatDateTime(value, systemTimezone) },
            {
              title: t('操作'),
              fixed: 'right',
              width: 150,
              render: (_, row) => (
                <Space size={0}>
                  <UiButton type="link" onClick={() => openEditModal(row)}>
                    {t('编辑')}
                  </UiButton>
                  <UiButton
                    type="link"
                    danger
                    loading={deleteMutation.isPending}
                    onClick={() => {
                      void modalApi.confirm({
                        title: t('确认删除关怀提示 {id}？', { id: row.id }),
                        content: t('删除后不可恢复。'),
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
        title={editingRow ? t('编辑关怀提示') : t('新增关怀提示')}
        open={editorOpen}
        onCancel={() => {
          setEditorOpen(false)
          setEditingRow(null)
        }}
        onOk={() => editorForm.submit()}
        confirmLoading={saveMutation.isPending}
      >
        <Form form={editorForm} layout="vertical" onFinish={(values) => void saveMutation.mutateAsync(values)}>
          <Form.Item label={t('提示类型')} name="tipType" rules={[{ required: true, message: t('请选择提示类型') }]}>
            <Select options={TIP_TYPE_OPTIONS.map((item) => ({ value: item.value, label: t(item.label) }))} />
          </Form.Item>
          <Form.Item
            label={t('中文文案')}
            name="contentZh"
            rules={[
              { required: true, message: t('请输入中文文案') },
              {
                validator: async (_, value) => {
                  if (isWarmTipLengthValid(String(value || ''))) return
                  throw new Error(t('文案需控制在3到20个词内'))
                },
              },
            ]}
          >
            <Input.TextArea rows={3} maxLength={160} showCount />
          </Form.Item>
          <Form.Item
            label={t('英文文案')}
            name="contentEn"
            rules={[
              { required: true, message: t('请输入英文文案') },
              {
                validator: async (_, value) => {
                  if (isWarmTipLengthValid(String(value || ''))) return
                  throw new Error(t('文案需控制在3到20个词内'))
                },
              },
            ]}
          >
            <Input.TextArea rows={3} maxLength={240} showCount />
          </Form.Item>
          <Form.Item label={t('排序')} name="sort" rules={[{ required: true, message: t('请输入排序') }]}>
            <InputNumber min={0} max={100000} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label={t('状态')} name="status" rules={[{ required: true, message: t('请选择状态') }]}>
            <Select
              options={[
                { value: 1, label: t('启用') },
                { value: 0, label: t('禁用') },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
