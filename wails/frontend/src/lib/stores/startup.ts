import type { Bridge, MockControls, Unsubscribe } from '../bridge'
import type { DevQuery } from '../devQuery'
import type { FileNode } from '../domain'
import { fileTree, openFile } from './files'
import { toggleBreakpoint } from './debug'
import { connectStores, updateSettings } from './index'

const firstGoFile = (tree: FileNode): FileNode | undefined =>
  tree.children.find((node) => node.name === 'main.go') ??
  tree.children.find((node) => node.name.endsWith('.go'))

const applyDevQuery = async (
  bridge: Bridge,
  mock: MockControls,
  query: DevQuery,
  path: string | undefined
): Promise<void> => {
  if (query.language) await updateSettings(bridge, { language: query.language })
  if (query.theme) await updateSettings(bridge, { theme: query.theme })
  if (query.scenario === 'debug' && path) await toggleBreakpoint(bridge, path, 6)
  await mock.play(query.scenario)
}

/** Connects the stores, opens the project folder and its first Go file. */
export const startApp = async (
  bridge: Bridge,
  mock: MockControls | null,
  query: DevQuery
): Promise<Unsubscribe> => {
  const stop = await connectStores(bridge)
  const tree = await bridge.files.openFolder()
  fileTree.set(tree)
  const first = firstGoFile(tree)
  if (first) await openFile(bridge, first.path)
  if (mock) await applyDevQuery(bridge, mock, query, first?.path)
  return stop
}
