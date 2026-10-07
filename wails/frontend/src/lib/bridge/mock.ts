// Mock bridge: lets `npm run dev` show the whole app in a browser without Go.
import { resolveLanguage, systemLanguage } from '../language'
import { createEmitter } from './emitter'
import { mockAssistant } from './mockAssistant'
import { mockConsole } from './mockConsole'
import { defaultSettings, sampleLocation, sampleSymbols } from './mockData'
import { mockDebug, mockRun, type MockState } from './mockExecution'
import { mockFiles } from './mockFiles'
import { mockRefactor } from './mockRefactor'
import { mockProjects } from './mockProjects'
import { sampleInlayHints } from './mockInlay'
import { codeLanguageOfPath, mockCodeLanguages, mockPackages } from './mockCodeLanguages'
import type { Emit, SampleLanguage, Scenario } from './mockScenarios'
import { mockSearch } from './mockSearch'
import { mockSettings } from './mockSettings'
import { mockTerminal } from './mockTerminal'
import { mockUpdates } from './mockUpdates'
import type { Bridge, FilesApi, LanguageApi } from './types'

export interface MockControls {
  /** Replays one of the three states of the prototype. */
  play: (scenario: Scenario) => Promise<void>
  scenario: () => Scenario
}

export interface MockBridge {
  bridge: Bridge
  controls: MockControls
}

const mockLanguage = (emit: Emit, files: FilesApi): LanguageApi => {
  const documents = new Map<string, string>() // what the editor has open, like a real server
  return {
    ...mockRefactor(files, documents),
    ...mockBasics(emit, documents)
  }
}

const mockBasics = (
  emit: Emit,
  documents: Map<string, string>
): Omit<LanguageApi, 'prepareRename' | 'rename' | 'references'> => ({
  openDocument: async (path, text) => {
    documents.set(path, text)
    emit('lsp:status', { codeLanguage: codeLanguageOfPath(path), status: 'ready' })
  },
  changeDocument: async (path, text) => void documents.set(path, text),
  closeDocument: async (path) => void documents.delete(path),
  untitledFile: async (name) => `untitled/${name}`,
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
  documentSymbols: async () => sampleSymbols(),
  inlayHints: async (visible) => sampleInlayHints(visible)
})

export interface MockOptions {
  /** The sample project the demo opens (`?language=python` in the browser). */
  sampleLanguage?: SampleLanguage
}

export const createMockBridge = ({ sampleLanguage = 'go' }: MockOptions = {}): MockBridge => {
  const { on, emit } = createEmitter()
  const state: MockState = {
    settings: defaultSettings(sampleLanguage),
    scenario: 'write',
    sampleLanguage,
    debugLine: 6,
    debugging: false
  }
  const debug = mockDebug(state, emit)
  const files = mockFiles(sampleLanguage)
  const bridge: Bridge = {
    isMock: true,
    run: mockRun(state, emit),
    packages: mockPackages(),
    codeLanguages: mockCodeLanguages(state),
    debug,
    language: mockLanguage(emit, files),
    assistant: mockAssistant(
      () => state.scenario,
      () => resolveLanguage(state.settings.language, systemLanguage()),
      emit
    ),
    console: mockConsole(),
    projects: mockProjects(files),
    files,
    search: mockSearch(files),
    settings: mockSettings(state, emit),
    updates: mockUpdates(emit),
    terminal: mockTerminal(emit),
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
