import type { ReplaceResult, SearchOptions, SearchResult } from '../domainSearch'

/** Mirrors bridge.SearchService (Go). A bad regular expression rejects with `search.invalidPattern`. */
export interface SearchApi {
  /** The occurrences in the text files under `root` (skips build output, dependencies, binaries). */
  search: (root: string, query: string, options: SearchOptions) => Promise<SearchResult>
  /**
   * Rewrites `paths` on disk (never anything outside `root`). With `line` > 0 only the occurrence
   * that starts at that line and column changes; with 0 every occurrence of each file does.
   */
  replace: (
    root: string,
    query: string,
    options: SearchOptions,
    replacement: string,
    paths: string[],
    line: number,
    column: number
  ) => Promise<ReplaceResult>
}
