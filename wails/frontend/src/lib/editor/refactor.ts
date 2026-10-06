// F2 renames the symbol under the cursor and Shift+F12 lists where it is used.
import type { EditorState, Extension } from '@codemirror/state'
import { EditorView, keymap } from '@codemirror/view'
import type { RenameTarget, SourceLocation } from '../domain'
import { locationAt, offsetOf, type DocumentContext } from './documentContext'
import { openRenameBox, renameBox } from './renameBox'

/** What the editor needs to refactor; the shell connects it to the backend and the panels. */
export interface RefactorWiring {
  /** Whether the symbol can be renamed. When it cannot it says why itself and answers null. */
  prepareRename: (at: SourceLocation) => Promise<RenameTarget | null>
  /** Renames the symbol everywhere; false when nothing was renamed. */
  renameSymbol: (at: SourceLocation, newName: string) => Promise<boolean>
  /** Shows where the symbol is used. */
  findReferences: (at: SourceLocation, symbol: string) => Promise<void>
}

export interface RefactorCommands {
  extension: Extension
  rename: (view: EditorView) => void
  references: (view: EditorView) => void
}

/** The name the server reports (its range), or the word under the cursor. */
const symbolAt = (
  state: EditorState,
  pos: number,
  target: RenameTarget
): { from: number; name: string } | null => {
  if (target.range) {
    const from = offsetOf(state, target.range.start.line, target.range.start.column)
    const to = offsetOf(state, target.range.end.line, target.range.end.column)
    return { from, name: target.placeholder || state.sliceDoc(from, to) }
  }
  const word = state.wordAt(pos)
  return word ? { from: word.from, name: state.sliceDoc(word.from, word.to) } : null
}

const startRename = async (
  view: EditorView,
  file: DocumentContext,
  wiring: RefactorWiring
): Promise<void> => {
  const path = file.path()
  if (!path) return
  await file.flush()
  const pos = view.state.selection.main.head
  const at = locationAt(view.state, pos, path)
  const target = await wiring.prepareRename(at)
  if (!target) return
  const symbol = symbolAt(view.state, pos, target)
  if (!symbol) return
  openRenameBox(view, {
    from: symbol.from,
    name: symbol.name,
    onSubmit: (name) => {
      if (name !== symbol.name) void wiring.renameSymbol(at, name)
    }
  })
}

const showReferences = async (
  view: EditorView,
  file: DocumentContext,
  wiring: RefactorWiring
): Promise<void> => {
  const path = file.path()
  if (!path) return
  await file.flush()
  const pos = view.state.selection.main.head
  const word = view.state.wordAt(pos)
  const symbol = word ? view.state.sliceDoc(word.from, word.to) : ''
  await wiring.findReferences(locationAt(view.state, pos, path), symbol)
}

export const refactoring = (file: DocumentContext, wiring: RefactorWiring): RefactorCommands => {
  const rename = (view: EditorView): void => void startRename(view, file, wiring)
  const references = (view: EditorView): void => void showReferences(view, file, wiring)
  const extension = [
    renameBox,
    keymap.of([
      {
        key: 'F2',
        run: (view) => {
          rename(view)
          return true
        }
      },
      {
        key: 'Shift-F12',
        run: (view) => {
          references(view)
          return true
        }
      }
    ])
  ]
  return { extension, rename, references }
}
