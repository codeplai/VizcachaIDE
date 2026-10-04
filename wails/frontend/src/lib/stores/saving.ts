import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import { profileOf } from './codeLanguages'
import { confirmCloseChanges } from './confirm'
import { activePath, baseName, buffers, closeFile, dirty, fileTree, openTabs } from './files'
import { askNativeDialog } from './nativeDialogs'
import { dismissNotice, notice, showNotice } from './notice'
import { rememberRecentFile } from './recentFiles'
import { settings } from './settings'
import { isUntitled } from './untitled'

const SAVE_NOTICE_KEYS = ['errors.saveFailed', 'errors.formatRejected']

const errorReason = (error: unknown): string =>
  error instanceof Error ? error.message : String(error ?? '')

/** First "line:column" in a formatter error such as "main.go:5:2: expected ...". */
export const lineFromFormatError = (error: unknown): number | null => {
  const match = /:(\d+):\d+/.exec(errorReason(error))
  return match?.[1] ? Number(match[1]) : null
}

/** Formats the text when the user asked for it and the file's language has a formatter. */
const textToSave = async (bridge: Bridge, path: string, text: string): Promise<string> => {
  if (!get(settings)?.formatOnSave) return text
  if (!profileOf(path)?.capabilities.format) return text
  try {
    return await bridge.run.format(path, text)
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

const writeFile = async (
  bridge: Bridge,
  path: string,
  text: string,
  retry: () => void
): Promise<boolean> => {
  try {
    await bridge.files.saveFile(path, text)
    return true
  } catch (error) {
    showNotice({
      messageKey: 'errors.saveFailed',
      values: { file: baseName(path), reason: errorReason(error) },
      actions: [{ labelKey: 'errors.saveRetry', run: retry }]
    })
    return false
  }
}

const parentFolder = (path: string): string => path.replace(/[\\/][^\\/]*$/, '')

/** The folder the "Save as" dialog starts in: the file's own, or the open folder for a new file. */
const startFolder = (path: string): string =>
  isUntitled(path) ? (get(fileTree)?.path ?? '') : parentFolder(path)

const without = <T>(all: Record<string, T>, path: string): Record<string, T> =>
  Object.fromEntries(Object.entries(all).filter(([key]) => key !== path))

/** Makes the tab of `from` the tab of `to`: same text, no unsaved mark, same place in the tabs. */
const moveTab = async (bridge: Bridge, from: string, to: string, text: string): Promise<void> => {
  const wasOpen = to in get(buffers)
  buffers.update((all) => ({ ...without(all, from), [to]: text }))
  dirty.update((all) => ({ ...without(all, from), [to]: false }))
  openTabs.update((tabs) => tabs.flatMap((tab) => (tab === from ? [to] : tab === to ? [] : [tab])))
  if (get(activePath) === from) activePath.set(to)
  await bridge.language.closeDocument(from)
  if (wasOpen) await bridge.language.closeDocument(to)
  await bridge.language.openDocument(to, text)
  await rememberRecentFile(bridge, to)
}

/**
 * Asks where to save and saves there; the tab becomes that file (Save as, and Save on a new file).
 * Returns the new path, or null when the user cancelled or the save failed.
 */
export const saveAs = async (bridge: Bridge, path: string): Promise<string | null> => {
  const target = await askNativeDialog(
    () => bridge.files.saveFileDialog(baseName(path), startFolder(path)),
    ''
  )
  if (!target) return null
  clearSaveNotice()
  const text = await textToSave(bridge, path, get(buffers)[path] ?? '')
  if (!(await writeFile(bridge, target, text, () => void saveAs(bridge, path)))) return null
  await moveTab(bridge, path, target, text)
  return target
}

/** Saves a tab (asking where when it is new). Returns the path it has now, or null when it failed. */
export const saveTab = async (bridge: Bridge, path: string): Promise<string | null> => {
  if (isUntitled(path)) return saveAs(bridge, path)
  clearSaveNotice()
  const text = await textToSave(bridge, path, get(buffers)[path] ?? '')
  if (!(await writeFile(bridge, path, text, () => void saveFile(bridge, path)))) return null
  buffers.update((all) => ({ ...all, [path]: text }))
  dirty.update((all) => ({ ...all, [path]: false }))
  return path
}

/** Formats (if asked), writes a file to disk and clears its modified mark. False when it failed. */
export const saveFile = async (bridge: Bridge, path: string): Promise<boolean> =>
  (await saveTab(bridge, path)) !== null

/** Saves the file that is open now (Ctrl+S). */
export const saveActiveFile = async (bridge: Bridge): Promise<void> => {
  const path = get(activePath)
  if (path) await saveFile(bridge, path)
}

/** Closes a tab, asking first when it has changes that were not saved. */
export const requestCloseTab = async (bridge: Bridge, path: string): Promise<void> => {
  let current = path
  if (get(dirty)[path]) {
    const choice = await confirmCloseChanges(baseName(path))
    if (choice === 'cancel') return
    if (choice === 'save') {
      const saved = await saveTab(bridge, path)
      if (!saved) return
      current = saved
    }
  }
  await closeFile(bridge, current)
}
