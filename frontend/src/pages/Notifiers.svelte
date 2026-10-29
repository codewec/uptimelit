<script lang="ts">
  import { onMount } from 'svelte'
  import { api, type Notifier } from '../lib/api'
  import { searchQuery } from '../lib/stores'
  import { toast, toastError } from '../lib/toasts'
  import AppLayout from '../layouts/AppLayout.svelte'
  import NotifierFormModal from '../components/NotifierFormModal.svelte'
  import ConfirmModal from '../components/ConfirmModal.svelte'
  import EmptyState from '../components/EmptyState.svelte'
  import Button from '../components/Button.svelte'
  import Icon from '../components/Icon.svelte'
  let notifiers = $state<Notifier[]>([]),
    loading = $state(true),
    error = $state(''),
    formOpen = $state(false),
    editing = $state<Notifier | null>(null),
    deleteOpen = $state(false),
    deleting = $state<Notifier | null>(null),
    sending = $state<number[]>([])
  let controller: AbortController
  let loadID = 0
  const filtered = $derived(notifiers.filter((n) => `${n.name} ${n.type}`.toLowerCase().includes($searchQuery.toLowerCase())))
  async function load() {
    const id = ++loadID
    try {
      const rows = await api.notifiers(controller?.signal)
      if (id !== loadID) return
      notifiers = rows ?? []
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
  function edit(n: Notifier | null) {
    editing = n
    formOpen = true
  }
  async function test(n: Notifier) {
    sending = [...sending, n.id]
    try {
      await api.testNotifier(n.id)
      toast.success('Test notification sent')
    } catch (e) {
      toastError(e)
    } finally {
      sending = sending.filter((id) => id !== n.id)
    }
  }
  async function remove() {
    if (!deleting) return
    await api.deleteNotifier(deleting.id)
    toast.success('Notifier deleted')
    await load()
  }
</script>

<AppLayout showSearch>
  <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
    <div>
      <h1 class="text-xl font-semibold tracking-tight">Notifiers</h1>
      <p class="mt-1 text-sm text-text-muted">The right alerts, in the right place.</p>
    </div>
    <Button onclick={() => edit(null)}><Icon name="plus" size={15} />New notifier</Button>
  </div>
  {#if error}<div role="alert" class="mb-4 flex items-center justify-between rounded-card bg-danger-soft px-4 py-3 text-sm text-danger">
      <span>{error}</span><button type="button" onclick={load} class="underline">Retry</button>
    </div>{/if}
  {#if loading}<p class="py-16 text-center text-sm text-text-muted">Loading notifiers…</p>
  {:else if !notifiers.length && !error}<EmptyState
      icon="bell"
      title="Stay in the loop"
      description="Connect Telegram, Pushover or Redis, then attach a channel to your monitors."
      ><Button onclick={() => edit(null)}><Icon name="plus" size={15} />Create notifier</Button></EmptyState
    >
  {:else if !filtered.length && !error}<EmptyState icon="search" title="No matching notifiers" description="Try a different search."
      ><Button variant="secondary" onclick={() => searchQuery.set('')}>Clear search</Button></EmptyState
    >
  {:else}<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {#each filtered as notifier (notifier.id)}<section
          class="flex flex-col rounded-card border border-border-soft bg-surface/80 p-4 backdrop-blur-sm"
        >
          <header class="mb-4 flex items-center gap-3">
            <span class="rounded-lg border border-border-soft bg-bg-elevated p-2 text-primary"><Icon name="bell" size={20} /></span>
            <div class="min-w-0">
              <h2 class="truncate text-sm font-semibold tracking-tight">{notifier.name}</h2>
              <span class="mt-1 block text-[10px] uppercase tracking-wider text-text-subtle">{notifier.type}</span>
            </div>
          </header>
          <p class="mb-5 flex-1 text-xs leading-relaxed text-text-muted">
            {{
              telegram: 'Outage and recovery messages delivered to your Telegram chat.',
              pushover: 'Push notifications sent directly to your devices.',
              redis: 'Notification events recorded in your Redis database.',
            }[notifier.type]}
          </p>
          <div class="flex items-center justify-between gap-2 border-t border-border-soft pt-3">
            <Button
              class="px-2.5 py-1.5 text-xs"
              variant="secondary"
              disabled={sending.includes(notifier.id)}
              onclick={() => test(notifier)}
              ><Icon name="send" size={13} />{sending.includes(notifier.id) ? 'Sending…' : 'Send test'}</Button
            >
            <div class="flex items-center gap-1">
              <Button class="px-2.5 py-1.5 text-xs" variant="ghost" onclick={() => edit(notifier)}>Edit</Button><Button
                class="px-2.5 py-1.5 text-xs"
                variant="danger"
                onclick={() => {
                  deleting = notifier
                  deleteOpen = true
                }}>Delete</Button
              >
            </div>
          </div>
        </section>{/each}
    </div>{/if}
</AppLayout>
<NotifierFormModal bind:open={formOpen} notifier={editing} onSaved={load} />
<ConfirmModal
  bind:open={deleteOpen}
  title="Delete notifier"
  message={`Delete “${deleting?.name ?? ''}”? It will be unlinked from your monitors.`}
  onConfirm={remove}
/>
