// More shared flows with the mock backend: one zoom, one first-build message, one way to start.
import { cleanup, render, waitFor } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import type { Bridge } from '../bridge'
import { createMockBridge } from '../bridge/mock'
import { SAMPLE_MAIN } from '../bridge/mockData'
import { parseDevQuery } from '../devQuery'
import EditorPane from '../editor/EditorPane.svelte'
import { registerEditorShortcuts } from '../editor/editorShortcuts'
import { zoomEditor } from '../editor/zoom'
import { setupI18n } from '../i18n'
import {
  activePath,
  buffers,
  connectRun,
  connectStores,
  dirty,
  fileTree,
  openFile,
  openTabs,
  outputLines,
  resetRun,
  settings
} from '.'
import { startApp } from './startup'

let stop: (() => void) | undefined

beforeAll(() => setupI18n('en'))
beforeEach(() => {
  openTabs.set([])
  activePath.set(null)
  buffers.set({})
  dirty.set({})
  fileTree.set(null)
  resetRun()
})
afterEach(() => {
  cleanup()
  stop?.()
  stop = undefined
})

const setUp = async (): Promise<Bridge> => {
  const { bridge } = createMockBridge()
  stop = await connectStores(bridge)
  return bridge
}

describe('one font size', () => {
  it('zoom is saved in Settings and the editor shows that size', async () => {
    const bridge = await setUp()
    const view = render(EditorPane)
    await openFile(bridge, SAMPLE_MAIN)
    await zoomEditor(bridge, 'in')
    await zoomEditor(bridge, 'in')
    expect((await bridge.settings.get()).fontSize).toBe(16)
    await waitFor(() => {
      const host = view.container.querySelector<HTMLElement>('.code > div')
      expect(host?.style.getPropertyValue('--editor-font-size')).toBe('16px')
    })
    expect(document.documentElement.style.getPropertyValue('--editor-font-size')).toBe('')
    expect(get(settings)?.fontSize).toBe(16)
  })

  it('Ctrl+= and Ctrl+0 zoom through the keyboard', async () => {
    const bridge = await setUp()
    const off = registerEditorShortcuts(bridge)
    window.dispatchEvent(new KeyboardEvent('keydown', { key: '=', ctrlKey: true }))
    await waitFor(() => expect(get(settings)?.fontSize).toBe(15))
    window.dispatchEvent(new KeyboardEvent('keydown', { key: '0', ctrlKey: true }))
    await waitFor(() => expect(get(settings)?.fontSize).toBe(14))
    off()
  })
})

describe('one first-build message', () => {
  const FIRST_BUILD = 'Preparing Go for the first time.'

  it('is the backend line only: no second message from a timer', () => {
    vi.useFakeTimers()
    const handlers: Record<string, (payload: unknown) => void> = {}
    const fake = {
      on: (name: string, handler: (payload: unknown) => void) => {
        handlers[name] = handler
        return () => {}
      }
    } as unknown as Bridge
    const off = connectRun(fake)
    handlers['run:started']?.({ target: 'main.go' })
    vi.advanceTimersByTime(60_000)
    expect(get(outputLines).filter((line) => line.key === 'run.firstBuild')).toHaveLength(0)
    handlers['run:output']?.({ stream: 'stdout', text: `${FIRST_BUILD}\n` })
    handlers['run:output']?.({ stream: 'stdout', text: 'Hola\n' })
    vi.advanceTimersByTime(60_000)
    const shown = get(outputLines).filter((line) => line.text === FIRST_BUILD)
    expect(shown).toHaveLength(1)
    expect(get(outputLines).some((line) => line.key === 'run.firstBuild')).toBe(false)
    off()
    vi.useRealTimers()
  })
})

describe('start', () => {
  it('asks for the last folder with listTree("") and never opens the folder dialog', async () => {
    const { bridge } = createMockBridge()
    const list = vi.spyOn(bridge.files, 'listTree')
    const dialog = vi.spyOn(bridge.files, 'openFolder')
    stop = await startApp(bridge, null, parseDevQuery(''))
    expect(list).toHaveBeenCalledWith('')
    expect(dialog).not.toHaveBeenCalled()
    expect(get(fileTree)?.name).toBe('hola-go')
    expect(get(activePath)).toBe(SAMPLE_MAIN)
  })

  it('starts empty when the backend has no last folder', async () => {
    const { bridge } = createMockBridge()
    vi.spyOn(bridge.files, 'listTree').mockResolvedValue({
      name: '',
      path: '',
      isDir: true,
      children: []
    })
    stop = await startApp(bridge, null, parseDevQuery(''))
    expect(get(fileTree)).toBeNull()
  })
})
