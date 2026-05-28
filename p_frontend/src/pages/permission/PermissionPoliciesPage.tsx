import { useMutation, useQuery } from '@tanstack/react-query'
import { Card, Form, Input, InputNumber, Modal, Select, Space, Table, Tag, Tree, message } from 'antd'
import type { DataNode } from 'antd/es/tree'
import type { TablePaginationConfig } from 'antd'
import type { SorterResult } from 'antd/es/table/interface'
import { useMemo, useState } from 'react'
import { UiButton } from '../../components/ui'
import { useI18n } from '../../i18n/messages'
import { useUiSettingsStore } from '../../store/uiSettings'
import { formatDateTime } from '../../utils/timezone'
import {
  createPermissionMenu,
  deletePermissionMenu,
  getPermissionMenu,
  getPermissionMenus,
  getPermissionMenuTree,
  syncPermissionMenus,
  updatePermissionMenu,
  type CreatePermissionMenuPayload,
  type PermissionMenu,
  type PermissionMenuTreeNode,
  type PermissionMenusFilters,
} from '../../services/api/permission'

type FilterFormValues = {
  keyword?: string
  menuType?: 'directory' | 'menu' | 'button'
  deleted?: 'yes' | 'no'
}

type MenuFormValues = {
  parentId: number
  name: string
  routePath: string
  componentPath: string
  menuType: 'directory' | 'menu' | 'button'
  permissionKey: string
  sort: number
}

function menuTreeToAntTree(nodes: PermissionMenuTreeNode[]): DataNode[] {
  return nodes.map((node) => ({
    key: String(node.id),
    title: node.name,
    children: menuTreeToAntTree(node.children),
  }))
}

function parseMenuIDFromTreeKey(key: unknown): number {
  if (typeof key === 'number') return Number.isFinite(key) ? key : 0
  if (typeof key !== 'string') return 0
  const trimmed = key.trim()
  if (!trimmed) return 0
  const direct = Number(trimmed)
  return Number.isFinite(direct) ? direct : 0
}

function flattenTree(nodes: PermissionMenuTreeNode[]): Array<{ id: number; name: string }> {
  const result: Array<{ id: number; name: string }> = [{ id: 0, name: '根节点' }]
  const walk = (items: PermissionMenuTreeNode[]) => {
    for (const item of items) {
      result.push({ id: item.id, name: item.name })
      if (item.children.length > 0) {
        walk(item.children)
      }
    }
  }
  walk(nodes)
  return result
}

export function PermissionPoliciesPage() {
  const { t } = useI18n()
  const systemTimezone = useUiSettingsStore((state) => state.systemTimezone)
  const [messageApi, contextHolder] = message.useMessage()
  const [modalApi, modalContextHolder] = Modal.useModal()
  const [filterForm] = Form.useForm<FilterFormValues>()
  const [menuForm] = Form.useForm<MenuFormValues>()
  const [queryTrigger, setQueryTrigger] = useState(0)
  const [filters, setFilters] = useState<PermissionMenusFilters>({ deleted: 'no' })
  const [pageNo, setPageNo] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [orderField, setOrderField] = useState<'id' | 'name' | 'menu_type' | 'sort' | 'updated_at' | undefined>()
  const [orderType, setOrderType] = useState<'asc' | 'desc' | undefined>()
  const [selectedTreeId, setSelectedTreeId] = useState<number>(0)
  const [expandedTreeKeys, setExpandedTreeKeys] = useState<string[]>([])
  const [editingRow, setEditingRow] = useState<PermissionMenu | null>(null)
  const [editorMode, setEditorMode] = useState<'create' | 'edit'>('create')
  const [editorOpen, setEditorOpen] = useState(false)
  const selectedFilterTreeId = selectedTreeId > 0 ? selectedTreeId : undefined

  const treeQuery = useQuery({
    queryKey: ['permission-menu-tree'],
    queryFn: () => getPermissionMenuTree(),
  })
  const treeData = useMemo(() => menuTreeToAntTree(treeQuery.data ?? []), [treeQuery.data])

  const menusQuery = useQuery({
    queryKey: ['permission-menus', pageNo, pageSize, orderField, orderType, filters, selectedFilterTreeId, queryTrigger],
    queryFn: () =>
      getPermissionMenus(pageNo, pageSize, orderField, orderType, {
        ...filters,
        treeNodeId: selectedFilterTreeId,
      }),
  })

  const menuOptions = useMemo(
    () => flattenTree(treeQuery.data ?? []).map((item) => ({ value: item.id, label: item.id === 0 ? t('根节点') : item.name })),
    [treeQuery.data, t],
  )
  const openCreateModal = (menuType: 'directory' | 'menu' | 'button') => {
    setEditorMode('create')
    setEditingRow(null)
    menuForm.setFieldsValue({
      parentId: selectedFilterTreeId ?? 0,
      name: '',
      routePath: '',
      componentPath: '',
      menuType,
      permissionKey: '',
      sort: 0,
    })
    setEditorOpen(true)
  }

  const openEditModal = async (row: PermissionMenu) => {
    const detail = await getPermissionMenu(row.id)
    setEditorMode('edit')
    setEditingRow(row)
    menuForm.setFieldsValue({
      parentId: detail.parentId,
      name: detail.name,
      routePath: detail.routePath,
      componentPath: detail.componentPath,
      menuType: detail.menuType,
      permissionKey: detail.permissionKey,
      sort: detail.sort,
    })
    setEditorOpen(true)
  }

  const saveMutation = useMutation({
    mutationFn: async (values: MenuFormValues) => {
      const payload: CreatePermissionMenuPayload = {
        parentId: values.parentId,
        name: values.name.trim(),
        routePath: values.routePath.trim(),
        componentPath: values.componentPath.trim(),
        menuType: values.menuType,
        permissionKey: values.permissionKey.trim(),
        sort: values.sort,
      }
      if (editorMode === 'create') {
        await createPermissionMenu(payload)
        return
      }
      if (!editingRow) return
      await updatePermissionMenu(editingRow.id, payload)
    },
    onSuccess: async () => {
      void messageApi.success(t(editorMode === 'create' ? '新增菜单成功' : '编辑菜单成功'))
      setEditorOpen(false)
      await Promise.all([menusQuery.refetch(), treeQuery.refetch()])
    },
  })

  const syncMutation = useMutation({
    mutationFn: async () => syncPermissionMenus(),
    onSuccess: async () => {
      void messageApi.success(t('菜单同步完成'))
      await Promise.all([menusQuery.refetch(), treeQuery.refetch()])
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => {
      await deletePermissionMenu(id)
    },
    onSuccess: async () => {
      void messageApi.success(t('删除菜单成功'))
      await Promise.all([menusQuery.refetch(), treeQuery.refetch()])
    },
  })

  const applyFilters = (values: FilterFormValues) => {
    setFilters({
      keyword: values.keyword?.trim() || undefined,
      menuType: values.menuType,
      deleted: values.deleted ?? 'no',
    })
    setPageNo(1)
    setQueryTrigger((current) => current + 1)
  }

  return (
    <>
      {contextHolder}
      {modalContextHolder}

      <Space direction="vertical" size={16} className="full-width permission-menu-page table-scroll-page">
        <Space wrap>
          <UiButton type="primary" onClick={() => openCreateModal('menu')}>
            {t('新增菜单')}
          </UiButton>
          <UiButton onClick={() => openCreateModal('button')}>{t('新增按钮权限')}</UiButton>
          <UiButton loading={syncMutation.isPending} onClick={() => void syncMutation.mutateAsync()}>
            {t('同步前端路由')}
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
            <Form.Item label={t('关键词')} name="keyword">
              <Input placeholder={t('菜单名称/权限标识')} allowClear onPressEnter={() => filterForm.submit()} />
            </Form.Item>
            <Form.Item label={t('菜单类型')} name="menuType">
              <Select
                allowClear
                style={{ width: 140 }}
                options={[
                  { value: 'directory', label: t('目录') },
                  { value: 'menu', label: t('菜单') },
                  { value: 'button', label: t('按钮') },
                ]}
              />
            </Form.Item>
            <Form.Item label={t('是否删除')} name="deleted" initialValue="no">
              <Select
                style={{ width: 120 }}
                options={[
                  { value: 'no', label: t('否') },
                  { value: 'yes', label: t('是') },
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
                    setFilters({ deleted: 'no' })
                    setSelectedTreeId(0)
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

        <div className="permission-menu-grid table-scroll-region">
          <div className="permission-menu-tree-column">
            <Card title={t('菜单树')} className="permission-menu-tree-card">
              <div className="permission-menu-tree-panel">
                <Tree
                  blockNode
                  expandAction="click"
                  virtual={false}
                  style={{ width: '100%' }}
                  selectedKeys={selectedTreeId > 0 ? [String(selectedTreeId)] : []}
                  expandedKeys={expandedTreeKeys}
                  onExpand={(keys) => setExpandedTreeKeys(keys.map(String))}
                  onSelect={(_, info) => {
                    setSelectedTreeId(info.selected ? parseMenuIDFromTreeKey(info.node.key) : 0)
                    setPageNo(1)
                  }}
                  treeData={treeData}
                />
              </div>
            </Card>
          </div>
          <div className="permission-menu-list-column">
            <Card title={t('菜单权限列表')} className="permission-menu-list-card compact-table-card system-table-card">
              <Table<PermissionMenu>
                rowKey="id"
                loading={menusQuery.isLoading || menusQuery.isFetching}
                dataSource={menusQuery.data?.items ?? []}
                scroll={{ x: 'max-content', y: 368 }}
                pagination={{
                  current: pageNo,
                  pageSize,
                  total: menusQuery.data?.total ?? 0,
                  showSizeChanger: true,
                  showTotal: (total) => t('共 {total} 条', { total }),
                }}
                onChange={(
                  pagination: TablePaginationConfig,
                  _filters,
                  sorter: SorterResult<PermissionMenu> | SorterResult<PermissionMenu>[],
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
                  const map: Record<string, 'id' | 'name' | 'menu_type' | 'sort' | 'updated_at'> = {
                    id: 'id',
                    name: 'name',
                    menuType: 'menu_type',
                    sort: 'sort',
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
                  { title: t('菜单ID'), dataIndex: 'id', sorter: true },
                  { title: t('菜单名称'), dataIndex: 'name', sorter: true },
                  { title: t('路由Path'), dataIndex: 'routePath', width: 220, ellipsis: true },
                  { title: t('组件Path'), dataIndex: 'componentPath', width: 260, ellipsis: true },
                  { title: t('菜单类型'), dataIndex: 'menuType', sorter: true, render: (menuType: PermissionMenu['menuType']) => menuType === 'directory' ? <Tag>{t('目录')}</Tag> : menuType === 'menu' ? <Tag color="blue">{t('菜单')}</Tag> : <Tag color="orange">{t('按钮')}</Tag> },
                  { title: t('权限标识'), dataIndex: 'permissionKey', width: 280, ellipsis: true },
                  { title: t('排序'), dataIndex: 'sort', sorter: true },
                  {
                    title: t('是否删除'),
                    dataIndex: 'deleted',
                    width: 110,
                    render: (deleted: boolean) => (
                      <Tag color={deleted ? 'red' : 'green'}>{deleted ? t('是') : t('否')}</Tag>
                    ),
                  },
                  {
                    title: t('删除时间'),
                    dataIndex: 'deletedAt',
                    width: 170,
                    render: (value: string) => (value ? formatDateTime(value, systemTimezone) : '-'),
                  },
                  {
                    title: t('更新时间'),
                    dataIndex: 'updatedAt',
                    sorter: true,
                    render: (value: string) => formatDateTime(value, systemTimezone),
                  },
                  {
                    title: t('操作'),
                    width: 180,
                    fixed: 'right',
                    render: (_, row) => (
                      <Space size={0}>
                        {!row.deleted ? (
                          <UiButton type="link" onClick={() => void openEditModal(row)}>
                            {t('编辑')}
                          </UiButton>
                        ) : null}
                        <UiButton
                          type="link"
                          danger
                          onClick={() => {
                            void modalApi.confirm({
                              title: row.deleted
                                ? t('确认彻底删除菜单 {name}', { name: row.name })
                                : t('确认删除菜单 {name}', { name: row.name }),
                              content: row.deleted
                                ? t('彻底删除不可恢复，且只能删除已删除超过1小时的菜单。')
                                : t('如果存在子菜单，需先删除子菜单。'),
                              okButtonProps: { danger: true },
                              onOk: async () => {
                                await deleteMutation.mutateAsync(row.id)
                              },
                            })
                          }}
                        >
                          {row.deleted ? t('彻底删除') : t('删除')}
                        </UiButton>
                      </Space>
                    ),
                  },
                ]}
              />
            </Card>
          </div>
        </div>
      </Space>

      <Modal
        title={t(editorMode === 'create' ? '新增菜单权限' : '编辑菜单权限')}
        open={editorOpen}
        onCancel={() => setEditorOpen(false)}
        onOk={() => menuForm.submit()}
        confirmLoading={saveMutation.isPending}
      >
        <Form form={menuForm} layout="vertical" onFinish={(values) => void saveMutation.mutateAsync(values)}>
          <Form.Item label={t('父级菜单')} name="parentId" rules={[{ required: true, message: t('请选择父级菜单') }]}>
            <Select options={menuOptions} showSearch optionFilterProp="label" />
          </Form.Item>
          <Form.Item label={t('菜单名称')} name="name" rules={[{ required: true, message: t('请输入菜单名称') }]}>
            <Input maxLength={64} />
          </Form.Item>
          <Form.Item label={t('菜单类型')} name="menuType" rules={[{ required: true, message: t('请选择菜单类型') }]}>
            <Select
              options={[
                { value: 'directory', label: t('目录') },
                { value: 'menu', label: t('菜单') },
                { value: 'button', label: t('按钮') },
              ]}
            />
          </Form.Item>
          <Form.Item label={t('路由 Path')} name="routePath">
            <Input maxLength={255} />
          </Form.Item>
          <Form.Item label={t('组件 Path')} name="componentPath">
            <Input maxLength={255} />
          </Form.Item>
          <Form.Item label={t('权限标识')} name="permissionKey">
            <Input maxLength={128} />
          </Form.Item>
          <Form.Item label={t('排序')} name="sort" rules={[{ required: true, message: t('请输入排序值') }]}>
            <InputNumber min={0} max={100000} precision={0} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
