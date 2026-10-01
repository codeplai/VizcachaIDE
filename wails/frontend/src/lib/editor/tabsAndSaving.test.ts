import { get } from 'svelte/store'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { Bridge } from '../bridge'
import { createMockBridge } from '../bridge/mock'
import { SAMPLE_CALC, SAMPLE_MAIN } from '../bridge/mockData'
import { activePath, buffers, dirty, editBuffer, openFile, openTabs } from '../stores'
import { gotoRequest, goToLocation } from './navigation'
import { lineFromFormatError, saveDocument, saveNotice } from './saving'
import { answerCloseRequest, closeRequest, requestCloseTab, tabItems } from './tabs'

const reset = (): void => {
  openTabs.set([])
  activePath.set(null)
  buffers.set({})
  dirty.set({})
  closeRequest.set(null)
  saveNotice.set(null)
  gotoRequest.set(null)
}

afterEach(reset)

const openBoth = async (bridge: Bridge): Promise<void> => {
  await openFile(bridge, SAMPLE_MAIN)
  await openFile(bridge, SAMPLE_CALC)
}

describe('tabs', () => {
  it('lists open files with the modified dot and the active one', async () => {
    const { bridge } = createMockBridge()
    await openBoth(bridge)
    editBuffer(SAMPLE_MAIN, 'package main\n// edit\n')
    expect(get(tabItems)).toEqual([
      { path: SAMPLE_MAIN, name: 'main.go', modified: true, active: false },
      { path: SAMPLE_CALC, name: 'calculadora.go', modified: false, active: true }
    ])
  })

  it('closes a clean tab at once', async () => {
    const { bridge } = createMockBridge()
    await openBoth(bridge)
    await requestCloseTab(bridge, SAMPLE_CALC)
    expect(get(openTabs)).toEqual([SAMPLE_MAIN])
    expect(get(closeRequest)).toBeNull()
  })

  it('asks before closing a tab with changes, and cancel keeps it', async () => {
    const { bridge } = createMockBridge()
    await openBoth(bridge)
    editBuffer(SAMPLE_CALC, 'changed')
    await requestCloseTab(bridge, SAMPLE_CALC)
    expect(get(closeRequest)).toEqual({ path: SAMPLE_CALC, file: 'calculadora.go' })
    expect(get(openTabs)).toHaveLength(2)
    await answerCloseRequest(bridge, 'cancel')
    expect(get(openTabs)).toHaveLength(2)
    expect(get(closeRequest)).toBeNull()
  })

  it('discarding forgets the unsaved text', async () => {
    const { bridge } = createMockBridge()
    await openBoth(bridge)
    editBuffer(SAMPLE_CALC, 'changed')
    await requestCloseTab(bridge, SAMPLE_CALC)
    await answerCloseRequest(bridge, 'discard')
    expect(get(openTabs)).toEqual([SAMPLE_MAIN])
    expect(get(buffers)[SAMPLE_CALC]).toBeUndefined()
    expect(get(dirty)[SAMPLE_CALC]).toBeUndefined()
  })

  it('saving first writes the file and then closes', async () => {
    const { bridge } = createMockBridge()
    const save = vi.spyOn(bridge.files, 'saveFile')
    await openBoth(bridge)
    editBuffer(SAMPLE_CALC, 'package main\n')
    await requestCloseTab(bridge, SAMPLE_CALC)
    await answerCloseRequest(bridge, 'save')
    expect(save).toHaveBeenCalledWith(SAMPLE_CALC, 'package main\n')
    expect(get(openTabs)).toEqual([SAMPLE_MAIN])
  })
})

describe('saving', () => {
  it('formats and then saves, clearing the modified mark', async () => {
    const { bridge } = createMockBridge()
    const save = vi.spyOn(bridge.files, 'saveFile')
    await openFile(bridge, SAMPLE_MAIN)
    editBuffer(SAMPLE_MAIN, 'func f() {\n    x()\n}')
    expect(await saveDocument(bridge, SAMPLE_MAIN)).toBe(true)
    expect(save).toHaveBeenCalledWith(SAMPLE_MAIN, 'func f() {\n\tx()\n}')
    expect(get(buffers)[SAMPLE_MAIN]).toBe('func f() {\n\tx()\n}')
    expect(get(dirty)[SAMPLE_MAIN]).toBe(false)
  })

  it('saves as it is when gofmt rejects the code, and reports the line', async () => {
    const { bridge } = createMockBridge()
    const save = vi.spyOn(bridge.files, 'saveFile')
    vi.spyOn(bridge.run, 'format').mockRejectedValue('main.go:7:3: expected declaration')
    await openFile(bridge, SAMPLE_MAIN)
    editBuffer(SAMPLE_MAIN, 'broken')
    await saveDocument(bridge, SAMPLE_MAIN)
    expect(save).toHaveBeenCalledWith(SAMPLE_MAIN, 'broken')
    expect(get(saveNotice)).toEqual({ kind: 'formatRejected', line: 7 })
  })

  it('reports a failed save and keeps the file modified', async () => {
    const { bridge } = createMockBridge()
    vi.spyOn(bridge.files, 'saveFile').mockRejectedValue(new Error('access denied'))
    await openFile(bridge, SAMPLE_MAIN)
    editBuffer(SAMPLE_MAIN, 'x')
    expect(await saveDocument(bridge, SAMPLE_MAIN)).toBe(false)
    expect(get(saveNotice)).toEqual({
      kind: 'saveFailed',
      file: 'main.go',
      reason: 'access denied'
    })
    expect(get(dirty)[SAMPLE_MAIN]).toBe(true)
  })

  it('reads the line from a gofmt error', () => {
    expect(lineFromFormatError(new Error('<standard input>:12:4: bad'))).toBe(12)
    expect(lineFromFormatError('no position')).toBeNull()
  })
})

describe('navigation', () => {
  it('opens the file, then asks the editor to move there', async () => {
    const { bridge } = createMockBridge()
    await openFile(bridge, SAMPLE_MAIN)
    await goToLocation(bridge, { file: SAMPLE_CALC, line: 3, column: 2 })
    expect(get(activePath)).toBe(SAMPLE_CALC)
    const first = get(gotoRequest)
    expect(first).toMatchObject({ file: SAMPLE_CALC, line: 3, column: 2 })
    await goToLocation(bridge, { file: SAMPLE_CALC, line: 3, column: 2 })
    expect(get(gotoRequest)?.nonce).not.toBe(first?.nonce)
  })
})
