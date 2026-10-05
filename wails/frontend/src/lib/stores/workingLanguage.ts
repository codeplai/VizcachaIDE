// The programming language the learner is working in, for File > New (Ctrl+N and the toolbar
// button): the open file's, else the open folder's, else the default of Settings. A new file in a
// Rust project is a .rs, not a .go.
import { derived } from 'svelte/store'
import type { CodeLanguage, FileNode, LanguageProfile } from '../domain'
import { codeLanguageOf, profiles } from './codeLanguages'
import { activePath, fileTree } from './files'
import { settings } from './settings'

/** Files that say what a folder is, checked in its root before counting sources. */
const PROJECT_MARKERS: [string, CodeLanguage][] = [
  ['Cargo.toml', 'rust'],
  ['go.mod', 'go'],
  ['compile_flags.txt', 'cpp'],
  ['compile_commands.json', 'cpp'],
  ['pyproject.toml', 'python'],
  ['requirements.txt', 'python']
]

/**
 * CMakeLists.txt marks a C++ folder only when the folder has C or C++ sources: other languages
 * (a Rust crate with a native library, a Python extension) have one too.
 */
const CMAKE_MARKER = 'CMakeLists.txt'

/** How deep the sources are counted: a project's own folders, not every dependency below. */
const COUNT_DEPTH = 3

const countSources = (
  node: FileNode,
  all: LanguageProfile[],
  counts: Map<CodeLanguage, number>,
  depth: number
): void => {
  for (const child of node.children) {
    if (child.isDir) {
      if (
        depth < COUNT_DEPTH &&
        !child.name.startsWith('.') &&
        child.name !== 'target' &&
        child.name !== 'build'
      ) {
        countSources(child, all, counts, depth + 1)
      }
      continue
    }
    const language = codeLanguageOf(child.path, all)
    if (language) counts.set(language, (counts.get(language) ?? 0) + 1)
  }
}

const hasCppSources = (tree: FileNode, all: LanguageProfile[]): boolean =>
  tree.children.some((node) => !node.isDir && codeLanguageOf(node.path, all) === 'cpp')

/** The language of a folder: its project marker, else the language most of its sources have. */
export const folderCodeLanguage = (
  tree: FileNode | null,
  all: LanguageProfile[]
): CodeLanguage | null => {
  if (!tree) return null
  for (const [file, language] of PROJECT_MARKERS) {
    if (tree.children.some((node) => !node.isDir && node.name === file)) return language
  }
  if (hasCppSources(tree, all) && tree.children.some((node) => node.name === CMAKE_MARKER)) {
    return 'cpp'
  }
  const counts = new Map<CodeLanguage, number>()
  countSources(tree, all, counts, 1)
  let best: CodeLanguage | null = null
  for (const [language, count] of counts) {
    if (best === null || count > (counts.get(best) ?? 0)) best = language
  }
  return best
}

/** The language of a new untitled file: the open file's, the open folder's, the default one. */
export const workingCodeLanguage = derived(
  [activePath, fileTree, profiles, settings],
  ([path, tree, all, current]): CodeLanguage =>
    (path ? codeLanguageOf(path, all) : null) ??
    folderCodeLanguage(tree, all) ??
    current?.defaultCodeLanguage ??
    'go'
)
