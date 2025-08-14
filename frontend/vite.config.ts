import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    // 只在开发环境启用DevTools
    process.env.NODE_ENV === 'development' && vueDevTools(),
  ].filter(Boolean),
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    port: 3000,
    host: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      }
    }
  },
  optimizeDeps: {
    include: ['vue', 'vue-router', 'pinia', 'element-plus']
  },
  build: {
    // 完全移除手动分块，让Vite自动优化
    rollupOptions: {
      output: {
        // 移除手动分块，避免依赖顺序问题
        // manualChunks: undefined
      }
    },
    // 确保生产构建的稳定性
    target: 'es2020', // 提升到es2020以获得更好的兼容性
    minify: 'esbuild',
    sourcemap: false,
    // 增加chunk大小限制，避免过度分割
    chunkSizeWarningLimit: 1000
  }
})
