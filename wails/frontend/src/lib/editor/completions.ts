// The editor's suggestions: the language server's items and the snippets, in one list.
import { autocompletion, type CompletionSource } from '@codemirror/autocomplete'
import type { Extension } from '@codemirror/state'
import type { DocumentContext, LanguageApi } from './documentContext'
import { goCompletionSource } from './lspCompletion'
import { snippetSource } from './snippets'

/**
 * Both sources answer together and CodeMirror merges and ranks what they return, so a snippet
 * never hides a language-server item. Without a language service only the snippets are offered.
 */
export const editorCompletion = (
  language: LanguageApi | null,
  file: DocumentContext
): Extension => {
  const sources: CompletionSource[] = [snippetSource(file)]
  if (language) sources.unshift(goCompletionSource(language, file))
  return autocompletion({ override: sources, icons: true })
}
