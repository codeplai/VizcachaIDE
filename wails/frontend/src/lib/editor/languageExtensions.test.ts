import { CompletionContext } from '@codemirror/autocomplete'
import { EditorState } from '@codemirror/state'
import { describe, expect, it, vi } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import type { CompletionItem, SignatureHelp } from '../domain'
import { locationAt, offsetOf, type DocumentContext } from './documentContext'
import { goCompletionSource, toCompletion } from './lspCompletion'
import type { ProblemMark } from './marks'
import { toLintDiagnostics } from './problemLint'
import { renderSignature } from './signatureHelp'

const file: DocumentContext = { path: () => 'main.go', flush: async () => {} }
const item = (label: string, kind: CompletionItem['kind'] = 'function'): CompletionItem => ({
  label,
  kind,
  detail: 'func()',
  documentation: 'Docs',
  insertText: ''
})

describe('positions', () => {
  it('converts offsets to 1-based lines and columns and back', () => {
    const state = EditorState.create({ doc: 'ab\ncde\n' })
    expect(locationAt(state, 4, 'main.go')).toEqual({ file: 'main.go', line: 2, column: 2 })
    expect(offsetOf(state, 2, 2)).toBe(4)
    expect(offsetOf(state, 99, 99)).toBe(state.doc.length)
  })

  it('counts columns in characters, as the backend does: an emoji is one column', () => {
    const state = EditorState.create({ doc: 'x := "😀" + y\n' })
    const y = state.doc.toString().indexOf('y') // UTF-16 offset: the emoji takes two units
    expect(locationAt(state, y, 'main.go')).toEqual({ file: 'main.go', line: 1, column: 12 })
    expect(offsetOf(state, 1, 12)).toBe(y)
  })
})

describe('completion', () => {
  it('maps gopls kinds to CodeMirror types', () => {
    expect(toCompletion(item('Println')).type).toBe('function')
    expect(toCompletion(item('Name', 'field')).type).toBe('property')
    expect(toCompletion(item('fmt', 'package')).type).toBe('namespace')
  })

  const sourceFor = (doc: string, items: CompletionItem[]) => {
    const { bridge } = createMockBridge()
    const completion = vi.spyOn(bridge.language, 'completion').mockResolvedValue(items)
    const source = goCompletionSource(bridge.language, file)
    const state = EditorState.create({ doc })
    const ask = (explicit: boolean) => source(new CompletionContext(state, doc.length, explicit))
    return { ask, completion }
  }

  it('asks gopls while a word is typed and offers from the word start', async () => {
    const { ask, completion } = sourceFor('fmt.Pri', [item('Println')])
    const result = await ask(false)
    expect(completion).toHaveBeenCalledWith({ file: 'main.go', line: 1, column: 8 })
    expect(result && 'options' in result && result.from).toBe(4)
  })

  it('opens right after a dot', async () => {
    const { ask, completion } = sourceFor('fmt.', [item('Println')])
    expect(await ask(false)).not.toBeNull()
    expect(completion).toHaveBeenCalled()
  })

  it('stays quiet on empty space unless Ctrl+Space was pressed', async () => {
    const { ask, completion } = sourceFor('x := ', [item('Println')])
    expect(await ask(false)).toBeNull()
    expect(completion).not.toHaveBeenCalled()
    expect(await ask(true)).not.toBeNull()
  })
})

describe('lint diagnostics', () => {
  const mark = (change: Partial<ProblemMark>): ProblemMark => ({
    line: 2,
    from: 5,
    to: 14,
    severity: 'error',
    title: 'Title',
    detail: 'raw message',
    ...change
  })
  const state = EditorState.create({ doc: 'package main\n    resultado := 1\n' })

  it('underlines the exact range of the problem', () => {
    const [diagnostic] = toLintDiagnostics(state, [mark({})])
    expect(state.sliceDoc(diagnostic?.from, diagnostic?.to)).toBe('resultado')
    expect(diagnostic?.severity).toBe('error')
    expect(diagnostic?.message).toBe('Title\nraw message')
  })

  it('gives zero-width ranges the rest of the word and skips unknown lines', () => {
    const [diagnostic] = toLintDiagnostics(state, [mark({ to: 5 })])
    expect(state.sliceDoc(diagnostic?.from, diagnostic?.to)).toBe('resultado')
    expect(toLintDiagnostics(state, [mark({ line: 40 })])).toEqual([])
  })

  it('shows a single line when the title is the message itself', () => {
    const [diagnostic] = toLintDiagnostics(state, [mark({ title: 'same', detail: 'same' })])
    expect(diagnostic?.message).toBe('same')
  })
})

describe('signature help', () => {
  const help: SignatureHelp = {
    label: 'func Println(a ...any) (n int, err error)',
    documentation: '',
    parameters: ['a ...any'],
    activeParameter: 0
  }

  it('bolds the parameter being typed', () => {
    const dom = renderSignature(help)
    expect(dom.querySelector('b')?.textContent).toBe('a ...any')
    expect(dom.textContent).toBe(help.label)
  })

  it('shows the plain label when the parameter is not found', () => {
    const dom = renderSignature({ ...help, activeParameter: 3 })
    expect(dom.querySelector('b')).toBeNull()
  })
})
