import { useQuery } from '@tanstack/react-query'
import { CopyOutlined } from '@ant-design/icons'
import {
  Card,
  DatePicker,
  Drawer,
  Dropdown,
  Form,
  Input,
  Modal,
  Select,
  Space,
  Spin,
  Table,
  Tag,
  Tooltip,
  Typography,
  message,
} from 'antd'
import type { MenuProps, TablePaginationConfig } from 'antd'
import { useMemo, useState } from 'react'
import type { SorterResult } from 'antd/es/table/interface'
import { UiButton, UiProtectedBadge } from '../../components/ui'
import { RoleSummary } from '../../components/permission/RoleSummary'
import { forceLogout, type UserSessionItem } from '../../services/api/auth'
import {
  assignOrganizationUserPosition,
  createOrganizationUser,
  deleteOrganizationUser,
  exportOrganizationUsers,
  getOrganizationDepartmentsTree,
  getOrganizationPositions,
  getOrganizationUsers,
  getOrganizationUserSessions,
  importOrganizationUsers,
  batchTransferOrganizationUsers,
  type ImportOrganizationUserItem,
  type OrganizationUsersFilters,
  resetOrganizationUserPassword,
  type OrganizationUser,
  type OrganizationUserStatus,
  updateOrganizationUserProfile,
  updateOrganizationUserStatus,
} from '../../services/api/organization'
import { useAuthStore } from '../../store/auth'
import { useUiSettingsStore } from '../../store/uiSettings'
import { useI18n } from '../../i18n/messages'
import { formatDateTime, toTimezoneDateTimeString } from '../../utils/timezone'
import { hasPermission, permissionKeys } from '../../utils/permissions'

type UserRow = OrganizationUser

const USERNAME_DISPLAY_LIMIT = 15
const USERNAME_PATTERN = /^[A-Za-z][A-Za-z0-9_-]*$/

function truncateUsername(value: string) {
  const characters = Array.from(value)
  return characters.length > USERNAME_DISPLAY_LIMIT
    ? `${characters.slice(0, USERNAME_DISPLAY_LIMIT).join('')}…`
    : value
}

type UserFormValues = {
  username: string
  password?: string
  displayName: string
  email: string
  phone: string
  status: OrganizationUserStatus
  departmentId?: number
  positionId?: number
}

type FilterFormValues = {
  keyword?: string
  phone?: string
  status?: 'active' | 'disabled' | 'deactivated'
  departmentId?: number
  positionId?: number
  createdAt?: [{ toISOString: () => string }, { toISOString: () => string }]
}

type TransferPositionFormValues = {
  departmentId: number
  positionId: number
}

function maskPhone(phone: string): string {
  const value = phone.trim()
  if (value.length <= 4) return value || '-'
  if (value.length <= 7) return `${value.slice(0, 2)}****${value.slice(-2)}`
  return `${value.slice(0, 3)}****${value.slice(-4)}`
}

function maskEmail(email: string): string {
  const value = email.trim()
  if (!value) return '-'
  const [name, domain] = value.split('@')
  if (!domain) return value.length <= 4 ? '****' : `${value.slice(0, 2)}****${value.slice(-2)}`
  const safeName =
    name.length <= 2 ? `${name.slice(0, 1)}***` : `${name.slice(0, 2)}***${name.slice(-1)}`
  return `${safeName}@${domain}`
}

async function copyText(value: string): Promise<void> {
  const text = value.trim()
  if (!text) return
  await navigator.clipboard.writeText(text)
}

function flattenDepartments(
  items: Awaited<ReturnType<typeof getOrganizationDepartmentsTree>>,
): Array<{ id: number; name: string; status: 'enabled' | 'disabled'; fullPath: string }> {
  const result: Array<{
    id: number
    name: string
    status: 'enabled' | 'disabled'
    fullPath: string
  }> = []
  const walk = (nodes: typeof items, parentNames: string[]) => {
    for (const node of nodes) {
      const currentNames = [...parentNames, node.name]
      result.push({
        id: node.id,
        name: node.name,
        status: node.status,
        fullPath: currentNames.join('-'),
      })
      if (node.children.length > 0) {
        walk(node.children, currentNames)
      }
    }
  }
  walk(items, [])
  return result
}

function renderAccountStatus(
  status: UserRow['accountStatus'],
  t: (text: string, params?: Record<string, string | number>) => string,
) {
  if (status === 'active') return <Tag color="green">{t('启用')}</Tag>
  if (status === 'disabled') return <Tag color="orange">{t('停用')}</Tag>
  return <Tag color="red">{t('已注销')}</Tag>
}

function renderOnlineStatus(
  status: UserRow['onlineStatus'],
  t: (text: string, params?: Record<string, string | number>) => string,
) {
  return (
    <Tag color={status === 'online' ? 'green' : 'default'}>
      {status === 'online' ? t('在线') : t('离线')}
    </Tag>
  )
}

function isLockedUser(row: UserRow) {
  return row.accountStatus === 'deactivated' || row.accountStatus === 'disabled'
}

export function OrganizationMembersPage() {
  const { t } = useI18n()
  const [messageApi, contextHolder] = message.useMessage()
  const [modalApi, modalContextHolder] = Modal.useModal()
  const [userForm] = Form.useForm<UserFormValues>()
  const [filterForm] = Form.useForm<FilterFormValues>()
  const [transferForm] = Form.useForm<TransferPositionFormValues>()
  const currentUser = useAuthStore((state) => state.currentUser)
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const systemTimezone = useUiSettingsStore((state) => state.systemTimezone)
  const [selectedRowKeys, setSelectedRowKeys] = useState<number[]>([])
  const [formOpen, setFormOpen] = useState(false)
  const [detailOpen, setDetailOpen] = useState(false)
  const [passwordOpen, setPasswordOpen] = useState(false)
  const [transferOpen, setTransferOpen] = useState(false)
  const [transferTargetUIDs, setTransferTargetUIDs] = useState<number[]>([])
  const [formSaving, setFormSaving] = useState(false)
  const [passwordSaving, setPasswordSaving] = useState(false)
  const [transferSaving, setTransferSaving] = useState(false)
  const [importing, setImporting] = useState(false)
  const [exporting, setExporting] = useState(false)
  const [editorMode, setEditorMode] = useState<'create' | 'edit'>('create')
  const [selected, setSelected] = useState<UserRow | null>(null)
  const [sessionItems, setSessionItems] = useState<UserSessionItem[]>([])
  const [sessionLoading, setSessionLoading] = useState(false)
  const [pageNo, setPageNo] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [filters, setFilters] = useState<OrganizationUsersFilters>({})
  const [queryTrigger, setQueryTrigger] = useState(0)
  const [orderField, setOrderField] = useState<
    | 'uid'
    | 'username'
    | 'display_name'
    | 'status'
    | 'active_session_count'
    | 'last_login_at'
    | undefined
  >()
  const [orderType, setOrderType] = useState<'asc' | 'desc' | undefined>()
  const canEditProfile = hasPermission(currentUser, permissionKeys.usersEditProfile)
  const canAssignPosition = hasPermission(currentUser, permissionKeys.usersAssignPosition)
  const canChangeStatus = hasPermission(currentUser, permissionKeys.usersChangeStatus)
  const canResetPassword = hasPermission(currentUser, permissionKeys.usersResetPassword)
  const canDeleteUser = hasPermission(currentUser, permissionKeys.usersDelete)
  const usersQuery = useQuery({
    queryKey: [
      'organization-users',
      pageNo,
      pageSize,
      orderField,
      orderType,
      filters,
      queryTrigger,
    ],
    queryFn: () => getOrganizationUsers(pageNo, pageSize, orderField, orderType, filters),
    enabled: isAuthenticated && Boolean(currentUser?.uid),
  })
  const departmentsQuery = useQuery({
    queryKey: ['organization-departments-tree-for-users'],
    queryFn: () => getOrganizationDepartmentsTree(),
    enabled: isAuthenticated && Boolean(currentUser?.uid),
  })
  const positionsQuery = useQuery({
    queryKey: ['organization-positions-all-for-users'],
    queryFn: () => getOrganizationPositions(1, 200),
    enabled: isAuthenticated && Boolean(currentUser?.uid),
  })
  const selectedFormDepartmentId = Form.useWatch('departmentId', userForm)
  const selectedTransferDepartmentId = Form.useWatch('departmentId', transferForm)
  const formDepartmentId = Number(selectedFormDepartmentId || 0)
  const transferDepartmentId = Number(selectedTransferDepartmentId || 0)
  const formPositionsQuery = useQuery({
    queryKey: ['organization-positions-for-user-form', formDepartmentId],
    queryFn: () =>
      getOrganizationPositions(1, 200, undefined, undefined, {
        departmentId: formDepartmentId,
        inheritParent: true,
      }),
    enabled:
      isAuthenticated &&
      Boolean(currentUser?.uid) &&
      canAssignPosition &&
      editorMode === 'create' &&
      formOpen &&
      formDepartmentId > 0,
  })
  const transferPositionsQuery = useQuery({
    queryKey: ['organization-positions-for-user-transfer', transferDepartmentId],
    queryFn: () =>
      getOrganizationPositions(1, 200, undefined, undefined, {
        departmentId: transferDepartmentId,
        inheritParent: true,
      }),
    enabled:
      isAuthenticated &&
      Boolean(currentUser?.uid) &&
      canAssignPosition &&
      transferOpen &&
      transferDepartmentId > 0,
  })
  const rows = usersQuery.data?.items ?? []
  const flattenedDepartments = useMemo(
    () => flattenDepartments(departmentsQuery.data ?? []),
    [departmentsQuery.data],
  )
  const departmentStatusMap = new Map(flattenedDepartments.map((item) => [item.id, item.status]))
  const departmentFullPathMap = new Map(
    flattenedDepartments.map((item) => [item.id, item.fullPath]),
  )
  const positionStatusMap = new Map(
    (positionsQuery.data?.items ?? []).map((item) => [item.id, item.status]),
  )
  const departmentFilterOptions = flattenedDepartments.map((item) => ({
    value: item.id,
    label: `${item.name}${item.status === 'disabled' ? ` [${t('停用')}]` : ''}`,
  }))
  const departmentFormOptions = flattenedDepartments.map((item) => ({
    value: item.id,
    label: `${item.name}${item.status === 'disabled' ? ` [${t('停用')}]` : ''}`,
    searchText: `${item.name} ${item.fullPath}`,
    disabled: item.status === 'disabled',
  }))
  const positionFilterOptions = (positionsQuery.data?.items ?? []).map((item) => ({
    value: item.id,
    label: `${item.name}（${item.departmentName || t('未分配部门')}）${item.status === 'disabled' ? ` [${t('停用')}]` : ''}`,
    departmentId: item.departmentId,
  }))
  const scopedPositionFormOptions = useMemo(() => {
    if (formDepartmentId <= 0) {
      return []
    }
    return (formPositionsQuery.data?.items ?? []).map((item) => ({
      value: item.id,
      label: `${item.name}（${item.departmentName || t('未分配部门')}）${item.status === 'disabled' ? ` [${t('停用')}]` : !item.canAssign ? ` [${t('无权分配')}]` : ''}`,
      disabled: item.status === 'disabled' || !item.canAssign,
    }))
  }, [formDepartmentId, formPositionsQuery.data?.items, t])
  const scopedTransferPositionOptions = useMemo(() => {
    if (transferDepartmentId <= 0) return []
    return (transferPositionsQuery.data?.items ?? []).map((item) => ({
      value: item.id,
      label: `${item.name}（${item.departmentName || t('未分配部门')}）${item.status === 'disabled' ? ` [${t('停用')}]` : !item.canAssign ? ` [${t('无权分配')}]` : ''}`,
      disabled: item.status === 'disabled' || !item.canAssign,
    }))
  }, [transferDepartmentId, transferPositionsQuery.data?.items, t])

  const forceLogoutUsers = async (uids: number[]) => {
    const filteredUIDs = uids.filter((uid) => {
      const row = rows.find((item) => item.uid === uid)
      if (!row) return false
      return canChangeStatus && row.canManage
    })
    const skipped = uids.length - filteredUIDs.length
    if (filteredUIDs.length === 0) {
      void messageApi.warning(t('没有可强制下线的用户'))
      return
    }
    const settled = await Promise.allSettled(filteredUIDs.map((uid) => forceLogout(uid)))
    const success = settled.filter((item) => item.status === 'fulfilled').length
    const failed = settled.length - success
    await usersQuery.refetch()
    if (success > 0) {
      void messageApi.success(t('强制下线成功 {count} 人', { count: success }))
    }
    if (failed > 0) {
      void messageApi.warning(t('强制下线失败 {count} 人，请重试', { count: failed }))
    }
    if (skipped > 0) {
      void messageApi.info(t('已自动跳过 {count} 个受保护用户', { count: skipped }))
    }
    setSelectedRowKeys([])
  }

  const deactivateUsers = async (uids: number[]) => {
    const filteredUIDs = uids.filter((uid) => {
      const row = rows.find((item) => item.uid === uid)
      if (!row) return false
      return canChangeStatus && row.canManage
    })
    const skipped = uids.length - filteredUIDs.length
    if (filteredUIDs.length === 0) {
      void messageApi.warning(t('没有可注销的用户'))
      return
    }
    const settled = await Promise.allSettled(
      filteredUIDs.map((uid) => updateOrganizationUserStatus(uid, 2)),
    )
    const success = settled.filter((item) => item.status === 'fulfilled').length
    const failed = settled.length - success
    await usersQuery.refetch()
    if (success > 0) {
      void messageApi.success(t('注销账号成功 {count} 人', { count: success }))
    }
    if (failed > 0) {
      void messageApi.warning(t('注销账号失败 {count} 人，请重试', { count: failed }))
    }
    if (skipped > 0) {
      void messageApi.info(t('已自动跳过 {count} 个受保护用户', { count: skipped }))
    }
    setSelectedRowKeys([])
  }

  const buildMoreActions = (row: UserRow): MenuProps['items'] => [
    {
      key: 'assign-position',
      label: t('调整岗位'),
      disabled: !canAssignPosition || !row.canManage || row.accountStatus === 'deactivated',
    },
    {
      key: 'change-status',
      label: row.accountStatus === 'disabled' ? t('启用账号') : t('停用账号'),
      disabled: !canChangeStatus || !row.canManage || row.accountStatus === 'deactivated',
    },
    {
      key: 'reset-password',
      label: t('重置密码'),
      disabled: !canResetPassword || !row.canManage || row.accountStatus === 'deactivated',
    },
    {
      key: 'force-logout',
      label: t('强制下线'),
      disabled:
        !canChangeStatus ||
        !row.canManage ||
        row.activeSessionCount === 0 ||
        row.accountStatus === 'deactivated',
    },
    {
      key: 'deactivate',
      label: t('注销账号'),
      danger: true,
      disabled: !canChangeStatus || !row.canManage || row.accountStatus === 'deactivated',
    },
    {
      key: 'delete-user',
      label: t('删除用户'),
      danger: true,
      disabled: !canDeleteUser || !row.canManage || row.accountStatus !== 'deactivated',
    },
  ]

  const onMoreActionClick = (row: UserRow, key: string) => {
    if (key === 'assign-position') {
      void openTransferPosition([row.uid])
      return
    }
    if (key === 'change-status') {
      const nextStatus: OrganizationUserStatus = row.accountStatus === 'disabled' ? 1 : 0
      void modalApi.confirm({
        title: t(nextStatus === 1 ? '确认启用账号 {name}' : '确认停用账号 {name}', {
          name: row.displayName || row.username,
        }),
        onOk: async () => {
          await updateOrganizationUserStatus(row.uid, nextStatus)
          void messageApi.success(t('账号状态已更新'))
          await usersQuery.refetch()
        },
      })
      return
    }
    if (key === 'reset-password') {
      setSelected(row)
      setPasswordOpen(true)
      return
    }
    if (key === 'force-logout') {
      modalApi.confirm({
        title: t('确认强制下线用户 {name}', { name: row.displayName || row.username }),
        onOk: async () => {
          await forceLogoutUsers([row.uid])
        },
      })
      return
    }
    if (key === 'deactivate') {
      modalApi.confirm({
        title: t('确认注销账号 {name}', { name: row.displayName || row.username }),
        content: t('注销后账号不可登录，且会话将全部失效。'),
        onOk: async () => {
          await deactivateUsers([row.uid])
        },
      })
      return
    }
    if (key === 'delete-user') {
      modalApi.confirm({
        title: t('确认删除用户 {name}', { name: row.displayName || row.username }),
        content: t('仅允许删除已注销满3个月用户。删除为软删除。'),
        okButtonProps: { danger: true },
        onOk: async () => {
          await deleteOrganizationUser(row.uid)
          void messageApi.success(t('删除用户成功'))
          void usersQuery.refetch()
        },
      })
    }
  }

  const loadUserSessionDetails = async () => {
    if (!selected) {
      return
    }
    setSessionLoading(true)
    try {
      if (!currentUser) {
        return
      }
      const items = await getOrganizationUserSessions(selected.uid, 'active', 10)
      setSessionItems(items)
      if (items.length === 0) {
        void messageApi.info(t('当前用户暂无会话数据'))
      }
    } finally {
      setSessionLoading(false)
    }
  }

  const openCreateForm = async () => {
    await departmentsQuery.refetch()
    if (canAssignPosition) {
      await positionsQuery.refetch()
    }
    setEditorMode('create')
    setSelected(null)
    userForm.setFieldsValue({
      username: '',
      password: '',
      displayName: '',
      email: '',
      phone: '',
      status: 1,
      departmentId: undefined,
      positionId: undefined,
    })
    setFormOpen(true)
  }

  const openEditForm = async (row: UserRow) => {
    setEditorMode('edit')
    setSelected(row)
    userForm.setFieldsValue({
      username: row.username,
      password: '',
      displayName: row.displayName,
      email: row.email,
      phone: row.phone,
      status: row.accountStatus === 'active' ? 1 : row.accountStatus === 'disabled' ? 0 : 2,
      departmentId: row.departmentId > 0 ? row.departmentId : undefined,
      positionId: row.positionId > 0 ? row.positionId : undefined,
    })
    setFormOpen(true)
  }

  const submitUserForm = async (values: UserFormValues) => {
    setFormSaving(true)
    try {
      if (editorMode === 'create') {
        if (values.departmentId && departmentStatusMap.get(values.departmentId) === 'disabled') {
          void messageApi.warning(t('不能选择已停用部门'))
          return
        }
        if (values.positionId && positionStatusMap.get(values.positionId) === 'disabled') {
          void messageApi.warning(t('不能选择已停用岗位'))
          return
        }
        await createOrganizationUser({
          username: values.username.trim(),
          password: values.password?.trim() ?? '',
          displayName: values.displayName.trim(),
          email: values.email.trim(),
          phone: values.phone.trim(),
          status: values.status,
          departmentId: values.departmentId || 0,
          positionId: values.positionId || 0,
        })
        void messageApi.success(t('新增用户成功'))
      } else if (selected) {
        await updateOrganizationUserProfile(selected.uid, {
          displayName: values.displayName.trim(),
          avatar: selected.avatar ?? '',
          email: values.email.trim(),
          phone: values.phone.trim(),
        })
        void messageApi.success(t('编辑用户成功'))
      }
      setFormOpen(false)
      void usersQuery.refetch()
    } finally {
      setFormSaving(false)
    }
  }

  const submitResetPassword = async () => {
    if (!selected) return
    setPasswordSaving(true)
    try {
      const tempPassword = await resetOrganizationUserPassword(selected.uid)
      setPasswordOpen(false)
      if (tempPassword) {
        const copyTempPassword = async () => {
          try {
            await copyText(tempPassword)
            void messageApi.success(t('已复制'))
          } catch {
            void messageApi.error(t('复制失败'))
          }
        }
        void messageApi.success({
          content: (
            <span>
              {t('重置密码成功，随机密码：')}
              <Tooltip title={t('点击复制')}>
                <Typography.Link
                  role="button"
                  tabIndex={0}
                  aria-label={`${t('点击复制')}：${tempPassword}`}
                  onClick={() => void copyTempPassword()}
                  onKeyDown={(event) => {
                    if (event.key === 'Enter' || event.key === ' ') {
                      event.preventDefault()
                      void copyTempPassword()
                    }
                  }}
                >
                  {tempPassword} <CopyOutlined />
                </Typography.Link>
              </Tooltip>
              {t('；该用户全部会话已下线')}
            </span>
          ),
          duration: 8,
        })
      } else {
        void messageApi.success(t('重置密码成功'))
      }
    } finally {
      setPasswordSaving(false)
    }
  }

  const openTransferPosition = async (uids = selectedRowKeys) => {
    const manageableUIDs = uids.filter((uid) => rows.find((row) => row.uid === uid)?.canManage)
    if (manageableUIDs.length === 0) {
      void messageApi.warning(t('请先勾选用户'))
      return
    }
    await Promise.all([departmentsQuery.refetch(), positionsQuery.refetch()])
    setTransferTargetUIDs(manageableUIDs)
    transferForm.resetFields()
    setTransferOpen(true)
  }

  const submitTransferPosition = async (values: TransferPositionFormValues) => {
    setTransferSaving(true)
    try {
      if (transferTargetUIDs.length === 1) {
        await assignOrganizationUserPosition(transferTargetUIDs[0], {
          departmentId: values.departmentId,
          positionId: values.positionId,
        })
      } else {
        await batchTransferOrganizationUsers(
          transferTargetUIDs,
          values.departmentId,
          values.positionId,
        )
      }
      setTransferOpen(false)
      setTransferTargetUIDs([])
      setSelectedRowKeys([])
      void messageApi.success(t('转移岗位成功'))
      await Promise.all([usersQuery.refetch(), positionsQuery.refetch()])
    } finally {
      setTransferSaving(false)
    }
  }

  const pickAndImportUsers = async () => {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = '.csv'
    input.onchange = async () => {
      const file = input.files?.[0]
      if (!file) {
        return
      }
      setImporting(true)
      try {
        const text = await file.text()
        const lines = text
          .split(/\r?\n/)
          .map((line) => line.trim())
          .filter((line) => line.length > 0)
        if (lines.length <= 1) {
          void messageApi.warning(t('导入文件无有效数据'))
          return
        }
        const rowsToImport: ImportOrganizationUserItem[] = []
        const statusMap: Record<string, 0 | 1 | 2> = {
          启用: 1,
          停用: 0,
          已注销: 2,
          '1': 1,
          '0': 0,
          '2': 2,
        }
        for (let i = 1; i < lines.length; i += 1) {
          const cols = lines[i].split(',').map((col) => col.trim())
          if (cols.length < 6) continue
          const mappedStatus = statusMap[cols[5]]
          if (mappedStatus === undefined) continue
          if (!USERNAME_PATTERN.test(cols[0])) continue
          rowsToImport.push({
            username: cols[0],
            password: cols[1] || 'Reset@123456',
            displayName: cols[2],
            email: cols[3],
            phone: cols[4],
            status: mappedStatus,
          })
        }
        if (rowsToImport.length === 0) {
          void messageApi.warning(t('未解析到可导入记录，请检查模板和枚举值'))
          return
        }
        await importOrganizationUsers(rowsToImport)
        void messageApi.success(t('批量导入成功，共 {count} 条', { count: rowsToImport.length }))
        void usersQuery.refetch()
      } finally {
        setImporting(false)
      }
    }
    input.click()
  }

  const handleExportUsers = async () => {
    setExporting(true)
    try {
      const blob = await exportOrganizationUsers(filters)
      const url = window.URL.createObjectURL(blob)
      const anchor = document.createElement('a')
      anchor.href = url
      anchor.download = `organization-users-${Date.now()}.csv`
      document.body.appendChild(anchor)
      anchor.click()
      document.body.removeChild(anchor)
      window.URL.revokeObjectURL(url)
      void messageApi.success(t('导出用户成功'))
    } finally {
      setExporting(false)
    }
  }

  const downloadImportTemplate = () => {
    const anchor = document.createElement('a')
    anchor.href = '/templates/organization_users_import_template.csv'
    anchor.download = 'organization_users_import_template.csv'
    document.body.appendChild(anchor)
    anchor.click()
    document.body.removeChild(anchor)
  }

  const importHoverMenu: MenuProps = {
    items: [
      {
        key: 'download-template',
        label: t('下载导入模板'),
      },
    ],
    onClick: ({ key }) => {
      if (key === 'download-template') {
        downloadImportTemplate()
      }
    },
  }

  const applyFilters = (values: FilterFormValues) => {
    setFilters({
      keyword: values.keyword?.trim() || undefined,
      phone: values.phone?.trim() || undefined,
      status: values.status,
      departmentId: values.departmentId,
      positionId: values.positionId,
      createdFrom: toTimezoneDateTimeString(values.createdAt?.[0], systemTimezone),
      createdTo: toTimezoneDateTimeString(values.createdAt?.[1], systemTimezone),
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
          {canEditProfile ? (
            <UiButton type="primary" onClick={openCreateForm}>
              {t('新增用户')}
            </UiButton>
          ) : null}
          {canEditProfile ? (
            <Dropdown menu={importHoverMenu} trigger={['hover']}>
              <UiButton loading={importing} onClick={pickAndImportUsers}>
                {t('批量导入')}
              </UiButton>
            </Dropdown>
          ) : null}
          <UiButton loading={exporting} onClick={handleExportUsers}>
            {t('导出用户')}
          </UiButton>
          {canAssignPosition ? (
            <UiButton onClick={() => void openTransferPosition()}>{t('转移岗位')}</UiButton>
          ) : null}
          {canChangeStatus ? (
            <UiButton
              onClick={() => {
                if (selectedRowKeys.length === 0) {
                  void messageApi.warning(t('请先勾选用户'))
                  return
                }
                modalApi.confirm({
                  title: t('确认强制下线选中 {count} 个用户？', { count: selectedRowKeys.length }),
                  onOk: async () => {
                    await forceLogoutUsers(selectedRowKeys)
                  },
                })
              }}
            >
              {t('批量强制下线')}
            </UiButton>
          ) : null}
          {canChangeStatus ? (
            <UiButton
              danger
              onClick={() => {
                if (selectedRowKeys.length === 0) {
                  void messageApi.warning(t('请先勾选用户'))
                  return
                }
                const validUIDs = selectedRowKeys.filter((uid) => {
                  const row = rows.find((item) => item.uid === uid)
                  return row && row.accountStatus !== 'deactivated'
                })
                const skipped = selectedRowKeys.length - validUIDs.length
                if (validUIDs.length === 0) {
                  void messageApi.warning(t('选中用户均已注销，无需重复操作'))
                  return
                }
                modalApi.confirm({
                  title: t('确认注销选中 {count} 个用户？', { count: validUIDs.length }),
                  content: t('注销后账号不可登录，且会话将全部失效。'),
                  onOk: async () => {
                    await deactivateUsers(validUIDs)
                    if (skipped > 0) {
                      await messageApi.info(
                        t('已自动跳过 {count} 个已注销用户', { count: skipped }),
                      )
                    }
                  },
                })
              }}
            >
              {t('批量注销')}
            </UiButton>
          ) : null}
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
              <Input
                placeholder={t('用户名/姓名')}
                allowClear
                onPressEnter={() => filterForm.submit()}
              />
            </Form.Item>
            <Form.Item label={t('手机号')} name="phone">
              <Input
                placeholder={t('请输入手机号')}
                allowClear
                onPressEnter={() => filterForm.submit()}
              />
            </Form.Item>
            <Form.Item label={t('状态')} name="status">
              <Select
                allowClear
                style={{ width: 150 }}
                options={[
                  { value: 'active', label: t('启用') },
                  { value: 'disabled', label: t('停用') },
                  { value: 'deactivated', label: t('已注销') },
                ]}
              />
            </Form.Item>
            <Form.Item label={t('部门')} name="departmentId">
              <Select
                allowClear
                style={{ width: 180 }}
                options={departmentFilterOptions}
                loading={departmentsQuery.isLoading}
              />
            </Form.Item>
            <Form.Item label={t('岗位')} name="positionId">
              <Select
                allowClear
                style={{ width: 240 }}
                options={positionFilterOptions}
                loading={positionsQuery.isLoading}
              />
            </Form.Item>
            <Form.Item label={t('创建时间')} name="createdAt">
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
                  }}
                >
                  {t('重置')}
                </UiButton>
              </Space>
            </Form.Item>
          </Form>
        </Card>

        <Card className="organization-users-table-card system-table-card table-scroll-region">
          <Table<UserRow>
            className="organization-users-table"
            rowKey="uid"
            size="small"
            loading={usersQuery.isLoading || usersQuery.isFetching}
            scroll={{ x: 'max-content', y: 392 }}
            rowSelection={{
              selectedRowKeys,
              onChange: (keys) => setSelectedRowKeys(keys as number[]),
              getCheckboxProps: (row) => ({ disabled: !row.canManage }),
            }}
            dataSource={rows}
            pagination={{
              current: pageNo,
              pageSize,
              total: usersQuery.data?.total ?? 0,
              showSizeChanger: true,
              showTotal: (total) => t('共 {total} 条', { total }),
            }}
            onChange={(
              pagination: TablePaginationConfig,
              _filters,
              sorter: SorterResult<UserRow> | SorterResult<UserRow>[],
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
              const fieldMap: Record<
                string,
                | 'uid'
                | 'username'
                | 'display_name'
                | 'status'
                | 'active_session_count'
                | 'last_login_at'
              > = {
                uid: 'uid',
                username: 'username',
                displayName: 'display_name',
                accountStatus: 'status',
                activeSessionCount: 'active_session_count',
                lastLoginAt: 'last_login_at',
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
              { title: t('用户ID'), dataIndex: 'uid', sorter: true },
              {
                title: t('用户名'),
                dataIndex: 'username',
                sorter: true,
                render: (value: string, row) => (
                  <Space size={6}>
                    <Tooltip title={t('点击复制文本')}>
                      <Typography.Link
                        onClick={() =>
                          void copyText(value).then(() => messageApi.success(t('已复制')))
                        }
                      >
                        {truncateUsername(value)}
                      </Typography.Link>
                    </Tooltip>
                    {row.isProtected ? (
                      <UiProtectedBadge
                        ariaLabel={t('受保护')}
                        tooltip={t('用户从岗位继承了受保护角色，仅超级管理员可执行敏感操作。')}
                      />
                    ) : null}
                  </Space>
                ),
              },
              { title: t('姓名'), dataIndex: 'displayName', sorter: true },
              {
                title: t('手机号'),
                dataIndex: 'phone',
                width: 150,
                render: (value: string) => (
                  value ? (
                    <Tooltip title={t('点击复制文本')}>
                      <Typography.Link
                        onClick={() =>
                          void copyText(value).then(() => messageApi.success(t('已复制')))
                        }
                      >
                        {maskPhone(value)}
                      </Typography.Link>
                    </Tooltip>
                  ) : (
                    '-'
                  )
                ),
              },
              {
                title: t('邮箱'),
                dataIndex: 'email',
                width: 180,
                ellipsis: true,
                render: (value: string) => (
                  value ? (
                    <Tooltip title={t('点击复制文本')}>
                      <Typography.Link
                        ellipsis
                        style={{ display: 'inline-block', maxWidth: 138 }}
                        onClick={() =>
                          void copyText(value).then(() => messageApi.success(t('已复制')))
                        }
                      >
                        {maskEmail(value)}
                      </Typography.Link>
                    </Tooltip>
                  ) : (
                    '-'
                  )
                ),
              },
              {
                title: t('部门'),
                render: (_, row) => (
                  <Space size={6}>
                    <span>{row.departmentName || '-'}</span>
                    {row.departmentId > 0 &&
                      departmentStatusMap.get(row.departmentId) === 'disabled' && (
                        <Tag color="red">{t('停用')}</Tag>
                      )}
                  </Space>
                ),
              },
              {
                title: t('岗位'),
                render: (_, row) => (
                  <Space size={6}>
                    <span>{row.positionName || '-'}</span>
                    {row.positionId > 0 && positionStatusMap.get(row.positionId) === 'disabled' && (
                      <Tag color="red">{t('停用')}</Tag>
                    )}
                  </Space>
                ),
              },
              {
                title: t('关联角色'),
                render: (_, row) => <RoleSummary roleNames={row.roleNames} />,
              },
              {
                title: t('账号状态'),
                dataIndex: 'accountStatus',
                render: (status: UserRow['accountStatus']) => renderAccountStatus(status, t),
                sorter: true,
              },
              {
                title: t('在线状态'),
                dataIndex: 'onlineStatus',
                render: (status: UserRow['onlineStatus']) => renderOnlineStatus(status, t),
              },
              { title: t('活跃会话数'), dataIndex: 'activeSessionCount', sorter: true },
              {
                title: t('最近登录时间'),
                dataIndex: 'lastLoginAt',
                sorter: true,
                render: (value: string) => formatDateTime(value, systemTimezone),
              },
              {
                title: t('操作'),
                className: 'organization-users-actions-col',
                fixed: 'right',
                width: 220,
                render: (_, row) => (
                  <Space size={0}>
                    <UiButton
                      type="link"
                      onClick={() => {
                        setSelected(row)
                        setSessionItems([])
                        setDetailOpen(true)
                      }}
                    >
                      {t('查看详情')}
                    </UiButton>
                    <Tooltip
                      title={
                        !row.canManage
                          ? t('不能操作同级、更高权限或受保护账号')
                          : isLockedUser(row)
                            ? t('已停用/已注销用户不可编辑')
                            : ''
                      }
                    >
                      <UiButton
                        type="link"
                        disabled={!canEditProfile || !row.canManage || isLockedUser(row)}
                        onClick={() => {
                          if (!canEditProfile || !row.canManage || isLockedUser(row)) {
                            return
                          }
                          void openEditForm(row)
                        }}
                      >
                        {t('编辑')}
                      </UiButton>
                    </Tooltip>
                    <Dropdown
                      menu={{
                        items: buildMoreActions(row),
                        onClick: ({ key }) => onMoreActionClick(row, key),
                      }}
                      trigger={['hover']}
                    >
                      <UiButton type="link">{t('更多')}</UiButton>
                    </Dropdown>
                  </Space>
                ),
              },
            ]}
          />
        </Card>
      </Space>

      <Drawer
        title={t('用户详情')}
        open={detailOpen}
        onClose={() => setDetailOpen(false)}
        width={520}
      >
        <Typography.Paragraph>
          {t('用户ID：')}
          {selected?.uid ?? '-'}
        </Typography.Paragraph>
        <Typography.Paragraph>
          {t('用户名：')}
          {selected?.username ?? '-'}
        </Typography.Paragraph>
        <Typography.Paragraph>
          {t('姓名：')}
          {selected?.displayName ?? '-'}
        </Typography.Paragraph>
        <Typography.Paragraph>
          {t('手机号：')}
          {selected?.phone ?? '-'}
        </Typography.Paragraph>
        <Typography.Paragraph>
          {t('邮箱：')}
          {selected?.email ?? '-'}
        </Typography.Paragraph>
        <Typography.Paragraph>
          {t('部门：')}
          {selected?.departmentName ?? '-'}
        </Typography.Paragraph>
        <Typography.Paragraph>
          {t('岗位：')}
          {selected?.positionName ?? '-'}
        </Typography.Paragraph>
        <Typography.Paragraph>
          {t('账号状态：')}
          {selected ? renderAccountStatus(selected.accountStatus, t) : '-'}
        </Typography.Paragraph>
        <Typography.Paragraph>
          {t('在线状态：')}
          {selected ? renderOnlineStatus(selected.onlineStatus, t) : '-'}
        </Typography.Paragraph>
        <Typography.Paragraph>
          {t('活跃会话数：')}
          {selected?.activeSessionCount ?? '-'}
        </Typography.Paragraph>
        <Typography.Paragraph>
          {t('最近登录IP：')}
          {selected?.lastLoginIp ?? '-'}
        </Typography.Paragraph>
        <Typography.Paragraph>
          {t('最近登录时间：')}
          {formatDateTime(selected?.lastLoginAt, systemTimezone)}
        </Typography.Paragraph>

        <Space direction="vertical" size={12} style={{ width: '100%', marginTop: 8 }}>
          <Typography.Title level={5} style={{ margin: 0 }}>
            {t('会话信息')}
          </Typography.Title>
          <Typography.Text type="secondary">{t('获取最新十条会话数据。')}</Typography.Text>
          <UiButton onClick={() => void loadUserSessionDetails()}>{t('查看会话明细')}</UiButton>
          {sessionLoading ? (
            <Spin size="small" />
          ) : (
            sessionItems.map((item) => (
              <Card size="small" key={item.sessionId}>
                <Typography.Text>
                  {item.status} | {item.loginIp} | {formatDateTime(item.lastSeenAt, systemTimezone)}
                </Typography.Text>
              </Card>
            ))
          )}
        </Space>
      </Drawer>

      <Drawer
        title={t(editorMode === 'edit' ? '编辑用户' : '新增用户')}
        open={formOpen}
        onClose={() => setFormOpen(false)}
        width={500}
      >
        <Form form={userForm} layout="vertical" onFinish={(values) => void submitUserForm(values)}>
          <Form.Item
            label={t('用户名')}
            name="username"
            rules={[
              { required: true, message: t('请输入用户名') },
              {
                pattern: USERNAME_PATTERN,
                message: t('用户名必须以字母开头，且只能包含字母、数字、下划线和减号'),
              },
            ]}
          >
            <Input disabled={editorMode === 'edit'} maxLength={64} />
          </Form.Item>
          {editorMode === 'create' && (
            <Form.Item
              label={t('初始密码')}
              name="password"
              rules={[{ required: true, message: t('请输入初始密码') }]}
            >
              <Input.Password />
            </Form.Item>
          )}
          <Form.Item
            label={t('姓名')}
            name="displayName"
            rules={[{ required: true, message: t('请输入姓名') }]}
          >
            <Input />
          </Form.Item>
          <Form.Item label={t('手机号')} name="phone">
            <Input />
          </Form.Item>
          <Form.Item label={t('邮箱')} name="email">
            <Input />
          </Form.Item>
          {editorMode === 'create' && canAssignPosition ? (
            <>
              <Form.Item
                label={t('所属部门')}
                name="departmentId"
                rules={[{ required: true, message: t('请选择所属部门') }]}
              >
                <Select
                  options={departmentFormOptions}
                  loading={departmentsQuery.isLoading}
                  placeholder={t('请选择')}
                  allowClear
                  showSearch
                  optionFilterProp="searchText"
                  onChange={() => {
                    userForm.setFieldValue('positionId', undefined)
                  }}
                />
              </Form.Item>
              {formDepartmentId > 0 && departmentFullPathMap.get(formDepartmentId) ? (
                <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 16 }}>
                  {t('部门路径')}：{departmentFullPathMap.get(formDepartmentId)}
                </Typography.Text>
              ) : null}
              <Form.Item
                label={
                  <Space size={6}>
                    <span>{t('所属岗位')}</span>
                    <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                      {t('无权分配：岗位包含当前账号不可继续授予的权限')}
                    </Typography.Text>
                  </Space>
                }
                name="positionId"
                rules={[{ required: true, message: t('请选择所属岗位') }]}
              >
                <Select
                  options={scopedPositionFormOptions}
                  loading={formPositionsQuery.isLoading || formPositionsQuery.isFetching}
                  placeholder={formDepartmentId > 0 ? t('请选择') : t('请先选择部门')}
                  notFoundContent={
                    formPositionsQuery.isFetching ? t('加载中...') : t('该部门暂无可分配岗位')
                  }
                  disabled={formDepartmentId <= 0}
                  allowClear
                />
              </Form.Item>
            </>
          ) : null}
          {editorMode === 'create' ? (
            <Form.Item
              label={t('账号状态')}
              name="status"
              rules={[{ required: true, message: t('请选择账号状态') }]}
            >
              <Select
                options={[
                  { value: 1, label: t('启用') },
                  { value: 0, label: t('停用') },
                  { value: 2, label: t('已注销') },
                ]}
              />
            </Form.Item>
          ) : null}
          <UiButton type="primary" loading={formSaving} onClick={() => userForm.submit()}>
            {t('保存')}
          </UiButton>
        </Form>
      </Drawer>

      <Modal
        title={t('重置密码')}
        open={passwordOpen}
        onCancel={() => {
          setPasswordOpen(false)
        }}
        onOk={() => void submitResetPassword()}
        confirmLoading={passwordSaving}
      >
        <Typography.Text>
          {t('将为用户 {name} 生成随机密码，并下线该用户全部会话。', {
            name: selected?.displayName || selected?.username || '-',
          })}
        </Typography.Text>
      </Modal>
      <Modal
        title={t('转移岗位')}
        open={transferOpen}
        onCancel={() => setTransferOpen(false)}
        onOk={() => transferForm.submit()}
        confirmLoading={transferSaving}
      >
        <Form
          form={transferForm}
          layout="vertical"
          onFinish={(values) => void submitTransferPosition(values)}
        >
          <Typography.Paragraph type="secondary">
            {t('将选中的 {count} 个用户转移到指定部门和岗位。', {
              count: transferTargetUIDs.length,
            })}
          </Typography.Paragraph>
          <Form.Item
            label={t('所属部门')}
            name="departmentId"
            rules={[{ required: true, message: t('请选择所属部门') }]}
          >
            <Select
              options={departmentFormOptions}
              loading={departmentsQuery.isLoading}
              placeholder={t('请选择')}
              onChange={() => transferForm.setFieldValue('positionId', undefined)}
            />
          </Form.Item>
          <Form.Item
            label={t('所属岗位')}
            name="positionId"
            rules={[{ required: true, message: t('请选择所属岗位') }]}
          >
            <Select
              options={scopedTransferPositionOptions}
              loading={transferPositionsQuery.isLoading || transferPositionsQuery.isFetching}
              placeholder={transferDepartmentId > 0 ? t('请选择') : t('请先选择部门')}
              notFoundContent={
                transferPositionsQuery.isFetching ? t('加载中...') : t('该部门暂无可分配岗位')
              }
              disabled={transferDepartmentId <= 0}
            />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
