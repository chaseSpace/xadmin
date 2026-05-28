import { theme } from 'antd'
import type { ThemeConfig } from 'antd'
import type { ThemeMode } from '../store/theme'

export function getAppTheme(mode: ThemeMode): ThemeConfig {
  return {
    algorithm: mode === 'dark' ? theme.darkAlgorithm : theme.defaultAlgorithm,
    token: {
      colorPrimary: '#0f6dff',
      borderRadius: 12,
      fontFamily: 'Avenir Next, Segoe UI, PingFang SC, Microsoft YaHei, sans-serif',
    },
    components: {
      Table: {
        cellPaddingBlock: 10,
        cellPaddingInline: 12,
        cellPaddingBlockMD: 8,
        cellPaddingInlineMD: 10,
        cellPaddingBlockSM: 6,
        cellPaddingInlineSM: 8,
        fontSize: 13,
        fontSizeSM: 12,
      },
    },
  }
}
