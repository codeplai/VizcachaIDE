import { get } from 'svelte/store'
import { beforeEach, describe, expect, it } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import { SAMPLE_CALC, SAMPLE_MAIN } from '../bridge/mockData'
import type { Bridge } from '../bridge'
import {
  activePath,
  buffers,
  connectStores,
  dirty,
  editBuffer,
  forgetRename,
  hasFreshRename,
  notice,
  openFile,
  openTabs,
  renameSymbol,
  undoRename
} from '.'

let bridge: Bridge
let originalMain = ''
let originalCalc = ''
const atSumar = { file: SAMPLE_MAIN, line: 5, column: 7 }

beforeEach(async () => {
  bridge = createMockBridge().bridge
  buffers.set({})
  dirty.set({})
  openTabs.set([])
  activePath.set(null)
  notice.set(null)
  forgetRename()
  await connectStores(bridge)
  originalMain = await bridge.files.readFile(SAMPLE_MAIN)
  originalCalc = await bridge.files.readFile(SAMPLE_CALC)
  await openFile(bridge, SAMPLE_MAIN) // calculadora.go stays closed: edited on disk
  await renameSymbol(bridge, atSumar, 'total')
})

describe('undoing a rename as a whole', () => {
  it('offers Undo in the notice and is fresh right after the rename', () => {
    expect(get(notice)?.actions.map((action) => action.labelKey)).toEqual(['refactor.undo'])
    expect(hasFreshRename()).toBe(true)
  })

  it('puts back every file: the open buffer and the file on disk', async () => {
    expect(await bridge.files.readFile(SAMPLE_CALC)).toContain('total(')
    expect(await undoRename(bridge)).toBe(true)
    expect(get(buffers)[SAMPLE_MAIN]).toBe(originalMain)
    expect(await bridge.files.readFile(SAMPLE_CALC)).toBe(originalCalc)
    expect(get(dirty)[SAMPLE_MAIN]).toBe(false) // it was saved before the rename
    expect(get(notice)).toMatchObject({ messageKey: 'refactor.undone', tone: 'info' })
    expect(hasFreshRename()).toBe(false)
  })

  it('the Undo button of the notice does the same', async () => {
    get(notice)?.actions[0]?.run()
    await new Promise((resolve) => setTimeout(resolve, 20))
    expect(get(buffers)[SAMPLE_MAIN]).toBe(originalMain)
    expect(await bridge.files.readFile(SAMPLE_CALC)).toBe(originalCalc)
  })

  it('is not fresh once the open file was edited: the editor undoes as usual', async () => {
    editBuffer(SAMPLE_MAIN, `${get(buffers)[SAMPLE_MAIN]}// more\n`)
    expect(hasFreshRename()).toBe(false)
    expect(await undoRename(bridge)).toBe(false)
    expect(get(notice)?.messageKey).toBe('refactor.undoStale')
    expect(await bridge.files.readFile(SAMPLE_CALC)).toContain('total(') // nothing was half-undone
  })

  it('does not overwrite a closed file that changed after the rename', async () => {
    await bridge.files.saveFile(SAMPLE_CALC, 'package main\n// edited elsewhere\n')
    expect(await undoRename(bridge)).toBe(true)
    expect(get(buffers)[SAMPLE_MAIN]).toBe(originalMain)
    expect(await bridge.files.readFile(SAMPLE_CALC)).toBe('package main\n// edited elsewhere\n')
    expect(get(notice)).toMatchObject({ messageKey: 'refactor.undoPartial', values: { files: 1 } })
  })

  it('is not fresh when the file on screen is not one of the renamed files', async () => {
    await openFile(bridge, SAMPLE_CALC) // now open, but the rename did not edit its buffer
    activePath.set(null)
    expect(hasFreshRename()).toBe(false)
  })
})
