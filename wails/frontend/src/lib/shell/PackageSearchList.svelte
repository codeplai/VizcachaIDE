<script lang="ts">
  import { t } from '../i18n'
  import type { PackageInfo } from '../domain'
  import type { PackageSearchState } from '../stores'

  interface Props {
    state: PackageSearchState
    listId: string
    optionId: (index: number) => string
    onChoose: (found: PackageInfo) => void
  }
  let { state, listId, optionId, onChoose }: Props = $props()
</script>

{#if state.status === 'searching'}
  <p class="hint" role="status">{$t('packages.searching')}</p>
{:else if state.status === 'failed'}
  <p class="hint bad" role="status">{$t('packages.searchFailed')}</p>
{:else if state.status === 'done' && state.results.length === 0}
  <p class="hint" role="status">{$t('packages.noResults', { values: { term: state.term } })}</p>
{:else if state.status === 'done'}
  <ul id={listId} role="listbox" aria-label={$t('packages.results')}>
    {#each state.results as found, index (found.name)}
      <li
        id={optionId(index)}
        role="option"
        tabindex="-1"
        aria-selected={index === state.active}
        class:active={index === state.active}
        onmousedown={(event) => event.preventDefault()}
        onclick={() => onChoose(found)}
        onkeydown={(event) => {
          if (event.key !== 'Enter' && event.key !== ' ') return
          event.preventDefault()
          onChoose(found)
        }}
      >
        <span class="head">
          <span class="name">{found.name}</span>
          {#if found.version}<span class="version">{found.version}</span>{/if}
        </span>
        {#if found.description}<span class="description">{found.description}</span>{/if}
      </li>
    {/each}
  </ul>
{/if}

<style>
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    max-height: 220px;
    overflow-y: auto;
    border: 1px solid var(--line);
    border-radius: 7px;
    background: var(--win);
  }
  li {
    display: grid;
    gap: 2px;
    padding: 6px 10px;
    cursor: pointer;
    border-bottom: 1px solid var(--line);
  }
  li:last-child {
    border-bottom: 0;
  }
  li:hover,
  li:focus-visible,
  li.active {
    background: var(--go-soft);
  }
  .head {
    display: flex;
    gap: 8px;
    align-items: baseline;
  }
  .name {
    font: 700 13.5px var(--mono);
    overflow-wrap: anywhere;
  }
  .version {
    font-size: 12px;
    color: var(--muted);
  }
  .description {
    font-size: 12.5px;
    color: var(--muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .bad {
    color: var(--err);
  }
</style>
