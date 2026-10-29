<script lang="ts">
  import { onMount } from 'svelte'
  import { Toaster } from 'svelte-french-toast'
  import { api, ApiError } from './lib/api'
  import { user, searchQuery, startTheme, systemInfo } from './lib/stores'
  import { routePath, replace, startRouter } from './lib/router'
  import Login from './pages/Login.svelte'
  import Monitors from './pages/Monitors.svelte'
  import Notifiers from './pages/Notifiers.svelte'
  import Profile from './pages/Profile.svelte'
  import Admin from './pages/Admin.svelte'
  import CustomThemeModal from './components/CustomThemeModal.svelte'
  import AppLayout from './layouts/AppLayout.svelte'
  let loading = $state(true),
    error = $state('')
  onMount(() => {
    const stopRouter = startRouter(),
      stopTheme = startTheme()
    const expired = () => {
      user.set(null)
      searchQuery.set('')
      replace('/login')
    }
    addEventListener('uptimelit:unauthorized', expired)
    api
      .me()
      .then((u) => {
        user.set(u)
      })
      .catch((e) => {
        if (!(e instanceof ApiError && e.status === 401)) error = e.message
      })
      .finally(() => (loading = false))
    return () => {
      stopRouter()
      stopTheme()
      removeEventListener('uptimelit:unauthorized', expired)
    }
  })
  $effect(() => {
    if (loading || error) return
    if (!$user && $routePath !== '/login') replace('/login' + location.search)
    if ($user && $routePath === '/login') replace('/')
    if ($user && $routePath === '/admin' && $user.role !== 'admin') replace('/')
  })
  $effect(() => {
    if ($user)
      api
        .status()
        .then(systemInfo.set)
        .catch(() => {})
  })
  $effect(() => {
    const path = $routePath
    searchQuery.set('')
    document.title =
      ({ '/': 'Monitors', '/notifiers': 'Notifiers', '/admin': 'Administration', '/profile': 'Profile', '/login': 'Sign in' }[path] ??
        'Page not found') + ' · UptimeLit'
  })
</script>

{#if loading}<div class="flex min-h-dvh items-center justify-center text-sm text-text-muted">Loading workspace…</div>
{:else if error}<div class="mx-auto mt-24 max-w-sm rounded-card border border-border bg-surface p-6 text-center">
    <h1 class="text-lg font-semibold">Unable to connect</h1>
    <p role="alert" class="my-4 text-sm text-danger">{error}</p>
    <button class="text-sm text-primary" onclick={() => location.reload()}>Try again</button>
  </div>
{:else if !$user}<Login />
{:else}{#key $user.id}
    {#if $routePath === '/'}<Monitors />{:else if $routePath === '/notifiers'}<Notifiers />{:else if $routePath === '/profile'}<Profile
      />{:else if $routePath === '/admin' && $user.role === 'admin'}<Admin />{:else}<AppLayout
        ><div class="py-20 text-center">
          <h1 class="text-xl font-semibold">Page not found</h1>
          <a href="/" class="mt-4 inline-block text-sm text-primary">Back to monitors</a>
        </div></AppLayout
      >{/if}
  {/key}{/if}
<CustomThemeModal />
<Toaster
  position="bottom-right"
  toastOptions={{
    style:
      'background: var(--color-surface); color: var(--color-text); border: 1px solid var(--color-border); border-radius: 0.75rem; font-size: 13px;',
  }}
/>
