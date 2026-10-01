import type {
  Breakpoint,
  CompletionItem,
  Diagnostic,
  DocumentSymbol,
  ExplainedDiagnostic,
  FileNode,
  RunConfiguration,
  Settings,
  SignatureHelp,
  SourceLocation,
  SourceRange,
  ToolchainInfo
} from '../domain'
import type { EventName, EventPayloads } from '../events'

export type Unsubscribe = () => void

/** Mirrors bridge.RunService (Go). */
export interface RunApi {
  run: (path: string, programArgs: string[]) => Promise<RunConfiguration>
  runUntitled: (source: string, programArgs: string[]) => Promise<RunConfiguration>
  build: (path: string, programArgs: string[]) => Promise<RunConfiguration>
  /** Splits the "program arguments" text like a shell (quotes group words). */
  splitArguments: (text: string) => Promise<string[]>
  modInit: (workingDir: string, modulePath: string) => Promise<void>
  modGet: (workingDir: string, pkg: string) => Promise<void>
  modTidy: (workingDir: string) => Promise<void>
  stop: () => Promise<void>
  writeInput: (text: string) => Promise<void>
  format: (text: string) => Promise<string>
  toolchain: () => Promise<ToolchainInfo>
}

/** Mirrors bridge.DebugService (Go). */
export interface DebugApi {
  start: (path: string, breakpoints: Breakpoint[], argsText?: string) => Promise<void>
  setBreakpoints: (file: string, lines: number[]) => Promise<void>
  stepOver: () => Promise<void>
  stepInto: () => Promise<void>
  stepOut: () => Promise<void>
  resume: () => Promise<void>
  runTo: (location: SourceLocation) => Promise<void>
  requestVariables: (reference: number) => Promise<void>
  stop: () => Promise<void>
}

/** Mirrors bridge.LanguageService (Go). */
export interface LanguageApi {
  openDocument: (path: string, text: string) => Promise<void>
  changeDocument: (path: string, text: string, version: number) => Promise<void>
  closeDocument: (path: string) => Promise<void>
  completion: (at: SourceLocation) => Promise<CompletionItem[]>
  hover: (at: SourceLocation) => Promise<string>
  definition: (at: SourceLocation) => Promise<SourceLocation | null>
  signatureHelp: (at: SourceLocation) => Promise<SignatureHelp | null>
  documentHighlights: (at: SourceLocation) => Promise<SourceRange[]>
  documentSymbols: (path: string) => Promise<DocumentSymbol[]>
}

/** Mirrors bridge.AssistantService (Go). */
export interface AssistantApi {
  explain: (rawOutput: string, workingDir: string) => Promise<ExplainedDiagnostic[]>
  explainDiagnostics: (diagnostics: Diagnostic[]) => Promise<ExplainedDiagnostic[]>
}

/** Mirrors bridge.FilesService (Go). */
export interface FilesApi {
  openFolder: () => Promise<FileNode>
  listTree: (root: string) => Promise<FileNode>
  readFile: (path: string) => Promise<string>
  saveFile: (path: string, text: string) => Promise<void>
}

/** Mirrors bridge.SettingsService (Go). */
export type ToolId = 'go' | 'dlv' | 'gopls'

export interface SettingsApi {
  get: () => Promise<Settings>
  save: (settings: Settings) => Promise<void>
  /** Native file dialog for a tool; saves the path and returns the tools detected again. */
  pickExecutable: (tool: ToolId) => Promise<ToolchainInfo>
  /** The language the backend resolved ("auto" becomes the system language). */
  resolvedLanguage: () => Promise<'en' | 'es'>
}

/** Things only the desktop shell can do (the Wails runtime in the app, the browser in the mock). */
export interface SystemApi {
  /** Opens a web page in the default browser (Wails BrowserOpenURL). */
  openUrl: (url: string) => void
}

/**
 * Everything the UI can ask of the backend, plus the event subscription.
 * The Wails implementation and the mock implementation both satisfy it.
 */
export interface Bridge {
  /** True when there is no Go backend (browser development with demo data). */
  readonly isMock: boolean
  run: RunApi
  debug: DebugApi
  language: LanguageApi
  assistant: AssistantApi
  files: FilesApi
  settings: SettingsApi
  system: SystemApi
  on: <E extends EventName>(name: E, handler: (payload: EventPayloads[E]) => void) => Unsubscribe
}
