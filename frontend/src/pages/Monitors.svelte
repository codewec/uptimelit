<script lang="ts">
  import { onMount } from 'svelte'
  import { api, type Monitor, type Notifier } from '../lib/api'
  import { searchQuery } from '../lib/stores'
  import { toast, toastError } from '../lib/toasts'
  import AppLayout from '../layouts/AppLayout.svelte'
  import MonitorCard from '../components/MonitorCard.svelte'
  import MonitorFormModal from '../components/MonitorFormModal.svelte'
  import HistoryModal from '../components/HistoryModal.svelte'
  import ConfirmModal from '../components/ConfirmModal.svelte'
  import EmptyState from '../components/EmptyState.svelte'
  import Button from '../components/Button.svelte'
  type StatusFilter = 'all' | 'up' | 'down' | 'paused'
  const statuses = [
    { value: 'all', label: 'All', color: 'bg-primary' },
    { value: 'up', label: 'Up', color: 'bg-success' },
    { value: 'down', label: 'Down', color: 'bg-danger' },
    { value: 'paused', label: 'Paused', color: 'bg-text-subtle' },
  ] as const
  let monitors = $state<Monitor[]>([]),
    notifiers = $state<Notifier[]>([]),
    loading = $state(true),
    error = $state(''),
    filter = $state<StatusFilter>('all')
  let formOpen = $state(false),
    editing = $state<Monitor | null>(null),
    historyOpen = $state(false),
    historyMonitor = $state<Monitor | null>(null),
    deleteOpen = $state(false),
    deleting = $state<Monitor | null>(null),
    busy = $state<number[]>([])
  let controller: AbortController
  let loadID = 0
  const counts = $derived({
    all: monitors.length,
    up: monitors.filter((m) => m.is_up === true && !m.paused).length,
    down: monitors.filter((m) => m.is_up === false && !m.paused).length,
    paused: monitors.filter((m) => m.paused).length,
  })
  const filtered = $derived(
    monitors.filter(
      (m) =>
        (filter === 'all' ||
          (filter === 'paused' && m.paused) ||
          (filter === 'up' && m.is_up === true && !m.paused) ||
          (filter === 'down' && m.is_up === false && !m.paused)) &&
        `${m.name} ${m.url}`.toLowerCase().includes($searchQuery.toLowerCase()),
    ),
  )
  async function load() {
    const id = ++loadID
    try {
      const [rows, channels] = await Promise.all([api.monitors(controller?.signal), api.notifiers(controller?.signal)])
      if (id !== loadID) return
      monitors = rows ?? []
      notifiers = channels ?? []
      error = ''
    } catch (e) {
      if (id === loadID && e instanceof Error && e.name !== 'AbortError') error = e.message
    } finally {
      if (id === loadID) loading = false
    }
  }
  onMount(() => {
    controller = new AbortController()
    let timer: ReturnType<typeof setTimeout>
    let disposed = false
    async function poll() {
      if (!document.hidden) await load()
      if (!disposed) timer = setTimeout(poll, 5000)
    }
    void poll()
    return () => {
      disposed = true
      clearTimeout(timer)
      controller.abort()
    }
  })
  function edit(m: Monitor | null) {
    editing = m
    formOpen = true
  }
  function selectFilter(value: StatusFilter) {
    filter = value
    if (value === 'all') searchQuery.set('')
  }
  function history(m: Monitor) {
    historyMonitor = m
    historyOpen = true
  }
  async function pause(m: Monitor) {
    busy = [...busy, m.id]
    try {
      await api.pauseMonitor(m.id)
      toast.success(m.paused ? 'Monitoring resumed' : 'Monitor paused')
      await load()
    } catch (e) {
      toastError(e)
    } finally {
      busy = busy.filter((id) => id !== m.id)
    }
  }
  async function remove() {
    if (!deleting) return
    await api.deleteMonitor(deleting.id)
    toast.success('Monitor deleted')
    await load()
  }
</script>

<AppLayout showSearch onNewMonitor={() => edit(null)}>
  <div role="group" aria-label="Filter monitors by status" class="mb-6 flex flex-wrap items-center gap-2">
    {#each statuses as status (status.value)}
      <button
        type="button"
        aria-label={status.label + ' monitors'}
        aria-pressed={filter === status.value}
        onclick={() => selectFilter(status.value)}
        class={'inline-flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-xs font-medium transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary ' +
          (filter === status.value
            ? 'border-primary/40 bg-primary/10 text-primary'
            : 'border-border-soft bg-surface/60 text-text-muted hover:border-border hover:bg-surface-2 hover:text-text')}
      >
        <span class={'h-1.5 w-1.5 rounded-full ' + status.color}></span>
        {status.label}<span class="tabular-nums opacity-70">{counts[status.value]}</span>
      </button>
    {/each}
  </div>
  {#if error}<div
      role="alert"
      class="mb-4 flex items-center justify-between gap-3 rounded-card bg-danger-soft px-4 py-3 text-sm text-danger"
    >
      <span>{error}</span><button type="button" onclick={load} class="shrink-0 font-medium underline">Retry</button>
    </div>{/if}
  {#if loading}<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3" aria-label="Loading monitors">
      {#each [1, 2, 3] as i}<div class="h-56 animate-pulse rounded-card border border-border-soft bg-surface/80"></div>{/each}
    </div>
  {:else if !monitors.length && !error}<EmptyState
      title="Your first monitor"
      description="Add a website, API or TCP service. We'll check it regularly and let you know when it needs attention."
    />
  {:else if !filtered.length && !error}<EmptyState
      icon="search"
      title="No matching monitors"
      description="Try another search or change the status filter."
      ><Button variant="secondary" onclick={() => selectFilter('all')}>Clear filters</Button></EmptyState
    >
  {:else}<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {#each filtered as monitor (monitor.id)}<MonitorCard
          {monitor}
          {notifiers}
          busy={busy.includes(monitor.id)}
          onEdit={() => edit(monitor)}
          onPause={() => pause(monitor)}
          onDelete={() => {
            deleting = monitor
            deleteOpen = true
          }}
          onHistory={() => history(monitor)}
        />{/each}
    </div>{/if}
</AppLayout>
<MonitorFormModal bind:open={formOpen} monitor={editing} {notifiers} onSaved={load} />
<HistoryModal bind:open={historyOpen} monitor={historyMonitor} />
<ConfirmModal
  bind:open={deleteOpen}
  title="Delete monitor"
  message={`Delete “${deleting?.name ?? ''}” and all its check history? This cannot be undone.`}
  onConfirm={remove}
/>
