import { derived, writable, get } from 'svelte/store'
import type { User } from './api'
import {
  defaultCustomTheme,
  loadGuestTheme,
  loadGuestCustomTheme,
  guestHasCustomTheme,
  saveGuestTheme,
  normalizeCustomTheme,
  type Theme,
  type CustomThemeConfig,
} from './themes'
export const user = writable<User | null>(null)
export const searchQuery = writable('')
export const systemInfo = writable<{ version: string } | null>(null)
export const theme = writable<Theme>(loadGuestTheme())
export const customTheme = writable<CustomThemeConfig>(loadGuestCustomTheme())
export const hasCustomTheme = writable(guestHasCustomTheme())
export const customThemeEditorOpen = writable(false)
const systemDark = writable(matchMedia('(prefers-color-scheme: dark)').matches)
export const resolvedTheme = derived([theme, systemDark], ([$theme, $dark]) =>
  $theme === 'system' ? ($dark ? 'frappe' : 'crema') : $theme,
)
export const setTheme = (value: Theme) => theme.set(value)
export const openCustomThemeEditor = () => customThemeEditorOpen.set(true)
export function selectCustomTheme() {
  if (get(hasCustomTheme)) theme.set('custom')
  else openCustomThemeEditor()
}
export function setCustomTheme(value: CustomThemeConfig) {
  customTheme.set(normalizeCustomTheme(value))
  hasCustomTheme.set(true)
  theme.set('custom')
}
export function deleteCustomTheme() {
  hasCustomTheme.set(false)
  customTheme.set({ ...defaultCustomTheme })
  theme.set('system')
}
export function startTheme() {
  const media = matchMedia('(prefers-color-scheme: dark)')
  const handle = (event: MediaQueryListEvent) => systemDark.set(event.matches)
  media.addEventListener('change', handle)
  const unsubscribe = derived([theme, resolvedTheme, customTheme, hasCustomTheme], (values) => values).subscribe(
    ([selected, resolved, custom, exists]) => {
      const root = document.documentElement
      root.dataset.theme = resolved
      root.style.colorScheme =
        resolved === 'crema' || resolved === 'latte' || (resolved === 'custom' && custom.scheme === 'light') ? 'light' : 'dark'
      const fields = ['bg', 'surface', 'text', 'primary', 'accent', 'danger', 'success'] as const
      for (const field of fields) {
        if (resolved === 'custom') root.style.setProperty('--color-' + field, custom[field])
        else root.style.removeProperty('--color-' + field)
      }
      root.style.setProperty(
        '--page-bg-image',
        resolved === 'custom' && custom.backgroundImage ? `url(${JSON.stringify(custom.backgroundImage)})` : 'none',
      )
      try {
        saveGuestTheme(selected, custom, exists)
      } catch {
        /* Preferences still work when browser storage is unavailable. */
      }
    },
  )
  return () => {
    media.removeEventListener('change', handle)
    unsubscribe()
  }
}
