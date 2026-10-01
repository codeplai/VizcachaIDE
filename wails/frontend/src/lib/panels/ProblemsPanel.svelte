<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { baseName, goToLocation, problemKey, problems } from '../stores'
</script>

<div class="list">
  {#if $problems.length > 0}
    <ul aria-label={$t('a11y.problemsList')}>
      {#each $problems as item (problemKey(item.diagnostic))}
        {@const location = item.diagnostic.location}
        <li>
          <button
            type="button"
            class="problem"
            disabled={!location}
            onclick={() => location && goToLocation(bridge, location)}
          >
            {#if location}
              <span class="loc">{baseName(location.file)}:{location.line}:{location.column}</span>
            {/if}
            <span class="msg">{item.explanation?.title ?? item.diagnostic.message}</span>
          </button>
        </li>
      {/each}
    </ul>
  {:else}
    <p class="empty">{$t('empty.problems')}</p>
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
  .problem {
    width: 100%;
    border: 0;
    background: none;
    cursor: pointer;
    display: flex;
    gap: 12px;
    padding: 3px 4px;
    font-size: 13px;
    text-align: left;
    border-radius: 6px;
  }
  .problem:hover:not(:disabled) {
    background: var(--err-soft);
  }
  .loc {
    font: 600 12.5px var(--mono);
    color: var(--err);
  }
  .empty {
    margin: 0;
    color: var(--muted);
    font-size: 13px;
  }
</style>
