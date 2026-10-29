<script lang="ts">
  import type { HTMLButtonAttributes } from 'svelte/elements'
  import { cn } from '../lib/cn'
  let {
    variant = 'primary',
    class: className = '',
    children,
    ...rest
  }: HTMLButtonAttributes & { variant?: 'primary' | 'secondary' | 'ghost' | 'danger' } = $props()
  const styles = {
    primary: 'bg-primary text-primary-fg hover:bg-primary-hover',
    secondary: 'border border-border bg-transparent text-text hover:bg-surface-2',
    ghost: 'text-text-muted hover:bg-surface-2 hover:text-text',
    danger: 'text-danger hover:bg-danger-soft',
  }
</script>

<button
  type="button"
  class={cn(
    'inline-flex items-center justify-center gap-2 rounded-btn px-3.5 py-2 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-50',
    styles[variant],
    className,
  )}
  {...rest}>{@render children?.()}</button
>
