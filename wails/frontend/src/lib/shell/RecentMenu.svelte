<script lang="ts">
  import { DropdownMenu } from 'bits-ui'
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { baseName, clearRecentFiles, openRecentFile, parentName, settings } from '../stores'

  const recent = $derived($settings?.recentFiles ?? [])
</script>

<DropdownMenu.Sub>
  <DropdownMenu.SubTrigger class="menu-item recent-trigger">
    {$t('recent.title')}
    <span aria-hidden="true">▸</span>
  </DropdownMenu.SubTrigger>
  <DropdownMenu.Portal>
    <DropdownMenu.SubContent class="menu" sideOffset={4}>
      {#each recent as path (path)}
        <DropdownMenu.Item
          class="menu-item recent-entry"
          title={path}
          onSelect={() => void openRecentFile(bridge, path)}
        >
          {baseName(path)}
          <span class="recent-folder">{parentName(path)}</span>
        </DropdownMenu.Item>
      {:else}
        <div class="recent-empty">{$t('recent.empty')}</div>
      {/each}
      {#if recent.length > 0}
        <DropdownMenu.Separator class="recent-separator" />
        <DropdownMenu.Item class="menu-item" onSelect={() => void clearRecentFiles(bridge)}>
          {$t('recent.clear')}
        </DropdownMenu.Item>
      {/if}
    </DropdownMenu.SubContent>
  </DropdownMenu.Portal>
</DropdownMenu.Sub>

<style>
  :global(.recent-trigger) {
    display: flex;
    justify-content: space-between;
    gap: 12px;
  }
  :global(.recent-entry) {
    display: flex;
    gap: 10px;
    align-items: baseline;
  }
  .recent-folder {
    color: var(--muted);
    font-size: 12px;
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .recent-empty {
    padding: 7px 12px;
    color: var(--muted);
    font-size: 14px;
  }
  :global(.recent-separator) {
    height: 1px;
    margin: 4px 0;
    background: var(--line);
  }
</style>
