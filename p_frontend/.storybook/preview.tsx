import type { Preview } from '@storybook/react-vite'
import { ConfigProvider } from 'antd'
import 'antd/dist/reset.css'
import '../src/styles/tokens.css'
import '../src/styles/global.css'
import { appTheme } from '../src/styles/theme'

const preview: Preview = {
  decorators: [
    (Story) => (
      <ConfigProvider theme={appTheme}>
        <div style={{ padding: 24 }}>
          <Story />
        </div>
      </ConfigProvider>
    ),
  ],
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
  },
}

export default preview
