<script lang="ts">
  import { onMount } from 'svelte'
  import { bridge } from '../bridge'
  import DebugToolbar from '../shell/DebugToolbar.svelte'
  import {
    activePath,
    activeText,
    breakpoints,
    baseName,
    buffers,
    closeSplit,
    cursor,
    debugActive,
    debugState,
    editBuffer,
    errorsRevealed,
    editorSplit,
    openTabs,
    problems,
    revealRequest,
    showErrorsInCode,
    showInSplit,
    toggleBreakpoint
  } from '../stores'
  import { t } from '../i18n'
  import Editor from './Editor.svelte'
  import { registerEditorShortcuts } from './editorShortcuts'
  import TabBar from './TabBar.svelte'
  import { toMarks } from './toMarks'
  import { languageWiring } from './wiring'
  import { editorFontSize } from './zoom'

  const marks = $derived(
    toMarks(
      $activePath,
      $showErrorsInCode ? $problems : [],
      ($activePath && $breakpoints[$activePath]) || [],
      $debugState?.frames[0]?.location?.file ?? null,
      $debugState?.frames[0]?.location?.line ?? null
    )
  )
  const splitMarks = $derived(
    toMarks(
      $editorSplit?.path ?? null,
      $showErrorsInCode ? $problems : [],
      ($editorSplit && $breakpoints[$editorSplit.path]) || [],
      $debugState?.frames[0]?.location?.file ?? null,
      $debugState?.frames[0]?.location?.line ?? null
    )
  )
  const wiring = languageWiring(bridge)
  // Typing hides again the problems a run had revealed (when the setting keeps them off).
  const editedByUser = (path: string | null, text: string): void => {
    if (!path) return
    errorsRevealed.set(false)
    editBuffer(path, text)
  }
  // Left and up put the second editor first; right and down put it after the first.
  const stacked = $derived($editorSplit?.direction === 'down' || $editorSplit?.direction === 'up')
  const splitFirst = $derived(
    $editorSplit?.direction === 'left' || $editorSplit?.direction === 'up'
  )

  onMount(() => registerEditorShortcuts(bridge))
</script>

<section class="editor">
  <TabBar />
  <div class="code" class:stacked>
    {#if $activePath}
      <div class="group" style:order={splitFirst ? 2 : 1}>
        <Editor
          path={$activePath}
          text={$activeText}
          {marks}
          {wiring}
          fontSize={$editorFontSize}
          goto={$revealRequest}
          onChange={(text) => editedByUser($activePath, text)}
          onCursor={(line, column) => cursor.set({ line, column })}
          onToggleBreakpoint={(line) => $activePath && toggleBreakpoint(bridge, $activePath, line)}
        />
      </div>
    {/if}
    {#if $editorSplit}
      {@const split = $editorSplit}
      <div class="group split" style:order={splitFirst ? 1 : 2}>
        <div class="split-bar">
          <select
            aria-label={$t('editor.split.file')}
            value={split.path}
            onchange={(event) => showInSplit(event.currentTarget.value)}
          >
            {#each $openTabs as path (path)}
              <option value={path}>{baseName(path)}</option>
            {/each}
          </select>
          <button type="button" aria-label={$t('editor.split.close')} onclick={closeSplit}>×</button
          >
        </div>
        <div class="split-editor">
          {#key split.path}
            <Editor
              path={split.path}
              text={$buffers[split.path] ?? ''}
              marks={splitMarks}
              {wiring}
              fontSize={$editorFontSize}
              goto={null}
              onChange={(text) => editedByUser(split.path, text)}
              onCursor={(line, column) => cursor.set({ line, column })}
              onToggleBreakpoint={(line) => toggleBreakpoint(bridge, split.path, line)}
            />
          {/key}
        </div>
      </div>
    {/if}
  </div>
  <!-- The step buttons only make sense while the program is paused. -->
  {#if $debugActive && $debugState}<DebugToolbar />{/if}
</section>

<style>
  .editor {
    position: relative;
    height: 100%;
    display: grid;
    grid-template-rows: 36px 1fr;
    min-width: 0;
    min-height: 0;
  }
  .code {
    min-height: 0;
    overflow: hidden;
    display: flex;
  }
  .code.stacked {
    flex-direction: column;
  }
  .group {
    flex: 1 1 0;
    min-width: 0;
    min-height: 0;
    overflow: hidden;
  }
  .group.split {
    display: grid;
    grid-template-rows: 28px 1fr;
    border-left: 1px solid var(--line);
  }
  .stacked .group.split {
    border-left: 0;
    border-top: 1px solid var(--line);
  }
  /* "left" and "up" put the second editor first, so the divider sits on its other side. */
  .group.split[style*='order: 1'] {
    border: 0;
    border-right: 1px solid var(--line);
  }
  .stacked .group.split[style*='order: 1'] {
    border-right: 0;
    border-bottom: 1px solid var(--line);
  }
  .split-bar {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 0 8px;
    background: var(--chrome);
    border-bottom: 1px solid var(--line);
  }
  .split-bar select {
    flex: 1;
    min-width: 0;
    font: 600 13px var(--ui);
    color: var(--ink);
    background: var(--chrome);
    border: 0;
  }
  .split-bar button {
    border: 0;
    background: none;
    color: var(--muted);
    font-size: 16px;
    cursor: pointer;
    border-radius: 6px;
  }
  .split-bar button:hover {
    background: var(--go-soft);
    color: var(--ink);
  }
  .split-editor {
    min-height: 0;
    overflow: hidden;
  }
</style>
