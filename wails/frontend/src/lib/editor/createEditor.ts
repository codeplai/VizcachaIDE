import { Compartment, EditorSelection, EditorState, type Extension } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import type { LanguageProfile } from '../domain'
import { offsetOf } from './documentContext'
import {
  editorExtensions,
  reportCursor,
  type EditorHandlers,
  type LanguageWiring
} from './extensions'
import { external } from './externalEdit'
import { languageExtensionsFor } from './languageSupport'
import { setMarks, type EditorMarks } from './marks'
import { pushDiagnostics } from './problemLint'
import { minimalChange } from './textChange'
import type { Change } from '../textEdits'

export type { EditorHandlers, LanguageWiring } from './extensions'

export interface EditorHandle {
  /** Shows a file. Each file keeps its own cursor and undo history while it stays open. */
  show: (path: string, text: string) => void
  setMarks: (marks: EditorMarks) => void
  setPhrases: (phrases: Extension) => void
  setFontSize: (pixels: number) => void
  /** Shows or hides the inlay hints (inferred types and parameter names). */
  setInlayHints: (enabled: boolean) => void
  /** Asks the language server for the inlay hints again (it finished analysing the file). */
  refreshInlayHints: () => void
  goTo: (line: number, column: number) => void
  /** Applies changes to an open file (shown or not) as one undoable step; false: not held here. */
  applyChanges: (path: string, changes: Change[]) => boolean
  /** Starts renaming the symbol at the cursor (F2). */
  renameSymbol: () => void
  /** Lists where the symbol at the cursor is used (Shift+F12). */
  findReferences: () => void
  destroy: () => void
}

/** The state of a file: the shared extensions plus the ones of its own language. */
const newFileState = (
  path: string,
  text: string,
  extensions: Extension[],
  profile: LanguageProfile | null
): EditorState =>
  EditorState.create({
    doc: text,
    extensions: [extensions, languageExtensionsFor(path, profile)]
  })

/** Applies changes to the file on screen, or to the saved state of another open file. */
const applyChangesTo = (
  view: EditorView,
  states: Map<string, EditorState>,
  shown: string | null,
  path: string,
  changes: Change[]
): boolean => {
  if (path === shown) {
    view.dispatch({ changes })
    return true
  }
  const state = states.get(path)
  if (!state) return false
  states.set(path, state.update({ changes }).state)
  return true
}

const showMarks = (view: EditorView, marks: EditorMarks): void =>
  view.dispatch({ effects: [setMarks.of(marks), ...pushDiagnostics(view.state, marks.problems)] })

/** Puts the cursor on a 1-based line and column, scrolls to it and focuses the editor. */
const goToIn = (view: EditorView, line: number, column: number): void => {
  const pos = offsetOf(view.state, line, column)
  view.dispatch({
    selection: EditorSelection.cursor(pos),
    effects: EditorView.scrollIntoView(pos, { y: 'center' })
  })
  view.focus()
}

export const createEditor = (
  parent: HTMLElement,
  handlers: EditorHandlers,
  wiring: LanguageWiring | null = null,
  /** The profile of a file's language, for its syntax and indentation. */
  profileFor: (path: string) => LanguageProfile | null = () => null
): EditorHandle => {
  const phrases = new Compartment()
  const states = new Map<string, EditorState>()
  let currentPath: string | null = null
  let currentPhrases: Extension = []
  const { extensions, inlay, refactor } = editorExtensions(
    handlers,
    phrases,
    wiring,
    () => currentPath
  )
  const view = new EditorView({ parent, state: EditorState.create({ extensions }) })

  const applyText = (text: string): void => {
    const change = minimalChange(view.state.doc.toString(), text)
    if (change) view.dispatch({ changes: change, annotations: external.of(true) })
  }

  const switchTo = (path: string, text: string): void => {
    if (currentPath) states.set(currentPath, view.state)
    currentPath = path
    view.setState(states.get(path) ?? newFileState(path, text, extensions, profileFor(path)))
    view.dispatch({ effects: phrases.reconfigure(currentPhrases) })
    reportCursor(view.state, handlers)
  }

  return {
    applyChanges: (path, changes) => applyChangesTo(view, states, currentPath, path, changes),
    renameSymbol: () => refactor?.rename(view),
    findReferences: () => refactor?.references(view),
    show: (path, text) => {
      if (path !== currentPath) switchTo(path, text)
      applyText(text)
    },
    setMarks: (marks) => showMarks(view, marks),
    setPhrases: (next) => {
      currentPhrases = next
      view.dispatch({ effects: phrases.reconfigure(next) })
    },
    setFontSize: (pixels) => parent.style.setProperty('--editor-font-size', `${pixels}px`),
    setInlayHints: (enabled) => inlay?.setEnabled(enabled),
    refreshInlayHints: () => inlay?.refresh(),
    goTo: (line, column) => goToIn(view, line, column),
    destroy: () => view.destroy()
  }
}
