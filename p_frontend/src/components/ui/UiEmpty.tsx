import { Empty } from 'antd'
import type { EmptyProps } from 'antd'

export type UiEmptyProps = EmptyProps

export function UiEmpty(props: UiEmptyProps) {
  return (
    <div className="ui-empty">
      <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} {...props} />
    </div>
  )
}
