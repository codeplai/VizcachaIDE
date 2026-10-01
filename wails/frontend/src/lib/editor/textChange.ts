export interface TextChange {
  from: number
  to: number
  insert: string
}

/**
 * The smallest edit that turns `before` into `after` (common start and end kept).
 * Replacing only that part keeps the cursor, the scroll and the undo history sensible.
 */
export const minimalChange = (before: string, after: string): TextChange | null => {
  if (before === after) return null
  const shortest = Math.min(before.length, after.length)
  let start = 0
  while (start < shortest && before[start] === after[start]) start++
  let end = 0
  while (
    end < shortest - start &&
    before[before.length - 1 - end] === after[after.length - 1 - end]
  ) {
    end++
  }
  return { from: start, to: before.length - end, insert: after.slice(start, after.length - end) }
}
