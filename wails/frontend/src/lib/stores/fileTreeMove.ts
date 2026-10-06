// Moving entries of the Files panel to another folder (drag and drop, and Cut + Paste).
import { get, writable } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { FileNode } from '../domain'
import { selectOnly, topLevelPaths } from './fileSelection'
import { basename, dirname, findNode, isInside, joinPath } from './fileTreeEdit'
import { failed, refreshTree } from './fileTreeCommands'
import { moveOpenFiles } from './fileTreeTabs'
import { fileTree } from './files'
import { collapsedFolders } from './layout'
import { showNotice } from './notice'

export type MoveProblem = 'itself' | 'same' | 'exists'

/**
 * Why `from` cannot go into `folder`, or null when it can. `same` means it is already there
 * (nothing to do, nothing to say); `itself` is a folder dropped into itself or a descendant.
 */
export const moveProblem = (
  tree: FileNode | null,
  from: string,
  folder: string
): MoveProblem | null => {
  if (isInside(folder, from)) return 'itself'
  if (dirname(from) === folder) return 'same'
  const name = basename(from).toLowerCase()
  const taken = findNode(tree, folder)?.children.some((c) => c.name.toLowerCase() === name)
  return taken ? 'exists' : null
}

/** True when every entry can be dropped on `folder` (to highlight the drop target). */
export const canDrop = (paths: string[], folder: string): boolean => {
  const tree = get(fileTree)
  const moving = topLevelPaths(paths, tree?.path)
  if (moving.length === 0 || findNode(tree, folder)?.isDir !== true) return false
  return moving.every((path) => moveProblem(tree, path, folder) === null)
}

const explain = (problem: MoveProblem, from: string): void => {
  const key = problem === 'itself' ? 'tree.moveIntoItself' : 'tree.moveExists'
  showNotice({ messageKey: key, values: { name: basename(from) }, actions: [] })
}

const moveOne = async (bridge: Bridge, from: string, folder: string): Promise<string | null> => {
  const problem = moveProblem(get(fileTree), from, folder)
  if (problem === 'same') return null
  if (problem) {
    explain(problem, from)
    return null
  }
  const to = joinPath(folder, basename(from))
  try {
    await bridge.files.rename(from, to)
  } catch (error) {
    failed('tree.moveFailed', basename(from), error)
    return null
  }
  await moveOpenFiles(bridge, from, to)
  return to
}

/**
 * Moves the entries into `folder` (a rename, so it stays on one volume: the operating system
 * refuses a move across volumes and the panel shows that error). Open tabs follow the files.
 * Returns the new paths.
 */
export const moveEntries = async (
  bridge: Bridge,
  paths: string[],
  folder: string
): Promise<string[]> => {
  const moved: string[] = []
  for (const from of topLevelPaths(paths, get(fileTree)?.path)) {
    const to = await moveOne(bridge, from, folder)
    if (to) moved.push(to)
  }
  await refreshTree(bridge)
  collapsedFolders.update((all) => new Set([...all].filter((path) => path !== folder)))
  if (moved.length > 0) selectOnly(moved)
  return moved
}

/** The entries being dragged, and the folder under the pointer when it can take them. */
export const dragPaths = writable<string[]>([])
export const dropTarget = writable<string | null>(null)

/** Drag start: dragging a row outside the selection drags only that row. */
export const startDrag = (path: string, selected: string[]): string[] => {
  const paths = selected.includes(path) ? selected : [path]
  dragPaths.set(paths)
  return paths
}

export const endDrag = (): void => {
  dragPaths.set([])
  dropTarget.set(null)
}

/** Drop on a folder (or the root): moves what is dragged, or says why it cannot. */
export const dropEntries = async (bridge: Bridge, folder: string): Promise<void> => {
  const paths = get(dragPaths)
  endDrag()
  if (canDrop(paths, folder)) {
    await moveEntries(bridge, paths, folder)
    return
  }
  const tree = get(fileTree)
  for (const from of topLevelPaths(paths, tree?.path)) {
    const problem = moveProblem(tree, from, folder)
    if (problem && problem !== 'same') explain(problem, from)
  }
}
