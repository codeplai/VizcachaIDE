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
  CPP_COMPILING_NOTICE,
  CPP_DIR,
  CPP_MAIN,
  cppCrashExplained,
  cppDebugState,
  cppDiagnostic,
  cppExplanation
} from './mockCpp'
import {
  PYTHON_DIR,
  PYTHON_MAIN,
  pythonDebugState,
  pythonDiagnostic,
  pythonExplanation
} from './mockPython'

/** `crash` only plays for C++: the program compiles, runs and dies (Segmentation fault). */
export type Scenario = 'write' | 'error' | 'debug' | 'crash'

/** The languages the demo has a sample project for. */
export type SampleLanguage = Extract<CodeLanguage, 'go' | 'python' | 'cpp'>

export type Emit = <E extends EventName>(name: E, payload: EventPayloads[E]) => void

interface Sample {
  main: string
  configuration: RunConfiguration
  greeting: string
  diagnostic: () => Diagnostic
  explanation: (language: Language) => ErrorExplanation
  debugState: (line?: number) => DebugState
  /** The compiler's notice line printed before the program (compiled languages). */
  notice?: string
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
  // Python and C++ run in a pseudoterminal that echoes what is typed; Go reads a pipe.
  echo: codeLanguage !== 'go'
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
  },
  cpp: {
    main: CPP_MAIN,
    configuration: configurationOf('cpp', CPP_MAIN, CPP_DIR),
    greeting: 'Hola, C++',
    diagnostic: cppDiagnostic,
    explanation: cppExplanation,
    debugState: (line) => cppDebugState(line),
    notice: CPP_COMPILING_NOTICE
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
  if (sample.notice) emit('run:output', { stream: 'stdout', text: `${sample.notice}\n` })
  emit('run:output', { stream: 'stdout', text: `${sample.greeting}\n` })
  emit('run:finished', { exitCode: 0, durationMs: 400 })
}

/** Like a run, but only the compiler's part: the program does not start. */
export const emitSuccessfulBuild = (emit: Emit, codeLanguage: SampleLanguage = 'go'): void => {
  const sample = sampleOf(codeLanguage)
  clearProblems(emit, codeLanguage)
  emit('run:started', sample.configuration)
  if (sample.notice) emit('run:output', { stream: 'stdout', text: `${sample.notice}\n` })
  emit('run:finished', { exitCode: 0, durationMs: 300 })
}

/** The C++ program that prints, then dies on a bad memory access (exit code of SIGSEGV). */
export const emitCrashedRun = (emit: Emit, language: Language): void => {
  const sample = sampleOf('cpp')
  clearProblems(emit, 'cpp')
  emit('run:started', sample.configuration)
  emit('run:output', { stream: 'stdout', text: `${sample.notice}\n${sample.greeting}\n` })
  emit('run:output', { stream: 'stderr', text: 'Segmentation fault\n' })
  emit('assistant:explained', cppCrashExplained(language))
  emit('run:finished', { exitCode: 139, durationMs: 350 })
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
  if (sample.notice) emit('run:output', { stream: 'stdout', text: `${sample.notice}\n` })
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
