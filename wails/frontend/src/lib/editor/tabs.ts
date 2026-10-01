// Open-file tabs: modified dot and the "save before closing?" request that the shell paints.
import { derived, get, writable } from 'svelte/store'
import type { Bridge } from '../bridge'
import { activePath, baseName, buffers, closeTab, dirty, openTabs } from '../stores/files'
import { saveDocument } from './saving'

export interface TabItem {
  path: string
  name: string
  modified: boolean
  active: boolean
}

export const tabItems = derived([openTabs, activePath, dirty], ([paths, active, changed]) =>
  paths.map((path): TabItem => ({
    path,
    name: baseName(path),
    modified: !!changed[path],
    active: path === active
  }))
)

/** Set when the user closes a tab with unsaved changes. The shell shows `confirm.closeChanges`. */
export interface CloseRequest {
  path: string
  file: string
}

export type CloseAnswer = 'save' | 'discard' | 'cancel'

export const closeRequest = writable<CloseRequest | null>(null)

const without = <T>(all: Record<string, T>, path: string): Record<string, T> =>
  Object.fromEntries(Object.entries(all).filter(([key]) => key !== path))

const removeTab = async (bridge: Bridge, path: string, forgetChanges: boolean): Promise<void> => {
  closeTab(path)
  if (forgetChanges) {
    buffers.update((all) => without(all, path))
    dirty.update((all) => without(all, path))
  }
  await bridge.language.closeDocument(path)
}

/** Closes the tab, or asks first when it has unsaved changes. */
export const requestCloseTab = async (bridge: Bridge, path: string): Promise<void> => {
  if (get(dirty)[path]) {
    closeRequest.set({ path, file: baseName(path) })
    return
  }
  await removeTab(bridge, path, false)
}

export const answerCloseRequest = async (bridge: Bridge, answer: CloseAnswer): Promise<void> => {
  const request = get(closeRequest)
  closeRequest.set(null)
  if (!request || answer === 'cancel') return
  if (answer === 'save' && !(await saveDocument(bridge, request.path))) return
  await removeTab(bridge, request.path, answer === 'discard')
}

export const saveActiveDocument = async (bridge: Bridge): Promise<void> => {
  const path = get(activePath)
  if (path) await saveDocument(bridge, path)
}
