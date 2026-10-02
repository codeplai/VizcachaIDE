// The only module (with ./mock.ts) that touches the generated wailsjs code.
import * as AssistantService from '../../../wailsjs/go/bridge/AssistantService'
import * as ConsoleService from '../../../wailsjs/go/bridge/ConsoleService'
import * as DebugService from '../../../wailsjs/go/bridge/DebugService'
import * as FilesService from '../../../wailsjs/go/bridge/FilesService'
import * as LanguageService from '../../../wailsjs/go/bridge/LanguageService'
import * as RunService from '../../../wailsjs/go/bridge/RunService'
import * as SettingsService from '../../../wailsjs/go/bridge/SettingsService'
import { BrowserOpenURL, EventsOn } from '../../../wailsjs/runtime/runtime'
import type { Bridge, ConsoleApi, DebugApi, FilesApi, RunApi } from './types'

// The generated classes and our plain interfaces describe the same JSON.
const fromWire = <T>(value: unknown): T => value as T
const toWire = (value: unknown): never => value as never

export const hasWailsBackend = (): boolean =>
  typeof window !== 'undefined' && 'go' in window && window.go !== undefined

const createRunApi = (): RunApi => ({
  run: async (path, args) => fromWire(await RunService.Run(path, args)),
  runUntitled: async (source, args) => fromWire(await RunService.RunUntitled(source, args)),
  build: async (path, args) => fromWire(await RunService.Build(path, args)),
  splitArguments: (text) => RunService.SplitArguments(text),
  modInit: (dir, modulePath) => RunService.ModInit(dir, modulePath),
  modGet: (dir, pkg) => RunService.ModGet(dir, pkg),
  modTidy: (dir) => RunService.ModTidy(dir),
  vet: (config) => RunService.Vet(toWire(config)),
  stop: () => RunService.Stop(),
  writeInput: (text) => RunService.WriteInput(text),
  format: (text) => RunService.Format(text),
  toolchain: async () => fromWire(await RunService.Toolchain())
})

const createConsoleApi = (): ConsoleApi => ({
  eval: async (code) => fromWire(await ConsoleService.Eval(code)),
  reset: () => ConsoleService.Reset()
})

const createDebugApi = (): DebugApi => ({
  start: (path, breakpoints, argsText) =>
    DebugService.Start(path, toWire(breakpoints), argsText ?? ''),
  setBreakpoints: (file, lines) => DebugService.SetBreakpoints(file, lines),
  stepOver: () => DebugService.StepOver(),
  stepInto: () => DebugService.StepInto(),
  stepOut: () => DebugService.StepOut(),
  resume: () => DebugService.Resume(),
  runTo: (location) => DebugService.RunTo(toWire(location)),
  requestVariables: (reference) => DebugService.RequestVariables(reference),
  frameVariables: async (frameId) => fromWire(await DebugService.FrameVariables(frameId)),
  stop: () => DebugService.Stop()
})

const createFilesApi = (): FilesApi => ({
  openFolder: async () => fromWire(await FilesService.OpenFolder()),
  listTree: async (root) => fromWire(await FilesService.ListTree(root)),
  readFile: (path) => FilesService.ReadFile(path),
  saveFile: (path, text) => FilesService.SaveFile(path, text),
  watchFiles: (paths) => FilesService.WatchFiles(paths),
  openFileDialog: () => FilesService.OpenFileDialog(),
  saveFileDialog: (suggestedName, folder) => FilesService.SaveFileDialog(suggestedName, folder)
})

export const createWailsBridge = (): Bridge => ({
  isMock: false,
  run: createRunApi(),
  debug: createDebugApi(),
  language: {
    openDocument: (path, text) => LanguageService.OpenDocument(path, text),
    changeDocument: (path, text, version) => LanguageService.ChangeDocument(path, text, version),
    closeDocument: (path) => LanguageService.CloseDocument(path),
    completion: async (at) => fromWire(await LanguageService.Completion(toWire(at))),
    hover: (at) => LanguageService.Hover(toWire(at)),
    definition: async (at) => fromWire(await LanguageService.Definition(toWire(at))),
    signatureHelp: async (at) => fromWire(await LanguageService.SignatureHelp(toWire(at))),
    documentHighlights: async (at) =>
      fromWire(await LanguageService.DocumentHighlights(toWire(at))),
    documentSymbols: async (path) => fromWire(await LanguageService.DocumentSymbols(path))
  },
  assistant: {
    explain: async (raw, dir) => fromWire(await AssistantService.Explain(raw, dir)),
    explainDiagnostics: async (diagnostics) =>
      fromWire(await AssistantService.ExplainDiagnostics(toWire(diagnostics)))
  },
  console: createConsoleApi(),
  files: createFilesApi(),
  settings: {
    get: async () => fromWire(await SettingsService.Get()),
    save: (settings) => SettingsService.Save(toWire(settings)),
    pickExecutable: async (tool) => fromWire(await SettingsService.PickExecutable(tool)),
    resolvedLanguage: async () =>
      (await SettingsService.ResolvedLanguage()) === 'es' ? 'es' : 'en'
  },
  system: { openUrl: (url) => BrowserOpenURL(url) },
  on: (name, handler) => EventsOn(name, handler)
})
