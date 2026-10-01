import { Compartment, EditorSelection, EditorState, type Extension } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { offsetOf } from './documentContext'
import {
  editorExtensions,
  reportCursor,
  type EditorHandlers,
  type LanguageWiring
} from './extensions'
import { external } from './externalEdit'
import { setMarks, type EditorMarks } from './marks'
import { pushDiagnostics } from './problemLint'
import { minimalChange } from './textChange'

export type { EditorHandlers, LanguageWiring } from './extensions'

export interface EditorHandle {
  /** Shows a file. Each file keeps its own cursor and undo history while it stays open. */
  show: (path: string, text: string) => void
  setMarks: (marks: EditorMarks) => void
  setPhrases: (phrases: Extension) => void
  setFontSize: (pixels: number) => void
  goTo: (line: number, column: number) => void
  forget: (path: string) => void
  destroy: () => void
}

export const createEditor = (
  parent: HTMLElement,
  handlers: EditorHandlers,
  wiring: LanguageWiring | null = null
): EditorHandle => {
  const phrases = new Compartment()
  const states = new Map<string, EditorState>()
  let currentPath: string | null = null
  let currentPhrases: Extension = []
  const { extensions } = editorExtensions(handlers, phrases, wiring, () => currentPath)
  const view = new EditorView({ parent, state: EditorState.create({ extensions }) })

  const applyText = (text: string): void => {
    const change = minimalChange(view.state.doc.toString(), text)
    if (change) view.dispatch({ changes: change, annotations: external.of(true) })
  }

  const switchTo = (path: string, text: string): void => {
    if (currentPath) states.set(currentPath, view.state)
    currentPath = path
    view.setState(states.get(path) ?? EditorState.create({ doc: text, extensions }))
    view.dispatch({ effects: phrases.reconfigure(currentPhrases) })
    reportCursor(view.state, handlers)
  }

  return {
    show: (path, text) => {
      if (path !== currentPath) switchTo(path, text)
      applyText(text)
    },
    setMarks: (marks) =>
      view.dispatch({
        effects: [setMarks.of(marks), ...pushDiagnostics(view.state, marks.problems)]
      }),
    setPhrases: (next) => {
      currentPhrases = next
      view.dispatch({ effects: phrases.reconfigure(next) })
    },
    setFontSize: (pixels) => parent.style.setProperty('--editor-font-size', `${pixels}px`),
    goTo: (line, column) => {
      const pos = offsetOf(view.state, line, column)
      view.dispatch({
        selection: EditorSelection.cursor(pos),
        effects: EditorView.scrollIntoView(pos, { y: 'center' })
      })
      view.focus()
    },
    forget: (path) => states.delete(path),
    destroy: () => view.destroy()
  }
}
