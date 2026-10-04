import type {
  CodeLanguage,
  DebugState,
  Diagnostic,
  ErrorExplanation,
  RunConfiguration
} from '../domain'
import type { EventName, EventPayloads } from '../events'
import type { Language } from '../language'
import { SAMPLE_MAIN, sampleDebugState, sampleDiagnostic, sampleExplanation } from './mockData'
import {
  PYTHON_DIR,
  PYTHON_MAIN,
  pythonDebugState,
  pythonDiagnostic,
  pythonExplanation
} from './mockPython'

export type Scenario = 'write' | 'error' | 'debug'

/** The languages the demo has a sample project for. */
export type SampleLanguage = Extract<CodeLanguage, 'go' | 'python'>

export type Emit = <E extends EventName>(name: E, payload: EventPayloads[E]) => void

interface Sample {
  main: string
  configuration: RunConfiguration
  greeting: string
  diagnostic: () => Diagnostic
  explanation: (language: Language) => ErrorExplanation
  debugState: (line?: number) => DebugState
}

const configurationOf = (
  codeLanguage: SampleLanguage,
  target: string,
  workingDir: string
): RunConfiguration => ({
  codeLanguage,
  target,
  workingDir,
  mode: 'file',
  programArgs: [],
  project: null,
  // Python runs in a pseudoterminal that echoes what is typed; Go reads a pipe.
  echo: codeLanguage === 'python'
})

const SAMPLES: Record<SampleLanguage, Sample> = {
  go: {
    main: SAMPLE_MAIN,
    configuration: configurationOf('go', SAMPLE_MAIN, 'hola-go'),
    greeting: 'Hola, Go',
    diagnostic: sampleDiagnostic,
    explanation: sampleExplanation,
    debugState: (line) => sampleDebugState(line)
  },
  python: {
    main: PYTHON_MAIN,
    configuration: configurationOf('python', PYTHON_MAIN, PYTHON_DIR),
    greeting: 'Hola, Python',
    diagnostic: pythonDiagnostic,
    explanation: pythonExplanation,
    debugState: (line) => pythonDebugState(line)
  }
}

export const sampleOf = (codeLanguage: SampleLanguage): Sample => SAMPLES[codeLanguage]

export const clearProblems = (emit: Emit, codeLanguage: SampleLanguage = 'go'): void => {
  const { main } = sampleOf(codeLanguage)
  emit('lsp:diagnostics', { path: main, diagnostics: [] })
  emit('assistant:explained', [])
  emit('file:changed', { path: main }) // the sample is the same on disk: stores ignore it
}

export const emitSuccessfulRun = (emit: Emit, codeLanguage: SampleLanguage = 'go'): void => {
  const sample = sampleOf(codeLanguage)
  clearProblems(emit, codeLanguage)
  emit('run:started', sample.configuration)
  emit('run:output', { stream: 'stdout', text: `${sample.greeting}\n` })
  emit('run:finished', { exitCode: 0, durationMs: 400 })
}

export const emitFailedRun = (
  emit: Emit,
  language: Language,
  codeLanguage: SampleLanguage = 'go'
): void => {
  const sample = sampleOf(codeLanguage)
  const diagnostic = sample.diagnostic()
  const explained = [{ diagnostic, explanation: sample.explanation(language) }]
  emit('run:started', sample.configuration)
  emit('run:output', { stream: 'stderr', text: `${diagnostic.rawText}\n` })
  emit('lsp:diagnostics', { path: sample.main, diagnostics: [diagnostic] })
  emit('assistant:explained', explained)
  emit('run:finished', { exitCode: 1, durationMs: 300 })
}

export const emitDebugStart = (emit: Emit, codeLanguage: SampleLanguage = 'go'): void => {
  clearProblems(emit, codeLanguage)
  emit('debug:output', { text: 'Starting the debugger…', category: 'console' })
  emit('debug:stopped', sampleOf(codeLanguage).debugState())
}
