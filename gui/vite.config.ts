import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { execSync } from 'child_process'

const gitVersion = (() => {
  try { return execSync('git describe --tags --always --dirty').toString().trim() }
  catch { return 'dev' }
})()

export default defineConfig({
  plugins: [react()],
  define: {
    __APP_VERSION__: JSON.stringify(gitVersion),
    __GIT_COMMIT__: JSON.stringify(''),
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
  build: {
    // mermaid + its deps (cytoscape, katex) form a single lazy chunk ~750 kB gzip.
    // It's only fetched when a ```mermaid block appears in a message, so the
    // warning is a false positive for initial load. Raise the limit to suppress it.
    chunkSizeWarningLimit: 3000,
    rollupOptions: {
      output: {
        manualChunks: (id) => {
          // Heavy diagram / visualisation libraries — load separately.
          // mermaid pulls in cytoscape + katex as transitive deps; keep them
          // together so the lazy import('mermaid') fetches one coherent chunk.
          if (id.includes('mermaid') || id.includes('cytoscape') || id.includes('katex')) return 'vendor-mermaid';
          // Markdown + syntax highlighting
          if (id.includes('react-markdown') || id.includes('remark') || id.includes('rehype') ||
              id.includes('react-syntax-highlighter')) return 'vendor-markdown';
          // Core React stack
          if (id.includes('node_modules/react/') || id.includes('node_modules/react-dom/') ||
              id.includes('scheduler') || id.includes('@tanstack')) return 'vendor-react';
        },
      },
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test-setup.ts',
  }
})
