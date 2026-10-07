import { undo } from '@codemirror/commands'
import { EditorView } from '@codemirror/view'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import type { RenameTarget } from '../domain'
import { createEditor, type EditorHandle } from './createEditor'
import type { RefactorWiring } from './refactor'

const SOURCE = 'package main\n\nfunc sumar(a, b int) int {\n\treturn a + b\n}\n'
const FILE = 'main.go'

// jsdom does not lay text out; CodeMirror measures it with these.
beforeAll(() => {
  Range.prototype.getClientRects = () => [] as unknown as DOMRectList
  Range.prototype.getBoundingClientRect = () => new DOMRect()
})

let parent: HTMLElement
let handle: EditorHandle
let refactor: RefactorWiring
const allowed: RenameTarget = { refusal: '', placeholder: '' }

beforeEach(() => {
  parent = document.createElement('div')
  document.body.append(parent)
  refactor = {
    prepareRename: vi.fn(async () => allowed),
    renameSymbol: vi.fn(async () => true),
    findReferences: vi.fn(async () => {})
  }
  const { bridge } = createMockBridge()
  handle = createEditor(
    parent,
    { onChange: () => {}, onCursor: () => {}, onToggleBreakpoint: () => {} },
    {
      language: bridge.language,
      openLocation: () => {},
      refactor
    }
  )
  handle.show(FILE, SOURCE)
  handle.goTo(3, 7) // inside "sumar"
})

afterEach(() => {
  handle.destroy()
  parent.remove()
})

const press = (key: string, shift = false): void => {
  const content = parent.querySelector<HTMLElement>('.cm-content')
  content?.dispatchEvent(new KeyboardEvent('keydown', { key, shiftKey: shift, bubbles: true }))
}

const box = async (): Promise<HTMLInputElement> => {
  await vi.waitFor(() => expect(parent.querySelector('.cm-rename input')).not.toBeNull())
  return parent.querySelector<HTMLInputElement>('.cm-rename input') as HTMLInputElement
}

const type = (input: HTMLInputElement, key: string): void => {
  input.dispatchEvent(new KeyboardEvent('keydown', { key, bubbles: true }))
}

describe('rename in the editor (F2)', () => {
  it('opens an input at the symbol, prefilled with its name', async () => {
    press('F2')
    const input = await box()
    expect(input.value).toBe('sumar')
    expect(input.getAttribute('aria-label')).toBe('Rename symbol')
    expect(refactor.prepareRename).toHaveBeenCalledWith({ file: FILE, line: 3, column: 7 })
  })

  it('renames when Enter is pressed, with the typed name', async () => {
    press('F2')
    const input = await box()
    input.value = '  sumarDos '
    type(input, 'Enter')
    expect(refactor.renameSymbol).toHaveBeenCalledWith(
      { file: FILE, line: 3, column: 7 },
      'sumarDos'
    )
    await vi.waitFor(() => expect(parent.querySelector('.cm-rename')).toBeNull())
  })

  it('does nothing when Escape is pressed or the name did not change', async () => {
    press('F2')
    type(await box(), 'Escape')
    await vi.waitFor(() => expect(parent.querySelector('.cm-rename')).toBeNull())
    press('F2')
    type(await box(), 'Enter')
    expect(refactor.renameSymbol).not.toHaveBeenCalled()
  })

  it('uses the range the server reports instead of the word under the cursor', async () => {
    vi.mocked(refactor.prepareRename).mockResolvedValue({
      refusal: '',
      placeholder: '',
      range: {
        start: { file: FILE, line: 3, column: 6 },
        end: { file: FILE, line: 3, column: 11 }
      }
    })
    press('F2')
    expect((await box()).value).toBe('sumar')
  })

  it('opens nothing when the server refuses (it already said why)', async () => {
    vi.mocked(refactor.prepareRename).mockResolvedValue(null)
    press('F2')
    await vi.waitFor(() => expect(refactor.prepareRename).toHaveBeenCalled())
    await Promise.resolve()
    expect(parent.querySelector('.cm-rename')).toBeNull()
  })
})

describe('Ctrl+Z right after a rename', () => {
  const ctrlZ = (): void => {
    parent
      .querySelector('.cm-content')
      ?.dispatchEvent(new KeyboardEvent('keydown', { key: 'z', ctrlKey: true, bubbles: true }))
  }

  it('undoes the whole rename instead of only this file', () => {
    refactor.undoRename = vi.fn(() => true)
    handle.applyChanges(FILE, [{ from: 19, to: 24, insert: 'suma' }])
    ctrlZ()
    expect(refactor.undoRename).toHaveBeenCalled()
    expect(parent.querySelector('.cm-content')?.textContent).toContain('func suma(') // not undone alone
  })

  it('leaves Ctrl+Z to the editor when there is no fresh rename', () => {
    refactor.undoRename = vi.fn(() => false)
    handle.applyChanges(FILE, [{ from: 19, to: 24, insert: 'suma' }])
    ctrlZ()
    expect(parent.querySelector('.cm-content')?.textContent).toContain('func sumar(')
  })
})

describe('find references in the editor (Shift+F12)', () => {
  it('asks for the uses of the word under the cursor', async () => {
    press('F12', true)
    await vi.waitFor(() =>
      expect(refactor.findReferences).toHaveBeenCalledWith(
        { file: FILE, line: 3, column: 7 },
        'sumar'
      )
    )
  })
})

describe('applying changes to open files', () => {
  it('can be undone with one Ctrl+Z', () => {
    const view = EditorView.findFromDOM(parent) as EditorView
    handle.applyChanges(FILE, [
      { from: 19, to: 24, insert: 'suma' },
      { from: 49, to: 50, insert: 'x' }
    ])
    expect(view.state.doc.toString()).not.toBe(SOURCE)
    undo(view)
    expect(view.state.doc.toString()).toBe(SOURCE)
  })

  it('edits the file on screen as one undoable step and another open file in its state', () => {
    expect(handle.applyChanges(FILE, [{ from: 19, to: 24, insert: 'suma' }])).toBe(true)
    expect(parent.querySelector('.cm-content')?.textContent).toContain('func suma(')
    expect(handle.applyChanges('other.go', [])).toBe(false) // not held by the editor
    handle.show('other.go', 'package other\n')
    handle.show(FILE, SOURCE)
    expect(handle.applyChanges('other.go', [{ from: 8, to: 13, insert: 'otro' }])).toBe(true)
    handle.show('other.go', 'package otro\n') // the buffer text the store keeps
    expect(parent.querySelector('.cm-content')?.textContent).toContain('package otro')
  })
})
