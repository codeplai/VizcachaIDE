// The editor's extension list, in one place.
import { closeBrackets, closeBracketsKeymap, completionKeymap } from '@codemirror/autocomplete'
import { defaultKeymap, history, historyKeymap, indentWithTab } from '@codemirror/commands'
import { bracketMatching } from '@codemirror/language'
import { lintKeymap } from '@codemirror/lint'
import { gotoLine, search, searchKeymap } from '@codemirror/search'
import { Prec, type Compartment, type EditorState, type Extension } from '@codemirror/state'
import { EditorView, drawSelection, keymap, lineNumbers } from '@codemirror/view'
import { breakpointGutter } from './breakpointGutter'
import type { DocumentContext, LanguageApi } from './documentContext'
import { isExternalEdit } from './externalEdit'
import { documentSync } from './documentSync'
import { goCompletion } from './lspCompletion'
import { goToDefinition, type OpenLocation } from './goToDefinition'
import { hoverDocs } from './hoverDocs'
import { inlayHints, type InlayHints } from './inlayHints'
import { inlineHints } from './inlineHint'
import { emptyMarks, marksField } from './marks'
import { problemLint } from './problemLint'
import { refactoring, type RefactorCommands, type RefactorWiring } from './refactor'
import { popupTheme } from './popupTheme'
import { signatureHelp } from './signatureHelp'
import { editorHighlighting, editorTheme } from './theme'

export interface EditorHandlers {
  onChange: (text: string) => void
  onCursor: (line: number, column: number) => void
  onToggleBreakpoint: (line: number) => void
}

export interface LanguageWiring {
  language: LanguageApi
  openLocation: OpenLocation
  /** Rename (F2) and find references (Shift+F12); absent in editors without a backend. */
  refactor?: RefactorWiring
}

const keys = Prec.high(
  keymap.of([
    { key: 'Mod-g', run: gotoLine, preventDefault: true },
    ...closeBracketsKeymap,
    ...completionKeymap,
    ...searchKeymap.filter((binding) => binding.key !== 'Mod-g'),
    ...lintKeymap,
    indentWithTab
  ])
)

const languageExtensions = (
  wiring: LanguageWiring | null,
  file: DocumentContext,
  inlay: InlayHints | null,
  refactor: RefactorCommands | null
): Extension[] => {
  if (!wiring) return []
  const { language, openLocation } = wiring
  return [
    ...(inlay ? [inlay.extension] : []),
    ...(refactor ? [refactor.extension] : []),
    goCompletion(language, file),
    hoverDocs(language, file),
    signatureHelp(language, file),
    goToDefinition(language, file, openLocation)
  ]
}

export const reportCursor = (state: EditorState, handlers: EditorHandlers): void => {
  const head = state.selection.main.head
  const line = state.doc.lineAt(head)
  handlers.onCursor(line.number, head - line.from + 1)
}

export interface EditorExtensions {
  extensions: Extension[]
  /** Sends pending edits to gopls. */
  flush: () => Promise<void>
  /** The inlay hints, when the editor has a language service. */
  inlay: InlayHints | null
  /** Rename and find references, when the editor has a language service. */
  refactor: RefactorCommands | null
}

export const editorExtensions = (
  handlers: EditorHandlers,
  phrases: Compartment,
  wiring: LanguageWiring | null,
  getPath: () => string | null
): EditorExtensions => {
  const sync = wiring ? documentSync(wiring.language, getPath) : null
  const file: DocumentContext = { path: getPath, flush: sync?.flush ?? (async () => {}) }
  const inlay = wiring ? inlayHints(wiring.language, file) : null
  const refactor = wiring?.refactor ? refactoring(file, wiring.refactor) : null
  const extensions = [
    marksField.init(() => emptyMarks),
    phrases.of([]),
    breakpointGutter(handlers.onToggleBreakpoint),
    lineNumbers(),
    history(),
    drawSelection(),
    bracketMatching(),
    closeBrackets(),
    // Replace all runs CodeMirror's own command; asking `confirmReplaceAll` first would need a
    // custom search panel, so the shell's confirmation is not connected to it (see the report).
    search({ top: true }),
    problemLint,
    keys,
    keymap.of([...defaultKeymap, ...historyKeymap]),
    editorTheme,
    popupTheme,
    editorHighlighting,
    inlineHints,
    ...(sync ? [sync.extension] : []),
    ...languageExtensions(wiring, file, inlay, refactor),
    EditorView.updateListener.of((update) => {
      if (update.docChanged && !isExternalEdit(update))
        handlers.onChange(update.state.doc.toString())
      if (update.selectionSet || update.docChanged) reportCursor(update.state, handlers)
    })
  ]
  return { extensions, flush: file.flush, inlay, refactor }
}
