<script lang="ts">
  import Modal from './Modal.svelte'
  import Button from './Button.svelte'
  import { api, type Notifier } from '../lib/api'
  import { toast, toastError } from '../lib/toasts'
  import { field, label } from '../lib/styles'
  let {
    open = $bindable(false),
    notifier = null,
    onSaved,
  }: { open?: boolean; notifier?: Notifier | null; onSaved: () => void | Promise<void> } = $props()
  let name = $state(''),
    type = $state<Notifier['type']>('telegram'),
    token = $state(''),
    chat = $state(''),
    appToken = $state(''),
    userKey = $state(''),
    priority = $state(0),
    address = $state('localhost:6379'),
    password = $state(''),
    db = $state(0),
    prefix = $state('uptimelit'),
    saving = $state(false),
    error = $state('')
  $effect(() => {
    if (open) {
      let cfg: Record<string, unknown> = {}
      try {
        cfg = JSON.parse(notifier?.config_json ?? '{}')
      } catch {
        /* Server validates provider settings when saving. */
      }
      name = notifier?.name ?? ''
      type = notifier?.type ?? 'telegram'
      token = String(cfg.bot_token ?? '')
      chat = String(cfg.chat_id ?? '')
      appToken = String(cfg.app_token ?? '')
      userKey = String(cfg.user_key ?? '')
      priority = Number(cfg.priority ?? 0)
      address = String(cfg.address ?? 'localhost:6379')
      password = String(cfg.password ?? '')
      db = Number(cfg.db ?? 0)
      prefix = String(cfg.key_prefix ?? 'uptimelit')
      error = ''
    }
  })
  async function save(event: SubmitEvent) {
    event.preventDefault()
    saving = true
    error = ''
    const config =
      type === 'telegram'
        ? { bot_token: token.trim(), chat_id: chat.trim() }
        : type === 'pushover'
          ? { app_token: appToken.trim(), user_key: userKey.trim(), priority }
          : { address: address.trim(), password, db, key_prefix: prefix.trim() || 'uptimelit' }
    try {
      await api.saveNotifier({ name: name.trim(), type, config_json: JSON.stringify(config) }, notifier?.id)
      open = false
      toast.success(notifier ? 'Notifier updated' : 'Notifier created')
      await onSaved()
    } catch (e) {
      error = e instanceof Error ? e.message : 'Could not save notifier'
      toastError(e)
    } finally {
      saving = false
    }
  }
</script>

<Modal
  bind:open
  title={notifier ? 'Edit notifier' : 'New notifier'}
  description="Choose where your outage and recovery alerts are delivered."
>
  <form class="space-y-4" onsubmit={save}>
    {#if error}<p role="alert" class="rounded-lg bg-danger-soft px-3 py-2 text-sm text-danger">{error}</p>{/if}
    <label class="block"
      ><span class={label}>Name</span><input class={field} bind:value={name} required placeholder="Operations team" /></label
    >
    <label class="block"
      ><span class={label}>Provider</span><select class={field} bind:value={type}
        ><option value="telegram">Telegram</option><option value="pushover">Pushover</option><option value="redis">Redis</option></select
      ></label
    >
    {#if type === 'telegram'}
      <label class="block"
        ><span class={label}>Bot token</span><input
          class={field}
          type="password"
          bind:value={token}
          autocomplete="off"
          required
          placeholder="123456:ABC…"
        /></label
      >
      <label class="block"
        ><span class={label}>Chat ID</span><input class={field} bind:value={chat} required placeholder="-1001234567890" /></label
      >
    {:else if type === 'pushover'}
      <label class="block"
        ><span class={label}>Application token</span><input
          class={field}
          type="password"
          bind:value={appToken}
          autocomplete="off"
          required
        /></label
      >
      <label class="block"
        ><span class={label}>User key</span><input class={field} type="password" bind:value={userKey} autocomplete="off" required /></label
      >
      <label class="block"
        ><span class={label}>Priority</span><select class={field} bind:value={priority}
          ><option value={-1}>Low</option><option value={0}>Normal</option><option value={1}>High</option></select
        ></label
      >
    {:else}
      <label class="block"
        ><span class={label}>Redis address</span><input class={field} bind:value={address} required placeholder="redis:6379" /></label
      >
      <label class="block"
        ><span class={label}>Password (optional)</span><input
          class={field}
          type="password"
          bind:value={password}
          autocomplete="off"
        /></label
      >
      <div class="grid grid-cols-2 gap-3">
        <label class="block"
          ><span class={label}>Database</span><input class={field} type="number" min="0" step="1" bind:value={db} required /></label
        ><label class="block"><span class={label}>Key prefix</span><input class={field} bind:value={prefix} /></label>
      </div>
    {/if}
    <div class="flex justify-end gap-2">
      <Button variant="secondary" disabled={saving} onclick={() => (open = false)}>Cancel</Button><Button type="submit" disabled={saving}
        >{saving ? 'Saving…' : notifier ? 'Save changes' : 'Create notifier'}</Button
      >
    </div>
  </form>
</Modal>
