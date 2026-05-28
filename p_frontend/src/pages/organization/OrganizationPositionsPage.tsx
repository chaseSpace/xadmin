import { useMutation, useQuery } from '@tanstack/react-query'
import { Card, Form, Input, InputNumber, Modal, Select, Space, Table, Tag, Tooltip, message } from 'antd'
import { ExclamationCircleOutlined } from '@ant-design/icons'
import type { TablePaginationConfig } from 'antd'
import type { SorterResult } from 'antd/es/table/interface'
import { useMemo, useState } from 'react'
import { UiButton } from '../../components/ui'
import { useI18n } from '../../i18n/messages'
import { useUiSettingsStore } from '../../store/uiSettings'
import { formatDateTime } from '../../utils/timezone'
import { getPermissionRoles } from '../../services/api/permission'
import {
  createOrganizationPosition,
  deleteOrganizationPosition,
  getOrganizationDepartmentsTree,
  getOrganizationPositions,
  toggleOrganizationPositionStatus,
  updateOrganizationPosition,
  type OrganizationPosition,
  type OrganizationPositionFilters,
} from '../../services/api/organization'

type PositionFormValues = {
  name: string
  code: string
  departmentId: number
  level: string
  hc: number
  staffed: number
  status: 'enabled' | 'disabled'
  roleIds: number[]
}

type FilterFormValues = {
  keyword?: string
  departmentId?: number
  level?: string
  status?: 'enabled' | 'disabled'
}

function flattenDepartments(
  items: Awaited<ReturnType<typeof getOrganizationDepartmentsTree>>,
): Array<{ id: number; name: string; status: 'enabled' | 'disabled' }> {
  const result: Array<{ id: number; name: string; status: 'enabled' | 'disabled' }> = []
  const walk = (nodes: typeof items) => {
    for (const node of nodes) {
      result.push({ id: node.id, name: node.name, status: node.status })
      if (node.children.length > 0) {
        walk(node.children)
      }
    }
  }
  walk(items)
  return result
}

export function OrganizationPositionsPage() {
  const { t } = useI18n()
  const systemTimezone = useUiSettingsStore((state) => state.systemTimezone)
  const [messageApi, contextHolder] = message.useMessage()
  const [modalApi, modalContextHolder] = Modal.useModal()
  const [filterForm] = Form.useForm<FilterFormValues>()
  const [positionForm] = Form.useForm<PositionFormValues>()
  const [pageNo, setPageNo] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [filters, setFilters] = useState<OrganizationPositionFilters>({})
  const [queryTrigger, setQueryTrigger] = useState(0)
  const [selectedRowKeys, setSelectedRowKeys] = useState<number[]>([])
  const [selected, setSelected] = useState<OrganizationPosition | null>(null)
  const [roleFallbackOptions, setRoleFallbackOptions] = useState<Array<{ value: number; label: string }>>([])
  const [formOpen, setFormOpen] = useState(false)
  const [editorMode, setEditorMode] = useState<'create' | 'edit'>('create')
  const [orderField, setOrderField] = useState<'id' | 'name' | 'level' | 'status' | 'updated_at' | undefined>()
  const [orderType, setOrderType] = useState<'asc' | 'desc' | undefined>()

  const departmentsQuery = useQuery({
    queryKey: ['organization-departments-tree-for-positions'],
    queryFn: () => getOrganizationDepartmentsTree(),
  })

  const departmentFilterOptions = flattenDepartments(departmentsQuery.data ?? []).map((item) => ({
    value: item.id,
    label: `${item.name}${item.status === 'disabled' ? ` [${t('停用')}]` : ''}`,
  }))
  const departmentFormOptions = flattenDepartments(departmentsQuery.data ?? []).map((item) => ({
    value: item.id,
    label: `${item.name}${item.status === 'disabled' ? ` [${t('停用')}]` : ''}`,
    disabled: item.status === 'disabled',
  }))

  const positionsQuery = useQuery({
    queryKey: ['organization-positions', pageNo, pageSize, orderField, orderType, filters, queryTrigger],
    queryFn: () => getOrganizationPositions(pageNo, pageSize, orderField, orderType, filters),
  })
  const rolesQuery = useQuery({
    queryKey: ['permission-roles-for-positions'],
    queryFn: () => getPermissionRoles(1, 200),
  })
  const roleOptions = useMemo(() => {
    const optionMap = new Map<number, string>()
    ;(rolesQuery.data?.items ?? []).forEach((item) => {
      const id = Number(item.id)
      if (!Number.isFinite(id) || id <= 0) return
      if (!optionMap.has(id)) {
        optionMap.set(id, item.roleName)
      }
    })
    return Array.from(optionMap.entries()).map(([value, label]) => ({ value, label }))
  }, [rolesQuery.data?.items])
  const roleSelectOptions = useMemo(() => {
    const optionMap = new Map<number, string>(roleOptions.map((item) => [item.value, item.label]))
    roleFallbackOptions.forEach((item) => {
      if (!optionMap.has(item.value)) {
        optionMap.set(item.value, item.label)
      }
    })
    return Array.from(optionMap.entries()).map(([value, label]) => ({ value, label }))
  }, [roleFallbackOptions, roleOptions])

  const renderRoleSummary = (roleNames: string[], roleIds: number[]) => {
    if (!roleIds || roleIds.length === 0) {
      return '-'
    }
    const names = roleNames.length === roleIds.length ? roleNames : roleIds.map((id, index) => roleNames[index] || t('角色#{id}', { id }))
    return <span style={{ whiteSpace: 'pre-line' }}>{names.join('\n')}</span>
  }

  const saveMutation = useMutation({
    mutationFn: async (values: PositionFormValues) => {
      const normalizedRoleIDs = Array.from(
        new Set((values.roleIds || []).map((id) => Number(id)).filter((id) => Number.isFinite(id) && id > 0)),
      )
      if (editorMode === 'create') {
        await createOrganizationPosition({
          name: values.name.trim(),
          code: values.code.trim(),
          departmentId: values.departmentId,
          level: values.level.trim(),
          hc: values.hc,
          staffed: values.staffed,
          roleIds: normalizedRoleIDs,
        })
        return
      }
      if (!selected) return
      await updateOrganizationPosition(selected.id, {
        name: values.name.trim(),
        code: values.code.trim(),
        departmentId: values.departmentId,
        level: values.level.trim(),
        hc: values.hc,
        staffed: values.staffed,
        status: values.status,
        roleIds: normalizedRoleIDs,
      })
    },
    onSuccess: async () => {
      void messageApi.success(t(editorMode === 'create' ? '新增岗位成功' : '编辑岗位成功'))
      setFormOpen(false)
      setSelected(null)
      await positionsQuery.refetch()
    },
  })

  const toggleStatusMutation = useMutation({
    mutationFn: async (row: OrganizationPosition) => {
      await toggleOrganizationPositionStatus(row.id, row.status !== 'enabled')
    },
    onSuccess: async () => {
      void messageApi.success(t('岗位状态已更新'))
      await positionsQuery.refetch()
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => {
      await deleteOrganizationPosition(id)
    },
    onSuccess: async () => {
      void messageApi.success(t('岗位删除成功'))
      await positionsQuery.refetch()
    },
  })

  const applyFilters = (values: FilterFormValues) => {
    setFilters({
      keyword: values.keyword?.trim() || undefined,
      departmentId: values.departmentId,
      level: values.level?.trim() || undefined,
      status: values.status,
    })
    setPageNo(1)
    setQueryTrigger((current) => current + 1)
  }

  const openCreateForm = async () => {
    await Promise.all([departmentsQuery.refetch(), rolesQuery.refetch()])
    setEditorMode('create')
    setSelected(null)
    setRoleFallbackOptions([])
    positionForm.setFieldsValue({
      name: '',
      code: '',
      departmentId: undefined,
      level: '',
      hc: 1,
      staffed: 0,
      status: 'enabled',
      roleIds: [],
    })
    setFormOpen(true)
  }

  const openEditForm = async (row: OrganizationPosition) => {
    await Promise.all([departmentsQuery.refetch(), rolesQuery.refetch()])
    setEditorMode('edit')
    setSelected(row)
    setRoleFallbackOptions(
      (row.roleIds || [])
        .map((id, index) => {
          const numericID = Number(id)
          if (!Number.isFinite(numericID) || numericID <= 0) {
            return null
          }
          return {
            value: numericID,
            label: row.roleNames[index] || t('角色#{id}', { id: numericID }),
          }
        })
        .filter((item): item is { value: number; label: string } => item !== null),
    )
    positionForm.setFieldsValue({
      name: row.name,
      code: row.code,
      departmentId: row.departmentId,
      level: row.level,
      hc: row.hc,
      staffed: row.staffed,
      status: row.status,
      roleIds: (row.roleIds || []).map((id) => Number(id)).filter((id) => Number.isFinite(id) && id > 0),
    })
    setFormOpen(true)
  }

  const handleBatchStatus = async (enabled: boolean) => {
    if (selectedRowKeys.length === 0) {
      void messageApi.warning(t('请先勾选岗位'))
      return
    }
    const rows = positionsQuery.data?.items ?? []
    const selectedRows = rows.filter((item) => selectedRowKeys.includes(item.id))
    const settled = await Promise.allSettled(selectedRows.map((row) => toggleOrganizationPositionStatus(row.id, enabled)))
    const success = settled.filter((item) => item.status === 'fulfilled').length
    const failed = settled.length - success
    if (success > 0) {
      void messageApi.success(t(enabled ? '批量启用成功 {count} 条' : '批量停用成功 {count} 条', { count: success }))
    }
    if (failed > 0) {
      void messageApi.warning(t('批量操作失败 {count} 条，请重试', { count: failed }))
    }
    setSelectedRowKeys([])
    await positionsQuery.refetch()
  }

  return (
    <>
      {contextHolder}
      {modalContextHolder}

      <Space direction="vertical" size={16} className="full-width table-scroll-page">
      <Space wrap>
        <UiButton type="primary" onClick={openCreateForm}>
          {t('新增岗位')}
        </UiButton>
        <UiButton onClick={() => void handleBatchStatus(true)}>{t('批量启用')}</UiButton>
        <UiButton onClick={() => void handleBatchStatus(false)}>{t('批量停用')}</UiButton>
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
          <Form.Item label={t('关键词')} name="keyword">
            <Input placeholder={t('岗位名称/岗位编码')} allowClear onPressEnter={() => filterForm.submit()} />
          </Form.Item>
          <Form.Item label={t('职级')} name="level">
            <Input placeholder={t('如 P5 / M1')} allowClear onPressEnter={() => filterForm.submit()} />
          </Form.Item>
          <Form.Item label={t('部门')} name="departmentId">
            <Select style={{ width: 220 }} allowClear options={departmentFilterOptions} loading={departmentsQuery.isLoading} />
          </Form.Item>
          <Form.Item label={t('状态')} name="status">
            <Select
              style={{ width: 140 }}
              allowClear
              options={[
                { value: 'enabled', label: t('启用') },
                { value: 'disabled', label: t('停用') },
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

      <Card className="compact-table-card system-table-card table-scroll-region">
        <Table<OrganizationPosition>
          rowKey="id"
          loading={positionsQuery.isLoading || positionsQuery.isFetching}
          dataSource={positionsQuery.data?.items ?? []}
          rowSelection={{
            selectedRowKeys,
            onChange: (keys) => setSelectedRowKeys(keys as number[]),
          }}
          scroll={{ x: 'max-content', y: 392 }}
          pagination={{
            current: pageNo,
            pageSize,
            total: positionsQuery.data?.total ?? 0,
            showSizeChanger: true,
          }}
          onChange={(pagination: TablePaginationConfig, _filters, sorter: SorterResult<OrganizationPosition> | SorterResult<OrganizationPosition>[]) => {
            if (pagination.current) {
              setPageNo(pagination.current)
            }
            if (pagination.pageSize && pagination.pageSize !== pageSize) {
              setPageSize(pagination.pageSize)
              setPageNo(1)
            }
            const singleSorter = Array.isArray(sorter) ? sorter[0] : sorter
            if (!singleSorter || !singleSorter.field || !singleSorter.order) {
              setOrderField(undefined)
              setOrderType(undefined)
              return
            }
            const fieldMap: Record<string, 'id' | 'name' | 'level' | 'status' | 'updated_at'> = {
              id: 'id',
              name: 'name',
              level: 'level',
              status: 'status',
              updatedAt: 'updated_at',
            }
            const mapped = fieldMap[String(singleSorter.field)]
            if (!mapped) {
              setOrderField(undefined)
              setOrderType(undefined)
              return
            }
            setOrderField(mapped)
            setOrderType(singleSorter.order === 'ascend' ? 'asc' : 'desc')
          }}
          columns={[
            { title: t('岗位ID'), dataIndex: 'id', sorter: true },
            { title: t('岗位名称'), dataIndex: 'name', sorter: true },
            { title: t('岗位编码'), dataIndex: 'code' },
            { title: t('所属部门'), dataIndex: 'departmentName' },
            {
              title: t('绑定角色'),
              render: (_, row) => renderRoleSummary(row.roleNames, row.roleIds),
            },
            { title: t('职级'), dataIndex: 'level', sorter: true },
            {
              title: (
                <Space size={4}>{t('关联人数')}<Tooltip title={t('当前岗位关联的全部未删除账号数量')}><ExclamationCircleOutlined /></Tooltip></Space>
              ),
              dataIndex: 'relatedCount',
            },
            {
              title: (
                <Space size={4}>{t('在岗人数')}<Tooltip title={t('当前岗位下启用状态账号数量')}><ExclamationCircleOutlined /></Tooltip></Space>
              ),
              dataIndex: 'staffed',
            },
            {
              title: t('状态'),
              dataIndex: 'status',
              sorter: true,
              render: (status: OrganizationPosition['status']) => (
                <Tag color={status === 'enabled' ? 'green' : 'default'}>{status === 'enabled' ? t('启用') : t('停用')}</Tag>
              ),
            },
            {
              title: t('更新时间'),
              dataIndex: 'updatedAt',
              sorter: true,
              render: (value: string) => formatDateTime(value, systemTimezone),
            },
            {
              title: t('操作'),
              width: 200,
              fixed: 'right',
              render: (_, row) => (
                <Space size={0}>
                  <UiButton
                    type="link"
                    onClick={() => {
                      void openEditForm(row)
                    }}
                  >
                    {t('编辑')}
                  </UiButton>
                  <UiButton
                    type="link"
                    onClick={() => {
                      void toggleStatusMutation.mutateAsync(row)
                    }}
                  >
                    {row.status === 'enabled' ? t('停用') : t('启用')}
                  </UiButton>
                  <UiButton
                    type="link"
                    danger
                    onClick={() => {
                      void modalApi.confirm({
                        title: t('确认删除岗位 {name}', { name: row.name }),
                        content: t('删除后不可恢复。'),
                        okButtonProps: { danger: true },
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
        title={t(editorMode === 'create' ? '新增岗位' : '编辑岗位')}
        open={formOpen}
        onCancel={() => {
          setFormOpen(false)
          setRoleFallbackOptions([])
        }}
        onOk={() => positionForm.submit()}
        confirmLoading={saveMutation.isPending}
      >
        <Form form={positionForm} layout="vertical" onFinish={(values) => void saveMutation.mutateAsync(values)}>
          <Form.Item label={t('岗位名称')} name="name" rules={[{ required: true, message: t('请输入岗位名称') }]}>
            <Input maxLength={64} />
          </Form.Item>
          <Form.Item label={t('岗位编码')} name="code" rules={[{ required: true, message: t('请输入岗位编码') }]}>
            <Input maxLength={64} />
          </Form.Item>
          <Form.Item label={t('所属部门')} name="departmentId" rules={[{ required: true, message: t('请选择所属部门') }]}>
            <Select options={departmentFormOptions} loading={departmentsQuery.isLoading} />
          </Form.Item>
          <Form.Item label={t('职级')} name="level" rules={[{ required: true, message: t('请输入岗位职级') }]}>
            <Input maxLength={32} />
          </Form.Item>
          <Space className="full-width" size={12}>
            <Form.Item label={t('编制人数')} name="hc" rules={[{ required: true, message: t('请输入编制人数') }]} style={{ flex: 1 }}>
              <InputNumber min={0} max={10000} precision={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item label={t('在岗人数')} name="staffed" rules={[{ required: true, message: t('请输入在岗人数') }]} style={{ flex: 1 }}>
              <InputNumber min={0} max={10000} precision={0} style={{ width: '100%' }} />
            </Form.Item>
          </Space>
          <Form.Item label={t('状态')} name="status" rules={[{ required: true, message: t('请选择状态') }]}>
            <Select
              options={[
                { value: 'enabled', label: t('启用') },
                { value: 'disabled', label: t('停用') },
              ]}
            />
          </Form.Item>
          <Form.Item label={t('关联角色')} name="roleIds">
            <Select mode="multiple" allowClear options={roleSelectOptions} loading={rolesQuery.isLoading} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
