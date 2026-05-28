import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Card, Drawer, Form, Input, Modal, Space, Table, Tag, Tree, Typography, message } from 'antd'
import type { DataNode } from 'antd/es/tree'
import { useMemo, useState } from 'react'
import {
  computeDepartmentStats,
  createOrganizationDepartment,
  deleteOrganizationDepartment,
  getOrganizationDepartment,
  getOrganizationDepartmentsTree,
  toggleOrganizationDepartmentStatus,
  updateOrganizationDepartment,
  type OrganizationDepartment,
} from '../../services/api/organization'
import { UiButton } from '../../components/ui'
import { useI18n } from '../../i18n/messages'
import { useUiSettingsStore } from '../../store/uiSettings'
import { formatDateTime } from '../../utils/timezone'

type DepartmentFormValues = {
  name: string
  code: string
}

type DepartmentLevelStats = {
  level1: number
  level2: number
  level3: number
  level4: number
}

function parseDepartmentIDFromTreeKey(key: unknown): number {
  if (typeof key === 'number') return Number.isFinite(key) ? key : 0
  if (typeof key !== 'string') return 0
  const trimmed = key.trim()
  if (!trimmed) return 0
  const direct = Number(trimmed)
  if (Number.isFinite(direct)) return direct
  const matched = trimmed.match(/\d+/g)
  if (!matched || matched.length === 0) return 0
  const fallback = Number(matched[matched.length - 1])
  return Number.isFinite(fallback) ? fallback : 0
}

function toTreeNodes(nodes: OrganizationDepartment[]): DataNode[] {
  return nodes.map((node) => ({
    key: String(node.id),
    title: node.name,
    children: toTreeNodes(node.children),
  }))
}

function findNodeByID(nodes: OrganizationDepartment[], id: number): OrganizationDepartment | null {
  for (const node of nodes) {
    if (node.id === id) return node
    const found = findNodeByID(node.children, id)
    if (found) return found
  }
  return null
}

function hasDepartmentChildren(nodes: OrganizationDepartment[], id: number): boolean {
  return (findNodeByID(nodes, id)?.children.length ?? 0) > 0
}

function listChildrenByParentID(nodes: OrganizationDepartment[], parentID: number): OrganizationDepartment[] {
  const parent = findNodeByID(nodes, parentID)
  return parent?.children ?? []
}

function buildDepartmentMap(nodes: OrganizationDepartment[]): Map<number, OrganizationDepartment> {
  const map = new Map<number, OrganizationDepartment>()
  const walk = (items: OrganizationDepartment[]) => {
    for (const item of items) {
      map.set(item.id, item)
      if (item.children.length > 0) {
        walk(item.children)
      }
    }
  }
  walk(nodes)
  return map
}

function computeDepartmentTreeLevelStats(nodes: OrganizationDepartment[]): DepartmentLevelStats {
  const stats: DepartmentLevelStats = { level1: 0, level2: 0, level3: 0, level4: 0 }
  const walk = (items: OrganizationDepartment[], level: number) => {
    for (const item of items) {
      if (level === 1) stats.level1 += 1
      if (level === 2) stats.level2 += 1
      if (level === 3) stats.level3 += 1
      if (level === 4) stats.level4 += 1
      if (item.children.length > 0) {
        walk(item.children, level + 1)
      }
    }
  }
  walk(nodes, 1)
  return stats
}

function getDepartmentLevel(node: OrganizationDepartment, departmentMap: Map<number, OrganizationDepartment>): number {
  let level = 1
  let currentParentID = node.parentId
  while (currentParentID > 0) {
    const parent = departmentMap.get(currentParentID)
    if (!parent) break
    level += 1
    currentParentID = parent.parentId
  }
  return level
}

export function OrganizationStructurePage() {
  const { t } = useI18n()
  const systemTimezone = useUiSettingsStore((state) => state.systemTimezone)
  const queryClient = useQueryClient()
  const [messageApi, contextHolder] = message.useMessage()
  const [modalApi, modalContextHolder] = Modal.useModal()
  const [form] = Form.useForm<DepartmentFormValues>()

  const [selectedID, setSelectedID] = useState<number>(0)
  const [expandedTreeKeys, setExpandedTreeKeys] = useState<string[]>([])
  const [detailDepartmentID, setDetailDepartmentID] = useState<number>(0)
  const [formMode, setFormMode] = useState<'create-root' | 'create-child' | 'edit'>('create-root')
  const [modalOpen, setModalOpen] = useState(false)

  const treeQuery = useQuery({
    queryKey: ['organization-departments-tree'],
    queryFn: () => getOrganizationDepartmentsTree(),
  })

  const selectedNode = useMemo(() => {
    if (!selectedID) return null
    return findNodeByID(treeQuery.data ?? [], selectedID)
  }, [selectedID, treeQuery.data])

  const detailQuery = useQuery({
    queryKey: ['organization-department-detail', detailDepartmentID],
    enabled: detailDepartmentID > 0,
    queryFn: () => getOrganizationDepartment(detailDepartmentID),
  })

  const createMutation = useMutation({
    mutationFn: async (values: DepartmentFormValues) => {
      const parentID = formMode === 'create-child' ? selectedID : 0
      await createOrganizationDepartment({
        parentId: parentID,
        name: values.name,
        code: values.code,
      })
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['organization-departments-tree'] })
      if (selectedID > 0) {
        await queryClient.invalidateQueries({ queryKey: ['organization-department-detail', selectedID] })
      }
      void messageApi.success(t('部门创建成功'))
      setModalOpen(false)
    },
  })

  const updateMutation = useMutation({
    mutationFn: async (values: DepartmentFormValues) => {
      if (selectedID <= 0) return
      await updateOrganizationDepartment(selectedID, {
        name: values.name,
        code: values.code,
      })
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['organization-departments-tree'] })
      if (selectedID > 0) {
        await queryClient.invalidateQueries({ queryKey: ['organization-department-detail', selectedID] })
      }
      void messageApi.success(t('部门更新成功'))
      setModalOpen(false)
    },
  })

  const toggleStatusMutation = useMutation({
    mutationFn: async () => {
      if (!selectedNode) return
      await toggleOrganizationDepartmentStatus(selectedNode.id, selectedNode.status !== 'enabled')
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['organization-departments-tree'] })
      if (selectedID > 0) {
        await queryClient.invalidateQueries({ queryKey: ['organization-department-detail', selectedID] })
      }
      void messageApi.success(t('部门状态已更新'))
    },
  })

  const deleteMutation = useMutation<void, Error, boolean>({
    mutationFn: async (force) => {
      if (selectedID <= 0) return
      await deleteOrganizationDepartment(selectedID, force)
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['organization-departments-tree'] })
      setSelectedID(0)
      void messageApi.success(t('部门删除成功'))
    },
  })

  const treeData = useMemo(() => toTreeNodes(treeQuery.data ?? []), [treeQuery.data])
  const departmentMap = useMemo(() => buildDepartmentMap(treeQuery.data ?? []), [treeQuery.data])
  const departmentLevelStats = useMemo(() => computeDepartmentTreeLevelStats(treeQuery.data ?? []), [treeQuery.data])
  const childRows = useMemo(() => {
    if (!selectedID) return treeQuery.data ?? []
    return listChildrenByParentID(treeQuery.data ?? [], selectedID)
  }, [selectedID, treeQuery.data])
  const departmentTreeSummary = useMemo(() => {
    const parts: string[] = []
    if (departmentLevelStats.level1 > 0) {
      parts.push(t('共 {count} 个1级部门', { count: departmentLevelStats.level1 }))
    }
    if (departmentLevelStats.level2 > 0) {
      parts.push(t('{count} 个2级部门', { count: departmentLevelStats.level2 }))
    }
    if (departmentLevelStats.level3 > 0) {
      parts.push(t('{count} 个3级部门', { count: departmentLevelStats.level3 }))
    }
    if (departmentLevelStats.level4 > 0) {
      parts.push(t('{count} 个4级部门', { count: departmentLevelStats.level4 }))
    }
    return parts.join('，')
  }, [departmentLevelStats.level1, departmentLevelStats.level2, departmentLevelStats.level3, departmentLevelStats.level4, t])

  const openCreateRoot = () => {
    setFormMode('create-root')
    form.resetFields()
    setModalOpen(true)
  }

  const openCreateChild = () => {
    if (!selectedNode) {
      void messageApi.warning(t('请先在左侧选择父部门'))
      return
    }
    const selectedLevel = getDepartmentLevel(selectedNode, departmentMap)
    if (selectedLevel >= 4) {
      void messageApi.warning(t('最多只支持4级部门，当前部门下不可继续新增下级'))
      return
    }
    setFormMode('create-child')
    form.resetFields()
    setModalOpen(true)
  }

  const openEdit = () => {
    if (!selectedNode) {
      void messageApi.warning(t('请先选择部门'))
      return
    }
    setFormMode('edit')
    form.setFieldsValue({
      name: selectedNode.name,
      code: selectedNode.code,
    })
    setModalOpen(true)
  }

  const openDetail = () => {
    if (!selectedNode) {
      void messageApi.warning(t('请先选择部门'))
      return
    }
    setDetailDepartmentID(selectedNode.id)
  }

  const submitForm = async () => {
    const values = await form.validateFields()
    if (formMode === 'edit') {
      await updateMutation.mutateAsync(values)
      return
    }
    await createMutation.mutateAsync(values)
  }

  const confirmToggleStatus = () => {
    if (!selectedNode) {
      void messageApi.warning(t('请先选择部门'))
      return
    }
    const actionName = selectedNode.status === 'enabled' ? t('停用') : t('启用')
    void modalApi.confirm({
      title: t('确认{action}部门', { action: actionName }),
      content: t('{action}后将影响该部门在组织架构中的可见状态。', { action: actionName }),
      onOk: async () => toggleStatusMutation.mutateAsync(),
    })
  }

  const confirmDelete = () => {
    if (!selectedNode) {
      void messageApi.warning(t('请先选择部门'))
      return
    }
    void modalApi.confirm({
      title: t('确认删除部门'),
      content: t('若部门下存在成员则无法删除；若仅存在下级部门或岗位，将二次确认后继续删除。'),
      onOk: async () => {
        try {
          await deleteMutation.mutateAsync(false)
        } catch (error) {
          const message = error instanceof Error ? error.message : String(error)
          if (!message.includes('force=true')) {
            throw error
          }
          await modalApi.confirm({
            title: t('该部门存在下级部门或岗位，是否继续删除？'),
            content: t('继续删除仅删除当前部门，不会自动删除下级部门或岗位。'),
            okButtonProps: { danger: true },
            onOk: async () => deleteMutation.mutateAsync(true),
          })
        }
      },
    })
  }

  const modalTitle =
    formMode === 'edit'
      ? t('编辑部门')
      : formMode === 'create-child'
        ? t('新增下级部门（父级：{name}）', { name: selectedNode?.name ?? '-' })
        : t('新增一级部门')

  return (
    <>
      {contextHolder}
      {modalContextHolder}

      <Space direction="vertical" size={16} className="full-width organization-structure-page table-scroll-page">
      <Space wrap>
        <UiButton type="primary" onClick={openCreateRoot}>
          {t('新增一级部门')}
        </UiButton>
        <UiButton onClick={openCreateChild}>{t('新增下级部门')}</UiButton>
        <UiButton onClick={openDetail}>{t('查看详情')}</UiButton>
        <UiButton onClick={openEdit}>{t('编辑部门')}</UiButton>
        <UiButton onClick={confirmToggleStatus}>{t('停用/启用')}</UiButton>
        <UiButton danger onClick={confirmDelete}>
          {t('删除部门')}
        </UiButton>
      </Space>

      <div className="organization-structure-grid table-scroll-region">
        <div className="organization-structure-tree-column">
          <Card
            className="organization-structure-tree-card"
            title={
              <Space size={6} wrap className="organization-structure-tree-title">
                <span>{t('部门树')}</span>
                {departmentTreeSummary ? (
                  <Typography.Text type="secondary" className="organization-structure-tree-summary">
                    {departmentTreeSummary}
                  </Typography.Text>
                ) : null}
              </Space>
            }
            loading={treeQuery.isLoading}
          >
            <div className="organization-structure-tree-panel">
              <Tree
                selectedKeys={selectedID ? [String(selectedID)] : []}
                expandedKeys={expandedTreeKeys}
                treeData={treeData}
                blockNode
                onExpand={(keys) => setExpandedTreeKeys(keys.map(String))}
                onSelect={(_, info) => {
                  if (!info) {
                    setSelectedID(0)
                    return
                  }
                  const nextID = parseDepartmentIDFromTreeKey(info.node.key)
                  if (hasDepartmentChildren(treeQuery.data ?? [], nextID)) {
                    const nextKey = String(nextID)
                    setExpandedTreeKeys((prev) => (prev.includes(nextKey) ? prev.filter((key) => key !== nextKey) : [...prev, nextKey]))
                  }
                  setSelectedID(info.selected ? nextID : 0)
                }}
              />
            </div>
          </Card>
        </div>

        <div className="organization-structure-list-column">
          <Card
            title={selectedNode ? t('下级部门（{name}）', { name: selectedNode.name }) : t('一级部门列表')}
            className="organization-structure-list-card compact-table-card system-table-card"
          >
            <Table<OrganizationDepartment>
              className="organization-structure-table"
              rowKey="id"
              dataSource={childRows}
              loading={treeQuery.isLoading}
              pagination={false}
              scroll={{ y: 392 }}
              tableLayout="fixed"
              columns={[
                { title: t('部门名称'), dataIndex: 'name', width: 150, ellipsis: true },
                { title: t('部门编码'), dataIndex: 'code', width: 120, ellipsis: true },
                {
                  title: t('直属岗位数量'),
                  width: 115,
                  render: (_, row) => {
                    const source = departmentMap.get(row.id) ?? row
                    return computeDepartmentStats(source).directPositionCount
                  },
                },
                {
                  title: t('所有岗位数量'),
                  width: 115,
                  render: (_, row) => {
                    const source = departmentMap.get(row.id) ?? row
                    return computeDepartmentStats(source).totalPositionCount
                  },
                },
                {
                  title: t('直属岗位成员数量'),
                  width: 125,
                  render: (_, row) => {
                    const source = departmentMap.get(row.id) ?? row
                    return computeDepartmentStats(source).directMemberCount
                  },
                },
                {
                  title: t('全部成员数量'),
                  width: 115,
                  render: (_, row) => {
                    const source = departmentMap.get(row.id) ?? row
                    return computeDepartmentStats(source).totalMemberCount
                  },
                },
                {
                  title: t('状态'),
                  dataIndex: 'status',
                  width: 100,
                  render: (status: OrganizationDepartment['status']) => (
                    <Tag color={status === 'enabled' ? 'green' : 'default'}>{status === 'enabled' ? t('启用') : t('停用')}</Tag>
                  ),
                },
                {
                  title: t('更新时间'),
                  dataIndex: 'updatedAt',
                  width: 170,
                  fixed: 'right',
                  render: (value: string) => formatDateTime(value, systemTimezone),
                },
                {
                  title: t('操作'),
                  width: 80,
                  fixed: 'right',
                  render: (_, row) => (
                    <UiButton
                      type="link"
                      onClick={(event) => {
                        event.stopPropagation()
                        setDetailDepartmentID(row.id)
                      }}
                    >
                      {t('详情')}
                    </UiButton>
                  ),
                },
              ]}
            />
          </Card>
        </div>
      </div>
      </Space>

      <Drawer
        title={t('部门详情')}
        open={detailDepartmentID > 0}
        onClose={() => setDetailDepartmentID(0)}
        width={420}
      >
        {(() => {
          const detailSource = detailDepartmentID > 0 ? departmentMap.get(detailDepartmentID) ?? detailQuery.data : detailQuery.data
          const stats = detailSource ? computeDepartmentStats(detailSource) : null
          return (
            <Space direction="vertical" size={12} className="full-width">
              <Typography.Text>{t('部门名称：')}{detailQuery.data?.name ?? '-'}</Typography.Text>
              <Typography.Text>{t('部门编码：')}{detailQuery.data?.code ?? '-'}</Typography.Text>
              <Typography.Text>{t('直属岗位数量：')}{stats ? stats.directPositionCount : '-'}</Typography.Text>
              <Typography.Text>{t('所有岗位数量：')}{stats ? stats.totalPositionCount : '-'}</Typography.Text>
              <Typography.Text>{t('直属岗位成员数量：')}{stats ? stats.directMemberCount : '-'}</Typography.Text>
              <Typography.Text>{t('全部成员数量：')}{stats ? stats.totalMemberCount : '-'}</Typography.Text>
              <Typography.Text>
                {t('状态：')}{detailQuery.data?.status === 'enabled' ? <Tag color="green">{t('启用')}</Tag> : <Tag>{t('停用')}</Tag>}
              </Typography.Text>
              <Typography.Text>{t('更新时间：')}{formatDateTime(detailQuery.data?.updatedAt, systemTimezone)}</Typography.Text>
            </Space>
          )
        })()}
      </Drawer>

      <Modal
        title={modalTitle}
        open={modalOpen}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        onCancel={() => setModalOpen(false)}
        onOk={() => void submitForm()}
      >
        <Form form={form} layout="vertical">
          <Form.Item label={t('部门名称')} name="name" rules={[{ required: true, message: t('请输入部门名称') }]}>
            <Input maxLength={64} placeholder={t('请输入部门名称')} />
          </Form.Item>
          <Form.Item label={t('部门编码')} name="code" rules={[{ required: true, message: t('请输入部门编码') }]}>
            <Input maxLength={64} placeholder={t('请输入部门编码')} />
          </Form.Item>
          {formMode === 'create-child' ? (
            <Typography.Text type="secondary">
              {t('当前仅支持创建到4级部门。')}
            </Typography.Text>
          ) : null}
        </Form>
      </Modal>
    </>
  )
}
