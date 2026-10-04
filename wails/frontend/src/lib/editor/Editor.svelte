<script lang="ts">
  import { onMount } from 'svelte'
  import { t } from '../i18n'
  import {
    createEditor,
    type EditorHandle,
    type EditorHandlers,
    type LanguageWiring
  } from './createEditor'
  import { profileOf } from '../stores/codeLanguages'
  import type { RevealRequest } from '../stores/navigation'
  import type { EditorMarks } from './marks'
  import { editorPhrases } from './phrases'

  interface Props extends EditorHandlers {
    path: string
    text: string
    marks: EditorMarks
    fontSize: number
    goto: RevealRequest | null
    wiring: LanguageWiring | null
  }

  let { path, text, marks, fontSize, goto, wiring, ...handlers }: Props = $props()

  let host: HTMLDivElement
  let handle: EditorHandle | undefined
  let lastGoto = 0

  onMount(() => {
    handle = createEditor(host, handlers, wiring, (file) => profileOf(file))
    return () => handle?.destroy()
  })

  $effect(() => handle?.show(path, text))
  $effect(() => handle?.setMarks(marks))
  $effect(() => handle?.setPhrases(editorPhrases($t)))
  $effect(() => handle?.setFontSize(fontSize))
  $effect(() => {
    if (!goto || goto.nonce === lastGoto || goto.location.file !== path) return
    lastGoto = goto.nonce
    handle?.goTo(goto.location.line, goto.location.column)
  })
</script>

<div class="host" bind:this={host}></div>

<style>
  .host {
    height: 100%;
    min-height: 0;
    overflow: hidden;
  }
</style>
