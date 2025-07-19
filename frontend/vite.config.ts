import { defineConfig } from 'vite'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  build: {
    outDir: 'build',
    minify: 'esbuild',
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8080'
    }
  },
  esbuild: {
    pure: ['console.log'],    // example: have esbuild remove any console.log
    minifyIdentifiers: false, // but keep variable names
  },
  plugins: [
    tailwindcss(),
  ],
})
