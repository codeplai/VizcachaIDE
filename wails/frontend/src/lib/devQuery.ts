import type { Scenario } from './bridge'
import type { CodeLanguage, LanguageSetting, ThemeSetting } from './domain'
import type { DialogName } from './stores/layout'

export interface DevQuery {
  scenario: Scenario
  language: LanguageSetting | null
  /** `language=python` opens the Python sample project (`lang` is the interface language). */
  codeLanguage: Extract<CodeLanguage, 'go' | 'python'> | null
  theme: ThemeSetting | null
  /** `folder=none` starts with no folder open (the empty state of the Files panel). */
  noFolder: boolean
  /** `firstrun=1` shows the first-start wizard. */
  firstRun: boolean
  /** `closechanges=1` edits the open file and tries to close it (the "save changes?" dialog). */
  closeChanges: boolean
  /** `dialog=settings|about|modules` opens that dialog. */
  dialog: DialogName | null
}

const pick = <T extends string>(value: string | null, allowed: readonly T[]): T | null =>
  allowed.find((option) => option === value) ?? null

/**
 * Browser-only options for the mock backend, for example `?scenario=error&lang=es&theme=dark`.
 * They make the states reachable without clicking (screenshots, demos).
 */
export const parseDevQuery = (search: string): DevQuery => {
  const params = new URLSearchParams(search)
  return {
    scenario: pick(params.get('scenario'), ['write', 'error', 'debug'] as const) ?? 'write',
    language: pick(params.get('lang'), ['en', 'es'] as const),
    codeLanguage: pick(params.get('language'), ['go', 'python'] as const),
    theme: pick(params.get('theme'), ['light', 'dark'] as const),
    noFolder: params.get('folder') === 'none',
    firstRun: params.get('firstrun') === '1',
    closeChanges: params.get('closechanges') === '1',
    dialog: pick(params.get('dialog'), ['settings', 'about', 'packages'] as const)
  }
}

/** The line the dev bar's "debug" state pauses on: the sample's own (see bridge/mockPython.ts). */
export const demoBreakpointLine = (path: string): number => (path.endsWith('.py') ? 8 : 6)
