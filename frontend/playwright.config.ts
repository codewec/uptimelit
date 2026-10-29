import { defineConfig } from '@playwright/test'
export default defineConfig({
  testDir: './tests',
  workers: 1,
  fullyParallel: false,
  timeout: 60000,
  use: { baseURL: 'http://127.0.0.1:18093', headless: true, screenshot: 'only-on-failure', trace: 'retain-on-failure' },
  webServer: {
    command: 'node tests/server.mjs',
    url: 'http://127.0.0.1:18093/api/auth/config',
    reuseExistingServer: false,
    timeout: 30000,
  },
  reporter: 'list',
})
