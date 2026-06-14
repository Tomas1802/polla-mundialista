import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// base: './' keeps asset paths relative, so the built site works under any
// GitHub Pages subpath (https://user.github.io/<repo>/) without hardcoding the
// repository name.
export default defineConfig({
  base: './',
  plugins: [react()],
  server: {
    port: 5173,
  },
})
