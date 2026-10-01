import { hoverTooltip } from '@codemirror/view'
import type { Extension } from '@codemirror/state'
import { locationAt, type DocumentContext, type LanguageApi } from './documentContext'

const docsBox = (text: string): HTMLElement => {
  const box = document.createElement('div')
  box.className = 'cm-doc-tooltip'
  box.textContent = text
  return box
}

/** Documentation tooltip when the mouse rests on a name. */
export const hoverDocs = (language: LanguageApi, file: DocumentContext): Extension =>
  hoverTooltip(async (view, pos) => {
    const path = file.path()
    if (!path) return null
    await file.flush()
    const text = (await language.hover(locationAt(view.state, pos, path))).trim()
    if (!text) return null
    const word = view.state.wordAt(pos)
    return {
      pos: word?.from ?? pos,
      end: word?.to ?? pos,
      above: true,
      create: () => ({ dom: docsBox(text) })
    }
  })
