// TypeScript mirror of internal/domain (JSON tags in camelCase). FROZEN contract.

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

export interface Goroutine {
  goroutineId: number
  name: string
  location: SourceLocation | null
}

export type StopReason = 'entry' | 'breakpoint' | 'step' | 'pause' | 'panic'

export interface DebugState {
  reason: StopReason
  frames: StackFrame[]
  variables: Variable[]
  goroutines: Goroutine[]
  currentGoroutine: number | null
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

export interface GoModule {
  root: string
  modulePath: string
}

export type RunTarget = 'file' | 'package'

export interface RunConfiguration {
  target: string
  workingDir: string
  mode: RunTarget
  programArgs: string[]
  module: GoModule | null
}

export interface FileNode {
  name: string
  path: string
  isDir: boolean
  children: FileNode[]
}

export type ToolSource = 'configured' | 'bundled' | 'path' | 'missing'

export interface ToolchainInfo {
  goVersion: string
  delveVersion: string
  goplsVersion: string
  goSource: ToolSource
  delveSource: ToolSource
  goplsSource: ToolSource
}

export type ServerStatus = 'starting' | 'ready' | 'unavailable'
export type LanguageSetting = 'auto' | 'en' | 'es'
export type ThemeSetting = 'system' | 'light' | 'dark'

export interface Settings {
  language: LanguageSetting
  theme: ThemeSetting
  fontSize: number
  goPath: string
  delvePath: string
  goplsPath: string
  firstRun: boolean
  lastFolder: string
  formatOnSave: boolean
  /** Last opened files, newest first (at most 10). */
  recentFiles: string[]
}
