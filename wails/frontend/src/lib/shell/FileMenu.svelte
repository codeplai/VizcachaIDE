<script lang="ts">
  import { DropdownMenu } from 'bits-ui'
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import {
    activePath,
    closeActive,
    closeAll,
    newFile,
    openFileFromDialog,
    openFolder,
    openTabs,
    saveActiveAs,
    saveActiveFile,
    saveAll
  } from '../stores'
  import RecentMenu from './RecentMenu.svelte'

  interface Entry {
    label: string
    shortcut?: string
    run: () => void
    needs?: 'file' | 'tabs'
    separatorBefore?: boolean
  }

  const entries: Entry[] = [
    { label: 'file.new', shortcut: 'Ctrl+N', run: () => void newFile(bridge) },
    { label: 'file.open', shortcut: 'Ctrl+O', run: () => void openFileFromDialog(bridge) },
    { label: 'shell.openFolder', run: () => void openFolder(bridge) },
    {
      label: 'file.save',
      shortcut: 'Ctrl+S',
      run: () => void saveActiveFile(bridge),
      needs: 'file',
      separatorBefore: true
    },
    {
      label: 'file.saveAs',
      shortcut: 'Ctrl+Shift+S',
      run: () => void saveActiveAs(bridge),
      needs: 'file'
    },
    { label: 'file.saveAll', run: () => void saveAll(bridge), needs: 'tabs' },
    {
      label: 'file.close',
      shortcut: 'Ctrl+W',
      run: () => void closeActive(bridge),
      needs: 'file',
      separatorBefore: true
    },
    { label: 'file.closeAll', run: () => void closeAll(bridge), needs: 'tabs' }
  ]

  const disabled = (entry: Entry): boolean =>
    (entry.needs === 'file' && !$activePath) || (entry.needs === 'tabs' && $openTabs.length === 0)
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger class="file-trigger">
    {$t('file.menu')}
  </DropdownMenu.Trigger>
  <DropdownMenu.Portal>
    <DropdownMenu.Content class="menu" align="start" sideOffset={6}>
      {#each entries as entry (entry.label)}
        {#if entry.separatorBefore}
          <DropdownMenu.Separator class="file-separator" />
        {/if}
        <DropdownMenu.Item
          class="menu-item file-entry"
          disabled={disabled(entry)}
          onSelect={entry.run}
        >
          <span>{$t(entry.label)}</span>
          {#if entry.shortcut}<span class="shortcut">{entry.shortcut}</span>{/if}
        </DropdownMenu.Item>
        {#if entry.label === 'shell.openFolder'}
          <RecentMenu />
        {/if}
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Portal>
</DropdownMenu.Root>

<style>
  :global(.file-trigger) {
    border: 1px solid transparent;
    background: none;
    color: var(--muted);
    font: 600 14px var(--ui);
    padding: 7px 8px;
    border-radius: 8px;
    cursor: pointer;
    white-space: nowrap;
  }
  :global(.file-trigger:hover),
  :global(.file-trigger[data-state='open']) {
    color: var(--ink);
    background: var(--rail);
  }
  :global(.file-entry) {
    display: flex;
    justify-content: space-between;
    gap: 24px;
  }
  :global(.file-entry[data-disabled]) {
    opacity: 0.45;
    cursor: default;
  }
  .shortcut {
    color: var(--muted);
    font: 600 12px var(--mono);
  }
  :global(.file-separator) {
    height: 1px;
    margin: 4px 0;
    background: var(--line);
  }
</style>
