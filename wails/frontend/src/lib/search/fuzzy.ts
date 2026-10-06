// Fuzzy matching of a short query against a relative file path (Quick open).

export interface FuzzyMatch {
  score: number
  /** Positions of `target` that matched, ascending. */
  indices: number[]
}

const BOUNDARY = /[/\\._\- ]/

const isBoundary = (target: string, at: number): boolean => {
  if (at === 0) return true
  const before = target[at - 1] ?? ''
  const here = target[at] ?? ''
  return BOUNDARY.test(before) || (before === before.toLowerCase() && here !== here.toLowerCase())
}

const GAP_PENALTY = 1
const CONSECUTIVE_BONUS = 8
const BOUNDARY_BONUS = 10
const NAME_BONUS = 6
const CASE_BONUS = 1
const NONE = Number.NEGATIVE_INFINITY

interface Row {
  score: number[]
  /** Where the previous query letter matched, to rebuild the indices. */
  from: number[]
}

/**
 * Letters of `query` must appear in `target` in order (case does not matter). The score favours
 * consecutive letters, letters that start a word or a path segment, and letters of the file name.
 * Returns null when it does not match.
 */
export const fuzzyMatch = (query: string, target: string): FuzzyMatch | null => {
  const q = query.toLowerCase()
  const t = target.toLowerCase()
  if (q === '') return { score: 0, indices: [] }
  if (q.length > t.length) return null
  const nameStart = Math.max(target.lastIndexOf('/'), target.lastIndexOf('\\')) + 1
  const bonusAt = (i: number, j: number): number =>
    1 +
    (isBoundary(target, j) ? BOUNDARY_BONUS : 0) +
    (j >= nameStart ? NAME_BONUS : 0) +
    (target[j] === query[i] ? CASE_BONUS : 0)

  const rows: Row[] = []
  for (let i = 0; i < q.length; i++)
    rows.push(scoreRow(q[i] ?? '', t, rows[i - 1], (j) => bonusAt(i, j)))
  return best(rows)
}

/** The score of putting one query letter at each position of the target, given the row before. */
const scoreRow = (
  letter: string,
  t: string,
  previous: Row | undefined,
  bonus: (j: number) => number
): Row => {
  const row: Row = {
    score: new Array<number>(t.length).fill(NONE),
    from: new Array<number>(t.length).fill(-1)
  }
  let carry = NONE // best earlier-row score, minus the gap it would cost to jump here
  let carryAt = -1
  for (let j = 0; j < t.length; j++) {
    carry -= GAP_PENALTY
    const jump = (previous?.score[j - 2] ?? NONE) - GAP_PENALTY
    if (j >= 2 && jump > carry) {
      carry = jump
      carryAt = j - 2
    }
    if (t[j] !== letter) continue
    if (!previous) {
      row.score[j] = bonus(j) - Math.min(j, 3) * GAP_PENALTY
      continue
    }
    const adjacent = j >= 1 ? (previous.score[j - 1] ?? NONE) + CONSECUTIVE_BONUS : NONE
    const top = Math.max(adjacent, carry)
    if (top === NONE) continue
    row.score[j] = top + bonus(j)
    row.from[j] = adjacent >= carry ? j - 1 : carryAt
  }
  return row
}

/** The best end position of the last row, and the path of letters that led to it. */
const best = (rows: Row[]): FuzzyMatch | null => {
  const last = rows[rows.length - 1]
  if (!last) return null
  let end = -1
  for (let j = 0; j < last.score.length; j++)
    if ((last.score[j] ?? NONE) > (last.score[end] ?? NONE)) end = j
  if (end < 0 || (last.score[end] ?? NONE) === NONE) return null
  const indices: number[] = []
  for (let i = rows.length - 1, j = end; i >= 0; i--) {
    indices.unshift(j)
    j = rows[i]?.from[j] ?? -1
  }
  return { score: last.score[end] ?? 0, indices }
}
