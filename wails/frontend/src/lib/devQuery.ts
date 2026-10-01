import type { Scenario } from './bridge'
import type { LanguageSetting, ThemeSetting } from './domain'

export interface DevQuery {
  scenario: Scenario
  language: LanguageSetting | null
  theme: ThemeSetting | null
}

const pick = <T extends string>(value: string | null, allowed: readonly T[]): T | null =>
  allowed.find((option) => option === value) ?? null

/**
 * Browser-only options for the mock backend, for example `?scenario=error&lang=es&theme=dark`.
 * They make the three states reachable without clicking (screenshots, demos).
 */
export const parseDevQuery = (search: string): DevQuery => {
  const params = new URLSearchParams(search)
  return {
    scenario: pick(params.get('scenario'), ['write', 'error', 'debug'] as const) ?? 'write',
    language: pick(params.get('lang'), ['en', 'es'] as const),
    theme: pick(params.get('theme'), ['light', 'dark'] as const)
  }
}
