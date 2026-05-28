import { useMutation, useQuery } from '@tanstack/react-query'
import { Card, Form, Input, Modal, Select, Space, Switch, Table, Tabs, Tag, Tooltip, Typography, message } from 'antd'
import type { TablePaginationConfig } from 'antd'
import type { SorterResult } from 'antd/es/table/interface'
import { useState } from 'react'
import { UiButton } from '../../components/ui'
import { useI18n } from '../../i18n/messages'
import {
  deleteSystemAlertBot,
  deleteSystemAlertScene,
  deleteSystemAlertTemplate,
  getSystemAlertBots,
  getSystemAlertScenes,
  getSystemAlertTemplates,
  saveSystemAlertBot,
  saveSystemAlertScene,
  saveSystemAlertTemplate,
  testSendAlertScene,
  type SystemAlertBotItem,
  type SystemAlertSceneItem,
  type SystemAlertTemplateItem,
} from '../../services/api/system'
import { useUiSettingsStore } from '../../store/uiSettings'
import { formatDateTime } from '../../utils/timezone'

// ===== Bot Tab =====

type BotEditorFormValues = {
  name: string
  username: string
  token: string
  botType: string
  enabled: boolean
}

const BOT_TYPE_OPTIONS = [
  { value: 'telegram', label: 'Telegram' },
  { value: 'feishu', label: '飞书' },
]

function BotTab() {
  const { t } = useI18n()
  const systemTimezone = useUiSettingsStore((state) => state.systemTimezone)
  const [messageApi, contextHolder] = message.useMessage()
  const [modalApi, modalHolder] = Modal.useModal()
  const [filterForm] = Form.useForm()
  const [editorForm] = Form.useForm<BotEditorFormValues>()
  const watchedBotType = Form.useWatch('botType', editorForm)
  const [pageNo, setPageNo] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [orderField, setOrderField] = useState<string | undefined>('created_at')
  const [orderType, setOrderType] = useState<'asc' | 'desc' | undefined>('desc')
  const [keyword, setKeyword] = useState('')
  const [botTypeFilter, setBotTypeFilter] = useState('')
  const [queryTrigger, setQueryTrigger] = useState(0)
  const [editorOpen, setEditorOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<SystemAlertBotItem | null>(null)

  const listQuery = useQuery({
    queryKey: ['system-alert-bots', pageNo, pageSize, orderField, orderType, keyword, botTypeFilter, queryTrigger],
    queryFn: () => getSystemAlertBots(pageNo, pageSize, orderField, orderType, keyword, botTypeFilter),
  })

  const saveMutation = useMutation({
    mutationFn: async (values: BotEditorFormValues) => {
      await saveSystemAlertBot({
        id: editingItem?.id,
        name: values.name,
        username: values.username,
        token: values.token,
        botType: values.botType,
        enabled: values.enabled,
      })
    },
    onSuccess: async () => {
      void messageApi.success(editingItem ? t('保存成功') : t('创建成功'))
      setEditorOpen(false)
      await listQuery.refetch()
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => deleteSystemAlertBot(id),
    onSuccess: async () => {
      void messageApi.success(t('删除成功'))
      await listQuery.refetch()
    },
  })

  const handleTableChange = (pagination: TablePaginationConfig, _: any, sorter: SorterResult<SystemAlertBotItem> | SorterResult<SystemAlertBotItem>[]) => {
    if (pagination.current) setPageNo(pagination.current)
    if (pagination.pageSize && pagination.pageSize !== pageSize) { setPageSize(pagination.pageSize); setPageNo(1) }
    const single = Array.isArray(sorter) ? sorter[0] : sorter
    if (!single?.field || !single.order) { setOrderField(undefined); setOrderType(undefined); return }
    const fieldMap: Record<string, string> = { id: 'id', name: 'name', botType: 'bot_type', enabled: 'enabled', createdAt: 'created_at', updatedAt: 'updated_at' }
    setOrderField(fieldMap[single.field as string] || (single.field as string))
    setOrderType(single.order === 'ascend' ? 'asc' : 'desc')
  }

  const openEditor = (item?: SystemAlertBotItem) => {
    setEditingItem(item || null)
    editorForm.resetFields()
    editorForm.setFieldsValue({
      name: item?.name || '',
      username: item?.username || '',
      token: item?.token || '',
      botType: item?.botType || 'telegram',
      enabled: item?.enabled ?? true,
    })
    setEditorOpen(true)
  }

  return (
    <>
      {contextHolder}
      {modalHolder}
      <Space direction="vertical" size={16} className="full-width">
        <Typography.Text type="secondary" style={{ fontSize: 13 }}>
          {t('说明：新增您熟悉的平台机器人，配置好Token信息，以便下一步调用。')}
        </Typography.Text>
        <Space wrap>
          <UiButton type="primary" onClick={() => openEditor()}>{t('新增')}</UiButton>
        </Space>
        <Card>
          <Form form={filterForm} layout="inline" style={{ rowGap: 12 }} onFinish={(v: any) => { setKeyword(v.keyword || ''); setBotTypeFilter(v.botType || ''); setPageNo(1); setQueryTrigger((n) => n + 1) }}>
            <Form.Item label={t('关键字')} name="keyword">
              <Input placeholder={t('名称 / 用户名')} allowClear onPressEnter={() => filterForm.submit()} />
            </Form.Item>
            <Form.Item label={t('类型')} name="botType">
              <Select style={{ width: 140 }} allowClear options={BOT_TYPE_OPTIONS.map((o) => ({ value: o.value, label: t(o.label) }))} />
            </Form.Item>
            <Form.Item>
              <Space size={8}>
                <UiButton type="primary" onClick={() => filterForm.submit()}>{t('查询')}</UiButton>
                <UiButton onClick={() => { filterForm.resetFields(); setKeyword(''); setBotTypeFilter(''); setPageNo(1); setQueryTrigger((n) => n + 1) }}>{t('重置')}</UiButton>
              </Space>
            </Form.Item>
          </Form>
        </Card>
        <Card className="compact-table-card system-table-card table-scroll-region">
          <Table<SystemAlertBotItem>
            rowKey="id"
            loading={listQuery.isLoading || listQuery.isFetching}
            dataSource={listQuery.data?.items ?? []}
            scroll={{ x: 1100, y: 340 }}
            pagination={{ current: pageNo, pageSize, total: listQuery.data?.total ?? 0, showSizeChanger: true, showTotal: (total) => t('共 {total} 条', { total }) }}
            onChange={handleTableChange}
            columns={[
              { title: 'ID', dataIndex: 'id', sorter: true, width: 70 },
              { title: t('名称'), dataIndex: 'name', sorter: true, width: 150 },
              { title: t('用户名'), dataIndex: 'username', width: 130 },
              { title: t('类型'), dataIndex: 'botType', width: 100, render: (v: string) => <Tag color={v === 'telegram' ? 'blue' : v === 'feishu' ? 'green' : 'default'}>{v}</Tag> },
              { title: t('状态'), dataIndex: 'enabled', width: 80, render: (v: boolean) => <Tag color={v ? 'green' : 'default'}>{v ? t('启用') : t('停用')}</Tag> },
              { title: t('已关联场景'), dataIndex: 'linkedSceneKeys', width: 200, render: (v: string[]) => {
                if (!v?.length) return '-'
                const show = v.slice(0, 2)
                const rest = v.length - 2
                return (<Space size={[4, 4]} wrap>{show.map((k) => <Tag key={k}>{k}</Tag>)}{rest > 0 && <Tooltip title={v.join(', ')}><Tag style={{ cursor: 'pointer' }}>+{rest}</Tag></Tooltip>}</Space>)
              } },
              { title: t('创建时间'), dataIndex: 'createdAt', sorter: true, width: 170, defaultSortOrder: 'descend', render: (v: string) => formatDateTime(v, systemTimezone) },
              { title: t('更新时间'), dataIndex: 'updatedAt', sorter: true, width: 170, render: (v: string) => formatDateTime(v, systemTimezone) },
              { title: t('操作'), fixed: 'right', width: 120, render: (_, row) => (
                <Space size={0}>
                  <UiButton type="link" onClick={() => openEditor(row)}>{t('编辑')}</UiButton>
                  <UiButton type="link" danger loading={deleteMutation.isPending} onClick={() => { void modalApi.confirm({ title: t('确认删除 {name}？', { name: row.name }), onOk: async () => { await deleteMutation.mutateAsync(row.id) } }) }}>{t('删除')}</UiButton>
                </Space>
              ) },
            ]}
          />
        </Card>
      </Space>

      <Modal title={editingItem ? t('编辑机器人配置') : t('新增机器人配置')} open={editorOpen} onCancel={() => setEditorOpen(false)} onOk={() => editorForm.submit()} confirmLoading={saveMutation.isPending} destroyOnHidden={false}>
        <Form form={editorForm} layout="vertical" onFinish={(v) => saveMutation.mutate(v)} autoComplete="off">
          <Form.Item name="name" label={t('名称')} tooltip={t('机器人的显示名称')} rules={[{ required: true, message: t('请输入名称') }, { max: 20 }]}>
            <Input maxLength={20} showCount />
          </Form.Item>
          <Form.Item name="username" label={t('用户名')} tooltip={t('TG Bot的username，用于标识')}>
            <Input placeholder="Bot username" maxLength={30} showCount />
          </Form.Item>
          <Form.Item name="token" label="Token" tooltip={watchedBotType === 'feishu' ? t('创建飞书群机器人后复制其中的Token部分') : t('从BotFather获取的API Token')} rules={[{ required: true, message: t('请输入Token') }, { max: 100 }, watchedBotType === 'feishu' ? { pattern: /^[a-z0-9-]{36,}$/, message: t('飞书Token格式不正确') } : { pattern: /^[0-9]+:[A-Za-z0-9_-]{35,}$/, message: t('TG Token格式不正确') }]}>
            <Input maxLength={100} showCount autoComplete="off" />
          </Form.Item>
          <Form.Item name="botType" label={t('类型')} tooltip={t('机器人平台类型')} rules={[{ required: true }]} initialValue="telegram">
            <Select options={BOT_TYPE_OPTIONS.map((o) => ({ value: o.value, label: t(o.label) }))} onChange={() => { setTimeout(() => editorForm.validateFields(['token']), 0) }} />
          </Form.Item>
          <Form.Item name="enabled" label={t('启用')} tooltip={t('停用后关联的场景无法发送')} valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}

// ===== Scene Tab =====

type SceneEditorFormValues = {
  sceneKey: string
  botId: number
  parseMode: string
  groupName: string
  groupId: string
  notifyTemplate: string
}

function SceneTab() {
  const { t } = useI18n()
  const systemTimezone = useUiSettingsStore((state) => state.systemTimezone)
  const [messageApi, contextHolder] = message.useMessage()
  const [modalApi, modalHolder] = Modal.useModal()
  const [editorForm] = Form.useForm<SceneEditorFormValues>()
  const [pageNo, setPageNo] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [orderField, setOrderField] = useState<string | undefined>('created_at')
  const [orderType, setOrderType] = useState<'asc' | 'desc' | undefined>('desc')
  const [keyword, setKeyword] = useState('')
  const [queryTrigger, setQueryTrigger] = useState(0)
  const [editorOpen, setEditorOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<SystemAlertSceneItem | null>(null)

  const botsQuery = useQuery({
    queryKey: ['system-alert-bots-enabled'],
    queryFn: () => getSystemAlertBots(1, 100),
  })

  const watchedBotId = Form.useWatch('botId', editorForm)
  const isBotDisabled = !!(watchedBotId && !(botsQuery.data?.items || []).find((b) => b.id === watchedBotId)?.enabled)
  const selectedBotType = (watchedBotId as number) ? (botsQuery.data?.items || []).find((b) => b.id === watchedBotId)?.botType : undefined

  const templatesQuery = useQuery({
    queryKey: ['system-alert-templates-for-scene', selectedBotType],
    queryFn: () => getSystemAlertTemplates(1, 100, undefined, undefined, undefined, selectedBotType),
    enabled: false,
  })
  const templateOptions = templatesQuery.data?.items || []

  const listQuery = useQuery({
    queryKey: ['system-alert-scenes', pageNo, pageSize, orderField, orderType, queryTrigger],
    queryFn: () => getSystemAlertScenes(pageNo, pageSize, orderField, orderType, keyword),
  })

  const saveMutation = useMutation({
    mutationFn: async (values: SceneEditorFormValues) => {
      await saveSystemAlertScene({ id: editingItem?.id, sceneKey: values.sceneKey, botId: values.botId || 0, parseMode: values.parseMode || '', groupName: values.groupName, groupId: values.groupId, notifyTemplate: values.notifyTemplate })
    },
    onSuccess: async () => {
      void messageApi.success(editingItem ? t('保存成功') : t('创建成功'))
      setEditorOpen(false)
      await listQuery.refetch()
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => deleteSystemAlertScene(id),
    onSuccess: async () => {
      void messageApi.success(t('删除成功'))
      await listQuery.refetch()
    },
  })

  const [testSendOpen, setTestSendOpen] = useState(false)
  const [testSendItem, setTestSendItem] = useState<SystemAlertSceneItem | null>(null)
  const [testSendForm] = Form.useForm()
  const testSendVars = testSendItem?.notifyTemplate?.match(/\{(\w+)\}/g)?.map((m) => m.slice(1, -1)) || []

  const testSendMutation = useMutation({
    mutationFn: async (variables: Record<string, string>) => {
      await testSendAlertScene(testSendItem!.id, variables)
    },
    onSuccess: () => {
      void messageApi.success(t('发送成功'))
      setTestSendOpen(false)
    },
  })

  const openTestSend = (item: SystemAlertSceneItem) => {
    setTestSendItem(item)
    testSendForm.resetFields()
    setTestSendOpen(true)
  }

  const handleTableChange = (pagination: TablePaginationConfig, _: any, sorter: SorterResult<SystemAlertSceneItem> | SorterResult<SystemAlertSceneItem>[]) => {
    if (pagination.current) setPageNo(pagination.current)
    if (pagination.pageSize && pagination.pageSize !== pageSize) { setPageSize(pagination.pageSize); setPageNo(1) }
    const single = Array.isArray(sorter) ? sorter[0] : sorter
    if (!single?.field || !single.order) { setOrderField(undefined); setOrderType(undefined); return }
    const fieldMap: Record<string, string> = { id: 'id', sceneKey: 'scene_key', groupName: 'group_name', createdAt: 'created_at', updatedAt: 'updated_at' }
    setOrderField(fieldMap[single.field as string] || (single.field as string))
    setOrderType(single.order === 'ascend' ? 'asc' : 'desc')
  }

  const openEditor = (item?: SystemAlertSceneItem) => {
    setEditingItem(item || null)
    editorForm.resetFields()
    editorForm.setFieldsValue({
      sceneKey: item?.sceneKey || '',
      botId: item?.botId || undefined as any,
      parseMode: item?.parseMode || '',
      groupName: item?.groupName || '',
      groupId: item?.groupId || '',
      notifyTemplate: item?.notifyTemplate || '',
    })
    void botsQuery.refetch()
    setEditorOpen(true)
  }

  return (
    <>
      {contextHolder}
      {modalHolder}
      <Space direction="vertical" size={16} className="full-width">
        <Typography.Text type="secondary" style={{ fontSize: 13 }}>
          {t('说明：提前在页面配置好场景key信息，然后在代码中通过场景key获取对应机器人对象进行通知或告警。')}
        </Typography.Text>
        <Space wrap>
          <UiButton type="primary" onClick={() => openEditor()}>{t('新增')}</UiButton>
        </Space>
        <Card>
          <Space wrap style={{ rowGap: 12 }}>
            <Input placeholder={t('搜索场景Key/群名称')} allowClear value={keyword} onChange={(e) => setKeyword(e.target.value)} onPressEnter={() => { setPageNo(1); setQueryTrigger((n) => n + 1) }} style={{ width: 240 }} />
            <UiButton type="primary" onClick={() => { setPageNo(1); setQueryTrigger((n) => n + 1) }}>{t('查询')}</UiButton>
            <UiButton onClick={() => { setKeyword(''); setPageNo(1); setQueryTrigger((n) => n + 1) }}>{t('重置')}</UiButton>
          </Space>
        </Card>
        <Card className="compact-table-card system-table-card table-scroll-region">
          <Table<SystemAlertSceneItem>
            rowKey="id"
            loading={listQuery.isLoading || listQuery.isFetching}
            dataSource={listQuery.data?.items ?? []}
            scroll={{ x: 900, y: 340 }}
            pagination={{ current: pageNo, pageSize, total: listQuery.data?.total ?? 0, showSizeChanger: true, showTotal: (total) => t('共 {total} 条', { total }) }}
            onChange={handleTableChange}
            columns={[
              { title: 'ID', dataIndex: 'id', sorter: true, width: 70 },
              { title: t('场景Key'), dataIndex: 'sceneKey', sorter: true, width: 140 },
              { title: t('关联机器人'), dataIndex: 'botId', width: 130, render: (v: number) => (botsQuery.data?.items || []).find((b) => b.id === v)?.name || '-' },
              { title: t('群名称'), dataIndex: 'groupName', width: 130 },
              { title: t('群ID'), dataIndex: 'groupId', width: 130 },
              { title: t('通知模板'), dataIndex: 'notifyTemplate', width: 220, ellipsis: true },
              { title: t('创建时间'), dataIndex: 'createdAt', sorter: true, width: 170, defaultSortOrder: 'descend', render: (v: string) => formatDateTime(v, systemTimezone) },
              { title: t('更新时间'), dataIndex: 'updatedAt', sorter: true, width: 170, render: (v: string) => formatDateTime(v, systemTimezone) },
              { title: t('操作'), fixed: 'right', width: 180, render: (_, row) => (
                <Space size={0}>
                  <UiButton type="link" onClick={() => openEditor(row)}>{t('编辑')}</UiButton>
                  <UiButton type="link" onClick={() => openTestSend(row)}>{t('测试发送')}</UiButton>
                  <UiButton type="link" danger loading={deleteMutation.isPending} onClick={() => { void modalApi.confirm({ title: t('确认删除 {name}？', { name: row.sceneKey }), onOk: async () => { await deleteMutation.mutateAsync(row.id) } }) }}>{t('删除')}</UiButton>
                </Space>
              ) },
            ]}
          />
        </Card>
      </Space>

      <Modal title={editingItem ? t('编辑场景配置') : t('新增场景配置')} open={editorOpen} onCancel={() => setEditorOpen(false)} onOk={() => editorForm.submit()} confirmLoading={saveMutation.isPending} okButtonProps={{ disabled: isBotDisabled }} destroyOnHidden={false}>
        <Form form={editorForm} layout="vertical" onFinish={(v) => saveMutation.mutate(v)} autoComplete="off">
          <Form.Item name="sceneKey" label={t('场景Key')} tooltip={t('唯一标识，代码通过Key查找对应的机器人')} rules={[{ required: true, message: t('请输入场景Key') }, { max: 20 }, { pattern: /^[a-zA-Z0-9_]+$/, message: t('仅支持英文数字下划线') }]}>
            <Input maxLength={20} showCount />
          </Form.Item>
          <Form.Item name="botId" label={t('关联机器人')} tooltip={t('选择用于发送消息的机器人')}>
            <Select
              allowClear
              placeholder={t('选择机器人')}
              className={isBotDisabled ? 'select-disabled-value' : undefined}
              options={(botsQuery.data?.items || []).map((b) => ({ value: b.id, label: `${b.name} (${b.botType})${b.enabled ? '' : ' - ' + t('停用')}` }))}
            />
          </Form.Item>
          <Form.Item name="parseMode" label={t('解析模式')} tooltip={t('根据机器人类型选择，用于透传给IM服务器解析模板内容')} initialValue="">
            <Select options={
              (() => {
                const selectedBot = watchedBotId ? (botsQuery.data?.items || []).find((b) => b.id === watchedBotId) : undefined
                if (selectedBot?.botType === 'telegram') return TG_PARSE_MODES.map((o) => ({ value: o.value, label: t(o.label) }))
                if (selectedBot?.botType === 'feishu') return FEISHU_PARSE_MODES.map((o) => ({ value: o.value, label: t(o.label) }))
                return [{ value: '', label: t('默认空') }]
              })()
            } />
          </Form.Item>
          <Form.Item name="groupName" label={t('群名称')} tooltip={t('接收消息的群组名称，仅备注用')}>
            <Input maxLength={20} showCount />
          </Form.Item>
          <Form.Item name="groupId" label={t('群ID')} tooltip={t('群唯一标识，作为发送目标，必须从对应官方渠道获取')}>
            <Input maxLength={20} showCount />
          </Form.Item>
          <Form.Item label={t('通知模板')} tooltip={t('支持{变量}占位符，发送时替换')}>
            <Space.Compact style={{ width: '100%', marginBottom: 8 }}>
              <Select
                style={{ flex: 1 }}
                placeholder={t('从模板列表选择')}
                value={undefined}
                loading={templatesQuery.isFetching}
                options={(templateOptions || []).map((tpl) => {
                  const modeLabel = t(PARSE_MODE_LABEL[tpl.parseMode] || tpl.parseMode || '默认')
                  return { value: tpl.id, label: `${tpl.id}. ${modeLabel} - ${tpl.content.length > 50 ? tpl.content.slice(0, 50) + '...' : tpl.content}` }
                })}
                onChange={(id) => {
                  const tpl = (templateOptions || []).find((t) => t.id === id)
                  if (tpl) editorForm.setFieldsValue({ parseMode: tpl.parseMode, notifyTemplate: tpl.content })
                }}
              />
              <UiButton onClick={() => { void templatesQuery.refetch() }} disabled={!selectedBotType} loading={templatesQuery.isFetching}>{t('刷新')}</UiButton>
            </Space.Compact>
            <Form.Item name="notifyTemplate" noStyle rules={[{ max: 1000 }]}>
              <Input.TextArea maxLength={1000} showCount rows={3} placeholder={'服务器 {host} 于 {time} 触发告警：{message}'} />
            </Form.Item>
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={t('测试发送')}
        open={testSendOpen}
        onCancel={() => setTestSendOpen(false)}
        onOk={() => { if (testSendVars.length === 0) { testSendMutation.mutate({}) } else { testSendForm.submit() } }}
        confirmLoading={testSendMutation.isPending}
        okButtonProps={{ disabled: !testSendItem?.botId || !(botsQuery.data?.items || []).find((b) => b.id === testSendItem?.botId)?.enabled }}
        destroyOnHidden={false}
      >
        <p style={{ marginBottom: 12 }}>{t('通知模板')}: <code>{testSendItem?.notifyTemplate}</code></p>
        {testSendVars.length === 0 ? (
          <p>{t('该模板无变量，将直接发送原文')}</p>
        ) : (
          <Form form={testSendForm} layout="vertical" onFinish={(v) => testSendMutation.mutate(v)}>
            {testSendVars.map((varName) => (
              <Form.Item key={varName} name={varName} label={varName} rules={[{ required: true, message: t('请输入') + ' ' + varName }]}>
                <Input />
              </Form.Item>
            ))}
          </Form>
        )}
      </Modal>
    </>
  )
}

// ===== Template Tab =====

const TG_PARSE_MODES = [
  { value: '', label: '默认空' },
  { value: 'Markdown', label: 'TG-Markdown' },
  { value: 'MarkdownV2', label: 'TG-MarkdownV2' },
  { value: 'HTML', label: 'TG-HTML' },
]

const FEISHU_PARSE_MODES = [
  { value: '', label: '文本' },
  { value: 'post', label: '富文本' },
  { value: 'interactive', label: '消息卡片' },
]

const PARSE_MODE_LABEL: Record<string, string> = {
  '': '默认', Markdown: 'TG-Markdown', MarkdownV2: 'TG-MarkdownV2', HTML: 'TG-HTML',
  post: '富文本', interactive: '消息卡片',
}

function TemplateTab() {
  const { t } = useI18n()
  const systemTimezone = useUiSettingsStore((state) => state.systemTimezone)
  const [messageApi, contextHolder] = message.useMessage()
  const [modalApi, modalHolder] = Modal.useModal()
  const [editorForm] = Form.useForm()
  const watchedTemplateBotType = Form.useWatch('botType', editorForm)
  const [pageNo, setPageNo] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [orderField, setOrderField] = useState<string | undefined>('created_at')
  const [orderType, setOrderType] = useState<'asc' | 'desc' | undefined>('desc')
  const [keyword, setKeyword] = useState('')
  const [botTypeFilter, setBotTypeFilter] = useState('')
  const [queryTrigger, setQueryTrigger] = useState(0)
  const [editorOpen, setEditorOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<SystemAlertTemplateItem | null>(null)

  const listQuery = useQuery({
    queryKey: ['system-alert-templates', pageNo, pageSize, orderField, orderType, queryTrigger],
    queryFn: () => getSystemAlertTemplates(pageNo, pageSize, orderField, orderType, keyword, botTypeFilter),
  })

  const saveMutation = useMutation({
    mutationFn: async (values: any) => {
      await saveSystemAlertTemplate({ id: editingItem?.id, botType: values.botType, name: values.name, parseMode: values.parseMode || '', content: values.content })
    },
    onSuccess: async () => {
      void messageApi.success(editingItem ? t('保存成功') : t('创建成功'))
      setEditorOpen(false)
      await listQuery.refetch()
    },
    onError: (err: any) => { void messageApi.error(err?.message || t('操作失败')) },
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => deleteSystemAlertTemplate(id),
    onSuccess: async () => { void messageApi.success(t('删除成功')); await listQuery.refetch() },
  })

  const handleTableChange = (pagination: TablePaginationConfig, _: any, sorter: SorterResult<SystemAlertTemplateItem> | SorterResult<SystemAlertTemplateItem>[]) => {
    if (pagination.current) setPageNo(pagination.current)
    if (pagination.pageSize && pagination.pageSize !== pageSize) { setPageSize(pagination.pageSize); setPageNo(1) }
    const single = Array.isArray(sorter) ? sorter[0] : sorter
    if (!single?.field || !single.order) { setOrderField(undefined); setOrderType(undefined); return }
    const fieldMap: Record<string, string> = { id: 'id', botType: 'bot_type', name: 'name', parseMode: 'parse_mode', createdAt: 'created_at', updatedAt: 'updated_at' }
    setOrderField(fieldMap[single.field as string] || (single.field as string))
    setOrderType(single.order === 'ascend' ? 'asc' : 'desc')
  }

  const openEditor = (item?: SystemAlertTemplateItem) => {
    setEditingItem(item || null)
    editorForm.resetFields()
    editorForm.setFieldsValue({
      botType: item?.botType || 'telegram',
      name: item?.name || '',
      parseMode: item?.parseMode || '',
      content: item?.content || '',
    })
    setEditorOpen(true)
  }

  return (
    <>
      {contextHolder}
      {modalHolder}
      <Space direction="vertical" size={16} className="full-width">
        <Typography.Text type="secondary" style={{ fontSize: 13 }}>
          {t('说明：可在此页面添加常用的通知模板，以便在编辑场景配置时快速选择，并稍作修改后使用。')}
        </Typography.Text>
        <Space wrap>
          <UiButton type="primary" onClick={() => openEditor()}>{t('新增')}</UiButton>
        </Space>
        <Card>
          <Space wrap style={{ rowGap: 12 }}>
            <Input placeholder={t('搜索模板名称/内容')} allowClear value={keyword} onChange={(e) => setKeyword(e.target.value)} onPressEnter={() => { setPageNo(1); setQueryTrigger((n) => n + 1) }} style={{ width: 260 }} />
            <Select style={{ width: 140 }} allowClear placeholder={t('机器人类型')} value={botTypeFilter || undefined} onChange={(v) => setBotTypeFilter(v || '')} options={BOT_TYPE_OPTIONS.map((o) => ({ value: o.value, label: t(o.label) }))} />
            <UiButton type="primary" onClick={() => { setPageNo(1); setQueryTrigger((n) => n + 1) }}>{t('查询')}</UiButton>
            <UiButton onClick={() => { setKeyword(''); setBotTypeFilter(''); setPageNo(1); setQueryTrigger((n) => n + 1) }}>{t('重置')}</UiButton>
          </Space>
        </Card>
        <Card className="compact-table-card system-table-card table-scroll-region">
          <Table<SystemAlertTemplateItem>
            rowKey="id"
            loading={listQuery.isLoading || listQuery.isFetching}
            dataSource={listQuery.data?.items ?? []}
            scroll={{ x: 900, y: 340 }}
            pagination={{ current: pageNo, pageSize, total: listQuery.data?.total ?? 0, showSizeChanger: true, showTotal: (total) => t('共 {total} 条', { total }) }}
            onChange={handleTableChange}
            columns={[
              { title: 'ID', dataIndex: 'id', sorter: true, width: 70 },
              { title: t('机器人类型'), dataIndex: 'botType', width: 110, render: (v: string) => <Tag color={v === 'telegram' ? 'blue' : v === 'feishu' ? 'green' : 'default'}>{v}</Tag> },
              { title: t('模板名称'), dataIndex: 'name', sorter: true, width: 150 },
              { title: t('解析模式'), dataIndex: 'parseMode', width: 120, render: (v: string) => t(PARSE_MODE_LABEL[v] || v || '默认') },
              { title: t('模板内容'), dataIndex: 'content', width: 250, ellipsis: true, render: (v: string) => (
                <Space size={4}>
                  <Typography.Text ellipsis style={{ maxWidth: 200 }}>{v}</Typography.Text>
                  <Typography.Text copyable={{ text: v, tooltips: [t('复制'), t('已复制')] }} style={{ fontSize: 14 }} />
                </Space>
              ) },
              { title: t('创建时间'), dataIndex: 'createdAt', sorter: true, width: 170, defaultSortOrder: 'descend', render: (v: string) => formatDateTime(v, systemTimezone) },
              { title: t('更新时间'), dataIndex: 'updatedAt', sorter: true, width: 170, render: (v: string) => formatDateTime(v, systemTimezone) },
              { title: t('操作'), fixed: 'right', width: 120, render: (_, row) => (
                <Space size={0}>
                  <UiButton type="link" onClick={() => openEditor(row)}>{t('编辑')}</UiButton>
                  <UiButton type="link" danger loading={deleteMutation.isPending} onClick={() => { void modalApi.confirm({ title: t('确认删除 {name}？', { name: row.name }), onOk: async () => { await deleteMutation.mutateAsync(row.id) } }) }}>{t('删除')}</UiButton>
                </Space>
              ) },
            ]}
          />
        </Card>
      </Space>

      <Modal title={editingItem ? t('编辑模板') : t('新增模板')} open={editorOpen} onCancel={() => setEditorOpen(false)} onOk={() => editorForm.submit()} confirmLoading={saveMutation.isPending} destroyOnHidden={false}>
        <Form form={editorForm} layout="vertical" onFinish={(v) => saveMutation.mutate(v)} autoComplete="off">
          <Form.Item name="botType" label={t('机器人类型')} tooltip={t('模板适用的机器人平台')} rules={[{ required: true }]} initialValue="telegram">
            <Select options={BOT_TYPE_OPTIONS.map((o) => ({ value: o.value, label: t(o.label) }))} />
          </Form.Item>
          <Form.Item name="name" label={t('模板名称')} tooltip={t('便于识别的模板名称')} rules={[{ required: true }, { max: 30 }]}>
            <Input maxLength={30} showCount />
          </Form.Item>
          <Form.Item name="parseMode" label={t('解析模式')} tooltip={watchedTemplateBotType === 'feishu' ? t('飞书消息格式') : t('TG消息格式：默认为纯文本')} initialValue="" rules={[{ validator: (_, v) => { const pm = v || ''; if (watchedTemplateBotType === 'telegram' && !TG_PARSE_MODES.some((o) => o.value === pm)) return Promise.reject(t('解析模式与机器人类型不匹配')); if (watchedTemplateBotType === 'feishu' && !FEISHU_PARSE_MODES.some((o) => o.value === pm)) return Promise.reject(t('解析模式与机器人类型不匹配')); return Promise.resolve() } }]}>
            <Select options={
              (watchedTemplateBotType === 'feishu' ? FEISHU_PARSE_MODES : TG_PARSE_MODES).map((o) => ({ value: o.value, label: t(o.label) }))
            } />
          </Form.Item>
          <Form.Item name="content" label={t('模板内容')} tooltip={t('支持{变量}占位符和对应格式的标记语法')} rules={[{ required: true }, { max: 1000 }]}>
            <Input.TextArea maxLength={1000} showCount rows={5} placeholder={'*【{level}】{title}*\n_{message}_\n`{time}`'} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}

// ===== Page =====

export function SystemAlertBotsPage() {
  const { t } = useI18n()

  return (
    <div className="table-scroll-page" style={{ display: 'flex', flexDirection: 'column' }}>
      <div style={{ marginBottom: 8, color: 'var(--ant-color-text-secondary)', fontSize: 13 }}>
        {t('使用流程')}: ① {t('机器人配置')} → ② {t('场景配置')}
      </div>
      <Tabs
        defaultActiveKey="bots"
        className="alert-config-tabs"
        items={[
          { key: 'bots', label: t('机器人配置'), children: <BotTab /> },
          { key: 'scenes', label: t('场景配置'), children: <SceneTab /> },
          { key: 'templates', label: t('模板配置'), children: <TemplateTab /> },
        ]}
      />
    </div>
  )
}
