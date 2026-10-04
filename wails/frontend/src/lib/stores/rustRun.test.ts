import { get } from 'svelte/store'
import { afterEach, describe, expect, it } from 'vitest'
import { rustPanicDiagnostic } from '../bridge/mockRustErrors'
import { compileProblemCount, runDiagnostics } from './assistant'
import { outputLines } from './output'
import { lastRunConfiguration, resetRun, runResult } from './run'

afterEach(() => {
  runDiagnostics.set([])
  lastRunConfiguration.set(null)
  resetRun()
})

describe('the end of a Rust run', () => {
  it('does not count a panic as a compile failure: the verdict is the exit code', () => {
    runDiagnostics.set([rustPanicDiagnostic()])
    expect(get(compileProblemCount)).toBe(0)
    runResult.set({ exitCode: 101, durationMs: 20 })
    expect(get(outputLines).at(-1)).toMatchObject({ key: 'run.exitCode', values: { code: 101 } })
  })

  it('still counts a compiler error', () => {
    runDiagnostics.set([{ ...rustPanicDiagnostic(), source: 'compiler', code: 'E0382' }])
    runResult.set({ exitCode: 1, durationMs: 20 })
    expect(get(outputLines).at(-1)?.key).toBe('run.compileFailed')
  })
})
