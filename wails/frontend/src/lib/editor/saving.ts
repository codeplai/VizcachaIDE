// Ctrl+S: format with gofmt, then save. If gofmt rejects the code the file is still saved as it is.
import { get, writable } from 'svelte/store'
import type { Bridge } from '../bridge'
import { baseName, buffers, dirty } from '../stores/files'

export type SaveNotice =
  | { kind: 'formatRejected'; line: number | null }
  | { kind: 'saveFailed'; file: string; reason: string }

/** Latest save problem, for the shell to show (errors.formatRejected / errors.saveFailed). */
export const saveNotice = writable<SaveNotice | null>(null)

const messageOf = (error: unknown): string =>
  error instanceof Error ? error.message : String(error)

/** First "line:column" in a gofmt error such as "main.go:5:2: expected ..." */
export const lineFromFormatError = (error: unknown): number | null => {
  const match = /:(\d+):\d+/.exec(messageOf(error))
  return match?.[1] ? Number(match[1]) : null
}

const formatOrKeep = async (bridge: Bridge, text: string): Promise<string> => {
  try {
    return await bridge.run.format(text)
  } catch (error) {
    saveNotice.set({ kind: 'formatRejected', line: lineFromFormatError(error) })
    return text
  }
}

export const saveDocument = async (bridge: Bridge, path: string): Promise<boolean> => {
  saveNotice.set(null)
  const text = await formatOrKeep(bridge, get(buffers)[path] ?? '')
  try {
    await bridge.files.saveFile(path, text)
  } catch (error) {
    saveNotice.set({ kind: 'saveFailed', file: baseName(path), reason: messageOf(error) })
    return false
  }
  buffers.update((all) => ({ ...all, [path]: text }))
  dirty.update((all) => ({ ...all, [path]: false }))
  return true
}
