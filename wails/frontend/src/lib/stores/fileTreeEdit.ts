// State and rules of the inline name editing in the file tree (create and rename).
import { writable } from 'svelte/store'
import type { FileNode } from '../domain'

export type TreeEditKind = 'file' | 'folder' | 'rename'

export interface TreeEdit {
  kind: TreeEditKind
  /** Folder that will contain the new entry (create), or the entry being renamed. */
  path: string
  /** Rename only: true when the entry is a folder. */
  isDir: boolean
  /** Translated-by-key error shown under the input. */
  error: { key: string; values: Record<string, string> } | null
}

export const treeEdit = writable<TreeEdit | null>(null)

export const FORBIDDEN_CHARACTERS = '/ \\ : * ? " < > |'
const FORBIDDEN = /[/\\:*?"<>|]/
export const GO_TEMPLATE = 'package main\n\nfunc main() {\n}\n'

export const separatorOf = (path: string): string => (path.includes('\\') ? '\\' : '/')
export const joinPath = (folder: string, name: string): string =>
  folder + separatorOf(folder) + name
export const dirname = (path: string): string => path.slice(0, path.lastIndexOf(separatorOf(path)))
export const isInside = (path: string, folder: string): boolean =>
  path === folder || path.startsWith(folder + separatorOf(folder))

export const findNode = (tree: FileNode | null, path: string): FileNode | null => {
  if (!tree) return null
  if (tree.path === path) return tree
  for (const child of tree.children) {
    const found = findNode(child, path)
    if (found) return found
  }
  return null
}

export const countFiles = (node: FileNode): number =>
  node.isDir ? node.children.reduce((total, child) => total + countFiles(child), 0) : 1

/** A new file without an extension becomes a Go file. */
export const withDefaultExtension = (name: string): string =>
  name.includes('.') ? name : `${name}.go`

export const finalName = (kind: TreeEditKind, typed: string): string => {
  const name = typed.trim()
  return kind === 'file' && name !== '' ? withDefaultExtension(name) : name
}

export interface NameProblem {
  key: string
  values: Record<string, string>
}

/**
 * Checks a typed name against the other entries of the target folder.
 * `renaming` is the entry being renamed (it may keep its own name, or change only its letter case).
 */
export const nameProblem = (
  kind: TreeEditKind,
  typed: string,
  folder: FileNode | null,
  renaming: string | null,
  isDir: boolean
): NameProblem | null => {
  const name = finalName(kind, typed)
  if (name === '' || name === '.' || name === '..') {
    return { key: 'tree.nameEmpty', values: {} }
  }
  if (FORBIDDEN.test(name) || name.endsWith('.')) {
    return { key: 'tree.nameInvalid', values: { chars: FORBIDDEN_CHARACTERS } }
  }
  const taken = folder?.children.some(
    (child) => child.path !== renaming && child.name.toLowerCase() === name.toLowerCase()
  )
  if (!taken) return null
  return { key: isDir ? 'tree.folderExists' : 'tree.nameExists', values: { name } }
}
