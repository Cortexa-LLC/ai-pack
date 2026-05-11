import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { execSync } from 'child_process'

const gitVersion = (() => {
  try { return execSync('git describe --tags --always --dirty').toString().trim() }
  catch { return 'dev' }
})()

// Agent server base URL — override via VITE_API_BASE_URL env var.
// The compiled-in default matches the server's own default port (SERVER_PORT).
const apiBase = process.env.VITE_API_BASE_URL ?? 'http://localhost:8082'

export default defineConfig({
  plugins: [react()],
  define: {
    __APP_VERSION__: JSON.stringify(gitVersion),
    __GIT_COMMIT__: JSON.stringify(''),
  },
  server: {
    port: 3000,
    proxy: {
      '/a2a': apiBase,
      '/stream': apiBase,
      '/logs': apiBase,
      '/metrics': apiBase,
      '/health': apiBase,
      '/graphql': apiBase,
      '/playground': apiBase,
      '/api': apiBase,
    }
  },
  preview: {
    port: 3000,
    proxy: {
      '/a2a': apiBase,
      '/stream': apiBase,
      '/logs': apiBase,
      '/metrics': apiBase,
      '/health': apiBase,
      '/graphql': apiBase,
      '/playground': apiBase,
      '/api': apiBase,
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
