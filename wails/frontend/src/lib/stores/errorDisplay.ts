// Whether the editor draws the problems of the code (underlines and messages under the lines).
import { derived, writable } from 'svelte/store'
import { settings } from './settings'

/** A build or run shows the problems whatever the setting says, until the user edits again. */
export const errorsRevealed = writable(false)

/** True when the editor draws the problems: the setting is on (it is by default) or a run revealed them. */
export const showErrorsInCode = derived(
  [settings, errorsRevealed],
  ([current, revealed]) => (current?.inlineErrors ?? true) || revealed
)
