// Search part of the mock bridge: reads the mock files, so `npm run dev` can search and replace.
import { findInText, replaceInText } from '../search/matcher'
import { projectFiles } from '../search/skip'
import type { FilesApi, SearchApi } from './types'

export const mockSearch = (files: FilesApi): SearchApi => ({
  search: async (root, query, options) => {
    const result = {
      files: [] as { path: string; matches: ReturnType<typeof findInText> }[],
      truncated: false
    }
    if (query === '') return result
    for (const node of projectFiles(await files.listTree(root))) {
      const matches = findInText(await files.readFile(node.path), query, options)
      if (matches.length > 0) result.files.push({ path: node.path, matches })
    }
    return result
  },
  replace: async (root, query, options, replacement, paths, line, column) => {
    const allowed = new Set(projectFiles(await files.listTree(root)).map((node) => node.path))
    const total = { files: 0, matches: 0 }
    for (const path of paths.filter((item) => allowed.has(item))) {
      const at = line > 0 ? { line, column } : undefined
      const done = replaceInText(await files.readFile(path), query, options, replacement, at)
      if (done.count === 0) continue
      await files.saveFile(path, done.text)
      total.files++
      total.matches += done.count
    }
    return total
  }
})
