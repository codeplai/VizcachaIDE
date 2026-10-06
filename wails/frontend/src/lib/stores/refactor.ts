// Renaming a symbol: asks the language server, then edits every file it touches. Open files are
// edited in their buffers (the editor keeps its undo history), the others on disk.
import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { FileEdit, RenameRefusal, RenameTarget, SourceLocation } from '../domain'
import { samePath } from '../samePath'
import { applyChanges, editsToChanges, type Change } from '../textEdits'
import { codeLanguageOf } from './codeLanguages'
import { editorBridge } from './editorBridge'
import { activePath, baseName, buffers, dirty } from './files'
import { showNotice } from './notice'

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

const openPathOf = (file: string): string | null =>
  Object.keys(get(buffers)).find((path) => samePath(path, file)) ?? null

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
}

/** Edits an open file: through the editor when it has the file (undoable), else its text. */
const editOpenFile = async (
  bridge: Bridge,
  path: string,
  edits: FileEdit['edits']
): Promise<number> => {
  const before = get(buffers)[path] ?? ''
  const changes: Change[] = editsToChanges(before, edits)
  if (changes.length === 0) return 0
  const handled = editorBridge()?.applyChanges(path, changes) ?? false
  const text = applyChanges(before, changes)
  if (!handled || path !== get(activePath)) {
    buffers.update((all) => ({ ...all, [path]: text }))
    await bridge.language.changeDocument(path, text, 0) // the active file's editor syncs itself
  }
  dirty.update((all) => ({ ...all, [path]: true }))
  return changes.length
}

/** The language server reads closed files from disk: let it know they changed. */
const refreshClosedFile = async (bridge: Bridge, path: string): Promise<void> => {
  const text = await bridge.files.readFile(path)
  await bridge.language.openDocument(path, text)
  await bridge.language.closeDocument(path)
}

/** Applies the edits of a rename to every file: open ones in their buffers, the others on disk. */
export const applyFileEdits = async (bridge: Bridge, files: FileEdit[]): Promise<Applied> => {
  const closed: FileEdit[] = []
  let places = 0
  for (const file of files) {
    const open = openPathOf(file.file)
    if (open) places += await editOpenFile(bridge, open, file.edits)
    else closed.push(file)
  }
  if (closed.length > 0) {
    places += (await bridge.files.applyTextEdits(closed)).edits
    for (const file of closed) await refreshClosedFile(bridge, file.file)
  }
  return { places, files: files.length }
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
    showNotice({
      messageKey: 'refactor.renamed',
      values: { places: applied.places, files: applied.files },
      actions: [],
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
