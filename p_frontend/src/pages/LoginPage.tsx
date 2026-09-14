import { zodResolver } from '@hookform/resolvers/zod'
import { useNavigate } from '@tanstack/react-router'
import { Alert, Card, Checkbox, Input, Space, Typography, message } from 'antd'
import { Controller, useForm } from 'react-hook-form'
import { useRef, useState } from 'react'
import { useAuthStore } from '../store/auth'
import { UiButton } from '../components/ui'
import { useI18n } from '../i18n/messages'
import { useLoginMutation } from '../services/api/auth'
import { normalizeApiError } from '../services/api/error'
import { loginRoute } from '../app/router'
import { loginRequestSchema, type LoginRequestInput } from '../services/schemas/auth'
import { LoginTextPressureBackground } from './login/LoginTextPressureBackground'
import {
  clearRememberedLoginCredentials,
  getRememberedLoginCredentials,
  saveRememberedLoginCredentials,
} from './login/rememberedCredentials'

const isDemoMode = import.meta.env.VITE_DEMO === 'true'

export function LoginPage() {
  const navigate = useNavigate()
  const search = loginRoute.useSearch()
  const login = useAuthStore((state) => state.login)
  const loginMutation = useLoginMutation()
  const { t } = useI18n()
  const [messageApi, contextHolder] = message.useMessage()
  const submittingRef = useRef(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [rememberedCredentials] = useState(() => getRememberedLoginCredentials())
  const [rememberPassword, setRememberPassword] = useState(Boolean(rememberedCredentials))

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginRequestInput>({
    resolver: zodResolver(loginRequestSchema),
    defaultValues: {
      username: rememberedCredentials?.username ?? '',
      password: rememberedCredentials?.password ?? '',
    },
  })

  const onSubmit = async (values: LoginRequestInput) => {
    if (submittingRef.current) return
    submittingRef.current = true
    setIsSubmitting(true)

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
          permissionKeys: [],
          isSuperAdmin: false,
          warmTip: null,
          menuLoaded: false,
          menuLoadError: '',
        },
      })

      if (rememberPassword) {
        saveRememberedLoginCredentials(values)
      } else {
        clearRememberedLoginCredentials()
      }

      await messageApi.success(t('登录成功'))

      if (search.redirect) {
        window.location.assign(search.redirect)
        return
      }

      await navigate({ to: '/' })
    } catch (error) {
      submittingRef.current = false
      setIsSubmitting(false)
      const errorMessage = normalizeApiError(error).message || t('登录失败，请稍后重试')
      void messageApi.error(errorMessage)
    }
  }

  const handleRememberPasswordChange = (nextRememberPassword: boolean) => {
    setRememberPassword(nextRememberPassword)
    if (!nextRememberPassword) {
      clearRememberedLoginCredentials()
    }
  }

  return (
    <div className="login-page">
      <LoginTextPressureBackground />
      {contextHolder}
      <Card className="login-card" variant="borderless">
        <form onSubmit={handleSubmit(onSubmit)}>
          <Space direction="vertical" size={16} className="full-width">
            <Typography.Title level={3} style={{ margin: 0 }}>
              {t('XAdmin 登录')}
            </Typography.Title>
            <Typography.Text type="secondary">{t('请输入您的账户名和密码')}</Typography.Text>
            {isDemoMode ? (
              <Typography.Text type="secondary">
                {t('使用默认演示账号可直接进入后台框架。')}
              </Typography.Text>
            ) : null}
            {isDemoMode ? (
              <Alert message={t('演示账号：admin / 123456')} type="info" showIcon />
            ) : null}
            {search.reason === 'expired' ? (
              <Alert message={t('登录已过期，请重新登录')} type="warning" showIcon />
            ) : null}

            <Controller
              name="username"
              control={control}
              render={({ field }) => (
                <Input
                  {...field}
                  size="large"
                  placeholder={t('请输入用户名')}
                  autoComplete="username"
                />
              )}
            />
            {errors.username ? (
              <Typography.Text type="danger">{errors.username.message}</Typography.Text>
            ) : null}

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
            {errors.password ? (
              <Typography.Text type="danger">{errors.password.message}</Typography.Text>
            ) : null}

            <Checkbox
              checked={rememberPassword}
              onChange={(event) => handleRememberPasswordChange(event.target.checked)}
            >
              {t('记住密码')}
            </Checkbox>

            <UiButton
              type="primary"
              htmlType="submit"
              size="large"
              loading={isSubmitting}
              requestLoadingDelayMs={1000}
              className="full-width"
            >
              {t('登录')}
            </UiButton>
          </Space>
        </form>
      </Card>
    </div>
  )
}
