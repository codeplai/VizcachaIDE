// Mock of the Files service: an in-memory tree that create, rename and delete really change.
import type { FileNode } from '../domain'
import { SAMPLE_DIR, SAMPLE_SOURCES, sampleTree } from './mockData'
import { PYTHON_DIR, PYTHON_SOURCES, pythonTree } from './mockPython'
import type { SampleLanguage } from './mockScenarios'
import type { FilesApi } from './types'

/** The folder the demo opens with: Go's, or Python's with `?language=python`. */
const PROJECTS: Record<
  SampleLanguage,
  { dir: string; tree: () => FileNode; sources: Record<string, string>; dialogFile: string }
> = {
  go: { dir: SAMPLE_DIR, tree: sampleTree, sources: SAMPLE_SOURCES, dialogFile: 'saludo.go' },
  python: { dir: PYTHON_DIR, tree: pythonTree, sources: PYTHON_SOURCES, dialogFile: 'saludo.py' }
}

const separatorOf = (path: string): string => (path.includes('\\') ? '\\' : '/')
const dirnameOf = (path: string): string => path.slice(0, path.lastIndexOf(separatorOf(path)))
const basenameOf = (path: string): string => path.slice(path.lastIndexOf(separatorOf(path)) + 1)
const isInside = (path: string, folder: string): boolean =>
  path === folder || path.startsWith(folder + separatorOf(folder))

const sortNodes = (nodes: FileNode[]): FileNode[] =>
  [...nodes].sort(
    (a, b) =>
      Number(b.isDir) - Number(a.isDir) || a.name.toLowerCase().localeCompare(b.name.toLowerCase())
  )

const find = (node: FileNode, path: string): FileNode | undefined => {
  if (node.path === path) return node
  for (const child of node.children) {
    const found = find(child, path)
    if (found) return found
  }
  return undefined
}

const rebase = (node: FileNode, from: string, to: string): FileNode => ({
  ...node,
  path: to + node.path.slice(from.length),
  children: node.children.map((child) => rebase(child, from, to))
})

interface TreeEditor {
  add: (path: string, isDir: boolean, text?: string) => void
  remove: (path: string) => FileNode
  rename: (from: string, to: string) => Promise<void>
}

const createTreeEditor = (tree: FileNode, texts: Record<string, string>): TreeEditor => {
  const parentOf = (path: string): FileNode => {
    const parent = find(tree, dirnameOf(path))
    if (!parent) throw new Error(`The folder ${dirnameOf(path)} does not exist.`)
    return parent
  }
  const add: TreeEditor['add'] = (path, isDir, text = '') => {
    const parent = parentOf(path)
    const name = basenameOf(path)
    if (parent.children.some((child) => child.name.toLowerCase() === name.toLowerCase())) {
      throw new Error(`${name} already exists`)
    }
    parent.children = sortNodes([...parent.children, { name, path, isDir, children: [] }])
    if (!isDir) texts[path] = text
  }
  const remove: TreeEditor['remove'] = (path) => {
    const parent = parentOf(path)
    const node = parent.children.find((child) => child.path === path)
    if (!node) throw new Error(`${path} does not exist`)
    parent.children = parent.children.filter((child) => child.path !== path)
    return node
  }
  const rename: TreeEditor['rename'] = async (from, to) => {
    if (from.toLowerCase() !== to.toLowerCase() && find(tree, to)) {
      throw new Error(`${basenameOf(to)} already exists`)
    }
    const moved = { ...rebase(remove(from), from, to), name: basenameOf(to) }
    const parent = parentOf(to)
    parent.children = sortNodes([...parent.children, moved])
    for (const key of Object.keys(texts).filter((key) => isInside(key, from))) {
      texts[to + key.slice(from.length)] = texts[key] ?? ''
      delete texts[key]
    }
  }
  return { add, remove, rename }
}

export const mockFiles = (codeLanguage: SampleLanguage = 'go'): FilesApi => {
  const project = PROJECTS[codeLanguage]
  const tree = project.tree()
  const texts: Record<string, string> = { ...project.sources }
  const editor = createTreeEditor(tree, texts)
  return {
    openFolder: async () => structuredClone(tree),
    listTree: async () => structuredClone(tree),
    readFile: async (path) => texts[path] ?? '',
    saveFile: async (path, text) => void (texts[path] = text),
    watchFiles: async () => {},
    openFileDialog: async () => `${project.dir}/${project.dialogFile}`,
    saveFileDialog: async (name, folder) => `${folder || project.dir}/${name}`,
    createFile: async (path, text) => editor.add(path, false, text),
    createFolder: async (path) => editor.add(path, true),
    rename: editor.rename,
    moveToTrash: async (path) => void editor.remove(path),
    revealInExplorer: async () => {}
  }
}
