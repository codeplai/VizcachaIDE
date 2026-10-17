import { pickedCompletion, type Completion, type CompletionSource } from '@codemirror/autocomplete'
import type { CompletionItem, CompletionKind } from '../domain'
import { locationAt, type DocumentContext, type LanguageApi } from './documentContext'

const COMPLETION_TYPES: Record<CompletionKind, string> = {
  keyword: 'keyword',
  function: 'function',
  method: 'method',
  variable: 'variable',
  constant: 'constant',
  field: 'property',
  type: 'type',
  package: 'namespace',
  other: 'text'
}

/**
 * Inserts the text and leaves the cursor right after it, back in the editor: the list can take
 * the focus (a click on it), and an insertion alone would not say where the cursor goes.
 */
const insertItem =
  (text: string): Completion['apply'] =>
  (view, completion, from, to) => {
    view.dispatch({
      changes: { from, to, insert: text },
      selection: { anchor: from + text.length },
      scrollIntoView: true,
      userEvent: 'input.complete',
      annotations: pickedCompletion.of(completion)
    })
    view.focus()
  }

export const toCompletion = (item: CompletionItem): Completion => ({
  label: item.label,
  type: COMPLETION_TYPES[item.kind],
  detail: item.detail || undefined,
  apply: insertItem(item.insertText || item.label),
  info: item.documentation
    ? () => {
        const box = document.createElement('div')
        box.className = 'cm-doc-tooltip'
        box.textContent = item.documentation
        return box
      }
    : undefined
})

/**
 * Asks gopls for suggestions. Typing a word or a "." opens the list on its own,
 * and Ctrl+Space opens it anywhere.
 */
export const goCompletionSource =
  (language: LanguageApi, file: DocumentContext): CompletionSource =>
  async (context) => {
    const path = file.path()
    const word = context.matchBefore(/\w*/)
    if (!path || !word) return null
    const afterDot = word.from > 0 && context.state.sliceDoc(word.from - 1, word.from) === '.'
    if (!context.explicit && word.text === '' && !afterDot) return null
    await file.flush()
    const items = await language.completion(locationAt(context.state, context.pos, path))
    if (context.aborted || items.length === 0) return null
    return { from: word.from, options: items.map(toCompletion), validFor: /^\w*$/ }
  }
