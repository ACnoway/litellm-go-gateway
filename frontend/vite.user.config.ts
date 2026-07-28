import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  root: './user',
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './user/src'),
    },
  },
  build: {
    outDir: '../dist/user',
    emptyOutDir: true
  },
  server: {
    port: 5174,
    proxy: {
      '/v1': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
