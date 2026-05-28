import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useRouterState } from '@tanstack/react-router'
import { QuestionCircleOutlined } from '@ant-design/icons'
import { Card, Form, Input, InputNumber, Modal, Select, Slider, Space, Switch, Tooltip, Typography, message } from 'antd'
import { useEffect, useMemo, useRef } from 'react'
import { UiButton } from '../../components/ui'
import { useI18n } from '../../i18n/messages'
import { useUiSettingsStore } from '../../store/uiSettings'
import { getSystemSettings, updateSystemSettings, type SystemSettings } from '../../services/api/account'
import { systemSettingsKeys } from '../../services/api/queryKeys'

const initialValues = {
  siteName: 'XAdmin 管理后台',
  timezone: 'Asia/Shanghai',
  loginLockThreshold: 5,
  passwordMinLength: 8,
  sessionTimeout: 30,
  passwordPolicy: ['uppercase', 'number'],
  globalWatermarkEnabled: false,
  globalWatermarkFontSize: 16,
}

const COMMON_TIMEZONE_OPTIONS = [
  { value: 'Asia/Shanghai', label: 'Asia/Shanghai (UTC+08:00)' },
  { value: 'Asia/Tokyo', label: 'Asia/Tokyo (UTC+09:00)' },
  { value: 'Asia/Singapore', label: 'Asia/Singapore (UTC+08:00)' },
  { value: 'Asia/Dubai', label: 'Asia/Dubai (UTC+04:00)' },
  { value: 'Europe/London', label: 'Europe/London (UTC+00:00 / +01:00)' },
  { value: 'Europe/Paris', label: 'Europe/Paris (UTC+01:00 / +02:00)' },
  { value: 'America/New_York', label: 'America/New_York (UTC-05:00 / -04:00)' },
  { value: 'America/Chicago', label: 'America/Chicago (UTC-06:00 / -05:00)' },
  { value: 'America/Los_Angeles', label: 'America/Los_Angeles (UTC-08:00 / -07:00)' },
  { value: 'Australia/Sydney', label: 'Australia/Sydney (UTC+10:00 / +11:00)' },
]

function normalizeValues(values: typeof initialValues): typeof initialValues {
  return {
    ...values,
    passwordPolicy: [...(values.passwordPolicy ?? [])],
    globalWatermarkEnabled: Boolean(values.globalWatermarkEnabled),
    globalWatermarkFontSize: Number(values.globalWatermarkFontSize) || 16,
    sessionTimeout: Number(values.sessionTimeout) || 30,
  }
}

function mapSystemSettingsToFormValues(settings: SystemSettings): typeof initialValues {
  return normalizeValues({
    siteName: settings.siteName,
    timezone: settings.timezone,
    loginLockThreshold: settings.loginLockThreshold,
    passwordMinLength: settings.passwordMinLength,
    sessionTimeout: settings.sessionTimeout,
    passwordPolicy: settings.passwordPolicy,
    globalWatermarkEnabled: settings.globalWatermarkEnabled,
    globalWatermarkFontSize: settings.globalWatermarkFontSize,
  })
}

function mapFormValuesToSystemSettings(values: typeof initialValues): SystemSettings {
  return {
    siteName: values.siteName,
    locale: 'zh-CN',
    timezone: values.timezone,
    serverTimezone: 'Local',
    loginLockThreshold: values.loginLockThreshold,
    passwordMinLength: values.passwordMinLength,
    sessionTimeout: values.sessionTimeout,
    passwordPolicy: [...(values.passwordPolicy ?? [])],
    globalWatermarkEnabled: Boolean(values.globalWatermarkEnabled),
    globalWatermarkFontSize: Number(values.globalWatermarkFontSize) || 16,
  }
}

export function SystemSettingsPage() {
  const [form] = Form.useForm()
  const { t } = useI18n()
  const queryClient = useQueryClient()
  const [messageApi, contextHolder] = message.useMessage()
  const [modalApi, modalContextHolder] = Modal.useModal()
  const pathname = useRouterState({ select: (state) => state.location.pathname })
  const prevIsActivePageRef = useRef(false)
  const systemLocaleRef = useRef('zh-CN')
  const watermarkEnabled = useUiSettingsStore((state) => state.globalWatermarkEnabled)
  const watermarkFontSize = useUiSettingsStore((state) => state.globalWatermarkFontSize)
  const setGlobalWatermarkEnabled = useUiSettingsStore((state) => state.setGlobalWatermarkEnabled)
  const setGlobalWatermarkFontSize = useUiSettingsStore((state) => state.setGlobalWatermarkFontSize)
  const setSystemTimezone = useUiSettingsStore((state) => state.setSystemTimezone)
  const setServerTimezone = useUiSettingsStore((state) => state.setServerTimezone)
  const formInitialValues = useMemo(
    () =>
      normalizeValues({
        ...initialValues,
        globalWatermarkEnabled: watermarkEnabled,
        globalWatermarkFontSize: watermarkFontSize,
      }),
    [watermarkEnabled, watermarkFontSize],
  )
  const settingsQuery = useQuery({
    queryKey: systemSettingsKeys.detail,
    queryFn: getSystemSettings,
    enabled: false,
    staleTime: 60_000,
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
  })
  const { data: systemSettings, refetch: refetchSystemSettings } = settingsQuery
  const isActivePage = pathname.startsWith('/system/settings')
  const savedValues = useMemo(
    () => (systemSettings ? mapSystemSettingsToFormValues(systemSettings) : formInitialValues),
    [formInitialValues, systemSettings],
  )
  const watchedValues = Form.useWatch((values: Partial<typeof initialValues>) => values, form)
  const currentValues = useMemo(
    () => normalizeValues({ ...savedValues, ...(watchedValues ?? {}) }),
    [savedValues, watchedValues],
  )
  const hasPendingChanges = useMemo(
    () => JSON.stringify(currentValues) !== JSON.stringify(savedValues),
    [currentValues, savedValues],
  )

  useEffect(() => {
    form.setFieldsValue({
      globalWatermarkEnabled: watermarkEnabled,
      globalWatermarkFontSize: watermarkFontSize,
    })
  }, [form, watermarkEnabled, watermarkFontSize])

  useEffect(() => {
    if (!systemSettings) {
      return
    }
    systemLocaleRef.current = systemSettings.locale || 'zh-CN'
    const values = mapSystemSettingsToFormValues(systemSettings)
    form.setFieldsValue(values)
    setGlobalWatermarkEnabled(values.globalWatermarkEnabled)
    setGlobalWatermarkFontSize(values.globalWatermarkFontSize)
    setSystemTimezone(values.timezone)
    setServerTimezone(systemSettings.serverTimezone)
  }, [form, setGlobalWatermarkEnabled, setGlobalWatermarkFontSize, setServerTimezone, setSystemTimezone, systemSettings])

  useEffect(() => {
    if (isActivePage && !prevIsActivePageRef.current) {
      void refetchSystemSettings()
    }
    prevIsActivePageRef.current = isActivePage
  }, [isActivePage, refetchSystemSettings])

  return (
    <Space direction="vertical" size={16} className="full-width">
      {contextHolder}
      {modalContextHolder}

      <Form
        form={form}
        layout="vertical"
        initialValues={formInitialValues}
      >
        <Space direction="vertical" size={20} className="full-width">
          <Card title={t('基础设置')} className="system-settings-card system-settings-card-base">
            <Form.Item label={t('站点名称')} name="siteName" rules={[{ required: true, message: t('请输入站点名称') }]}>
              <Input placeholder={t('请输入站点名称')} maxLength={20} showCount />
            </Form.Item>
            <Space wrap className="full-width">
              <Form.Item
                label={
                  <Space size={4}>
                    <span>{t('前端展示时区')}</span>
                    <Tooltip title={t('将会应用到所有页面中涉及时间字段的时区适配（仅前端修改）')}>
                      <QuestionCircleOutlined />
                    </Tooltip>
                  </Space>
                }
                name="timezone"
                style={{ minWidth: 320 }}
              >
                <Select
                  showSearch
                  optionFilterProp="label"
                  options={COMMON_TIMEZONE_OPTIONS}
                />
              </Form.Item>
              <Form.Item label={t('服务器时区')} style={{ minWidth: 220 }}>
                <div className="system-settings-readonly-field">
                  <Typography.Text type="secondary">{systemSettings?.serverTimezone || 'Asia/Shanghai'}</Typography.Text>
                </div>
              </Form.Item>
            </Space>
          </Card>

          <Card title={t('安全设置')} className="system-settings-card system-settings-card-security">
            <Space wrap className="full-width">
              <Form.Item label={t('登录失败锁定阈值')} name="loginLockThreshold">
                <InputNumber min={3} max={20} />
              </Form.Item>
              <Form.Item label={t('密码最小长度')} name="passwordMinLength">
                <InputNumber min={6} max={32} />
              </Form.Item>
              <Form.Item label={t('会话过期时间（分钟）')} name="sessionTimeout">
                <InputNumber min={5} max={1440} />
              </Form.Item>
            </Space>
            <Form.Item label={t('密码复杂度策略')} name="passwordPolicy">
              <Select
                mode="multiple"
                placeholder={t('请选择策略')}
                options={[
                  { value: 'uppercase', label: t('大写字母') },
                  { value: 'number', label: t('数字') },
                  { value: 'symbol', label: t('特殊字符') },
                ]}
              />
            </Form.Item>
            <Form.Item label={t('全局页面水印')}>
              <Space size={16} wrap>
                <Form.Item name="globalWatermarkEnabled" valuePropName="checked" noStyle>
                  <Switch />
                </Form.Item>
                <Typography.Text type="secondary">{t('字体大小')}</Typography.Text>
                <Form.Item name="globalWatermarkFontSize" noStyle>
                  <Slider
                    min={12}
                    max={32}
                    step={1}
                    style={{ width: 220, margin: 0 }}
                  />
                </Form.Item>
                <Typography.Text>{watermarkFontSize}px</Typography.Text>
              </Space>
            </Form.Item>
          </Card>

        </Space>
      </Form>

      <Space>
        <UiButton
          type={hasPendingChanges ? 'primary' : 'default'}
          disabled={!hasPendingChanges}
          onClick={() => {
            void form
              .validateFields()
              .then(async (values) => {
                const normalized = normalizeValues(values as typeof initialValues)
                const saved = await updateSystemSettings({
                  ...mapFormValuesToSystemSettings(normalized),
                  serverTimezone: systemSettings?.serverTimezone || 'Local',
                  locale: systemLocaleRef.current,
                })
                systemLocaleRef.current = saved.locale || systemLocaleRef.current
                const current = mapSystemSettingsToFormValues(saved)
                queryClient.setQueryData(systemSettingsKeys.detail, saved)
                form.setFieldsValue(current)
                setGlobalWatermarkEnabled(current.globalWatermarkEnabled)
                setGlobalWatermarkFontSize(current.globalWatermarkFontSize)
                setSystemTimezone(current.timezone)
                setServerTimezone(saved.serverTimezone)
                void messageApi.success(t('配置保存成功'))
              })
              .catch(() => {
                void messageApi.error(t('请先修正表单校验项'))
              })
          }}
        >
          {t('保存配置')}
        </UiButton>
        <UiButton
          onClick={() => {
            void modalApi.confirm({
              title: t('确认恢复默认配置？'),
              onOk: async () => {
                const current = normalizeValues(initialValues)
                const saved = await updateSystemSettings({
                  ...mapFormValuesToSystemSettings(current),
                  serverTimezone: systemSettings?.serverTimezone || 'Local',
                  locale: systemLocaleRef.current,
                })
                systemLocaleRef.current = saved.locale || systemLocaleRef.current
                const nextValues = mapSystemSettingsToFormValues(saved)
                queryClient.setQueryData(systemSettingsKeys.detail, saved)
                form.setFieldsValue(nextValues)
                setGlobalWatermarkEnabled(nextValues.globalWatermarkEnabled)
                setGlobalWatermarkFontSize(nextValues.globalWatermarkFontSize)
                setSystemTimezone(nextValues.timezone)
                setServerTimezone(saved.serverTimezone)
                await messageApi.success(t('已恢复默认配置'))
              },
            })
          }}
        >
          {t('恢复默认')}
        </UiButton>
      </Space>
    </Space>
  )
}
