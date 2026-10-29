<script lang="ts">
  import Modal from './Modal.svelte'
  import Button from './Button.svelte'
  import { customTheme, customThemeEditorOpen, setCustomTheme, deleteCustomTheme, hasCustomTheme } from '../lib/stores'
  import type { CustomThemeConfig } from '../lib/themes'
  import { field, label } from '../lib/styles'
  let draft = $state<CustomThemeConfig>({ ...$customTheme })
  const colors = [
    ['bg', 'Background'],
    ['surface', 'Tiles'],
    ['text', 'Text'],
    ['primary', 'Primary'],
    ['accent', 'Accent'],
    ['danger', 'Danger'],
    ['success', 'Success'],
  ] as const
  $effect(() => {
    if ($customThemeEditorOpen) draft = { ...$customTheme }
  })
  function save(event: SubmitEvent) {
    event.preventDefault()
    setCustomTheme(draft)
    customThemeEditorOpen.set(false)
  }
</script>

<Modal
  bind:open={$customThemeEditorOpen}
  title="Custom theme"
  description="Make this workspace feel like home. Preferences are saved in this browser."
>
  <form onsubmit={save} class="space-y-4">
    <label class="block"
      ><span class={label}>Color scheme</span><select class={field} bind:value={draft.scheme}
        ><option value="light">Light</option><option value="dark">Dark</option></select
      ></label
    >
    <div class="grid grid-cols-2 gap-3">
      {#each colors as [key, name]}<label class="flex items-center gap-2 rounded-btn border border-border-soft p-2 text-xs"
          ><input type="color" class="h-7 w-7 cursor-pointer bg-transparent" bind:value={draft[key]} />{name}</label
        >{/each}
    </div>
    <label class="block"
      ><span class={label}>Background image URL</span><input
        type="url"
        class={field}
        bind:value={draft.backgroundImage}
        placeholder="https://…"
      /></label
    >
    <div class="flex justify-end gap-2">
      {#if $hasCustomTheme}<Button
          variant="danger"
          onclick={() => {
            deleteCustomTheme()
            customThemeEditorOpen.set(false)
          }}>Reset</Button
        >{/if}<Button variant="secondary" onclick={() => customThemeEditorOpen.set(false)}>Cancel</Button><Button type="submit"
        >Save theme</Button
      >
    </div>
  </form>
</Modal>
