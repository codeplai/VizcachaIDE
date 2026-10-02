<script lang="ts">
  import { ContextMenu } from 'bits-ui'
  import type { Snippet } from 'svelte'
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import {
    copyPath,
    deleteEntry,
    fileTree,
    folderOf,
    revealEntry,
    startCreate,
    startRename
  } from '../stores'

  interface Props {
    /** The row this menu belongs to; null for the empty area of the tree (the root folder). */
    path: string | null
    /** Receives the props the trigger element must spread (`{...props}`). */
    row: Snippet<[Record<string, unknown>]>
  }
  let { path, row }: Props = $props()

  const target = $derived(path ?? $fileTree?.path ?? null)
  /** The root folder cannot be renamed or deleted. */
  const canChange = $derived(path !== null && path !== $fileTree?.path)
  const isMac = typeof navigator !== 'undefined' && /mac/i.test(navigator.platform)
</script>

<ContextMenu.Root>
  <ContextMenu.Trigger>
    {#snippet child({ props })}
      {@render row(props)}
    {/snippet}
  </ContextMenu.Trigger>
  <ContextMenu.Portal>
    <ContextMenu.Content class="menu">
      <ContextMenu.Item class="menu-item" onSelect={() => startCreate('file', folderOf(path))}>
        {$t('tree.newFileHere')}
      </ContextMenu.Item>
      <ContextMenu.Item class="menu-item" onSelect={() => startCreate('folder', folderOf(path))}>
        {$t('tree.newFolder')}
      </ContextMenu.Item>
      {#if canChange && path}
        <ContextMenu.Separator class="menu-sep" />
        <ContextMenu.Item class="menu-item tree-menu-row" onSelect={() => startRename(path)}>
          {$t('tree.rename')}<kbd>F2</kbd>
        </ContextMenu.Item>
        <ContextMenu.Item
          class="menu-item tree-menu-row"
          onSelect={() => void deleteEntry(bridge, path)}
        >
          {$t('tree.delete')}<kbd>{$t('tree.deleteKey')}</kbd>
        </ContextMenu.Item>
      {/if}
      {#if target}
        <ContextMenu.Separator class="menu-sep" />
        <ContextMenu.Item class="menu-item" onSelect={() => void revealEntry(bridge, target)}>
          {$t(isMac ? 'tree.revealMac' : 'tree.reveal')}
        </ContextMenu.Item>
        <ContextMenu.Item class="menu-item" onSelect={() => void copyPath(target)}>
          {$t('tree.copyPath')}
        </ContextMenu.Item>
      {/if}
    </ContextMenu.Content>
  </ContextMenu.Portal>
</ContextMenu.Root>

<style>
  :global(.tree-menu-row) {
    display: flex !important;
    justify-content: space-between;
    gap: 24px;
  }
  :global(.tree-menu-row kbd) {
    font: 12px var(--ui);
    color: var(--muted);
  }
  :global(.menu-sep) {
    height: 1px;
    background: var(--line);
    margin: 4px 2px;
  }
</style>
