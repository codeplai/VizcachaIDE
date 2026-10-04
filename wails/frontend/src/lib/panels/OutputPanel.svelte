<script lang="ts">
  import { tick } from 'svelte'
  import { bridge } from '../bridge'
  import { formatSeconds, locale, t } from '../i18n'
  import {
    clearOutput,
    copyText,
    goToLocation,
    outputLines,
    readClipboard,
    programInputOpen,
    selectAllIn,
    stdinDraft,
    type OutputLine
  } from '../stores'
  import PanelMenu, { type PanelAction } from './PanelMenu.svelte'
  import StdinInput from './StdinInput.svelte'

  let terminal: HTMLDivElement | undefined = $state()

  const valuesOf = (line: OutputLine, language: string | null | undefined) => {
    const values: Record<string, string | number> = { ...line.values }
    if (line.seconds !== undefined) values.seconds = formatSeconds(language ?? 'en', line.seconds)
    return values
  }

  const pasteIntoInput = async (): Promise<void> => {
    const text = await readClipboard(bridge)
    stdinDraft.update((draft) => draft + text.replace(/\r?\n$/, ''))
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
      run: () => void copyText(bridge, terminal?.innerText || terminal?.textContent || '')
    },
    {
      label: 'panels.menuPasteInput',
      disabled: !$programInputOpen,
      run: () => void pasteIntoInput()
    },
    { label: 'panels.menuSelectAll', run: () => selectAllIn(terminal) },
    { label: 'panels.menuClear', separated: true, run: clearOutput }
  ]

  // Follow the program: keep the newest line in view.
  $effect(() => {
    void $outputLines
    void tick().then(() => {
      if (terminal) terminal.scrollTop = terminal.scrollHeight
    })
  })
</script>

<div class="pane">
  <PanelMenu {actions}>
    <div
      class="term"
      role="log"
      aria-live="polite"
      aria-label={$t('a11y.output')}
      bind:this={terminal}
    >
      {#each $outputLines as line, index (index)}
        <div class={line.tone}>
          {#if line.key}
            {$t(line.key, { values: valuesOf(line, $locale) })}
          {:else}
            {#each line.segments ?? [] as segment, position (position)}
              {#if segment.location}
                {@const location = segment.location}
                <button
                  type="button"
                  class="place"
                  title={$t('a11y.goToPlace', { values: { place: segment.text } })}
                  onclick={() => goToLocation(bridge, location)}
                >
                  {segment.text}
                </button>
              {:else if segment.color || segment.background || segment.decorations}
                <span
                  class={(segment.decorations ?? []).map((name) => `ansi-${name}`).join(' ')}
                  style:color={segment.color}
                  style:background-color={segment.background}>{segment.text}</span
                >
              {:else}
                {segment.text}
              {/if}
            {/each}
          {/if}
        </div>
      {:else}
        <div class="system">{$t('empty.output')}</div>
      {/each}
    </div>
  </PanelMenu>
  {#if $programInputOpen}<StdinInput />{/if}
</div>

<style>
  .pane {
    height: 100%;
    display: flex;
    flex-direction: column;
  }
  .term {
    flex: 1;
    min-height: 0;
    font: 400 13px/1.6 var(--mono);
    padding: 8px 14px;
    overflow: auto;
    white-space: pre-wrap;
  }
  .system {
    color: var(--muted);
  }
  .error {
    color: var(--err);
  }
  .success {
    color: var(--ok);
  }
  .term :global(.ansi-bold) {
    font-weight: 700;
  }
  .term :global(.ansi-dim) {
    opacity: 0.7;
  }
  .term :global(.ansi-italic) {
    font-style: italic;
  }
  .term :global(.ansi-underline) {
    text-decoration: underline;
  }
  .term :global(.ansi-strikethrough) {
    text-decoration: line-through;
  }
  .place {
    border: 0;
    background: none;
    padding: 0;
    font: inherit;
    color: var(--go);
    text-decoration: underline;
    cursor: pointer;
  }
  .place:hover {
    background: var(--go-soft);
  }
</style>
