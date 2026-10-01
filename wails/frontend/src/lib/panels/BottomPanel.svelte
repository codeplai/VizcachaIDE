<script lang="ts">
  import { t } from '../i18n'
  import { outputTab, problemCount } from '../stores'
  import OutputPanel from './OutputPanel.svelte'
  import ProblemsPanel from './ProblemsPanel.svelte'
</script>

<section class="output">
  <div class="otabs" role="tablist">
    <button
      type="button"
      role="tab"
      class="otab"
      class:on={$outputTab === 'output'}
      aria-selected={$outputTab === 'output'}
      onclick={() => outputTab.set('output')}
    >
      {$t('panels.output')}
    </button>
    <button
      type="button"
      role="tab"
      class="otab"
      class:on={$outputTab === 'problems'}
      aria-selected={$outputTab === 'problems'}
      onclick={() => outputTab.set('problems')}
    >
      {$t('panels.problems')}
      {#if $problemCount > 0}<span class="count">{$problemCount}</span>{/if}
    </button>
  </div>
  <div class="body">
    {#if $outputTab === 'output'}<OutputPanel />{:else}<ProblemsPanel />{/if}
  </div>
</section>

<style>
  .output {
    height: 100%;
    display: grid;
    grid-template-rows: 32px 1fr;
    min-width: 0;
    min-height: 0;
    background: var(--win);
  }
  .body {
    min-height: 0;
  }
  .otabs {
    display: flex;
    gap: 2px;
    padding: 0 10px;
    background: var(--chrome);
    border-bottom: 1px solid var(--line);
    align-items: end;
  }
  .otab {
    border: 0;
    background: none;
    cursor: pointer;
    padding: 6px 12px;
    font-size: 12.5px;
    font-weight: 600;
    color: var(--muted);
    border-radius: 7px 7px 0 0;
    display: flex;
    gap: 6px;
    align-items: center;
  }
  .otab.on {
    background: var(--win);
    color: var(--ink);
    box-shadow: inset 0 2px 0 var(--go);
  }
  .count {
    font-size: 11px;
    background: var(--err);
    color: var(--win);
    border-radius: 999px;
    padding: 0 6px;
  }
</style>
