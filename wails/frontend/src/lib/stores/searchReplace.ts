// Replace in the Search panel. Open files change in their buffers (the editor shows the change,
// it can be undone and the file is marked as modified); the others change on disk.
import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { FileMatches, SearchMatch } from '../domainSearch'
import { replaceInText, type ReplaceAt } from '../search/matcher'
import { confirmReplaceInFiles } from './confirm'
import { activePath, buffers, dirty } from './files'
import { folderRoot, matchCount, patchSearch, runSearch, searchState } from './search'

/** Puts new text in an open file: not saved yet, and the language server is told. */
const applyToBuffer = async (bridge: Bridge, path: string, text: string): Promise<void> => {
  buffers.update((all) => ({ ...all, [path]: text }))
  dirty.update((all) => ({ ...all, [path]: true }))
  if (get(activePath) === path) return // the editor sends the change to the language server
  await bridge.language.closeDocument(path)
  await bridge.language.openDocument(path, text)
}

interface Done {
  files: number
  matches: number
}

/** Replaces in the paths that are open (buffers) and the rest (disk); `at` limits it to one match. */
const replaceIn = async (bridge: Bridge, paths: string[], at?: ReplaceAt): Promise<Done> => {
  const { query, options, replacement } = get(searchState)
  const open = paths.filter((path) => path in get(buffers))
  const done: Done = { files: 0, matches: 0 }
  for (const path of open) {
    const result = replaceInText(get(buffers)[path] ?? '', query, options, replacement, at)
    if (result.count === 0) continue
    await applyToBuffer(bridge, path, result.text)
    done.files++
    done.matches += result.count
  }
  const onDisk = paths.filter((path) => !open.includes(path))
  if (onDisk.length > 0) {
    const args = [folderRoot(), query, options, replacement, onDisk] as const
    const result = await bridge.search.replace(...args, at?.line ?? 0, at?.column ?? 0)
    done.files += result.files
    done.matches += result.matches
  }
  return done
}

const finish = async (bridge: Bridge, done: Done): Promise<void> => {
  patchSearch({ replaced: done })
  await runSearch(bridge) // the list shows what is left (the replaced text may still match)
}

const attempt = async (bridge: Bridge, work: () => Promise<Done>): Promise<void> => {
  try {
    await finish(bridge, await work())
  } catch {
    patchSearch({ status: 'error', error: 'search.replaceFailed' })
  }
}

/** Replaces one occurrence (its row of the results). */
export const replaceMatch = (bridge: Bridge, path: string, match: SearchMatch): Promise<void> =>
  attempt(bridge, () => replaceIn(bridge, [path], { line: match.line, column: match.column }))

/** Replaces every occurrence of one file. */
export const replaceInFile = (bridge: Bridge, file: FileMatches): Promise<void> =>
  attempt(bridge, () => replaceIn(bridge, [file.path]))

/** Replaces every occurrence in every file of the results, after asking. */
export const replaceEverywhere = async (bridge: Bridge): Promise<boolean> => {
  const { files } = get(searchState)
  if (files.length === 0) return false
  const answer = await confirmReplaceInFiles(matchCount(files), files.length)
  if (answer !== 'replace') return false
  await attempt(bridge, () =>
    replaceIn(
      bridge,
      files.map((file) => file.path)
    )
  )
  return true
}
