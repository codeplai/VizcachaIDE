// Undo of a whole rename. A rename touches several files, some of them not open, and the editor's
// own undo only knows the file on screen: undoing just that one would leave the project half
// renamed and not compiling. So the last rename is remembered as a unit and undone as a unit,
// from the notice's Undo button or with Ctrl+Z in the editor right after it.
import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import { minimalChange } from '../editor/textChange'
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

let last: FileSnapshot[] | null = null

export const rememberRename = (snapshots: FileSnapshot[]): void => {
  last = snapshots
}

export const forgetRename = (): void => {
  last = null
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

/**
 * Whether Ctrl+Z should undo the whole rename: the file on screen is one of its files and no open
 * file has been edited since (anything typed after is undone by the editor as usual).
 */
export const hasFreshRename = (): boolean => {
  if (!last) return false
  const open = last.filter((file) => file.open)
  const shown = get(activePath)
  const current = get(buffers)
  return (
    shown !== null &&
    open.some((file) => file.path === shown) &&
    open.every((file) => current[file.path] === file.after)
  )
}

const restoreOpenFile = async (bridge: Bridge, file: FileSnapshot): Promise<void> => {
  const current = get(buffers)[file.path] ?? ''
  const change = minimalChange(current, file.before)
  if (change) {
    const handled = editorBridge()?.applyChanges(file.path, [change]) ?? false
    if (!handled || file.path !== get(activePath)) {
      buffers.update((all) => ({ ...all, [file.path]: file.before }))
      await bridge.language.changeDocument(file.path, file.before, 0)
    }
  }
  dirty.update((all) => ({ ...all, [file.path]: file.wasDirty }))
}

/** A closed file is put back only if nobody changed it since the rename. */
const restoreClosedFile = async (bridge: Bridge, file: FileSnapshot): Promise<boolean> => {
  const current = await bridge.files.readFile(file.path).catch(() => null)
  if (current !== file.after) return false
  await bridge.files.saveFile(file.path, file.before)
  await refreshClosedFile(bridge, file.path)
  return true
}

/** Undoes the last rename in every file. False when there is nothing (fresh) to undo. */
export const undoRename = async (bridge: Bridge): Promise<boolean> => {
  const snapshots = last
  if (!snapshots) return false
  const current = get(buffers)
  if (snapshots.some((file) => file.open && current[file.path] !== file.after)) {
    last = null // edited since: a blind restore would lose that work
    showNotice({ messageKey: 'refactor.undoStale', values: {}, actions: [] })
    return false
  }
  last = null
  let skipped = 0
  for (const file of snapshots) {
    if (file.open) await restoreOpenFile(bridge, file)
    else if (!(await restoreClosedFile(bridge, file))) skipped++
  }
  showNotice(
    skipped === 0
      ? { messageKey: 'refactor.undone', values: {}, actions: [], tone: 'info' }
      : { messageKey: 'refactor.undoPartial', values: { files: skipped }, actions: [] }
  )
  return true
}
