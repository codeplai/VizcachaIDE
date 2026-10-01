<script lang="ts">
  import { bridge } from '../bridge'
  import DebugToolbar from '../shell/DebugToolbar.svelte'
  import {
    activePath,
    baseName,
    breakpoints,
    cursor,
    debugActive,
    debugState,
    editBuffer,
    activeText,
    openTabs,
    problems,
    toggleBreakpoint,
    openFile
  } from '../stores'
  import Editor from './Editor.svelte'
  import { toMarks } from './toMarks'

  const marks = $derived(
    toMarks(
      $activePath,
      $problems,
      ($activePath && $breakpoints[$activePath]) || [],
      $debugState?.frames[0]?.location?.file ?? null,
      $debugState?.frames[0]?.location?.line ?? null
    )
  )
</script>

<section class="editor">
  <div class="tabs" role="tablist">
    {#each $openTabs as path (path)}
      <button
        type="button"
        role="tab"
        class="tab"
        class:on={path === $activePath}
        aria-selected={path === $activePath}
        onclick={() => openFile(bridge, path)}
      >
        {baseName(path)}
      </button>
    {/each}
  </div>
  <div class="code">
    {#if $activePath}
      <Editor
        text={$activeText}
        {marks}
        onChange={(text) => $activePath && editBuffer($activePath, text)}
        onCursor={(line, column) => cursor.set({ line, column })}
        onToggleBreakpoint={(line) => $activePath && toggleBreakpoint(bridge, $activePath, line)}
      />
    {/if}
  </div>
  {#if $debugActive}<DebugToolbar />{/if}
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
  .tabs {
    display: flex;
    background: var(--chrome);
    border-bottom: 1px solid var(--line);
    overflow: hidden;
  }
  .tab {
    border: 0;
    border-right: 1px solid var(--line);
    background: none;
    padding: 0 16px;
    display: flex;
    align-items: center;
    font-size: 13px;
    color: var(--muted);
    cursor: pointer;
  }
  .tab.on {
    background: var(--win);
    color: var(--ink);
    font-weight: 600;
    box-shadow: inset 0 2px 0 var(--go);
  }
  .code {
    min-height: 0;
    overflow: hidden;
  }
</style>
