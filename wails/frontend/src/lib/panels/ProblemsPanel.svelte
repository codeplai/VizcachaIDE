<script lang="ts">
  import { t } from '../i18n'
  import { problems } from '../stores'
</script>

<div class="list">
  {#each $problems as item (item.diagnostic.rawText)}
    <div class="problem">
      <span class="loc">
        {item.diagnostic.location?.file.split('/').pop()}:{item.diagnostic.location?.line}
      </span>
      <span class="msg">{item.explanation?.title ?? item.diagnostic.message}</span>
    </div>
  {:else}
    <p class="empty">{$t('empty.problems')}</p>
  {/each}
</div>

<style>
  .list {
    padding: 8px 14px;
    overflow: auto;
    height: 100%;
  }
  .problem {
    display: flex;
    gap: 12px;
    padding: 3px 0;
    font-size: 13px;
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
