<script lang="ts">
  import { onMount, tick } from 'svelte'
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import {
    matchCount,
    patchSearch,
    replaceEverywhere,
    runSearch,
    searchFocus,
    searchState
  } from '../stores'
  import type { SearchOptions } from '../domainSearch'
  import SearchResults from './SearchResults.svelte'
  import SearchStatus from './SearchStatus.svelte'

  let field: HTMLInputElement | undefined = $state()
  let timer: ReturnType<typeof setTimeout> | undefined
  let focused = 0

  const total = $derived(matchCount($searchState.files))

  /** Searches shortly after the student stops typing, or at once (Enter). */
  const schedule = (now = false): void => {
    clearTimeout(timer)
    timer = setTimeout(() => void runSearch(bridge), now ? 0 : 250)
  }

  const toggle = (name: keyof SearchOptions): void => {
    patchSearch({ options: { ...$searchState.options, [name]: !$searchState.options[name] } })
    schedule(true)
  }

  const OPTIONS: { name: keyof SearchOptions; mark: string; label: string }[] = [
    { name: 'caseSensitive', mark: 'Aa', label: 'search.caseSensitive' },
    { name: 'wholeWord', mark: 'ab', label: 'search.wholeWord' },
    { name: 'regex', mark: '.*', label: 'search.regex' }
  ]

  // Ctrl+Shift+F asks for the focus; the field may not exist yet when the panel was hidden.
  $effect(() => {
    if ($searchFocus === focused) return
    focused = $searchFocus
    void tick().then(() => field?.select())
  })
  onMount(() => () => clearTimeout(timer))
</script>

<section class="side" aria-label={$t('panels.search')}>
  <h2 class="side-h">{$t('panels.search')}</h2>
  <div class="form">
    <div class="line">
      <button
        type="button"
        class="chev"
        aria-expanded={$searchState.showReplace}
        aria-label={$t('search.toggleReplace')}
        title={$t('search.toggleReplace')}
        onclick={() => patchSearch({ showReplace: !$searchState.showReplace })}
        >{$searchState.showReplace ? '▾' : '▸'}</button
      >
      <input
        bind:this={field}
        class="input"
        type="text"
        value={$searchState.query}
        placeholder={$t('search.placeholder')}
        aria-label={$t('search.label')}
        spellcheck="false"
        oninput={(event) => {
          patchSearch({ query: event.currentTarget.value, replaced: null })
          schedule()
        }}
        onkeydown={(event) => event.key === 'Enter' && schedule(true)}
      />
      {#each OPTIONS as option (option.name)}
        <button
          type="button"
          class="opt"
          aria-pressed={$searchState.options[option.name]}
          title={$t(option.label)}
          aria-label={$t(option.label)}
          onclick={() => toggle(option.name)}>{option.mark}</button
        >
      {/each}
    </div>
    {#if $searchState.showReplace}
      <div class="line">
        <span class="chev"></span>
        <input
          class="input"
          type="text"
          value={$searchState.replacement}
          placeholder={$t('search.replacePlaceholder')}
          aria-label={$t('search.replaceLabel')}
          spellcheck="false"
          oninput={(event) => patchSearch({ replacement: event.currentTarget.value })}
        />
        <button
          type="button"
          class="all"
          disabled={total === 0}
          title={$t('search.replaceAllTip')}
          onclick={() => void replaceEverywhere(bridge)}>{$t('search.replaceAll')}</button
        >
      </div>
    {/if}
  </div>

  <SearchStatus />
  <SearchResults />
</section>

<style>
  .side {
    height: 100%;
    background: var(--chrome);
    padding: 10px 0 0;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }
  .side-h {
    margin: 0;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--muted);
    padding: 2px 14px 6px;
  }
  .form {
    display: grid;
    gap: 4px;
    padding: 0 8px;
  }
  .line {
    display: flex;
    gap: 3px;
    align-items: center;
  }
  .chev {
    width: 18px;
    flex: none;
    border: 0;
    background: none;
    color: var(--muted);
    cursor: pointer;
    padding: 0;
  }
  .input {
    flex: 1;
    min-width: 0;
    border: 1px solid var(--line);
    background: var(--win);
    color: var(--ink);
    border-radius: 6px;
    padding: 5px 7px;
    font: 13px var(--ui);
  }
  .input:focus-visible {
    outline: 2px solid var(--go);
    outline-offset: -1px;
  }
  .opt,
  .all {
    flex: none;
    border: 1px solid var(--line);
    background: var(--win);
    color: var(--ink);
    border-radius: 5px;
    padding: 3px 6px;
    font: 700 12px var(--ui);
    cursor: pointer;
  }
  .opt[aria-pressed='true'] {
    border-color: var(--go);
    background: var(--go-soft, rgba(0, 120, 200, 0.14));
  }
  .all:disabled {
    opacity: 0.5;
    cursor: default;
  }
  .opt:focus-visible,
  .all:focus-visible,
  .chev:focus-visible {
    outline: 2px solid var(--go);
  }
</style>
