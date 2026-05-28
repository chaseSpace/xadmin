import type { Meta, StoryObj } from '@storybook/react-vite'
import type { TableColumnsType } from 'antd'
import { Tag } from 'antd'
import { UiTable } from './UiTable'

type AdminUserRow = {
  id: string
  name: string
  email: string
  role: 'owner' | 'admin' | 'viewer'
}

const columns: TableColumnsType<AdminUserRow> = [
  {
    title: '姓名',
    dataIndex: 'name',
    key: 'name',
  },
  {
    title: '邮箱',
    dataIndex: 'email',
    key: 'email',
  },
  {
    title: '角色',
    dataIndex: 'role',
    key: 'role',
    render: (role: AdminUserRow['role']) => {
      if (role === 'owner') {
        return <Tag color="gold">Owner</Tag>
      }
      if (role === 'admin') {
        return <Tag color="blue">Admin</Tag>
      }
      return <Tag>Viewer</Tag>
    },
  },
]

const dataSource: AdminUserRow[] = [
  {
    id: 'u_01',
    name: 'Lynn Chen',
    email: 'lynn@xadmin.local',
    role: 'owner',
  },
  {
    id: 'u_02',
    name: 'Alex Wu',
    email: 'alex@xadmin.local',
    role: 'admin',
  },
  {
    id: 'u_03',
    name: 'Mia Zhang',
    email: 'mia@xadmin.local',
    role: 'viewer',
  },
]

const meta = {
  title: 'UI/UiTable',
  component: UiTable,
} satisfies Meta<typeof UiTable>

export default meta

type Story = StoryObj<typeof meta>

export const Basic: Story = {
  render: () => (
    <UiTable<AdminUserRow>
      columns={columns}
      dataSource={dataSource}
      pagination={false}
      rowKey="id"
    />
  ),
}
