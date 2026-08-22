import { Tag, Tooltip } from 'antd'
import type { CSSProperties, ReactNode } from 'react'

export type UiProtectedBadgeProps = {
  tooltip: ReactNode
  ariaLabel?: string
  className?: string
  style?: CSSProperties
}

export function UiProtectedBadge({
  tooltip,
  ariaLabel,
  className,
  style,
}: UiProtectedBadgeProps) {
  return (
    <Tooltip title={tooltip}>
      <Tag
        aria-label={ariaLabel}
        className={className}
        color="red"
        style={{ cursor: 'help', fontWeight: 600, marginInlineEnd: 0, ...style }}
      >
        保
      </Tag>
    </Tooltip>
  )
}
