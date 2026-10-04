import { derived, get, writable } from 'svelte/store'
import type { Bridge, Unsubscribe } from '../bridge'
import type { LanguageSetting, Settings, ToolSource } from '../domain'

export const settings = writable<Settings | null>(null)
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

/** i18n key that says where the backend found a tool (the `source` of a ToolStatus). */
export const toolSourceKey = (source: ToolSource): string => TOOL_SOURCE_KEYS[source]

const refreshLanguage = async (bridge: Bridge): Promise<void> => {
  resolvedLanguage.set(await bridge.settings.resolvedLanguage())
}

/** Loads the saved settings and the UI language, then keeps the store in sync with the backend. */
export const connectSettings = async (bridge: Bridge): Promise<Unsubscribe> => {
  const off = bridge.on('settings:changed', (next) => {
    const before = get(settings)
    settings.set(next)
    if (before && before.language !== next.language) void refreshLanguage(bridge)
  })
  settings.set(await bridge.settings.get())
  await refreshLanguage(bridge)
  return off
}

export const updateSettings = async (bridge: Bridge, change: Partial<Settings>): Promise<void> => {
  const current = get(settings)
  if (!current) return
  await bridge.settings.save({ ...current, ...change })
}
