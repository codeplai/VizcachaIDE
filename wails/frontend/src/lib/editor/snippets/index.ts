// Snippets in the completion list: typing `for`, `if`, `def`... offers a template with tab stops.
import {
  snippetCompletion,
  type Completion,
  type CompletionContext,
  type CompletionResult,
  type CompletionSource
} from '@codemirror/autocomplete'
import { syntaxTree } from '@codemirror/language'
import { get } from 'svelte/store'
import type { CodeLanguage } from '../../domain'
import { t } from '../../i18n'
import type { DocumentContext } from '../documentContext'
import { cppSnippets } from './cpp'
import { goSnippets } from './go'
import { pythonSnippets } from './python'
import { rustSnippets } from './rust'
import type { SnippetDef } from './types'

export type { SnippetDef }

export const SNIPPETS: Record<CodeLanguage, SnippetDef[]> = {
  go: goSnippets,
  python: pythonSnippets,
  cpp: cppSnippets,
  rust: rustSnippets
}

const EXTENSIONS: Record<string, CodeLanguage> = {
  go: 'go',
  py: 'python',
  pyw: 'python',
  c: 'cpp',
  cc: 'cpp',
  cpp: 'cpp',
  cxx: 'cpp',
  h: 'cpp',
  hh: 'cpp',
  hpp: 'cpp',
  rs: 'rust'
}

/** The language of a file by its extension (the snippets do not wait for the profiles to load). */
export const snippetLanguage = (path: string): CodeLanguage | null =>
  EXTENSIONS[path.slice(path.lastIndexOf('.') + 1).toLowerCase()] ?? null

type Translate = (key: string) => string

const translator = (): Translate => {
  const translate = get(t)
  return (key) => {
    try {
      return translate(key)
    } catch {
      return key // translations not loaded yet
    }
  }
}

/** Completion items of a snippet list, ranked above the language server's items for the same word. */
export const snippetCompletions = (defs: SnippetDef[], translate: Translate): Completion[] =>
  defs.map((def) =>
    snippetCompletion(def.template, {
      label: def.label,
      type: 'snippet',
      detail: translate('snippets.kind'),
      boost: 10,
      info: () => {
        const box = document.createElement('div')
        box.className = 'cm-doc-tooltip'
        box.textContent = translate(def.description)
        return box
      }
    })
  )

const NOT_CODE = /Comment|String/

/** Typing inside a comment or a string, or after a dot (a member), is not the place for a snippet. */
const inCodePosition = (context: CompletionContext, from: number): boolean => {
  if (from > 0 && context.state.sliceDoc(from - 1, from) === '.') return false
  return !NOT_CODE.test(syntaxTree(context.state).resolveInner(context.pos, -1).name)
}

export const snippetSource =
  (file: DocumentContext): CompletionSource =>
  (context): CompletionResult | null => {
    const path = file.path()
    const language = path ? snippetLanguage(path) : null
    const word = context.matchBefore(/\w+/)
    if (!language || !word || !inCodePosition(context, word.from)) return null
    return {
      from: word.from,
      options: snippetCompletions(SNIPPETS[language], translator()),
      validFor: /^\w*$/
    }
  }
