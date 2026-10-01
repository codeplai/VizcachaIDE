<script lang="ts">
  import { t } from '../i18n'
  import { currentLine, firstProblem, problemCount, uiMode } from '../stores'
  import CallStackPanel from './CallStackPanel.svelte'
  import ErrorCard from './ErrorCard.svelte'
  import TipsCard from './TipsCard.svelte'
  import VariablesPanel from './VariablesPanel.svelte'
</script>

<aside class="guide">
  <div class="guide-h">
    <h2>{$t('panels.assistant')}</h2>
    {#if $uiMode === 'write'}
      <span class="chip idle">{$t('assistant.chipReady')}</span>
    {:else if $uiMode === 'error'}
      <span class="chip err">
        {$t('assistant.chipProblems', { values: { count: $problemCount } })}
      </span>
    {:else}
      <span class="chip run">
        {$t('assistant.chipPaused', { values: { line: $currentLine ?? 0 } })}
      </span>
    {/if}
  </div>

  {#if $uiMode === 'write'}
    <TipsCard />
  {:else if $uiMode === 'error' && $firstProblem}
    <ErrorCard item={$firstProblem} />
  {:else if $uiMode === 'debug'}
    <VariablesPanel />
    <CallStackPanel />
  {/if}
</aside>

<style>
  .guide {
    height: 100%;
    background: var(--chrome);
    padding: 14px;
    display: grid;
    align-content: start;
    gap: 14px;
    overflow: auto;
  }
  .guide-h {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  h2 {
    margin: 0;
    font-size: 14px;
    font-weight: 700;
  }
  .chip {
    font: 600 11px var(--ui);
    border-radius: 999px;
    padding: 2px 9px;
  }
  .chip.err {
    background: var(--err-soft);
    color: var(--err);
  }
  .chip.run {
    background: var(--sand-soft);
    color: var(--ink);
  }
  .chip.idle {
    background: var(--go-soft);
    color: var(--go);
  }
</style>
