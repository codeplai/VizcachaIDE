import { get, writable } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { FileMatches, SearchOptions } from '../domainSearch'
import { findInText, isValidQuery } from '../search/matcher'
import { isInsideFolder } from '../search/skip'
import { buffers, fileTree } from './files'
import { sidebarOpen, sidebarView } from './layout'

export type SearchStatus = 'idle' | 'searching' | 'done' | 'error'

export interface SearchState {
  query: string
  replacement: string
  options: SearchOptions
  /** The replace field is shown. */
  showReplace: boolean
  files: FileMatches[]
  truncated: boolean
  status: SearchStatus
  /** An i18n key when status is 'error'. */
  error: string
  /** What the last replacement changed, for the panel's status line. */
  replaced: { files: number; matches: number } | null
}

export const initialSearch = (): SearchState => ({
  query: '',
  replacement: '',
  options: { caseSensitive: false, wholeWord: false, regex: false },
  showReplace: false,
  files: [],
  truncated: false,
  status: 'idle',
  error: '',
  replaced: null
})

export const searchState = writable<SearchState>(initialSearch())

export const patchSearch = (patch: Partial<SearchState>): void =>
  searchState.update((state) => ({ ...state, ...patch }))

/** Changes each time the Search panel should take the keyboard focus. */
export const searchFocus = writable(0)

/** Shows the Search panel and focuses its field (Ctrl+Shift+F). */
export const showSearch = (): void => {
  sidebarView.set('search')
  sidebarOpen.set(true)
  searchFocus.update((n) => n + 1)
}

export const folderRoot = (): string => get(fileTree)?.path ?? ''

/** The open files of the folder, searched in their buffers (they may hold unsaved text). */
const searchOpenFiles = (root: string, query: string, options: SearchOptions): FileMatches[] =>
  Object.entries(get(buffers))
    .filter(([path]) => isInsideFolder(root, path))
    .flatMap(([path, text]) => {
      const matches = findInText(text, query, options)
      return matches.length > 0 ? [{ path, matches }] : []
    })

let latest = 0

/** Searches the open folder with the current query and options. */
export const runSearch = async (bridge: Bridge): Promise<void> => {
  const { query, options } = get(searchState)
  const root = folderRoot()
  const id = ++latest
  if (query === '' || root === '') {
    patchSearch({ files: [], truncated: false, status: 'idle', error: '' })
    return
  }
  if (!isValidQuery(query, options)) {
    patchSearch({ files: [], truncated: false, status: 'error', error: 'search.invalidPattern' })
    return
  }
  patchSearch({ status: 'searching', error: '' })
  try {
    const found = await bridge.search.search(root, query, options)
    if (id !== latest) return
    const open = searchOpenFiles(root, query, options)
    const openPaths = new Set(Object.keys(get(buffers)))
    const files = [...found.files.filter((file) => !openPaths.has(file.path)), ...open].sort(
      (a, b) => a.path.localeCompare(b.path)
    )
    patchSearch({ files, truncated: found.truncated, status: 'done' })
  } catch (error) {
    if (id !== latest) return
    const invalid = String(error).includes('search.invalidPattern')
    patchSearch({
      files: [],
      status: 'error',
      error: invalid ? 'search.invalidPattern' : 'search.failed'
    })
  }
}

export const matchCount = (files: FileMatches[]): number =>
  files.reduce((sum, file) => sum + file.matches.length, 0)
