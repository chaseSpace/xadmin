import { Typography, theme } from 'antd'
import type { CSSProperties, ReactNode } from 'react'
import { useThemeStore } from '../../store/theme'

export type UiAsteriskHintProps = {
  children: ReactNode
  className?: string
  style?: CSSProperties
}

export function UiAsteriskHint({ children, className, style }: UiAsteriskHintProps) {
  const { token } = theme.useToken()
  const themeMode = useThemeStore((state) => state.mode)
  const color =
    themeMode === 'light'
      ? `color-mix(in srgb, ${token.colorWarningActive} 55%, ${token.colorErrorActive} 45%)`
      : token.colorWarningText

  return (
    <Typography.Text
      className={className}
      role="note"
      style={{ color, display: 'block', marginTop: 8, ...style }}
    >
      <span aria-hidden="true">*</span> {children}
    </Typography.Text>
  )
}
