import type { UpdateState } from '../domain'

/** Mirrors bridge.UpdatesService (Go). Progress and results also arrive as `update:state`. */
export interface UpdatesApi {
  state: () => Promise<UpdateState>
  check: () => Promise<UpdateState>
  /** Starts the download in the background. */
  download: () => Promise<void>
  /** Runs the installer and closes the IDE, or shows the downloaded file (portable copies). */
  install: () => Promise<void>
}

/** Mirrors bridge.TerminalService (Go). Output and exit arrive as `terminal:output` and `terminal:exit`. */
export interface TerminalApi {
  /** Opens a shell in `dir` (the home folder when empty) and returns the id of the session. */
  start: (dir: string, cols: number, rows: number) => Promise<string>
  /** What the user typed or pasted, unchanged. */
  write: (id: string, data: string) => Promise<void>
  resize: (id: string, cols: number, rows: number) => Promise<void>
  /** Ends the shell and everything it started. */
  close: (id: string) => Promise<void>
}
