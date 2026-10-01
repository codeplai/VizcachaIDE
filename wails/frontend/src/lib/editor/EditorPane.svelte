<script lang="ts">
  import { onMount } from 'svelte'
  import { bridge } from '../bridge'
  import DebugToolbar from '../shell/DebugToolbar.svelte'
  import {
    activePath,
    activeText,
    breakpoints,
    cursor,
    debugActive,
    debugState,
    editBuffer,
    problems,
    revealRequest,
    toggleBreakpoint
  } from '../stores'
  import Editor from './Editor.svelte'
  import { registerEditorShortcuts } from './editorShortcuts'
  import TabBar from './TabBar.svelte'
  import { toMarks } from './toMarks'
  import { languageWiring } from './wiring'
  import { editorFontSize } from './zoom'

  const marks = $derived(
    toMarks(
      $activePath,
      $problems,
      ($activePath && $breakpoints[$activePath]) || [],
      $debugState?.frames[0]?.location?.file ?? null,
      $debugState?.frames[0]?.location?.line ?? null
    )
  )
  const wiring = languageWiring(bridge)

  onMount(() => registerEditorShortcuts(bridge))
</script>

<section class="editor">
  <TabBar />
  <div class="code">
    {#if $activePath}
      <Editor
        path={$activePath}
        text={$activeText}
        {marks}
        {wiring}
        fontSize={$editorFontSize}
        goto={$revealRequest}
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
  .code {
    min-height: 0;
    overflow: hidden;
  }
</style>
