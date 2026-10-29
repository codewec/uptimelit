<script lang="ts">
  import { onMount } from 'svelte'
  import { api, type AuthConfig } from '../lib/api'
  import { user } from '../lib/stores'
  import { replace } from '../lib/router'
  import { toastError } from '../lib/toasts'
  import AuthLayout from '../layouts/AuthLayout.svelte'
  import ThemeFab from '../components/ThemeFab.svelte'
  import logo from '../assets/uptimelit.svg'
  import { field, label } from '../lib/styles'
  let config = $state<AuthConfig | null>(null),
    loading = $state(true),
    saving = $state(false),
    username = $state(''),
    password = $state(''),
    error = $state('')
  onMount(() => {
    const controller = new AbortController()
    const oidcError = new URLSearchParams(location.search).get('oidc_error')
    if (oidcError) {
      error = oidcError
      replace('/login')
    }
    api
      .authConfig(controller.signal)
      .then((c) => (config = c))
      .catch((e) => {
        if (e.name !== 'AbortError') error = e.message
      })
      .finally(() => (loading = false))
    return () => controller.abort()
  })
  async function submit(event: SubmitEvent) {
    event.preventDefault()
    saving = true
    error = ''
    try {
      const result = await api.login(username.trim(), password)
      password = ''
      user.set(result.user)
      replace('/')
    } catch (e) {
      error = e instanceof Error ? e.message : 'Sign in failed'
      toastError(e)
    } finally {
      saving = false
    }
  }
</script>

<AuthLayout>
  <form class="rounded-card border border-border-soft bg-surface p-6 shadow-xl" onsubmit={submit}>
    <div class="mb-6 text-center">
      <img src={logo} alt="" class="mx-auto mb-3 h-16 w-16 rounded-xl" />
      <h1 class="text-xl font-semibold tracking-tight">UptimeLit</h1>
      <p class="mt-1 text-sm text-text-muted">{config?.bootstrap ? 'Set up your monitoring workspace' : 'Sign in to your workspace'}</p>
    </div>
    {#if error}<p role="alert" class="mb-3 rounded-lg bg-danger-soft px-3 py-2 text-sm text-danger">{error}</p>{/if}
    {#if loading}<p class="py-8 text-center text-sm text-text-muted">Loading…</p>{:else if config}
      {#if config.bootstrap}<p
          class="mb-4 rounded-lg border border-border-soft bg-bg-elevated px-3 py-2.5 text-xs leading-relaxed text-text-muted"
        >
          Your first sign-in creates the administrator account. Additional users are created by an administrator.
        </p>{/if}
      {#if config.passwordLoginEnabled}
        <label class="mb-3 block"
          ><span class={label}>Username</span><input
            class={field}
            bind:value={username}
            autocomplete="username"
            required
            maxlength="64"
            pattern="[a-zA-Z0-9._-]+"
          /></label
        >
        <label class="mb-4 block"
          ><span class={label}>Password</span><input
            class={field}
            type="password"
            bind:value={password}
            autocomplete={config.bootstrap ? 'new-password' : 'current-password'}
            required
            minlength={config.bootstrap ? 8 : undefined}
          />{#if config.bootstrap}<span class="mt-1 block text-[11px] text-text-subtle"
              >8–72 bytes. Usernames use letters, digits, dots, underscores or dashes.</span
            >{/if}</label
        >
        <button
          type="submit"
          disabled={saving}
          class="w-full rounded-btn bg-primary py-2.5 text-sm font-medium text-primary-fg hover:bg-primary-hover disabled:opacity-60"
          >{saving ? 'Signing in…' : config.bootstrap ? 'Create administrator' : 'Sign in'}</button
        >
      {/if}
      {#if config.oidcEnabled}
        {#if config.passwordLoginEnabled}<div
            class="my-4 flex items-center gap-3 text-xs text-text-subtle before:h-px before:flex-1 before:bg-border after:h-px after:flex-1 after:bg-border"
          >
            or
          </div>{/if}
        <a
          href="/api/auth/oidc/login"
          class="flex w-full items-center justify-center rounded-btn border border-border bg-bg-elevated py-2.5 text-sm font-medium text-text hover:bg-surface-2"
          >{config.oidcButtonTitle}</a
        >
      {/if}
    {:else}<button type="button" class="w-full text-sm text-primary" onclick={() => location.reload()}>Try again</button>{/if}
  </form>
</AuthLayout>
<ThemeFab />
