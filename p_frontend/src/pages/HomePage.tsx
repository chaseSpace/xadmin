import { Card, Col, Divider, Row, Space, Tag, Typography } from 'antd'
import { motion } from 'framer-motion'
import { useI18n } from '../i18n/messages'

const container = {
  hidden: { opacity: 0, y: 18 },
  show: {
    opacity: 1,
    y: 0,
    transition: {
      duration: 0.32,
      staggerChildren: 0.08,
    },
  },
}

const item = {
  hidden: { opacity: 0, y: 14 },
  show: { opacity: 1, y: 0, transition: { duration: 0.24 } },
}

const frontendStack = ['React 19', 'TypeScript 5', 'TanStack Router', 'TanStack Query', 'Ant Design 5', 'Vite', 'pnpm']
const backendStack = ['Go', 'Fiber', 'GORM', 'PostgreSQL', 'Protocol Buffers', 'JWT', 'Cron']
const engineeringFeatures = [
  '统一鉴权与会话管理',
  'API级权限校验（RBAC + 永久缓存）',
  '菜单/角色权限模型',
  '组织与岗位管理能力',
  '操作审计与安全处置链路',
  'IP黑名单（内存匹配 + 命中持久化）',
  '告警配置（机器人/场景/模板 + TG发送）',
  '前后端类型与接口契约同步',
]

export function HomePage() {
  const { t } = useI18n()

  return (
    <motion.section className="hello-page" variants={container} initial="hidden" animate="show">
      <motion.div variants={item}>
        <Space direction="vertical" size={12}>
          <Tag color="blue">{t('技术概览')}</Tag>
          <Typography.Title level={2} className="hello-title">
            {t('XAdmin 全栈技术栈与能力全景')}
          </Typography.Title>
          <Typography.Paragraph className="hello-subtitle">
            {t('当前概览页聚焦项目的前后端技术选型、工程规范与核心功能特性，便于快速理解系统架构。')}
          </Typography.Paragraph>
        </Space>
      </motion.div>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <motion.div variants={item}>
            <Card title={t('前端技术栈')} className="hello-card">
              <Space wrap>
                {frontendStack.map((tech) => (
                  <Tag key={tech}>{tech}</Tag>
                ))}
              </Space>
            </Card>
          </motion.div>
        </Col>
        <Col xs={24} lg={12}>
          <motion.div variants={item}>
            <Card title={t('后端技术栈')} className="hello-card">
              <Space wrap>
                {backendStack.map((tech) => (
                  <Tag color="geekblue" key={tech}>
                    {tech}
                  </Tag>
                ))}
              </Space>
            </Card>
          </motion.div>
        </Col>
      </Row>

      <motion.div variants={item}>
        <Card title={t('核心功能特性')} className="hello-card">
          <Row gutter={[16, 12]}>
            {engineeringFeatures.map((feature, index) => (
              <Col xs={24} md={12} xl={8} key={feature}>
                <Typography.Text strong>{`0${index + 1}`}</Typography.Text>
                <Divider style={{ margin: '8px 0' }} />
                <Typography.Paragraph style={{ marginBottom: 0 }}>{t(feature)}</Typography.Paragraph>
              </Col>
            ))}
          </Row>
        </Card>
      </motion.div>
    </motion.section>
  )
}
