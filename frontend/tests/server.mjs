import { spawn } from 'node:child_process'
import { mkdtempSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
const directory = mkdtempSync(join(tmpdir(), 'uptimelit-e2e-'))
const app = fileURLToPath(new URL('../../app', import.meta.url))
const child = spawn(app, ['-addr', '127.0.0.1:18093', '-db', join(directory, 'test.db')], {
  stdio: 'inherit',
  env: {
    ...process.env,
    OIDC_ISSUER: '',
    OIDC_CLIENT_ID: '',
    OIDC_CLIENT_SECRET: '',
    OIDC_REDIRECT_URL: '',
    DISABLE_PASSWORD_LOGIN: 'false',
    JWT_SECRET: 'uptimelit-browser-tests-only-signing-key',
  },
})
const stop = () => child.kill('SIGTERM')
process.on('SIGTERM', stop)
process.on('SIGINT', stop)
child.on('error', (error) => {
  console.error(error)
  rmSync(directory, { recursive: true, force: true })
  process.exit(1)
})
child.on('exit', (code) => {
  rmSync(directory, { recursive: true, force: true })
  process.exit(code ?? 0)
})
