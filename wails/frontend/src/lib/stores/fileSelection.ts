// Which rows of the Files panel are selected: click, Ctrl/Cmd+click and Shift+click rules.
import { derived, get, writable } from 'svelte/store'
import type { FileNode } from '../domain'
import { findNode, isInside } from './fileTreeEdit'
import { fileTree } from './files'
import { collapsedFolders, selectedNode } from './layout'

/** The paths selected with Ctrl/Cmd+click or Shift+click (a plain click selects only one). */
export const selection = writable<string[]>([])
/** Where a Shift+click range starts: the row clicked last without Shift. */
const rangeStart = writable<string | null>(null)

export interface ClickModifiers {
  /** Ctrl (Cmd on macOS) was held: toggle the row. */
  toggle: boolean
  /** Shift was held: select from the range start to the row. */
  range: boolean
}

/** The selected rows: the selection, or the focused row alone when it is outside of it. */
export const selectedPaths = derived([selection, selectedNode], ([paths, focused]) =>
  focused && !paths.includes(focused) ? [focused] : paths
)

/** Paths of the rows on screen, top to bottom (the children of collapsed folders are not). */
export const visiblePaths = (tree: FileNode | null, collapsed: Set<string>): string[] => {
  if (!tree) return []
  const below = collapsed.has(tree.path)
    ? []
    : tree.children.flatMap((c) => visiblePaths(c, collapsed))
  return [tree.path, ...below]
}

/** The new selection after a click on `path` (pure: the rules of the panel). */
export const selectionAfterClick = (
  current: string[],
  start: string | null,
  path: string,
  visible: string[],
  modifiers: ClickModifiers
): { paths: string[]; start: string | null } => {
  if (modifiers.range && start && visible.includes(start) && visible.includes(path)) {
    const [one, other] = [visible.indexOf(start), visible.indexOf(path)]
    return { paths: visible.slice(Math.min(one, other), Math.max(one, other) + 1), start }
  }
  if (modifiers.toggle) {
    const paths = current.includes(path) ? current.filter((p) => p !== path) : [...current, path]
    return { paths, start: path }
  }
  return { paths: [path], start: path }
}

/** Applies a click on a row; the clicked row also becomes the focused one (`selectedNode`). */
export const selectEntry = (path: string, modifiers: ClickModifiers): void => {
  const current = get(selectedPaths)
  const visible = visiblePaths(get(fileTree), get(collapsedFolders))
  const next = selectionAfterClick(current, get(rangeStart), path, visible, modifiers)
  selection.set(next.paths)
  rangeStart.set(next.start)
  selectedNode.set(path)
}

/** Selects exactly these paths (after a paste or a move, the new entries). */
export const selectOnly = (paths: string[]): void => {
  selection.set(paths)
  rangeStart.set(paths[0] ?? null)
  selectedNode.set(paths[paths.length - 1] ?? null)
}

/** Escape: nothing selected. */
export const clearSelection = (): void => {
  selection.set([])
  rangeStart.set(null)
  selectedNode.set(null)
}

/** Without the open folder itself and without what is inside another selected folder. */
export const topLevelPaths = (paths: string[], root: string | undefined): string[] => {
  const chosen = [...new Set(paths)].filter((path) => path !== root)
  return chosen.filter((path) => !chosen.some((other) => other !== path && isInside(path, other)))
}

/** The entries an action (delete, cut, copy) applies to when the row `path` is acted on. */
export const actionTargets = (path: string): string[] => {
  const chosen = get(selectedPaths)
  const paths = chosen.includes(path) ? chosen : [path]
  return topLevelPaths(paths, get(fileTree)?.path)
}

/** Drops from the selection the entries that no longer exist (after the tree is read again). */
export const pruneSelection = (tree: FileNode | null): void => {
  selection.update((paths) => paths.filter((path) => findNode(tree, path)))
}
