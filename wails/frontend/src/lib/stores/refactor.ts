// Renaming a symbol: asks the language server, then edits every file it touches. Open files are
// edited in their buffers (the editor keeps its undo history), the others on disk.
import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { FileEdit, RenameRefusal, RenameTarget, SourceLocation } from '../domain'
import { applyChanges, editsToChanges, type Change } from '../textEdits'
import { codeLanguageOf } from './codeLanguages'
import { editorBridge } from './editorBridge'
import { activePath, baseName, buffers, dirty } from './files'
import { showNotice } from './notice'
import {
  openPathOf,
  refreshClosedFile,
  rememberRename,
  undoRename,
  type FileSnapshot
} from './refactorUndo'

const REFUSAL_KEYS: Record<Exclude<RenameRefusal, ''>, string> = {
  notRenameable: 'refactor.notRenameable',
  unsupported: 'refactor.unsupported',
  failed: 'refactor.failed'
}

/** Tells why the rename did not happen, in the user's language. */
const explainRefusal = (path: string, refusal: RenameRefusal, detail = ''): void => {
  if (refusal === '') return
  if (refusal === 'unsupported' && codeLanguageOf(path) === 'python') {
    showNotice({
      messageKey: 'refactor.unsupportedPython',
      values: {},
      actions: [],
      detail: 'pip install rope'
    })
    return
  }
  showNotice({ messageKey: REFUSAL_KEYS[refusal], values: { reason: detail }, actions: [] })
}

/** Whether the symbol can be renamed. When it cannot, the notice says why and this is null. */
export const prepareRename = async (
  bridge: Bridge,
  at: SourceLocation
): Promise<RenameTarget | null> => {
  const target = await bridge.language.prepareRename(at)
  if (target.refusal === '') return target
  explainRefusal(at.file, target.refusal, target.detail)
  return null
}

interface Applied {
  places: number
  files: number
  /** What each file was before and after, to undo the rename as a whole. */
  snapshots: FileSnapshot[]
}

/** Edits an open file: through the editor when it has the file (undoable), else its text. */
const editOpenFile = async (
  bridge: Bridge,
  path: string,
  edits: FileEdit['edits']
): Promise<{ places: number; snapshot: FileSnapshot | null }> => {
  const before = get(buffers)[path] ?? ''
  const wasDirty = get(dirty)[path] ?? false
  const changes: Change[] = editsToChanges(before, edits)
  if (changes.length === 0) return { places: 0, snapshot: null }
  const handled = editorBridge()?.applyChanges(path, changes) ?? false
  if (!handled || path !== get(activePath)) {
    const text = applyChanges(before, changes)
    buffers.update((all) => ({ ...all, [path]: text }))
    await bridge.language.changeDocument(path, text, 0) // the active file's editor syncs itself
  }
  dirty.update((all) => ({ ...all, [path]: true }))
  const after = get(buffers)[path] ?? ''
  return { places: changes.length, snapshot: { path, open: true, before, after, wasDirty } }
}

/** Edits the closed files on disk and tells the language server; remembers their old text. */
const editClosedFiles = async (bridge: Bridge, files: FileEdit[]): Promise<Applied> => {
  const snapshots: FileSnapshot[] = []
  const olds = await Promise.all(files.map((file) => bridge.files.readFile(file.file)))
  const summary = await bridge.files.applyTextEdits(files)
  for (const [index, file] of files.entries()) {
    const after = await bridge.files.readFile(file.file)
    snapshots.push({
      path: file.file,
      open: false,
      before: olds[index] ?? '',
      after,
      wasDirty: false
    })
    await refreshClosedFile(bridge, file.file)
  }
  return { places: summary.edits, files: files.length, snapshots }
}

/** Applies the edits of a rename to every file: open ones in their buffers, the others on disk. */
export const applyFileEdits = async (bridge: Bridge, files: FileEdit[]): Promise<Applied> => {
  const closed: FileEdit[] = []
  const applied: Applied = { places: 0, files: files.length, snapshots: [] }
  for (const file of files) {
    const open = openPathOf(file.file)
    if (!open) {
      closed.push(file)
      continue
    }
    const done = await editOpenFile(bridge, open, file.edits)
    applied.places += done.places
    if (done.snapshot) applied.snapshots.push(done.snapshot)
  }
  if (closed.length > 0) {
    const onDisk = await editClosedFiles(bridge, closed)
    applied.places += onDisk.places
    applied.snapshots.push(...onDisk.snapshots)
  }
  return applied
}

/** Renames the symbol at a position everywhere and says how much changed. */
export const renameSymbol = async (
  bridge: Bridge,
  at: SourceLocation,
  newName: string
): Promise<boolean> => {
  try {
    const result = await bridge.language.rename(at, newName)
    if (result.refusal !== '') {
      explainRefusal(at.file, result.refusal, result.detail)
      return false
    }
    const applied = await applyFileEdits(bridge, result.files)
    rememberRename(applied.snapshots)
    showNotice({
      messageKey: 'refactor.renamed',
      values: { places: applied.places, files: applied.files },
      actions: [{ labelKey: 'refactor.undo', run: () => void undoRename(bridge) }],
      tone: 'info'
    })
    return true
  } catch (error) {
    const reason = error instanceof Error ? error.message : String(error)
    showNotice({
      messageKey: 'refactor.failed',
      values: { reason: `${baseName(at.file)}: ${reason}` },
      actions: []
    })
    return false
  }
}
