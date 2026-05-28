import type { Meta, StoryObj } from '@storybook/react-vite'
import { useState } from 'react'
import { UiButton } from './UiButton'
import { UiModal } from './UiModal'

const meta = {
  title: 'UI/UiModal',
  component: UiModal,
} satisfies Meta<typeof UiModal>

export default meta

type Story = StoryObj<typeof meta>

function ModalDemo() {
  const [open, setOpen] = useState(false)

  return (
    <>
      <UiButton type="primary" onClick={() => setOpen(true)}>
        打开弹窗
      </UiButton>
      <UiModal
        title="确认操作"
        open={open}
        onCancel={() => setOpen(false)}
        onOk={() => setOpen(false)}
      >
        这是一个通用 Modal 组件示例。
      </UiModal>
    </>
  )
}

export const Basic: Story = {
  render: () => <ModalDemo />,
}
