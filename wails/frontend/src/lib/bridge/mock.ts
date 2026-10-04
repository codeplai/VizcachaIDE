// Mock bridge: lets `npm run dev` show the whole app in a browser without Go.
import type { RunConfiguration, Settings } from '../domain'
import { TERMINATED_BY_USER } from '../events'
import { resolveLanguage, systemLanguage, type Language } from '../language'
import { createEmitter } from './emitter'
import { mockAssistant } from './mockAssistant'
import { mockConsole } from './mockConsole'
import { mockFiles } from './mockFiles'
import {
  defaultSettings,
  sampleDebugState,
  SAMPLE_STRUCT_REFERENCE,
  sampleLocation,
  sampleStructFields,
  sampleSymbols
} from './mockData'
import {
  emitDebugStart,
  emitFailedRun,
  emitSuccessfulRun,
  type Emit,
  type Scenario
} from './mockScenarios'
import { sampleFrameVariables } from './mockFrames'
import { codeLanguageOfPath, mockCodeLanguages, mockPackages } from './mockCodeLanguages'
import { mockSettings, toolchainFor } from './mockSettings'
import type { Bridge, DebugApi, LanguageApi, RunApi } from './types'

export interface MockControls {
  /** Replays one of the three states of the prototype. */
  play: (scenario: Scenario) => Promise<void>
  scenario: () => Scenario
}

export interface MockBridge {
  bridge: Bridge
  controls: MockControls
}

interface MockState {
  settings: Settings
  scenario: Scenario
  debugLine: number
  debugging: boolean
}

/** The text of app.ErrUnsupported: the language's adapter does not exist yet. */
const UNSUPPORTED_ACTION = 'this language does not support the action'
const LAST_LINE = 7
const FIELDS_DELAY_MS = 250

const configurationFor = (path: string): RunConfiguration => ({
  codeLanguage: codeLanguageOfPath(path),
  target: path,
  workingDir: 'hola-go',
  mode: 'file',
  programArgs: [],
  project: null,
  echo: false
})

const mockRun = (state: MockState, emit: Emit): RunApi => {
  const language = (): Language => resolveLanguage(state.settings.language, systemLanguage())
  const run: RunApi['run'] = async (path) => {
    // Like the 2.1 backend: Python and C++ have a profile but no adapter yet.
    if (codeLanguageOfPath(path) !== 'go') throw new Error(UNSUPPORTED_ACTION)
    if (state.scenario === 'error') emitFailedRun(emit, language())
    else emitSuccessfulRun(emit)
    return configurationFor(path)
  }
  return {
    run,
    runUntitled: (path, _source, args) => run(path, args),
    build: async (path) => configurationFor(path),
    splitArguments: async (text) => text.split(/\s+/).filter(Boolean),
    modInit: async () => {},
    modGet: async () => {},
    modTidy: async () => {},
    check: async () => '',
    stop: async () => emit('run:finished', { exitCode: TERMINATED_BY_USER, durationMs: 0 }),
    writeInput: async () => {},
    format: async (_path, text) =>
      text.replace(/^( {4})+/gm, (indent) => '	'.repeat(indent.length / 4)),
    toolchain: async () => toolchainFor(state.settings)
  }
}

const mockDebug = (state: MockState, emit: Emit): DebugApi => {
  const finish = (exitCode: number): void => {
    state.debugging = false
    emit('debug:terminated', { exitCode })
  }
  const stepTo = async (line: number): Promise<void> => {
    state.debugLine = line
    emit('debug:stopped', sampleDebugState(line))
  }
  return {
    start: async () => {
      state.debugging = true
      state.debugLine = 6
      emitDebugStart(emit)
    },
    setBreakpoints: async () => {},
    stepOver: () => stepTo(Math.min(state.debugLine + 1, LAST_LINE)),
    stepInto: () => stepTo(Math.min(state.debugLine + 1, LAST_LINE)),
    stepOut: () => stepTo(LAST_LINE),
    resume: async () => finish(0),
    runTo: (location) => stepTo(location.line),
    requestVariables: async (reference) => {
      const variables = reference === SAMPLE_STRUCT_REFERENCE ? sampleStructFields() : []
      setTimeout(() => emit('debug:variables', { reference, variables }), FIELDS_DELAY_MS)
    },
    frameVariables: async (frameId) => sampleFrameVariables(frameId),
    stop: async () => finish(TERMINATED_BY_USER)
  }
}

const mockLanguage = (emit: Emit): LanguageApi => ({
  openDocument: async () => emit('lsp:status', 'ready'),
  changeDocument: async () => {},
  closeDocument: async () => {},
  completion: async () => [
    {
      label: 'Println',
      kind: 'function',
      detail: 'func(a ...any)',
      documentation: 'Println formats using the default formats and writes to standard output.',
      insertText: ''
    },
    {
      label: 'Printf',
      kind: 'function',
      detail: 'func(format string, a ...any)',
      documentation: '',
      insertText: ''
    },
    {
      label: 'Sprintf',
      kind: 'function',
      detail: 'func(format string, a ...any) string',
      documentation: '',
      insertText: ''
    }
  ],
  hover: async () => 'func sumar(a, b int) int',
  definition: async () => sampleLocation(5, 6),
  signatureHelp: async () => ({
    label: 'func Println(a ...any) (n int, err error)',
    documentation: '',
    parameters: ['a ...any'],
    activeParameter: 0
  }),
  documentHighlights: async () => [],
  documentSymbols: async () => sampleSymbols()
})

export const createMockBridge = (): MockBridge => {
  const { on, emit } = createEmitter()
  const state: MockState = {
    settings: defaultSettings(),
    scenario: 'write',
    debugLine: 6,
    debugging: false
  }
  const debug = mockDebug(state, emit)
  const bridge: Bridge = {
    isMock: true,
    run: mockRun(state, emit),
    packages: mockPackages(),
    codeLanguages: mockCodeLanguages(state),
    debug,
    language: mockLanguage(emit),
    assistant: mockAssistant(
      () => state.scenario,
      () => resolveLanguage(state.settings.language, systemLanguage()),
      emit
    ),
    console: mockConsole(),
    files: mockFiles(),
    settings: mockSettings(state, emit),
    system: {
      openUrl: (url) => void window.open(url, '_blank', 'noopener'),
      readClipboard: () => navigator.clipboard.readText(),
      writeClipboard: (text) => navigator.clipboard.writeText(text)
    },
    on
  }
  const play = async (next: Scenario): Promise<void> => {
    state.scenario = next
    if (state.debugging) await debug.stop()
    if (next === 'debug') await debug.start('', [])
    else await bridge.run.run('', [])
  }
  return { bridge, controls: { play, scenario: () => state.scenario } }
}
