<script lang="ts">
  import { onMount } from 'svelte'
  import { api, type User, type AuthConfig } from '../lib/api'
  import { user } from '../lib/stores'
  import { replace } from '../lib/router'
  import { toast, toastError } from '../lib/toasts'
  import { field, label } from '../lib/styles'
  import AppLayout from '../layouts/AppLayout.svelte'
  import Modal from '../components/Modal.svelte'
  import ConfirmModal from '../components/ConfirmModal.svelte'
  import Button from '../components/Button.svelte'
  import Icon from '../components/Icon.svelte'
  let users = $state<User[]>([]),
    config = $state<AuthConfig | null>(null),
    loading = $state(true),
    error = $state(''),
    formOpen = $state(false),
    editing = $state<User | null>(null),
    name = $state(''),
    role = $state<User['role']>('user'),
    password = $state(''),
    resetOIDC = $state(false),
    saving = $state(false),
    formError = $state(''),
    deleteOpen = $state(false),
    deleting = $state<User | null>(null)
  let controller: AbortController
  let loadID = 0
  async function load() {
    const id = ++loadID
    try {
      const [rows, cfg] = await Promise.all([api.users(controller?.signal), api.authConfig(controller?.signal)])
      if (id !== loadID) return
      users = rows
      config = cfg
      error = ''
    } catch (e) {
      if (id === loadID && e instanceof Error && e.name !== 'AbortError') error = e.message
    } finally {
      if (id === loadID) loading = false
    }
  }
  onMount(() => {
    controller = new AbortController()
    void load()
    return () => controller.abort()
  })
  function edit(account: User | null) {
    editing = account
    name = account?.username ?? ''
    role = account?.role ?? 'user'
    password = ''
    resetOIDC = false
    formError = ''
    formOpen = true
  }
  async function save(event: SubmitEvent) {
    event.preventDefault()
    saving = true
    formError = ''
    try {
      await api.saveUser({ username: name.trim(), role, ...(editing ? { newPassword: password, resetOIDC } : { password }) }, editing?.id)
      formOpen = false
      if (editing?.id === $user?.id) {
        user.set(null)
        replace('/login')
        return
      }
      toast.success(editing ? 'User updated' : 'User created')
      await load()
    } catch (e) {
      formError = e instanceof Error ? e.message : 'Could not save user'
      toastError(e)
    } finally {
      saving = false
    }
  }
  async function remove() {
    if (!deleting) return
    await api.deleteUser(deleting.id)
    toast.success('User and their data deleted')
    await load()
  }
</script>

<AppLayout>
  <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
    <div>
      <h1 class="text-xl font-semibold tracking-tight">Administration</h1>
      <p class="mt-1 text-sm text-text-muted">Manage users and their access to UptimeLit.</p>
    </div>
    <Button onclick={() => edit(null)}><Icon name="plus" size={15} />Create user</Button>
  </div>
  {#if error}<p role="alert" class="mb-4 rounded-card bg-danger-soft px-4 py-3 text-sm text-danger">
      {error} <button type="button" class="underline" onclick={load}>Retry</button>
    </p>{/if}
  {#if loading}<p class="py-16 text-center text-sm text-text-muted">Loading users…</p>{:else}
    <section class="rounded-card border border-border-soft bg-surface">
      <div class="border-b border-border-soft px-5 py-4">
        <h2 class="text-sm font-semibold">Users <span class="ml-1 text-xs font-normal text-text-subtle">{users.length}</span></h2>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full min-w-xl text-left text-sm">
          <thead class="text-xs text-text-muted"
            ><tr class="border-b border-border-soft"
              ><th class="px-5 py-3 font-medium">Login</th><th class="px-5 py-3 font-medium">Role</th><th class="px-5 py-3 font-medium"
                >Sign-in methods</th
              ><th class="px-5 py-3 text-right font-medium">Actions</th></tr
            ></thead
          ><tbody
            >{#each users as account (account.id)}<tr class="border-b border-border-soft last:border-0 hover:bg-surface-2"
                ><td class="px-5 py-3 font-medium"
                  >{account.username}{#if account.id === $user?.id}<span class="ml-2 text-[10px] font-normal text-text-subtle">You</span
                    >{/if}</td
                ><td class="px-5 py-3 capitalize text-text-muted">{account.role}</td><td class="px-5 py-3 text-text-muted"
                  >{account.hasPassword && account.hasOIDC
                    ? 'Password + OIDC'
                    : account.hasOIDC
                      ? 'OIDC'
                      : account.hasPassword
                        ? 'Password'
                        : 'Awaiting OIDC'}</td
                ><td class="px-5 py-3"
                  ><div class="flex justify-end gap-1">
                    <Button class="px-2.5 py-1.5 text-xs" variant="secondary" onclick={() => edit(account)}>Edit</Button><Button
                      class="px-2.5 py-1.5 text-xs"
                      variant="danger"
                      disabled={account.id === $user?.id}
                      onclick={() => {
                        deleting = account
                        deleteOpen = true
                      }}>Delete</Button
                    >
                  </div></td
                ></tr
              >{/each}</tbody
          >
        </table>
      </div>
    </section>
    <p class="mt-4 text-xs text-text-subtle">
      Administrators manage accounts; monitoring data remains private to each owner. The last administrator cannot be removed or demoted.
    </p>
  {/if}
</AppLayout>
<Modal
  bind:open={formOpen}
  title={editing ? 'Edit user' : 'Create user'}
  description={editing
    ? 'Saving account changes revokes this user’s sessions.'
    : 'Create a password account or pre-provision a username for OIDC.'}
>
  <form class="space-y-4" onsubmit={save}>
    {#if formError}<p role="alert" class="rounded-lg bg-danger-soft px-3 py-2 text-sm text-danger">{formError}</p>{/if}
    <label class="block"
      ><span class={label}>Username</span><input
        class={field}
        bind:value={name}
        maxlength="64"
        pattern="[a-zA-Z0-9._-]+"
        required
        autocomplete="off"
      /></label
    >
    <label class="block"
      ><span class={label}>Role</span><select class={field} bind:value={role}
        ><option value="user">User</option><option value="admin">Administrator</option></select
      ></label
    >
    <label class="block"
      ><span class={label}>{editing ? 'New password (leave empty to keep)' : 'Password'}</span><input
        class={field}
        type="password"
        bind:value={password}
        autocomplete="new-password"
        required={!editing && !config?.oidcEnabled}
      /><span class="mt-1 block text-[11px] text-text-subtle"
        >{!editing && config?.oidcEnabled ? 'Optional for OIDC accounts. Otherwise use 8–72 bytes.' : '8–72 bytes.'}</span
      ></label
    >
    {#if editing?.hasOIDC}<label
        class="flex items-start gap-2 rounded-lg border border-border-soft bg-bg-elevated p-3 text-xs text-text-muted"
        ><input type="checkbox" class="mt-0.5 accent-primary" bind:checked={resetOIDC} />Reset OIDC binding. An available password login
        method is required.</label
      >{/if}
    {#if !editing && config?.oidcEnabled}<p class="text-xs leading-relaxed text-text-subtle">
        OIDC links to this account when the provider's normalized username matches. No public registration is enabled.
      </p>{/if}
    <div class="flex justify-end gap-2">
      <Button variant="secondary" disabled={saving} onclick={() => (formOpen = false)}>Cancel</Button><Button
        type="submit"
        disabled={saving}>{saving ? 'Saving…' : editing ? 'Save changes' : 'Create user'}</Button
      >
    </div>
  </form>
</Modal>
<ConfirmModal
  bind:open={deleteOpen}
  title="Delete user"
  message={`Delete “${deleting?.username ?? ''}” and all their monitors, history and notifiers? This cannot be undone.`}
  onConfirm={remove}
/>
