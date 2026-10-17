// The editor split: a second editor beside (or above) the first one, as in VS Code. Both show
// the same open files, so a file open in both stays in step through its shared text.
import { derived, get, writable } from 'svelte/store'
import { activePath, buffers } from './files'

export type SplitDirection = 'right' | 'left' | 'down' | 'up'

export interface SplitState {
  direction: SplitDirection
  /** The file the second editor shows. */
  path: string
}

const state = writable<SplitState | null>(null)

/** The split on screen; none when its file was closed. */
export const editorSplit = derived([state, buffers], ([split, texts]) =>
  split && split.path in texts ? split : null
)

/** Splits the editor: the second one shows the open file (or keeps its own, when already split). */
export const splitEditor = (direction: SplitDirection): void => {
  const path = get(editorSplit)?.path ?? get(activePath)
  if (path) state.set({ direction, path })
}

/** Shows another open file in the second editor. */
export const showInSplit = (path: string): void =>
  state.update((split) => (split ? { ...split, path } : split))

export const closeSplit = (): void => state.set(null)

// A closed file leaves no split behind that would come back when the file is opened again.
buffers.subscribe((texts) => {
  const split = get(state)
  if (split && !(split.path in texts)) state.set(null)
})
