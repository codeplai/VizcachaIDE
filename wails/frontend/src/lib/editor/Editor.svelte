<script lang="ts">
  import { onMount } from 'svelte'
  import { t } from '../i18n'
  import {
    createEditor,
    type EditorHandle,
    type EditorHandlers,
    type LanguageWiring
  } from './createEditor'
  import { lspStatus, profileOf } from '../stores/codeLanguages'
  import { registerEditorBridge } from '../stores/editorBridge'
  import { settings } from '../stores/settings'
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
    const editor = createEditor(host, handlers, wiring, (file) => profileOf(file))
    handle = editor
    const unregister = registerEditorBridge({
      applyChanges: editor.applyChanges,
      undoFile: editor.undoFile,
      redoFile: editor.redoFile,
      renameSymbol: editor.renameSymbol,
      findReferences: editor.findReferences
    })
    return () => {
      unregister()
      editor.destroy()
    }
  })

  $effect(() => handle?.show(path, text))
  $effect(() => handle?.setMarks(marks))
  $effect(() => handle?.setPhrases(editorPhrases($t)))
  $effect(() => handle?.setFontSize(fontSize))
  $effect(() => handle?.setInlayHints($settings?.inlayHints ?? true))
  // New diagnostics or a ready server mean the file was analysed: hints asked while the server
  // was still loading (rust-analyzer takes seconds) came back empty, so ask again.
  $effect(() => {
    void marks
    void $lspStatus
    handle?.refreshInlayHints()
  })
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
