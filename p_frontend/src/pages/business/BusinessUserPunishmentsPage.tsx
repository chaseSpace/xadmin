import { Card, DatePicker, Drawer, Form, Input, Modal, Select, Space, Table, Tag, Typography, message } from 'antd'
import { useState } from 'react'
import type { Key } from 'react'
import { useDebouncedCallback } from '../../hooks/useDebouncedCallback'
import { UiButton } from '../../components/ui'
import { useI18n } from '../../i18n/messages'

type PunishmentRecord = {
  id: string
  userId: string
  userName: string
  type: 'mute' | 'restrict-login' | 'ban'
  level: 'L1' | 'L2' | 'L3'
  startAt: string
  endAt: string
  reason: string
  operator: string
  status: 'active' | 'revoked' | 'expired'
}

const data: PunishmentRecord[] = [
  {
    id: 'p-001',
    userId: 'bu1002',
    userName: '橙子汽水',
    type: 'mute',
    level: 'L1',
    startAt: '2026-04-15 09:00',
    endAt: '2026-04-20 09:00',
    reason: '评论区违规发言',
    operator: 'risk_admin',
    status: 'active',
  },
  {
    id: 'p-002',
    userId: 'bu0933',
    userName: 'DawnWind',
    type: 'restrict-login',
    level: 'L2',
    startAt: '2026-04-01 10:00',
    endAt: '2026-04-08 10:00',
    reason: '疑似异常登录行为',
    operator: 'risk_admin',
    status: 'expired',
  },
  {
    id: 'p-003',
    userId: 'bu0668',
    userName: 'BlueRain',
    type: 'ban',
    level: 'L3',
    startAt: '2026-03-10 12:00',
    endAt: '2099-12-31 00:00',
    reason: '多次严重违规',
    operator: 'super_admin',
    status: 'revoked',
  },
]

type Translate = (text: string, params?: Record<string, string | number>) => string

function renderType(type: PunishmentRecord['type'], t: Translate) {
  if (type === 'mute') return <Tag>{t('禁言')}</Tag>
  if (type === 'restrict-login') return <Tag color="orange">{t('限制登录')}</Tag>
  return <Tag color="red">{t('封禁')}</Tag>
}

function renderStatus(status: PunishmentRecord['status'], t: Translate) {
  if (status === 'active') return <Tag color="red">{t('生效')}</Tag>
  if (status === 'revoked') return <Tag color="blue">{t('撤销')}</Tag>
  return <Tag>{t('到期')}</Tag>
}

export function BusinessUserPunishmentsPage() {
  const { t } = useI18n()
  const [messageApi, contextHolder] = message.useMessage()
  const [modalApi, modalContextHolder] = Modal.useModal()
  const [filterForm] = Form.useForm()
  const onQuery = useDebouncedCallback(() => {
    void messageApi.success(t('已按筛选条件查询'))
  })
  const [selectedRowKeys, setSelectedRowKeys] = useState<Key[]>([])
  const [createOpen, setCreateOpen] = useState(false)
  const [detailOpen, setDetailOpen] = useState(false)
  const [selected, setSelected] = useState<PunishmentRecord | null>(null)

  return (
    <>
      {contextHolder}
      {modalContextHolder}

      <Space direction="vertical" size={16} className="full-width table-scroll-page">
      <Space wrap>
        <UiButton type="primary" onClick={() => setCreateOpen(true)}>
          {t('新增惩罚')}
        </UiButton>
        <UiButton
          onClick={() => {
            if (selectedRowKeys.length === 0) {
              void messageApi.warning(t('请先勾选需要撤销的惩罚记录'))
              return
            }
            void messageApi.success(t('批量撤销成功，共 {count} 条', { count: selectedRowKeys.length }))
          }}
        >
          {t('批量撤销')}
        </UiButton>
        <UiButton onClick={() => void messageApi.success(t('导出记录成功'))}>{t('导出记录')}</UiButton>
      </Space>

      <Card>
        <Form
          form={filterForm}
          layout="inline"
          style={{ rowGap: 12 }}
          onFinish={onQuery}
          onKeyDown={(event) => {
            if (event.key === 'Enter') {
              event.preventDefault()
              filterForm.submit()
            }
          }}
        >
          <Form.Item label={t('用户ID/手机号')} name="user">
            <Input placeholder={t('请输入用户ID或手机号')} allowClear onPressEnter={() => filterForm.submit()} />
          </Form.Item>
          <Form.Item label={t('惩罚类型')} name="type">
            <Select
              style={{ width: 170 }}
              allowClear
              options={[
                { value: 'mute', label: t('禁言') },
                { value: 'restrict-login', label: t('限制登录') },
                { value: 'ban', label: t('封禁') },
              ]}
            />
          </Form.Item>
          <Form.Item label={t('状态')} name="status">
            <Select
              style={{ width: 140 }}
              allowClear
              options={[
                { value: 'active', label: t('生效') },
                { value: 'revoked', label: t('撤销') },
                { value: 'expired', label: t('到期') },
              ]}
            />
          </Form.Item>
          <Form.Item label={t('时间范围')} name="time">
            <DatePicker.RangePicker />
          </Form.Item>
          <Form.Item>
            <Space size={8}>
              <UiButton type="primary" htmlType="submit">
                {t('查询')}
              </UiButton>
              <UiButton
                onClick={() => {
                  filterForm.resetFields()
                  void messageApi.success(t('已重置筛选条件'))
                }}
              >
                {t('重置')}
              </UiButton>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      <Card className="compact-table-card system-table-card table-scroll-region">
        <Table<PunishmentRecord>
          scroll={{ x: 'max-content', y: 392 }}
          rowKey="id"
          rowSelection={{
            selectedRowKeys,
            onChange: (keys) => setSelectedRowKeys(keys),
          }}
          dataSource={data}
          pagination={{ pageSize: 10 }}
          columns={[
            { title: t('记录ID'), dataIndex: 'id' },
            { title: t('用户ID'), dataIndex: 'userId' },
            { title: t('用户姓名'), dataIndex: 'userName' },
            { title: t('惩罚类型'), dataIndex: 'type', render: (value: PunishmentRecord['type']) => renderType(value, t) },
            { title: t('惩罚等级'), dataIndex: 'level' },
            { title: t('生效时间'), dataIndex: 'startAt' },
            { title: t('失效时间'), dataIndex: 'endAt' },
            { title: t('触发原因'), dataIndex: 'reason', render: (value: string) => t(value) },
            { title: t('执行人'), dataIndex: 'operator' },
            { title: t('状态'), dataIndex: 'status', render: (value: PunishmentRecord['status']) => renderStatus(value, t) },
            {
              title: t('操作'),
              render: (_, row) => (
                <Space size={0}>
                  <UiButton
                    type="link"
                    onClick={() => {
                      setSelected(row)
                      setDetailOpen(true)
                    }}
                  >
                    {t('查看详情')}
                  </UiButton>
                  <UiButton
                    type="link"
                    onClick={() => {
                      void modalApi.confirm({
                        title: t('确认撤销惩罚 {id}', { id: row.id }),
                        onOk: async () => {
                          await messageApi.success(t('撤销成功'))
                        },
                      })
                    }}
                  >
                    {t('撤销')}
                  </UiButton>
                </Space>
              ),
            },
          ]}
        />
      </Card>
      </Space>

      <Modal
        title={t('新增惩罚')}
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={() => {
          setCreateOpen(false)
          void messageApi.success(t('新增惩罚成功'))
        }}
      >
        <Form layout="vertical">
          <Form.Item label={t('用户ID')} required>
            <Input />
          </Form.Item>
          <Form.Item label={t('惩罚类型')} required>
            <Select
              options={[
                { value: 'mute', label: t('禁言') },
                { value: 'restrict-login', label: t('限制登录') },
                { value: 'ban', label: t('封禁') },
              ]}
            />
          </Form.Item>
          <Form.Item label={t('触发原因')} required>
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer title={t('惩罚详情')} open={detailOpen} onClose={() => setDetailOpen(false)} width={420}>
        <Typography.Paragraph>{t('记录ID：')}{selected?.id ?? '-'}</Typography.Paragraph>
        <Typography.Paragraph>{t('用户：')}{selected?.userName ?? '-'}</Typography.Paragraph>
        <Typography.Paragraph>{t('类型：')}{selected ? t(selected.type === 'mute' ? '禁言' : selected.type === 'restrict-login' ? '限制登录' : '封禁') : '-'}</Typography.Paragraph>
        <Typography.Paragraph>{t('原因：')}{selected?.reason ? t(selected.reason) : '-'}</Typography.Paragraph>
      </Drawer>
    </>
  )
}
