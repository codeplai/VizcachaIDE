<script lang="ts">
  import { tick } from 'svelte'
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import {
    activeCodeLanguage,
    clearConsoleScreen,
    consoleEntries,
    copyText,
    evalConsole,
    readClipboard,
    resetConsole,
    selectAllIn,
    stepConsoleHistory
  } from '../stores'
  import PanelMenu, { type PanelAction } from './PanelMenu.svelte'

  // Python's console looks like python's own; Go keeps its single-chevron prompt and its texts.
  const python = $derived($activeCodeLanguage === 'python')
  const prompt = $derived(python ? '>>>' : '›')
  const text = (name: string): string => (python ? `console.${name}Python` : `console.${name}`)

  let draft = $state('')
  let scroller: HTMLDivElement | undefined = $state()

  const submit = async (): Promise<void> => {
    const code = draft
    draft = ''
    await evalConsole(bridge, code)
    await tick()
    if (scroller) scroller.scrollTop = scroller.scrollHeight
  }

  // The arrows only walk the history from the first/last line of a multi-line draft.
  const atEdge = (direction: -1 | 1, field: HTMLTextAreaElement): boolean =>
    direction === -1
      ? !draft.slice(0, field.selectionStart).includes('\n')
      : !draft.slice(field.selectionEnd).includes('\n')

  const recall = (direction: -1 | 1, event: KeyboardEvent): void => {
    if (!atEdge(direction, event.currentTarget as HTMLTextAreaElement)) return
    const text = stepConsoleHistory(direction)
    if (text === null) return
    event.preventDefault()
    draft = text
  }

  const paste = async (): Promise<void> => {
    draft += await readClipboard(bridge)
  }

  const actions = (selection: string): PanelAction[] => [
    {
      label: 'panels.menuCopy',
      keys: 'Ctrl+C',
      disabled: !selection,
      run: (text) => void copyText(bridge, text)
    },
    {
      label: 'panels.menuCopyAll',
      run: () => void copyText(bridge, scroller?.innerText || scroller?.textContent || '')
    },
    { label: 'panels.menuPaste', run: () => void paste() },
    { label: 'panels.menuSelectAll', run: () => selectAllIn(scroller) },
    { label: 'panels.menuClearScreen', separated: true, run: clearConsoleScreen },
    { label: 'console.reset', run: () => void resetConsole(bridge) }
  ]

  const onKeydown = (event: KeyboardEvent): void => {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      void submit()
    } else if (event.key === 'ArrowUp') recall(-1, event)
    else if (event.key === 'ArrowDown') recall(1, event)
  }
</script>

<div class="console">
  <PanelMenu {actions}>
    <div class="scroll" bind:this={scroller}>
      {#if $consoleEntries.length === 0}
        <p class="empty">{$t(text('empty'))}</p>
        <p class="note">{$t(text('note'))}</p>
      {:else}
        {#each $consoleEntries as entry (entry.id)}
          <div class="entry">
            <pre class="code"><span class="prompt">{prompt}</span> {entry.code}</pre>
            {#if entry.output}<pre class="out">{entry.output}</pre>{/if}
            {#if entry.result}<pre class="result">{entry.result}</pre>{/if}
            {#if entry.error}<pre class="err">{entry.error}</pre>{/if}
          </div>
        {/each}
      {/if}
    </div>
  </PanelMenu>
  <div class="bar">
    <span class="prompt">{prompt}</span>
    <textarea
      rows="1"
      spellcheck="false"
      aria-label={$t(text('inputLabel'))}
      placeholder={$t(text('inputPlaceholder'))}
      bind:value={draft}
      onkeydown={onKeydown}></textarea>
    <button
      type="button"
      class="reset"
      title={$t('console.resetHint')}
      onclick={() => void resetConsole(bridge)}>{$t('console.reset')}</button
    >
  </div>
</div>

<style>
  .console {
    height: 100%;
    display: grid;
    grid-template-rows: 1fr auto;
    min-height: 0;
  }
  .scroll {
    overflow: auto;
    padding: 8px 14px;
    font: 13px var(--mono);
  }
  pre {
    margin: 0;
    white-space: pre-wrap;
    word-break: break-word;
    font: inherit;
  }
  .prompt {
    color: var(--go);
    font-weight: 700;
  }
  .code {
    color: var(--muted);
  }
  .result,
  .out {
    color: var(--ink);
  }
  .err {
    color: var(--err);
  }
  .entry {
    padding: 2px 0;
  }
  .empty,
  .note {
    margin: 0 0 4px;
    font: 13px var(--ui);
    color: var(--muted);
  }
  .note {
    font-size: 12px;
  }
  .bar {
    display: flex;
    gap: 8px;
    align-items: start;
    padding: 6px 14px;
    border-top: 1px solid var(--line);
    background: var(--chrome);
  }
  textarea {
    flex: 1;
    resize: none;
    border: 0;
    outline: 0;
    background: none;
    color: var(--ink);
    font: 13px var(--mono);
    field-sizing: content;
    max-height: 120px;
  }
  .reset {
    font: 600 12.5px var(--ui);
    border: 1px solid var(--line);
    background: var(--win);
    color: var(--ink);
    border-radius: 7px;
    padding: 4px 10px;
    cursor: pointer;
  }
</style>
