import { defineConfig } from 'vite'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  build: {
    outDir: 'build',
    minify: 'esbuild',
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/uploads': 'http://localhost:8080'
    }
  },
  esbuild: {
    pure: ['console.log'],
    minifyIdentifiers: false,
  },
  plugins: [
    tailwindcss(),
  ],
})
