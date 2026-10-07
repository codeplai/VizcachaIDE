<script lang="ts">
  import { t } from '../i18n'
  import { capabilities, outputTab, problemCount, referenceCount } from '../stores'
  import ConsolePanel from './ConsolePanel.svelte'
  import OutputPanel from './OutputPanel.svelte'
  import ProblemsPanel from './ProblemsPanel.svelte'
  import ReferencesPanel from './ReferencesPanel.svelte'
  import TerminalPanel from './TerminalPanel.svelte'

  // The terminal is created the first time its tab is shown and then stays alive (hidden), so
  // switching tabs keeps its shells and what they printed.
  let terminalSeen = $state(false)
  $effect(() => {
    if ($outputTab === 'terminal') terminalSeen = true
  })
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
    <button
      type="button"
      role="tab"
      class="otab"
      class:on={$outputTab === 'references'}
      aria-selected={$outputTab === 'references'}
      onclick={() => outputTab.set('references')}
    >
      {$t('panels.references')}
      {#if $referenceCount > 0}<span class="count info">{$referenceCount}</span>{/if}
    </button>
    {#if $capabilities?.console}
      <button
        type="button"
        role="tab"
        class="otab"
        class:on={$outputTab === 'console'}
        aria-selected={$outputTab === 'console'}
        onclick={() => outputTab.set('console')}
      >
        {$t('console.title')}
      </button>
    {/if}
    <button
      type="button"
      role="tab"
      class="otab"
      class:on={$outputTab === 'terminal'}
      aria-selected={$outputTab === 'terminal'}
      onclick={() => outputTab.set('terminal')}
    >
      {$t('panels.terminal')}
    </button>
  </div>
  <div class="body">
    {#if $outputTab === 'output'}<OutputPanel
      />{:else if $outputTab === 'console' && $capabilities?.console}<ConsolePanel
      />{:else if $outputTab === 'problems'}<ProblemsPanel
      />{:else if $outputTab === 'references'}<ReferencesPanel />{/if}
    {#if terminalSeen}
      <div class="terminal-slot" hidden={$outputTab !== 'terminal'}>
        <TerminalPanel visible={$outputTab === 'terminal'} />
      </div>
    {/if}
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
  .terminal-slot {
    height: 100%;
  }
  .terminal-slot[hidden] {
    display: none;
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
  .count.info {
    background: var(--go);
  }
</style>
