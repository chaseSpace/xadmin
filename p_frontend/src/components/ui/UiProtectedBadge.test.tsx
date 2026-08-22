import { renderToStaticMarkup } from 'react-dom/server'
import type { CSSProperties, ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { UiProtectedBadge } from './UiProtectedBadge'

vi.mock('antd', () => ({
  Tooltip: ({ children, title }: { children: ReactNode; title: ReactNode }) => (
    <span data-tooltip={String(title)}>{children}</span>
  ),
  Tag: ({
    children,
    color,
    style,
    ...props
  }: {
    children: ReactNode
    color?: string
    style?: CSSProperties
    'aria-label'?: string
  }) => (
    <span data-color={color} style={style} {...props}>
      {children}
    </span>
  ),
}))

describe('UiProtectedBadge', () => {
  it('renders the shared protected badge with tooltip', () => {
    const html = renderToStaticMarkup(
      <UiProtectedBadge ariaLabel="Protected" tooltip="Super admins only" />,
    )

    expect(html).toContain('data-tooltip="Super admins only"')
    expect(html).toContain('data-color="red"')
    expect(html).toContain('aria-label="Protected"')
    expect(html).toContain('保')
  })
})
