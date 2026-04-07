import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  css: {
    preprocessorOptions: {
      less: {
        modifyVars: {
          'primary-color': '#1890ff',
          'border-radius-base': '4px'
        },
        javascriptEnabled: true
      }
    }
  },
  // PostCSS 配置会自动读取 postcss.config.js
  // 这里不需要额外配置
})
