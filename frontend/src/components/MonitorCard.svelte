<script lang="ts">
  import { networkStatus } from '../lib/checks'
  import { DropdownMenu } from 'bits-ui'
  import type { Monitor, Notifier } from '../lib/api'
  import Icon from './Icon.svelte'
  import HistoryBars from './HistoryBars.svelte'
  let {
    monitor,
    notifiers,
    onEdit,
    onPause,
    onDelete,
    onHistory,
    busy = false,
  }: {
    monitor: Monitor
    notifiers: Notifier[]
    onEdit: () => void
    onPause: () => void
    onDelete: () => void
    onHistory: () => void
    busy?: boolean
  } = $props()
  const state = $derived(monitor.paused ? 'Paused' : monitor.is_up === true ? 'Up' : monitor.is_up === false ? 'Down' : 'Pending')
  const color = $derived(
    monitor.paused
      ? 'text-text-subtle'
      : monitor.is_up === true
        ? 'text-success'
        : monitor.is_up === false
          ? 'text-danger'
          : 'text-text-muted',
  )
  const linked = $derived(notifiers.filter((n) => monitor.notifier_ids?.includes(n.id)))
</script>

<section class="relative h-full rounded-card border border-border-soft bg-surface/80 p-4 backdrop-blur-sm">
  <header class="mb-4 flex items-start gap-2">
    <span class="mt-0.5 rounded-md bg-bg-elevated p-1.5 text-text-muted"
      ><Icon name={monitor.type === 'http' ? 'monitor' : 'activity'} size={18} /></span
    >
    <div class="min-w-0 flex-1">
      <h2 class="truncate text-sm font-semibold tracking-tight text-text" title={monitor.name}>{monitor.name}</h2>
      {#if monitor.type === 'http'}<a
          href={monitor.url}
          target="_blank"
          rel="noopener noreferrer"
          class="mt-0.5 block truncate text-[11px] text-text-subtle hover:text-primary"
          title={monitor.url}>{monitor.url}</a
        >{:else}<p class="mt-0.5 truncate text-[11px] text-text-subtle">{monitor.url}</p>{/if}
    </div>
    <DropdownMenu.Root
      ><DropdownMenu.Trigger
        disabled={busy}
        class="flex h-7 w-7 shrink-0 items-center justify-center rounded-btn text-text-subtle hover:bg-surface-2 hover:text-text disabled:opacity-40"
        aria-label={'Actions for ' + monitor.name}><Icon name="more" /></DropdownMenu.Trigger
      ><DropdownMenu.Portal
        ><DropdownMenu.Content
          class="z-50 min-w-40 rounded-xl border border-border bg-surface p-1 shadow-xl outline-none"
          align="end"
          sideOffset={5}
        >
          {#each [{ label: 'History', icon: 'history', action: onHistory }, { label: 'Edit', icon: 'edit', action: onEdit }, { label: monitor.paused ? 'Resume' : 'Pause', icon: monitor.paused ? 'play' : 'pause', action: onPause }] as item}<DropdownMenu.Item
              class="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-xs outline-none data-highlighted:bg-surface-2"
              onSelect={item.action}><Icon name={item.icon} size={14} />{item.label}</DropdownMenu.Item
            >{/each}
          <DropdownMenu.Separator class="my-1 h-px bg-border-soft" /><DropdownMenu.Item
            class="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-xs text-danger outline-none data-highlighted:bg-danger-soft"
            onSelect={onDelete}><Icon name="trash" size={14} />Delete</DropdownMenu.Item
          >
        </DropdownMenu.Content></DropdownMenu.Portal
      ></DropdownMenu.Root
    >
  </header>
  <div class="mb-4 flex items-center justify-between gap-2">
    <span
      class={'inline-flex items-center gap-1.5 rounded-full border border-current/15 bg-current/5 px-2 py-0.5 text-[11px] font-medium ' +
        color}><span class="h-1.5 w-1.5 rounded-full bg-current"></span>{state}</span
    ><span class="text-[10px] uppercase tracking-wider text-text-subtle">{monitor.type} · {monitor.interval_seconds}s</span>
  </div>
  <HistoryBars
    checks={monitor.recent_checks ?? []}
    paused={monitor.paused}
    monitorName={monitor.name}
    protocol={monitor.type}
    {onHistory}
  />
  <div class="mt-3 grid grid-cols-3 divide-x divide-border-soft text-center">
    <div>
      <p class="text-[10px] text-text-subtle">Uptime · 24h</p>
      <p class="mt-0.5 text-xs font-semibold tabular-nums">{monitor.uptime_24h == null ? '—' : monitor.uptime_24h.toFixed(2) + '%'}</p>
    </div>
    <div>
      <p class="text-[10px] text-text-subtle">Response</p>
      <p class="mt-0.5 text-xs font-semibold tabular-nums">{monitor.response_time_ms == null ? '—' : monitor.response_time_ms + ' ms'}</p>
    </div>
    <div>
      <p class="text-[10px] text-text-subtle">
        {monitor.type === 'http' ? 'HTTP status' : monitor.type === 'ping' ? 'ICMP' : 'Connection'}
      </p>
      <p class="mt-0.5 text-xs font-semibold tabular-nums">
        {monitor.type === 'http' ? (monitor.status_code ?? '—') : networkStatus(monitor.is_up, monitor.error_code, monitor.type)}
      </p>
    </div>
  </div>
  <div class="mt-4 flex items-center justify-between gap-2 border-t border-border-soft pt-3 text-[10px] text-text-subtle">
    <span class="flex items-center gap-1" title={linked.map((n) => n.name).join(', ') || 'No notifications'}
      ><Icon name="bell" size={12} />{linked.length
        ? `${linked.length} notifier${linked.length === 1 ? '' : 's'}`
        : 'No notifications'}</span
    ><span title={monitor.checked_at ? new Date(monitor.checked_at).toLocaleString() : ''}
      >{monitor.checked_at
        ? 'Checked ' + new Date(monitor.checked_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
        : 'Waiting for first check'}</span
    >
  </div>
</section>
