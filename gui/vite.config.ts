import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { execSync } from 'child_process'

const gitCommit = (() => {
  try { return execSync('git rev-parse --short HEAD').toString().trim() }
  catch { return 'unknown' }
})()

export default defineConfig({
  plugins: [react()],
  define: {
    __APP_VERSION__: JSON.stringify(process.env.npm_package_version || '1.0.0'),
    __GIT_COMMIT__: JSON.stringify(gitCommit),
  },
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
