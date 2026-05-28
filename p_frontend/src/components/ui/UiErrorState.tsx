import { Result, Space, Typography } from 'antd'
import { UiButton } from './UiButton'

type UiErrorStateProps = {
  title?: string
  description?: string
  onRetry?: () => void
}

export function UiErrorState({
  title = '请求失败',
  description = '请稍后重试或联系管理员。',
  onRetry,
}: UiErrorStateProps) {
  return (
    <Result
      status="error"
      title={title}
      subTitle={<Typography.Text type="secondary">{description}</Typography.Text>}
      extra={
        onRetry ? (
          <Space>
            <UiButton type="primary" onClick={onRetry}>
              重试
            </UiButton>
          </Space>
        ) : null
      }
    />
  )
}
