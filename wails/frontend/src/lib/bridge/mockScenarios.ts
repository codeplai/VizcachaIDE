import type { ExplainedDiagnostic } from '../domain'
import type { EventName, EventPayloads } from '../events'
import type { Language } from '../language'
import { SAMPLE_MAIN, sampleDebugState, sampleDiagnostic, sampleExplanation } from './mockData'

export type Scenario = 'write' | 'error' | 'debug'

export type Emit = <E extends EventName>(name: E, payload: EventPayloads[E]) => void

export const clearProblems = (emit: Emit): void => {
  emit('lsp:diagnostics', { path: SAMPLE_MAIN, diagnostics: [] })
  emit('assistant:explained', [])
}

export const emitSuccessfulRun = (emit: Emit): void => {
  clearProblems(emit)
  emit('run:started', {
    target: SAMPLE_MAIN,
    workingDir: 'hola-go',
    mode: 'file',
    programArgs: [],
    module: null
  })
  emit('run:output', { stream: 'stdout', text: 'Hola, Go\n' })
  emit('run:finished', { exitCode: 0, durationMs: 400 })
}

export const emitFailedRun = (emit: Emit, language: Language): void => {
  const diagnostic = sampleDiagnostic()
  const explained: ExplainedDiagnostic[] = [
    { diagnostic, explanation: sampleExplanation(language) }
  ]
  emit('run:started', {
    target: SAMPLE_MAIN,
    workingDir: 'hola-go',
    mode: 'file',
    programArgs: [],
    module: null
  })
  emit('run:output', { stream: 'stderr', text: `${diagnostic.rawText}\n` })
  emit('lsp:diagnostics', { path: SAMPLE_MAIN, diagnostics: [diagnostic] })
  emit('assistant:explained', explained)
  emit('run:finished', { exitCode: 1, durationMs: 300 })
}

export const emitDebugStart = (emit: Emit): void => {
  clearProblems(emit)
  emit('debug:output', { text: 'Starting the debugger…', category: 'console' })
  emit('debug:stopped', sampleDebugState())
}
