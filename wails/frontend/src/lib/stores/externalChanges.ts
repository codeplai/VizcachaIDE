import { get } from 'svelte/store'
import type { Bridge, Unsubscribe } from '../bridge'
import { Events } from '../events'
import { confirmReloadFile } from './confirm'
import { activePath, baseName, buffers, dirty, openTabs } from './files'
import { isUntitled } from './untitled'

/** Replaces the text of an open file with what is on disk and clears its modified mark. */
const applyDiskText = async (bridge: Bridge, path: string, text: string): Promise<void> => {
  buffers.update((all) => ({ ...all, [path]: text }))
  dirty.update((all) => ({ ...all, [path]: false }))
  if (get(activePath) === path) return // the editor sends the change to the language server
  await bridge.language.closeDocument(path)
  await bridge.language.openDocument(path, text)
}

const asking = new Set<string>()

const readDisk = async (bridge: Bridge, path: string): Promise<string | null> => {
  try {
    return await bridge.files.readFile(path)
  } catch {
    return null // deleted or being replaced: the next event tells
  }
}

/** A watched file changed on disk: reload it, asking first when it has unsaved changes. */
const onFileChanged = async (bridge: Bridge, path: string): Promise<void> => {
  if (!(path in get(buffers)) || asking.has(path)) return
  const text = await readDisk(bridge, path)
  if (text === null || text === get(buffers)[path]) return
  if (!get(dirty)[path]) {
    await applyDiskText(bridge, path, text)
    return
  }
  asking.add(path)
  try {
    const choice = await confirmReloadFile(baseName(path))
    if (choice === 'reload') await applyDiskText(bridge, path, text)
  } finally {
    asking.delete(path)
  }
}

/** Keeps the backend watching the open tabs and reacts to `file:changed`. */
export const connectExternalChanges = (bridge: Bridge): Unsubscribe => {
  let last = ''
  const offTabs = openTabs.subscribe((tabs) => {
    const paths = tabs.filter((path) => !isUntitled(path))
    const key = JSON.stringify(paths)
    if (key === last) return
    last = key
    void bridge.files.watchFiles(paths).catch(() => {})
  })
  const offEvent = bridge.on(Events.fileChanged, ({ path }) => void onFileChanged(bridge, path))
  return () => {
    offTabs()
    offEvent()
  }
}
