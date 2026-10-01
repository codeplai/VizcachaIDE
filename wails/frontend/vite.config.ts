import { defineConfig } from 'vitest/config'
import { svelte } from '@sveltejs/vite-plugin-svelte'

export default defineConfig(({ mode }) => ({
  plugins: [svelte()],
  // Svelte components must resolve to the browser build in tests.
  resolve: mode === 'test' ? { conditions: ['browser'] } : {},
  test: {
    environment: 'jsdom',
    include: ['src/**/*.test.ts'],
    globals: false
  }
}))
