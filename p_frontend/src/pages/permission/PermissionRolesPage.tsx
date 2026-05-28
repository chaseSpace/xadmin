import { useMutation, useQuery } from '@tanstack/react-query'
import { Card, Drawer, Form, Input, Modal, Select, Space, Table, Tag, Tooltip, Tree, message } from 'antd'
import type { DataNode } from 'antd/es/tree'
import type { TablePaginationConfig } from 'antd'
import type { SorterResult } from 'antd/es/table/interface'
import { useState } from 'react'
import type { Key } from 'react'
import { UiButton } from '../../components/ui'
import { useI18n } from '../../i18n/messages'
import { useUiSettingsStore } from '../../store/uiSettings'
import { formatDateTime } from '../../utils/timezone'
import {
  createPermissionRole,
  deletePermissionRole,
  getPermissionRole,
  getPermissionRoleMenus,
  getPermissionRoles,
  getPermissionMenuTree,
  updatePermissionRole,
  updatePermissionRoleMenus,
  type CreatePermissionRolePayload,
  type PermissionRole,
  type PermissionRoleType,
  type PermissionRolesFilters,
  type PermissionMenuTreeNode,
} from '../../services/api/permission'

type RoleFormValues = {
  roleName: string
  roleType: PermissionRoleType
}

type FilterFormValues = {
  keyword?: string
  roleType?: PermissionRoleType
}

function mapTree(nodes: PermissionMenuTreeNode[]): DataNode[] {
  return nodes.map((node) => ({
    key: String(node.id),
    title: node.name,
    children: mapTree(node.children),
  }))
}

function getDescendantKeys(nodes: PermissionMenuTreeNode[], parentId: string): string[] {
  const keys: string[] = []
  const walk = (list: PermissionMenuTreeNode[]) => {
    for (const node of list) {
      keys.push(String(node.id))
      walk(node.children)
    }
  }
  const find = (list: PermissionMenuTreeNode[]): PermissionMenuTreeNode | undefined => {
    for (const node of list) {
      if (String(node.id) === parentId) return node
      const found = find(node.children)
      if (found) return found
    }
    return undefined
  }
  const target = find(nodes)
  if (target) walk(target.children)
  return keys
}


export function PermissionRolesPage() {
  const { t } = useI18n()
  const systemTimezone = useUiSettingsStore((state) => state.systemTimezone)
  const [messageApi, contextHolder] = message.useMessage()
  const [modalApi, modalContextHolder] = Modal.useModal()
  const [filterForm] = Form.useForm<FilterFormValues>()
  const [roleForm] = Form.useForm<RoleFormValues>()

  const [pageNo, setPageNo] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [filters, setFilters] = useState<PermissionRolesFilters>({})
  const [queryTrigger, setQueryTrigger] = useState(0)
  const [orderField, setOrderField] = useState<'id' | 'role_name' | 'role_type' | 'updated_at' | 'users' | undefined>()
  const [orderType, setOrderType] = useState<'asc' | 'desc' | undefined>()

  const [selected, setSelected] = useState<PermissionRole | null>(null)
  const [editorMode, setEditorMode] = useState<'create' | 'edit' | 'copy'>('create')
  const [roleModalOpen, setRoleModalOpen] = useState(false)
  const [permissionDrawerOpen, setPermissionDrawerOpen] = useState(false)
  const [checkedMenuKeys, setCheckedMenuKeys] = useState<Key[]>([])

  const isSystemRole = (role?: PermissionRole | null) => Boolean(role) && role?.roleType === 'system'
  const isRootAdminRole = (role?: PermissionRole | null) =>
    Boolean(role) && role?.roleType === 'system' && role.roleName.trim() === '超级管理员'

  const rolesQuery = useQuery({
    queryKey: ['permission-roles', pageNo, pageSize, orderField, orderType, filters, queryTrigger],
    queryFn: () => getPermissionRoles(pageNo, pageSize, orderField, orderType, filters),
  })

  const menuTreeQuery = useQuery({
    queryKey: ['permission-menu-tree-for-role'],
    queryFn: () => getPermissionMenuTree(),
  })

  const saveRoleMutation = useMutation({
    mutationFn: async (values: RoleFormValues) => {
      const payload: CreatePermissionRolePayload = {
        roleName: values.roleName.trim(),
        roleType: values.roleType,
      }
      if (editorMode === 'create' || editorMode === 'copy') {
        await createPermissionRole(payload)
        return
      }
      if (!selected) return
      await updatePermissionRole(selected.id, payload)
    },
    onSuccess: async () => {
      void messageApi.success(t(editorMode === 'create' ? '新建角色成功' : editorMode === 'copy' ? '复制角色成功' : '编辑角色成功'))
      setRoleModalOpen(false)
      await rolesQuery.refetch()
    },
  })

  const deleteRoleMutation = useMutation({
    mutationFn: async (id: number) => {
      await deletePermissionRole(id)
    },
    onSuccess: async () => {
      void messageApi.success(t('角色删除成功'))
      await rolesQuery.refetch()
    },
  })

  const saveRoleMenusMutation = useMutation({
    mutationFn: async () => {
      if (!selected) return
      if (isRootAdminRole(selected)) {
        throw new Error('super role menu readonly')
      }
      const normalized = Array.from(
        new Set(
          checkedMenuKeys
            .map((key) => Number(key))
            .filter((value) => Number.isInteger(value) && value > 0),
        ),
      )
      await updatePermissionRoleMenus(selected.id, normalized)
    },
    onSuccess: async () => {
      void messageApi.success(t('角色菜单权限更新成功'))
      setPermissionDrawerOpen(false)
    },
  })

  const openRoleModal = async (mode: 'create' | 'edit' | 'copy', row?: PermissionRole) => {
    if (mode === 'edit' && isSystemRole(row)) {
      void messageApi.warning(t('系统角色不可编辑'))
      return
    }
    setEditorMode(mode)
    setSelected(row ?? null)
    if (!row) {
      roleForm.setFieldsValue({ roleName: '', roleType: 'custom' })
      setRoleModalOpen(true)
      return
    }
    const detail = await getPermissionRole(row.id)
    roleForm.setFieldsValue({
      roleName: mode === 'copy' ? t('{name}-副本', { name: detail.roleName }) : detail.roleName,
      roleType: detail.roleType,
    })
    setRoleModalOpen(true)
  }

  const openPermissionDrawer = async (row: PermissionRole) => {
    if (isRootAdminRole(row)) {
      void messageApi.warning(t('超管角色权限不可修改'))
      return
    }
    setSelected(row)
    const menuIds = await getPermissionRoleMenus(row.id)
    setCheckedMenuKeys(menuIds.map((id) => String(id)))
    setPermissionDrawerOpen(true)
  }

  const applyFilters = (values: FilterFormValues) => {
    setFilters({
      keyword: values.keyword?.trim() || undefined,
      roleType: values.roleType,
    })
    setPageNo(1)
    setQueryTrigger((current) => current + 1)
  }

  return (
    <>
      {contextHolder}
      {modalContextHolder}

      <Space direction="vertical" size={16} className="full-width table-scroll-page">
      <Space wrap>
        <UiButton type="primary" onClick={() => void openRoleModal('create')}>
          {t('新建角色')}
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
          <Form.Item label={t('角色名称')} name="keyword">
            <Input placeholder={t('请输入角色名称')} allowClear onPressEnter={() => filterForm.submit()} />
          </Form.Item>
          <Form.Item label={t('角色类型')} name="roleType">
            <Select
              style={{ width: 160 }}
              allowClear
              options={[
                { value: 'system', label: t('系统角色') },
                { value: 'custom', label: t('自定义角色') },
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
        <Table<PermissionRole>
          rowKey="id"
          loading={rolesQuery.isLoading || rolesQuery.isFetching}
          dataSource={rolesQuery.data?.items ?? []}
          scroll={{ x: 'max-content', y: 392 }}
          pagination={{
            current: pageNo,
            pageSize,
            total: rolesQuery.data?.total ?? 0,
            showSizeChanger: true,
            showTotal: (total) => t('共 {total} 条', { total }),
          }}
          onChange={(
            pagination: TablePaginationConfig,
            _filters,
            sorter: SorterResult<PermissionRole> | SorterResult<PermissionRole>[],
          ) => {
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
            const map: Record<string, 'id' | 'role_name' | 'role_type' | 'updated_at' | 'users'> = {
              id: 'id',
              roleName: 'role_name',
              roleType: 'role_type',
              users: 'users',
              updatedAt: 'updated_at',
            }
            const field = map[String(singleSorter.field)]
            if (!field) {
              setOrderField(undefined)
              setOrderType(undefined)
              return
            }
            setOrderField(field)
            setOrderType(singleSorter.order === 'ascend' ? 'asc' : 'desc')
          }}
          columns={[
            { title: t('角色ID'), dataIndex: 'id', sorter: true },
            { title: t('角色名称'), dataIndex: 'roleName', sorter: true },
            { title: t('角色类型'), dataIndex: 'roleType', sorter: true, render: (roleType: PermissionRole['roleType']) => <Tag color={roleType === 'system' ? 'blue' : 'default'}>{roleType === 'system' ? t('系统角色') : t('自定义角色')}</Tag> },
            { title: t('绑定用户数'), dataIndex: 'users', sorter: true },
            {
              title: t('更新时间'),
              dataIndex: 'updatedAt',
              sorter: true,
              render: (value: string) => formatDateTime(value, systemTimezone),
            },
            {
              title: t('操作'),
              width: 320,
              fixed: 'right',
              render: (_, row) => (
                <Space size={0}>
                  <UiButton type="link" disabled={isRootAdminRole(row)} onClick={() => void openPermissionDrawer(row)}>
                    {t('配置菜单')}
                  </UiButton>
                  {isSystemRole(row) ? (
                    <Tooltip title={t('系统角色不可编辑')}>
                      <span>
                        <UiButton type="link" disabled>
                          {t('编辑')}
                        </UiButton>
                      </span>
                    </Tooltip>
                  ) : (
                    <UiButton type="link" onClick={() => void openRoleModal('edit', row)}>
                      {t('编辑')}
                    </UiButton>
                  )}
                  <UiButton type="link" onClick={() => void openRoleModal('copy', row)}>
                    {t('复制')}
                  </UiButton>
                  <UiButton
                    type="link"
                    danger
                    disabled={isSystemRole(row)}
                    onClick={() => {
                      const boundUsers = Number.isFinite(row.users) ? row.users : 0
                      void modalApi.confirm({
                        title: t('确认删除角色 {name}？', { name: row.roleName }),
                        content: t('当前角色绑定了 {count} 个用户，删除后将强制相关用户下线并移除该角色绑定。', { count: boundUsers }),
                        okButtonProps: { danger: true },
                        onOk: async () => {
                          await deleteRoleMutation.mutateAsync(row.id)
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
        title={t(editorMode === 'create' ? '新建角色' : editorMode === 'edit' ? '编辑角色' : '复制角色')}
        open={roleModalOpen}
        onCancel={() => setRoleModalOpen(false)}
        onOk={() => roleForm.submit()}
        confirmLoading={saveRoleMutation.isPending}
      >
        <Form form={roleForm} layout="vertical" onFinish={(values) => void saveRoleMutation.mutateAsync(values)}>
          <Form.Item label={t('角色名称')} name="roleName" rules={[{ required: true, message: t('请输入角色名称') }]}>
            <Input maxLength={64} />
          </Form.Item>
          <Form.Item label={t('角色类型')} name="roleType" rules={[{ required: true, message: t('请选择角色类型') }]}>
            <Select
              options={[
                { value: 'system', label: t('系统角色') },
                { value: 'custom', label: t('自定义角色') },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        title={`${t('角色菜单权限配置')}${selected ? ` - ${selected.roleName}` : ''}`}
        open={permissionDrawerOpen}
        width={460}
        onClose={() => setPermissionDrawerOpen(false)}
        extra={
          <UiButton
            type="primary"
            disabled={isRootAdminRole(selected)}
            loading={saveRoleMenusMutation.isPending}
            onClick={() => void saveRoleMenusMutation.mutateAsync()}
          >
            {t('保存')}
          </UiButton>
        }
      >
        <Tree
          checkable
          checkStrictly
          defaultExpandAll
          checkedKeys={{ checked: checkedMenuKeys, halfChecked: [] }}
          onCheck={(keys) => {
            const newKeys = Array.isArray(keys) ? keys : keys.checked
            const newKeysStr = newKeys.map(String)
            const prevKeys = checkedMenuKeys.map(String)
            // 新增勾选：自动勾选所有下级
            const added = newKeysStr.filter((k) => !prevKeys.includes(k))
            let result = [...newKeysStr]
            for (const key of added) {
              const descendants = getDescendantKeys(menuTreeQuery.data ?? [], key)
              for (const d of descendants) {
                if (!result.includes(d)) result.push(d)
              }
            }
            // 取消勾选：自动取消所有下级
            const removed = prevKeys.filter((k) => !newKeysStr.includes(k))
            for (const key of removed) {
              const descendants = getDescendantKeys(menuTreeQuery.data ?? [], key)
              result = result.filter((k) => !descendants.includes(k))
            }
            setCheckedMenuKeys(result)
          }}
          treeData={mapTree(menuTreeQuery.data ?? [])}
          onSelect={(_, { node }) => {
            const key = String(node.key)
            const isChecked = checkedMenuKeys.map(String).includes(key)
            const descendants = getDescendantKeys(menuTreeQuery.data ?? [], key)
            if (isChecked) {
              setCheckedMenuKeys(checkedMenuKeys.filter((k) => String(k) !== key && !descendants.includes(String(k))))
            } else {
              const result = [...checkedMenuKeys.map(String), key, ...descendants]
              setCheckedMenuKeys([...new Set(result)])
            }
          }}
        />
      </Drawer>
    </>
  )
}
