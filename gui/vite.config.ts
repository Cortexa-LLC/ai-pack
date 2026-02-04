import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      '/a2a': 'http://localhost:8080',
      '/stream': 'http://localhost:8080',
      '/logs': 'http://localhost:8080',
      '/metrics': 'http://localhost:8080',
      '/health': 'http://localhost:8080',
      '/graphql': 'http://localhost:8080',
      '/playground': 'http://localhost:8080',
      '/api': 'http://localhost:8080',
    }
  },
  preview: {
    port: 3000,
    proxy: {
      '/a2a': 'http://localhost:8080',
      '/stream': 'http://localhost:8080',
      '/logs': 'http://localhost:8080',
      '/metrics': 'http://localhost:8080',
      '/health': 'http://localhost:8080',
      '/graphql': 'http://localhost:8080',
      '/playground': 'http://localhost:8080',
      '/api': 'http://localhost:8080',
    }
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test-setup.ts',
  }
})
