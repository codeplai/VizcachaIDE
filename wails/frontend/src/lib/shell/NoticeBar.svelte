<script lang="ts">
  import { t } from '../i18n'
  import { dismissNotice, notice } from '../stores'
</script>

{#if $notice}
  {@const current = $notice}
  <div
    class="notice"
    class:info={current.tone === 'info'}
    role={current.tone === 'info' ? 'status' : 'alert'}
  >
    <div class="text">
      <span>{$t(current.messageKey, { values: current.values })}</span>
      {#if current.detail}<code>{current.detail}</code>{/if}
    </div>
    {#each current.actions as action (action.labelKey)}
      <button type="button" class="act" onclick={action.run}>{$t(action.labelKey)}</button>
    {/each}
    <button type="button" class="close" aria-label={$t('shell.dismiss')} onclick={dismissNotice}>
      ×
    </button>
  </div>
{/if}

<style>
  .notice {
    display: flex;
    gap: 10px;
    align-items: center;
    padding: 8px 14px;
    background: var(--err-soft);
    border-top: 1px solid var(--line);
    font-size: 13.5px;
  }
  .notice.info {
    background: var(--go-soft);
  }
  .text {
    display: grid;
    gap: 4px;
    flex: 1;
  }
  code {
    font: 400 12.5px var(--mono);
    color: var(--muted);
    overflow-wrap: anywhere;
  }
  .act {
    border: 1px solid var(--line);
    background: var(--win);
    border-radius: 7px;
    padding: 4px 10px;
    font: 600 12.5px var(--ui);
    cursor: pointer;
  }
  .close {
    border: 0;
    background: none;
    font-size: 18px;
    line-height: 1;
    cursor: pointer;
    color: var(--muted);
  }
</style>
