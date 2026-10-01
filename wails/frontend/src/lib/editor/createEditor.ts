import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { go } from '@codemirror/lang-go'
import { Annotation, EditorState } from '@codemirror/state'
import { EditorView, keymap, lineNumbers } from '@codemirror/view'
import {
  breakpointGutter,
  emptyMarks,
  marksDecorations,
  marksField,
  setMarks,
  type EditorMarks
} from './marks'
import { editorHighlighting, editorTheme } from './theme'

/** Marks transactions that come from the app, so they are not echoed back as user edits. */
const external = Annotation.define<boolean>()

export interface EditorHandlers {
  onChange: (text: string) => void
  onCursor: (line: number, column: number) => void
  onToggleBreakpoint: (line: number) => void
}

export interface EditorHandle {
  setText: (text: string) => void
  setMarks: (marks: EditorMarks) => void
  destroy: () => void
}

export const createEditor = (
  parent: HTMLElement,
  initialText: string,
  handlers: EditorHandlers
): EditorHandle => {
  const view = new EditorView({
    parent,
    state: EditorState.create({
      doc: initialText,
      extensions: [
        marksField.init(() => emptyMarks),
        breakpointGutter(handlers.onToggleBreakpoint),
        lineNumbers(),
        history(),
        keymap.of([...defaultKeymap, ...historyKeymap]),
        go(),
        editorTheme,
        editorHighlighting,
        marksDecorations,
        EditorView.updateListener.of((update) => {
          if (update.docChanged && !update.transactions.some((tr) => tr.annotation(external))) {
            handlers.onChange(update.state.doc.toString())
          }
          if (update.selectionSet || update.docChanged) reportCursor(update.state, handlers)
        })
      ]
    })
  })
  return {
    setText: (text) => {
      if (text === view.state.doc.toString()) return
      view.dispatch({
        changes: { from: 0, to: view.state.doc.length, insert: text },
        annotations: external.of(true)
      })
    },
    setMarks: (marks) => view.dispatch({ effects: setMarks.of(marks) }),
    destroy: () => view.destroy()
  }
}

const reportCursor = (state: EditorState, handlers: EditorHandlers): void => {
  const head = state.selection.main.head
  const line = state.doc.lineAt(head)
  handlers.onCursor(line.number, head - line.from + 1)
}
