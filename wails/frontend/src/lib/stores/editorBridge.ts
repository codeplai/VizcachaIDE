// What the stores may ask of the editor on screen. The editor registers itself when it starts; the
// stores (renaming files of a project) and the menus (More) reach it through here.
import type { Change } from '../textEdits'

export interface EditorBridge {
  /**
   * Applies changes to an open file the editor holds, as one undoable step. False when the editor
   * has no state for that file (then the caller edits the file's text itself).
   */
  applyChanges: (path: string, changes: Change[]) => boolean
  /** Undoes the last history step of an open file (shown or not): pops it, appends nothing. */
  undoFile: (path: string) => boolean
  /** Redoes the step undone last in an open file. */
  redoFile: (path: string) => boolean
  /** Starts renaming the symbol at the cursor (F2). */
  renameSymbol: () => void
  /** Finds every use of the symbol at the cursor (Shift+F12). */
  findReferences: () => void
}

let current: EditorBridge | null = null

/** Registers the editor; returns what to call when it goes away. */
export const registerEditorBridge = (editor: EditorBridge): (() => void) => {
  current = editor
  return () => {
    if (current === editor) current = null
  }
}

export const editorBridge = (): EditorBridge | null => current
