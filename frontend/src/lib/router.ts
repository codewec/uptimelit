import { writable } from 'svelte/store'
function normalize(path: string) {
  return path.replace(/\/+$/, '') || '/'
}
export const routePath = writable(normalize(location.pathname))
function navigate(to: string, replace = false) {
  const url = new URL(to, location.origin)
  if (url.origin !== location.origin) {
    location.assign(url)
    return
  }
  history[replace ? 'replaceState' : 'pushState'](null, '', url.pathname + url.search + url.hash)
  routePath.set(normalize(url.pathname))
  window.scrollTo({ top: 0, left: 0 })
}
export const push = (to: string) => navigate(to)
export const replace = (to: string) => navigate(to, true)
export function startRouter() {
  const pop = () => routePath.set(normalize(location.pathname))
  const click = (event: MouseEvent) => {
    if (event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return
    const a = event.target instanceof Element ? event.target.closest<HTMLAnchorElement>('a[href]') : null
    if (!a || a.target || a.download || a.hasAttribute('data-native-navigation')) return
    const url = new URL(a.href, location.href)
    if (url.origin !== location.origin || url.pathname.startsWith('/api/') || url.hash) return
    event.preventDefault()
    push(url.pathname + url.search)
  }
  addEventListener('popstate', pop)
  document.addEventListener('click', click)
  return () => {
    removeEventListener('popstate', pop)
    document.removeEventListener('click', click)
  }
}
