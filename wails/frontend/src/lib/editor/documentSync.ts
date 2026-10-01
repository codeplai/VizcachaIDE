import type { Extension } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import type { LanguageApi } from './documentContext'

export interface DocumentSync {
  extension: Extension
  flush: () => Promise<void>
}

const SYNC_DELAY_MS = 150

/** Keeps gopls' copy of the document current: edits are batched, and `flush` sends them at once. */
export const documentSync = (language: LanguageApi, getPath: () => string | null): DocumentSync => {
  const versions = new Map<string, number>()
  let pending: { path: string; text: string } | null = null
  let timer: ReturnType<typeof setTimeout> | undefined

  const flush = async (): Promise<void> => {
    clearTimeout(timer)
    if (!pending) return
    const { path, text } = pending
    pending = null
    const version = (versions.get(path) ?? 0) + 1
    versions.set(path, version)
    await language.changeDocument(path, text, version)
  }

  const extension = EditorView.updateListener.of((update) => {
    const path = getPath()
    if (!update.docChanged || !path) return
    pending = { path, text: update.state.doc.toString() }
    clearTimeout(timer)
    timer = setTimeout(() => void flush(), SYNC_DELAY_MS)
  })
  return { extension, flush }
}
