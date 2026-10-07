import { derived, get, writable } from 'svelte/store'
import type { Bridge } from '../bridge'
import { projectFiles } from '../search/skip'
import { rankFiles, type QuickOpenItem } from '../search/quickOpenRank'
import { fileTree, openFile } from './files'
import { settings } from './settings'

/** The Quick open palette (Ctrl+P) is shown. */
export const quickOpenVisible = writable(false)
export const quickOpenQuery = writable('')

export const openQuickOpen = async (): Promise<void> => {
  if (get(fileTree) === null) return // no folder, nothing to offer
  quickOpenQuery.set('')
  quickOpenVisible.set(true)
}

export const closeQuickOpen = (): void => quickOpenVisible.set(false)

/** The files the palette lists for what is typed. */
export const quickOpenItems = derived(
  [fileTree, quickOpenQuery, settings],
  ([tree, query, current]): QuickOpenItem[] =>
    tree
      ? rankFiles(
          tree.path,
          projectFiles(tree).map((node) => node.path),
          query,
          current?.recentFiles ?? []
        )
      : []
)

/** Opens the chosen file in the editor and closes the palette. */
export const chooseQuickOpen = async (bridge: Bridge, item: QuickOpenItem): Promise<void> => {
  closeQuickOpen()
  await openFile(bridge, item.path)
}
