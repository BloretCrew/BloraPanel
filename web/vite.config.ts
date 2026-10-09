import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  // @xterm/headless 6.0.0 publishes its ESM build here but has an incorrect
  // package.module field. Resolve the pinned build without patching packages.
  resolve:{alias:{'@xterm/headless':'@xterm/headless/lib-headless/xterm-headless.mjs'}},
  optimizeDeps:{include:['@xterm/headless','@xterm/addon-serialize']},
  server: { proxy: { '/api': { target: process.env.BLORA_API_TARGET || 'http://127.0.0.1:37861', secure: process.env.BLORA_DEV_SELF_SIGNED !== '1', ws: true } } },
  test: { include: ['tests/**/*.test.ts'] },
  build: { target: 'es2022' },
})
