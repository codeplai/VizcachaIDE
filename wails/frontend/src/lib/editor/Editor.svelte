<script lang="ts">
  import { onMount } from 'svelte'
  import { createEditor, type EditorHandle, type EditorHandlers } from './createEditor'
  import type { EditorMarks } from './marks'

  interface Props extends EditorHandlers {
    text: string
    marks: EditorMarks
  }

  let { text, marks, onChange, onCursor, onToggleBreakpoint }: Props = $props()

  let host: HTMLDivElement
  let handle: EditorHandle | undefined

  onMount(() => {
    handle = createEditor(host, text, { onChange, onCursor, onToggleBreakpoint })
    handle.setMarks(marks)
    return () => handle?.destroy()
  })

  $effect(() => handle?.setText(text))
  $effect(() => handle?.setMarks(marks))
</script>

<div class="host" bind:this={host}></div>

<style>
  .host {
    height: 100%;
    min-height: 0;
    overflow: hidden;
  }
</style>
