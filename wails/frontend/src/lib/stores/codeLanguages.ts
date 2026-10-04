// The programming languages the IDE knows (not the UI language: that is ../language.ts).
import { derived, get, writable } from 'svelte/store'
import type { Bridge, ToolId, Unsubscribe } from '../bridge'
import type { Capabilities, CodeLanguage, LanguageProfile, Settings, ToolStatus } from '../domain'
import { debuggedPath } from './debug'
import { activePath } from './files'
import { settings } from './settings'

/** The profiles the backend declared, loaded once at start. */
export const profiles = writable<LanguageProfile[]>([])
/** The tools the backend found, one entry per ToolSpec of every profile. */
export const tools = writable<ToolStatus[]>([])

const FALLBACK_CODE_LANGUAGE: CodeLanguage = 'go'

const extensionOf = (path: string): string => {
  const name = path.split(/[\\/]/).pop() ?? path
  const dot = name.lastIndexOf('.')
  return dot < 0 ? '' : name.slice(dot).toLowerCase()
}

/** The profile that owns a file, by its lower-case extension; null when none does. */
export const profileOf = (
  path: string,
  all: LanguageProfile[] = get(profiles)
): LanguageProfile | null => {
  const extension = extensionOf(path)
  if (!extension) return null
  return all.find((profile) => profile.extensions.includes(extension)) ?? null
}

/** The programming language of a file; null for a file no profile claims. Untitled names work too. */
export const codeLanguageOf = (
  path: string,
  all: LanguageProfile[] = get(profiles)
): CodeLanguage | null => profileOf(path, all)?.id ?? null

const defaultCodeLanguageOf = (current: Settings | null): CodeLanguage =>
  current?.defaultCodeLanguage || FALLBACK_CODE_LANGUAGE

/** The language of the open file; with no file (or an unknown one) the default for new files. */
export const activeCodeLanguage = derived(
  [activePath, profiles, settings],
  ([path, all, current]): CodeLanguage =>
    (path ? codeLanguageOf(path, all) : null) ?? defaultCodeLanguageOf(current)
)

export const activeProfile = derived(
  [activeCodeLanguage, profiles],
  ([id, all]): LanguageProfile | null => all.find((profile) => profile.id === id) ?? null
)

/** What buttons and panels apply to the open file's language; null until the profiles load. */
export const capabilities = derived(activeProfile, (profile): Capabilities | null =>
  profile ? profile.capabilities : null
)

/** The profiles the student chose to use (`settings.enabledCodeLanguages`); empty means all of them. */
export const enabledProfiles = derived(
  [profiles, settings],
  ([all, current]): LanguageProfile[] => {
    const chosen = current?.enabledCodeLanguages ?? []
    return chosen.length === 0 ? all : all.filter((profile) => chosen.includes(profile.id))
  }
)

/** True when the program being debugged can read the keyboard (its language says so). */
export const debugInputEnabled = derived([debuggedPath, profiles], ([path, all]): boolean => {
  const profile = path ? profileOf(path, all) : null
  return profile?.capabilities.debugInput ?? false
})

/** The status of one tool; undefined until the backend reported it. */
export const toolStatus = (id: string, all: ToolStatus[] = get(tools)): ToolStatus | undefined =>
  all.find((status) => status.id === id)

/** Asks again which tools exist and where they come from (after a tool path changed). */
export const refreshTools = async (bridge: Bridge): Promise<void> => {
  tools.set(await bridge.codeLanguages.tools())
}

const toolPathsDiffer = (a: Settings, b: Settings): boolean => {
  const ids = new Set([...Object.keys(a.toolPaths), ...Object.keys(b.toolPaths)])
  return [...ids].some((id) => (a.toolPaths[id] ?? '') !== (b.toolPaths[id] ?? ''))
}

/** Loads the profiles and the tools, and reloads the tools whenever a tool path changes. */
export const connectCodeLanguages = async (bridge: Bridge): Promise<Unsubscribe> => {
  let before = get(settings)
  const off = settings.subscribe((next) => {
    if (before && next && toolPathsDiffer(before, next)) void refreshTools(bridge)
    before = next
  })
  profiles.set(await bridge.codeLanguages.profiles())
  await refreshTools(bridge)
  return off
}

/** "Choose…" next to a tool path: the native dialog saves the path, then the tools are re-detected. */
export const pickTool = async (bridge: Bridge, toolId: ToolId): Promise<void> => {
  await bridge.settings.pickExecutable(toolId)
  await refreshTools(bridge)
}
