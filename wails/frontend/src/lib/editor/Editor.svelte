<script lang="ts">
  import { onMount } from 'svelte'
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import {
    createEditor,
    type EditorHandle,
    type EditorHandlers,
    type LanguageWiring
  } from './createEditor'
  import { lspStatus, profileOf } from '../stores/codeLanguages'
  import { registerEditorBridge, type EditorBridge } from '../stores/editorBridge'
  import { splitEditor, type SplitDirection } from '../stores/split'
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
  let menu = $state<{ x: number; y: number; selected: boolean } | null>(null)

  const openMenu = (event: MouseEvent): void => {
    event.preventDefault()
    menu = { x: event.clientX, y: event.clientY, selected: !!handle?.selectedText() }
  }
  const closeMenu = (): void => {
    menu = null
  }
  const menuEntries = [
    { id: 'cut', label: 'editor.cut', key: 'X', needsSelection: true },
    { id: 'copy', label: 'editor.copy', key: 'C', needsSelection: true },
    { id: 'paste', label: 'editor.paste', key: 'V', needsSelection: false },
    { id: 'selectAll', label: 'editor.selectAll', key: 'A', needsSelection: false }
  ] as const
  const splitEntries: { direction: SplitDirection; label: string }[] = [
    { direction: 'right', label: 'editor.split.right' },
    { direction: 'left', label: 'editor.split.left' },
    { direction: 'down', label: 'editor.split.down' },
    { direction: 'up', label: 'editor.split.up' }
  ]
  const split = (direction: SplitDirection): void => {
    closeMenu()
    splitEditor(direction)
  }
  const modifier = /Mac|iPhone|iPad/.test(navigator.platform) ? '⌘' : 'Ctrl+'
  const runEdit = async (id: (typeof menuEntries)[number]['id']): Promise<void> => {
    closeMenu()
    if (!handle) return
    const text = handle.selectedText()
    if (id === 'selectAll') handle.selectAll()
    else if (id === 'copy') await bridge.system.writeClipboard(text).catch(() => {})
    else if (id === 'cut') {
      await bridge.system.writeClipboard(text).catch(() => {})
      handle.replaceSelection('', 'delete.cut')
    } else {
      const pasted = await bridge.system.readClipboard().catch(() => '')
      if (pasted) handle.replaceSelection(pasted, 'input.paste')
    }
    handle.focus()
  }

  onMount(() => {
    const editor = createEditor(host, handlers, wiring, (file) => profileOf(file))
    handle = editor
    const editorBridge: EditorBridge = {
      applyChanges: editor.applyChanges,
      undoFile: editor.undoFile,
      redoFile: editor.redoFile,
      renameSymbol: editor.renameSymbol,
      findReferences: editor.findReferences
    }
    const unregister = registerEditorBridge(editorBridge)
    // With the editor split, the one the user works in is the one the menus act on.
    const claim = (): void => void registerEditorBridge(editorBridge)
    host.addEventListener('focusin', claim)
    return () => {
      host.removeEventListener('focusin', claim)
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

<svelte:window onpointerdown={closeMenu} onblur={closeMenu} />

<div class="host" role="presentation" bind:this={host} oncontextmenu={openMenu}></div>

{#if menu}
  <div
    class="context-menu"
    role="menu"
    tabindex="-1"
    style="left: {Math.min(menu.x, window.innerWidth - 210)}px; top: {Math.min(
      menu.y,
      window.innerHeight - 240
    )}px"
    onpointerdown={(event) => event.stopPropagation()}
  >
    {#each menuEntries as entry (entry.id)}
      <button
        type="button"
        role="menuitem"
        disabled={entry.needsSelection && !menu.selected}
        onclick={() => void runEdit(entry.id)}
      >
        <span>{$t(entry.label)}</span>
        <kbd>{modifier}{entry.key}</kbd>
      </button>
    {/each}
    <hr />
    {#each splitEntries as entry (entry.direction)}
      <button type="button" role="menuitem" onclick={() => split(entry.direction)}>
        <span>{$t(entry.label)}</span>
        {#if entry.direction === 'right'}<kbd>{modifier}\</kbd>{/if}
      </button>
    {/each}
  </div>
{/if}

<style>
  .host {
    height: 100%;
    min-height: 0;
    overflow: hidden;
  }
  .context-menu {
    position: fixed;
    z-index: 50;
    min-width: 200px;
    padding: 4px;
    display: grid;
    background: var(--win);
    color: var(--ink);
    border: 1px solid var(--line);
    border-radius: 10px;
    box-shadow: 0 10px 24px -14px rgba(8, 18, 28, 0.55);
  }
  .context-menu button {
    border: 0;
    background: none;
    color: inherit;
    font: 400 13px var(--ui);
    text-align: left;
    display: flex;
    justify-content: space-between;
    gap: 24px;
    padding: 6px 10px;
    border-radius: 6px;
    cursor: pointer;
  }
  .context-menu hr {
    width: 100%;
    margin: 4px 0;
    border: 0;
    border-top: 1px solid var(--line);
  }
  .context-menu kbd {
    font: 400 12px var(--ui);
    color: var(--muted);
  }
  .context-menu button:hover:not(:disabled) {
    background: var(--go-soft);
  }
  .context-menu button:disabled {
    color: var(--muted);
    cursor: default;
  }
</style>
