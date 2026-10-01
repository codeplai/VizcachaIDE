// Mock bridge: lets `npm run dev` show the whole app in a browser without Go.
import type { RunConfiguration, Settings } from '../domain'
import { TERMINATED_BY_USER } from '../events'
import { resolveLanguage, systemLanguage, type Language } from '../language'
import { createEmitter } from './emitter'
import {
  SAMPLE_SOURCES,
  defaultSettings,
  sampleDebugState,
  sampleLocation,
  sampleSymbols,
  sampleTree
} from './mockData'
import {
  emitDebugStart,
  emitFailedRun,
  emitSuccessfulRun,
  type Emit,
  type Scenario
} from './mockScenarios'
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

const TOOLCHAIN = { goVersion: '1.25.5', delveVersion: '1.27.2', goplsVersion: '0.21.1' }
const LAST_LINE = 7

const configurationFor = (path: string): RunConfiguration => ({
  target: path,
  workingDir: 'hola-go',
  mode: 'file',
  programArgs: [],
  module: null
})

const mockRun = (state: MockState, emit: Emit): RunApi => {
  const language = (): Language => resolveLanguage(state.settings.language, systemLanguage())
  const run: RunApi['run'] = async (path) => {
    if (state.scenario === 'error') emitFailedRun(emit, language())
    else emitSuccessfulRun(emit)
    return configurationFor(path)
  }
  return {
    run,
    runUntitled: (_source, args) => run('untitled/main.go', args),
    stop: async () => {},
    writeInput: async () => {},
    format: async (text) => text,
    toolchain: async () => TOOLCHAIN
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
    requestVariables: async (reference) => emit('debug:variables', { reference, variables: [] }),
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
      documentation: '',
      insertText: ''
    }
  ],
  hover: async () => 'func sumar(a, b int) int',
  definition: async () => sampleLocation(5, 6),
  signatureHelp: async () => null,
  documentHighlights: async () => [],
  documentSymbols: async () => sampleSymbols()
})

const mockSettings = (state: MockState, emit: Emit): Bridge['settings'] => ({
  get: async () => state.settings,
  save: async (next) => {
    state.settings = next
    emit('settings:changed', next)
  }
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
    debug,
    language: mockLanguage(emit),
    assistant: { explain: async () => [], explainDiagnostics: async () => [] },
    files: {
      openFolder: async () => sampleTree(),
      listTree: async () => sampleTree(),
      readFile: async (path) => SAMPLE_SOURCES[path] ?? '',
      saveFile: async () => {}
    },
    settings: mockSettings(state, emit),
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
