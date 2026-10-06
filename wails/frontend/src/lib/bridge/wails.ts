// The only module (with ./mock.ts) that touches the generated wailsjs code.
import * as AssistantService from '../../../wailsjs/go/bridge/AssistantService'
import * as ConsoleService from '../../../wailsjs/go/bridge/ConsoleService'
import * as DebugService from '../../../wailsjs/go/bridge/DebugService'
import * as FilesService from '../../../wailsjs/go/bridge/FilesService'
import * as LanguageService from '../../../wailsjs/go/bridge/LanguageService'
import * as RunService from '../../../wailsjs/go/bridge/RunService'
import * as SettingsService from '../../../wailsjs/go/bridge/SettingsService'
import * as TerminalService from '../../../wailsjs/go/bridge/TerminalService'
import * as UpdatesService from '../../../wailsjs/go/bridge/UpdatesService'
import {
  BrowserOpenURL,
  ClipboardGetText,
  ClipboardSetText,
  EventsOn
} from '../../../wailsjs/runtime/runtime'
import type {
  Bridge,
  ConsoleApi,
  DebugApi,
  FilesApi,
  RunApi,
  TerminalApi,
  UpdatesApi
} from './types'
import { createCodeLanguagesApi, createPackagesApi, createProjectsApi } from './wailsCodeLanguages'

// The generated classes and our plain interfaces describe the same JSON.
const fromWire = <T>(value: unknown): T => value as T
const toWire = (value: unknown): never => value as never

export const hasWailsBackend = (): boolean =>
  typeof window !== 'undefined' && 'go' in window && window.go !== undefined

const createRunApi = (): RunApi => ({
  run: async (path, args) => fromWire(await RunService.Run(path, args)),
  runUntitled: async (path, source, args) =>
    fromWire(await RunService.RunUntitled(path, source, args)),
  build: async (path, args) => fromWire(await RunService.Build(path, args)),
  runMember: async (path, member, args) => fromWire(await RunService.RunMember(path, member, args)),
  splitArguments: (text) => RunService.SplitArguments(text),
  check: (config) => RunService.Check(toWire(config)),
  stop: () => RunService.Stop(),
  writeInput: (text) => RunService.WriteInput(text),
  format: (path, text) => RunService.Format(path, text)
})

const createConsoleApi = (): ConsoleApi => ({
  eval: async (codeLanguage, code) =>
    fromWire(await ConsoleService.Eval(toWire(codeLanguage), code)),
  reset: (codeLanguage) => ConsoleService.Reset(toWire(codeLanguage))
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
  chooseFolder: () => FilesService.ChooseFolder(),
  listTree: async (root) => fromWire(await FilesService.ListTree(root)),
  readFile: (path) => FilesService.ReadFile(path),
  saveFile: (path, text) => FilesService.SaveFile(path, text),
  watchFiles: (paths) => FilesService.WatchFiles(paths),
  openFileDialog: () => FilesService.OpenFileDialog(),
  saveFileDialog: (suggestedName, folder) => FilesService.SaveFileDialog(suggestedName, folder),
  createFile: (path, text) => FilesService.CreateFile(path, text),
  createFolder: (path) => FilesService.CreateFolder(path),
  rename: (from, to) => FilesService.Rename(from, to),
  copy: (from, to) => FilesService.Copy(from, to),
  moveToTrash: (path) => FilesService.MoveToTrash(path),
  revealInExplorer: (path) => FilesService.RevealInExplorer(path)
})

const createUpdatesApi = (): UpdatesApi => ({
  state: async () => fromWire(await UpdatesService.State()),
  check: async () => fromWire(await UpdatesService.Check()),
  download: () => UpdatesService.Download(),
  install: () => UpdatesService.Install()
})

const createTerminalApi = (): TerminalApi => ({
  start: (dir, cols, rows) => TerminalService.Start(dir, cols, rows),
  write: (id, data) => TerminalService.Write(id, data),
  resize: (id, cols, rows) => TerminalService.Resize(id, cols, rows),
  close: (id) => TerminalService.Close(id)
})

export const createWailsBridge = (): Bridge => ({
  isMock: false,
  run: createRunApi(),
  packages: createPackagesApi(),
  projects: createProjectsApi(),
  codeLanguages: createCodeLanguagesApi(),
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
    documentSymbols: async (path) => fromWire(await LanguageService.DocumentSymbols(path)),
    inlayHints: async (visible) => fromWire(await LanguageService.InlayHints(toWire(visible)))
  },
  assistant: {
    explain: async (codeLanguage, raw, dir) =>
      fromWire(await AssistantService.Explain(toWire(codeLanguage), raw, dir)),
    explainDiagnostics: async (codeLanguage, diagnostics) =>
      fromWire(await AssistantService.ExplainDiagnostics(codeLanguage, toWire(diagnostics)))
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
  updates: createUpdatesApi(),
  terminal: createTerminalApi(),
  system: {
    openUrl: (url) => BrowserOpenURL(url),
    readClipboard: () => ClipboardGetText(),
    writeClipboard: async (text) => {
      if (!(await ClipboardSetText(text))) throw new Error('clipboard refused the text')
    }
  },
  on: (name, handler) => EventsOn(name, handler)
})
