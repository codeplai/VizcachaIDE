import type { Bridge } from '../bridge'
import { baseName, openFile } from './files'
import { showNotice } from './notice'
import { forgetRecentFile } from './recentFiles'

/** Opens a file from "Open recent"; if it is gone, drops it from the list and says so. */
export const openRecentFile = async (bridge: Bridge, path: string): Promise<void> => {
  try {
    await openFile(bridge, path)
  } catch {
    await forgetRecentFile(bridge, path)
    showNotice({ messageKey: 'recent.missing', values: { file: baseName(path) }, actions: [] })
  }
}
