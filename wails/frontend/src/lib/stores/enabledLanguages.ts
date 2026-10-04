// Which programming languages the student uses (Settings > General and the first-run wizard).
import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { CodeLanguage } from '../domain'
import { enabledProfiles, profiles } from './codeLanguages'
import { settings, updateSettings } from './settings'

const FALLBACK: CodeLanguage = 'go'

/**
 * The list to save after the student ticked or cleared one language. At least one stays ticked
 * (null: nothing changes), and ticking every language saves the empty list, which means "all",
 * so a language added by a later version is not hidden.
 */
export const toggledLanguages = (
  all: CodeLanguage[],
  enabled: CodeLanguage[],
  id: CodeLanguage
): CodeLanguage[] | null => {
  const next = enabled.includes(id) ? enabled.filter((one) => one !== id) : [...enabled, id]
  if (next.length === 0) return null
  return all.every((one) => next.includes(one)) ? [] : all.filter((one) => next.includes(one))
}

/** Ticks or clears a language; the default language for new files follows if it was cleared. */
export const toggleEnabledLanguage = async (bridge: Bridge, id: CodeLanguage): Promise<void> => {
  const all = get(profiles).map((profile) => profile.id)
  const enabled = get(enabledProfiles).map((profile) => profile.id)
  const next = toggledLanguages(all, enabled, id)
  if (next === null) return
  const current = get(settings)?.defaultCodeLanguage
  const kept = next.length === 0 || (current !== undefined && next.includes(current))
  await updateSettings(bridge, {
    enabledCodeLanguages: next,
    ...(kept ? {} : { defaultCodeLanguage: next[0] })
  })
}

/** The language a new file starts in: the default one if it is enabled, else the first enabled. */
export const startingCodeLanguage = (preferred: CodeLanguage | undefined): CodeLanguage => {
  const enabled = get(enabledProfiles).map((profile) => profile.id)
  if (preferred && enabled.includes(preferred)) return preferred
  return enabled[0] ?? preferred ?? FALLBACK
}
