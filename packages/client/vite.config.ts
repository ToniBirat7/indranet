import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  // Tauri dev server
  server: {
    port: 1420,
    strictPort: true,
  },
  // Output to dist/ (matches tauri.conf.json frontendDist)
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  clearScreen: false,
})
