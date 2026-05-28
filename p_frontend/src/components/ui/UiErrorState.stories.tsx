import type { Meta, StoryObj } from '@storybook/react-vite'
import { UiErrorState } from './UiErrorState'

const meta = {
  title: 'UI/UiErrorState',
  component: UiErrorState,
  args: {
    title: '请求失败',
    description: '请稍后重试或联系管理员。',
  },
} satisfies Meta<typeof UiErrorState>

export default meta

type Story = StoryObj<typeof meta>

export const Basic: Story = {}

export const WithRetry: Story = {
  args: {
    onRetry: () => {
      console.info('retry clicked')
    },
  },
}
