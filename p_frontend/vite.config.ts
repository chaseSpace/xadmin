import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    rolldownOptions: {
      output: {
        codeSplitting: {
          groups: [
            {
              name: 'react-vendor',
              test: /node_modules[\\/](react|react-dom|scheduler)[\\/]/,
              priority: 30,
            },
            {
              name: 'antd-vendor',
              test: /node_modules[\\/](antd|@ant-design|rc-[\w-]+|rc-util)[\\/]/,
              priority: 20,
            },
            {
              name: 'vendor',
              test: /node_modules[\\/]/,
              priority: 10,
              maxSize: 240 * 1024,
            },
          ],
        },
      },
    },
  },
  server: {
    host: '0.0.0.0',
    port: 8633,
    proxy: {
      '/v1': {
        target: 'http://localhost:6666',
        changeOrigin: true,
      },
    },
    hmr: {
      host: 'localhost',
      protocol: 'ws',
      clientPort: 8633,
    },
    watch: {
      usePolling: true,
      interval: 120,
    },
  },
})
