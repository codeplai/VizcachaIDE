import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { bridge } from '../bridge'
import { setupI18n } from '../i18n'
import { registerShortcuts } from '../shell/shortcuts'
import {
  connectTerminal,
  outputTab,
  resetTerminals,
  terminalSessions,
  type TerminalSession
} from '../stores'
import BottomPanel from './BottomPanel.svelte'
import TerminalPanel from './TerminalPanel.svelte'

// xterm needs a real layout engine: a stand-in records what the panel asks of it.
interface FakeTerminal {
  written: string[]
  selection: string
  typed: (data: string) => void
  keyHandler: (event: KeyboardEvent) => boolean
  disposed: boolean
}
const created = vi.hoisted(() => [] as FakeTerminal[])

vi.mock('@xterm/xterm', () => ({
  Terminal: class {
    cols = 80
    rows = 24
    options: Record<string, unknown> = {}
    fake: FakeTerminal = {
      written: [],
      selection: '',
      typed: () => {},
      keyHandler: () => true,
      disposed: false
    }
    constructor(options: Record<string, unknown>) {
      this.options = options
      created.push(this.fake)
    }
    loadAddon() {}
    open() {}
    focus() {}
    write(data: string) {
      this.fake.written.push(data)
    }
    paste(data: string) {
      this.fake.written.push(`paste:${data}`)
    }
    hasSelection() {
      return this.fake.selection !== ''
    }
    getSelection() {
      return this.fake.selection
    }
    clearSelection() {
      this.fake.selection = ''
    }
    attachCustomKeyEventHandler(handler: (event: KeyboardEvent) => boolean) {
      this.fake.keyHandler = handler
    }
    onData(listener: (data: string) => void) {
      this.fake.typed = listener
    }
    onResize() {}
    dispose() {
      this.fake.disposed = true
    }
  }
}))
vi.mock('@xterm/addon-fit', () => ({
  FitAddon: class {
    fit() {}
  }
}))

class FakeResizeObserver {
  observe() {}
  disconnect() {}
}

let disconnect: () => void

beforeAll(() => {
  setupI18n('en')
  vi.stubGlobal('ResizeObserver', FakeResizeObserver)
})

beforeEach(() => {
  resetTerminals()
  created.length = 0
  outputTab.set('output')
  disconnect = connectTerminal(bridge)
})

afterEach(() => {
  cleanup()
  disconnect()
  vi.restoreAllMocks()
})

const sessions = (): TerminalSession[] => get(terminalSessions)
const fake = (index = 0): FakeTerminal => {
  const found = created[index]
  if (!found) throw new Error(`terminal ${index} was not created`)
  return found
}
const written = (index = 0): string => fake(index).written.join('')

describe('Terminal panel', () => {
  it('starts one session the first time it is shown, and not again', async () => {
    const first = render(TerminalPanel, { visible: true })
    await waitFor(() => expect(sessions()).toHaveLength(1))
    expect(screen.getByRole('tab', { name: /^1: / })).toBeTruthy()
    first.unmount()
    render(TerminalPanel, { visible: true })
    await Promise.resolve()
    expect(sessions()).toHaveLength(1)
  })

  it('draws the output events in xterm and sends the typed keys to the shell', async () => {
    const write = vi.spyOn(bridge.terminal, 'write')
    render(TerminalPanel, { visible: true })
    await waitFor(() => expect(written()).toContain('Mock terminal'))
    fake(0).typed('ls')
    expect(write).toHaveBeenCalledWith(sessions()[0]?.id, 'ls')
    await waitFor(() => expect(written()).toContain('ls'))
  })

  it('opens another session with New terminal and closes the active one with Kill terminal', async () => {
    const close = vi.spyOn(bridge.terminal, 'close')
    render(TerminalPanel, { visible: true })
    await waitFor(() => expect(sessions()).toHaveLength(1))
    await fireEvent.click(screen.getByRole('button', { name: '+ New terminal' }))
    await waitFor(() => expect(sessions()).toHaveLength(2))
    expect(screen.getByRole('tab', { name: /^2: / })).toBeTruthy()
    const [first, second] = sessions() as [TerminalSession, TerminalSession]
    await fireEvent.click(screen.getByRole('button', { name: 'Kill terminal' }))
    await waitFor(() => expect(sessions()).toHaveLength(1))
    expect(close).toHaveBeenCalledWith(second.id)
    expect(sessions()[0]?.id).toBe(first.id)
    expect(fake(1).disposed).toBe(true)
  })

  it('shows a muted line when the shell exits', async () => {
    render(TerminalPanel, { visible: true })
    await waitFor(() => expect(sessions()).toHaveLength(1))
    fake(0).typed('exit\r')
    await waitFor(() => expect(written()).toContain('[process exited with code 0]'))
    expect(sessions()[0]?.exitCode).toBe(0)
  })

  it('keeps its shells when another tab is shown and back', async () => {
    outputTab.set('terminal')
    render(BottomPanel)
    await waitFor(() => expect(sessions()).toHaveLength(1))
    await fireEvent.click(screen.getByRole('tab', { name: 'Output' }))
    expect(fake(0).disposed).toBe(false)
    await fireEvent.click(screen.getByRole('tab', { name: 'Terminal' }))
    expect(sessions()).toHaveLength(1)
  })
})

describe('Ctrl+` shortcut', () => {
  it('shows the Terminal tab and starts a shell', async () => {
    const stop = registerShortcuts(bridge)
    window.dispatchEvent(new KeyboardEvent('keydown', { key: '`', ctrlKey: true }))
    await waitFor(() => expect(sessions()).toHaveLength(1))
    expect(get(outputTab)).toBe('terminal')
    stop()
  })

  it('leaves the other editor shortcuts out while the terminal has the focus', () => {
    const stop = registerShortcuts(bridge)
    const host = document.createElement('div')
    host.className = 'xterm'
    document.body.append(host)
    const event = new KeyboardEvent('keydown', {
      key: 'w',
      ctrlKey: true,
      bubbles: true,
      cancelable: true
    })
    host.dispatchEvent(event)
    expect(event.defaultPrevented).toBe(false)
    host.remove()
    stop()
  })
})
