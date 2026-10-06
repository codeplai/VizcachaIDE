<script lang="ts">
  import { tick } from 'svelte'
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import Highlighted from './Highlighted.svelte'
  import {
    chooseQuickOpen,
    closeQuickOpen,
    quickOpenItems,
    quickOpenQuery,
    quickOpenVisible
  } from '../stores'

  let input: HTMLInputElement | undefined = $state()
  let active = $state(0)
  const items = $derived($quickOpenItems)
  const activeId = $derived(items[active] ? `quick-open-${active}` : undefined)

  // Opening the palette focuses the field; typing starts again at the best match.
  let before: Element | null = null
  $effect(() => {
    if ($quickOpenVisible) {
      before = document.activeElement
      void tick().then(() => input?.focus())
    } else if (before instanceof HTMLElement) {
      before.focus() // back to the editor (or wherever the student was)
      before = null
    }
  })
  $effect(() => {
    void $quickOpenQuery
    active = 0
  })

  const move = (step: number): void => {
    if (items.length === 0) return
    active = (active + step + items.length) % items.length
    void tick().then(() =>
      document.getElementById(`quick-open-${active}`)?.scrollIntoView({ block: 'nearest' })
    )
  }

  const onKeydown = (event: KeyboardEvent): void => {
    const keys: Record<string, () => void> = {
      ArrowDown: () => move(1),
      ArrowUp: () => move(-1),
      Home: () => move(-active),
      End: () => move(items.length - 1 - active),
      Escape: closeQuickOpen,
      Enter: () => {
        const item = items[active]
        if (item) void chooseQuickOpen(bridge, item)
      }
    }
    const run = keys[event.key]
    if (!run || (event.isComposing && event.key === 'Enter')) return
    event.preventDefault()
    event.stopPropagation()
    run()
  }
</script>

{#if $quickOpenVisible}
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div class="backdrop" onclick={closeQuickOpen}>
    <div
      class="palette"
      role="dialog"
      aria-modal="true"
      aria-label={$t('quickOpen.title')}
      tabindex="-1"
      onclick={(event) => event.stopPropagation()}
    >
      <input
        bind:this={input}
        bind:value={$quickOpenQuery}
        class="field"
        type="text"
        role="combobox"
        aria-autocomplete="list"
        aria-expanded="true"
        aria-controls="quick-open-list"
        aria-activedescendant={activeId}
        aria-label={$t('quickOpen.label')}
        placeholder={$t('quickOpen.placeholder')}
        autocomplete="off"
        spellcheck="false"
        onkeydown={onKeydown}
      />
      <ul id="quick-open-list" class="list" role="listbox" aria-label={$t('quickOpen.results')}>
        {#each items as item, index (item.path)}
          {@const nameStart = item.relative.length - item.name.length}
          <li
            id={`quick-open-${index}`}
            class="row"
            class:active={index === active}
            role="option"
            aria-selected={index === active}
            onmousemove={() => (active = index)}
            onclick={() => void chooseQuickOpen(bridge, item)}
          >
            <span class="name"
              ><Highlighted text={item.name} indices={item.indices} offset={nameStart} /></span
            >
            {#if item.dir}
              <span class="dir"><Highlighted text={item.dir} indices={item.indices} /></span>
            {/if}
          </li>
        {/each}
      </ul>
      {#if items.length === 0}
        <p class="empty">{$t('quickOpen.empty')}</p>
      {/if}
      <p class="sr-only" aria-live="polite">
        {$t('quickOpen.count', { values: { count: items.length } })}
      </p>
    </div>
  </div>
{/if}

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    z-index: 40;
    display: flex;
    justify-content: center;
    align-items: flex-start;
    padding-top: 12vh;
    background: rgba(8, 18, 28, 0.25);
  }
  .palette {
    width: min(560px, calc(100vw - 32px));
    background: var(--win);
    color: var(--ink);
    border: 1px solid var(--line);
    border-radius: 12px;
    box-shadow: 0 24px 60px -20px rgba(8, 18, 28, 0.6);
    overflow: hidden;
  }
  .field {
    width: 100%;
    box-sizing: border-box;
    border: 0;
    border-bottom: 1px solid var(--line);
    background: transparent;
    color: inherit;
    padding: 12px 14px;
    font: 15px var(--ui);
    outline: none;
  }
  .field:focus-visible {
    box-shadow: inset 0 -2px 0 var(--go);
  }
  .list {
    list-style: none;
    margin: 0;
    padding: 4px;
    max-height: 340px;
    overflow: auto;
  }
  .row {
    display: flex;
    gap: 10px;
    align-items: baseline;
    padding: 6px 10px;
    border-radius: 6px;
    cursor: pointer;
    font: 14px var(--ui);
  }
  .row.active {
    background: var(--go-soft, rgba(0, 120, 200, 0.14));
    outline: 1px solid var(--go);
  }
  .name {
    font-weight: 600;
    white-space: nowrap;
  }
  .dir {
    color: var(--muted);
    font-size: 12.5px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .empty {
    margin: 0;
    padding: 12px 14px;
    color: var(--muted);
    font-size: 14px;
  }
</style>
