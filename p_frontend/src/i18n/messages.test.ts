import { describe, expect, it } from 'vitest'
import { translateText } from './messages'

describe('translateText', () => {
  it('returns english translation for mapped chinese source text', () => {
    expect(translateText('系统设置', 'en-US')).toBe('System Settings')
    expect(translateText('岗位管理', 'en-US')).toBe('Positions')
    expect(translateText('IP黑名单', 'en-US')).toBe('IP Blacklist')
  })

  it('keeps chinese source text for zh-CN locale and unknown text fallback', () => {
    expect(translateText('系统设置', 'zh-CN')).toBe('系统设置')
    expect(translateText('未配置的文案', 'en-US')).toBe('未配置的文案')
  })

  it('replaces template params after translation', () => {
    expect(
      translateText('将使当前账号在其他 {count} 处在线会话退出登录。', 'en-US', { count: 3 }),
    ).toBe('This will sign out 3 other active sessions for the current account.')
  })

  it('translates additional business and operation copy used by management pages', () => {
    expect(translateText('业务用户列表', 'en-US')).toBe('Business Users')
    expect(translateText('新增黑名单IP', 'en-US')).toBe('Add Blacklist IP')
    expect(translateText('批量启用', 'en-US')).toBe('Enable Selected')
  })

  it('translates the file timezone hint text', () => {
    expect(
      translateText('根据设置，当前使用的是 {timezone} 时区', 'en-US', {
        timezone: 'Asia/Shanghai',
      }),
    ).toBe('According to settings, the current time zone is Asia/Shanghai.')
  })
})
