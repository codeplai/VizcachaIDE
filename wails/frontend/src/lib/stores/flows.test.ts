// The flows that the editor and the shell share, checked end to end with the mock backend:
// one navigation, one Ctrl+S and one close-with-changes dialog.
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import type { Bridge } from '../bridge'
import { createMockBridge } from '../bridge/mock'
import { SAMPLE_CALC, SAMPLE_MAIN } from '../bridge/mockData'
import EditorPane from '../editor/EditorPane.svelte'
import { registerEditorShortcuts } from '../editor/editorShortcuts'
import { setupI18n } from '../i18n'
import ConfirmHost from '../shell/ConfirmHost.svelte'
import NoticeBar from '../shell/NoticeBar.svelte'
import { registerShortcuts } from '../shell/shortcuts'
import {
  activePath,
  buffers,
  connectStores,
  cursor,
  dirty,
  editBuffer,
  fileTree,
  goToLocation,
  notice,
  openFile,
  openTabs,
  requestCloseTab,
  resetRun,
  revealRequest,
  saveActiveFile,
  updateSettings
} from '.'

let stop: (() => void) | undefined

const reset = (): void => {
  openTabs.set([])
  activePath.set(null)
  buffers.set({})
  dirty.set({})
  fileTree.set(null)
  notice.set(null)
  revealRequest.set(null)
  resetRun()
}

const setUp = async (): Promise<Bridge> => {
  const { bridge } = createMockBridge()
  stop = await connectStores(bridge)
  return bridge
}

beforeAll(() => setupI18n('en'))
beforeEach(reset)
afterEach(() => {
  cleanup()
  stop?.()
  reset()
})

const pressCtrlS = (): void => {
  window.dispatchEvent(new KeyboardEvent('keydown', { key: 's', ctrlKey: true, cancelable: true }))
}

describe('one navigation', () => {
  it('opens a closed file, activates it and asks to reveal the place (what every panel does)', async () => {
    const bridge = await setUp()
    await openFile(bridge, SAMPLE_MAIN)
    await goToLocation(bridge, { file: SAMPLE_CALC, line: 3, column: 2 })
    expect(get(openTabs)).toEqual([SAMPLE_MAIN, SAMPLE_CALC])
    expect(get(activePath)).toBe(SAMPLE_CALC)
    const first = get(revealRequest)
    expect(first?.location).toEqual({ file: SAMPLE_CALC, line: 3, column: 2 })
    await goToLocation(bridge, { file: SAMPLE_CALC, line: 3, column: 2 })
    expect(get(revealRequest)?.nonce).not.toBe(first?.nonce)
  })

  it('moves the cursor of the editor, also in a file that was not open', async () => {
    const bridge = await setUp()
    render(EditorPane)
    await openFile(bridge, SAMPLE_MAIN)
    await goToLocation(bridge, { file: SAMPLE_MAIN, line: 4, column: 1 })
    await waitFor(() => expect(get(cursor)).toEqual({ line: 4, column: 1 }))
    await goToLocation(bridge, { file: SAMPLE_CALC, line: 2, column: 1 })
    await waitFor(() => expect(get(cursor).line).toBe(2))
    expect(get(activePath)).toBe(SAMPLE_CALC)
  })
})

describe('one Ctrl+S', () => {
  it('saves exactly once with both shortcut handlers registered', async () => {
    const bridge = await setUp()
    const save = vi.spyOn(bridge.files, 'saveFile')
    const offShell = registerShortcuts(bridge)
    const offEditor = registerEditorShortcuts(bridge)
    await openFile(bridge, SAMPLE_MAIN)
    editBuffer(SAMPLE_MAIN, 'func f() {\n    x()\n}')
    pressCtrlS()
    await waitFor(() => expect(get(dirty)[SAMPLE_MAIN]).toBe(false))
    expect(save).toHaveBeenCalledTimes(1)
    expect(save).toHaveBeenCalledWith(SAMPLE_MAIN, 'func f() {\n\tx()\n}')
    offShell()
    offEditor()
  })

  it('does not format when the setting is off', async () => {
    const bridge = await setUp()
    await updateSettings(bridge, { formatOnSave: false })
    const save = vi.spyOn(bridge.files, 'saveFile')
    await openFile(bridge, SAMPLE_MAIN)
    editBuffer(SAMPLE_MAIN, 'a\n    b')
    await saveActiveFile(bridge)
    expect(save).toHaveBeenCalledWith(SAMPLE_MAIN, 'a\n    b')
  })

  it('says why it was not formatted, saves anyway, and shows one message', async () => {
    const bridge = await setUp()
    vi.spyOn(bridge.run, 'format').mockRejectedValue(new Error('main.go:7:3: expected declaration'))
    const save = vi.spyOn(bridge.files, 'saveFile')
    render(NoticeBar)
    await openFile(bridge, SAMPLE_MAIN)
    editBuffer(SAMPLE_MAIN, 'broken')
    await saveActiveFile(bridge)
    expect(save).toHaveBeenCalledWith(SAMPLE_MAIN, 'broken')
    expect(get(notice)?.messageKey).toBe('errors.formatRejected')
    expect(await screen.findAllByRole('alert')).toHaveLength(1)
    expect(screen.getByText(/Go couldn't read line 7/)).toBeTruthy()
  })

  it('shows the save error with Try again, and the retry saves', async () => {
    const bridge = await setUp()
    const failing = vi.spyOn(bridge.files, 'saveFile').mockRejectedValueOnce(new Error('disk full'))
    render(NoticeBar)
    await openFile(bridge, SAMPLE_MAIN)
    editBuffer(SAMPLE_MAIN, 'x')
    await saveActiveFile(bridge)
    expect(get(dirty)[SAMPLE_MAIN]).toBe(true)
    await fireEvent.click(await screen.findByRole('button', { name: 'Try again' }))
    await waitFor(() => expect(get(dirty)[SAMPLE_MAIN]).toBe(false))
    expect(failing).toHaveBeenCalledTimes(2)
  })
})

describe('one close-with-changes dialog', () => {
  const closeDirty = async (bridge: Bridge): Promise<{ closing: Promise<void> }> => {
    await openFile(bridge, SAMPLE_MAIN)
    await openFile(bridge, SAMPLE_CALC)
    editBuffer(SAMPLE_CALC, 'package main\n')
    render(ConfirmHost)
    return { closing: requestCloseTab(bridge, SAMPLE_CALC) }
  }

  it('Cancel keeps the tab and the text', async () => {
    const bridge = await setUp()
    const { closing } = await closeDirty(bridge)
    await fireEvent.click(await screen.findByRole('button', { name: 'Cancel' }))
    await closing
    expect(get(openTabs)).toEqual([SAMPLE_MAIN, SAMPLE_CALC])
    expect(get(buffers)[SAMPLE_CALC]).toBe('package main\n')
  })

  it("Don't save closes and forgets the text, also in the language server", async () => {
    const bridge = await setUp()
    const closeDocument = vi.spyOn(bridge.language, 'closeDocument')
    const save = vi.spyOn(bridge.files, 'saveFile')
    const { closing } = await closeDirty(bridge)
    await fireEvent.click(await screen.findByRole('button', { name: /Don.t save/ }))
    await closing
    expect(get(openTabs)).toEqual([SAMPLE_MAIN])
    expect(get(buffers)[SAMPLE_CALC]).toBeUndefined()
    expect(get(dirty)[SAMPLE_CALC]).toBeUndefined()
    expect(closeDocument).toHaveBeenCalledWith(SAMPLE_CALC)
    expect(save).not.toHaveBeenCalled()
  })

  it('Save writes the file once and then closes', async () => {
    const bridge = await setUp()
    const save = vi.spyOn(bridge.files, 'saveFile')
    const { closing } = await closeDirty(bridge)
    await fireEvent.click(await screen.findByRole('button', { name: 'Save' }))
    await closing
    expect(save).toHaveBeenCalledTimes(1)
    expect(get(openTabs)).toEqual([SAMPLE_MAIN])
  })

  it('a clean tab closes without asking, and can be opened again', async () => {
    const bridge = await setUp()
    await openFile(bridge, SAMPLE_MAIN)
    await requestCloseTab(bridge, SAMPLE_MAIN)
    expect(get(openTabs)).toEqual([])
    await openFile(bridge, SAMPLE_MAIN)
    expect(get(buffers)[SAMPLE_MAIN]).toBeTruthy()
  })
})
