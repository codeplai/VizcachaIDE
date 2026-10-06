// Search and replace in the files of the open folder (mirrors internal/domain/search.go).

export interface SearchOptions {
  caseSensitive: boolean
  wholeWord: boolean
  regex: boolean
}

/** One occurrence: 1-based line and column; column and length count UTF-16 units. */
export interface SearchMatch {
  line: number
  column: number
  length: number
  /** The line that holds it (cut when very long). */
  text: string
}

export interface FileMatches {
  path: string
  matches: SearchMatch[]
}

export interface SearchResult {
  files: FileMatches[]
  /** The search stopped at its limit of matches. */
  truncated: boolean
}

export interface ReplaceResult {
  files: number
  matches: number
}
