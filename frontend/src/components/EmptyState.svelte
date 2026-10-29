<script lang="ts">
  import type { Snippet } from 'svelte'
  import Icon from './Icon.svelte'
  let {
    icon = 'monitor',
    title,
    description,
    children,
  }: { icon?: string; title: string; description: string; children?: Snippet } = $props()
</script>

<div class="rounded-card border border-dashed border-border bg-surface/60 px-5 py-16 text-center">
  <span class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-xl border border-border-soft bg-bg-elevated text-text-muted"
    ><Icon name={icon} size={23} /></span
  >
  <h2 class="text-sm font-semibold text-text">{title}</h2>
  <p class="mx-auto mt-2 max-w-sm text-sm leading-relaxed text-text-muted">{description}</p>
  <div class="mt-5">{@render children?.()}</div>
</div>
