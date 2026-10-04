<script lang="ts">
  import { t } from '../i18n'
  import { selectSidebarView, sidebarOpen, sidebarView, type SidebarView } from '../stores'

  const items: { view: SidebarView; icon: string; label: string }[] = [
    { view: 'files', icon: '▤', label: 'panels.files' },
    { view: 'outline', icon: '☰', label: 'panels.outline' },
    { view: 'search', icon: '⌕', label: 'panels.search' }
  ]
</script>

<nav class="rail" aria-label={$t('a11y.sidebar')}>
  {#each items as item (item.view)}
    <button
      type="button"
      class:on={$sidebarOpen && $sidebarView === item.view}
      aria-pressed={$sidebarOpen && $sidebarView === item.view}
      aria-label={$t(item.label)}
      title={$t(item.label)}
      onclick={() => selectSidebarView(item.view)}
    >
      {item.icon}
    </button>
  {/each}
</nav>

<style>
  .rail {
    width: 48px;
    flex: none;
    background: var(--rail);
    display: grid;
    align-content: start;
    justify-items: center;
    gap: 6px;
    padding-top: 10px;
    border-right: 1px solid var(--line);
  }
  button {
    width: 32px;
    height: 32px;
    border: 0;
    border-radius: 8px;
    background: none;
    display: grid;
    place-items: center;
    color: var(--muted);
    font-size: 15px;
    cursor: pointer;
  }
  button.on {
    background: var(--win);
    color: var(--go);
    box-shadow: inset 3px 0 0 var(--go);
  }
</style>
