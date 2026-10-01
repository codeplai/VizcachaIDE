<script lang="ts">
  import { bridge } from '../bridge'
  import type { Variable } from '../domain'
  import { t } from '../i18n'
  import { expandedReferences, loadingReferences, toggleVariable } from '../stores'
  import VariableCard from './VariableCard.svelte'

  let { variable }: { variable: Variable } = $props()

  const expandable = $derived(variable.reference > 0)
  const open = $derived(expandable && $expandedReferences.has(variable.reference))
  const loading = $derived(open && $loadingReferences.has(variable.reference))
</script>

<div class="var" class:changed={variable.changed}>
  {#if expandable}
    <button
      type="button"
      class="n toggle"
      aria-expanded={open}
      aria-label={$t(open ? 'panels.collapseVariable' : 'panels.expandVariable', {
        values: { name: variable.name }
      })}
      onclick={() => toggleVariable(bridge, variable)}
    >
      <span aria-hidden="true">{open ? '▾' : '▸'}</span>
      {variable.name}
    </button>
  {:else}
    <span class="n">{variable.name}</span>
  {/if}
  <span class="t">
    {variable.changed ? `${variable.typeName} · ${$t('panels.justChanged')}` : variable.typeName}
  </span>
  <span class="v">{variable.value}</span>
  {#if open}
    <div class="kids" aria-live="polite">
      {#if loading}
        <span class="loading">{$t('panels.loading')}</span>
      {/if}
      {#each variable.children as child, position (position)}
        <VariableCard variable={child} />
      {/each}
    </div>
  {/if}
</div>

<style>
  .var {
    background: var(--win);
    border: 1px solid var(--line);
    border-radius: 8px;
    padding: 7px 10px;
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 2px 8px;
    font-size: 13px;
  }
  .var.changed {
    border-color: var(--sand);
    background: var(--sand-soft);
    box-shadow: inset 3px 0 0 var(--sand);
  }
  .n {
    font: 600 13px var(--mono);
  }
  .toggle {
    border: 0;
    background: none;
    padding: 0;
    text-align: left;
    cursor: pointer;
    color: inherit;
  }
  .t {
    font: 400 11.5px var(--mono);
    color: var(--muted);
    grid-column: 1;
  }
  .var.changed .t {
    color: var(--ink);
  }
  .v {
    font: 600 15px var(--mono);
    color: var(--go);
    grid-row: 1 / span 2;
    grid-column: 2;
    align-self: center;
  }
  .kids {
    grid-column: 1 / -1;
    display: grid;
    gap: 6px;
    margin-top: 6px;
    padding-left: 10px;
    border-left: 2px solid var(--line);
  }
  .loading {
    color: var(--muted);
    font-size: 12.5px;
  }
</style>
