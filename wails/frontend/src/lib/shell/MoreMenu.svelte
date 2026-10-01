<script lang="ts">
  import { DropdownMenu } from 'bits-ui'
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { openDialog, openFolder } from '../stores'

  const entries = [
    { label: 'shell.openFolder', run: () => void openFolder(bridge) },
    { label: 'shell.goModules', run: () => openDialog.set('modules') },
    { label: 'panels.settings', run: () => openDialog.set('settings') },
    { label: 'shell.about', run: () => openDialog.set('about') }
  ]
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
