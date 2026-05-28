import type { Meta, StoryObj } from '@storybook/react-vite'
import { Button, Form, Input, Space } from 'antd'
import { UiForm } from './UiForm'

type LoginValues = {
  email: string
  password: string
}

const meta = {
  title: 'UI/UiForm',
  component: UiForm,
} satisfies Meta<typeof UiForm>

export default meta

type Story = StoryObj<typeof meta>

export const LoginForm: Story = {
  render: () => (
    <UiForm<LoginValues> style={{ maxWidth: 420 }}>
      <Form.Item<LoginValues>
        label="邮箱"
        name="email"
        rules={[{ required: true, message: '请输入邮箱' }]}
      >
        <Input placeholder="admin@xadmin.local" />
      </Form.Item>
      <Form.Item<LoginValues>
        label="密码"
        name="password"
        rules={[{ required: true, message: '请输入密码' }]}
      >
        <Input.Password placeholder="请输入密码" />
      </Form.Item>
      <Form.Item>
        <Space>
          <Button type="primary" htmlType="submit">
            登录
          </Button>
          <Button htmlType="button">重置</Button>
        </Space>
      </Form.Item>
    </UiForm>
  ),
}
