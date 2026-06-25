/// <reference types="vitest/config" />
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      // 'prompt' (not 'autoUpdate'): never silently reload — an exec may have an
      // approval form open. The ReloadPrompt UI lets them choose when to update.
      registerType: 'prompt',
      devOptions: { enabled: false },
      manifest: {
        name: 'Gigmann Executive Cockpit',
        short_name: 'Ahenfie',
        description: 'AI-native executive chief of staff for the Gigmann healthcare network',
        theme_color: '#0b5cad',
        background_color: '#0f172a',
        display: 'standalone',
        start_url: '/',
        icons: [
          { src: 'pwa-192x192.png', sizes: '192x192', type: 'image/png' },
          { src: 'pwa-512x512.png', sizes: '512x512', type: 'image/png' },
          { src: 'pwa-512x512.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
        ],
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
        navigateFallback: 'index.html',
        // Live data must never come from the SW cache. Navigations to /api or
        // /healthz skip the SPA fallback...
        navigateFallbackDenylist: [/\/api(\/|$)/, /\/healthz$/],
        // ...and every /api + /healthz fetch is forced network-only (no read,
        // no write of cache). This enforces the "never a stale figure" rule.
        runtimeCaching: [
          {
            urlPattern: ({ url }) =>
              url.pathname === '/api' ||
              url.pathname.startsWith('/api/') ||
              url.pathname === '/healthz',
            handler: 'NetworkOnly',
          },
        ],
      },
    }),
  ],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8080',
      '/healthz': 'http://localhost:8080',
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: './src/test/setup.ts',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
      include: ['src/**/*.{ts,tsx}'],
      exclude: [
        'src/**/*.test.{ts,tsx}',
        'src/main.tsx',
        'src/test/**',
        'src/**/*.d.ts',
        'src/api/**',
        'src/app/ReloadPrompt.tsx',
      ],
      thresholds: { statements: 80, lines: 80, functions: 80, branches: 70 },
    },
  },
})
