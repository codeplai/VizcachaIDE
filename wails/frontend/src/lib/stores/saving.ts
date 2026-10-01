import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import { confirmCloseChanges } from './confirm'
import { baseName, buffers, closeTab, dirty } from './files'
import { showNotice } from './notice'
import { formatOnSave } from './settings'
import { isUntitled } from './untitled'

const errorReason = (error: unknown): string => (error instanceof Error ? error.message : '')

/** Formats the text when the user asked for it. A file Go cannot read is saved as it is. */
const textToSave = async (bridge: Bridge, text: string): Promise<string> => {
  if (!get(formatOnSave)) return text
  try {
    return await bridge.run.format(text)
  } catch {
    return text
  }
}

/** Writes a file to disk. Returns false (and says why) when it could not. */
export const saveFile = async (bridge: Bridge, path: string): Promise<boolean> => {
  if (isUntitled(path)) return false
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

/** Closes a tab, asking first when it has changes that were not saved. */
export const requestCloseTab = async (bridge: Bridge, path: string): Promise<void> => {
  if (get(dirty)[path] && !isUntitled(path)) {
    const choice = await confirmCloseChanges(baseName(path))
    if (choice === 'cancel') return
    if (choice === 'save' && !(await saveFile(bridge, path))) return
  }
  closeTab(path)
}
