<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import {
    baseName,
    goToLocation,
    parentName,
    referenceCount,
    referenceGroups,
    referencesState
  } from '../stores'
</script>

<div class="list">
  {#if !$referencesState}
    <p class="empty">{$t('references.hint')}</p>
  {:else if $referencesState.loading}
    <p class="empty" role="status">{$t('references.searching')}</p>
  {:else if $referenceCount === 0}
    <p class="empty" role="status">
      {$t('references.empty', { values: { symbol: $referencesState.symbol } })}
    </p>
  {:else}
    <h2 class="summary" role="status">
      {$t('references.summary', {
        values: {
          symbol: $referencesState.symbol,
          count: $referenceCount,
          files: $referenceGroups.length
        }
      })}
    </h2>
    <ul aria-label={$t('a11y.referencesList')}>
      {#each $referenceGroups as group (group.file)}
        <li class="group">
          <div class="file">
            <span class="name">{baseName(group.file)}</span>
            <span class="folder">{parentName(group.file)}</span>
            <span class="n">{group.items.length}</span>
          </div>
          <ul>
            {#each group.items as item (`${item.range.start.line}:${item.range.start.column}`)}
              <li>
                <button
                  type="button"
                  class="ref"
                  onclick={() => goToLocation(bridge, item.range.start)}
                >
                  <span class="loc">{item.range.start.line}:{item.range.start.column}</span>
                  <span class="text">{item.preview}</span>
                </button>
              </li>
            {/each}
          </ul>
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .list {
    padding: 8px 14px;
    overflow: auto;
    height: 100%;
  }
  ul {
    margin: 0;
    padding: 0;
  }
  li {
    list-style: none;
  }
  .summary {
    margin: 0 0 6px;
    font-size: 12.5px;
    font-weight: 600;
    color: var(--muted);
  }
  .group + .group {
    margin-top: 6px;
  }
  .file {
    display: flex;
    gap: 8px;
    align-items: baseline;
    padding: 2px 4px;
    font-size: 13px;
  }
  .name {
    font-weight: 700;
  }
  .folder {
    color: var(--muted);
    font-size: 12px;
  }
  .n {
    margin-left: auto;
    font-size: 11px;
    color: var(--muted);
  }
  .ref {
    width: 100%;
    border: 0;
    background: none;
    cursor: pointer;
    display: flex;
    gap: 12px;
    padding: 3px 4px 3px 16px;
    font-size: 13px;
    text-align: left;
    border-radius: 6px;
  }
  .ref:hover {
    background: var(--go-soft);
  }
  .loc {
    font: 600 12.5px var(--mono);
    color: var(--go);
    flex: none;
    min-width: 52px;
  }
  .text {
    font: 400 12.5px var(--mono);
    white-space: pre;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .empty {
    margin: 0;
    color: var(--muted);
    font-size: 13px;
  }
</style>
