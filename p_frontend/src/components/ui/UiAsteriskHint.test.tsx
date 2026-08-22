import { renderToStaticMarkup } from 'react-dom/server'
import type { CSSProperties, ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { UiAsteriskHint } from './UiAsteriskHint'

vi.mock('antd', () => ({
  Typography: {
    Text: ({
      children,
      className,
      role,
      style,
    }: {
      children: ReactNode
      className?: string
      role?: string
      style?: CSSProperties
    }) => (
      <span className={className} role={role} style={style}>
        {children}
      </span>
    ),
  },
  theme: {
    useToken: () => ({
      token: {
        colorWarningActive: '#d48806',
        colorErrorActive: '#d9363e',
        colorWarningText: '#faad14',
      },
    }),
  },
}))

vi.mock('../../store/theme', () => ({
  useThemeStore: (selector: (state: { mode: 'light' }) => unknown) =>
    selector({ mode: 'light' }),
}))

describe('UiAsteriskHint', () => {
  it('renders a shared asterisk note style', () => {
    const html = renderToStaticMarkup(<UiAsteriskHint>Protected scope</UiAsteriskHint>)

    expect(html).toContain('role="note"')
    expect(html).toContain('aria-hidden="true"')
    expect(html).toContain('*')
    expect(html).toContain('Protected scope')
    expect(html).toContain('margin-top:8px')
    expect(html).toContain('color-mix')
  })
})
