// Mirror of internal/bridge/events.go. FROZEN contract: a Go test checks that both
// files list the same event names. Change them together (Contract change request).
import type {
  DebugState,
  ExplainedDiagnostic,
  RunConfiguration,
  ServerStatus,
  Settings,
  Variable,
  Diagnostic
} from './domain'

export const Events = {
  runOutput: 'run:output',
  runStarted: 'run:started',
  runFinished: 'run:finished',
  debugStopped: 'debug:stopped',
  debugVariables: 'debug:variables',
  debugOutput: 'debug:output',
  debugTerminated: 'debug:terminated',
  lspDiagnostics: 'lsp:diagnostics',
  lspStatus: 'lsp:status',
  assistantExplained: 'assistant:explained',
  settingsChanged: 'settings:changed',
  fileChanged: 'file:changed'
} as const

export type EventName = (typeof Events)[keyof typeof Events]

export interface RunOutputPayload {
  stream: 'stdout' | 'stderr'
  text: string
}

export interface RunFinishedPayload {
  exitCode: number
  durationMs: number
}

export interface DebugVariablesPayload {
  reference: number
  variables: Variable[]
}

export interface DebugOutputPayload {
  text: string
  category: 'stdout' | 'stderr' | 'console'
}

/** `exitCode` is -1 when the user stopped the session. */
export interface DebugTerminatedPayload {
  exitCode: number
}

export interface DiagnosticsPayload {
  path: string
  diagnostics: Diagnostic[]
}

/** An open file changed on disk outside the IDE. */
export interface FileChangedPayload {
  path: string
}

export const TERMINATED_BY_USER = -1

export interface EventPayloads {
  'run:output': RunOutputPayload
  'run:started': RunConfiguration
  'run:finished': RunFinishedPayload
  'debug:stopped': DebugState
  'debug:variables': DebugVariablesPayload
  'debug:output': DebugOutputPayload
  'debug:terminated': DebugTerminatedPayload
  'lsp:diagnostics': DiagnosticsPayload
  'lsp:status': ServerStatus
  'assistant:explained': ExplainedDiagnostic[]
  'settings:changed': Settings
  'file:changed': FileChangedPayload
}
