import { get } from 'svelte/store'
import { afterEach, describe, expect, it } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import type { Diagnostic, ExplainedDiagnostic, Variable } from '../domain'
import {
  connectStores,
  debugActive,
  mergeChildren,
  mergeProblems,
  outputLines,
  problemCount,
  splitChunk,
  toggleBreakpointLine,
  uiMode
} from '.'

const variable = (name: string, reference = 0, children: Variable[] = []): Variable => ({
  name,
  typeName: 'int',
  value: '1',
  reference,
  changed: false,
  children
})

describe('pure helpers', () => {
  it('splits chunks into lines', () => {
    expect(splitChunk('a\nb\n')).toEqual(['a', 'b'])
    expect(splitChunk('a\r\nb')).toEqual(['a', 'b'])
  })

  it('toggles breakpoints keeping lines sorted', () => {
    const one = toggleBreakpointLine({}, 'main.go', 9)
    const two = toggleBreakpointLine(one, 'main.go', 3)
    expect(two['main.go']).toEqual([3, 9])
    expect(toggleBreakpointLine(two, 'main.go', 3)['main.go']).toEqual([9])
  })

  it('puts lazily loaded children under the right variable', () => {
    const tree = [variable('a'), variable('p', 7, [variable('inner', 8)])]
    const merged = mergeChildren(tree, 8, [variable('leaf')])
    expect(merged[1]?.children[0]?.children[0]?.name).toBe('leaf')
  })

  it('merges gopls diagnostics with explanations without duplicates', () => {
    const diagnostic: Diagnostic = {
      location: { file: 'main.go', line: 11, column: 5 },
      severity: 'error',
      message: 'declared and not used: x',
      rawText: 'raw',
      source: 'go',
      code: '',
      end: null
    }
    const explained: ExplainedDiagnostic[] = [{ diagnostic, explanation: null }]
    expect(mergeProblems({ 'main.go': [diagnostic] }, explained)).toHaveLength(1)
    expect(mergeProblems({}, explained)).toHaveLength(1)
  })
})

describe('stores connected to the mock bridge', () => {
  let stop: (() => void) | undefined
  afterEach(() => stop?.())

  const connect = async () => {
    const mock = createMockBridge()
    stop = await connectStores(mock.bridge)
    return mock
  }

  it('shows the write state after a successful run', async () => {
    const { controls } = await connect()
    await controls.play('write')
    expect(get(uiMode)).toBe('write')
    expect(get(outputLines).map((line) => line.tone)).toEqual(['system', 'plain', 'success'])
  })

  it('shows the error state and the compile failure line', async () => {
    const { controls } = await connect()
    await controls.play('error')
    expect(get(uiMode)).toBe('error')
    expect(get(problemCount)).toBe(1)
    const last = get(outputLines).at(-1)
    expect(last?.key).toBe('run.compileFailed')
    expect(last?.values).toEqual({ count: 1 })
  })

  it('shows the debug state and leaves it when the user stops', async () => {
    const { bridge, controls } = await connect()
    await controls.play('debug')
    expect(get(uiMode)).toBe('debug')
    expect(get(debugActive)).toBe(true)
    await bridge.debug.stop()
    expect(get(debugActive)).toBe(false)
  })
})
