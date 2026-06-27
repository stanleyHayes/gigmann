import { defineConfig, devices } from '@playwright/test'

// E2E config (GEC-53/101). Starts the in-memory API + the Vite dev server (which
// proxies /api → :8080), then drives the demo narrative in Chromium. Runs in CI
// (the e2e workflow installs browsers); locally: `npx playwright test`.
export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  retries: process.env.CI ? 1 : 0,
  use: { baseURL: 'http://localhost:5173', trace: 'on-first-retry' },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: [
    {
      command: 'cd ../backend && go run ./cmd/api',
      port: 8080,
      reuseExistingServer: !process.env.CI,
      timeout: 60_000,
      env: { JWT_SECRET: 'e2e-secret', APP_ENV: 'development' },
    },
    {
      command: 'npm run dev',
      port: 5173,
      reuseExistingServer: !process.env.CI,
      timeout: 60_000,
    },
  ],
})
