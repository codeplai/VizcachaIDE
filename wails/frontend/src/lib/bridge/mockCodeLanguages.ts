// Language profiles, detected tools and package commands of the mock bridge (see mock.ts).
import type { CodeLanguage } from '../domain'
import { languageProfiles } from './languageProfiles'
import { type SettingsHolder, toolsFor } from './mockSettings'
import { searchDemoIndex } from './mockPackageIndex'
import type { CodeLanguagesApi, PackagesApi } from './types'

/** The language of a path by its extension; Go when nothing matches (untitled names too). */
export const codeLanguageOfPath = (path: string): CodeLanguage => {
  const lower = path.toLowerCase()
  const profile = languageProfiles.find((p) => p.extensions.some((ext) => lower.endsWith(ext)))
  return profile?.id ?? 'go'
}

export const mockCodeLanguages = (state: SettingsHolder): CodeLanguagesApi => ({
  profiles: async () => languageProfiles,
  tools: async () => toolsFor(state.settings)
})

/** The demo has no real packages: every verb succeeds silently. */
export const mockPackages = (): PackagesApi => ({
  init: async () => {},
  add: async () => {},
  remove: async () => {},
  tidy: async () => {},
  list: async () => {},
  search: async (codeLanguage, query) => searchDemoIndex(codeLanguage, query)
})
