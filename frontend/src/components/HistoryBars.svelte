<script lang="ts">
  import { connectionStatus } from '../lib/checks'
  import { tick } from 'svelte'
  import { Tooltip } from 'bits-ui'
  import type { Check } from '../lib/api'
  import Icon from './Icon.svelte'

  let {
    checks = [],
    paused = false,
    monitorName,
    protocol,
    onHistory,
  }: {
    checks?: Check[]
    paused?: boolean
    monitorName: string
    protocol: 'http' | 'tcp'
    onHistory: () => void
  } = $props()

  const recent = $derived(checks.slice(-50))
  const slots = $derived([
    ...Array.from({ length: Math.max(0, 50 - recent.length) }, (_, index) => ({ key: `empty-${index}`, check: null })),
    ...recent.map((check) => ({ key: check.checked_at, check })),
  ])
  const tether = Tooltip.createTether<Check>()
  let focusedKey = $state<string | null>(null)
  let row: HTMLDivElement
  const tabKey = $derived(focusedKey && recent.some((check) => check.checked_at === focusedKey) ? focusedKey : recent.at(-1)?.checked_at)

  async function navigate(event: KeyboardEvent, key: string) {
    if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return
    event.preventDefault()
    const index = recent.findIndex((check) => check.checked_at === key)
    const next =
      event.key === 'Home'
        ? 0
        : event.key === 'End'
          ? recent.length - 1
          : Math.max(0, Math.min(recent.length - 1, index + (event.key === 'ArrowRight' ? 1 : -1)))
    const nextKey = recent[next]?.checked_at
    if (!nextKey) return
    focusedKey = nextKey
    await tick()
    Array.from(row.querySelectorAll<HTMLButtonElement>('[data-check-key]'))
      .find((button) => button.dataset.checkKey === nextKey)
      ?.focus()
  }
</script>

<Tooltip.Provider delayDuration={150} skipDelayDuration={300}>
  <div role="group" aria-label={'Check history for ' + monitorName}>
    <div bind:this={row} class="flex h-7 items-stretch">
      {#each slots as slot (slot.key)}
        {#if slot.check}
          {@const check = slot.check}
          <Tooltip.Trigger
            {tether}
            payload={check}
            data-check-key={slot.key}
            tabindex={slot.key === tabKey ? 0 : -1}
            aria-label={`${new Date(check.checked_at).toLocaleString()} · ${check.is_up ? 'Up' : 'Down'}`}
            onfocus={() => (focusedKey = slot.key)}
            onkeydown={(event) => navigate(event, slot.key)}
            class="history-trigger data-[state=instant-open]:z-10 data-[state=delayed-open]:z-10 relative min-w-0 flex-1 cursor-pointer px-[1.5px] outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 focus-visible:ring-offset-surface"
          >
            <span
              aria-hidden="true"
              class={'history-bar pointer-events-none block h-full rounded-[3px] ' +
                (paused ? 'bg-text-subtle/40' : check.is_up ? 'bg-success/75' : 'bg-danger/80')}
            ></span>
          </Tooltip.Trigger>
        {:else}
          <span aria-hidden="true" class="min-w-0 flex-1 px-[1.5px]"><span class="block h-full rounded-[3px] bg-border-soft"></span></span>
        {/if}
      {/each}
    </div>
    <Tooltip.Root {tether}>
      {#snippet children({ payload: check })}
        {#if check}
          <Tooltip.Portal>
            <Tooltip.Content
              role="tooltip"
              side="top"
              sideOffset={12}
              collisionPadding={12}
              class="history-tooltip z-50 w-64 max-w-[calc(100vw-1.5rem)] rounded-xl border border-border bg-surface p-3.5 text-xs text-text shadow-xl outline-none"
            >
              <div class="mb-3 flex items-center justify-between gap-3">
                <span class="min-w-0 truncate font-semibold" title={monitorName}>{monitorName}</span>
                <span
                  class={'inline-flex shrink-0 items-center gap-1.5 rounded-full px-2 py-0.5 text-[10px] font-medium ' +
                    (check.is_up ? 'bg-success/10 text-success' : 'bg-danger-soft text-danger')}
                  ><span class="h-1.5 w-1.5 rounded-full bg-current"></span>{check.is_up ? 'Up' : 'Down'}</span
                >
              </div>
              <dl class="space-y-2">
                <div class="flex items-start justify-between gap-3">
                  <dt class="shrink-0 text-text-muted">Checked at</dt>
                  <dd class="text-right tabular-nums">
                    <time datetime={check.checked_at}>{new Date(check.checked_at).toLocaleString()}</time>
                  </dd>
                </div>
                <div class="flex justify-between gap-3">
                  <dt class="text-text-muted">Protocol</dt>
                  <dd class="uppercase">{protocol}</dd>
                </div>
                {#if protocol === 'http'}<div class="flex justify-between gap-3">
                    <dt class="text-text-muted">HTTP status</dt>
                    <dd class="font-medium tabular-nums">{check.status_code ?? 'No response'}</dd>
                  </div>{:else}<div class="flex justify-between gap-3">
                    <dt class="text-text-muted">Connection</dt>
                    <dd class="text-right font-medium">{connectionStatus(check.is_up, check.error_code)}</dd>
                  </div>{/if}
                <div class="flex justify-between gap-3">
                  <dt class="text-text-muted">Response time</dt>
                  <dd class="font-medium tabular-nums">{check.response_time_ms == null ? '—' : `${check.response_time_ms} ms`}</dd>
                </div>
              </dl>
              {#if paused}<p class="mt-3 border-t border-border-soft pt-2 text-[10px] text-text-subtle">
                  Monitoring is currently paused.
                </p>{/if}
              <Tooltip.Arrow class="text-surface" width={12} height={6} />
            </Tooltip.Content>
          </Tooltip.Portal>
        {/if}
      {/snippet}
    </Tooltip.Root>
    <div class="mt-2 flex items-center justify-between gap-2 text-[10px] text-text-subtle">
      <span>{recent.length ? 'Recent checks' : 'No check history yet'}</span>
      <button
        type="button"
        onclick={onHistory}
        class="inline-flex items-center gap-1 rounded text-text-muted hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
        aria-label={'View history for ' + monitorName}><Icon name="history" size={11} />View history</button
      >
    </div>
  </div>
</Tooltip.Provider>

<style>
  .history-bar {
    transform-origin: center;
  }
  :global(.history-trigger[data-state='instant-open']) .history-bar,
  :global(.history-trigger[data-state='delayed-open']) .history-bar {
    transform: scale(1.5, 1.3);
    transition: transform 160ms cubic-bezier(0.2, 0.8, 0.2, 1);
  }
  :global(.history-tooltip) {
    transform-origin: var(--bits-tooltip-content-transform-origin);
  }
  :global(.history-tooltip[data-state='delayed-open']) {
    animation: history-tooltip-in 140ms ease-out;
  }
  :global(.history-tooltip[data-state='closed']) {
    animation: history-tooltip-out 120ms ease-in forwards;
  }
  @keyframes history-tooltip-out {
    from {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
    to {
      opacity: 0;
      transform: translateY(3px) scale(0.97);
    }
  }
  @keyframes history-tooltip-in {
    from {
      opacity: 0;
      transform: translateY(3px) scale(0.97);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .history-bar {
      transition: none !important;
    }
    :global(.history-tooltip) {
      animation: none !important;
    }
  }
</style>
