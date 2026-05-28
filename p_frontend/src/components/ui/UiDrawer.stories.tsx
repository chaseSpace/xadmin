import type { Meta, StoryObj } from '@storybook/react-vite'
import { useState } from 'react'
import { UiButton } from './UiButton'
import { UiDrawer } from './UiDrawer'

const meta = {
  title: 'UI/UiDrawer',
  component: UiDrawer,
} satisfies Meta<typeof UiDrawer>

export default meta

type Story = StoryObj<typeof meta>

function DrawerDemo() {
  const [open, setOpen] = useState(false)

  return (
    <>
      <UiButton type="primary" onClick={() => setOpen(true)}>
        打开抽屉
      </UiButton>
      <UiDrawer
        title="筛选条件"
        open={open}
        width={420}
        onClose={() => setOpen(false)}
      >
        这里可以放置表单筛选项、操作说明或辅助配置内容。
      </UiDrawer>
    </>
  )
}

export const Basic: Story = {
  render: () => <DrawerDemo />,
}
