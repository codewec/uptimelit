import { test, expect, type Page } from '@playwright/test'
import type { Monitor } from '../src/lib/api'
const credentials = { username: 'e2e-admin', password: 'test-admin-password' }
async function login(page: Page, username = credentials.username, password = credentials.password) {
  await page.goto('/login')
  await page.getByLabel(/^Username/).fill(username)
  await page.getByLabel(/^Password/).fill(password)
  await page.getByRole('button', { name: /^(Sign in|Create administrator)$/ }).click()
  await expect(page.getByRole('group', { name: 'Filter monitors by status' })).toBeVisible()
}
async function ensureAdmin(page: Page) {
  const cfg = await (await page.request.get('/api/auth/config')).json()
  if (cfg.bootstrap) {
    const response = await page.request.post('/api/auth/login', { data: credentials, headers: { 'X-UptimeLit-Request': '1' } })
    expect(response.ok()).toBeTruthy()
  }
}

test('bootstrap, monitor and notifier CRUD, user isolation and profile sessions', async ({ page, browser }) => {
  const errors: string[] = []
  page.on('pageerror', (e) => errors.push(e.message))
  await page.goto('/login')
  await expect(page.getByText('Your first sign-in creates the administrator account.', { exact: false })).toBeVisible()
  await login(page)
  await expect(page.getByRole('heading', { name: 'Your first monitor' })).toBeVisible()
  // Choose and persist the same theme presets as DashLit.
  await page.getByRole('button', { name: 'Choose theme' }).click()
  await page.getByRole('menuitem', { name: 'Mocha', exact: true }).click()
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'mocha')
  await page.reload()
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'mocha')
  // Create a notifier without sending anything to an external provider.
  await page.goto('/notifiers')
  await page.getByRole('button', { name: 'New notifier', exact: true }).click()
  let dialog = page.getByRole('dialog', { name: 'New notifier', exact: true })
  await dialog.getByLabel('Name', { exact: true }).fill('API alerts')
  await dialog.getByLabel('Bot token').fill('test-only-token')
  await dialog.getByLabel('Chat ID').fill('12345')
  await dialog.getByRole('button', { name: 'Create notifier', exact: true }).click()
  await expect(page.getByRole('heading', { name: 'API alerts', exact: true })).toBeVisible()
  await expect(dialog).toBeHidden()
  await page.getByRole('button', { name: 'Edit', exact: true }).click()
  dialog = page.getByRole('dialog', { name: 'Edit notifier' })
  await dialog.getByLabel('Chat ID').fill('54321')
  await dialog.getByRole('button', { name: 'Save changes' }).click()
  await expect(dialog).toBeHidden()
  await page.route('**/api/notifiers/*/test', (route) => route.fulfill({ json: { message: 'test sent successfully' } }))
  await page.getByRole('button', { name: 'Send test' }).click()
  await expect(page.getByText('Test notification sent', { exact: true })).toBeVisible()
  // Create an HTTP monitor and link the owner’s notifier.
  await page.goto('/')
  await page.getByRole('button', { name: 'New monitor', exact: true }).click()
  dialog = page.getByRole('dialog', { name: 'New monitor', exact: true })
  await dialog.getByLabel('Name', { exact: true }).fill('Production API')
  await dialog.getByLabel('URL', { exact: true }).fill('http://127.0.0.1:18093/api/auth/config')
  await dialog.getByRole('checkbox', { name: /API alerts/ }).check()
  await dialog.getByRole('button', { name: 'Create monitor', exact: true }).click()
  await expect(page.getByRole('heading', { name: 'Production API', exact: true })).toBeVisible()
  await expect(dialog).toBeHidden()
  await expect(page.getByText('1 notifier', { exact: true })).toBeVisible()
  await expect(page.locator('section').getByText('Up', { exact: true })).toBeVisible({ timeout: 15000 })
  // Check details use a tooltip without opening the full history dialog.
  const timeline = page.getByRole('group', { name: 'Check history for Production API', exact: true })
  const latestBar = timeline.locator('[data-check-key]').last()
  await latestBar.hover()
  const tooltip = page.getByRole('tooltip')
  await expect(tooltip).toBeVisible()
  await expect(tooltip.getByText('HTTP status', { exact: true })).toBeVisible()
  await expect(tooltip.getByText('200', { exact: true })).toBeVisible()
  await expect(page.getByRole('dialog')).toHaveCount(0)
  await page.keyboard.press('Escape')
  await expect(tooltip).toBeHidden()
  await page.mouse.move(0, 0)
  await latestBar.focus()
  await expect(tooltip).toBeVisible()
  await expect(latestBar).toHaveAttribute('aria-describedby', (await tooltip.getAttribute('id')) as string)
  await page.keyboard.press('Tab')
  await expect(timeline.getByRole('button', { name: 'View history for Production API' })).toBeFocused()
  await expect(tooltip).toBeHidden()
  await page.screenshot({ path: 'test-results/desktop-monitors.png', fullPage: true })
  // Edit, search, pause and inspect history.
  await page.getByRole('button', { name: 'Actions for Production API' }).click()
  await page.getByRole('menuitem', { name: 'Edit', exact: true }).click()
  dialog = page.getByRole('dialog', { name: 'Edit monitor' })
  await dialog.getByLabel('Name', { exact: true }).fill('Core API')
  await dialog.getByRole('button', { name: 'Save changes' }).click()
  await expect(page.getByRole('heading', { name: 'Core API', exact: true })).toBeVisible()
  await page.getByRole('textbox', { name: 'Search', exact: true }).fill('does not exist')
  await expect(page.getByRole('heading', { name: 'No matching monitors' })).toBeVisible()
  await page.getByRole('button', { name: 'Clear filters' }).click()
  await expect(page.getByRole('combobox', { name: 'Filter monitors', exact: true })).toHaveCount(0)
  await page.getByRole('button', { name: 'Down monitors', exact: true }).click()
  await expect(page.getByRole('button', { name: 'Down monitors', exact: true })).toHaveAttribute('aria-pressed', 'true')
  await expect(page.getByRole('heading', { name: 'No matching monitors' })).toBeVisible()
  await page.getByRole('textbox', { name: 'Search', exact: true }).fill('another search')
  await page.getByRole('button', { name: 'All monitors', exact: true }).click()
  await expect(page.getByRole('textbox', { name: 'Search', exact: true })).toHaveValue('')
  await expect(page.getByRole('button', { name: 'All monitors', exact: true })).toHaveAttribute('aria-pressed', 'true')
  await expect(page.getByRole('heading', { name: 'Core API', exact: true })).toBeVisible()

  await page.getByRole('button', { name: 'Actions for Core API' }).click()
  await page.getByRole('menuitem', { name: 'Pause', exact: true }).click()
  await expect(page.locator('section').getByText('Paused', { exact: true })).toBeVisible()
  await page.getByRole('button', { name: 'Paused monitors', exact: true }).click()
  await expect(page.getByRole('button', { name: 'Paused monitors', exact: true })).toHaveAttribute('aria-pressed', 'true')
  await expect(page.getByRole('heading', { name: 'Core API', exact: true })).toBeVisible()
  await page.getByRole('button', { name: 'Up monitors', exact: true }).click()
  await expect(page.getByRole('heading', { name: 'No matching monitors' })).toBeVisible()
  await page.getByRole('button', { name: 'All monitors', exact: true }).click()

  await page.getByRole('button', { name: 'View history for Core API' }).click()
  dialog = page.getByRole('dialog', { name: 'Core API', exact: true })
  await expect(dialog).toBeVisible()
  await expect(dialog.getByRole('cell', { name: '200', exact: true }).first()).toBeVisible()
  await dialog.getByRole('button', { name: 'Close', exact: true }).click()
  // Provision an ordinary user from the administrative page.
  await page.goto('/admin')
  await page.getByRole('button', { name: 'Create user', exact: true }).click()
  dialog = page.getByRole('dialog', { name: 'Create user', exact: true })
  await dialog.getByLabel('Username', { exact: true }).fill('e2e-alice')
  await dialog.getByLabel(/^Password/).fill('test-alice-password')
  await dialog.getByRole('button', { name: 'Create user', exact: true }).click()
  await expect(page.getByRole('cell', { name: 'e2e-alice', exact: true })).toBeVisible()
  const aliceContext = await browser.newContext({ baseURL: 'http://127.0.0.1:18093' })
  const alice = await aliceContext.newPage()
  await login(alice, 'e2e-alice', 'test-alice-password')
  await expect(alice.getByRole('heading', { name: 'Your first monitor' })).toBeVisible()
  await expect(alice.getByRole('link', { name: 'Administration', exact: true })).toHaveCount(0)
  await alice.goto('/admin')
  await expect(alice.getByRole('group', { name: 'Filter monitors by status' })).toBeVisible()
  await alice.goto('/profile')
  await alice.getByLabel('Current password', { exact: true }).fill('test-alice-password')
  await alice.getByLabel(/^New password/).fill('test-alice-updated')
  await alice.getByLabel('Confirm new password', { exact: true }).fill('test-alice-updated')
  await alice.getByRole('button', { name: 'Update password' }).click()
  await expect(alice.getByRole('button', { name: 'Sign in', exact: true })).toBeVisible()
  await login(alice, 'e2e-alice', 'test-alice-updated')
  await aliceContext.close()
  // Delete through accessible confirmation dialogs.
  await page.goto('/')
  await page.getByRole('button', { name: 'Actions for Core API' }).click()
  await page.getByRole('menuitem', { name: 'Delete', exact: true }).click()
  await page.getByRole('alertdialog').getByRole('button', { name: 'Delete', exact: true }).click()
  await expect(page.getByRole('heading', { name: 'Your first monitor' })).toBeVisible()
  await page.goto('/notifiers')
  await page.getByRole('button', { name: 'Delete', exact: true }).click()
  await page.getByRole('alertdialog').getByRole('button', { name: 'Delete', exact: true }).click()
  await expect(page.getByRole('heading', { name: 'Stay in the loop' })).toBeVisible()
  await page.goto('/admin')
  await page
    .getByRole('row')
    .filter({ has: page.getByRole('cell', { name: 'e2e-alice', exact: true }) })
    .getByRole('button', { name: 'Delete', exact: true })
    .click()
  await page.getByRole('alertdialog').getByRole('button', { name: 'Delete', exact: true }).click()
  await expect(page.getByRole('cell', { name: 'e2e-alice', exact: true })).toHaveCount(0)
  await page.getByRole('button', { name: 'Sign out', exact: true }).click()
  await expect(page.getByRole('button', { name: 'Sign in', exact: true })).toBeVisible()
  expect(errors).toEqual([])
})

test('mobile navigation, dialogs and responsive layout', async ({ page }) => {
  await ensureAdmin(page)
  await page.setViewportSize({ width: 390, height: 844 })
  await login(page)
  await page.getByRole('button', { name: 'Menu', exact: true }).click()
  await page.getByRole('menuitem', { name: 'Notifiers', exact: true }).click()
  await expect(page.getByRole('heading', { name: 'Notifiers', exact: true })).toBeVisible()
  await page.getByRole('button', { name: 'New notifier', exact: true }).click()
  const dialog = page.getByRole('dialog', { name: 'New notifier', exact: true })
  await expect(dialog).toBeVisible()
  const bounds = await dialog.boundingBox()
  expect(bounds?.width).toBeLessThanOrEqual(390)
  await page.keyboard.press('Escape')
  await expect(dialog).toBeHidden()
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBeTruthy()
  await page.screenshot({ path: 'test-results/mobile-notifiers.png', fullPage: true })
})

test('timeline has continuous hit areas and one shared tooltip during fast pointer movement', async ({ page }) => {
  await ensureAdmin(page)
  const checks = Array.from({ length: 50 }, (_, index) => ({
    checked_at: new Date(Date.UTC(2026, 9, 29, 12, index)).toISOString(),
    is_up: index % 2 === 0,
    status_code: index % 2 === 0 ? 200 : 503,
    response_time_ms: index,
  }))
  const monitor: Monitor = {
    id: 42,
    owner_id: 'fixture',
    name: 'Timeline test',
    type: 'http',
    url: 'https://example.com',
    interval_seconds: 60,
    fail_threshold: 3,
    paused: false,
    created_at: checks[0].checked_at,
    notifier_ids: [],
    status_code: 503,
    response_time_ms: 49,
    is_up: false,
    checked_at: checks[49].checked_at,
    uptime_24h: 50,
    recent_checks: checks,
  }
  await page.route('**/api/endpoints', (route) => route.fulfill({ json: [monitor] }))
  await login(page)
  const timeline = page.getByRole('group', { name: 'Check history for Timeline test', exact: true })
  const bars = timeline.locator('[data-check-key]')
  await expect(bars).toHaveCount(50)
  const openBars = timeline.locator('[data-check-key]:not([data-state="closed"])')
  const tooltip = page.getByRole('tooltip')
  await bars.first().hover()
  await expect(tooltip).toBeVisible()
  await expect(tooltip).toHaveCSS('animation-name', /history-tooltip-in/)
  const tooltipID = await tooltip.getAttribute('id')
  // Cross former gaps and jump between non-neighbouring bars without hover delays.
  for (const index of [1, 32, 4, 49, 0, 25]) {
    const bar = bars.nth(index)
    const bounds = (await bar.boundingBox())!
    // Both padding edges must belong to this bar, not the row or its neighbours.
    for (const x of [bounds.x + 0.2, bounds.x + bounds.width - 0.2]) {
      await page.mouse.move(x, bounds.y + bounds.height / 2)
      await expect(openBars).toHaveCount(1)
      await expect(openBars).toHaveAttribute('data-check-key', checks[index].checked_at)
      await expect(tooltip.locator('time')).toHaveAttribute('datetime', checks[index].checked_at)
      await expect(tooltip).toHaveAttribute('id', tooltipID!)
      await expect(tooltip).toHaveCSS('animation-name', 'none')
      // Only the bar associated with the tooltip may be enlarged.
      await expect
        .poll(() => bars.locator('span').evaluateAll((spans) => spans.filter((span) => getComputedStyle(span).transform !== 'none').length))
        .toBe(1)
    }
  }
  await page.screenshot({ path: 'test-results/timeline-tooltip.png' })
  await tooltip.evaluate((element) => {
    element.addEventListener('animationstart', (event) => {
      document.documentElement.dataset.tooltipAnimation = (event as AnimationEvent).animationName
    })
  })
  await page.mouse.move(0, 0)
  await expect(tooltip).toBeHidden()
  await expect(page.locator('html')).toHaveAttribute('data-tooltip-animation', /history-tooltip-out/)
  await expect(openBars).toHaveCount(0)
  await bars.last().focus()
  await expect(tooltip).toBeVisible()
  await page.keyboard.press('Home')
  await expect(bars.first()).toBeFocused()
  await expect(openBars).toHaveCount(1)
  await expect(tooltip.locator('time')).toHaveAttribute('datetime', checks[0].checked_at)
  await expect(tooltip.getByText('0 ms', { exact: true })).toBeVisible()
  await page.keyboard.press('End')
  await expect(bars.last()).toBeFocused()
  await page.keyboard.press('Escape')
  await expect(tooltip).toBeHidden()
  await expect(openBars).toHaveCount(0)
  // Arrow polygons use currentColor, so their text color must match the surface.
  for (const themeName of ['Crema', 'Mocha']) {
    await page.getByRole('button', { name: 'Choose theme' }).click()
    await page.getByRole('menuitem', { name: themeName, exact: true }).click()
    await bars.first().hover()
    await expect(tooltip).toBeVisible()
    const colors = await tooltip.evaluate((element) => ({
      background: getComputedStyle(element).backgroundColor,
      arrow: getComputedStyle(element.querySelector('polygon')!).fill,
    }))
    expect(colors.arrow).toBe(colors.background)
    await page.mouse.move(0, 0)
    await expect(tooltip).toBeHidden()
  }
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await bars.first().focus()
  await expect(tooltip).toBeVisible()
  await expect(tooltip).toHaveCSS('animation-name', 'none')
})

test('Bits UI v2 menus and dialogs retain keyboard navigation, focus and dismissal', async ({ page }) => {
  const errors: string[] = []
  page.on('pageerror', (error) => errors.push(error.message))
  await ensureAdmin(page)
  await page.context().clearCookies()
  await page.goto('/login')
  // The floating theme menu is also used before signing in.
  const themeButton = page.getByRole('button', { name: 'Choose theme' })
  await themeButton.focus()
  await page.keyboard.press('ArrowDown')
  await expect(page.getByRole('menu')).toBeVisible()
  await page.keyboard.press('End')
  await expect(page.getByRole('menuitem', { name: 'Edit colors…', exact: true })).toBeFocused()
  await page.keyboard.press('Enter')
  const customTheme = page.getByRole('dialog', { name: 'Custom theme', exact: true })
  await expect(customTheme).toBeVisible()
  await expect(page.getByRole('menu')).toHaveCount(0)
  await expect(customTheme.getByLabel('Color scheme')).toBeFocused()
  await customTheme.getByLabel('Color scheme').selectOption('light')
  await page.keyboard.press('Escape')
  await expect(customTheme).toBeHidden()
  await expect(themeButton).toBeFocused()

  await login(page)
  const newMonitor = page.getByRole('button', { name: 'New monitor', exact: true })
  await newMonitor.click()
  const monitorDialog = page.getByRole('dialog', { name: 'New monitor', exact: true })
  await expect(monitorDialog.getByLabel('Name', { exact: true })).toBeFocused()
  await expect(monitorDialog).toHaveAccessibleDescription('Keep an eye on a website, API, TCP service or host.')
  await page.keyboard.press('Shift+Tab')
  await expect(monitorDialog.getByRole('button', { name: 'Create monitor', exact: true })).toBeFocused()
  await page.keyboard.press('Tab')
  await expect(monitorDialog.getByLabel('Name', { exact: true })).toBeFocused()
  await monitorDialog.getByLabel('Monitor type').selectOption('tcp')
  await expect(monitorDialog.getByLabel('Monitor type')).toHaveValue('tcp')
  await page.keyboard.press('Escape')
  await expect(monitorDialog).toBeHidden()
  await expect(newMonitor).toBeFocused()
  // Pointer dismissal still works and scroll locking is released.
  await newMonitor.click()
  await expect(monitorDialog).toBeVisible()
  await page.mouse.click(1, 1)
  await expect(monitorDialog).toBeHidden()
  await expect.poll(() => page.evaluate(() => getComputedStyle(document.body).overflow)).not.toBe('hidden')

  await page.goto('/notifiers')
  const newNotifier = page.getByRole('button', { name: 'New notifier', exact: true })
  await newNotifier.click()
  const notifierDialog = page.getByRole('dialog', { name: 'New notifier', exact: true })
  await notifierDialog.getByLabel('Provider').selectOption('pushover')
  await expect(notifierDialog.getByLabel('Priority')).toBeVisible()
  await notifierDialog.getByLabel('Priority').selectOption('1')
  await notifierDialog.getByRole('button', { name: 'Cancel', exact: true }).click()
  await expect(notifierDialog).toBeHidden()
  await expect(newNotifier).toBeFocused()

  await page.goto('/admin')
  await page.getByRole('button', { name: 'Create user', exact: true }).click()
  const userDialog = page.getByRole('dialog', { name: 'Create user', exact: true })
  await expect(userDialog).toBeVisible()
  await userDialog.getByRole('combobox', { name: 'Role', exact: true }).selectOption('admin')
  await expect(userDialog.getByRole('combobox', { name: 'Role', exact: true })).toHaveValue('admin')
  await page.keyboard.press('Escape')
  await expect(userDialog).toBeHidden()

  // The mobile menu uses the same theme items and opens the shared dialog.
  await page.setViewportSize({ width: 390, height: 844 })
  const mobileMenu = page.getByRole('button', { name: 'Menu', exact: true })
  await mobileMenu.click()
  await page.getByRole('menuitem', { name: 'Edit colors…', exact: true }).click()
  await expect(customTheme).toBeVisible()
  await expect(customTheme.getByLabel('Color scheme')).toBeFocused()
  await page.keyboard.press('Escape')
  await expect(customTheme).toBeHidden()
  await expect(mobileMenu).toBeFocused()
  expect(errors).toEqual([])
})

test('Bits UI v2 alert dialog stays open on failed asynchronous confirmation', async ({ page }) => {
  await ensureAdmin(page)
  await login(page)
  const response = await page.request.post('/api/notifiers', {
    headers: { 'X-UptimeLit-Request': '1' },
    data: { name: 'Confirmation test', type: 'telegram', config_json: JSON.stringify({ bot_token: 'test', chat_id: 'test' }) },
  })
  expect(response.ok()).toBeTruthy()
  const notifier = await response.json()
  try {
    await page.goto('/notifiers')
    await page
      .locator('section')
      .filter({ has: page.getByRole('heading', { name: 'Confirmation test', exact: true }) })
      .getByRole('button', { name: 'Delete', exact: true })
      .click()
    const confirmation = page.getByRole('alertdialog', { name: 'Delete notifier', exact: true })
    await expect(confirmation).toBeVisible()
    await expect(confirmation).toBeFocused()
    await page.mouse.click(1, 1)
    await expect(confirmation).toBeVisible()
    await confirmation.getByRole('button', { name: 'Cancel', exact: true }).click()
    await expect(confirmation).toBeHidden()
    await page
      .locator('section')
      .filter({ has: page.getByRole('heading', { name: 'Confirmation test', exact: true }) })
      .getByRole('button', { name: 'Delete', exact: true })
      .click()
    let release!: () => void
    const gate = new Promise<void>((resolve) => {
      release = resolve
    })
    await page.route(`**/api/notifiers/${notifier.id}`, async (route) => {
      await gate
      await route.fulfill({ status: 500, json: { error: 'Confirmation test failure' } })
    })
    await confirmation.getByRole('button', { name: 'Delete', exact: true }).click()
    try {
      await expect(confirmation).toBeVisible()
      await expect(confirmation.getByRole('button', { name: 'Cancel', exact: true })).toBeDisabled()
      await expect(confirmation.getByRole('button').last()).toBeDisabled()
    } finally {
      release()
    }
    await expect(page.getByText('Confirmation test failure', { exact: true })).toBeVisible()
    await expect(confirmation).toBeVisible()
    await expect(confirmation.getByRole('button', { name: 'Delete', exact: true })).toBeEnabled()
    await page.keyboard.press('Escape')
    await expect(confirmation).toBeHidden()
  } finally {
    await page.request.delete(`/api/notifiers/${notifier.id}`, { headers: { 'X-UptimeLit-Request': '1' } })
  }
})

test('TCP cards and check history show connection failure reasons instead of HTTP codes', async ({ page }) => {
  await ensureAdmin(page)
  const timestamp = '2026-10-30T12:00:00Z'
  const cases = [
    { name: 'Healthy TCP', up: true, error: '', label: 'Connected' },
    { name: 'Refused TCP', up: false, error: 'connection_refused', label: 'Connection refused' },
    { name: 'Slow TCP', up: false, error: 'timeout', label: 'Timeout' },
    { name: 'DNS TCP', up: false, error: 'dns_error', label: 'DNS error' },
    { name: 'Unreachable TCP', up: false, error: 'network_unreachable', label: 'Network unreachable' },
    { name: 'Reset TCP', up: false, error: 'connection_reset', label: 'Connection reset' },
    { name: 'Other TCP', up: false, error: 'connection_error', label: 'Connection error' },
    { name: 'Legacy TCP', up: false, error: '', label: 'Failed' },
    { name: 'Pending TCP', up: null, error: '', label: '—' },
  ]
  const monitors: Monitor[] = cases.map((item, index) => ({
    id: index + 1,
    owner_id: 'fixture',
    name: item.name,
    type: 'tcp',
    url: 'localhost:1234',
    interval_seconds: 60,
    fail_threshold: 3,
    paused: false,
    created_at: timestamp,
    notifier_ids: [],
    status_code: null,
    response_time_ms: item.up == null ? null : 0,
    is_up: item.up,
    checked_at: item.up == null ? null : timestamp,
    uptime_24h: null,
    error_code: item.error,
    recent_checks: item.up == null ? [] : [{ is_up: item.up, checked_at: timestamp, response_time_ms: 0, error_code: item.error }],
  }))
  await page.route('**/api/endpoints', (route) => route.fulfill({ json: monitors }))
  await page.route('**/api/endpoints/2/checks?*', (route) => route.fulfill({ json: monitors[1].recent_checks }))
  await login(page)
  for (const item of cases) {
    const card = page.locator('section').filter({ has: page.getByRole('heading', { name: item.name, exact: true }) })
    await expect(card.getByText('Connection', { exact: true })).toBeVisible()
    await expect(card.getByText('HTTP status', { exact: true })).toHaveCount(0)
    if (item.up != null) await expect(card.getByText(item.label, { exact: true })).toBeVisible()
  }
  const failedCard = page.locator('section').filter({ has: page.getByRole('heading', { name: 'Refused TCP', exact: true }) })
  await failedCard.locator('[data-check-key]').hover()
  const tooltip = page.getByRole('tooltip')
  await expect(tooltip.getByText('Connection refused', { exact: true })).toBeVisible()
  await expect(tooltip.getByText('HTTP status', { exact: true })).toHaveCount(0)
  await page.keyboard.press('Escape')
  await failedCard.getByRole('button', { name: 'View history for Refused TCP' }).click()
  const history = page.getByRole('dialog', { name: 'Refused TCP', exact: true })
  await expect(history.getByRole('columnheader', { name: 'Connection', exact: true })).toBeVisible()
  await expect(history.getByRole('cell', { name: 'Connection refused', exact: true })).toBeVisible()
})

test('Ping monitor checks ICMP and displays replies in the card, tooltip and history', async ({ page }) => {
  await ensureAdmin(page)
  await login(page)
  await page.getByRole('button', { name: 'New monitor', exact: true }).click()
  const form = page.getByRole('dialog', { name: 'New monitor', exact: true })
  await form.getByLabel('Name', { exact: true }).fill('Local ICMP')
  await form.getByLabel('Monitor type').selectOption('ping')
  await form.getByLabel('Host', { exact: true }).fill('127.0.0.1')
  await form.getByRole('button', { name: 'Create monitor', exact: true }).click()
  const card = page.locator('section').filter({ has: page.getByRole('heading', { name: 'Local ICMP', exact: true }) })
  await expect(card.getByText('Reply received', { exact: true })).toBeVisible({ timeout: 15000 })
  await expect(card.getByText('HTTP status', { exact: true })).toHaveCount(0)
  await expect(card.getByText('ICMP', { exact: true })).toBeVisible()
  await card.locator('[data-check-key]').last().hover()
  const tooltip = page.getByRole('tooltip')
  await expect(tooltip.getByText('Reply received', { exact: true })).toBeVisible()
  await expect(tooltip.getByText('ICMP', { exact: true })).toBeVisible()
  await page.keyboard.press('Escape')
  await card.getByRole('button', { name: 'View history for Local ICMP' }).click()
  const history = page.getByRole('dialog', { name: 'Local ICMP', exact: true })
  await expect(history.getByRole('columnheader', { name: 'ICMP', exact: true })).toBeVisible()
  await expect(history.getByRole('cell', { name: 'Reply received', exact: true }).first()).toBeVisible()
  await history.getByRole('button', { name: 'Close', exact: true }).click()
  await card.getByRole('button', { name: 'Actions for Local ICMP' }).click()
  await page.getByRole('menuitem', { name: 'Edit', exact: true }).click()
  const edit = page.getByRole('dialog', { name: 'Edit monitor', exact: true })
  await expect(edit.getByLabel('Monitor type')).toHaveValue('ping')
  await expect(edit.getByLabel('Host', { exact: true })).toHaveValue('127.0.0.1')
  await edit.getByLabel('Host', { exact: true }).fill('localhost')
  await edit.getByRole('button', { name: 'Save changes', exact: true }).click()
  await expect(card.getByText('localhost', { exact: true })).toBeVisible()
  await card.getByRole('button', { name: 'Actions for Local ICMP' }).click()
  await page.getByRole('menuitem', { name: 'Delete', exact: true }).click()
  await page.getByRole('alertdialog').getByRole('button', { name: 'Delete', exact: true }).click()
  await expect(card).toHaveCount(0)
})
