import type { Bridge, MockControls, Unsubscribe } from '../bridge'
import { demoBreakpointLine, type DevQuery } from '../devQuery'
import type { FileNode } from '../domain'
import { editBuffer, fileTree, openFile } from './files'
import { debuggedPath, toggleBreakpoint } from './debug'
import { connectStores } from './index'
import { openDialog } from './layout'
import { requestCloseTab } from './saving'
import { updateSettings } from './settings'

const EXTENSIONS: Record<string, string> = { python: '.py', cpp: '.cpp' }

/** The file to open at start: the project's `main` of the demo's language, else its first file. */
const firstFile = (tree: FileNode, codeLanguage: string): FileNode | undefined => {
  const extension = EXTENSIONS[codeLanguage] ?? '.go'
  return (
    tree.children.find((node) => node.name === `main${extension}`) ??
    tree.children.find((node) => node.name.endsWith(extension))
  )
}

const applyDevQuery = async (
  bridge: Bridge,
  mock: MockControls,
  query: DevQuery,
  path: string | undefined
): Promise<void> => {
  if (query.language) await updateSettings(bridge, { language: query.language })
  if (query.theme) await updateSettings(bridge, { theme: query.theme })
  if (query.firstRun) await updateSettings(bridge, { firstRun: true })
  if (query.dialog) openDialog.set(query.dialog)
  if (query.closeChanges && path) {
    editBuffer(path, '// edited\n')
    void requestCloseTab(bridge, path)
  }
  if (query.scenario === 'debug' && path) {
    debuggedPath.set(path)
    await toggleBreakpoint(bridge, path, demoBreakpointLine(path))
  }
  await mock.play(query.scenario)
}

/** The folder to show at start: the last one used (the backend remembers it), if there is one. */
const initialTree = async (bridge: Bridge, query: DevQuery): Promise<FileNode | null> => {
  if (bridge.isMock && query.noFolder) return null
  try {
    const tree = await bridge.files.listTree('')
    return tree.path ? tree : null
  } catch {
    return null // The last folder was moved or deleted: start without one.
  }
}

/** Connects the stores, opens the project folder and its first Go file. */
export const startApp = async (
  bridge: Bridge,
  mock: MockControls | null,
  query: DevQuery
): Promise<Unsubscribe> => {
  const stop = await connectStores(bridge)
  const tree = await initialTree(bridge, query)
  fileTree.set(tree)
  const first = tree ? firstFile(tree, query.codeLanguage ?? 'go') : undefined
  if (first) await openFile(bridge, first.path)
  if (mock) await applyDevQuery(bridge, mock, query, first?.path)
  return stop
}
