import { defineConfig } from 'vite';

export default defineConfig({
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  server: {
    proxy: {
      '/transactions': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});
