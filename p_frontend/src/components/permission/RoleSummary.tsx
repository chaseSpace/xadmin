import { Tag, Tooltip } from 'antd'

type RoleSummaryProps = {
  roleNames?: readonly string[]
  roleIds?: readonly number[]
  emptyText?: string
}

export function buildRoleLabels(
  roleNames: readonly string[] = [],
  roleIds: readonly number[] = [],
): string[] {
  const count = Math.max(roleNames.length, roleIds.length)
  const labels: string[] = []
  for (let index = 0; index < count; index += 1) {
    const name = String(roleNames[index] ?? '').trim()
    const id = Number(roleIds[index] ?? 0)
    const label = name || (Number.isInteger(id) && id > 0 ? `角色#${id}` : '')
    if (label && !labels.includes(label)) labels.push(label)
  }
  return labels
}

export function RoleSummary({ roleNames = [], roleIds = [], emptyText = '-' }: RoleSummaryProps) {
  const labels = buildRoleLabels(roleNames, roleIds)
  if (labels.length === 0) return <span>{emptyText}</span>

  const visible = labels.slice(0, 3)
  const overflow = labels.slice(3)

  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start', gap: 4 }}>
      {visible.map((label, index) => (
        <div key={label} style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          <Tag
            color="blue"
            style={{
              marginInlineEnd: 0,
              maxWidth: 160,
              overflow: 'hidden',
              textOverflow: 'ellipsis',
            }}
          >
            {label}
          </Tag>
          {index === visible.length - 1 && overflow.length > 0 ? (
            <Tooltip
              title={
                <div>
                  {overflow.map((item) => (
                    <div key={item}>{item}</div>
                  ))}
                </div>
              }
            >
              <Tag style={{ marginInlineEnd: 0, cursor: 'help' }}>+{overflow.length}</Tag>
            </Tooltip>
          ) : null}
        </div>
      ))}
    </div>
  )
}
