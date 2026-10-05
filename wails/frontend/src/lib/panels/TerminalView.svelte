<script lang="ts">
  import { FitAddon } from '@xterm/addon-fit'
  import { Terminal } from '@xterm/xterm'
  import '@xterm/xterm/css/xterm.css'
  import { onMount } from 'svelte'
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { attachTerminalOutput, terminalFocusRequests, type TerminalSession } from '../stores'
  import { orderedInput } from './terminalInput'
  import { terminalKeyHandler } from './terminalKeys'
  import { terminalFont, terminalTheme, watchTheme } from './terminalTheme'

  let { session, active }: { session: TerminalSession; active: boolean } = $props()

  let host: HTMLDivElement
  let term = $state.raw<Terminal>()
  let fit: FitAddon | undefined
  let announced = false

  const fitToPanel = (): void => {
    if (!term || !fit || !host.clientWidth || !host.clientHeight) return
    fit.fit()
  }

  const copySelection = (): void => {
    const text = term?.getSelection() ?? ''
    if (text) void bridge.system.writeClipboard(text)
    term?.clearSelection()
  }

  const paste = async (): Promise<void> => {
    const text = await bridge.system.readClipboard().catch(() => '')
    if (text) term?.paste(text)
  }

  /** Right click copies the selection, or pastes when nothing is selected (as Windows Terminal). */
  const rightClick = (event: MouseEvent): void => {
    event.preventDefault()
    if (term?.hasSelection()) copySelection()
    else void paste()
  }

  onMount(() => {
    const created = new Terminal({
      fontFamily: terminalFont(),
      fontSize: 13,
      cursorBlink: true,
      scrollback: 5000,
      theme: terminalTheme()
    })
    fit = new FitAddon()
    created.loadAddon(fit)
    created.open(host)
    created.attachCustomKeyEventHandler(
      terminalKeyHandler({
        hasSelection: () => created.hasSelection(),
        copy: copySelection,
        paste: () => void paste()
      })
    )
    created.onData(orderedInput((data) => bridge.terminal.write(session.id, data)))
    created.onResize(
      ({ cols, rows }) => void bridge.terminal.resize(session.id, cols, rows).catch(() => {})
    )
    const detach = attachTerminalOutput(session.id, (data) => created.write(data))
    const stopWatching = watchTheme(() => (created.options.theme = terminalTheme()))
    const observer = new ResizeObserver(fitToPanel)
    observer.observe(host)
    term = created
    fitToPanel()
    void bridge.terminal.resize(session.id, created.cols, created.rows).catch(() => {})
    return () => {
      observer.disconnect()
      stopWatching()
      detach()
      created.dispose()
    }
  })

  $effect(() => {
    if (term && session.exitCode !== null && !announced) {
      announced = true
      const message = $t('terminal.exited', { values: { code: session.exitCode } })
      term.write(`\r\n\x1b[2m${message}\x1b[0m\r\n`)
    }
  })

  $effect(() => {
    void $terminalFocusRequests
    if (active && term) {
      fitToPanel()
      term.focus()
    }
  })
</script>

<div
  class="view"
  class:hidden={!active}
  bind:this={host}
  role="presentation"
  oncontextmenu={rightClick}
></div>

<style>
  .view {
    position: absolute;
    inset: 0;
    padding: 4px 0 0 8px;
    box-sizing: border-box;
    background: var(--win);
    overflow: hidden;
  }
  .view.hidden {
    visibility: hidden;
  }
  .view :global(.xterm) {
    height: 100%;
  }
  .view :global(.xterm-viewport) {
    background: var(--win) !important;
  }
</style>
