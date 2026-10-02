// Commands of the Files panel: create, rename, delete (Recycle Bin), reveal and copy path.
import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import { confirmDelete } from './confirm'
import {
  GO_TEMPLATE,
  countFiles,
  dirname,
  findNode,
  finalName,
  joinPath,
  nameProblem,
  treeEdit,
  type TreeEditKind
} from './fileTreeEdit'
import { closeOpenFilesUnder, hasUnsavedUnder, moveOpenFiles } from './fileTreeTabs'
import { fileTree, openFile } from './files'
import { collapsedFolders, selectedNode } from './layout'
import { dismissNotice, notice, showNotice } from './notice'

const COPIED_NOTICE_MS = 2500

const reasonOf = (error: unknown): string =>
  error instanceof Error ? error.message : String(error ?? '')

const failed = (messageKey: string, name: string, error: unknown): void =>
  showNotice({ messageKey, values: { name }, actions: [], detail: reasonOf(error) })

/** Reads the tree again from disk (after any change made by the panel). */
export const refreshTree = async (bridge: Bridge): Promise<void> => {
  const root = get(fileTree)?.path
  if (!root) return
  fileTree.set(await bridge.files.listTree(root))
}

/** The folder where "New file" goes for a given row: the folder itself, or the one holding the file. */
export const folderOf = (path: string | null): string | null => {
  const tree = get(fileTree)
  if (!tree) return null
  const node = path ? findNode(tree, path) : null
  if (!node) return tree.path
  return node.isDir ? node.path : dirname(node.path)
}

export const cancelEdit = (): void => treeEdit.set(null)

/** Shows the empty name box inside `folderPath` (which is expanded first). */
export const startCreate = (kind: 'file' | 'folder', folderPath: string | null): void => {
  const folder = folderPath ?? get(fileTree)?.path
  if (!folder) return
  collapsedFolders.update((all) => new Set([...all].filter((path) => path !== folder)))
  treeEdit.set({ kind, path: folder, isDir: kind === 'folder', error: null })
}

/** Turns the row into a name box. The root folder cannot be renamed. */
export const startRename = (path: string): void => {
  const node = findNode(get(fileTree), path)
  if (!node || path === get(fileTree)?.path) return
  treeEdit.set({ kind: 'rename', path, isDir: node.isDir, error: null })
}

export const clearEditError = (): void =>
  treeEdit.update((edit) => (edit?.error ? { ...edit, error: null } : edit))

const createEntry = async (bridge: Bridge, kind: TreeEditKind, folder: string, name: string) => {
  const path = joinPath(folder, name)
  if (kind === 'folder') await bridge.files.createFolder(path)
  else await bridge.files.createFile(path, name.endsWith('.go') ? GO_TEMPLATE : '')
  await refreshTree(bridge)
  selectedNode.set(path)
  if (kind === 'file' && name.endsWith('.go')) await openFile(bridge, path)
}

const renameEntry = async (bridge: Bridge, from: string, name: string): Promise<void> => {
  const to = joinPath(dirname(from), name)
  if (to === from) return
  await bridge.files.rename(from, to)
  await refreshTree(bridge)
  await moveOpenFiles(bridge, from, to)
  selectedNode.set(to)
}

/** Applies the name typed in the box: shows the problem under it, or does the operation. */
export const commitEdit = async (bridge: Bridge, typed: string): Promise<void> => {
  const edit = get(treeEdit)
  if (!edit) return
  const renaming = edit.kind === 'rename'
  const folder = findNode(get(fileTree), renaming ? dirname(edit.path) : edit.path)
  const problem = nameProblem(edit.kind, typed, folder, renaming ? edit.path : null, edit.isDir)
  if (problem) {
    treeEdit.set({ ...edit, error: problem })
    return
  }
  const name = finalName(edit.kind, typed)
  treeEdit.set(null)
  try {
    if (renaming) await renameEntry(bridge, edit.path, name)
    else await createEntry(bridge, edit.kind, edit.path, name)
  } catch (error) {
    failed(renaming ? 'tree.renameFailed' : 'tree.createFailed', name, error)
  }
}

/** Asks, then sends the file or folder to the Recycle Bin and closes its tabs. */
export const deleteEntry = async (bridge: Bridge, path: string): Promise<void> => {
  const tree = get(fileTree)
  const node = findNode(tree, path)
  if (!node || path === tree?.path) return
  const answer = await confirmDelete(node.name, {
    files: node.isDir ? countFiles(node) : undefined,
    unsaved: hasUnsavedUnder(path)
  })
  if (answer !== 'delete') return
  try {
    await bridge.files.moveToTrash(path)
  } catch (error) {
    failed('tree.deleteFailed', node.name, error)
    return
  }
  await closeOpenFilesUnder(bridge, path)
  await refreshTree(bridge)
}

export const revealEntry = async (bridge: Bridge, path: string): Promise<void> => {
  try {
    await bridge.files.revealInExplorer(path)
  } catch (error) {
    failed('tree.revealFailed', path, error)
  }
}

export const copyPath = async (path: string): Promise<void> => {
  try {
    await navigator.clipboard.writeText(path)
  } catch (error) {
    failed('tree.copyFailed', path, error)
    return
  }
  showNotice({ messageKey: 'tree.pathCopied', values: {}, actions: [], tone: 'info' })
  const shown = get(notice)
  setTimeout(() => get(notice) === shown && dismissNotice(), COPIED_NOTICE_MS)
}
