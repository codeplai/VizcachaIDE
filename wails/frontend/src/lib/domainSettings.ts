// Part of the domain.ts contract: the user's settings and the update state. Import it through
// domain.ts.
import type { CodeLanguage } from './domainCodeLanguage'

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
  /** Look for a new version at start (at most once a day) and download it. */
  checkUpdates: boolean
  /** RFC 3339 time of the last successful update check; '' if never. */
  lastUpdateCheck: string
}

export type UpdateStatus =
  'idle' | 'checking' | 'upToDate' | 'available' | 'downloading' | 'ready' | 'failed'

/** A published version of the IDE and the file for this installation. */
export interface Release {
  version: string
  /** Release notes (Markdown). */
  notes: string
  notesUrl: string
  publishedAt: string
  asset: string
  assetSize: number
}

/** What the IDE knows about updates (event update:state). */
export interface UpdateState {
  status: UpdateStatus
  current: string
  latest: Release | null
  downloadedBytes: number
  totalBytes: number
  /** True when Install runs the installer and restarts; false: it shows the downloaded file. */
  installs: boolean
  checkedAt: string
  error: string
}
