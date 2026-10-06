import type { FileNode } from '../domain'

/** Folders that search, replace and Quick open never look into (same list as the backend). */
export const SKIPPED_FOLDERS = new Set([
  '.git',
  'node_modules',
  'build',
  'target',
  'dist',
  '__pycache__',
  '.venv',
  'venv',
  '.idea',
  '.vs'
])

/** Every file of the tree outside the skipped folders, in tree order. */
export const projectFiles = (tree: FileNode | null): FileNode[] => {
  const files: FileNode[] = []
  const visit = (node: FileNode, isRoot: boolean): void => {
    if (!node.isDir) files.push(node)
    else if (isRoot || !SKIPPED_FOLDERS.has(node.name))
      node.children.forEach((c) => visit(c, false))
  }
  if (tree) visit(tree, true)
  return files
}

const normalize = (path: string): string => path.replace(/\\/g, '/')

/** The path of a file relative to the folder, with `/`. */
export const relativePath = (root: string, path: string): string => {
  const from = normalize(root).replace(/\/$/, '')
  const target = normalize(path)
  return target.startsWith(`${from}/`) ? target.slice(from.length + 1) : target
}

/** True when `path` is a file inside `root` (not the root itself, not outside it). */
export const isInsideFolder = (root: string, path: string): boolean =>
  normalize(path).startsWith(`${normalize(root).replace(/\/$/, '')}/`)
