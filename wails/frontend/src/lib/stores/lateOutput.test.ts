import { get } from 'svelte/store'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { Bridge } from '../bridge'
import type { Diagnostic, RunConfiguration } from '../domain'
import { connectAssistant, LATE_OUTPUT_DELAY_MS, runDiagnostics } from './assistant'
import { connectRun } from './run'

type Handler = (payload: never) => void

const segfault: Diagnostic = {
  location: null,
  severity: 'error',
  message: 'Segmentation fault',
  rawText: 'Segmentation fault',
  source: 'runtime',
  code: 'CPP-SEGFAULT'
} as unknown as Diagnostic

/** A bridge whose events are emitted by the test, in any order. */
const fakeBridge = () => {
  const handlers = new Map<string, Handler[]>()
  const explain = vi.fn(async (_language: string, output: string) =>
    output.includes('Segmentation fault') ? [{ diagnostic: segfault }] : []
  )
  const bridge = {
    on: (name: string, handler: Handler) => {
      handlers.set(name, [...(handlers.get(name) ?? []), handler])
      return () => undefined
    },
    assistant: { explain, explainDiagnostics: vi.fn(async () => []) },
    run: { check: vi.fn(async () => '') }
  } as unknown as Bridge
  const emit = (name: string, payload: unknown): void =>
    (handlers.get(name) ?? []).forEach((handler) => handler(payload as never))
  return { bridge, emit, explain }
}

const config = {
  codeLanguage: 'cpp',
  target: 'C:\\qa\\caida',
  workingDir: 'C:\\qa\\caida',
  echo: true
} as unknown as RunConfiguration

beforeEach(() => {
  vi.useFakeTimers()
  runDiagnostics.set([])
})

afterEach(() => vi.useRealTimers())

describe('late output', () => {
  it('explains the run again when the crash line arrives after run:finished', async () => {
    const { bridge, emit, explain } = fakeBridge()
    connectRun(bridge)
    connectAssistant(bridge)
    emit('run:started', config)
    emit('run:output', { stream: 'stdout', text: 'antes de caer\n' })
    emit('run:finished', { exitCode: 3221225477, durationMs: 10 })
    await vi.advanceTimersByTimeAsync(0)
    expect(get(runDiagnostics)).toEqual([]) // explained without the line: nothing yet
    emit('run:output', { stream: 'stderr', text: 'Segmentation fault\n' })
    await vi.advanceTimersByTimeAsync(LATE_OUTPUT_DELAY_MS)
    expect(explain).toHaveBeenCalledTimes(2)
    expect(get(runDiagnostics)).toEqual([segfault])
  })

  it('does not explain again for output of the next run', async () => {
    const { bridge, emit, explain } = fakeBridge()
    connectRun(bridge)
    connectAssistant(bridge)
    emit('run:started', config)
    emit('run:finished', { exitCode: 0, durationMs: 10 })
    emit('run:started', config)
    emit('run:output', { stream: 'stdout', text: 'otra vez\n' })
    await vi.advanceTimersByTimeAsync(LATE_OUTPUT_DELAY_MS)
    expect(explain).not.toHaveBeenCalled()
  })
})
