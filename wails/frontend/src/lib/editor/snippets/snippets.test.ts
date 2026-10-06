import {
  CompletionContext,
  currentCompletions,
  startCompletion,
  type Completion,
  type CompletionResult
} from '@codemirror/autocomplete'
import { EditorState } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { get } from 'svelte/store'
import { beforeAll, describe, expect, it, vi } from 'vitest'
import { createMockBridge } from '../../bridge/mock'
import type { LanguageProfile } from '../../domain'
import { applyLanguage, setupI18n, t } from '../../i18n'
import en from '../../i18n/locales/en.json'
import es from '../../i18n/locales/es.json'
import type { DocumentContext } from '../documentContext'
import { editorCompletion } from '../completions'
import { languageExtensionsFor } from '../languageSupport'
import { SNIPPETS, snippetLanguage, snippetSource } from '.'

beforeAll(() => {
  setupI18n('en')
  // jsdom has no layout: CodeMirror measures text with these.
  Range.prototype.getClientRects = () => [] as unknown as DOMRectList
  Range.prototype.getBoundingClientRect = () => new DOMRect()
})

const FILES = { go: 'main.go', python: 'main.py', cpp: 'main.cpp', rust: 'main.rs' } as const

const EXPECTED: Record<keyof typeof FILES, string[]> = {
  go: ['for', 'range', 'if', 'while', 'func', 'main', 'println', 'struct', 'switch'],
  python: ['for', 'if', 'while', 'def', 'main', 'print', 'class', 'try', 'match'],
  cpp: [
    'for',
    'foreach',
    'if',
    'while',
    'func',
    'main',
    'cout',
    'class',
    'struct',
    'switch',
    'try'
  ],
  rust: ['for', 'if', 'while', 'fn', 'main', 'println', 'struct', 'match']
}

const fileOf = (path: string): DocumentContext => ({ path: () => path, flush: async () => {} })

const profileOf = (id: keyof typeof FILES): LanguageProfile =>
  ({ id, indent: { useTabs: false, size: 4 } }) as LanguageProfile

const stateOf = (id: keyof typeof FILES, doc: string): EditorState =>
  EditorState.create({ doc, extensions: languageExtensionsFor(FILES[id], profileOf(id)) })

const ask = async (id: keyof typeof FILES, doc: string): Promise<CompletionResult | null> => {
  const state = stateOf(id, doc)
  return (await snippetSource(fileOf(FILES[id]))(
    new CompletionContext(state, doc.length, false)
  )) as CompletionResult | null
}

describe('snippets per language', () => {
  it('knows the language of a file by its extension', () => {
    expect(snippetLanguage('a/b/main.go')).toBe('go')
    expect(snippetLanguage('x.PY')).toBe('python')
    expect(snippetLanguage('x.hpp')).toBe('cpp')
    expect(snippetLanguage('x.rs')).toBe('rust')
    expect(snippetLanguage('notes.txt')).toBeNull()
  })

  for (const [id, labels] of Object.entries(EXPECTED) as [keyof typeof FILES, string[]][]) {
    it(`${id} offers its snippets, typed as snippet with a description`, async () => {
      expect(SNIPPETS[id].map((snippet) => snippet.label)).toEqual(labels)
      const result = await ask(id, 'fo')
      expect(result?.from).toBe(0)
      const options = result?.options ?? []
      expect(options.map((option) => option.label)).toEqual(labels)
      for (const option of options) {
        expect(option.type).toBe('snippet')
        expect(option.detail).toBe('snippet')
      }
    })
  }

  it('every description exists in English and Spanish, and they differ', () => {
    const english = en as Record<string, string>
    const spanish = es as Record<string, string>
    const keys = new Set(Object.values(SNIPPETS).flatMap((list) => list.map((s) => s.description)))
    for (const key of keys) {
      expect(english[key], key).toBeTruthy()
      expect(spanish[key], key).toBeTruthy()
      expect(spanish[key]).not.toBe(english[key])
    }
    expect(english['snippets.forCount']).toBe('for: repeat a block a number of times')
  })

  it('shows the description in the language of the IDE', async () => {
    const describeFor = async (language: 'en' | 'es'): Promise<string> => {
      applyLanguage(language)
      const result = await ask('python', 'wh')
      const option = result?.options.find((item) => item.label === 'while') as Completion
      const info = option.info as () => HTMLElement
      return info().textContent ?? ''
    }
    expect(await describeFor('en')).toBe('while: repeat a block as long as a condition is true')
    expect(await describeFor('es')).toBe('while: repite un bloque mientras una condición se cumpla')
    applyLanguage('en')
    expect(get(t)('snippets.kind')).toBe('snippet')
  })

  it('stays quiet in comments, strings, after a dot and on empty space', async () => {
    expect(await ask('go', '// fo')).toBeNull()
    expect(await ask('go', 'x := "fo')).toBeNull()
    expect(await ask('python', 'obj.fo')).toBeNull()
    expect(await ask('python', '# fo')).toBeNull()
    expect(await ask('go', 'x := ')).toBeNull()
  })

  it('inserts the template with tab stops, the first one selected', async () => {
    const result = await ask('go', 'for')
    const option = result?.options.find((item) => item.label === 'for') as Completion
    const view = new EditorView({ state: stateOf('go', 'for') })
    ;(option.apply as (...args: unknown[]) => void)(view, option, 0, 3)
    // The test file's profile indents with four spaces: the template's tab became that.
    expect(view.state.doc.toString()).toBe('for i := 0; i < n; i++ {\n    \n}')
    const { from, to } = view.state.selection.main
    expect(view.state.sliceDoc(from, to)).toBe('i')
    view.destroy()
  })
})

describe('merging with the language server', () => {
  it('lists both the language items and the snippet, snippet first for an exact word', async () => {
    const { bridge } = createMockBridge()
    vi.spyOn(bridge.language, 'completion').mockResolvedValue([
      { label: 'format', kind: 'function', detail: '', documentation: '', insertText: '' },
      { label: 'for', kind: 'keyword', detail: '', documentation: '', insertText: '' }
    ])
    const view = new EditorView({
      parent: document.body,
      state: EditorState.create({
        doc: 'for',
        selection: { anchor: 3 },
        extensions: [
          languageExtensionsFor('main.go', null),
          editorCompletion(bridge.language, fileOf('main.go'))
        ]
      })
    })
    startCompletion(view)
    await vi.waitFor(() => expect(currentCompletions(view.state).length).toBeGreaterThan(1))
    const all = currentCompletions(view.state)
    const labels = all.map((item) => `${item.type}:${item.label}`)
    expect(labels).toContain('snippet:for')
    expect(labels).toContain('keyword:for')
    expect(labels).toContain('function:format')
    expect(labels[0]).toBe('snippet:for')
    view.destroy()
  })

  it('still offers the snippets when there is no language service', async () => {
    const view = new EditorView({
      parent: document.body,
      state: EditorState.create({
        doc: 'whi',
        selection: { anchor: 3 },
        extensions: [
          languageExtensionsFor('main.py', profileOf('python')),
          editorCompletion(null, fileOf('main.py'))
        ]
      })
    })
    startCompletion(view)
    await vi.waitFor(() => expect(currentCompletions(view.state).length).toBe(1))
    expect(currentCompletions(view.state)[0]?.label).toBe('while')
    view.destroy()
  })
})
