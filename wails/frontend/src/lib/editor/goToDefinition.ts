// Ctrl+click (Cmd+click) and F12 jump to where a name is defined.
import type { Extension } from '@codemirror/state'
import { EditorView, keymap } from '@codemirror/view'
import type { SourceLocation } from '../domain'
import { locationAt, type DocumentContext, type LanguageApi } from './documentContext'

export type OpenLocation = (location: SourceLocation) => void

const jumpFrom = async (
  language: LanguageApi,
  file: DocumentContext,
  view: EditorView,
  pos: number,
  open: OpenLocation
): Promise<void> => {
  const path = file.path()
  if (!path) return
  await file.flush()
  const target = await language.definition(locationAt(view.state, pos, path))
  if (target) open(target)
}

export const goToDefinition = (
  language: LanguageApi,
  file: DocumentContext,
  open: OpenLocation
): Extension => [
  EditorView.domEventHandlers({
    mousedown: (event, view) => {
      if (event.button !== 0 || !(event.ctrlKey || event.metaKey)) return false
      const pos = view.posAtCoords({ x: event.clientX, y: event.clientY })
      if (pos === null) return false
      event.preventDefault()
      void jumpFrom(language, file, view, pos, open)
      return true
    }
  }),
  keymap.of([
    {
      key: 'F12',
      run: (view) => {
        void jumpFrom(language, file, view, view.state.selection.main.head, open)
        return true
      }
    }
  ])
]
