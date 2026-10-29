<script lang="ts">
  import { onMount } from 'svelte'
  import { api, type AuthConfig } from '../lib/api'
  import { user } from '../lib/stores'
  import { replace } from '../lib/router'
  import { toast, toastError } from '../lib/toasts'
  import { field, label, panel } from '../lib/styles'
  import AppLayout from '../layouts/AppLayout.svelte'
  import Button from '../components/Button.svelte'
  let username = $state(''),
    currentPassword = $state(''),
    newPassword = $state(''),
    confirmPassword = $state(''),
    saving = $state(false),
    config = $state<AuthConfig | null>(null),
    error = $state('')
  onMount(() => {
    username = $user?.username ?? ''
    const controller = new AbortController()
    api
      .authConfig(controller.signal)
      .then((c) => (config = c))
      .catch((e) => {
        if (e.name !== 'AbortError') error = e.message
      })
    return () => controller.abort()
  })
  async function save(event: SubmitEvent, password = false) {
    event.preventDefault()
    if (password && newPassword !== confirmPassword) {
      toast.error('Passwords do not match')
      return
    }
    if (password) {
      const size = new TextEncoder().encode(newPassword).length
      if (size < 8 || size > 72) {
        toast.error('Password must contain 8–72 bytes')
        return
      }
    }
    saving = true
    try {
      await api.profile({
        username: password ? ($user?.username ?? username) : username.trim(),
        ...(password ? { newPassword, currentPassword } : {}),
      })
      user.set(null)
      toast.success('Profile updated. Please sign in again.')
      replace('/login')
    } catch (e) {
      toastError(e)
    } finally {
      saving = false
    }
  }
</script>

<AppLayout>
  <div class="mb-6">
    <h1 class="text-xl font-semibold tracking-tight">Profile</h1>
    <p class="mt-1 text-sm text-text-muted">Manage your account and sign-in settings.</p>
    <p class="mt-2 inline-flex items-center gap-1.5 rounded-full border border-border bg-surface px-2.5 py-1 text-xs text-text-muted">
      Signed in with <span class="font-medium text-text">{$user?.authMethod === 'oidc' ? 'OIDC' : 'password'}</span>
    </p>
  </div>
  {#if error}<p role="alert" class="mb-4 rounded-card bg-danger-soft px-4 py-3 text-sm text-danger">{error}</p>{/if}
  <div class={'grid items-start gap-4 ' + (config?.passwordLoginEnabled ? 'lg:grid-cols-2' : '')}>
    <form class={panel} onsubmit={(event) => save(event)}>
      <h2 class="text-sm font-semibold">Username</h2>
      <label class="mt-4 block"
        ><span class={label}>Login</span><input
          class={field}
          bind:value={username}
          maxlength="64"
          pattern="[a-zA-Z0-9._-]+"
          required
          autocomplete="username"
          disabled={$user?.authMethod === 'oidc' || saving}
        />{#if $user?.authMethod === 'oidc'}<span class="mt-1 block text-[11px] text-text-subtle"
            >Username is managed by the OIDC provider.</span
          >{/if}</label
      >{#if $user?.authMethod !== 'oidc'}<div class="mt-4 flex justify-end">
          <Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save username'}</Button>
        </div>{/if}
    </form>
    {#if config?.passwordLoginEnabled}<form class={panel} onsubmit={(event) => save(event, true)}>
        <h2 class="text-sm font-semibold">Password</h2>
        <div class="mt-4 space-y-3">
          {#if $user?.hasPassword}<label class="block"
              ><span class={label}>Current password</span><input
                class={field}
                type="password"
                bind:value={currentPassword}
                required
                autocomplete="current-password"
              /></label
            >{/if}
          <label class="block"
            ><span class={label}>New password</span><input
              class={field}
              type="password"
              bind:value={newPassword}
              required
              autocomplete="new-password"
            /><span class="mt-1 block text-[11px] text-text-subtle">8–72 bytes.</span></label
          >
          <label class="block"
            ><span class={label}>Confirm new password</span><input
              class={field}
              type="password"
              bind:value={confirmPassword}
              required
              autocomplete="new-password"
            /></label
          >
        </div>
        <div class="mt-4 flex justify-end"><Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Update password'}</Button></div>
      </form>{/if}
  </div>
  <p class="mt-4 text-xs text-text-subtle">Saving account changes signs out all your sessions. Sign in again to continue.</p>
</AppLayout>
