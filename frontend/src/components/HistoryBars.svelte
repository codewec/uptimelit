<script lang="ts">
  import type { Check } from '../lib/api'
  let { checks = [], paused = false }: { checks?: Check[]; paused?: boolean } = $props()
  const slots = $derived([...Array(Math.max(0, 50 - checks.length)).fill(null), ...checks.slice(-50)] as (Check | null)[])
</script>

<div
  class="flex h-7 items-stretch gap-[3px]"
  role="img"
  aria-label={checks.length ? `Last ${Math.min(checks.length, 50)} checks` : 'No check history yet'}
>
  {#each slots as check, index (index)}<span
      class={'min-w-0 flex-1 rounded-[3px] transition-colors ' +
        (!check ? 'bg-border-soft' : paused ? 'bg-text-subtle/40' : check.is_up ? 'bg-success/75' : 'bg-danger/80')}
      title={check
        ? `${new Date(check.checked_at).toLocaleString()} · ${check.is_up ? 'UP' : 'DOWN'}${check.status_code ? ' · HTTP ' + check.status_code : ''}${check.response_time_ms != null ? ' · ' + check.response_time_ms + ' ms' : ''}`
        : 'No check'}
    ></span>{/each}
</div>
