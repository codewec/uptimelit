<script lang="ts">
  import type { Snippet } from 'svelte'
  import logo from '../assets/uptimelit.svg'
  let { children }: { children?: Snippet } = $props()
</script>

<div class="flex min-h-dvh flex-col p-4">
  <div class="flex flex-1 items-center justify-center"><div class="w-full max-w-sm">{@render children?.()}</div></div>
  <footer class="flex items-center justify-center gap-2 pt-4 text-xs text-text-subtle">
    <img src={logo} alt="" class="h-3.5 w-3.5 opacity-70" /><span>UptimeLit · © {new Date().getFullYear()} ·</span><a
      href="https://github.com/codewec/uptimelit"
      target="_blank"
      rel="noreferrer"
      class="opacity-60 hover:opacity-100">GitHub</a
    >
  </footer>
</div>
