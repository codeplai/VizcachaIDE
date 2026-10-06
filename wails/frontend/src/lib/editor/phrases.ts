// Translates the texts CodeMirror draws itself (search panel, go to line, suggestions list).
import { EditorState, type Extension } from '@codemirror/state'

/** CodeMirror's English phrase -> our i18n key (namespace editor.*). */
const PHRASE_KEYS: Record<string, string> = {
  Find: 'editor.search.find',
  Replace: 'editor.search.replace',
  next: 'editor.search.next',
  previous: 'editor.search.previous',
  all: 'editor.search.all',
  'match case': 'editor.search.matchCase',
  regexp: 'editor.search.regexp',
  'by word': 'editor.search.byWord',
  replace: 'editor.search.replace',
  'replace all': 'editor.search.replaceAll',
  close: 'editor.search.close',
  'Go to line': 'editor.search.goToLine',
  go: 'editor.search.go',
  'current match': 'editor.search.currentMatch',
  'on line': 'editor.search.onLine',
  'replaced $ matches': 'editor.search.replacedMatches',
  'replaced match on line $': 'editor.search.replacedMatch',
  Completions: 'editor.completions',
  'Rename symbol': 'editor.rename.label',
  'Enter to rename, Esc to cancel': 'editor.rename.hint'
}

export type Translate = (key: string) => string

export const phrasesFor = (translate: Translate): Record<string, string> =>
  Object.fromEntries(Object.entries(PHRASE_KEYS).map(([phrase, key]) => [phrase, translate(key)]))

export const editorPhrases = (translate: Translate): Extension =>
  EditorState.phrases.of(phrasesFor(translate))
