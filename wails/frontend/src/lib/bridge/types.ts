import type {
  Breakpoint,
  CodeLanguage,
  ConsoleResult,
  CompletionItem,
  Diagnostic,
  DocumentSymbol,
  EditSummary,
  InlayHint,
  ExplainedDiagnostic,
  FileEdit,
  FileNode,
  FrameVariables,
  PackageInfo,
  LanguageProfile,
  NewProject,
  RunConfiguration,
  Settings,
  SignatureHelp,
  SourceLocation,
  SourceRange,
  ToolStatus
} from '../domain'
import type { EventName, EventPayloads } from '../events'
import type { RefactorApi } from './typesRefactor'
import type { SearchApi } from './typesSearch'
import type { TerminalApi, UpdatesApi } from './typesShell'

export type { RefactorApi, SearchApi, TerminalApi, UpdatesApi }

export type Unsubscribe = () => void

/** Mirrors bridge.RunService (Go). The language comes from the path's extension. */
export interface RunApi {
  run: (path: string, programArgs: string[]) => Promise<RunConfiguration>
  /** path is the untitled name (e.g. "untitled-1.py"): its extension decides the language. */
  runUntitled: (path: string, source: string, programArgs: string[]) => Promise<RunConfiguration>
  build: (path: string, programArgs: string[]) => Promise<RunConfiguration>
  /** Runs the member of a Cargo workspace chosen after `run` rejected with run.chooseMember. */
  runMember: (path: string, member: string, programArgs: string[]) => Promise<RunConfiguration>
  /** Splits the "program arguments" text like a shell (quotes group words). */
  splitArguments: (text: string) => Promise<string[]>
  /** Runs the checker of config.codeLanguage (go vet, ruff check); "" when it found nothing. */
  check: (config: RunConfiguration) => Promise<string>
  stop: () => Promise<void>
  writeInput: (text: string) => Promise<void>
  format: (path: string, text: string) => Promise<string>
}

/** Mirrors bridge.PackagesService (Go). Verbs a language lacks reject with "unsupported". */
export interface PackagesApi {
  init: (codeLanguage: CodeLanguage, dir: string, name: string) => Promise<void>
  add: (codeLanguage: CodeLanguage, dir: string, pkg: string) => Promise<void>
  remove: (codeLanguage: CodeLanguage, dir: string, pkg: string) => Promise<void>
  tidy: (codeLanguage: CodeLanguage, dir: string) => Promise<void>
  list: (codeLanguage: CodeLanguage, dir: string) => Promise<void>
  /** Packages of the language's index whose name matches, at most 10. Rejects when the index
   *  cannot be searched (no network...); the student can still type the exact name. */
  search: (codeLanguage: CodeLanguage, query: string) => Promise<PackageInfo[]>
}

/** Mirrors bridge.ProjectsService (Go). Errors start with an i18n key (`project.errorExists`). */
export interface ProjectsApi {
  /** Creates <location>/<name> with the language's template; returns the folder and main file. */
  create: (codeLanguage: CodeLanguage, location: string, name: string) => Promise<NewProject>
}

/** Mirrors bridge.CodeLanguagesService (Go). Not to be confused with LanguageApi (the LSP). */
export interface CodeLanguagesApi {
  profiles: () => Promise<LanguageProfile[]>
  tools: () => Promise<ToolStatus[]>
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
  /** Arguments and locals of any frame of the paused stack (the Calls view). */
  frameVariables: (frameId: number) => Promise<FrameVariables>
  stop: () => Promise<void>
}

/** Mirrors bridge.LanguageService (Go): code intelligence, routed by path. */
export interface LanguageApi extends RefactorApi {
  openDocument: (path: string, text: string) => Promise<void>
  changeDocument: (path: string, text: string, version: number) => Promise<void>
  closeDocument: (path: string) => Promise<void>
  completion: (at: SourceLocation) => Promise<CompletionItem[]>
  hover: (at: SourceLocation) => Promise<string>
  definition: (at: SourceLocation) => Promise<SourceLocation | null>
  signatureHelp: (at: SourceLocation) => Promise<SignatureHelp | null>
  documentHighlights: (at: SourceLocation) => Promise<SourceRange[]>
  documentSymbols: (path: string) => Promise<DocumentSymbol[]>
  /** The hints of the visible lines of an open file (`visible.start.file`). */
  inlayHints: (visible: SourceRange) => Promise<InlayHint[]>
  /** Where a new unsaved file named `name` lives while open (an absolute temporary path). */
  untitledFile: (name: string) => Promise<string>
}

/** Mirrors bridge.AssistantService (Go). */
export interface AssistantApi {
  explain: (
    codeLanguage: CodeLanguage,
    rawOutput: string,
    workingDir: string
  ) => Promise<ExplainedDiagnostic[]>
  explainDiagnostics: (
    codeLanguage: CodeLanguage | '',
    diagnostics: Diagnostic[]
  ) => Promise<ExplainedDiagnostic[]>
}

/** Mirrors bridge.ConsoleService (Go). */
export interface ConsoleApi {
  eval: (codeLanguage: CodeLanguage, code: string) => Promise<ConsoleResult>
  reset: (codeLanguage: CodeLanguage) => Promise<void>
}

/** Mirrors bridge.FilesService (Go). */
export interface FilesApi {
  openFolder: () => Promise<FileNode>
  /** Native folder dialog that opens nothing; "" when the user cancels. */
  chooseFolder: () => Promise<string>
  listTree: (root: string) => Promise<FileNode>
  readFile: (path: string) => Promise<string>
  saveFile: (path: string, text: string) => Promise<void>
  /** Files (the open tabs) to watch; changes made elsewhere arrive as `file:changed`. */
  watchFiles: (paths: string[]) => Promise<void>
  /** Native "open file" dialog; "" when the user cancels. */
  openFileDialog: () => Promise<string>
  /** Native "save as" dialog (filters by the profiles' extensions); "" when cancelled. */
  saveFileDialog: (suggestedName: string, folder: string) => Promise<string>
  // Files panel operations (create, rename, Recycle Bin, Show in Explorer).
  createFile: (path: string, text: string) => Promise<void>
  createFolder: (path: string) => Promise<void>
  /** Fails if the target already exists. */
  rename: (from: string, to: string) => Promise<void>
  /** Copies a file or folder (recursively); fails if the target exists or is inside the source. */
  copy: (from: string, to: string) => Promise<void>
  /** Sends the file or folder to the Recycle Bin; it never deletes permanently. */
  moveToTrash: (path: string) => Promise<void>
  revealInExplorer: (path: string) => Promise<void>
  /** Edits files on disk (the ones a rename touches that are not open) and tells the watcher. */
  applyTextEdits: (files: FileEdit[]) => Promise<EditSummary>
}

/** A ToolSpec.id ('go', 'dlv', 'gopls', 'python', 'clangd'...). */
export type ToolId = string

/** Mirrors bridge.SettingsService (Go). */
export interface SettingsApi {
  get: () => Promise<Settings>
  save: (settings: Settings) => Promise<void>
  /**
   * Native file dialog for a tool (a ToolSpec.id); saves the path and returns the tools of
   * every language detected again.
   */
  pickExecutable: (toolId: ToolId) => Promise<ToolStatus[]>
  /** The language the backend resolved ("auto" becomes the system language). */
  resolvedLanguage: () => Promise<'en' | 'es'>
}

/** Things only the desktop shell can do (the Wails runtime in the app, the browser in the mock). */
export interface SystemApi {
  /** Opens a web page in the default browser (Wails BrowserOpenURL). */
  openUrl: (url: string) => void
  /** The clipboard, through Wails (a WebView2 page cannot read it by itself). */
  readClipboard: () => Promise<string>
  writeClipboard: (text: string) => Promise<void>
}

/**
 * Everything the UI can ask of the backend, plus the event subscription.
 * The Wails implementation and the mock implementation both satisfy it.
 */
export interface Bridge {
  /** True when there is no Go backend (browser development with demo data). */
  readonly isMock: boolean
  run: RunApi
  packages: PackagesApi
  projects: ProjectsApi
  codeLanguages: CodeLanguagesApi
  debug: DebugApi
  language: LanguageApi
  assistant: AssistantApi
  console: ConsoleApi
  files: FilesApi
  search: SearchApi
  settings: SettingsApi
  updates: UpdatesApi
  terminal: TerminalApi
  system: SystemApi
  on: <E extends EventName>(name: E, handler: (payload: EventPayloads[E]) => void) => Unsubscribe
}
