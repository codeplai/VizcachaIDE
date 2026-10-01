import { derived, get, writable } from 'svelte/store'
import type { Bridge, ToolId, Unsubscribe } from '../bridge'
import type { LanguageSetting, Settings, ToolSource, ToolchainInfo } from '../domain'

export const settings = writable<Settings | null>(null)
export const toolchain = writable<ToolchainInfo | null>(null)
/** The language the backend resolved for "auto" (the system language), null until it answers. */
export const resolvedLanguage = writable<'en' | 'es' | null>(null)

/** The language to show: the one chosen, or the backend's answer when the choice is "auto". */
export const interfaceLanguage = derived(
  [settings, resolvedLanguage],
  ([current, resolved]): LanguageSetting =>
    current?.language === 'auto' && resolved ? resolved : (current?.language ?? 'auto')
)

export const DEFAULT_FONT_SIZE = 14
export const MIN_FONT_SIZE = 10
export const MAX_FONT_SIZE = 28

export const clampFontSize = (size: number): number =>
  Math.min(MAX_FONT_SIZE, Math.max(MIN_FONT_SIZE, Math.round(size) || MIN_FONT_SIZE))

const TOOL_SOURCE_KEYS: Record<ToolSource, string> = {
  configured: 'settings.originCustom',
  bundled: 'settings.originBundled',
  path: 'settings.originPath',
  missing: 'settings.originMissing'
}

/** i18n key that says where the backend found a tool (the `*Source` fields of ToolchainInfo). */
export const toolSourceKey = (source: ToolSource): string => TOOL_SOURCE_KEYS[source]

const toolPathsDiffer = (a: Settings, b: Settings): boolean =>
  a.goPath !== b.goPath || a.delvePath !== b.delvePath || a.goplsPath !== b.goplsPath

/** Asks again which tools exist and where they come from (after a tool path changed). */
const refreshToolchain = async (bridge: Bridge): Promise<void> => {
  toolchain.set(await bridge.run.toolchain())
}

const refreshLanguage = async (bridge: Bridge): Promise<void> => {
  resolvedLanguage.set(await bridge.settings.resolvedLanguage())
}

/** Loads the saved settings and the tool versions, then keeps the store in sync with the backend. */
export const connectSettings = async (bridge: Bridge): Promise<Unsubscribe> => {
  const off = bridge.on('settings:changed', (next) => {
    const before = get(settings)
    settings.set(next)
    if (before && toolPathsDiffer(before, next)) void refreshToolchain(bridge)
    if (before && before.language !== next.language) void refreshLanguage(bridge)
  })
  settings.set(await bridge.settings.get())
  await Promise.all([refreshToolchain(bridge), refreshLanguage(bridge)])
  return off
}

export const updateSettings = async (bridge: Bridge, change: Partial<Settings>): Promise<void> => {
  const current = get(settings)
  if (!current) return
  await bridge.settings.save({ ...current, ...change })
}

/** "Choose…" next to a tool path: the native dialog saves the path and re-detects the tools. */
export const pickTool = async (bridge: Bridge, tool: ToolId): Promise<void> => {
  toolchain.set(await bridge.settings.pickExecutable(tool))
}
