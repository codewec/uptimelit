<script lang="ts">
  let { name, size = 16, class: className = '' }: { name: string; size?: number; class?: string } = $props()
  const paths: Record<string, string> = {
    plus: 'M12 5v14M5 12h14',
    search: 'm21 21-4.4-4.4M11 3a8 8 0 1 0 0 16 8 8 0 0 0 0-16',
    monitor: 'M4 4h16a1 1 0 0 1 1 1v12H3V5a1 1 0 0 1 1-1Zm4 17h8m-4-4v4',
    bell: 'M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9M10 21h4',
    user: 'M12 3a4 4 0 1 0 0 8 4 4 0 0 0 0-8M4 21a8 8 0 0 1 16 0',
    shield: 'm12 3-8 4v5c0 5 3.4 8.3 8 9 4.6-.7 8-4 8-9V7l-8-4Zm-3 9 2 2 4-4',
    logout: 'M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4m7 12 5-5-5-5m5 5H9',
    theme: 'M12 3a9 9 0 1 0 0 18 9 9 0 0 0 0-18Zm0 0c-3 4-3 14 0 18',
    close: 'm6 6 12 12M6 18 18 6',
    check: 'm5 12 4 4L19 6',
    more: 'M5 12h.01M12 12h.01M19 12h.01',
    history: 'M3 11a9 9 0 1 1 2.4 7M3 4v7h7m2-4v5l3 2',
    edit: 'm16 3 5 5L8 21H3v-5L16 3Zm-2 2 5 5',
    pause: 'M8 5v14M16 5v14',
    play: 'm7 4 14 8-14 8V4Z',
    trash: 'M3 6h18M9 6V3h6v3M5 6l1 15h12l1-15M10 10v7m4-7v7',
    activity: 'M2 12h5l3-8 4 16 3-8h5',
    arrow: 'm15 18-6-6 6-6',
    menu: 'M4 6h16M4 12h16M4 18h16',
    send: 'm22 2-7 20-4-9-9-4 20-7Zm0 0L11 13',
  }
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width="1.8"
  stroke-linecap="round"
  stroke-linejoin="round"
  class={className}
  aria-hidden="true"><path d={paths[name] || paths.monitor} /></svg
>
