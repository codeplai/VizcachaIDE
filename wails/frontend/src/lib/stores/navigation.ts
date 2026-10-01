import { get, writable } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { FileNode, SourceLocation } from '../domain'
import { fileTree, openFile } from './files'

/** Asks the editor to put the cursor on a place. The editor reacts to a new `nonce`. */
export interface RevealRequest {
  location: SourceLocation
  nonce: number
}

export const revealRequest = writable<RevealRequest | null>(null)

let nonce = 0

const normalize = (path: string): string => path.replace(/\\/g, '/').replace(/^\.\//, '')

const findNode = (node: FileNode, wanted: string): FileNode | null => {
  const path = normalize(node.path)
  if (!node.isDir && (path === wanted || path.endsWith(`/${wanted}`))) return node
  for (const child of node.children) {
    const found = findNode(child, wanted)
    if (found) return found
  }
  return null
}

/** Turns what Go prints ("./main.go") into the path of a file of the open project. */
export const resolveProjectPath = (tree: FileNode | null, file: string): string =>
  (tree && findNode(tree, normalize(file))?.path) ?? file

export const goToLocation = async (bridge: Bridge, location: SourceLocation): Promise<void> => {
  const path = resolveProjectPath(get(fileTree), location.file)
  await openFile(bridge, path)
  revealRequest.set({ location: { ...location, file: path }, nonce: ++nonce })
}
