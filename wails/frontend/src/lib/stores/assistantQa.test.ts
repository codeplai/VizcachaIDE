// Assistant and Output gaps found by the parity QA with the real backend (docs/wails/QA_WAILS.md).
import { cleanup, render, screen, waitFor } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { bridge as appBridge, mockControls, type Bridge } from '../bridge'
import { SAMPLE_MAIN } from '../bridge/mockData'
import type { Diagnostic } from '../domain'
import { setupI18n } from '../i18n'
import AssistantPanel from '../panels/AssistantPanel.svelte'
import {
  activePath,
  addChunk,
  assistantProblems,
  buffers,
  compileProblemCount,
  connectStores,
  diagnosticsByFile,
  explained,
  openFile,
  openTabs,
  problemCount,
  outputLines,
  problemKey,
  programArguments,
  runDiagnostics,
  runLines,
  running
} from '.'

let stop: (() => void) | undefined

const setUp = async (): Promise<Bridge> => {
  const bridge = appBridge
  stop = await connectStores(bridge)
  await openFile(bridge, SAMPLE_MAIN)
  return bridge
}

const diagnostic = (file: string, message: string, source = 'gopls'): Diagnostic => ({
  location: { file, line: 3, column: 1 },
  end: null,
  severity: 'error',
  message,
  rawText: message,
  source,
  code: ''
})

beforeAll(() => setupI18n('en'))
beforeEach(() => {
  openTabs.set([])
  activePath.set(null)
  buffers.set({})
  programArguments.set('')
  runLines.set([])
  running.set(false)
  diagnosticsByFile.set({})
  explained.set([])
  runDiagnostics.set([])
})
afterEach(() => {
  cleanup()
  stop?.()
})

describe('the Assistant receives the run output and the gopls problems', () => {
  it('counts only the problems Go printed in this run, not the ones gopls shows elsewhere', async () => {
    await setUp()
    await mockControls?.play('error')
    diagnosticsByFile.update((all) => ({
      ...all,
      'other.go': [diagnostic('other.go', 'main redeclared')]
    }))
    await waitFor(() => expect(get(compileProblemCount)).toBe(1))
    expect(get(problemCount)).toBe(2)
    expect(get(outputLines).at(-1)?.values).toEqual({ count: 1 })
  })

  it('explains the diagnostics of every file with one request when gopls publishes', async () => {
    const bridge = appBridge
    const controls = mockControls!
    stop = await connectStores(bridge)
    const explain = vi.spyOn(bridge.assistant, 'explainDiagnostics')
    await controls.play('error')
    await waitFor(() => expect(explain).toHaveBeenCalled())
    const sent = explain.mock.calls.at(-1)?.[1] ?? []
    expect(sent.some((item) => item.message.includes('resultado'))).toBe(true)
  })
})

describe('problems of other files', () => {
  it('stay out of the Assistant but are listed, told apart by place, in Problems', async () => {
    expect(problemKey(diagnostic('a.go', 'main redeclared'))).not.toBe(
      problemKey(diagnostic('b.go', 'main redeclared'))
    )
    await setUp()
    diagnosticsByFile.set({
      [SAMPLE_MAIN]: [diagnostic(SAMPLE_MAIN, 'declared and not used: x')],
      'other.go': [diagnostic('other.go', 'main redeclared')]
    })
    expect(get(problemCount)).toBe(2)
    expect(get(assistantProblems).map((item) => item.diagnostic.message)).toEqual([
      'declared and not used: x'
    ])
    render(AssistantPanel)
    expect(await screen.findAllByText('declared and not used: x')).not.toHaveLength(0)
    expect(screen.queryByText('main redeclared')).toBeNull()
  })
})

describe('output chunks that end in the middle of a line', () => {
  type Chunk = ['stdout' | 'stderr', string]
  const feed = (...chunks: Chunk[]) => {
    let state: ReturnType<typeof addChunk> = { lines: [], tail: null }
    for (const [stream, text] of chunks) state = addChunk(state.lines, state.tail, stream, text)
    return state
  }

  it('joins the two writes Go uses for a panic message', () => {
    const { lines } = feed(
      ['stderr', 'panic: '],
      ['stderr', 'assignment to entry in nil map\n\n'],
      ['stderr', 'goroutine 1\n']
    )
    expect(lines.map((line) => line.text)).toEqual([
      'panic: assignment to entry in nil map',
      '',
      'goroutine 1'
    ])
  })

  it('keeps a prompt open so the typed answer lands on the same line', () => {
    const { lines, tail } = feed(['stdout', 'Name: '], ['stdout', 'Ana\n'], ['stdout', 'Age: '])
    expect(lines.map((line) => line.text)).toEqual(['Name: Ana', 'Age: '])
    expect(tail).toBe('stdout')
  })

  it('does not glue stdout to stderr and ends on a full line', () => {
    const { lines, tail } = feed(['stdout', 'a'], ['stderr', 'b\n'])
    expect(lines).toEqual([
      { kind: 'stdout', text: 'a' },
      { kind: 'stderr', text: 'b' }
    ])
    expect(tail).toBeNull()
  })
})
