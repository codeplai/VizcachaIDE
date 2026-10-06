import { fuzzyMatch } from './fuzzy'
import { relativePath } from './skip'

export interface QuickOpenItem {
  path: string
  /** Path inside the open folder, with `/`. */
  relative: string
  name: string
  /** Folder part of `relative` ("" for a file at the top). */
  dir: string
  /** Positions of `relative` that matched the query (to highlight). */
  indices: number[]
}

export const MAX_QUICK_OPEN = 50
const RECENT_BONUS = 4

const itemOf = (root: string, path: string, indices: number[]): QuickOpenItem => {
  const relative = relativePath(root, path)
  const cut = relative.lastIndexOf('/')
  return {
    path,
    relative,
    name: relative.slice(cut + 1),
    dir: relative.slice(0, Math.max(cut, 0)),
    indices
  }
}

/**
 * The files to offer for what was typed. An empty query lists the recently opened files first
 * (newest first), then the others by path. Otherwise the best fuzzy matches come first.
 */
export const rankFiles = (
  root: string,
  paths: string[],
  query: string,
  recents: string[],
  limit = MAX_QUICK_OPEN
): QuickOpenItem[] => {
  const recentRank = new Map(recents.map((path, index) => [path, index]))
  const typed = query.trim()
  if (typed === '') {
    const items = paths.map((path) => itemOf(root, path, []))
    return items
      .sort((a, b) => {
        const ra = recentRank.get(a.path) ?? Infinity
        const rb = recentRank.get(b.path) ?? Infinity
        return ra - rb || a.relative.localeCompare(b.relative)
      })
      .slice(0, limit)
  }
  const scored = paths.flatMap((path) => {
    const item = itemOf(root, path, [])
    const match = fuzzyMatch(typed, item.relative)
    if (!match) return []
    const bonus = recentRank.has(path) ? RECENT_BONUS : 0
    return [{ item: { ...item, indices: match.indices }, score: match.score + bonus }]
  })
  return scored
    .sort(
      (a, b) =>
        b.score - a.score ||
        a.item.relative.length - b.item.relative.length ||
        a.item.relative.localeCompare(b.item.relative)
    )
    .slice(0, limit)
    .map((entry) => entry.item)
}
