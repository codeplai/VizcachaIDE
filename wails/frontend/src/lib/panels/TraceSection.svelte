<script lang="ts">
  import { t } from '../i18n'
  import { debugTab, type DebugTab } from '../stores'
  import CallStackPanel from './CallStackPanel.svelte'
  import GoroutinesPanel from './GoroutinesPanel.svelte'

  const tabs: { id: DebugTab; label: string }[] = [
    { id: 'stack', label: 'panels.callStack' },
    { id: 'goroutines', label: 'panels.goroutines' }
  ]
</script>

<section class="block">
  <div class="tabs" role="tablist">
    {#each tabs as tab (tab.id)}
      <button
        type="button"
        role="tab"
        class="tab"
        class:on={$debugTab === tab.id}
        aria-selected={$debugTab === tab.id}
        onclick={() => debugTab.set(tab.id)}
      >
        {$t(tab.label)}
      </button>
    {/each}
  </div>
  <div role="tabpanel">
    {#if $debugTab === 'stack'}<CallStackPanel />{:else}<GoroutinesPanel />{/if}
  </div>
</section>

<style>
  .block {
    display: grid;
    gap: 10px;
  }
  .tabs {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 14px;
  }
  .tab {
    border: 0;
    background: none;
    padding: 0 0 3px;
    cursor: pointer;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--muted);
    white-space: nowrap;
    border-bottom: 2px solid transparent;
  }
  .tab.on {
    color: var(--ink);
    border-bottom-color: var(--sand);
  }
</style>
