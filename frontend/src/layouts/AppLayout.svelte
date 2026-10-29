<script lang="ts">
  import type { Snippet } from 'svelte'
  import { DropdownMenu } from 'bits-ui'
  import { api } from '../lib/api'
  import { routePath, push, replace } from '../lib/router'
  import { user, searchQuery, theme, systemInfo } from '../lib/stores'
  import { themeLabel } from '../lib/themes'
  import { toastError } from '../lib/toasts'
  import ThemeItems from '../components/ThemeItems.svelte'
  import Icon from '../components/Icon.svelte'
  import logo from '../assets/uptimelit.svg'
  let { children, showSearch = false, onNewMonitor }: { children?: Snippet; showSearch?: boolean; onNewMonitor?: () => void } = $props()
  let signingOut = $state(false)
  async function logout() {
    signingOut = true
    try {
      await api.logout()
      user.set(null)
      searchQuery.set('')
      replace('/login')
    } catch (e) {
      toastError(e, 'Could not sign out')
    } finally {
      signingOut = false
    }
  }
</script>

{#snippet newMonitorButton()}
  {#if onNewMonitor}
    <button
      type="button"
      class="flex h-8 w-8 items-center justify-center rounded-btn text-text-muted hover:bg-surface-2 hover:text-text"
      onclick={onNewMonitor}
      title="New monitor"
      aria-label="New monitor"><Icon name="plus" /></button
    >
  {/if}
{/snippet}

<div class="flex min-h-dvh flex-col">
  <header class="sticky top-0 z-40 border-b border-border-soft bg-bg-elevated/80 backdrop-blur-md">
    <div class="mx-auto flex max-w-6xl items-center gap-2 px-4 py-2.5 sm:gap-3">
      <a href="/" class="flex shrink-0 items-center gap-2 text-sm font-semibold tracking-tight text-text"
        ><img src={logo} alt="" class="h-6 w-6" width="24" height="24" /><span>UptimeLit</span></a
      >
      <div class="min-w-0 flex-1">
        {#if showSearch}<label class="relative block"
            ><span class="pointer-events-none absolute inset-y-0 left-2.5 flex items-center text-text-subtle"
              ><Icon name="search" size={14} /></span
            ><input
              aria-label="Search"
              class="h-8 w-full rounded-btn border border-border-soft bg-surface/70 pl-8 pr-3 text-xs outline-none placeholder:text-text-subtle focus:border-primary"
              bind:value={$searchQuery}
              placeholder="Search…"
            /></label
          >{/if}
      </div>
      <div class="ml-auto hidden items-center gap-1 sm:flex">
        {@render newMonitorButton()}
        <DropdownMenu.Root
          ><DropdownMenu.Trigger
            class="flex h-8 w-8 items-center justify-center rounded-btn text-text-muted hover:bg-surface-2 hover:text-text"
            title={'Theme: ' + themeLabel($theme)}
            aria-label="Choose theme"><Icon name="theme" /></DropdownMenu.Trigger
          ><DropdownMenu.Portal
            ><DropdownMenu.Content
              class="z-50 min-w-44 rounded-xl border border-border bg-surface p-1 shadow-xl outline-none"
              sideOffset={6}
              align="end"><ThemeItems /></DropdownMenu.Content
            ></DropdownMenu.Portal
          ></DropdownMenu.Root
        >
        <a
          href="/notifiers"
          aria-current={$routePath === '/notifiers' ? 'page' : undefined}
          class={'flex h-8 w-8 items-center justify-center rounded-btn hover:bg-surface-2 hover:text-text ' +
            ($routePath === '/notifiers' ? 'bg-surface-2 text-primary' : 'text-text-muted')}
          title="Notifiers"
          aria-label="Notifiers"><Icon name="bell" /></a
        >
        {#if $user?.role === 'admin'}<a
            href="/admin"
            class="flex h-8 w-8 items-center justify-center rounded-btn text-text-muted hover:bg-surface-2 hover:text-text"
            title="Administration"
            aria-label="Administration"><Icon name="shield" /></a
          >{/if}
        <a
          href="/profile"
          class="flex h-8 w-8 items-center justify-center rounded-btn text-text-muted hover:bg-surface-2 hover:text-text"
          title="Profile"
          aria-label="Profile"><Icon name="user" /></a
        >
        <button
          type="button"
          class="flex h-8 w-8 items-center justify-center rounded-btn text-text-muted hover:bg-surface-2 hover:text-text disabled:opacity-50"
          onclick={logout}
          disabled={signingOut}
          title="Sign out"
          aria-label="Sign out"><Icon name="logout" /></button
        >
      </div>
      <div class="flex items-center gap-1 sm:hidden">
        {@render newMonitorButton()}
        <DropdownMenu.Root
          ><DropdownMenu.Trigger
            aria-label="Menu"
            class="inline-flex h-8 w-8 items-center justify-center rounded-btn border border-border text-text"
            ><Icon name="menu" /></DropdownMenu.Trigger
          ><DropdownMenu.Portal
            ><DropdownMenu.Content
              class="z-50 max-h-[80vh] min-w-48 overflow-auto rounded-xl border border-border bg-surface p-1 shadow-xl outline-none"
              sideOffset={6}
              align="end"
            >
              {#each [{ label: 'Notifiers', path: '/notifiers' }, { label: 'Profile', path: '/profile' }, ...($user?.role === 'admin' ? [{ label: 'Administration', path: '/admin' }] : [])] as item}<DropdownMenu.Item
                  class="cursor-pointer rounded-lg px-2.5 py-2 text-sm outline-none data-highlighted:bg-surface-2"
                  onSelect={() => {
                    searchQuery.set('')
                    push(item.path)
                  }}>{item.label}</DropdownMenu.Item
                >{/each}
              <DropdownMenu.Separator class="my-1 h-px bg-border-soft" /><ThemeItems /><DropdownMenu.Separator
                class="my-1 h-px bg-border-soft"
              /><DropdownMenu.Item
                disabled={signingOut}
                class="cursor-pointer rounded-lg px-2.5 py-2 text-sm text-danger outline-none data-highlighted:bg-surface-2"
                onSelect={logout}>Sign out</DropdownMenu.Item
              >
            </DropdownMenu.Content></DropdownMenu.Portal
          ></DropdownMenu.Root
        >
      </div>
    </div>
  </header>
  <main class="mx-auto w-full max-w-6xl flex-1 px-4 py-6">{@render children?.()}</main>
  <footer class="border-t border-border-soft bg-bg-elevated/60">
    <div class="mx-auto flex max-w-6xl items-center justify-center gap-2 px-4 py-3 text-xs text-text-subtle">
      <img src={logo} alt="" class="h-3.5 w-3.5 opacity-70" /><span
        >UptimeLit {$systemInfo?.version ?? 'dev'} · © {new Date().getFullYear()} ·</span
      ><a href="https://github.com/codewec/uptimelit" target="_blank" rel="noreferrer" class="opacity-60 hover:opacity-100">GitHub</a>
    </div>
  </footer>
</div>
