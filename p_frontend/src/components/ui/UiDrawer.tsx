import { Drawer } from 'antd'
import type { DrawerProps } from 'antd'

export type UiDrawerProps = DrawerProps

export function UiDrawer(props: UiDrawerProps) {
  return <Drawer destroyOnHidden {...props} />
}
