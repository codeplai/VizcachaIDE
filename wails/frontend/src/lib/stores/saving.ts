import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import { confirmCloseChanges } from './confirm'
import { activePath, baseName, buffers, closeFile, dirty } from './files'
import { dismissNotice, notice, showNotice } from './notice'
import { settings } from './settings'
import { isUntitled } from './untitled'

const SAVE_NOTICE_KEYS = ['errors.saveFailed', 'errors.formatRejected']

const errorReason = (error: unknown): string =>
  error instanceof Error ? error.message : String(error ?? '')

/** First "line:column" in a gofmt error such as "main.go:5:2: expected ...". */
export const lineFromFormatError = (error: unknown): number | null => {
  const match = /:(\d+):\d+/.exec(errorReason(error))
  return match?.[1] ? Number(match[1]) : null
}

/** Formats the text when the user asked for it. A file Go cannot read is saved as it is. */
const textToSave = async (bridge: Bridge, text: string): Promise<string> => {
  if (!get(settings)?.formatOnSave) return text
  try {
    return await bridge.run.format(text)
  } catch (error) {
    const line = lineFromFormatError(error)
    showNotice({ messageKey: 'errors.formatRejected', values: { line: line ?? '?' }, actions: [] })
    return text
  }
}

const clearSaveNotice = (): void => {
  const current = get(notice)
  if (current && SAVE_NOTICE_KEYS.includes(current.messageKey)) dismissNotice()
}

/** Formats (if asked), writes a file to disk and clears its modified mark. False when it failed. */
export const saveFile = async (bridge: Bridge, path: string): Promise<boolean> => {
  if (isUntitled(path)) return false
  clearSaveNotice()
  const text = await textToSave(bridge, get(buffers)[path] ?? '')
  try {
    await bridge.files.saveFile(path, text)
  } catch (error) {
    showNotice({
      messageKey: 'errors.saveFailed',
      values: { file: baseName(path), reason: errorReason(error) },
      actions: [{ labelKey: 'errors.saveRetry', run: () => void saveFile(bridge, path) }]
    })
    return false
  }
  buffers.update((all) => ({ ...all, [path]: text }))
  dirty.update((all) => ({ ...all, [path]: false }))
  return true
}

/** Saves the file that is open now (Ctrl+S). */
export const saveActiveFile = async (bridge: Bridge): Promise<void> => {
  const path = get(activePath)
  if (path) await saveFile(bridge, path)
}

/** Closes a tab, asking first when it has changes that were not saved. */
export const requestCloseTab = async (bridge: Bridge, path: string): Promise<void> => {
  if (get(dirty)[path] && !isUntitled(path)) {
    const choice = await confirmCloseChanges(baseName(path))
    if (choice === 'cancel') return
    if (choice === 'save' && !(await saveFile(bridge, path))) return
  }
  await closeFile(bridge, path)
}
