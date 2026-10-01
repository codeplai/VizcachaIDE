import { get, writable } from 'svelte/store'
import type { Bridge, Unsubscribe } from '../bridge'
import type { Settings, ToolchainInfo } from '../domain'

export const settings = writable<Settings | null>(null)
export const toolchain = writable<ToolchainInfo | null>(null)

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
