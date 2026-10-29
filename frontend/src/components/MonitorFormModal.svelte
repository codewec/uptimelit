<script lang="ts">
  import Modal from './Modal.svelte'
  import Button from './Button.svelte'
  import { api, type Monitor, type Notifier } from '../lib/api'
  import { toast, toastError } from '../lib/toasts'
  import { field, label } from '../lib/styles'
  let {
    open = $bindable(false),
    monitor = null,
    notifiers,
    onSaved,
  }: { open?: boolean; monitor?: Monitor | null; notifiers: Notifier[]; onSaved: () => void | Promise<void> } = $props()
  let name = $state(''),
    type = $state<'http' | 'tcp'>('http'),
    url = $state(''),
    interval = $state(60),
    threshold = $state(3),
    selected = $state<number[]>([]),
    saving = $state(false),
    error = $state('')
  $effect(() => {
    if (open) {
      name = monitor?.name ?? ''
      type = monitor?.type ?? 'http'
      url = monitor?.url ?? ''
      interval = monitor?.interval_seconds ?? 60
      threshold = monitor?.fail_threshold ?? 3
      selected = [...(monitor?.notifier_ids ?? [])]
      error = ''
    }
  })
  async function save(event: SubmitEvent) {
    event.preventDefault()
    saving = true
    error = ''
    try {
      await api.saveMonitor(
        { name: name.trim(), type, url: url.trim(), interval_seconds: interval, fail_threshold: threshold, notifier_ids: selected },
        monitor?.id,
      )
      open = false
      toast.success(monitor ? 'Monitor updated' : 'Monitor created')
      await onSaved()
    } catch (e) {
      error = e instanceof Error ? e.message : 'Could not save monitor'
      toastError(e)
    } finally {
      saving = false
    }
  }
</script>

<Modal bind:open title={monitor ? 'Edit monitor' : 'New monitor'} description="Keep an eye on a website, API or TCP service.">
  <form onsubmit={save} class="space-y-4">
    {#if error}<p role="alert" class="rounded-lg bg-danger-soft px-3 py-2 text-sm text-danger">{error}</p>{/if}
    <label class="block"
      ><span class={label}>Name</span><input class={field} bind:value={name} required placeholder="Production API" /></label
    >
    <label class="block"
      ><span class={label}>Monitor type</span><select class={field} bind:value={type}
        ><option value="http">HTTP / HTTPS</option><option value="tcp">TCP port</option></select
      ></label
    >
    <label class="block"
      ><span class={label}>{type === 'http' ? 'URL' : 'Address'}</span><input
        class={field}
        type={type === 'http' ? 'url' : 'text'}
        bind:value={url}
        placeholder={type === 'http' ? 'https://api.example.com/health' : 'db.example.com:5432'}
        required
      /></label
    >
    <div class="grid grid-cols-2 gap-3">
      <label class="block"
        ><span class={label}>Interval (seconds)</span><input
          class={field}
          type="number"
          min="10"
          step="1"
          required
          bind:value={interval}
        /></label
      ><label class="block"
        ><span class={label}>Failure threshold</span><input
          class={field}
          type="number"
          min="1"
          step="1"
          required
          bind:value={threshold}
        /></label
      >
    </div>
    <fieldset>
      <legend class={label}>Notification channels</legend>{#if notifiers.length}<div
          class="max-h-36 space-y-1 overflow-auto rounded-btn border border-border-soft p-2"
        >
          {#each notifiers as notifier (notifier.id)}<label
              class="flex cursor-pointer items-center gap-2 rounded-lg p-2 text-xs hover:bg-surface-2"
              ><input type="checkbox" class="accent-primary" bind:group={selected} value={notifier.id} /><span class="flex-1"
                >{notifier.name}</span
              ><span class="text-[10px] uppercase text-text-subtle">{notifier.type}</span></label
            >{/each}
        </div>{:else}<p class="rounded-btn border border-dashed border-border p-3 text-xs text-text-subtle">
          Create a channel in Notifiers to receive alerts.
        </p>{/if}
    </fieldset>
    <div class="flex justify-end gap-2">
      <Button variant="secondary" disabled={saving} onclick={() => (open = false)}>Cancel</Button><Button type="submit" disabled={saving}
        >{saving ? 'Saving…' : monitor ? 'Save changes' : 'Create monitor'}</Button
      >
    </div>
  </form>
</Modal>
