<script lang="ts">
  import { DropdownMenu } from 'bits-ui'
  import { t } from '../i18n'
  import { activeCodeLanguage, capabilities, openDialog } from '../stores'

  // The package manager of the open file's language; languages without one have no entry.
  const hasPackages = $derived(($capabilities?.packageActions.length ?? 0) > 0)
  const entries = $derived([
    ...(hasPackages
      ? [
          {
            label: $activeCodeLanguage === 'go' ? 'shell.goModules' : 'packages.title',
            run: () => openDialog.set('packages')
          }
        ]
      : []),
    { label: 'panels.settings', run: () => openDialog.set('settings') },
    { label: 'shell.about', run: () => openDialog.set('about') }
  ])
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger class="more-trigger">
    {$t('actions.more')}
  </DropdownMenu.Trigger>
  <DropdownMenu.Portal>
    <DropdownMenu.Content class="menu" align="end" sideOffset={6}>
      {#each entries as entry (entry.label)}
        <DropdownMenu.Item class="menu-item" onSelect={entry.run}>
          {$t(entry.label)}
        </DropdownMenu.Item>
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Portal>
</DropdownMenu.Root>

<style>
  :global(.more-trigger) {
    border: 1px solid transparent;
    background: none;
    color: var(--muted);
    font: 600 14px var(--ui);
    padding: 7px 8px;
    border-radius: 8px;
    cursor: pointer;
    white-space: nowrap;
  }
  :global(.more-trigger:hover),
  :global(.more-trigger[data-state='open']) {
    color: var(--ink);
    background: var(--rail);
  }
</style>
