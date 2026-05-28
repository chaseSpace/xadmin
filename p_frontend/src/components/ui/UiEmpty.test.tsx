import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it, vi } from 'vitest'
import { UiEmpty } from './UiEmpty'

vi.mock('antd', () => {
  const Empty = ({ description }: { description?: string }) => (
    <div className="ant-empty">{description}</div>
  )
  Empty.PRESENTED_IMAGE_SIMPLE = 'simple'
  return { Empty }
})

describe('UiEmpty', () => {
  it('renders provided description with wrapper class', () => {
    const html = renderToStaticMarkup(<UiEmpty description="No data" />)

    expect(html).toContain('No data')
    expect(html).toContain('ui-empty')
  })
})
