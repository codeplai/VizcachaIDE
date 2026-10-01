import { derived } from 'svelte/store'
import { debugActive } from './debug'
import { problemCount } from './diagnostics'

/** The three states of the prototype, derived from what is happening. */
export type UiMode = 'write' | 'error' | 'debug'

export const uiMode = derived([debugActive, problemCount], ([debugging, problems]): UiMode => {
  if (debugging) return 'debug'
  return problems > 0 ? 'error' : 'write'
})
