<script lang="ts">
  import { connectionStatus } from '../lib/checks'
  import Modal from './Modal.svelte'
  import Button from './Button.svelte'
  import { api, type Monitor, type Check } from '../lib/api'
  let { open = $bindable(false), monitor }: { open?: boolean; monitor: Monitor | null } = $props()
  let checks = $state<Check[]>([]),
    loading = $state(false),
    error = $state('')
  $effect(() => {
    if (!open || !monitor) return
    const controller = new AbortController()
    loading = true
    checks = []
    error = ''
    api
      .checks(monitor.id, controller.signal)
      .then((rows) => (checks = rows ?? []))
      .catch((e) => {
        if (e.name !== 'AbortError') error = e.message
      })
      .finally(() => (loading = false))
    return () => controller.abort()
  })
</script>

<Modal
  bind:open
  title={monitor?.name ?? 'Check history'}
  description="The most recent 100 checks, newest first."
  class="w-[min(44rem,calc(100vw-1.5rem))]"
>
  {#if loading}<p class="py-12 text-center text-sm text-text-muted">Loading checks…</p>{:else if error}<p
      role="alert"
      class="py-6 text-sm text-danger"
    >
      {error}
    </p>{:else if !checks.length}<p class="py-12 text-center text-sm text-text-muted">
      No checks yet. The first check will appear shortly.
    </p>{:else}<div class="max-h-[55vh] overflow-auto">
      <table class="w-full min-w-96 text-left text-xs">
        <thead class="sticky top-0 bg-surface text-text-muted"
          ><tr class="border-b border-border-soft"
            ><th class="px-2 py-3 font-medium">Time</th><th class="px-2 py-3 font-medium">Status</th><th class="px-2 py-3 font-medium"
              >{monitor?.type === 'tcp' ? 'Connection' : 'HTTP'}</th
            ><th class="px-2 py-3 text-right font-medium">Response</th></tr
          ></thead
        ><tbody
          >{#each checks as check, index (check.id ?? index)}<tr class="border-b border-border-soft last:border-0"
              ><td class="px-2 py-3 text-text-muted">{new Date(check.checked_at).toLocaleString()}</td><td
                class={'px-2 py-3 font-medium ' + (check.is_up ? 'text-success' : 'text-danger')}>{check.is_up ? 'Up' : 'Down'}</td
              ><td class="px-2 py-3 tabular-nums"
                >{monitor?.type === 'tcp' ? connectionStatus(check.is_up, check.error_code) : (check.status_code ?? '—')}</td
              ><td class="px-2 py-3 text-right tabular-nums">{check.response_time_ms ?? 0} ms</td></tr
            >{/each}</tbody
        >
      </table>
    </div>{/if}
  {#snippet footer()}<Button variant="secondary" onclick={() => (open = false)}>Close</Button>{/snippet}
</Modal>
