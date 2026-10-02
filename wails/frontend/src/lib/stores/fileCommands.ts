import { derived, get } from 'svelte/store'
import type { Bridge } from '../bridge'
import { activePath, dirty, openFile, openTabs } from './files'
import { showNotice } from './notice'
import { requestCloseTab, saveAs, saveFile } from './saving'
import { BLANK_PROGRAM, isUntitled, openUntitled } from './untitled'

/** True when the open file has changes to save, or is new and not on disk yet. */
export const activeNeedsSave = derived([activePath, dirty], ([path, marks]) =>
  path ? Boolean(marks[path]) || isUntitled(path) : false
)

/** New file (Ctrl+N): a tab without title with the smallest Go program. */
export const newFile = async (bridge: Bridge): Promise<void> => {
  await openUntitled(bridge, BLANK_PROGRAM)
}

/** Open file (Ctrl+O): asks for a file and opens it in a tab. Cancelling does nothing. */
export const openFileFromDialog = async (bridge: Bridge): Promise<void> => {
  const path = await bridge.files.openFileDialog()
  if (!path) return
  try {
    await openFile(bridge, path)
  } catch {
    showNotice({ messageKey: 'files.openFailed', values: { file: path }, actions: [] })
  }
}

/** Save as (Ctrl+Shift+S) for the open file. */
export const saveActiveAs = async (bridge: Bridge): Promise<void> => {
  const path = get(activePath)
  if (path) await saveAs(bridge, path)
}

/** Save all: stops at the first file that is not saved (for example, a cancelled dialog). */
export const saveAll = async (bridge: Bridge): Promise<void> => {
  for (const path of [...get(openTabs)]) {
    if (!get(dirty)[path]) continue
    if (!(await saveFile(bridge, path))) return
  }
}

/** Close (Ctrl+W) the open file. */
export const closeActive = async (bridge: Bridge): Promise<void> => {
  const path = get(activePath)
  if (path) await requestCloseTab(bridge, path)
}

/** Close all: asks about each file with changes; Cancel stops and keeps the rest open. */
export const closeAll = async (bridge: Bridge): Promise<void> => {
  for (const path of [...get(openTabs)]) {
    await requestCloseTab(bridge, path)
    if (get(openTabs).includes(path)) return
  }
}
