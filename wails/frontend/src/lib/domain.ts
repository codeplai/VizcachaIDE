// TypeScript mirror of internal/domain (JSON tags in camelCase). FROZEN contract, revised once as
// contract v3 in M0 (docs/PLAN_NUCLEO_MULTILENGUAJE.md section 3).
//
// Naming: "language" alone is the UI language (en/es, see ../language.ts); the programming
// language is a CodeLanguage.

import type { CodeLanguage } from './domainCodeLanguage'

export * from './domainCodeLanguage'

export type Severity = 'error' | 'warning' | 'info' | 'hint'

export interface SourceLocation {
  file: string
  line: number
  column: number
}

export interface SourceRange {
  start: SourceLocation
  end: SourceLocation
}

export interface Diagnostic {
  location: SourceLocation | null
  severity: Severity
  message: string
  rawText: string
  source: string
  code: string
  end: SourceLocation | null
}

export type SymbolKind =
  | 'function'
  | 'method'
  | 'struct'
  | 'interface'
  | 'type'
  | 'variable'
  | 'constant'
  | 'field'
  | 'package'
  | 'other'

export interface DocumentSymbol {
  name: string
  kind: SymbolKind
  location: SourceLocation
  range: SourceRange | null
  detail: string
  children: DocumentSymbol[]
}

export type CompletionKind =
  | 'keyword'
  | 'function'
  | 'method'
  | 'variable'
  | 'constant'
  | 'field'
  | 'type'
  | 'package'
  | 'other'

export interface CompletionItem {
  label: string
  kind: CompletionKind
  detail: string
  documentation: string
  insertText: string
}

export interface SignatureHelp {
  label: string
  documentation: string
  parameters: string[]
  activeParameter: number
}

export interface Breakpoint {
  location: SourceLocation
  condition: string
}

export interface Variable {
  name: string
  typeName: string
  value: string
  reference: number
  changed: boolean
  children: Variable[]
}

export interface StackFrame {
  frameId: number
  function: string
  location: SourceLocation | null
}

/** Arguments and locals of one stack frame (Go: domain.FrameVariables). */
export interface FrameVariables {
  arguments: Variable[]
  locals: Variable[]
}

/** One thread of the debugged program (a goroutine in Go). */
export interface Thread {
  threadId: number
  name: string
  location: SourceLocation | null
}

/** 'exception' is an uncaught error: a panic in Go, an exception in Python, a crash in C++. */
export type StopReason = 'entry' | 'breakpoint' | 'step' | 'pause' | 'exception'

export interface DebugState {
  reason: StopReason
  frames: StackFrame[]
  variables: Variable[]
  threads: Thread[]
  currentThread: number | null
  description: string
}

export interface ErrorExplanation {
  explanationId: string
  title: string
  body: string
  fixHint: string
  placeholders: Record<string, string>
}

export interface ExplainedDiagnostic {
  diagnostic: Diagnostic
  explanation: ErrorExplanation | null
}

export type ProjectKind = 'gomod' | 'folder' | 'pyproject'

/** The project around the file being run. For Go, name is the module path. */
export interface ProjectContext {
  root: string
  kind: ProjectKind
  name: string
}

export type RunTarget = 'file' | 'project'

export interface RunConfiguration {
  codeLanguage: CodeLanguage
  target: string
  workingDir: string
  mode: RunTarget
  programArgs: string[]
  project: ProjectContext | null
  /** True when a pseudoterminal already echoes the input: Output must not repeat it. */
  echo: boolean
}

export interface FileNode {
  name: string
  path: string
  isDir: boolean
  children: FileNode[]
}

export type ServerStatus = 'starting' | 'ready' | 'unavailable'
export type LanguageSetting = 'auto' | 'en' | 'es'
export type ThemeSetting = 'system' | 'light' | 'dark'

export interface Settings {
  language: LanguageSetting
  theme: ThemeSetting
  fontSize: number
  /** Executables chosen in Settings, by ToolSpec.id; missing or '' = find it automatically. */
  toolPaths: Record<string, string>
  firstRun: boolean
  lastFolder: string
  /** Language of new files. */
  defaultCodeLanguage: CodeLanguage
  /** Languages chosen in the first-run wizard; empty = all. */
  enabledCodeLanguages: CodeLanguage[]
  formatOnSave: boolean
  /** Last opened files, newest first (at most 10). */
  recentFiles: string[]
}

/** What the interactive console answers to one snippet. */
export interface ConsoleResult {
  /** Value of an expression, formatted like a REPL (strings quoted). Empty for statements. */
  result: string
  /** What the snippet printed. */
  output: string
  /** Plain error message, empty when the snippet ran fine. */
  error: string
}
