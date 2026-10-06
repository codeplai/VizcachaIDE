import { insertNewlineAndIndent } from '@codemirror/commands'
import { EditorState } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import { createEditor, type EditorHandle, type EditorHandlers } from './createEditor'
import { languageWiring } from './wiring'

let host: HTMLDivElement
let editor: EditorHandle
let handlers: EditorHandlers

const viewOf = (): EditorView => {
  const view = EditorView.findFromDOM(host)
  if (!view) throw new Error('no editor view')
  return view
}

beforeEach(() => {
  host = document.createElement('div')
  document.body.append(host)
  handlers = { onChange: vi.fn(), onCursor: vi.fn(), onToggleBreakpoint: vi.fn() }
})

afterEach(() => {
  editor.destroy()
  host.remove()
  vi.useRealTimers()
})

describe('createEditor', () => {
  it('shows a file without reporting it as a user edit', () => {
    editor = createEditor(host, handlers)
    editor.show('a.go', 'package main\n')
    expect(viewOf().state.doc.toString()).toBe('package main\n')
    expect(handlers.onChange).not.toHaveBeenCalled()
  })

  it('keeps each file with its own text and cursor when switching tabs', () => {
    editor = createEditor(host, handlers)
    editor.show('a.go', 'one\ntwo\n')
    editor.goTo(2, 2)
    editor.show('b.go', 'other')
    expect(viewOf().state.doc.toString()).toBe('other')
    editor.show('a.go', 'one\ntwo\n')
    expect(viewOf().state.selection.main.head).toBe(5)
  })

  it('applies new text from outside as a small edit that keeps the cursor', () => {
    editor = createEditor(host, handlers)
    editor.show('a.go', 'func f() {\n    x()\n}')
    editor.goTo(3, 2)
    editor.show('a.go', 'func f() {\n\tx()\n}')
    expect(viewOf().state.doc.toString()).toBe('func f() {\n\tx()\n}')
    expect(viewOf().state.doc.lineAt(viewOf().state.selection.main.head).number).toBe(3)
    expect(handlers.onChange).not.toHaveBeenCalled()
  })

  it('allows several selections (snippet fields), added with Alt+click, not Ctrl+click', () => {
    editor = createEditor(host, handlers)
    editor.show('a.go', 'package main\n')
    const state = viewOf().state
    expect(state.facet(EditorState.allowMultipleSelections)).toBe(true)
    const adds = state.facet(EditorView.clickAddsSelectionRange)
    expect(adds.some((add) => add(new MouseEvent('mousedown', { altKey: true })))).toBe(true)
    expect(adds.some((add) => add(new MouseEvent('mousedown', { ctrlKey: true })))).toBe(false)
  })

  it('reports the cursor position with 1-based line and column', () => {
    editor = createEditor(host, handlers)
    editor.show('a.go', 'ab\ncde')
    editor.goTo(2, 3)
    expect(handlers.onCursor).toHaveBeenLastCalledWith(2, 3)
  })

  it('indents with a tab after "{" and puts the closing brace on its own line', () => {
    editor = createEditor(host, handlers)
    editor.show('a.go', 'package main\n\nfunc main() {}')
    const view = viewOf()
    view.dispatch({ selection: { anchor: view.state.doc.length - 1 } })
    insertNewlineAndIndent(view)
    expect(view.state.doc.toString()).toBe('package main\n\nfunc main() {\n\t\n}')
  })

  it('draws problems as lint ranges and an inline explanation', () => {
    editor = createEditor(host, handlers)
    editor.show('a.go', 'x := 1\nresultado := 2\n')
    editor.setMarks({
      debugLine: 2,
      breakpoints: [1],
      problems: [
        { line: 2, from: 1, to: 10, severity: 'error', title: 'Never used', detail: 'raw' }
      ]
    })
    expect(host.querySelector('.cm-lintRange-error')?.textContent).toBe('resultado')
    expect(host.querySelector('.cm-inline-hint')?.textContent).toContain('Never used')
    expect(host.querySelector('.cm-debug-line')).not.toBeNull()
    expect(host.querySelector('.cm-bp')).not.toBeNull()
  })

  it('applies the zoom through a CSS variable', () => {
    editor = createEditor(host, handlers)
    editor.setFontSize(18)
    expect(host.style.getPropertyValue('--editor-font-size')).toBe('18px')
  })
})

describe('createEditor with the language service', () => {
  it('sends edits to gopls after a short pause', async () => {
    vi.useFakeTimers()
    const { bridge } = createMockBridge()
    const change = vi.spyOn(bridge.language, 'changeDocument')
    editor = createEditor(host, handlers, languageWiring(bridge))
    editor.show('a.go', 'package main\n')
    viewOf().dispatch({ changes: { from: 0, insert: '// hi\n' } })
    expect(change).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(200)
    expect(change).toHaveBeenCalledWith('a.go', '// hi\npackage main\n', 1)
  })
})
