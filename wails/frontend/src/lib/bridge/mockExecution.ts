// Run and debug of the mock bridge: the sample of the file's language plays the three scenarios.
import type { CodeLanguage, RunConfiguration, Settings } from '../domain'
import { TERMINATED_BY_USER } from '../events'
import { resolveLanguage, systemLanguage, type Language } from '../language'
import { codeLanguageOfPath } from './mockCodeLanguages'
import { SAMPLE_STRUCT_REFERENCE, sampleStructFields } from './mockData'
import { sampleFrameVariables } from './mockFrames'
import { PYTHON_BREAKPOINT_LINE, PYTHON_LAST_LINE, pythonFrameVariables } from './mockPython'
import {
  emitDebugStart,
  emitFailedRun,
  emitSuccessfulRun,
  sampleOf,
  type Emit,
  type SampleLanguage,
  type Scenario
} from './mockScenarios'
import type { DebugApi, RunApi } from './types'

export interface MockState {
  settings: Settings
  scenario: Scenario
  /** The language of the sample project the demo opened (`?language=python`). */
  sampleLanguage: SampleLanguage
  debugLine: number
  debugging: boolean
}

/** The text of app.ErrUnsupported: the language's adapter does not exist yet. */
const UNSUPPORTED_ACTION = 'this language does not support the action'
const FIELDS_DELAY_MS = 250

/** Where the demo pauses first and where it ends, per language. */
const DEBUG_LINES: Record<SampleLanguage, { first: number; last: number }> = {
  go: { first: 6, last: 7 },
  python: { first: PYTHON_BREAKPOINT_LINE, last: PYTHON_LAST_LINE }
}

/** The language of a path; the demo's own sample when the path is empty (the dev bar). */
const languageOf = (state: MockState, path: string): CodeLanguage =>
  path ? codeLanguageOfPath(path) : state.sampleLanguage

/** Like the backend: C++ has a profile but no adapter yet. */
const sampleFor = (state: MockState, path: string): SampleLanguage => {
  const codeLanguage = languageOf(state, path)
  if (codeLanguage === 'cpp') throw new Error(UNSUPPORTED_ACTION)
  return codeLanguage
}

const configurationFor = (state: MockState, path: string): RunConfiguration => {
  const { configuration } = sampleOf(sampleFor(state, path))
  return { ...configuration, target: path || configuration.target }
}

export const mockRun = (state: MockState, emit: Emit): RunApi => {
  const language = (): Language => resolveLanguage(state.settings.language, systemLanguage())
  const run: RunApi['run'] = async (path) => {
    const sample = sampleFor(state, path)
    if (state.scenario === 'error') emitFailedRun(emit, language(), sample)
    else emitSuccessfulRun(emit, sample)
    return configurationFor(state, path)
  }
  return {
    run,
    runUntitled: (path, _source, args) => run(path, args),
    build: async (path) => configurationFor(state, path),
    splitArguments: async (text) => text.split(/\s+/).filter(Boolean),
    check: async () => '',
    stop: async () => emit('run:finished', { exitCode: TERMINATED_BY_USER, durationMs: 0 }),
    writeInput: async () => {},
    format: async (path, text) =>
      languageOf(state, path) === 'go'
        ? text.replace(/^( {4})+/gm, (indent) => '	'.repeat(indent.length / 4))
        : text
  }
}

export const mockDebug = (state: MockState, emit: Emit): DebugApi => {
  let sample: SampleLanguage = 'go'
  const finish = (exitCode: number): void => {
    state.debugging = false
    emit('debug:terminated', { exitCode })
  }
  const stepTo = async (line: number): Promise<void> => {
    state.debugLine = line
    emit('debug:stopped', sampleOf(sample).debugState(line))
  }
  const lastLine = (): number => DEBUG_LINES[sample].last
  return {
    start: async (path) => {
      sample = sampleFor(state, path)
      state.debugging = true
      state.debugLine = DEBUG_LINES[sample].first
      emitDebugStart(emit, sample)
    },
    setBreakpoints: async () => {},
    stepOver: () => stepTo(Math.min(state.debugLine + 1, lastLine())),
    stepInto: () => stepTo(Math.min(state.debugLine + 1, lastLine())),
    stepOut: () => stepTo(lastLine()),
    resume: async () => finish(0),
    runTo: (location) => stepTo(location.line),
    requestVariables: async (reference) => {
      const goStruct = sample === 'go' && reference === SAMPLE_STRUCT_REFERENCE
      const variables = goStruct ? sampleStructFields() : []
      setTimeout(() => emit('debug:variables', { reference, variables }), FIELDS_DELAY_MS)
    },
    frameVariables: async (frameId) =>
      sample === 'python' ? pythonFrameVariables(frameId) : sampleFrameVariables(frameId),
    stop: async () => finish(TERMINATED_BY_USER)
  }
}
