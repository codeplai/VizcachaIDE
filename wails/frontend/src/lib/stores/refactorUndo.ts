// Undo and redo of a whole rename. A rename touches several files, some of them not open, and the
// editor's own history only knows the file on screen: undoing just that one would leave the
// project half renamed and not compiling. So the last rename is remembered as a unit and undone
// as a unit, from the notice's Undo button or with Ctrl+Z in the editor right after it.
//
// The open files are undone with the editor's own history (the rename step is popped, nothing is
// appended), so one more Ctrl+Z goes on with the student's earlier edits. Redo right after such
// an undo redoes the rename in every file as a unit (the inverse, with the same guards); the
// editor's redo is never left to redo it in one file only.
import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import { samePath } from '../samePath'
import { editorBridge } from './editorBridge'
import { activePath, buffers, dirty } from './files'
import { showNotice } from './notice'

/** What one file of the rename was before and after. */
export interface FileSnapshot {
  path: string
  /** Open in the editor (its buffer was edited) or closed (edited on disk). */
  open: boolean
  before: string
  after: string
  wasDirty: boolean
}

let last: FileSnapshot[] | null = null // the rename that can be undone
let undone: FileSnapshot[] | null = null // the rename that was just undone and can be redone

export const rememberRename = (snapshots: FileSnapshot[]): void => {
  last = snapshots
  undone = null
}

export const forgetRename = (): void => {
  last = null
  undone = null
}

/** The path of an open file that a server may spell differently. */
export const openPathOf = (file: string): string | null =>
  Object.keys(get(buffers)).find((path) => samePath(path, file)) ?? null

/** The language server reads closed files from disk: let it know they changed. */
export const refreshClosedFile = async (bridge: Bridge, path: string): Promise<void> => {
  const text = await bridge.files.readFile(path)
  await bridge.language.openDocument(path, text)
  await bridge.language.closeDocument(path)
}

/** Every open file of the rename still holds its text of `side`: nothing was edited since. */
const holds = (snapshots: FileSnapshot[], side: 'before' | 'after'): boolean => {
  const current = get(buffers)
  return snapshots.every((file) => !file.open || current[file.path] === file[side])
}

/** `holds`, and the file on screen is one of the open files of the rename (for the keys). */
const isFresh = (snapshots: FileSnapshot[] | null, side: 'before' | 'after'): boolean => {
  const shown = get(activePath)
  return (
    snapshots !== null &&
    shown !== null &&
    snapshots.some((file) => file.open && file.path === shown) &&
    holds(snapshots, side)
  )
}

/** Whether Ctrl+Z should undo the whole rename (nothing in its open files was edited since). */
export const hasFreshRename = (): boolean => isFresh(last, 'after')

/** Whether Ctrl+Y should redo the whole rename that was just undone. */
export const hasFreshUndo = (): boolean => isFresh(undone, 'before')

/** Moves an open file to `target` through the editor's history; falls back to setting its text. */
const moveOpenFile = async (
  bridge: Bridge,
  file: FileSnapshot,
  target: string,
  how: 'undoFile' | 'redoFile'
): Promise<void> => {
  editorBridge()?.[how](file.path)
  if ((get(buffers)[file.path] ?? '') !== target) {
    buffers.update((all) => ({ ...all, [file.path]: target }))
    await bridge.language.changeDocument(file.path, target, 0)
  }
}

/** A closed file is rewritten only if nobody changed it since; false when it was changed. */
const moveClosedFile = async (
  bridge: Bridge,
  file: FileSnapshot,
  from: string,
  to: string
): Promise<boolean> => {
  const current = await bridge.files.readFile(file.path).catch(() => null)
  if (current !== from) return false
  await bridge.files.saveFile(file.path, to)
  await refreshClosedFile(bridge, file.path)
  return true
}

const finish = (skipped: number, messageKey: string): void =>
  showNotice(
    skipped === 0
      ? { messageKey, values: {}, actions: [], tone: 'info' }
      : { messageKey: 'refactor.undoPartial', values: { files: skipped }, actions: [] }
  )

/** Undoes the last rename in every file. False when there is nothing (fresh) to undo. */
export const undoRename = async (bridge: Bridge): Promise<boolean> => {
  const snapshots = last
  if (!snapshots) return false
  if (!holds(snapshots, 'after')) {
    last = null // edited since: a blind restore would lose that work
    showNotice({ messageKey: 'refactor.undoStale', values: {}, actions: [] })
    return false
  }
  last = null
  let skipped = 0
  for (const file of snapshots) {
    if (file.open) {
      await moveOpenFile(bridge, file, file.before, 'undoFile')
      dirty.update((all) => ({ ...all, [file.path]: file.wasDirty }))
    } else if (!(await moveClosedFile(bridge, file, file.after, file.before))) skipped++
  }
  undone = snapshots
  finish(skipped, 'refactor.undone')
  return true
}

/** Redoes the rename that was just undone, in every file. */
export const redoRename = async (bridge: Bridge): Promise<boolean> => {
  const snapshots = undone
  if (!snapshots || !holds(snapshots, 'before')) return false
  undone = null
  let skipped = 0
  for (const file of snapshots) {
    if (file.open) {
      await moveOpenFile(bridge, file, file.after, 'redoFile')
      dirty.update((all) => ({ ...all, [file.path]: true }))
    } else if (!(await moveClosedFile(bridge, file, file.before, file.after))) skipped++
  }
  last = snapshots
  finish(skipped, 'refactor.redone')
  return true
}
