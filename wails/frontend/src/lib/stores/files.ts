import { derived, get, writable } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { FileNode } from '../domain'
import { problems } from './diagnostics'
import { updateSettings } from './settings'

export const fileTree = writable<FileNode | null>(null)
export const openTabs = writable<string[]>([])
export const activePath = writable<string | null>(null)
/** Text of every open file, as the editor shows it. */
export const buffers = writable<Record<string, string>>({})
export const dirty = writable<Record<string, boolean>>({})

export const baseName = (path: string): string => path.split(/[\\/]/).pop() ?? path
export const parentName = (path: string): string => {
  const parts = path.split(/[\\/]/)
  return parts.length > 1 ? (parts[parts.length - 2] ?? '') : ''
}

export const activeFileName = derived(activePath, (path) => (path ? baseName(path) : ''))
export const activeText = derived([activePath, buffers], ([path, all]) =>
  path ? (all[path] ?? '') : ''
)

/** Names of the files that have at least one problem (for the dot in the file tree). */
export const filesWithProblems = derived(problems, (items) =>
  items.flatMap((item) => (item.diagnostic.location ? [item.diagnostic.location.file] : []))
)

/** Asks for a folder, shows its files and remembers it for the next start. */
export const openFolder = async (bridge: Bridge): Promise<void> => {
  const tree = await bridge.files.openFolder()
  if (!tree?.path) return
  fileTree.set(tree)
  await updateSettings(bridge, { lastFolder: tree.path })
}

export const openFile = async (bridge: Bridge, path: string): Promise<void> => {
  if (!(path in get(buffers))) {
    const text = await bridge.files.readFile(path)
    buffers.update((all) => ({ ...all, [path]: text }))
    await bridge.language.openDocument(path, text)
  }
  openTabs.update((tabs) => (tabs.includes(path) ? tabs : [...tabs, path]))
  activePath.set(path)
}

export const editBuffer = (path: string, text: string): void => {
  buffers.update((all) => ({ ...all, [path]: text }))
  dirty.update((all) => ({ ...all, [path]: true }))
}

const closeTab = (path: string): void => {
  const tabs = get(openTabs).filter((tab) => tab !== path)
  openTabs.set(tabs)
  if (get(activePath) === path) activePath.set(tabs[tabs.length - 1] ?? null)
}

const without = <T>(all: Record<string, T>, path: string): Record<string, T> =>
  Object.fromEntries(Object.entries(all).filter(([key]) => key !== path))

/** Closes the tab and forgets its text (saved or discarded), also for the language server. */
export const closeFile = async (bridge: Bridge, path: string): Promise<void> => {
  closeTab(path)
  buffers.update((all) => without(all, path))
  dirty.update((all) => without(all, path))
  await bridge.language.closeDocument(path)
}
