// Searching a package by name in the Packages dialog: after a quiet moment of typing the
// language's index is asked, the results are listed and the student picks one.
import { get, writable } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { CodeLanguage, PackageInfo } from '../domain'

/** How long the student must stop typing before the index is asked. */
export const SEARCH_DELAY_MS = 400
/** The shortest text worth searching. */
export const SEARCH_MIN_LENGTH = 2

export interface PackageSearchState {
  /** idle: nothing to show; searching: waiting for the index; done: `results`; failed: no answer. */
  status: 'idle' | 'searching' | 'done' | 'failed'
  results: PackageInfo[]
  /** The package the student picked from the list, if the text still names it. */
  chosen: PackageInfo | null
  /** Index in `results` of the option the arrow keys are on, -1 for none. */
  active: number
  /** The text that was searched, for "no results for ...". */
  term: string
}

const IDLE: PackageSearchState = { status: 'idle', results: [], chosen: null, active: -1, term: '' }

/** The name part of what the student typed: "numpy==2.0" or "pkg@v1.2" search for "numpy"/"pkg". */
export const searchTerm = (text: string): string => text.split(/[@=<>~!;[\s]/)[0]?.trim() ?? ''

/** The state with the arrow-key position moved one option (wrapping around). */
const stepActive = (current: PackageSearchState, step: 1 | -1): PackageSearchState => {
  const count = current.results.length
  if (count === 0) return current
  const first = step === 1 ? 0 : count - 1
  return {
    ...current,
    active: current.active < 0 ? first : (current.active + step + count) % count
  }
}

/** Asks the index and turns its answer, or its failure, into the state to show. */
const lookUp = async (
  bridge: Bridge,
  codeLanguage: CodeLanguage,
  term: string
): Promise<PackageSearchState> => {
  try {
    const results = await bridge.packages.search(codeLanguage, term)
    return { status: 'done', results, chosen: null, active: -1, term }
  } catch {
    return { ...IDLE, status: 'failed', term }
  }
}

export interface PackageSearch {
  subscribe: (run: (state: PackageSearchState) => void) => () => void
  /** The text of the package field changed because the student typed. */
  typed: (text: string) => void
  /** The student picked a result: it is shown as chosen and the list closes. */
  choose: (found: PackageInfo) => void
  /** Moves the arrow-key position by one (wraps around). */
  move: (step: 1 | -1) => void
  /** The option the arrow keys are on, or null. */
  activeOption: () => PackageInfo | null
  /** Forgets everything (the dialog closed or the language changed). */
  reset: () => void
}

/** Creates the search of one dialog; `codeLanguage` is read each time a search starts. */
export const createPackageSearch = (
  bridge: Bridge,
  codeLanguage: () => CodeLanguage
): PackageSearch => {
  const state = writable<PackageSearchState>(IDLE)
  let timer: ReturnType<typeof setTimeout> | undefined
  // Only the answer to the latest search counts: older ones arrive late and are ignored.
  let latest = 0

  const stop = () => {
    clearTimeout(timer)
    timer = undefined
    latest += 1
  }

  const ask = async (term: string) => {
    const mine = latest
    state.update((current) => ({ ...current, status: 'searching', term }))
    const answer = await lookUp(bridge, codeLanguage(), term)
    if (mine === latest) state.set(answer)
  }

  return {
    subscribe: state.subscribe,
    typed: (text) => {
      stop()
      const term = searchTerm(text)
      const { chosen } = get(state)
      if (chosen && term === chosen.name) return
      if (term.length < SEARCH_MIN_LENGTH) {
        state.set(IDLE)
        return
      }
      state.update((current) => ({ ...current, chosen: null }))
      timer = setTimeout(() => void ask(term), SEARCH_DELAY_MS)
    },
    choose: (found) => {
      stop()
      state.set({ ...IDLE, chosen: found })
    },
    move: (step) => state.update((current) => stepActive(current, step)),
    activeOption: () => {
      const { results, active } = get(state)
      return results[active] ?? null
    },
    reset: () => {
      stop()
      state.set(IDLE)
    }
  }
}
