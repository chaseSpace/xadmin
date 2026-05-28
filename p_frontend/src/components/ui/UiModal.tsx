import { Modal } from 'antd'
import type { ModalProps } from 'antd'

export type UiModalProps = ModalProps

export function UiModal(props: UiModalProps) {
  return <Modal destroyOnHidden {...props} />
}
