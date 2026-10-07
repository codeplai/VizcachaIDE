// Copy, Cut, Paste and Duplicate of the Files panel. The clipboard is the panel's own: it never
// touches the system clipboard (that is what "Copy path" is for), so the editor keeps its Ctrl+C.
import { get, writable } from 'svelte/store'
import type { Bridge } from '../bridge'
import { actionTargets, selectOnly, topLevelPaths } from './fileSelection'
import { basename, dirname, findNode, isInside, joinPath } from './fileTreeEdit'
import { failed, folderOf, refreshTree } from './fileTreeCommands'
import { moveEntries } from './fileTreeMove'
import { copyName } from './fileTreeNames'
import { fileTree } from './files'
import { collapsedFolders, selectedNode } from './layout'
import { showNotice } from './notice'

export interface PanelClipboard {
  mode: 'copy' | 'cut'
  paths: string[]
}

export const panelClipboard = writable<PanelClipboard | null>(null)

/** Copy: remembers the entries (the row acted on, or all the selected ones). */
export const copyEntries = (path: string): void =>
  panelClipboard.set({ mode: 'copy', paths: actionTargets(path) })

/** Cut: remembers the entries; they move when pasted. */
export const cutEntries = (path: string): void =>
  panelClipboard.set({ mode: 'cut', paths: actionTargets(path) })

const copyOne = async (bridge: Bridge, from: string, folder: string, taken: string[]) => {
  const node = findNode(get(fileTree), from)
  if (!node) {
    showNotice({ messageKey: 'tree.sourceMissing', values: { name: basename(from) }, actions: [] })
    return null
  }
  if (node.isDir && isInside(folder, from)) {
    showNotice({ messageKey: 'tree.moveIntoItself', values: { name: node.name }, actions: [] })
    return null
  }
  const name = copyName(node.name, node.isDir, taken, dirname(from) === folder)
  const to = joinPath(folder, name)
  try {
    await bridge.files.copy(from, to)
  } catch (error) {
    failed('tree.copyEntryFailed', node.name, error)
    return null
  }
  taken.push(name)
  return to
}

/** Copies each entry into the folder `folderFor` gives for it; returns the new paths. */
const copyAll = async (
  bridge: Bridge,
  sources: string[],
  folderFor: (source: string) => string
): Promise<string[]> => {
  const created: string[] = []
  const takenBy = new Map<string, string[]>()
  for (const from of sources) {
    const folder = folderFor(from)
    const taken =
      takenBy.get(folder) ?? findNode(get(fileTree), folder)?.children.map((c) => c.name) ?? []
    takenBy.set(folder, taken)
    const to = await copyOne(bridge, from, folder, taken)
    if (to) created.push(to)
  }
  await refreshTree(bridge)
  collapsedFolders.update(
    (all) => new Set([...all].filter((path) => !created.some((to) => dirname(to) === path)))
  )
  if (created.length > 0) selectOnly(created)
  return created
}

/** Paste into the selected folder, the folder of the selected file, or the root. */
export const pasteEntries = async (bridge: Bridge, folder?: string): Promise<void> => {
  const clip = get(panelClipboard)
  const target = folder ?? folderOf(get(selectedNode))
  if (!clip || !target) return
  if (clip.mode === 'copy') {
    await copyAll(bridge, topLevelPaths(clip.paths, get(fileTree)?.path), () => target)
    return
  }
  await moveEntries(bridge, clip.paths, target)
  panelClipboard.set(null)
}

/** Duplicate: a copy of each entry next to it, named `name (copy)`. */
export const duplicateEntries = async (bridge: Bridge, path: string): Promise<void> => {
  await copyAll(bridge, actionTargets(path), dirname)
}
