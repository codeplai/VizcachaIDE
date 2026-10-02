import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import { settings, updateSettings } from './settings'

/** Same limit as `domain.MaxRecentFiles` in the backend. */
export const MAX_RECENT_FILES = 10

const currentList = (): string[] => get(settings)?.recentFiles ?? []

/** Puts the file at the top of "Open recent" (once, newest first, at most 10). */
export const rememberRecentFile = async (bridge: Bridge, path: string): Promise<void> => {
  const list = currentList()
  if (list[0] === path) return
  const next = [path, ...list.filter((item) => item !== path)].slice(0, MAX_RECENT_FILES)
  await updateSettings(bridge, { recentFiles: next })
}

export const forgetRecentFile = async (bridge: Bridge, path: string): Promise<void> => {
  const list = currentList()
  if (!list.includes(path)) return
  await updateSettings(bridge, { recentFiles: list.filter((item) => item !== path) })
}

export const clearRecentFiles = async (bridge: Bridge): Promise<void> => {
  if (currentList().length === 0) return
  await updateSettings(bridge, { recentFiles: [] })
}
