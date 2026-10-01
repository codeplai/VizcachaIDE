import { defineConfig } from 'vitest/config'
import { svelte } from '@sveltejs/vite-plugin-svelte'
// The app version lives in wails.json (info.productVersion), next to this folder.
import wailsConfig from '../wails.json'

export default defineConfig(({ mode }) => ({
  plugins: [svelte()],
  define: { __APP_VERSION__: JSON.stringify(wailsConfig.info.productVersion) },
  // Svelte components must resolve to the browser build in tests.
  resolve: mode === 'test' ? { conditions: ['browser'] } : {},
  test: {
    environment: 'jsdom',
    include: ['src/**/*.test.ts'],
    globals: false
  }
}))
