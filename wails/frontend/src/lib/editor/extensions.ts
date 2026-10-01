// The editor's extension list, in one place.
import { closeBrackets, closeBracketsKeymap, completionKeymap } from '@codemirror/autocomplete'
import { defaultKeymap, history, historyKeymap, indentWithTab } from '@codemirror/commands'
import { go } from '@codemirror/lang-go'
import { bracketMatching, indentUnit } from '@codemirror/language'
import { lintKeymap } from '@codemirror/lint'
import { gotoLine, search, searchKeymap } from '@codemirror/search'
import { EditorState, Prec, type Compartment, type Extension } from '@codemirror/state'
import { EditorView, drawSelection, keymap, lineNumbers } from '@codemirror/view'
import { breakpointGutter } from './breakpointGutter'
import type { DocumentContext, LanguageApi } from './documentContext'
import { isExternalEdit } from './externalEdit'
import { documentSync } from './documentSync'
import { goCompletion } from './goCompletion'
import { goToDefinition, type OpenLocation } from './goToDefinition'
import { hoverDocs } from './hoverDocs'
import { inlineHints } from './inlineHint'
import { emptyMarks, marksField } from './marks'
import { problemLint } from './problemLint'
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
}

/** Go uses real tabs; Enter between braces puts the closing brace on its own line. */
const goIndentation: Extension = [indentUnit.of('\t'), EditorState.tabSize.of(4), closeBrackets()]

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

const languageExtensions = (wiring: LanguageWiring | null, file: DocumentContext): Extension[] => {
  if (!wiring) return []
  const { language, openLocation } = wiring
  return [
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
}

export const editorExtensions = (
  handlers: EditorHandlers,
  phrases: Compartment,
  wiring: LanguageWiring | null,
  getPath: () => string | null
): EditorExtensions => {
  const sync = wiring ? documentSync(wiring.language, getPath) : null
  const file: DocumentContext = { path: getPath, flush: sync?.flush ?? (async () => {}) }
  const extensions = [
    marksField.init(() => emptyMarks),
    phrases.of([]),
    breakpointGutter(handlers.onToggleBreakpoint),
    lineNumbers(),
    history(),
    drawSelection(),
    bracketMatching(),
    goIndentation,
    search({ top: true }),
    problemLint,
    keys,
    keymap.of([...defaultKeymap, ...historyKeymap]),
    go(),
    editorTheme,
    popupTheme,
    editorHighlighting,
    inlineHints,
    ...(sync ? [sync.extension] : []),
    ...languageExtensions(wiring, file),
    EditorView.updateListener.of((update) => {
      if (update.docChanged && !isExternalEdit(update))
        handlers.onChange(update.state.doc.toString())
      if (update.selectionSet || update.docChanged) reportCursor(update.state, handlers)
    })
  ]
  return { extensions, flush: file.flush }
}
