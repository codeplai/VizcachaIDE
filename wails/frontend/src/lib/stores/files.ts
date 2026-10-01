import { derived, get, writable } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { FileNode } from '../domain'

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

export const openFolder = async (bridge: Bridge): Promise<void> => {
  fileTree.set(await bridge.files.openFolder())
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

export const closeTab = (path: string): void => {
  const tabs = get(openTabs).filter((tab) => tab !== path)
  openTabs.set(tabs)
  if (get(activePath) === path) activePath.set(tabs[tabs.length - 1] ?? null)
}
