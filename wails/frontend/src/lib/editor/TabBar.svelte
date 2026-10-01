<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { openFile, requestCloseTab } from '../stores'
  import { tabItems } from './tabs'
</script>

<div class="tabs" role="tablist" aria-label={$t('editor.tab.list')}>
  {#each $tabItems as tab (tab.path)}
    <div class="tab" class:on={tab.active} role="presentation">
      <button
        type="button"
        role="tab"
        class="name"
        aria-selected={tab.active}
        onclick={() => openFile(bridge, tab.path)}
        onauxclick={(event) => event.button === 1 && requestCloseTab(bridge, tab.path)}
      >
        {tab.name}
        {#if tab.modified}<span class="dot" title={$t('editor.tab.modified')}></span>{/if}
      </button>
      <button
        type="button"
        class="close"
        aria-label={$t('editor.tab.close', { values: { file: tab.name } })}
        onclick={() => requestCloseTab(bridge, tab.path)}>×</button
      >
    </div>
  {/each}
</div>

<style>
  .tabs {
    display: flex;
    background: var(--chrome);
    border-bottom: 1px solid var(--line);
    overflow: hidden;
  }
  .tab {
    display: flex;
    align-items: center;
    border-right: 1px solid var(--line);
    color: var(--muted);
  }
  .tab.on {
    background: var(--win);
    color: var(--ink);
    font-weight: 600;
    box-shadow: inset 0 2px 0 var(--go);
  }
  .name {
    border: 0;
    background: none;
    color: inherit;
    font: inherit;
    font-size: 13px;
    padding: 0 6px 0 16px;
    height: 100%;
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
  }
  .dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--go);
  }
  .close {
    border: 0;
    background: none;
    color: var(--muted);
    font-size: 16px;
    line-height: 1;
    width: 22px;
    height: 22px;
    margin-right: 8px;
    border-radius: 6px;
    cursor: pointer;
    opacity: 0;
  }
  .tab:hover .close,
  .tab.on .close,
  .close:focus-visible {
    opacity: 1;
  }
  .close:hover {
    background: var(--go-soft);
    color: var(--ink);
  }
</style>
