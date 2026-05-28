import { Card, DatePicker, Drawer, Form, Input, Modal, Select, Space, Table, Tag, Typography, message } from 'antd'
import { useState } from 'react'
import type { Key } from 'react'
import { useDebouncedCallback } from '../../hooks/useDebouncedCallback'
import { UiButton } from '../../components/ui'
import { useI18n } from '../../i18n/messages'
import { useUiSettingsStore } from '../../store/uiSettings'
import { formatDateTime } from '../../utils/timezone'

type BusinessUser = {
  id: string
  nickname: string
  phone: string
  channel: string
  verified: boolean
  status: 'active' | 'frozen' | 'closed'
  orders: number
  activeAt: string
  createdAt: string
}

const data: BusinessUser[] = [
  {
    id: 'bu1001',
    nickname: '星海漫游',
    phone: '138****0011',
    channel: 'App',
    verified: true,
    status: 'active',
    orders: 42,
    activeAt: '2026-04-17 19:10',
    createdAt: '2025-09-10 14:11',
  },
  {
    id: 'bu1002',
    nickname: '橙子汽水',
    phone: '139****2234',
    channel: 'H5',
    verified: true,
    status: 'frozen',
    orders: 7,
    activeAt: '2026-04-16 14:03',
    createdAt: '2026-02-21 09:20',
  },
  {
    id: 'bu1003',
    nickname: 'SilentFox',
    phone: '137****8879',
    channel: 'MiniProgram',
    verified: false,
    status: 'closed',
    orders: 0,
    activeAt: '2026-04-10 08:41',
    createdAt: '2026-03-05 08:11',
  },
]

function renderStatus(status: BusinessUser['status'], t: (text: string, params?: Record<string, string | number>) => string) {
  if (status === 'active') return <Tag color="green">{t('正常')}</Tag>
  if (status === 'frozen') return <Tag color="orange">{t('冻结')}</Tag>
  return <Tag>{t('注销')}</Tag>
}
export function BusinessUsersPage() {
  const { t } = useI18n()
  const systemTimezone = useUiSettingsStore((state) => state.systemTimezone)
  const [messageApi, contextHolder] = message.useMessage()
  const [modalApi, modalContextHolder] = Modal.useModal()
  const [filterForm] = Form.useForm()
  const onQuery = useDebouncedCallback(() => {
    void messageApi.success(t('已按筛选条件查询'))
  })
  const [selectedRowKeys, setSelectedRowKeys] = useState<Key[]>([])
  const [detailOpen, setDetailOpen] = useState(false)
  const [tagOpen, setTagOpen] = useState(false)
  const [selected, setSelected] = useState<BusinessUser | null>(null)

  return (
    <>
      {contextHolder}
      {modalContextHolder}

      <Space direction="vertical" size={16} className="full-width table-scroll-page">
      <Space wrap>
        <UiButton type="primary" onClick={() => void messageApi.success(t('打开新增业务用户弹窗'))}>
          {t('新增业务用户')}
        </UiButton>
        <UiButton
          onClick={() => {
            if (selectedRowKeys.length === 0) {
              void messageApi.warning(t('请先勾选需要导出的用户'))
              return
            }
            void messageApi.success(t('批量导出成功，共 {count} 条', { count: selectedRowKeys.length }))
          }}
        >
          {t('批量导出')}
        </UiButton>
        <UiButton
          onClick={() => {
            if (selectedRowKeys.length === 0) {
              void messageApi.warning(t('请先勾选需要标记的用户'))
              return
            }
            void messageApi.success(t('批量标记成功，共 {count} 条', { count: selectedRowKeys.length }))
          }}
        >
          {t('批量标记')}
        </UiButton>
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
          <Form.Item label={t('用户ID')} name="userId">
            <Input placeholder={t('请输入用户ID')} allowClear onPressEnter={() => filterForm.submit()} />
          </Form.Item>
          <Form.Item label={t('昵称')} name="nickname">
            <Input placeholder={t('请输入昵称')} allowClear onPressEnter={() => filterForm.submit()} />
          </Form.Item>
          <Form.Item label={t('手机号')} name="phone">
            <Input placeholder={t('请输入手机号')} allowClear onPressEnter={() => filterForm.submit()} />
          </Form.Item>
          <Form.Item label={t('渠道来源')} name="channel">
            <Select
              style={{ width: 160 }}
              allowClear
              options={[
                { value: 'App', label: 'App' },
                { value: 'H5', label: 'H5' },
                { value: 'MiniProgram', label: 'MiniProgram' },
              ]}
            />
          </Form.Item>
          <Form.Item label={t('账号状态')} name="status">
            <Select
              style={{ width: 160 }}
              allowClear
              options={[
                { value: 'active', label: t('正常') },
                { value: 'frozen', label: t('冻结') },
                { value: 'closed', label: t('注销') },
              ]}
            />
          </Form.Item>
          <Form.Item label={t('注册时间')} name="createdAt">
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
        <Table<BusinessUser>
          scroll={{ x: 'max-content', y: 392 }}
          rowKey="id"
          rowSelection={{
            selectedRowKeys,
            onChange: (keys) => setSelectedRowKeys(keys),
          }}
          dataSource={data}
          pagination={{ pageSize: 10 }}
          columns={[
            { title: t('用户ID'), dataIndex: 'id' },
            { title: t('昵称'), dataIndex: 'nickname' },
            { title: t('手机号'), dataIndex: 'phone' },
            { title: t('注册渠道'), dataIndex: 'channel' },
            {
              title: t('实名状态'),
              dataIndex: 'verified',
              render: (verified: boolean) => (
                <Tag color={verified ? 'blue' : 'default'}>{verified ? t('已实名') : t('未实名')}</Tag>
              ),
            },
            {
              title: t('账号状态'),
              dataIndex: 'status',
              render: (status: BusinessUser['status']) => renderStatus(status, t),
            },
            { title: t('累计订单数'), dataIndex: 'orders' },
            { title: t('最近活跃时间'), dataIndex: 'activeAt', render: (value: string) => formatDateTime(value, systemTimezone) },
            { title: t('注册时间'), dataIndex: 'createdAt', render: (value: string) => formatDateTime(value, systemTimezone) },
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
                      setSelected(row)
                      setTagOpen(true)
                    }}
                  >
                    {t('编辑标签')}
                  </UiButton>
                  <UiButton
                    type="link"
                    onClick={() => {
                      void modalApi.confirm({
                        title: t('确认冻结账号 {name}', { name: row.nickname }),
                        onOk: async () => {
                          await messageApi.success(t('冻结账号成功'))
                        },
                      })
                    }}
                  >
                    {t('冻结账号')}
                  </UiButton>
                </Space>
              ),
            },
          ]}
        />
      </Card>
      </Space>

      <Drawer title={t('业务用户详情')} open={detailOpen} onClose={() => setDetailOpen(false)} width={420}>
        <Typography.Paragraph>{t('用户ID')}：{selected?.id ?? '-'}</Typography.Paragraph>
        <Typography.Paragraph>{t('昵称')}：{selected?.nickname ?? '-'}</Typography.Paragraph>
        <Typography.Paragraph>{t('手机号')}：{selected?.phone ?? '-'}</Typography.Paragraph>
        <Typography.Paragraph>{t('账号状态：')}{selected ? t(selected.status === 'active' ? '正常' : selected.status === 'frozen' ? '冻结' : '注销') : '-'}</Typography.Paragraph>
      </Drawer>

      <Modal
        title={t('编辑标签')}
        open={tagOpen}
        onCancel={() => setTagOpen(false)}
        onOk={() => {
          setTagOpen(false)
          void messageApi.success(t('标签更新成功'))
        }}
      >
        <Form layout="vertical">
          <Form.Item label={t('标签')}>
            <Input placeholder={t('请输入标签')} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
