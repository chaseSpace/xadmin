import type { Meta, StoryObj } from '@storybook/react-vite'
import { Space } from 'antd'
import { UiButton } from './UiButton'

const meta = {
  title: 'UI/UiButton',
  component: UiButton,
  args: {
    children: 'Primary Action',
    type: 'primary',
  },
} satisfies Meta<typeof UiButton>

export default meta

type Story = StoryObj<typeof meta>

export const Primary: Story = {}

export const Variants: Story = {
  render: () => (
    <Space>
      <UiButton type="primary">Primary</UiButton>
      <UiButton>Default</UiButton>
      <UiButton type="dashed">Dashed</UiButton>
      <UiButton type="link">Link</UiButton>
    </Space>
  ),
}
