import type { Meta, StoryObj } from '@storybook/react-vite'
import { UiEmpty } from './UiEmpty'

const meta = {
  title: 'UI/UiEmpty',
  component: UiEmpty,
  args: {
    description: '暂无数据',
  },
} satisfies Meta<typeof UiEmpty>

export default meta

type Story = StoryObj<typeof meta>

export const Default: Story = {}
