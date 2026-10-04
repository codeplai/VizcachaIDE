import { EditorState } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { InlayHint, SourceRange } from '../domain'
import { sampleInlayHints } from '../bridge/mockInlay'
import { SAMPLE_MAIN } from '../bridge/mockData'
import type { LanguageApi } from './documentContext'
import { INLAY_DELAY_MS, inlayHints } from './inlayHints'

const typeHint = (line: number, column: number, label = ': int'): InlayHint => ({
  line,
  column,
  label,
  kind: 'type',
  paddingLeft: true,
  paddingRight: false
})

let view: EditorView
let host: HTMLDivElement

const setup = (inlayHintsFn: LanguageApi['inlayHints'], doc = 'x := 5\ny := 6\n') => {
  const language = { inlayHints: inlayHintsFn } as LanguageApi
  const hints = inlayHints(language, { path: () => 'a.go', flush: async () => {} })
  host = document.createElement('div')
  document.body.append(host)
  view = new EditorView({
    parent: host,
    state: EditorState.create({ doc, extensions: [hints.extension] })
  })
  return hints
}

const settle = async (ms = INLAY_DELAY_MS): Promise<void> => {
  await vi.advanceTimersByTimeAsync(ms)
}

const shown = (): string[] =>
  [...host.querySelectorAll('.cm-inlay-hint')].map((el) => el.textContent ?? '')

beforeEach(() => vi.useFakeTimers())
afterEach(() => {
  view.destroy()
  host.remove()
  vi.useRealTimers()
})

describe('inlayHints', () => {
  it('asks once, after the delay, for the visible lines of the file and draws the hints', async () => {
    const ask = vi.fn(async (_visible: SourceRange) => [typeHint(1, 2)])
    setup(ask)
    await settle(INLAY_DELAY_MS - 1)
    expect(ask).not.toHaveBeenCalled()
    await settle(1)
    expect(ask).toHaveBeenCalledTimes(1)
    expect(ask.mock.calls[0]?.[0]).toEqual({
      start: { file: 'a.go', line: 1, column: 1 },
      end: { file: 'a.go', line: 3, column: 1 }
    })
    expect(shown()).toEqual([': int'])
    expect(host.querySelector('.cm-inlay-pad-left')).not.toBeNull()
    expect(view.state.doc.toString()).toBe('x := 5\ny := 6\n') // the code is untouched
  })

  it('asks again on refresh: a server still loading answered with none', async () => {
    let ready = false
    const ask = vi.fn(async () => (ready ? [typeHint(1, 2, ': Vec<i32>')] : []))
    const hints = setup(ask)
    await settle()
    expect(shown()).toEqual([])
    ready = true
    hints.refresh()
    await settle()
    expect(ask).toHaveBeenCalledTimes(2)
    expect(shown()).toEqual([': Vec<i32>'])
  })

  it('debounces a burst of edits into one request', async () => {
    const ask = vi.fn(async () => [] as InlayHint[])
    setup(ask)
    for (let i = 0; i < 5; i++) {
      view.dispatch({ changes: { from: 0, insert: 'a' } })
      await settle(100)
    }
    await settle(INLAY_DELAY_MS)
    expect(ask).toHaveBeenCalledTimes(1)
  })

  it('ignores an answer that arrives after a newer request', async () => {
    const answers: ((hints: InlayHint[]) => void)[] = []
    const ask = vi.fn(() => new Promise<InlayHint[]>((resolve) => answers.push(resolve)))
    setup(ask)
    await settle()
    view.dispatch({ changes: { from: 0, insert: 'a' } })
    await settle()
    expect(answers).toHaveLength(2)
    answers[1]?.([typeHint(1, 3, ': new')])
    await settle(0)
    answers[0]?.([typeHint(1, 2, ': old')])
    await settle(0)
    expect(shown()).toEqual([': new'])
  })

  it('puts the hint at its column counted in characters, not UTF-16 units', async () => {
    setup(async () => [typeHint(1, 5, ': char')], '"😀"x\n')
    await settle()
    expect(host.querySelector('.cm-line')?.textContent).toBe('"😀"x: char')
  })

  it('removes the widgets when turned off, and asks again when turned on', async () => {
    const ask = vi.fn(async () => [typeHint(1, 2)])
    const hints = setup(ask)
    await settle()
    expect(shown()).toHaveLength(1)
    hints.setEnabled(false)
    await settle(0)
    expect(shown()).toEqual([])
    await settle(INLAY_DELAY_MS * 2)
    expect(ask).toHaveBeenCalledTimes(1)
    hints.setEnabled(true)
    await settle()
    expect(ask).toHaveBeenCalledTimes(2)
    expect(shown()).toHaveLength(1)
  })

  it('draws nothing when the request fails', async () => {
    setup(async () => Promise.reject(new Error('no server')))
    await settle()
    expect(shown()).toEqual([])
  })
})

describe('sampleInlayHints (mock)', () => {
  it('gives hints for the Go sample only, inside the requested lines', () => {
    const range = (file: string, start: number, end: number): SourceRange => ({
      start: { file, line: start, column: 1 },
      end: { file, line: end, column: 1 }
    })
    expect(sampleInlayHints(range(SAMPLE_MAIN, 1, 20)).length).toBeGreaterThan(1)
    expect(sampleInlayHints(range(SAMPLE_MAIN, 1, 5))).toEqual([])
    expect(sampleInlayHints(range('other/main.py', 1, 20))).toEqual([])
  })
})
