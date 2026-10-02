// Keeps the open tabs, buffers, language server and recent files in step with
// what the file tree renames or deletes.
import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import { isInside } from './fileTreeEdit'
import { activePath, buffers, closeFile, dirty, openTabs } from './files'
import { selectedNode } from './layout'
import { forgetRecentFile } from './recentFiles'
import { settings, updateSettings } from './settings'

const openPathsUnder = (path: string): string[] =>
  get(openTabs).filter((tab) => isInside(tab, path))

/** True when a file inside `path` (or `path` itself) has changes that were not saved. */
export const hasUnsavedUnder = (path: string): boolean => {
  const changed = get(dirty)
  return openPathsUnder(path).some((tab) => changed[tab])
}

const rekey = <T>(all: Record<string, T>, rename: (key: string) => string): Record<string, T> =>
  Object.fromEntries(Object.entries(all).map(([key, value]) => [rename(key), value]))

/** After renaming `from` to `to` (a file or a folder), every open file under it follows. */
export const moveOpenFiles = async (bridge: Bridge, from: string, to: string): Promise<void> => {
  const moved = (path: string): string =>
    isInside(path, from) ? to + path.slice(from.length) : path
  const affected = openPathsUnder(from)
  const texts = get(buffers)
  buffers.update((all) => rekey(all, moved))
  dirty.update((all) => rekey(all, moved))
  openTabs.update((tabs) => tabs.map(moved))
  activePath.update((path) => (path ? moved(path) : path))
  selectedNode.update((path) => (path ? moved(path) : path))
  const recent = get(settings)?.recentFiles ?? []
  if (recent.some((path) => isInside(path, from))) {
    await updateSettings(bridge, { recentFiles: recent.map(moved) })
  }
  for (const path of affected) {
    await bridge.language.closeDocument(path)
    await bridge.language.openDocument(moved(path), texts[path] ?? '')
  }
}

/** After deleting `path`, its open tabs close and it leaves "Open recent". */
export const closeOpenFilesUnder = async (bridge: Bridge, path: string): Promise<void> => {
  for (const tab of openPathsUnder(path)) await closeFile(bridge, tab)
  for (const recent of get(settings)?.recentFiles ?? []) {
    if (isInside(recent, path)) await forgetRecentFile(bridge, recent)
  }
  selectedNode.update((current) => (current && isInside(current, path) ? null : current))
}
