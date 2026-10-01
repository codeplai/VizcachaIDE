import { get, writable } from 'svelte/store'
import type { Bridge, Unsubscribe } from '../bridge'
import type { Settings, ToolchainInfo } from '../domain'
import { readPreference, writePreference } from './preferences'

export const settings = writable<Settings | null>(null)
export const toolchain = writable<ToolchainInfo | null>(null)

/**
 * "Format when saving" is not part of the backend settings yet (see the Contract change request),
 * so it is kept in the browser storage.
 */
export const formatOnSave = writable<boolean>(readPreference('formatOnSave', true))
formatOnSave.subscribe((value) => writePreference('formatOnSave', value))

export const MIN_FONT_SIZE = 10
export const MAX_FONT_SIZE = 28

/** Where the IDE got a tool from: the path the user typed or the automatic search. */
export type ToolOrigin = 'custom' | 'automatic'

export const toolOrigin = (path: string): ToolOrigin =>
  path.trim() === '' ? 'automatic' : 'custom'

export const clampFontSize = (size: number): number =>
  Math.min(MAX_FONT_SIZE, Math.max(MIN_FONT_SIZE, Math.round(size) || MIN_FONT_SIZE))

/** Loads the saved settings and the tool versions, then keeps the store in sync with the backend. */
export const connectSettings = async (bridge: Bridge): Promise<Unsubscribe> => {
  const off = bridge.on('settings:changed', (next) => settings.set(next))
  settings.set(await bridge.settings.get())
  toolchain.set(await bridge.run.toolchain())
  return off
}

export const updateSettings = async (bridge: Bridge, change: Partial<Settings>): Promise<void> => {
  const current = get(settings)
  if (!current) return
  await bridge.settings.save({ ...current, ...change })
}
