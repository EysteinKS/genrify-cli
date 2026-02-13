import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  // Base path for deployment (e.g., '/genrify/' for github.io/genrify)
  // Override with BASE_URL env var or keep as '/' for root deployment
  base: process.env.BASE_URL || '/',

  plugins: [react()],

  server: {
    port: 5173,
  },

  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },

  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/__tests__/setup.ts',
  },
})
