// The rename and the editor's history together: a real editor, the stores and the mock bridge.
import { EditorView } from '@codemirror/view'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it } from 'vitest'
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
  notice,
  openFile,
  openTabs,
  registerEditorBridge,
  renameSymbol
} from '../stores'
import { createEditor, type EditorHandle } from './createEditor'
import { languageWiring } from './wiring'

beforeAll(() => {
  Range.prototype.getClientRects = () => [] as unknown as DOMRectList
  Range.prototype.getBoundingClientRect = () => new DOMRect()
})

let bridge: Bridge
let parent: HTMLElement
let handle: EditorHandle
let unregister: () => void
let originalMain = ''
let originalCalc = ''
const atSumar = { file: SAMPLE_MAIN, line: 5, column: 7 }

const view = (): EditorView => EditorView.findFromDOM(parent) as EditorView
const key = (name: string, shift = false): void => {
  parent
    .querySelector('.cm-content')
    ?.dispatchEvent(
      new KeyboardEvent('keydown', { key: name, ctrlKey: true, shiftKey: shift, bubbles: true })
    )
}
const settle = (): Promise<void> => new Promise((resolve) => setTimeout(resolve, 30))

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
  await openFile(bridge, SAMPLE_MAIN) // calculadora.go stays closed
  parent = document.createElement('div')
  document.body.append(parent)
  handle = createEditor(
    parent,
    {
      onChange: (text) => editBuffer(SAMPLE_MAIN, text),
      onCursor: () => {},
      onToggleBreakpoint: () => {}
    },
    languageWiring(bridge)
  )
  unregister = registerEditorBridge({
    applyChanges: handle.applyChanges,
    undoFile: handle.undoFile,
    redoFile: handle.redoFile,
    renameSymbol: handle.renameSymbol,
    findReferences: handle.findReferences
  })
  handle.show(SAMPLE_MAIN, originalMain)
  // The student types a comment at the end first, then renames.
  view().dispatch({
    changes: { from: originalMain.length, insert: '// hola\n' },
    userEvent: 'input.type'
  })
  await renameSymbol(bridge, atSumar, 'total')
})

afterEach(() => {
  unregister()
  handle.destroy()
  parent.remove()
})

describe('rename and the editor history', () => {
  it('Ctrl+Z undoes the whole rename and pops it: nothing is appended', async () => {
    key('z')
    await settle()
    expect(view().state.doc.toString()).toBe(`${originalMain}// hola\n`)
    expect(await bridge.files.readFile(SAMPLE_CALC)).toBe(originalCalc)
  })

  it('a second Ctrl+Z undoes the earlier typing and leaves the other files alone', async () => {
    key('z')
    await settle()
    key('z')
    await settle()
    expect(view().state.doc.toString()).toBe(originalMain)
    expect(await bridge.files.readFile(SAMPLE_CALC)).toBe(originalCalc)
    expect(get(buffers)[SAMPLE_MAIN]).toBe(originalMain)
  })

  it('Ctrl+Y right after the whole undo redoes the rename in every file', async () => {
    key('z')
    await settle()
    key('y')
    await settle()
    expect(view().state.doc.toString()).toContain('func total(a, b int) int {')
    expect(await bridge.files.readFile(SAMPLE_CALC)).toContain('return total(a, a)')
    expect(get(notice)?.messageKey).toBe('refactor.redone')
    key('z') // and it can be undone as a whole again
    await settle()
    expect(await bridge.files.readFile(SAMPLE_CALC)).toBe(originalCalc)
    expect(view().state.doc.toString()).toBe(`${originalMain}// hola\n`)
  })

  it('Ctrl+Y after going on undoing the typing redoes only the typing, not the rename', async () => {
    key('z')
    await settle()
    key('z')
    await settle()
    key('y')
    await settle()
    expect(view().state.doc.toString()).toBe(`${originalMain}// hola\n`)
    expect(await bridge.files.readFile(SAMPLE_CALC)).toBe(originalCalc)
  })

  it('the Undo button of the notice pops the editor history the same way', async () => {
    get(notice)?.actions[0]?.run()
    await settle()
    expect(view().state.doc.toString()).toBe(`${originalMain}// hola\n`)
    key('z')
    await settle()
    expect(view().state.doc.toString()).toBe(originalMain)
    expect(await bridge.files.readFile(SAMPLE_CALC)).toBe(originalCalc)
  })
})
