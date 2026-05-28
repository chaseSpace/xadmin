import { zodResolver } from '@hookform/resolvers/zod'
import { useNavigate } from '@tanstack/react-router'
import { Alert, Card, Input, Space, Typography, message } from 'antd'
import { Controller, useForm } from 'react-hook-form'
import { useAuthStore } from '../store/auth'
import { UiButton } from '../components/ui'
import { useI18n } from '../i18n/messages'
import { useLoginMutation } from '../services/api/auth'
import { normalizeApiError } from '../services/api/error'
import { loginRoute } from '../app/router'
import {
  loginRequestSchema,
  type LoginRequestInput,
} from '../services/schemas/auth'
import { WaveBackground } from './login/WaveBackground'

export function LoginPage() {
  const navigate = useNavigate()
  const search = loginRoute.useSearch()
  const login = useAuthStore((state) => state.login)
  const loginMutation = useLoginMutation()
  const { t } = useI18n()
  const [messageApi, contextHolder] = message.useMessage()

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginRequestInput>({
    resolver: zodResolver(loginRequestSchema),
    defaultValues: {
      username: 'admin',
      password: '123456',
    },
  })

  const onSubmit = async (values: LoginRequestInput) => {
    try {
      const result = await loginMutation.mutateAsync(values)
      login({
        token: result.token,
        user: {
          uid: result.uid,
          username: result.username,
          displayName: result.displayName,
          avatar: result.avatar,
          sessionId: result.sessionId,
          menuRoutes: [],
          menuItems: [],
          warmTip: null,
          menuLoaded: false,
          menuLoadError: '',
        },
      })
      await messageApi.success(t('登录成功'))

      if (search.redirect) {
        window.location.assign(search.redirect)
        return
      }

      await navigate({ to: '/' })
    } catch (error) {
      const errorMessage = normalizeApiError(error).message || t('登录失败，请稍后重试')
      void messageApi.error(errorMessage)
    }
  }

  return (
    <div className="login-page">
      <WaveBackground />
      {contextHolder}
      <Card className="login-card" variant="borderless">
        <Space direction="vertical" size={16} className="full-width">
          <Typography.Title level={3} style={{ margin: 0 }}>
            {t('XAdmin 登录')}
          </Typography.Title>
          <Typography.Text type="secondary">
            {t('使用默认演示账号可直接进入后台框架。')}
          </Typography.Text>
          <Alert message={t('演示账号：admin / 123456')} type="info" showIcon />
          {search.reason === 'expired' ? (
            <Alert message={t('登录已过期，请重新登录')} type="warning" showIcon />
          ) : null}

          <Controller
            name="username"
            control={control}
            render={({ field }) => (
              <Input {...field} size="large" placeholder={t('请输入用户名')} autoComplete="username" />
            )}
          />
          {errors.username ? <Typography.Text type="danger">{errors.username.message}</Typography.Text> : null}

          <Controller
            name="password"
            control={control}
            render={({ field }) => (
              <Input.Password
                {...field}
                size="large"
                placeholder={t('请输入密码')}
                autoComplete="current-password"
              />
            )}
          />
          {errors.password ? <Typography.Text type="danger">{errors.password.message}</Typography.Text> : null}

          <UiButton
            type="primary"
            size="large"
            loading={loginMutation.isPending}
            requestLoadingDelayMs={1000}
            onClick={handleSubmit(onSubmit)}
            className="full-width"
          >
            {t('登录')}
          </UiButton>
        </Space>
      </Card>
    </div>
  )
}
